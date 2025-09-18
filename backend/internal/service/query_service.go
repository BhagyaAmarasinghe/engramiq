package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/engramiq/engramiq-backend/internal/repository"
	"github.com/google/uuid"
)

type QueryService interface {
	ProcessQuery(userID uuid.UUID, siteID uuid.UUID, queryText string, queryType domain.QueryType) (*domain.UserQuery, error)
	ProcessEnhancedQuery(userID uuid.UUID, siteID uuid.UUID, queryText string) (*domain.EnhancedQueryResponse, error)
	GetQueryResult(queryID uuid.UUID) (*domain.UserQuery, error)
	GetQueryHistory(userID uuid.UUID, pagination *domain.Pagination) ([]*domain.UserQuery, error)
	SearchSimilarQueries(siteID uuid.UUID, queryText string, limit int) ([]*domain.UserQuery, error)
	GetQueryAnalytics(siteID uuid.UUID, startDate, endDate time.Time) (*domain.QueryAnalytics, error)
}

type queryService struct {
	queryRepo        repository.QueryRepository
	actionRepo       repository.ActionRepository
	docRepo          repository.DocumentRepository
	componentRepo    repository.ComponentRepository
	llmService       LLMService
	contentFilter    ContentFilterService
	sourceAttribution SourceAttributionService
}

type QueryIntent struct {
	Type       string                 `json:"type"`        // timeline, search, maintenance_history, component_status
	Entities   map[string]interface{} `json:"entities"`    // extracted entities (dates, components, etc.)
	Confidence float64                `json:"confidence"`
}


func NewQueryService(
	queryRepo repository.QueryRepository,
	actionRepo repository.ActionRepository,
	docRepo repository.DocumentRepository,
	componentRepo repository.ComponentRepository,
	llmService LLMService,
	contentFilter ContentFilterService,
	sourceAttribution SourceAttributionService,
) QueryService {
	return &queryService{
		queryRepo:        queryRepo,
		actionRepo:       actionRepo,
		docRepo:          docRepo,
		componentRepo:    componentRepo,
		llmService:       llmService,
		contentFilter:    contentFilter,
		sourceAttribution: sourceAttribution,
	}
}

func (s *queryService) ProcessEnhancedQuery(userID uuid.UUID, siteID uuid.UUID, queryText string) (*domain.EnhancedQueryResponse, error) {
	startTime := time.Now()

	// Step 1: Content filtering and validation
	validationResult, err := s.contentFilter.ValidateQuery(queryText)
	if err != nil {
		return nil, fmt.Errorf("query validation failed: %w", err)
	}

	if !validationResult.IsValid {
		return &domain.EnhancedQueryResponse{
			Answer:            fmt.Sprintf("I cannot process this query: %s", validationResult.Reason),
			ConfidenceScore:   0.0,
			Sources:          []domain.QuerySourceDetail{},
			RelatedConcepts:  []string{},
			ExtractedEntities: map[string][]string{},
			ResponseType:     "error",
			NoHallucination:  true,
			ProcessingTimeMs: int(time.Since(startTime).Milliseconds()),
		}, nil
	}

	// Step 2: Enhanced intent analysis using LLM
	intent, err := s.llmService.AnalyzeQueryIntent(queryText, siteID)
	if err != nil {
		return nil, fmt.Errorf("intent analysis failed: %w", err)
	}

	// Step 3: Retrieve relevant documents using RAG pattern
	sources, err := s.retrieveRelevantSources(siteID, queryText, intent)
	if err != nil {
		return nil, fmt.Errorf("source retrieval failed: %w", err)
	}

	// Step 4: Generate response using only retrieved sources
	response, err := s.llmService.GenerateEnhancedResponse(queryText, sources)
	if err != nil {
		return nil, fmt.Errorf("response generation failed: %w", err)
	}

	// Step 5: Apply professional tone enforcement
	response.Answer = s.contentFilter.EnforceProfessionalTone(response.Answer)
	response.Answer = s.contentFilter.SanitizeResponse(response.Answer)

	// Step 6: Store query and sources for traceability
	query := &domain.UserQuery{
		ID:               uuid.New(),
		UserID:           userID,
		SiteID:           siteID,
		QueryText:        queryText,
		QueryType:        domain.QueryType(intent.Type),
		ConfidenceScore:  response.ConfidenceScore,
		ExtractedEntities: convertToJSON(response.ExtractedEntities),
		CreatedAt:        time.Now(),
	}

	// Generate and store embedding
	embedding, _ := s.llmService.GenerateEmbedding(queryText)
	query.Embedding = embedding

	// Save query record
	err = s.queryRepo.Create(query)
	if err != nil {
		return nil, fmt.Errorf("failed to save query: %w", err)
	}

	// Store source attributions
	documents := make([]*domain.Document, len(sources))
	excerpts := make([]string, len(sources))
	relevanceScores := make([]float64, len(sources))

	for i, source := range sources {
		// This would normally retrieve the full document
		// For now, we'll create minimal document records
		documents[i] = &domain.Document{
			ID:    source.DocumentID,
			Title: source.DocumentTitle,
		}
		excerpts[i] = source.RelevantExcerpt
		relevanceScores[i] = source.RelevanceScore
	}

	err = s.sourceAttribution.AttributeSources(query.ID, documents, excerpts, relevanceScores)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to attribute sources: %v\n", err)
	}

	response.ProcessingTimeMs = int(time.Since(startTime).Milliseconds())
	return response, nil
}

