package handler

import (
	"strconv"
	"time"

	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/engramiq/engramiq-backend/internal/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ActionHandler struct {
	actionRepo repository.ActionRepository
}

func NewActionHandler(actionRepo repository.ActionRepository) *ActionHandler {
	return &ActionHandler{
		actionRepo: actionRepo,
	}
}

func (h *ActionHandler) GetAction(c *fiber.Ctx) error {
	// Get action ID from params
	actionIDParam := c.Params("id")
	actionID, err := uuid.Parse(actionIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid action ID",
		})
	}

	// Get action with components
	action, err := h.actionRepo.GetByID(actionID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Action not found",
		})
	}

	return c.JSON(action)
}

func (h *ActionHandler) ListActions(c *fiber.Ctx) error {
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
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	sort := c.Query("sort", "action_date DESC, created_at DESC")

	pagination := &domain.Pagination{
		Page:  page,
		Limit: limit,
		Sort:  sort,
	}

	// Parse filters
	filters := make(map[string]interface{})
	if actionType := c.Query("action_type"); actionType != "" {
		filters["action_type"] = actionType
	}
	if status := c.Query("action_status"); status != "" {
		filters["action_status"] = status
	}
	if workOrder := c.Query("work_order_number"); workOrder != "" {
		filters["work_order_number"] = workOrder
	}
	if componentID := c.Query("component_id"); componentID != "" {
		if id, err := uuid.Parse(componentID); err == nil {
			filters["component_id"] = id
		}
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		filters["date_from"] = dateFrom
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		filters["date_to"] = dateTo
	}

	// Get actions
	actions, err := h.actionRepo.ListBySite(siteID, pagination, filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"actions":    actions,
		"pagination": pagination,
	})
}

func (h *ActionHandler) GetActionsByComponent(c *fiber.Ctx) error {
	// Get component ID from params
	componentIDParam := c.Params("componentId")
	componentID, err := uuid.Parse(componentIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid component ID",
		})
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	sort := c.Query("sort", "action_date DESC")

	pagination := &domain.Pagination{
		Page:  page,
		Limit: limit,
		Sort:  sort,
	}

	// Get actions by component
	actions, err := h.actionRepo.ListByComponent(componentID, pagination)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"component_id": componentID,
		"actions":      actions,
		"pagination":   pagination,
	})
}

func (h *ActionHandler) GetActionsByWorkOrder(c *fiber.Ctx) error {
	// Get work order number from params
	workOrderNumber := c.Params("workOrder")
	if workOrderNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Work order number is required",
		})
	}

	// Get actions by work order
	actions, err := h.actionRepo.GetByWorkOrderNumber(workOrderNumber)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"work_order_number": workOrderNumber,
		"actions":           actions,
		"count":             len(actions),
	})
}

func (h *ActionHandler) GetActionTimeline(c *fiber.Ctx) error {
	// Get site ID from params
	siteIDParam := c.Params("siteId")
	siteID, err := uuid.Parse(siteIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid site ID",
		})
	}

	// Parse date range
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")

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

	// Get actions in date range
	actions, err := h.actionRepo.GetByDateRange(siteID, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"site_id": siteID,
		"date_range": fiber.Map{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		},
		"actions": actions,
		"count":   len(actions),
	})
}

func (h *ActionHandler) UpdateAction(c *fiber.Ctx) error {
	// Get action ID from params
	actionIDParam := c.Params("id")
	actionID, err := uuid.Parse(actionIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid action ID",
		})
	}

	// Parse update request
	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update action
	err = h.actionRepo.Update(actionID, updates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Get updated action
	action, err := h.actionRepo.GetByID(actionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(action)
}

func (h *ActionHandler) DeleteAction(c *fiber.Ctx) error {
	// Get action ID from params
	actionIDParam := c.Params("id")
	actionID, err := uuid.Parse(actionIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid action ID",
		})
	}

	// Delete action
	err = h.actionRepo.Delete(actionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

func (h *ActionHandler) SearchActions(c *fiber.Ctx) error {
	// Get site ID from params
	siteIDParam := c.Params("siteId")
	_, err := uuid.Parse(siteIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid site ID",
		})
	}

	// Get search query
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Search query is required",
		})
	}

	// Parse parameters
	_, _ = strconv.Atoi(c.Query("limit", "20"))
	threshold := 0.8
	if t := c.Query("threshold"); t != "" {
		if parsed, parseErr := strconv.ParseFloat(t, 64); parseErr == nil {
			threshold = parsed
		}
	}

	// Generate embedding for search (this would be done by the LLM service)
	// For now, return empty results as this requires the embedding generation
	actions := make([]*domain.ExtractedAction, 0)

	return c.JSON(fiber.Map{
		"query":     query,
		"actions":   actions,
		"count":     len(actions),
		"threshold": threshold,
	})
}