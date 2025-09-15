# Engramiq Reporting Agent Widget - Complete Architecture

## Project Overview

Engramiq is a **Reporting Agent Widget** designed for Solar Asset Managers to:
- Process and query unstructured operational data (emails, PDFs, reports)
- Extract key actions and relate them to specific site components
- Provide intelligent querying of site memory and operational context
- Visualize events on timelines for better operational insights
- Reduce manual data wrangling from 50% of daily work to minimal effort

## Technology Stack Selection

### Backend Stack (Go-based)
**Why Go?** Perfect for high-performance data processing, concurrent document ingestion, and enterprise-grade reliability.

```
Language: Go 1.21+
Framework: Fiber v2 (high-performance HTTP)
Database: PostgreSQL 15+ with pgvector extension
Document Processing: Tika + custom parsers
LLM Integration: OpenAI API with privacy controls
Search: Elasticsearch + pgvector hybrid
Queue: Redis for background jobs
Storage: MinIO (S3-compatible) for documents
Cache: Redis for query results
```

### Frontend Stack (Next.js)
**Why Next.js?** Excellent for complex data visualization, server-side rendering for performance, and enterprise UI requirements.

```
Framework: Next.js 14+ with App Router
Language: TypeScript 5+
Styling: Tailwind CSS (matching style guide)
Components: Custom components following style guide
State: Zustand for complex widget state
Data Fetching: TanStack Query with real-time updates
Charts: D3.js for timeline visualizations
Forms: React Hook Form with Zod validation
```

## Detailed Architecture

### 1. Document Ingestion Pipeline

```go
// Document ingestion service
type DocumentService struct {
    parser    DocumentParser
    extractor ActionExtractor
    llmClient LLMClient
    db        *Database
}

type Document struct {
    ID          string                 `json:"id"`
    Type        DocumentType          `json:"type"` // email, pdf, meeting_transcript
    Source      string                `json:"source"`
    Content     string                `json:"content"`
    Metadata    map[string]interface{} `json:"metadata"`
    ProcessedAt time.Time             `json:"processed_at"`
    SiteID      string                `json:"site_id"`
}

// Process workflow
func (ds *DocumentService) ProcessDocument(doc *Document) error {
    // 1. Parse document content
    parsedContent, err := ds.parser.Parse(doc)
    if err != nil {
        return err
    }
    
    // 2. Extract actions using LLM
    actions, err := ds.extractor.ExtractActions(parsedContent, doc.SiteID)
    if err != nil {
        return err
    }
    
    // 3. Relate to site components
    for _, action := range actions {
        components, err := ds.RelateToComponents(action, doc.SiteID)
        if err != nil {
            continue
        }
        action.RelatedComponents = components
    }
    
    // 4. Store in database
    return ds.db.StoreDocumentAndActions(doc, actions)
}
```

### 2. Site Component Management

Based on `inverter_nodes.json`, we need a flexible schema for equipment:

```sql
-- Site hierarchy table
CREATE TABLE sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    location_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Site components (equipment, inverters, etc.)
CREATE TABLE site_components (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES sites(id),
    external_id VARCHAR(255), -- Maps to "id" in JSON
    component_type VARCHAR(100), -- inverter, combiner, panel, etc.
    name VARCHAR(255),
    label VARCHAR(255),
    level INTEGER,
    group_name VARCHAR(255),
    
    -- Equipment specifications from JSON
    specifications JSONB DEFAULT '{}',
    electrical_data JSONB DEFAULT '{}',
    location_data JSONB DEFAULT '{}',
    
    -- Drawing references
    drawing_title VARCHAR(500),
    drawing_number VARCHAR(100),
    revision VARCHAR(50),
    revision_date DATE,
    spatial_id UUID,
    
    -- Search vectors
    embedding vector(1536),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Component relationships (parent-child hierarchy)
CREATE TABLE component_relationships (
    parent_id UUID REFERENCES site_components(id),
    child_id UUID REFERENCES site_components(id),
    relationship_type VARCHAR(50), -- connects_to, powers, controls
    PRIMARY KEY (parent_id, child_id)
);
```