func (s *queryService) ProcessQuery(userID uuid.UUID, siteID uuid.UUID, queryText string, queryType domain.QueryType) (*domain.UserQuery, error) {
	// Create query record
	query := &domain.UserQuery{
		ID:        uuid.New(),
		UserID:    userID,
		SiteID:    siteID,
		QueryText: queryText,
		QueryType: queryType,
		CreatedAt: time.Now(),
	}

	// Generate embedding for similarity search
	embedding, err := s.llmService.GenerateEmbedding(queryText)
	if err == nil {
		query.Embedding = embedding
	}

	// Save initial query
	err = s.queryRepo.Create(query)
	if err != nil {
		return nil, fmt.Errorf("failed to create query: %w", err)
	}

	// Process query based on type
	go s.processQueryAsync(query.ID, queryText, siteID)

	return query, nil
}

func (s *queryService) processQueryAsync(queryID uuid.UUID, queryText string, siteID uuid.UUID) {
	// Analyze query intent
	intent, err := s.analyzeQueryIntent(queryText)
	if err != nil {
		s.updateQueryError(queryID, "Failed to analyze query intent")
		return
	}

	// Execute appropriate search based on intent
	var result *QueryResult
	switch intent.Type {
	case "timeline":
		result, err = s.processTimelineQuery(siteID, intent)
	case "maintenance_history":
		result, err = s.processMaintenanceQuery(siteID, intent)
	case "component_status":
		result, err = s.processComponentQuery(siteID, intent)
	case "search":
		result, err = s.processSearchQuery(siteID, queryText, intent)
	default:
		result, err = s.processGeneralQuery(siteID, queryText)
	}

	if err != nil {
		s.updateQueryError(queryID, err.Error())
		return
	}

	// Convert result to JSON map
	var resultMap domain.JSON
	if result.Data != nil {
		// Convert the result data to a map
		jsonBytes, _ := json.Marshal(result.Data)
		json.Unmarshal(jsonBytes, &resultMap)
	} else {
		resultMap = domain.JSON{}
	}
	
	// Update query with results
	s.queryRepo.UpdateResults(queryID, resultMap, result.Count)
}

