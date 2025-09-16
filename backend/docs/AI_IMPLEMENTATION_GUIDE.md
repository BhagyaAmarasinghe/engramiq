# AI Services and Vector Search Implementation Guide

## Overview

EngramIQ implements a sophisticated AI-powered document processing and semantic search system using embeddings, vector storage, and natural language processing. This guide explains the technical implementation details and architecture of these AI services.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Embeddings and Vectors Explained](#embeddings-and-vectors-explained)
3. [Document Processing Pipeline](#document-processing-pipeline)
4. [Vector Storage with pgvector](#vector-storage-with-pgvector)
5. [Semantic Search Implementation](#semantic-search-implementation)
6. [RAG (Retrieval Augmented Generation)](#rag-retrieval-augmented-generation)
7. [AI Service Components](#ai-service-components)
8. [Performance Optimization](#performance-optimization)
9. [API Examples](#api-examples)

## Architecture Overview

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  File Upload    │────▶│ Text Extraction  │────▶│ Async Processing│
└─────────────────┘     └──────────────────┘     └─────────────────┘
                                                           │
                                                           ▼
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│ Vector Storage  │◀────│ OpenAI Embedding │◀────│ LLM Service     │
│  (pgvector)     │     │     API          │     │                 │
└─────────────────┘     └──────────────────┘     └─────────────────┘
         │
         ▼
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│ Semantic Search │────▶│ Query Processing │────▶│ Response with   │
│                 │     │      (RAG)       │     │ Source Citation │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

## Embeddings and Vectors Explained

### What are Embeddings?

Embeddings are numerical representations of text that capture semantic meaning. They transform human-readable text into machine-processable vectors in high-dimensional space.

**Key Concepts:**
- **Dimension**: EngramIQ uses 1536-dimensional vectors (OpenAI's text-embedding-ada-002 model)
- **Semantic Distance**: Similar content has vectors closer together in vector space
- **Language Understanding**: Embeddings capture context, meaning, and relationships

### Mathematical Foundation

```
Text: "Solar panel maintenance report"
         ↓
Embedding: [0.021, -0.043, 0.112, ..., -0.008] // 1536 dimensions
```

**Cosine Similarity Formula:**
```
similarity = cos(θ) = (A·B) / (||A|| × ||B||)

Where:
- A·B is the dot product of vectors A and B
- ||A|| and ||B|| are the magnitudes of the vectors
- Result ranges from -1 (opposite) to 1 (identical)
```

## Document Processing Pipeline

### 1. File Upload Handler
**Location**: `internal/handler/document_handler.go`

```go
func (h *DocumentHandler) UploadDocument(c *fiber.Ctx) error {
    // Parse multipart form
    file, err := c.FormFile("file")
    
    // Create document record
    document := &domain.Document{
        SiteID:      siteID,
        Title:       file.Filename,
        DocumentType: c.FormValue("document_type"),
        Status:      "pending",
    }
    
    // Trigger async processing
    go h.documentService.ProcessDocument(document.ID)
}
```

### 2. Text Extraction
**Location**: `internal/service/document_service.go`

```go
func (s *DocumentService) extractContent(filePath string) (string, error) {
    switch extension {
    case ".pdf":
        return extractPDFContent(filePath)
    case ".docx":
        return extractDOCXContent(filePath)
    case ".txt":
        return extractTextContent(filePath)
    }
}
```

### 3. Embedding Generation
**Location**: `internal/service/llm_service.go`

```go
func (s *LLMService) GenerateEmbedding(text string) (pgvector.Vector, error) {
    // Prepare OpenAI API request
    requestBody := map[string]interface{}{
        "model": "text-embedding-ada-002",
        "input": text,
    }
    
    // Call OpenAI API
    resp, err := s.httpClient.Post(
        "https://api.openai.com/v1/embeddings",
        "application/json",
        bytes.NewBuffer(jsonBody),
    )
    
    // Convert response to pgvector format
    embedding := make([]float32, len(floatEmbedding))
    for i, v := range floatEmbedding {
        embedding[i] = float32(v)
    }
    
    return pgvector.NewVector(embedding), nil
}
```

### 4. Document Processing
**Location**: `internal/service/document_service.go`

```go
func (s *DocumentService) ProcessDocument(documentID string) error {
    // Update status to processing
    doc.ProcessingStatus = "processing"
    
    // Generate embedding
    embedding, err := s.llmService.GenerateEmbedding(doc.ProcessedContent)
    
    // Extract actions (maintenance, repairs, etc.)
    actions, err := s.llmService.ExtractActions(doc.ProcessedContent)
    
    // Update document with results
    doc.Embedding = embedding
    doc.ProcessingStatus = "completed"
    
    return s.repo.Update(doc)
}
```

## Vector Storage with pgvector

### Database Schema

```sql
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Documents table with vector column
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    site_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    processed_content TEXT,
    embedding vector(1536),  -- 1536-dimensional vector
    processing_status VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create IVFFlat index for efficient similarity search
CREATE INDEX idx_documents_embedding ON documents 
USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100);
```

### GORM Model Implementation
**Location**: `internal/domain/document.go`

```go
type Document struct {
    ID               string           `gorm:"type:uuid;default:uuid_generate_v4()"`
    SiteID          string           `gorm:"type:uuid;not null"`
    Title           string           `gorm:"type:varchar(255);not null"`
    Content         string           `gorm:"type:text"`
    ProcessedContent string           `gorm:"type:text"`
    Embedding       pgvector.Vector  `gorm:"type:vector(1536)"`
    ProcessingStatus string          `gorm:"type:varchar(50)"`
    CreatedAt       time.Time        `gorm:"not null"`
}
```

## Semantic Search Implementation

### Search Repository
**Location**: `internal/repository/document_repository.go`

```go
func (r *DocumentRepository) SearchSemantic(
    siteID string, 
    queryVector pgvector.Vector, 
    threshold float64,
) ([]domain.Document, error) {
    var documents []domain.Document
    
    // Cosine similarity search using pgvector
    query := r.db.Model(&domain.Document{}).
        Where("site_id = ?", siteID).
        Where("embedding <=> ? < ?", queryVector, threshold).
        Order("embedding <=> ?", queryVector).
        Limit(10)
    
    return documents, query.Find(&documents).Error
}
```

### Query Service Implementation
**Location**: `internal/service/query_service.go`

```go
func (s *QueryService) ProcessEnhancedQuery(query string, siteID string) (*domain.EnhancedQueryResponse, error) {
    // 1. Generate query embedding
    queryEmbedding, err := s.llmService.GenerateEmbedding(query)
    
    // 2. Perform semantic search
    relevantDocs, err := s.documentRepo.SearchSemantic(
        siteID, 
        queryEmbedding, 
        0.5, // Similarity threshold
    )
    
    // 3. Fall back to full-text search if needed
    if len(relevantDocs) == 0 {
        relevantDocs, err = s.documentRepo.SearchFullText(siteID, query)
    }
    
    // 4. Generate response using RAG
    response, err := s.llmService.GenerateEnhancedResponse(
        query,
        relevantDocs,
        intentAnalysis,
    )
    
    return response, nil
}
```

## RAG (Retrieval Augmented Generation)

### Implementation Pattern

RAG combines retrieval (finding relevant documents) with generation (creating responses) to provide accurate, source-based answers.

**Location**: `internal/service/llm_service.go`

```go
func (s *LLMService) GenerateEnhancedResponse(
    query string,
    sources []domain.Document,
    intent *IntentAnalysis,
) (*domain.EnhancedQueryResponse, error) {
    // Build context from retrieved documents
    context := s.buildContextFromSources(sources)
    
    // Create constrained prompt
    messages := []Message{
        {
            Role: "system",
            Content: `You are a solar asset management assistant. 
                     Generate responses ONLY from the provided sources.
                     Include citations for all claims.
                     If information is not in sources, say so.`,
        },
        {
            Role: "user",
            Content: fmt.Sprintf(
                "Query: %s\n\nSources:\n%s\n\nProvide a response with citations.",
                query,
                context,
            ),
        },
    }
    
    // Generate response
    response := s.callOpenAI(messages)
    
    // Validate against sources
    validated := s.ValidateResponseAgainstSources(response, sources)
    
    return &domain.EnhancedQueryResponse{
        Answer:           validated.Answer,
        ConfidenceScore: validated.Confidence,
        Sources:         s.formatSources(sources),
        NoHallucination: validated.NoHallucination,
    }, nil
}
```

## AI Service Components

### 1. LLM Service
**Purpose**: Interface with OpenAI API for embeddings and text generation

**Key Methods**:
- `GenerateEmbedding()`: Convert text to vectors
- `ExtractActions()`: Extract maintenance actions from documents
- `AnalyzeQueryIntent()`: Understand user query intent
- `GenerateEnhancedResponse()`: Create RAG-based responses

### 2. Query Service
**Purpose**: Orchestrate query processing pipeline

**Key Methods**:
- `ProcessEnhancedQuery()`: Main query processing endpoint
- `FindSimilarQueries()`: Retrieve similar past queries
- `UpdateQueryWithResponse()`: Store query results

### 3. Content Filter Service
**Purpose**: Ensure professional, appropriate responses

**Key Methods**:
- `ValidateQuery()`: Check input appropriateness
- `EnforceProfessionalTone()`: Adjust response tone
- `FilterInappropriateContent()`: Remove sensitive data

### 4. Source Attribution Service
**Purpose**: Track and cite information sources

**Key Methods**:
- `AttributeSources()`: Link responses to source documents
- `FormatCitation()`: Create proper citations
- `ValidateSourceContent()`: Verify source relevance

## Performance Optimization

### 1. Vector Index Configuration

```sql
-- IVFFlat index with optimized parameters
CREATE INDEX idx_documents_embedding ON documents 
USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100);  -- Adjust based on dataset size

-- Recommended lists calculation:
-- lists = sqrt(number_of_rows) / 2
```

### 2. Batch Processing

```go
// Process multiple embeddings in parallel
func (s *DocumentService) ProcessDocumentsBatch(documentIDs []string) {
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 5) // Limit concurrent processing
    
    for _, id := range documentIDs {
        wg.Add(1)
        semaphore <- struct{}{}
        
        go func(docID string) {
            defer wg.Done()
            defer func() { <-semaphore }()
            
            s.ProcessDocument(docID)
        }(id)
    }
    
    wg.Wait()
}
```

### 3. Caching Strategy

```go
// Cache embeddings in Redis
func (s *LLMService) GetEmbeddingWithCache(text string) (pgvector.Vector, error) {
    // Generate cache key
    cacheKey := fmt.Sprintf("embedding:%s", generateHash(text))
    
    // Check cache
    if cached, err := s.redis.Get(cacheKey); err == nil {
        return deserializeVector(cached), nil
    }
    
    // Generate and cache
    embedding, err := s.GenerateEmbedding(text)
    s.redis.Set(cacheKey, serializeVector(embedding), 24*time.Hour)
    
    return embedding, err
}
```

## API Examples

### 1. Upload Document with Automatic Processing

```bash
POST /api/v1/sites/{siteId}/documents
Content-Type: multipart/form-data

file: maintenance_report.pdf
document_type: field_service_report
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "maintenance_report.pdf",
  "processing_status": "pending",
  "created_at": "2024-09-16T10:30:00Z"
}
```

### 2. Semantic Query with Source Attribution

```bash
POST /api/v1/sites/{siteId}/queries
Content-Type: application/json

{
  "query_text": "What maintenance was performed on inverter INV001?",
  "enhanced": true
}
```

**Response:**
```json
{
  "answer": "Based on the field service reports, inverter INV001 underwent the following maintenance:\n\n1. Quarterly preventive maintenance on March 15th [1]\n2. DC disconnect replacement on March 22nd [2]\n3. Firmware update to version 3.2.1 on March 22nd [2]",
  "confidence_score": 0.92,
  "sources": [
    {
      "document_id": "doc-001",
      "document_title": "Field Service Report - March 15, 2024",
      "relevant_excerpt": "Performed quarterly preventive maintenance on inverter INV001...",
      "citation": "[1] Field Service Report - March 15, 2024, p. 3"
    },
    {
      "document_id": "doc-002",
      "document_title": "Field Service Report - March 22, 2024",
      "relevant_excerpt": "Replaced faulty DC disconnect on INV001 and updated firmware...",
      "citation": "[2] Field Service Report - March 22, 2024, p. 2"
    }
  ],
  "no_hallucination": true,
  "processing_time_ms": 1250
}
```

### 3. Semantic Document Search

```bash
GET /api/v1/sites/{siteId}/documents/search?query=inverter+failure&semantic=true
```

**Response:**
```json
{
  "documents": [
    {
      "id": "doc-003",
      "title": "Emergency Service Report - INV003 Failure",
      "similarity_score": 0.89,
      "excerpt": "Complete inverter failure due to overheating..."
    },
    {
      "id": "doc-007",
      "title": "Maintenance Report - Inverter Issues",
      "similarity_score": 0.76,
      "excerpt": "Multiple inverters showing signs of degradation..."
    }
  ],
  "total": 2
}
```

## Best Practices

### 1. Document Processing
- Process documents asynchronously to avoid blocking uploads
- Implement retry logic for failed OpenAI API calls
- Store both original and processed content
- Use content hashing for deduplication

### 2. Embedding Management
- Cache frequently accessed embeddings
- Batch embedding generation when possible
- Monitor embedding dimension consistency
- Implement fallback for API failures

### 3. Search Optimization
- Use appropriate similarity thresholds (typically 0.3-0.7)
- Combine semantic and full-text search
- Limit result sets to improve performance
- Pre-filter by site/date when possible

### 4. Security Considerations
- Never expose raw API keys in responses
- Implement rate limiting for embedding generation
- Validate and sanitize all input text
- Use environment variables for sensitive configuration

## Troubleshooting

### Common Issues

1. **Embedding Generation Fails**
   - Check OpenAI API key validity
   - Verify network connectivity
   - Check rate limits
   - Ensure text length < 8191 tokens

2. **Poor Search Results**
   - Adjust similarity threshold
   - Verify embedding dimensions match
   - Check index creation
   - Ensure documents are fully processed

3. **Slow Query Performance**
   - Optimize pgvector index parameters
   - Implement result caching
   - Limit concurrent processing
   - Monitor database query plans

## Conclusion

EngramIQ's AI implementation provides a robust foundation for semantic document processing and intelligent query handling in solar asset management. The combination of embeddings, vector search, and RAG ensures accurate, source-based responses while maintaining high performance and reliability.

For additional technical details, refer to the source code in the `internal/service/` and `internal/repository/` directories.