### 3. Action Extraction and Event Management

```go
// Action extraction using LLM
type ActionExtractor struct {
    llmClient   LLMClient
    componentDB ComponentDatabase
}

type ExtractedAction struct {
    ID                string                 `json:"id"`
    DocumentID        string                 `json:"document_id"`
    ActionType        string                 `json:"action_type"` // maintenance, replacement, troubleshoot
    Description       string                 `json:"description"`
    Timestamp         time.Time              `json:"timestamp"`
    TechnicianName    string                 `json:"technician_name"`
    RelatedComponents []RelatedComponent     `json:"related_components"`
    WorkOrderNumber   string                 `json:"work_order_number"`
    Status           string                 `json:"status"`
    FollowUpActions  []string               `json:"follow_up_actions"`
    Metadata         map[string]interface{} `json:"metadata"`
}

type RelatedComponent struct {
    ComponentID   string  `json:"component_id"`
    ComponentType string  `json:"component_type"`
    Confidence    float64 `json:"confidence"`
    Mention       string  `json:"mention"` // How it was referenced in text
}

// LLM prompt for extraction
const ExtractionPrompt = `
Analyze this solar site maintenance document and extract structured actions.
For each action, identify:
1. What was done (action_type: maintenance, replacement, troubleshoot, inspection)
2. When it happened (timestamp)
3. Who performed it (technician_name)
4. Which equipment was involved (relate to component IDs from site data)
5. Work order numbers
6. Current status and any follow-up actions needed

Site Component Context:
%s

Document Content:
%s

Return JSON format following the ExtractedAction schema.
`
```

### 4. Event Timeline System

```sql
-- Events table for timeline visualization
CREATE TABLE site_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES sites(id),
    action_id UUID REFERENCES extracted_actions(id),
    event_type VARCHAR(100), -- maintenance, fault, replacement, scheduled
    title VARCHAR(500),
    description TEXT,
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    is_future BOOLEAN DEFAULT FALSE,
    priority VARCHAR(20) DEFAULT 'medium', -- low, medium, high, critical
    
    -- Related components
    primary_component_id UUID REFERENCES site_components(id),
    affected_components UUID[], -- Array of component IDs
    
    -- Metadata
    work_order_number VARCHAR(100),
    technician_name VARCHAR(255),
    source_document_id UUID,
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for timeline queries
CREATE INDEX idx_site_events_site_timeline ON site_events(site_id, start_time);
CREATE INDEX idx_site_events_component ON site_events(primary_component_id);
CREATE INDEX idx_site_events_future ON site_events(is_future, start_time);
```

### 5. Intelligent Query System

```go
// Query service for natural language queries
type QueryService struct {
    llmClient      LLMClient
    vectorDB       VectorDatabase
    textSearch     SearchEngine
    componentDB    ComponentDatabase
    eventDB        EventDatabase
}

type QueryRequest struct {
    Query  string `json:"query"`
    SiteID string `json:"site_id"`
    UserID string `json:"user_id"`
}

type QueryResponse struct {
    Answer         string           `json:"answer"`
    Sources        []Source         `json:"sources"`
    RelatedEvents  []Event          `json:"related_events"`
    Confidence     float64          `json:"confidence"`
    QueryType      string           `json:"query_type"`
    ExecutionTime  int              `json:"execution_time_ms"`
}

func (qs *QueryService) ProcessQuery(req QueryRequest) (*QueryResponse, error) {
    // 1. Classify query type
    queryType := qs.classifyQuery(req.Query)
    
    // 2. Vector similarity search
    vectorResults, err := qs.vectorDB.Search(req.Query, req.SiteID, 10)
    if err != nil {
        return nil, err
    }
    
    // 3. Full-text search for exact matches
    textResults, err := qs.textSearch.Search(req.Query, req.SiteID)
    if err != nil {
        return nil, err
    }
    
    // 4. Component-specific search if query mentions equipment
    components, err := qs.findMentionedComponents(req.Query, req.SiteID)
    if err != nil {
        return nil, err
    }
    
    // 5. Generate contextual answer using LLM
    context := qs.buildContext(vectorResults, textResults, components)
    answer, err := qs.generateAnswer(req.Query, context)
    if err != nil {
        return nil, err
    }
    
    return &QueryResponse{
        Answer:        answer.Text,
        Sources:       answer.Sources,
        RelatedEvents: qs.findRelatedEvents(components),
        Confidence:    answer.Confidence,
        QueryType:     queryType,
    }, nil
}

// Query classification
func (qs *QueryService) classifyQuery(query string) string {
    // Use LLM to classify:
    // - maintenance_history: "What work was done on inverter 001?"
    // - fault_analysis: "Why is inverter 31 failing?"
    // - scheduled_events: "What's coming up next week?"
    // - performance_metrics: "How is inverter performance?"
    // - component_status: "What's the status of inverter 40?"
}
```

