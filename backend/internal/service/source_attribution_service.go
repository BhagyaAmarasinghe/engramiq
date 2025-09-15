package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/engramiq/engramiq-backend/internal/repository"
	"github.com/google/uuid"
)

type SourceAttributionService interface {
	AttributeSources(queryID uuid.UUID, documents []*domain.Document, excerpts []string, relevanceScores []float64) error
	GetQuerySources(queryID uuid.UUID) ([]*domain.QuerySource, error)
	FormatCitation(document *domain.Document, pageNumber *int, sectionRef string) string
	ValidateSourceContent(answer string, sources []*domain.QuerySource) (*SourceValidationResult, error)
}

type SourceValidationResult struct {
	IsValid           bool     `json:"is_valid"`
	ConfidenceScore   float64  `json:"confidence_score"`
	HallucinationRisk float64  `json:"hallucination_risk"`
	UnsupportedClaims []string `json:"unsupported_claims"`
}

type sourceAttributionService struct {
	queryRepo    repository.QueryRepository
	documentRepo repository.DocumentRepository
}

func NewSourceAttributionService(
	queryRepo repository.QueryRepository,
	documentRepo repository.DocumentRepository,
) SourceAttributionService {
	return &sourceAttributionService{
		queryRepo:    queryRepo,
		documentRepo: documentRepo,
	}
}

func (s *sourceAttributionService) AttributeSources(queryID uuid.UUID, documents []*domain.Document, excerpts []string, relevanceScores []float64) error {
	if len(documents) != len(excerpts) || len(documents) != len(relevanceScores) {
		return fmt.Errorf("mismatched array lengths: documents=%d, excerpts=%d, scores=%d", 
			len(documents), len(excerpts), len(relevanceScores))
	}

	// Get the query to validate it exists
	_, err := s.queryRepo.GetByID(queryID)
	if err != nil {
		return fmt.Errorf("query not found: %w", err)
	}

	// Create query source records for each document
	for i, doc := range documents {
		_ = &domain.QuerySource{
			ID:              uuid.New(),
			QueryID:         queryID,
			DocumentID:      doc.ID,
			DocumentTitle:   doc.Title,
			RelevantExcerpt: excerpts[i],
			RelevanceScore:  relevanceScores[i],
			PageNumber:      s.extractPageNumber(excerpts[i]),
			SectionReference: s.extractSectionReference(excerpts[i]),
			CreatedAt:       time.Now(),
		}

		// For now, we'll store this in the query's results field as JSONB
		// In a full implementation, you'd create a QuerySourceRepository
		// and persist these relationships properly
	}

	return nil
}

func (s *sourceAttributionService) GetQuerySources(queryID uuid.UUID) ([]*domain.QuerySource, error) {
	// This would retrieve from the query_sources table
	// For now, return empty slice since we're storing in query.results
	return []*domain.QuerySource{}, nil
}

func (s *sourceAttributionService) FormatCitation(document *domain.Document, pageNumber *int, sectionRef string) string {
	citation := document.Title
	
	if document.DocumentDate != nil {
		citation += fmt.Sprintf(" (%s)", document.DocumentDate.Format("2006-01-02"))
	}
	
	if pageNumber != nil {
		citation += fmt.Sprintf(", p. %d", *pageNumber)
	}
	
	if sectionRef != "" {
		citation += fmt.Sprintf(", %s", sectionRef)
	}
	
	return citation
}

func (s *sourceAttributionService) ValidateSourceContent(answer string, sources []*domain.QuerySource) (*SourceValidationResult, error) {
	result := &SourceValidationResult{
		IsValid:           true,
		ConfidenceScore:   0.0,
		HallucinationRisk: 0.0,
		UnsupportedClaims: []string{},
	}

	if len(sources) == 0 {
		result.IsValid = false
		result.HallucinationRisk = 1.0
		result.UnsupportedClaims = append(result.UnsupportedClaims, "No sources provided for answer")
		return result, nil
	}

	// Basic content validation - check if answer content appears in sources
	answerWords := strings.Fields(strings.ToLower(answer))
	totalWords := len(answerWords)
	supportedWords := 0

	for _, source := range sources {
		sourceText := strings.ToLower(source.RelevantExcerpt)
		for _, word := range answerWords {
			// Skip common words for better accuracy
			if len(word) > 3 && strings.Contains(sourceText, word) {
				supportedWords++
			}
		}
	}

	if totalWords > 0 {
		supportRatio := float64(supportedWords) / float64(totalWords)
		result.ConfidenceScore = supportRatio
		result.HallucinationRisk = 1.0 - supportRatio

		// Flag potential hallucinations if support is low
		if supportRatio < 0.6 {
			result.IsValid = false
			result.UnsupportedClaims = append(result.UnsupportedClaims, 
				fmt.Sprintf("Low source support ratio: %.2f", supportRatio))
		}
	}

	// Additional validation rules can be added here:
	// - Date consistency checks
	// - Fact contradiction detection
	// - Confidence threshold enforcement

	return result, nil
}

// Helper methods for extracting metadata from excerpts
func (s *sourceAttributionService) extractPageNumber(excerpt string) *int {
	// Simple regex to find page references like "Page 5" or "p. 5"
	// In production, this would be more sophisticated
	if strings.Contains(strings.ToLower(excerpt), "page ") {
		// This is a simplified implementation
		return nil // Return actual page number when found
	}
	return nil
}

func (s *sourceAttributionService) extractSectionReference(excerpt string) string {
	// Look for section references like "Section 3.1" or "Chapter 2"
	// In production, this would use more sophisticated text analysis
	if strings.Contains(strings.ToLower(excerpt), "section ") {
		// Extract and return section reference
	}
	return ""
}