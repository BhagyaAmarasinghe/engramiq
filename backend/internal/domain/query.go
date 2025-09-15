package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// QueryType represents different types of natural language queries
// This helps us optimize query processing based on intent
type QueryType string

const (
	QueryTypeComponentStatus    QueryType = "component_status"
	QueryTypeMaintenanceHistory QueryType = "maintenance_history"
	QueryTypeFaultAnalysis      QueryType = "fault_analysis"
	QueryTypeScheduledEvents    QueryType = "scheduled_events"
	QueryTypePerformanceMetrics QueryType = "performance_metrics"
	QueryTypeGeneral           QueryType = "general"
)

// QueryRequest represents a natural language query from the user
type QueryRequest struct {
	Query               string `json:"query" validate:"required,min=3,max=500"`
	IncludeRelatedEvents bool   `json:"include_related_events"`
}

// QueryResponse contains the AI-generated answer with sources
type QueryResponse struct {
	Answer         string                `json:"answer"`
	Sources        []QuerySourceDetail   `json:"sources"`
	RelatedEvents  []TimelineEvent       `json:"related_events,omitempty"`
	Confidence     float64               `json:"confidence"`
	QueryType      QueryType             `json:"query_type"`
	ExecutionTime  int                   `json:"execution_time_ms"`
}


// QuerySuggestion for autocomplete/suggestions
type QuerySuggestion struct {
	Query       string    `json:"query"`
	QueryType   QueryType `json:"query_type"`
	Popularity  int       `json:"popularity"`
}

// QueryAnalytics tracks usage patterns for optimization
type QueryAnalytics struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SiteID            uuid.UUID  `json:"site_id" gorm:"type:uuid;not null"`
	UserID            *uuid.UUID `json:"user_id" gorm:"type:uuid"`
	QueryText         string     `json:"query_text" gorm:"not null"`
	QueryType         QueryType  `json:"query_type" gorm:"type:varchar(50)"`
	ResultsCount      int        `json:"results_count"`
	ResponseGenerated bool       `json:"response_generated" gorm:"default:false"`
	ExecutionTimeMs   int        `json:"execution_time_ms"`
	SearchTimeMs      int        `json:"search_time_ms"`
	LLMTimeMs         int        `json:"llm_time_ms"`
	SessionID         uuid.UUID  `json:"session_id" gorm:"type:uuid"`
	UserAgent         string     `json:"user_agent"`
	CreatedAt         time.Time  `json:"created_at"`
	
	// Aggregated analytics fields (computed, not stored)
	TotalQueries        int64             `json:"total_queries" gorm:"-"`
	SuccessfulQueries   int64             `json:"successful_queries" gorm:"-"`
	SuccessRate         float64           `json:"success_rate" gorm:"-"`
	AverageResponseTime float64           `json:"average_response_time" gorm:"-"`
	QueryTypeBreakdown  map[string]int64  `json:"query_type_breakdown" gorm:"-"`
}

func (QueryAnalytics) TableName() string {
	return "query_analytics"
}

// UserQuery represents a user's natural language query
type UserQuery struct {
	ID               uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID           uuid.UUID       `json:"user_id" gorm:"type:uuid;not null"`
	SiteID           uuid.UUID       `json:"site_id" gorm:"type:uuid;not null"`
	QueryText        string          `json:"query_text" gorm:"type:text;not null"`
	QueryType        QueryType       `json:"query_type" gorm:"type:varchar(50);not null"`
	Embedding        pgvector.Vector `json:"-" gorm:"type:vector(1536)"`
	Results          JSON            `json:"results" gorm:"type:jsonb"`
	ResultCount      int             `json:"result_count" gorm:"default:0"`
	ConfidenceScore  float64         `json:"confidence_score" gorm:"default:0"`
	ExtractedEntities JSON           `json:"extracted_entities" gorm:"type:jsonb;default:'{}'"`
	ProcessedAt      *time.Time      `json:"processed_at"`
	ErrorMessage     string          `json:"error_message,omitempty" gorm:"type:text"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	
	// Relationships
	Site         *Site          `json:"site,omitempty" gorm:"foreignKey:SiteID"`
	QuerySources []QuerySource  `json:"query_sources,omitempty" gorm:"foreignKey:QueryID"`
}

// QuerySource represents a document source used to answer a query
type QuerySource struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	QueryID          uuid.UUID  `json:"query_id" gorm:"type:uuid;not null"`
	DocumentID       uuid.UUID  `json:"document_id" gorm:"type:uuid;not null"`
	DocumentTitle    string     `json:"document_title" gorm:"type:varchar(500)"`
	RelevantExcerpt  string     `json:"relevant_excerpt" gorm:"type:text"`
	RelevanceScore   float64    `json:"relevance_score" gorm:"default:0"`
	PageNumber       *int       `json:"page_number"`
	SectionReference string     `json:"section_reference" gorm:"type:varchar(255)"`
	CreatedAt        time.Time  `json:"created_at"`
	
	// Relationships
	Query    *UserQuery `json:"query,omitempty" gorm:"foreignKey:QueryID"`
	Document *Document  `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
}

