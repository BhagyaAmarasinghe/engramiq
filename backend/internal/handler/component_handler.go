package handler

import (
	"strconv"

	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/engramiq/engramiq-backend/internal/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type ComponentHandler struct {
	componentRepo repository.ComponentRepository
	actionRepo    repository.ActionRepository
}

type CreateComponentRequest struct {
	ExternalID      string                 `json:"external_id" validate:"required"`
	Name            string                 `json:"name" validate:"required"`
	ComponentType   domain.ComponentType   `json:"component_type" validate:"required"`
	Label           string                 `json:"label"`
	GroupName       string                 `json:"group_name"`
	Specifications  map[string]interface{} `json:"specifications"`
	Level           int                    `json:"level"`
	CurrentStatus   domain.ComponentStatus `json:"current_status"`
}

func NewComponentHandler(componentRepo repository.ComponentRepository, actionRepo repository.ActionRepository) *ComponentHandler {
	return &ComponentHandler{
		componentRepo: componentRepo,
		actionRepo:    actionRepo,
	}
}

func (h *ComponentHandler) CreateComponent(c *fiber.Ctx) error {
	// Get site ID from params
	siteIDParam := c.Params("siteId")
	siteID, err := uuid.Parse(siteIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid site ID",
		})
	}

	// Parse request body
	var req CreateComponentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Create component
	component := &domain.SiteComponent{
		ID:              uuid.New(),
		SiteID:          siteID,
		ExternalID:      req.ExternalID,
		Name:            req.Name,
		ComponentType:   req.ComponentType,
		Label:           req.Label,
		Level:           req.Level,
		GroupName:       req.GroupName,
		Specifications:  domain.JSON(req.Specifications),
		ElectricalData:  domain.JSON{},
		PhysicalData:    domain.JSON{},
		CurrentStatus:   req.CurrentStatus,
		Embedding:       pgvector.NewVector(make([]float32, 1536)), // Initialize empty vector
	}

	// Set default status if not provided
	if component.CurrentStatus == "" {
		component.CurrentStatus = domain.ComponentStatusOperational
	}

	err = h.componentRepo.Create(component)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(component)
}

func (h *ComponentHandler) GetComponent(c *fiber.Ctx) error {
	// Get component ID from params
	componentIDParam := c.Params("id")
	componentID, err := uuid.Parse(componentIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid component ID",
		})
	}

	// Get component
	component, err := h.componentRepo.GetByID(componentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Component not found",
		})
	}

	return c.JSON(component)
}

func (h *ComponentHandler) ListComponents(c *fiber.Ctx) error {
	// Get site ID from params
	siteIDParam := c.Params("siteId")
	siteID, err := uuid.Parse(siteIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid site ID",
		})
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	sort := c.Query("sort", "level ASC, name ASC")

	pagination := &domain.Pagination{
		Page:  page,
		Limit: limit,
		Sort:  sort,
	}

	// Parse filters
	filters := make(map[string]interface{})
	if componentType := c.Query("component_type"); componentType != "" {
		filters["component_type"] = componentType
	}
	if status := c.Query("status"); status != "" {
		filters["current_status"] = status
	}
	if level := c.Query("level"); level != "" {
		if levelInt, err := strconv.Atoi(level); err == nil {
			filters["level"] = levelInt
		}
	}

	// Get components
	components, err := h.componentRepo.ListBySite(siteID, pagination, filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"components": components,
		"pagination": pagination,
	})
}

func (h *ComponentHandler) UpdateComponent(c *fiber.Ctx) error {
	// Get component ID from params
	componentIDParam := c.Params("id")
	componentID, err := uuid.Parse(componentIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid component ID",
		})
	}

	// Parse update request
	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update component
	err = h.componentRepo.Update(componentID, updates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Get updated component
	component, err := h.componentRepo.GetByID(componentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(component)
}

func (h *ComponentHandler) DeleteComponent(c *fiber.Ctx) error {
	// Get component ID from params
	componentIDParam := c.Params("id")
	componentID, err := uuid.Parse(componentIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid component ID",
		})
	}

	// Delete component
	err = h.componentRepo.Delete(componentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

func (h *ComponentHandler) GetComponentHierarchy(c *fiber.Ctx) error {
	// Get site ID from params
	siteIDParam := c.Params("siteId")
	siteID, err := uuid.Parse(siteIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid site ID",
		})
	}

	// Get hierarchy
	components, err := h.componentRepo.GetHierarchy(siteID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"components": components,
		"count":      len(components),
	})
}

func (h *ComponentHandler) GetComponentMaintenanceHistory(c *fiber.Ctx) error {
	// Get component ID from params
	componentIDParam := c.Params("id")
	componentID, err := uuid.Parse(componentIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid component ID",
		})
	}

	// Parse limit
	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	// Get maintenance history
	actions, err := h.actionRepo.GetMaintenanceHistory(componentID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"component_id": componentID,
		"actions":      actions,
		"count":        len(actions),
	})
}

func (h *ComponentHandler) BulkCreateComponents(c *fiber.Ctx) error {
	// Get site ID from params
	siteIDParam := c.Params("siteId")
	siteID, err := uuid.Parse(siteIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid site ID",
		})
	}

	// Parse request body
	var req struct {
		Components []CreateComponentRequest `json:"components" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Convert to domain models
	components := make([]*domain.SiteComponent, len(req.Components))
	for i, comp := range req.Components {
		components[i] = &domain.SiteComponent{
			ID:              uuid.New(),
			SiteID:          siteID,
			ExternalID:      comp.ExternalID,
			Name:            comp.Name,
			ComponentType:   comp.ComponentType,
			Label:           comp.Label,
			Level:           comp.Level,
			GroupName:       comp.GroupName,
			Specifications:  domain.JSON(comp.Specifications),
			CurrentStatus:   comp.CurrentStatus,
		}

		// Set default status if not provided
		if components[i].CurrentStatus == "" {
			components[i].CurrentStatus = domain.ComponentStatusOperational
		}
	}

	// Bulk create components
	err = h.componentRepo.BulkCreate(components)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Components created successfully",
		"count":      len(components),
		"components": components,
	})
}