### 6. Advanced Search with Privacy Controls

```go
// Privacy-preserving LLM integration
type LLMClient struct {
    apiKey          string
    endpoint        string
    privacySettings PrivacySettings
}

type PrivacySettings struct {
    StripPII          bool     `json:"strip_pii"`
    AllowedDataTypes  []string `json:"allowed_data_types"`
    DataRetention     string   `json:"data_retention"`
    LoggingEnabled    bool     `json:"logging_enabled"`
}

func (client *LLMClient) ProcessWithPrivacy(content string, prompt string) (*LLMResponse, error) {
    // 1. Strip PII if enabled
    if client.privacySettings.StripPII {
        content = client.stripPersonalInfo(content)
    }
    
    // 2. Check data types
    if !client.isAllowedDataType(content) {
        return nil, errors.New("data type not allowed for LLM processing")
    }
    
    // 3. Process with temporary context only
    response, err := client.callLLM(prompt, content)
    if err != nil {
        return nil, err
    }
    
    // 4. Clear temporary data immediately
    if !client.privacySettings.LoggingEnabled {
        client.clearTemporaryData()
    }
    
    return response, nil
}
```

### 7. Timeline Visualization API

```go
// Timeline API endpoints
func (h *TimelineHandler) GetTimeline(c *fiber.Ctx) error {
    siteID := c.Params("siteId")
    startDate := c.Query("start", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
    endDate := c.Query("end", time.Now().AddDate(0, 1, 0).Format("2006-01-02"))
    
    events, err := h.eventService.GetEventsByDateRange(siteID, startDate, endDate)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    
    // Group events by type and priority
    timeline := h.buildTimelineResponse(events)
    
    return c.JSON(fiber.Map{
        "success": true,
        "data": fiber.Map{
            "timeline": timeline,
            "summary": h.generateTimelineSummary(events),
        },
    })
}

type TimelineEvent struct {
    ID              string                 `json:"id"`
    Title           string                 `json:"title"`
    Description     string                 `json:"description"`
    StartTime       time.Time              `json:"start_time"`
    EndTime         *time.Time             `json:"end_time,omitempty"`
    EventType       string                 `json:"event_type"`
    Priority        string                 `json:"priority"`
    IsFuture        bool                   `json:"is_future"`
    Component       ComponentSummary       `json:"component"`
    Sources         []DocumentSource       `json:"sources"`
    FollowUpActions []string               `json:"follow_up_actions,omitempty"`
    Metadata        map[string]interface{} `json:"metadata"`
}
```

### 8. Real-time Processing Architecture