func (QuerySource) TableName() string {
	return "query_sources"
}

// EnhancedQueryResponse represents a structured response with sources per PRD requirements
type EnhancedQueryResponse struct {
	Answer           string               `json:"answer"`
	ConfidenceScore  float64              `json:"confidence_score"`
	Sources          []QuerySourceDetail  `json:"sources"`
	RelatedConcepts  []string             `json:"related_concepts"`
	ExtractedEntities map[string][]string `json:"extracted_entities"`
	ResponseType     string               `json:"response_type"` // summary, timeline, list, analysis
	NoHallucination  bool                 `json:"no_hallucination"` // Validation flag
	ProcessingTimeMs int                  `json:"processing_time_ms"`
}

// QuerySourceDetail provides detailed source information for responses
type QuerySourceDetail struct {
	DocumentID       uuid.UUID `json:"document_id"`
	DocumentTitle    string    `json:"document_title"`
	DocumentDate     time.Time `json:"document_date"`
	DocumentType     string    `json:"document_type"`
	RelevantExcerpt  string    `json:"relevant_excerpt"`
	RelevanceScore   float64   `json:"relevance_score"`
	PageNumber       *int      `json:"page_number,omitempty"`
	SectionReference string    `json:"section_reference,omitempty"`
	Citation         string    `json:"citation"` // Formatted citation string
}

// QueryIntent represents enhanced intent analysis
type QueryIntent struct {
	Type             string                 `json:"type"`        // timeline, search, maintenance_history, component_status, analysis
	Confidence       float64                `json:"confidence"`
	ExtractedEntities map[string][]string   `json:"extracted_entities"`
	RelatedConcepts  []string               `json:"related_concepts"`
	RequiredSources  []string               `json:"required_sources"` // Document types needed
	DateRange        *DateRange             `json:"date_range,omitempty"`
	ComponentFilters []string               `json:"component_filters,omitempty"`
}

// SearchRequest for different search types
type SearchRequest struct {
	Query     string              `json:"query" validate:"required"`
	Filters   map[string]string   `json:"filters"`
	DateRange *DateRange          `json:"date_range,omitempty"`
	Limit     int                 `json:"limit" validate:"min=1,max=100"`
}

// SemanticSearchRequest for AI-powered search
type SemanticSearchRequest struct {
	Query      string      `json:"query" validate:"required,min=3"`
	Limit      int         `json:"limit" validate:"min=1,max=50"`
	Threshold  float64     `json:"threshold" validate:"min=0,max=1"`
}

// SearchResult represents a search hit
type SearchResult struct {
	ID             uuid.UUID     `json:"id"`
	Type           string        `json:"type"` // document, component, action
	Title          string        `json:"title"`
	Excerpt        string        `json:"excerpt"`
	Score          float64       `json:"score"`
	SimilarityScore float64      `json:"similarity_score,omitempty"`
	Context        string        `json:"context,omitempty"`
	RecentIssues   []string      `json:"recent_issues,omitempty"`
	Highlights     []string      `json:"highlights,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
}