package handler

import (
	"strconv"

	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/engramiq/engramiq-backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type DocumentHandler struct {
	docService service.DocumentService
}

func NewDocumentHandler(docService service.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		docService: docService,
	}
}

func (h *DocumentHandler) UploadDocument(c *fiber.Ctx) error {
	// Get site ID from params
	siteIDParam := c.Params("siteId")
	siteID, err := uuid.Parse(siteIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid site ID",
		})
	}

	// Get document type from form
	docTypeStr := c.FormValue("document_type", "field_service_report")
	docType := domain.DocumentType(docTypeStr)

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	// Upload document
	document, err := h.docService.UploadDocument(siteID, file, docType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Trigger async document processing to generate embeddings
	go func() {
		if processErr := h.docService.ProcessDocument(document.ID); processErr != nil {
			// Log error but don't fail the upload response
			// TODO: Add proper logging here
		}
	}()

	return c.Status(fiber.StatusCreated).JSON(document)
}

func (h *DocumentHandler) GetDocument(c *fiber.Ctx) error {
	// Get document ID from params
	docIDParam := c.Params("id")
	docID, err := uuid.Parse(docIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid document ID",
		})
	}

	// Get document
	document, err := h.docService.GetDocument(docID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Document not found",
		})
	}

	return c.JSON(document)
}

func (h *DocumentHandler) ListDocuments(c *fiber.Ctx) error {
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
	sort := c.Query("sort", "created_at DESC")

	pagination := &domain.Pagination{
		Page:  page,
		Limit: limit,
		Sort:  sort,
	}

	// Parse filters
	filters := make(map[string]interface{})
	if docType := c.Query("document_type"); docType != "" {
		filters["document_type"] = docType
	}
	if status := c.Query("processing_status"); status != "" {
		filters["processing_status"] = status
	}

	// Get documents
	documents, err := h.docService.ListDocuments(siteID, pagination, filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"documents":  documents,
		"pagination": pagination,
	})
}

func (h *DocumentHandler) DeleteDocument(c *fiber.Ctx) error {
	// Get document ID from params
	docIDParam := c.Params("id")
	docID, err := uuid.Parse(docIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid document ID",
		})
	}

	// Delete document
	err = h.docService.DeleteDocument(docID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

func (h *DocumentHandler) ProcessDocument(c *fiber.Ctx) error {
	// Get document ID from params
	docIDParam := c.Params("id")
	docID, err := uuid.Parse(docIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid document ID",
		})
	}

	// Process document
	err = h.docService.ProcessDocument(docID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Document processing started",
	})
}

func (h *DocumentHandler) SearchDocuments(c *fiber.Ctx) error {
	// Get site ID from params
	siteIDParam := c.Params("siteId")
	siteID, err := uuid.Parse(siteIDParam)
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
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	searchType := c.Query("type", "fulltext") // fulltext or semantic

	var documents []*domain.Document

	if searchType == "semantic" {
		threshold := 0.8
		if t := c.Query("threshold"); t != "" {
			if parsed, parseErr := strconv.ParseFloat(t, 64); parseErr == nil {
				threshold = parsed
			}
		}
		documents, err = h.docService.SearchDocumentsSemantic(siteID, query, limit, threshold)
	} else {
		documents, err = h.docService.SearchDocuments(siteID, query, limit)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"query":     query,
		"type":      searchType,
		"documents": documents,
		"count":     len(documents),
	})
}