```go
// Background job processing
type JobProcessor struct {
    redis    *redis.Client
    services map[string]JobHandler
}

type DocumentProcessingJob struct {
    DocumentID string                 `json:"document_id"`
    SiteID     string                 `json:"site_id"`
    Priority   string                 `json:"priority"` // high, normal, low
    Options    map[string]interface{} `json:"options"`
}

func (jp *JobProcessor) ProcessDocumentAsync(job DocumentProcessingJob) {
    // Queue for background processing
    jobData, _ := json.Marshal(job)
    jp.redis.LPush(context.Background(), "document_processing_queue", jobData)
}

// Worker processes documents
func (jp *JobProcessor) DocumentWorker() {
    for {
        // Pop job from queue
        result := jp.redis.BRPop(context.Background(), 0, "document_processing_queue")
        if len(result.Val()) < 2 {
            continue
        }
        
        var job DocumentProcessingJob
        json.Unmarshal([]byte(result.Val()[1]), &job)
        
        // Process document
        err := jp.services["document"].Handle(job)
        if err != nil {
            // Handle failure - retry logic, error logging
            log.Error("Document processing failed", "job_id", job.DocumentID, "error", err)
        }
    }
}
```

### 9. Frontend Architecture

```typescript
// Timeline visualization component
interface TimelineProps {
  siteId: string;
  dateRange: DateRange;
  eventTypes: string[];
}

export const EventTimeline: React.FC<TimelineProps> = ({ 
  siteId, 
  dateRange, 
  eventTypes 
}) => {
  const { data: events, isLoading } = useQuery({
    queryKey: ['timeline', siteId, dateRange],
    queryFn: () => fetchTimeline(siteId, dateRange),
  });

  return (
    <div className="timeline-container">
      <TimelineHeader dateRange={dateRange} />
      <TimelineChart 
        events={events}
        onEventClick={handleEventClick}
        colorScheme={timelineColorScheme}
      />
      <EventDetails selectedEvent={selectedEvent} />
    </div>
  );
};

// Query interface component
export const QueryInterface: React.FC = () => {
  const [query, setQuery] = useState('');
  const [response, setResponse] = useState<QueryResponse | null>(null);
  
  const handleQuery = useMutation({
    mutationFn: (query: string) => submitQuery(query, currentSiteId),
    onSuccess: (data) => {
      setResponse(data);
      // Update timeline if query affects events
      queryClient.invalidateQueries(['timeline']);
    },
  });

  return (
    <div className="query-interface">
      <SearchInput 
        value={query}
        onChange={setQuery}
        onSubmit={() => handleQuery.mutate(query)}
        placeholder="Ask about site operations..."
      />
      {response && (
        <QueryResults 
          response={response}
          onSourceClick={handleSourceClick}
        />
      )}
    </div>
  );
};
```

### 10. Deployment Architecture

```yaml
# Docker Compose for development
version: '3.8'
services:
  engramiq-api:
    build: ./backend
    environment:
      - DATABASE_URL=postgresql://user:pass@postgres:5432/engramiq
      - REDIS_URL=redis://redis:6379
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    depends_on:
      - postgres
      - redis
      - elasticsearch

  engramiq-web:
    build: ./frontend
    environment:
      - NEXT_PUBLIC_API_URL=http://localhost:8080
    ports:
      - "3000:3000"

  postgres:
    image: pgvector/pgvector:pg15
    environment:
      POSTGRES_DB: engramiq
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

  elasticsearch:
    image: elasticsearch:8.8.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
    volumes:
      - es_data:/usr/share/elasticsearch/data

  minio:
    image: minio/minio
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: admin
      MINIO_ROOT_PASSWORD: password123
    volumes:
      - minio_data:/data

volumes:
  postgres_data:
  redis_data:
  es_data:
  minio_data:
```

## Key Features Implementation

### 1. Document Upload & Processing
- Drag-and-drop interface for PDFs, emails, transcripts
- Automatic parsing and content extraction
- Real-time processing status updates

### 2. Natural Language Querying
- "What maintenance was done on inverter 31 last month?"
- "Show me all arc fault incidents"
- "What's scheduled for next week?"

### 3. Timeline Visualization
- Interactive timeline with zoom/pan
- Color-coded by event type and priority
- Click for detailed information and sources

### 4. Component Integration
- Automatic linking of actions to equipment
- Equipment hierarchy visualization
- Status tracking and history

This architecture specifically addresses the Solar Asset Manager use case while maintaining enterprise-grade security, performance, and scalability requirements outlined in the PRD.