func (s *queryService) retrieveRelevantSources(siteID uuid.UUID, queryText string, intent *domain.QueryIntent) ([]domain.QuerySourceDetail, error) {
	sources := []domain.QuerySourceDetail{}

	// Generate embedding for semantic search
	embedding, err := s.llmService.GenerateEmbedding(queryText)
	if err != nil {
		return sources, err
	}

	// Search documents using semantic similarity
	documents, err := s.docRepo.SearchSemantic(siteID, embedding, 10, 0.7)
	if err != nil {
		return sources, err
	}
	
	// If no documents found with semantic search, try full-text search as fallback
	if len(documents) == 0 {
		fmt.Printf("No documents found with semantic search, trying full-text search for query '%s'\n", queryText)
		documents, err = s.docRepo.SearchFullText(siteID, queryText, 5)
		if err != nil {
			return sources, err
		}
		fmt.Printf("Full-text search found %d documents\n", len(documents))
	}
	
	// Log found documents for debugging
	fmt.Printf("Found %d documents for query '%s'\n", len(documents), queryText)
	for _, doc := range documents {
		fmt.Printf("- Document: %s (Type: %s, RawContent len: %d, ProcessedContent len: %d)\n", 
			doc.Title, doc.DocumentType, len(doc.RawContent), len(doc.ProcessedContent))
	}

	// Convert documents to source details
	for _, doc := range documents {
		// Load full document if content is missing
		if doc.ProcessedContent == "" && doc.RawContent == "" {
			// Try to reload the document with full content
			fullDoc, err := s.docRepo.GetByID(doc.ID)
			if err == nil && fullDoc != nil {
				doc = fullDoc
			}
		}
		
		// Extract relevant excerpt - try ProcessedContent first, then RawContent
		excerpt := doc.ProcessedContent
		if excerpt == "" {
			excerpt = doc.RawContent
		}

		// If still empty, use title and metadata as fallback
		if excerpt == "" {
			excerpt = fmt.Sprintf("Document: %s (Type: %s)", doc.Title, doc.DocumentType)
			if doc.DocumentMetadata != nil {
				// Add any useful metadata
				if summary, ok := doc.DocumentMetadata["summary"].(string); ok {
					excerpt += "\nSummary: " + summary
				}
			}
		}

		// Extract relevant chunk based on query instead of just truncating
		excerpt = s.extractRelevantChunk(excerpt, queryText, 8000) // Increased from 500 to 8000 chars

		source := domain.QuerySourceDetail{
			DocumentID:       doc.ID,
			DocumentTitle:    doc.Title,
			DocumentType:     string(doc.DocumentType),
			RelevantExcerpt:  excerpt,
			RelevanceScore:   0.8, // Would be calculated from similarity
			Citation:         s.sourceAttribution.FormatCitation(doc, nil, ""),
		}

		if doc.DocumentDate != nil {
			source.DocumentDate = *doc.DocumentDate
		}

		sources = append(sources, source)
	}

	// If we have specific component filters, also search actions
	if len(intent.ComponentFilters) > 0 {
		// Search for relevant maintenance actions - use action_type filter instead of component_type
		actions, err := s.actionRepo.ListBySite(siteID, &domain.Pagination{Limit: 5}, map[string]interface{}{
			"action_type": "maintenance",
		})
		if err == nil {
			for _, action := range actions {
				// Add action as a source
				source := domain.QuerySourceDetail{
					DocumentID:      action.ID, // Using action ID as document ID
					DocumentTitle:   fmt.Sprintf("Maintenance Action: %s", action.ActionType),
					DocumentType:    "maintenance_action",
					RelevantExcerpt: action.Description,
					RelevanceScore:  0.7,
					Citation:        fmt.Sprintf("Action %s (%s)", action.ActionType, action.CreatedAt.Format("2006-01-02")),
				}

				if action.ActionDate != nil {
					source.DocumentDate = *action.ActionDate
				}

				sources = append(sources, source)
			}
		}
	}

	return sources, nil
}

func (s *queryService) analyzeQueryIntent(queryText string) (*QueryIntent, error) {
	// Simple rule-based intent detection
	// In production, this would use ML models or LLM for better accuracy
	
	lowercaseQuery := strings.ToLower(queryText)
	
	intent := &QueryIntent{
		Entities:   make(map[string]interface{}),
		Confidence: 0.8,
	}

	// Timeline queries
	if containsAny(lowercaseQuery, []string{"timeline", "when", "history", "over time", "chronological"}) {
		intent.Type = "timeline"
		intent.Entities["date_range"] = s.extractDateRange(queryText)
		return intent, nil
	}

	// Maintenance history queries
	if containsAny(lowercaseQuery, []string{"maintenance", "repair", "service", "fix", "replace"}) {
		intent.Type = "maintenance_history"
		intent.Entities["components"] = s.extractComponents(queryText)
		return intent, nil
	}

	// Component status queries
	if containsAny(lowercaseQuery, []string{"status", "condition", "health", "performance", "inverter", "combiner"}) {
		intent.Type = "component_status"
		intent.Entities["components"] = s.extractComponents(queryText)
		return intent, nil
	}

	// Default to general search
	intent.Type = "search"
	return intent, nil
}

