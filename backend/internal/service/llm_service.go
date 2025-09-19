package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/engramiq/engramiq-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type QueryResult struct {
	Query       string      `json:"query"`
	ResultType  string      `json:"result_type"`
	Count       int         `json:"count"`
	Data        interface{} `json:"data"`
	Summary     string      `json:"summary"`
	Results     domain.JSON `json:"results"`
	ProcessedAt time.Time   `json:"processed_at"`
}

type LLMService interface {
	GenerateEmbedding(text string) (pgvector.Vector, error)
	ExtractActions(content string, siteID uuid.UUID) ([]*domain.ExtractedAction, error)
	ProcessNaturalLanguageQuery(query string, siteID uuid.UUID) (*QueryResult, error)
	SummarizeDocument(content string) (string, error)
	
	// Enhanced methods per PRD requirements
	AnalyzeQueryIntent(query string, siteID uuid.UUID) (*domain.QueryIntent, error)
	ExtractEntities(text string) (map[string][]string, error)
	GenerateEnhancedResponse(query string, sources []domain.QuerySourceDetail) (*domain.EnhancedQueryResponse, error)
	ValidateResponseAgainstSources(answer string, sources []domain.QuerySourceDetail) (float64, error)
}

type llmService struct {
	apiKey      string
	apiURL      string
	model       string
	client      *http.Client
	actionRepo  repository.ActionRepository
	componentRepo repository.ComponentRepository
}

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type EmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

type ActionExtractionResult struct {
	Actions []struct {
		ActionType        string    `json:"action_type"`
		Description       string    `json:"description"`
		ComponentType     string    `json:"component_type"`
		ComponentID       string    `json:"component_id,omitempty"`
		TechnicianNames   []string  `json:"technician_names"`
		WorkOrderNumber   string    `json:"work_order_number,omitempty"`
		ActionDate        string    `json:"action_date"`
		ActionStatus      string    `json:"action_status"`
		ConfidenceScore   float64   `json:"confidence_score"`
		Details           string    `json:"details"`
	} `json:"actions"`
}

func NewLLMService(
	apiKey string,
	apiURL string, 
	model string,
	actionRepo repository.ActionRepository,
	componentRepo repository.ComponentRepository,
) LLMService {
	return &llmService{
		apiKey:        apiKey,
		apiURL:        apiURL,
		model:         model,
		client:        &http.Client{Timeout: 120 * time.Second},
		actionRepo:    actionRepo,
		componentRepo: componentRepo,
	}
}

