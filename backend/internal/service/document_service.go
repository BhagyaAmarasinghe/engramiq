package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/engramiq/engramiq-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type DocumentService interface {
	UploadDocument(siteID uuid.UUID, file *multipart.FileHeader, documentType domain.DocumentType) (*domain.Document, error)
	GetDocument(id uuid.UUID) (*domain.Document, error)
	ListDocuments(siteID uuid.UUID, pagination *domain.Pagination, filters map[string]interface{}) ([]*domain.DocumentWithStats, error)
	DeleteDocument(id uuid.UUID) error
	ProcessDocument(id uuid.UUID) error
	SearchDocuments(siteID uuid.UUID, query string, limit int) ([]*domain.Document, error)
	SearchDocumentsSemantic(siteID uuid.UUID, queryText string, limit int, threshold float64) ([]*domain.Document, error)
	SearchDocumentsSemanticWithEmbedding(siteID uuid.UUID, embedding pgvector.Vector, limit int, threshold float64) ([]*domain.Document, error)
	GetPendingProcessing(limit int) ([]*domain.Document, error)
	UpdateProcessingStatus(id uuid.UUID, status domain.ProcessingStatus) error
}

type documentService struct {
	docRepo    repository.DocumentRepository
	siteRepo   repository.SiteRepository
	llmService LLMService
}

func NewDocumentService(
	docRepo repository.DocumentRepository,
	siteRepo repository.SiteRepository,
	llmService LLMService,
) DocumentService {
	return &documentService{
		docRepo:    docRepo,
		siteRepo:   siteRepo,
		llmService: llmService,
	}
}

func (s *documentService) UploadDocument(siteID uuid.UUID, file *multipart.FileHeader, documentType domain.DocumentType) (*domain.Document, error) {
	// Verify site exists
	_, err := s.siteRepo.GetByID(siteID)
	if err != nil {
		return nil, fmt.Errorf("site not found: %w", err)
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Read file content
	content, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Calculate content hash for deduplication
	hash := sha256.Sum256(content)
	contentHash := hex.EncodeToString(hash[:])

	// Check if document already exists
	existingDoc, err := s.docRepo.GetByContentHash(contentHash)
	if err == nil && existingDoc != nil {
		return existingDoc, nil
	}

	// Extract text content based on file type
	textContent, err := s.extractTextContent(content, filepath.Ext(file.Filename))
	if err != nil {
		return nil, fmt.Errorf("failed to extract text content: %w", err)
	}

	// Create document record
	document := &domain.Document{
		ID:               uuid.New(),
		SiteID:          siteID,
		Title:           file.Filename,
		OriginalFilename: file.Filename,
		ContentHash:     contentHash,
		FileSize:        file.Size,
		MimeType:        file.Header.Get("Content-Type"),
		DocumentType:    documentType,
		RawContent:      string(content), // Store original file content
		ProcessedContent: textContent,
		ProcessingStatus: domain.ProcessingStatusPending,
		DocumentMetadata: domain.JSON{}, // Initialize empty JSON
		Embedding:       pgvector.NewVector(make([]float32, 1536)), // Initialize empty vector
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Set document date based on filename or current time
	extractedDate := s.extractDateFromFilename(file.Filename)
	if !extractedDate.IsZero() {
		document.DocumentDate = &extractedDate
	} else {
		now := time.Now()
		document.DocumentDate = &now
	}

	// Store document
	err = s.docRepo.Create(document)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return document, nil
}

func (s *documentService) GetDocument(id uuid.UUID) (*domain.Document, error) {
	return s.docRepo.GetByID(id)
}

func (s *documentService) ListDocuments(siteID uuid.UUID, pagination *domain.Pagination, filters map[string]interface{}) ([]*domain.DocumentWithStats, error) {
	return s.docRepo.ListBySite(siteID, pagination, filters)
}

func (s *documentService) DeleteDocument(id uuid.UUID) error {
	return s.docRepo.Delete(id)
}

func (s *documentService) ProcessDocument(id uuid.UUID) error {
	// Get document
	document, err := s.docRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	// Update status to processing
	err = s.UpdateProcessingStatus(id, domain.ProcessingStatusProcessing)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Generate embeddings for semantic search
	embedding, err := s.llmService.GenerateEmbedding(document.ProcessedContent)
	if err != nil {
		s.UpdateProcessingStatus(id, domain.ProcessingStatusFailed)
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Extract actions from document content
	_, err = s.llmService.ExtractActions(document.ProcessedContent, document.SiteID)
	if err != nil {
		s.UpdateProcessingStatus(id, domain.ProcessingStatusFailed)
		return fmt.Errorf("failed to extract actions: %w", err)
	}

	// Update document with processing results
	updates := map[string]interface{}{
		"embedding":           embedding,
		"processing_status":   domain.ProcessingStatusCompleted,
		"processing_completed_at": time.Now(),
	}

	err = s.docRepo.Update(id, updates)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

func (s *documentService) SearchDocuments(siteID uuid.UUID, query string, limit int) ([]*domain.Document, error) {
	return s.docRepo.SearchFullText(siteID, query, limit)
}

func (s *documentService) SearchDocumentsSemantic(siteID uuid.UUID, queryText string, limit int, threshold float64) ([]*domain.Document, error) {
	// Generate embedding for search query
	embedding, err := s.llmService.GenerateEmbedding(queryText)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	return s.SearchDocumentsSemanticWithEmbedding(siteID, embedding, limit, threshold)
}

func (s *documentService) SearchDocumentsSemanticWithEmbedding(siteID uuid.UUID, embedding pgvector.Vector, limit int, threshold float64) ([]*domain.Document, error) {
	return s.docRepo.SearchSemantic(siteID, embedding, limit, threshold)
}

func (s *documentService) GetPendingProcessing(limit int) ([]*domain.Document, error) {
	return s.docRepo.GetPendingProcessing(limit)
}

func (s *documentService) UpdateProcessingStatus(id uuid.UUID, status domain.ProcessingStatus) error {
	return s.docRepo.UpdateProcessingStatus(id, status)
}

// Helper methods

func (s *documentService) extractTextContent(content []byte, fileExt string) (string, error) {
	switch fileExt {
	case ".txt":
		return string(content), nil
	case ".pdf":
		// TODO: Implement PDF text extraction using a library like pdfcpu or unipdf
		return string(content), nil // Placeholder
	case ".docx", ".doc":
		// TODO: Implement Word document text extraction
		return string(content), nil // Placeholder
	default:
		// For unknown types, try to extract as plain text
		return string(content), nil
	}
}

func (s *documentService) extractDateFromFilename(filename string) time.Time {
	// Common date patterns in filenames
	// Examples: "report_2023-12-15.pdf", "maintenance_20231215.txt"
	// This is a simplified implementation - real implementation would use regex
	
	// For now, return zero time to use current time
	return time.Time{}
}