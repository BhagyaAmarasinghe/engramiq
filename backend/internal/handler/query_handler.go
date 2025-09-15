package handler

import (
	"strconv"
	"time"

	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/engramiq/engramiq-backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type QueryHandler struct {
	queryService service.QueryService
}

type CreateQueryRequest struct {
	QueryText string            `json:"query_text" validate:"required"`
	QueryType domain.QueryType  `json:"query_type"`
	Enhanced  bool              `json:"enhanced,omitempty"` // Use enhanced processing per PRD
}

func NewQueryHandler(queryService service.QueryService) *QueryHandler {
	return &QueryHandler{
		queryService: queryService,
	}
}

func (h *QueryHandler) CreateQuery(c *fiber.Ctx) error {
	// Get site ID from params
	siteIDParam := c.Params("siteId")
	siteID, err := uuid.Parse(siteIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid site ID",
		})
	}

	// TODO: Get user ID from authentication context
	// For now, use a placeholder
	userID := uuid.New()

	// Parse request body
	var req CreateQueryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Set default query type if not provided
	if req.QueryType == "" {
		req.QueryType = domain.QueryTypeGeneral
	}

	// Use enhanced processing by default per PRD requirements
	if req.Enhanced || req.QueryType == "" {
		// Enhanced query processing with source attribution and no hallucination
		enhancedResponse, err := h.queryService.ProcessEnhancedQuery(userID, siteID, req.QueryText)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusCreated).JSON(enhancedResponse)
	} else {
		// Legacy query processing
		query, err := h.queryService.ProcessQuery(userID, siteID, req.QueryText, req.QueryType)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusCreated).JSON(query)
	}
}

func (h *QueryHandler) GetQuery(c *fiber.Ctx) error {
	// Get query ID from params
	queryIDParam := c.Params("id")
	queryID, err := uuid.Parse(queryIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid query ID",
		})
	}

	// Get query result
	query, err := h.queryService.GetQueryResult(queryID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Query not found",
		})
	}

	return c.JSON(query)
}

func (h *QueryHandler) GetQueryHistory(c *fiber.Ctx) error {
	// TODO: Get user ID from authentication context
	// For now, parse from query parameter
	userIDParam := c.Query("user_id")
	if userIDParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	sort := c.Query("sort", "created_at DESC")

	pagination := &domain.Pagination{
		Page:  page,
		Limit: limit,
		Sort:  sort,
	}

	// Get query history
	queries, err := h.queryService.GetQueryHistory(userID, pagination)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"queries":    queries,
		"pagination": pagination,
	})
}

func (h *QueryHandler) SearchSimilarQueries(c *fiber.Ctx) error {
	// Get site ID from params
	siteIDParam := c.Params("siteId")
	siteID, err := uuid.Parse(siteIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid site ID",
		})
	}

	// Get search query
	queryText := c.Query("q")
	if queryText == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query text is required",
		})
	}

	// Parse limit
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// Search similar queries
	queries, err := h.queryService.SearchSimilarQueries(siteID, queryText, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"query":           queryText,
		"similar_queries": queries,
		"count":          len(queries),
	})
}

func (h *QueryHandler) GetQueryAnalytics(c *fiber.Ctx) error {
	// Get site ID from params
	siteIDParam := c.Params("siteId")
	siteID, err := uuid.Parse(siteIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid site ID",
		})
	}

	// Parse date range
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Default to last 30 days if not provided
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	// Get analytics
	analytics, err := h.queryService.GetQueryAnalytics(siteID, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"analytics":  analytics,
		"date_range": fiber.Map{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		},
	})
}