func (s *llmService) GenerateEmbedding(text string) (pgvector.Vector, error) {
	reqBody := EmbeddingRequest{
		Model: "text-embedding-ada-002",
		Input: []string{text},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.apiURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return pgvector.Vector{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if embeddingResp.Error != nil {
		return pgvector.Vector{}, fmt.Errorf("OpenAI API error: %s", embeddingResp.Error.Message)
	}

	if len(embeddingResp.Data) == 0 {
		return pgvector.Vector{}, fmt.Errorf("no embedding data returned")
	}

	// Convert to pgvector format
	embedding := make([]float32, len(embeddingResp.Data[0].Embedding))
	for i, val := range embeddingResp.Data[0].Embedding {
		embedding[i] = float32(val)
	}

	return pgvector.NewVector(embedding), nil
}

func (s *llmService) ExtractActions(content string, siteID uuid.UUID) ([]*domain.ExtractedAction, error) {
	// Get site components for context
	components, err := s.componentRepo.ListBySite(siteID, &domain.Pagination{Limit: 100}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get site components: %w", err)
	}

	// Build component context for the LLM
	componentContext := s.buildComponentContext(components)

	prompt := fmt.Sprintf(`You are an expert at extracting maintenance actions from solar field service reports. 

CRITICAL: You must respond with ONLY valid JSON, no additional text before or after the JSON.

Context: This is a field service report for a solar site with the following components:
%s

Document content:
%s

Analyze the document and extract all maintenance actions, repairs, replacements, inspections, and other work performed.

Return ONLY this JSON structure (no other text):
{
  "actions": [
    {
      "action_type": "maintenance|repair|replacement|inspection|troubleshoot|other",
      "description": "Brief description of what was done",
      "component_type": "inverter|combiner|panel|transformer|meter|switchgear|monitoring|other",
      "component_id": "external ID if identifiable or empty string",
      "technician_names": ["name1", "name2"],
      "work_order_number": "WO number or empty string",
      "action_date": "2024-11-05T00:00:00Z",
      "action_status": "completed|pending|failed",
      "confidence_score": 0.95,
      "details": "Additional context or empty string"
    }
  ]
}

If no actions are found, return: {"actions": []}

REMEMBER: Return ONLY the JSON, nothing else.`, componentContext, content)

	messages := []Message{
		{Role: "system", Content: "You are a maintenance action extraction specialist for solar power systems."},
		{Role: "user", Content: prompt},
	}

	reqBody := OpenAIRequest{
		Model:    s.model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.apiURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if openAIResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	// Parse the JSON response
	responseContent := openAIResp.Choices[0].Message.Content
	fmt.Printf("LLM ExtractActions Response: %s\n", responseContent)
	fmt.Printf("DEBUG: Response length: %d\n", len(responseContent))
	
	var extractionResult ActionExtractionResult
	if err := json.Unmarshal([]byte(responseContent), &extractionResult); err != nil {
		fmt.Printf("JSON parsing error: %v\nRaw response: %s\n", err, responseContent)
		
		// Try to extract JSON from response if it's wrapped in text
		if startIdx := strings.Index(responseContent, "{"); startIdx >= 0 {
			if endIdx := strings.LastIndex(responseContent, "}"); endIdx > startIdx {
				jsonPart := responseContent[startIdx : endIdx+1]
				fmt.Printf("Attempting to parse extracted JSON: %s\n", jsonPart)
				if err := json.Unmarshal([]byte(jsonPart), &extractionResult); err == nil {
					fmt.Printf("Successfully parsed extracted JSON\n")
				} else {
					fmt.Printf("Failed to parse extracted JSON: %v\n", err)
					// Return empty result instead of failing
					return []*domain.ExtractedAction{}, nil
				}
			} else {
				// Return empty result instead of failing
				return []*domain.ExtractedAction{}, nil
			}
		} else {
			// Return empty result instead of failing
			return []*domain.ExtractedAction{}, nil
		}
	}

	// Convert to domain models
	actions := make([]*domain.ExtractedAction, 0, len(extractionResult.Actions))
	fmt.Printf("DEBUG: Found %d actions in extraction result\n", len(extractionResult.Actions))

	for _, result := range extractionResult.Actions {
		actionDate, _ := time.Parse(time.RFC3339, result.ActionDate)
		if actionDate.IsZero() {
			actionDate = time.Now()
		}

		// Find matching component if specified
		var primaryComponentID *uuid.UUID
		if result.ComponentID != "" {
			component, err := s.componentRepo.GetByExternalID(siteID, result.ComponentID)
			if err == nil {
				primaryComponentID = &component.ID
			}
		}

		action := &domain.ExtractedAction{
			ID:                   uuid.New(),
			SiteID:              siteID,
			ActionType:          domain.ActionType(result.ActionType),
			Title:               result.Description, // Set the required Title field
			Description:         result.Description,
			TechnicianNames:     result.TechnicianNames,
			WorkOrderNumber:     result.WorkOrderNumber,
			ActionDate:          &actionDate,
			ActionStatus:        domain.ActionStatus(result.ActionStatus),
			ExtractionConfidence: result.ConfidenceScore,
			ExtractionMetadata:  domain.JSON{"details": result.Details},
			PrimaryComponentID:  primaryComponentID,
			// Don't set Embedding - let GORM use database default (null)
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		actions = append(actions, action)
	}

	fmt.Printf("DEBUG: Returning %d actions from ExtractActions\n", len(actions))
	return actions, nil
}

func (s *llmService) ProcessNaturalLanguageQuery(query string, siteID uuid.UUID) (*QueryResult, error) {
	// This is a placeholder implementation
	// In a real implementation, this would:
	// 1. Analyze the query to determine intent
	// 2. Convert to appropriate database queries or searches
	// 3. Execute the queries and collect results
	// 4. Format results appropriately

	result := &QueryResult{
		Query:       query,
		ResultType:  "actions",
		Count:       0,
		Data:        []interface{}{},
		Summary:     "No results found",
		Results:     domain.JSON{},
		ProcessedAt: time.Now(),
	}

	return result, nil
}

func (s *llmService) SummarizeDocument(content string) (string, error) {
	messages := []Message{
		{Role: "system", Content: "You are a document summarization specialist for solar maintenance reports."},
		{Role: "user", Content: fmt.Sprintf("Please provide a concise summary of this solar field service report:\n\n%s", content)},
	}

	reqBody := OpenAIRequest{
		Model:    s.model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.apiURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if openAIResp.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

func (s *llmService) buildComponentContext(components []*domain.SiteComponent) string {
	var context string
	for _, comp := range components {
		context += fmt.Sprintf("- %s: %s (ID: %s)\n", comp.ComponentType, comp.Name, comp.ExternalID)
	}
	return context
}

// Enhanced methods per PRD requirements

func (s *llmService) AnalyzeQueryIntent(query string, siteID uuid.UUID) (*domain.QueryIntent, error) {
	// Get site components for context
	components, err := s.componentRepo.ListBySite(siteID, &domain.Pagination{Limit: 100}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get site components: %w", err)
	}

	componentContext := s.buildComponentContext(components)

	prompt := fmt.Sprintf(`You are an expert at analyzing natural language queries about solar asset management.

Site Context:
%s

Analyze the following query and return a JSON response with the query intent:

Query: "%s"

Return JSON with the following structure:
{
  "type": "timeline|maintenance_history|component_status|search|analysis",
  "confidence": 0.95,
  "extracted_entities": {
    "components": ["inverter001", "combiner05"],
    "dates": ["2023-12-15", "last month"],
    "maintenance_types": ["repair", "replacement"],
    "technicians": ["John Smith"],
    "work_orders": ["WO-12345"]
  },
  "related_concepts": ["power output", "electrical issues", "warranty"],
  "required_sources": ["field_service_report", "maintenance_log"],
  "date_range": {
    "start": "2023-11-01",
    "end": "2023-12-31"
  },
  "component_filters": ["inverter", "combiner"]
}`, componentContext, query)

	messages := []Message{
		{Role: "system", Content: "You are a solar asset management query analysis specialist. Always return valid JSON."},
		{Role: "user", Content: prompt},
	}

	reqBody := OpenAIRequest{
		Model:    s.model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.apiURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if openAIResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	// Parse the JSON response
	responseContent := openAIResp.Choices[0].Message.Content
	fmt.Printf("LLM Intent Response: %s\n", responseContent)
	
	var intent domain.QueryIntent
	if err := json.Unmarshal([]byte(responseContent), &intent); err != nil {
		// Log the raw response for debugging
		fmt.Printf("Failed to parse intent JSON. Raw response: %s\n", responseContent)
		return nil, fmt.Errorf("failed to parse intent result: %w", err)
	}

	return &intent, nil
}

func (s *llmService) ExtractEntities(text string) (map[string][]string, error) {
	prompt := fmt.Sprintf(`Extract entities from this solar asset management text. Return JSON format:

Text: "%s"

Return:
{
  "components": ["specific component IDs or names"],
  "dates": ["dates mentioned"],
  "maintenance_types": ["repair", "replacement", "inspection"],
  "technicians": ["technician names"],
  "work_orders": ["work order numbers"],
  "locations": ["site locations or areas"],
  "issues": ["problems or issues mentioned"]
}`, text)

	messages := []Message{
		{Role: "system", Content: "You are an entity extraction specialist for solar asset data. Always return valid JSON."},
		{Role: "user", Content: prompt},
	}

	reqBody := OpenAIRequest{
		Model:    s.model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.apiURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if openAIResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	// Parse the JSON response
	var entities map[string][]string
	if err := json.Unmarshal([]byte(openAIResp.Choices[0].Message.Content), &entities); err != nil {
		return nil, fmt.Errorf("failed to parse entities: %w", err)
	}

	return entities, nil
}

func (s *llmService) GenerateEnhancedResponse(query string, sources []domain.QuerySourceDetail) (*domain.EnhancedQueryResponse, error) {
	startTime := time.Now()

	// Build source context for the LLM
	sourceContext := ""
	for i, source := range sources {
		sourceContext += fmt.Sprintf("\nSource %d (%s - %s):\n%s\n", 
			i+1, source.DocumentTitle, source.DocumentType, source.RelevantExcerpt)
	}

	prompt := fmt.Sprintf(`You are a professional solar asset management assistant. Answer the query using ONLY the provided sources. Follow these requirements:

1. Only use information from the provided sources - no external knowledge
2. Include specific citations in your answer
3. If you cannot answer from the sources, say so explicitly
4. Maintain professional tone
5. Be concise but complete

Query: "%s"

Available Sources:
%s

Provide your response in the following JSON format:
{
  "answer": "Your detailed answer here with [Source 1] citations",
  "confidence_score": 0.95,
  "related_concepts": ["related terms mentioned"],
  "response_type": "summary|timeline|list|analysis"
}`, query, sourceContext)

	messages := []Message{
		{Role: "system", Content: "You are a professional solar asset management assistant. Always provide accurate, source-based answers with citations."},
		{Role: "user", Content: prompt},
	}

	reqBody := OpenAIRequest{
		Model:    s.model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.apiURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if openAIResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	// Parse the JSON response
	responseContent := openAIResp.Choices[0].Message.Content
	fmt.Printf("LLM Enhanced Response: %s\n", responseContent)
	
	var responseData struct {
		Answer         string   `json:"answer"`
		ConfidenceScore float64 `json:"confidence_score"`
		RelatedConcepts []string `json:"related_concepts"`
		ResponseType   string   `json:"response_type"`
	}

	if err := json.Unmarshal([]byte(responseContent), &responseData); err != nil {
		// Log the raw response for debugging
		fmt.Printf("Failed to parse enhanced response JSON. Raw response: %s\n", responseContent)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract entities from the query
	entities, _ := s.ExtractEntities(query)

	processingTime := int(time.Since(startTime).Milliseconds())

	// Validate response against sources
	confidence, _ := s.ValidateResponseAgainstSources(responseData.Answer, sources)

	response := &domain.EnhancedQueryResponse{
		Answer:            responseData.Answer,
		ConfidenceScore:   confidence,
		Sources:          sources,
		RelatedConcepts:  responseData.RelatedConcepts,
		ExtractedEntities: entities,
		ResponseType:     responseData.ResponseType,
		NoHallucination:  confidence > 0.7, // Flag if confidence is high
		ProcessingTimeMs: processingTime,
	}

	return response, nil
}

func (s *llmService) ValidateResponseAgainstSources(answer string, sources []domain.QuerySourceDetail) (float64, error) {
	if len(sources) == 0 {
		return 0.0, nil
	}

	// Basic validation - check if answer content appears in sources
	answerWords := strings.Fields(strings.ToLower(answer))
	supportedWords := 0

	// Remove citations and common words for better accuracy
	filteredWords := []string{}
	for _, word := range answerWords {
		// Skip citations like [Source 1]
		if !strings.Contains(word, "source") && !strings.Contains(word, "[") && 
		   !strings.Contains(word, "]") && len(word) > 3 {
			filteredWords = append(filteredWords, word)
		}
	}

	for _, source := range sources {
		sourceText := strings.ToLower(source.RelevantExcerpt)
		for _, word := range filteredWords {
			if strings.Contains(sourceText, word) {
				supportedWords++
			}
		}
	}

	if len(filteredWords) > 0 {
		return float64(supportedWords) / float64(len(filteredWords)), nil
	}

	return 0.5, nil // Default moderate confidence if no content to validate
}