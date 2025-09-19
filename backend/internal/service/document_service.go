package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/engramiq/engramiq-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
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
	docRepo      repository.DocumentRepository
	siteRepo     repository.SiteRepository
	actionRepo   repository.ActionRepository
	llmService   LLMService
}

func NewDocumentService(
	docRepo repository.DocumentRepository,
	siteRepo repository.SiteRepository,
	actionRepo repository.ActionRepository,
	llmService LLMService,
) DocumentService {
	return &documentService{
		docRepo:      docRepo,
		siteRepo:     siteRepo,
		actionRepo:   actionRepo,
		llmService:   llmService,
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

	// Determine what to store as raw content based on file type
	var rawContent string
	fileExt := filepath.Ext(file.Filename)
	if fileExt == ".pdf" || fileExt == ".docx" || fileExt == ".doc" {
		// For binary files, don't store raw content to avoid UTF-8 encoding issues
		rawContent = ""
	} else {
		// For text files, store the original content
		rawContent = string(content)
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
		RawContent:      rawContent,
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
	actions, err := s.llmService.ExtractActions(document.ProcessedContent, document.SiteID)
	if err != nil {
		s.UpdateProcessingStatus(id, domain.ProcessingStatusFailed)
		return fmt.Errorf("failed to extract actions: %w", err)
	}

	// Save extracted actions to database
	extractedCount := 0
	fmt.Printf("Attempting to save %d extracted actions\n", len(actions))
	for i, action := range actions {
		// Associate action with the document it came from
		action.DocumentID = document.ID
		fmt.Printf("Saving action %d: %s\n", i+1, action.Title)
		if err := s.actionRepo.Create(action); err != nil {
			fmt.Printf("Failed to save action %d: %v\n", i+1, err)
			// Continue processing other actions even if one fails
		} else {
			fmt.Printf("Successfully saved action %d\n", i+1)
			extractedCount++
		}
	}
	fmt.Printf("Total actions saved: %d\n", extractedCount)

	// Update document with processing results
	updates := map[string]interface{}{
		"embedding":           embedding,
		"processing_status":   domain.ProcessingStatusCompleted,
		"processing_completed_at": time.Now(),
		// "extracted_actions_count": extractedCount, // Column doesn't exist in database
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
		// Try to extract PDF text, but don't fail if it can't be parsed
		extracted, err := s.extractPDFText(content)
		if err != nil {
			// If PDF extraction fails, return a safe placeholder
			return "[PDF content - text extraction failed: " + err.Error() + "]", nil
		}
		return extracted, nil
	case ".docx", ".doc":
		// TODO: Implement Word document text extraction
		// For now, return empty string for binary Word documents to avoid encoding issues
		return "", fmt.Errorf("Word document text extraction not implemented yet")
	default:
		// For unknown types, check if content is valid UTF-8
		if strings.ToValidUTF8(string(content), "") != string(content) {
			return "", fmt.Errorf("file contains binary content that cannot be processed as text")
		}
		return string(content), nil
	}
}

func (s *documentService) extractPDFText(content []byte) (string, error) {
	// Create a reader from the byte slice
	reader := bytes.NewReader(content)

	// Create a PDF reader
	pdfReader, err := pdf.NewReader(reader, int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %w", err)
	}

	var textContent strings.Builder

	// Extract text from each page
	for pageNum := 1; pageNum <= pdfReader.NumPage(); pageNum++ {
		page := pdfReader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		// Get page content - pass an empty font map since we just want text
		pageText, err := page.GetPlainText(map[string]*pdf.Font{})
		if err != nil {
			// If we can't extract text from this page, continue with others
			continue
		}

		textContent.WriteString(pageText)
		textContent.WriteString("\n\n") // Add spacing between pages
	}

	extracted := strings.TrimSpace(textContent.String())
	if extracted == "" {
		// Return a placeholder instead of error to avoid UTF-8 issues
		return "[PDF content - text extraction failed]", nil
	}

	return extracted, nil
}

func (s *documentService) extractDateFromFilename(filename string) time.Time {
	// Common date patterns in filenames
	// Examples: "report_2023-12-15.pdf", "maintenance_20231215.txt"
	// This is a simplified implementation - real implementation would use regex
	
	// For now, return zero time to use current time
	return time.Time{}
}