func (s *queryService) processTimelineQuery(siteID uuid.UUID, intent *QueryIntent) (*QueryResult, error) {
	// Extract date range or use default (last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)
	
	if dateRange, ok := intent.Entities["date_range"].(map[string]time.Time); ok {
		if start, exists := dateRange["start"]; exists {
			startDate = start
		}
		if end, exists := dateRange["end"]; exists {
			endDate = end
		}
	}

	// Get actions in date range
	actions, err := s.actionRepo.GetByDateRange(siteID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return &QueryResult{
		ResultType: "timeline",
		Count:      len(actions),
		Data:       actions,
		Summary:    fmt.Sprintf("Found %d actions between %s and %s", len(actions), startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
	}, nil
}

func (s *queryService) processMaintenanceQuery(siteID uuid.UUID, intent *QueryIntent) (*QueryResult, error) {
	var actions []*domain.ExtractedAction
	var err error

	// If specific components mentioned, filter by them
	if components, ok := intent.Entities["components"].([]string); ok && len(components) > 0 {
		// Find component by name or type
		siteComponents, err := s.componentRepo.ListBySite(siteID, &domain.Pagination{Limit: 1000}, map[string]interface{}{
			"component_type": components[0], // Simplified - take first component
		})
		if err != nil {
			return nil, err
		}

		if len(siteComponents) > 0 {
			actions, err = s.actionRepo.GetMaintenanceHistory(siteComponents[0].ID, 50)
		}
	} else {
		// Get recent maintenance actions for the site
		actions, err = s.actionRepo.ListBySite(siteID, &domain.Pagination{Limit: 50}, map[string]interface{}{
			"action_type": "maintenance",
		})
	}

	if err != nil {
		return nil, err
	}

	return &QueryResult{
		ResultType: "actions",
		Count:      len(actions),
		Data:       actions,
		Summary:    fmt.Sprintf("Found %d maintenance actions", len(actions)),
	}, nil
}

func (s *queryService) processComponentQuery(siteID uuid.UUID, intent *QueryIntent) (*QueryResult, error) {
	// Get site components
	components, err := s.componentRepo.ListBySite(siteID, &domain.Pagination{Limit: 1000}, nil)
	if err != nil {
		return nil, err
	}

	// Filter by specific components if mentioned
	if componentTypes, ok := intent.Entities["components"].([]string); ok && len(componentTypes) > 0 {
		filtered := make([]*domain.SiteComponent, 0)
		for _, comp := range components {
			for _, compType := range componentTypes {
				if strings.Contains(strings.ToLower(string(comp.ComponentType)), strings.ToLower(compType)) {
					filtered = append(filtered, comp)
					break
				}
			}
		}
		components = filtered
	}

	return &QueryResult{
		ResultType: "components",
		Count:      len(components),
		Data:       components,
		Summary:    fmt.Sprintf("Found %d components", len(components)),
	}, nil
}

func (s *queryService) processSearchQuery(siteID uuid.UUID, queryText string, intent *QueryIntent) (*QueryResult, error) {
	// Search documents first
	documents, err := s.docRepo.SearchFullText(siteID, queryText, 20)
	if err != nil {
		return nil, err
	}

	// Search actions
	actions, err := s.actionRepo.ListBySite(siteID, &domain.Pagination{Limit: 20}, map[string]interface{}{
		"search": queryText,
	})
	if err != nil {
		return nil, err
	}

	// Combine results
	combinedResult := map[string]interface{}{
		"documents": documents,
		"actions":   actions,
	}

	totalCount := len(documents) + len(actions)

	return &QueryResult{
		ResultType: "mixed",
		Count:      totalCount,
		Data:       combinedResult,
		Summary:    fmt.Sprintf("Found %d documents and %d actions matching '%s'", len(documents), len(actions), queryText),
	}, nil
}

func (s *queryService) processGeneralQuery(siteID uuid.UUID, queryText string) (*QueryResult, error) {
	// Fallback to semantic search if available
	// This would use embeddings to find similar content
	return s.processSearchQuery(siteID, queryText, &QueryIntent{Type: "search"})
}

func (s *queryService) GetQueryResult(queryID uuid.UUID) (*domain.UserQuery, error) {
	return s.queryRepo.GetByID(queryID)
}

func (s *queryService) GetQueryHistory(userID uuid.UUID, pagination *domain.Pagination) ([]*domain.UserQuery, error) {
	return s.queryRepo.ListByUser(userID, pagination)
}

func (s *queryService) SearchSimilarQueries(siteID uuid.UUID, queryText string, limit int) ([]*domain.UserQuery, error) {
	// Generate embedding for the query
	embedding, err := s.llmService.GenerateEmbedding(queryText)
	if err != nil {
		return nil, err
	}

	return s.queryRepo.SearchSimilarQueries(siteID, embedding, limit, 0.8)
}

func (s *queryService) GetQueryAnalytics(siteID uuid.UUID, startDate, endDate time.Time) (*domain.QueryAnalytics, error) {
	return s.queryRepo.GetQueryAnalytics(siteID, startDate, endDate)
}

func (s *queryService) updateQueryError(queryID uuid.UUID, errorMsg string) {
	s.queryRepo.Update(queryID, map[string]interface{}{
		"error_message": errorMsg,
		"processed_at":  time.Now(),
	})
}

// Helper functions

func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func (s *queryService) extractDateRange(query string) map[string]time.Time {
	// Simple date extraction using regex
	// In production, use more sophisticated NLP
	dateRange := make(map[string]time.Time)
	
	// Look for date patterns like "2023-12-15", "last month", "this year"
	dateRegex := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})`)
	matches := dateRegex.FindAllString(query, -1)
	
	if len(matches) >= 2 {
		if start, err := time.Parse("2006-01-02", matches[0]); err == nil {
			dateRange["start"] = start
		}
		if end, err := time.Parse("2006-01-02", matches[1]); err == nil {
			dateRange["end"] = end
		}
	}
	
	return dateRange
}

func (s *queryService) extractComponents(query string) []string {
	// Extract component mentions from query
	components := make([]string, 0)
	
	componentKeywords := []string{"inverter", "combiner", "transformer", "string", "module", "panel"}
	lowercaseQuery := strings.ToLower(query)
	
	for _, keyword := range componentKeywords {
		if strings.Contains(lowercaseQuery, keyword) {
			components = append(components, keyword)
		}
	}
	
	return components
}

// Helper function to convert map[string][]string to domain.JSON
func convertToJSON(entities map[string][]string) domain.JSON {
	result := make(domain.JSON)
	for key, values := range entities {
		result[key] = values
	}
	return result
}

// extractRelevantChunk intelligently extracts the most relevant portion of a document
// based on the query, rather than just truncating
func (s *queryService) extractRelevantChunk(content, query string, maxChars int) string {
	if len(content) <= maxChars {
		return content
	}

	// Convert to lowercase for better matching
	lowercaseQuery := strings.ToLower(query)

	// Extract key terms from the query
	queryTerms := []string{}

	// Look for specific component references (inverter numbers, etc.)
	if strings.Contains(lowercaseQuery, "inverter") {
		// Extract inverter numbers like "31", "40", "16", etc.
		words := strings.Fields(lowercaseQuery)
		for _, word := range words {
			// Match patterns like "31", "inv-31", "inverter 31", etc.
			if regexp.MustCompile(`\d+`).MatchString(word) {
				queryTerms = append(queryTerms, word)
				queryTerms = append(queryTerms, "inv-"+word)
				queryTerms = append(queryTerms, "inverter "+word)
			}
		}
	}

	// Add technician references
	if strings.Contains(lowercaseQuery, "technician") {
		words := strings.Fields(lowercaseQuery)
		for i, word := range words {
			if word == "technician" && i+1 < len(words) {
				queryTerms = append(queryTerms, "technician "+words[i+1])
			}
		}
	}

	// Add date references
	if strings.Contains(lowercaseQuery, "/") || strings.Contains(lowercaseQuery, "-") {
		words := strings.Fields(lowercaseQuery)
		for _, word := range words {
			if regexp.MustCompile(`\d+[/-]\d+[/-]?\d*`).MatchString(word) {
				queryTerms = append(queryTerms, word)
			}
		}
	}

	// Generic important terms
	importantTerms := []string{"maintenance", "repair", "replace", "install", "issue", "problem", "fault", "work"}
	for _, term := range importantTerms {
		if strings.Contains(lowercaseQuery, term) {
			queryTerms = append(queryTerms, term)
		}
	}

	// Find the best starting position based on query term matches
	bestScore := 0
	bestStart := 0
	windowSize := maxChars

	// Slide a window through the content and score each position
	for i := 0; i+windowSize < len(content); i += 500 { // Step by 500 chars
		score := 0
		window := strings.ToLower(content[i : i+windowSize])

		// Score based on query term matches
		for _, term := range queryTerms {
			matches := strings.Count(window, term)
			score += matches * 10 // Weight query terms heavily
		}

		// Bonus for being near the beginning
		if i < len(content)/4 {
			score += 2
		}

		if score > bestScore {
			bestScore = score
			bestStart = i
		}
	}

	// Extract the chunk
	end := bestStart + windowSize
	if end > len(content) {
		end = len(content)
	}

	chunk := content[bestStart:end]

	// If we don't start at the beginning, add some context
	if bestStart > 0 {
		chunk = "..." + chunk
	}

	// If we don't end at the end, indicate truncation
	if end < len(content) {
		chunk = chunk + "..."
	}

	return chunk
}