package main

import (
	"log"
	"os"

	"github.com/engramiq/engramiq-backend/internal/config"
	"github.com/engramiq/engramiq-backend/internal/infrastructure/database"
	"github.com/engramiq/engramiq-backend/internal/infrastructure/cache"
	"github.com/engramiq/engramiq-backend/internal/handler"
	"github.com/engramiq/engramiq-backend/internal/repository"
	"github.com/engramiq/engramiq-backend/internal/service"
	"github.com/engramiq/engramiq-backend/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.New(cfg.Environment)
	defer log.Sync()

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations", "error", err)
	}

	// Initialize Redis cache
	_ = cache.NewRedis(cfg.Redis)

	// Initialize repositories
	siteRepo := repository.NewSiteRepository(db)
	componentRepo := repository.NewComponentRepository(db)
	documentRepo := repository.NewDocumentRepository(db)
	actionRepo := repository.NewActionRepository(db)
	_ = repository.NewEventRepository(db)
	queryRepo := repository.NewQueryRepository(db)
	_ = repository.NewUserRepository(db)

	// Initialize services
	llmService := service.NewLLMService(
		cfg.LLM.APIKey,
		"https://api.openai.com/v1", // Default OpenAI API URL
		cfg.LLM.Model,
		actionRepo,
		componentRepo,
	)
	
	// Initialize new PRD services
	contentFilterService := service.NewContentFilterService()
	sourceAttributionService := service.NewSourceAttributionService(queryRepo, documentRepo)
	
	documentService := service.NewDocumentService(documentRepo, siteRepo, actionRepo, llmService)
	queryService := service.NewQueryService(queryRepo, actionRepo, documentRepo, componentRepo, llmService, contentFilterService, sourceAttributionService)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
		AppName: "Engramiq Reporting Agent",
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.Server.CORSOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Health check
	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"service": "engramiq-reporting-agent",
		})
	})

	// API routes
	api := app.Group("/api/v1")

	// Initialize handlers
	siteHandler := handler.NewSiteHandler(siteRepo)
	documentHandler := handler.NewDocumentHandler(documentService)
	queryHandler := handler.NewQueryHandler(queryService)
	componentHandler := handler.NewComponentHandler(componentRepo, actionRepo)
	actionHandler := handler.NewActionHandler(actionRepo)

	// Site routes
	api.Get("/sites", siteHandler.ListSites)
	api.Get("/sites/:id", siteHandler.GetSite)

	// Document routes
	api.Post("/sites/:siteId/documents", documentHandler.UploadDocument)
	api.Get("/sites/:siteId/documents", documentHandler.ListDocuments)
	api.Get("/documents/:id", documentHandler.GetDocument)
	api.Delete("/documents/:id", documentHandler.DeleteDocument)
	api.Post("/documents/:id/process", documentHandler.ProcessDocument)
	api.Get("/sites/:siteId/documents/search", documentHandler.SearchDocuments)

	// Query routes - specific routes must come before parameterized routes
	api.Post("/sites/:siteId/queries", queryHandler.CreateQuery)
	api.Get("/queries/history", queryHandler.GetQueryHistory)
	api.Get("/queries/:id", queryHandler.GetQuery)
	api.Get("/sites/:siteId/queries/similar", queryHandler.SearchSimilarQueries)
	api.Get("/sites/:siteId/analytics/queries", queryHandler.GetQueryAnalytics)

	// Component routes
	api.Post("/sites/:siteId/components", componentHandler.CreateComponent)
	api.Get("/sites/:siteId/components", componentHandler.ListComponents)
	api.Get("/components/:id", componentHandler.GetComponent)
	api.Put("/components/:id", componentHandler.UpdateComponent)
	api.Delete("/components/:id", componentHandler.DeleteComponent)
	api.Get("/sites/:siteId/components/hierarchy", componentHandler.GetComponentHierarchy)
	api.Get("/components/:id/maintenance-history", componentHandler.GetComponentMaintenanceHistory)
	api.Post("/sites/:siteId/components/bulk", componentHandler.BulkCreateComponents)

	// Action routes
	api.Get("/sites/:siteId/actions", actionHandler.ListActions)
	api.Get("/actions/:id", actionHandler.GetAction)
	api.Get("/components/:componentId/actions", actionHandler.GetActionsByComponent)
	api.Get("/work-orders/:workOrder/actions", actionHandler.GetActionsByWorkOrder)
	api.Get("/sites/:siteId/timeline", actionHandler.GetActionTimeline)
	api.Put("/actions/:id", actionHandler.UpdateAction)
	api.Delete("/actions/:id", actionHandler.DeleteAction)
	api.Get("/sites/:siteId/actions/search", actionHandler.SearchActions)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Info("Starting server", "port", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Failed to start server", "error", err)
	}
}