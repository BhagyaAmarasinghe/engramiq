# EngramIQ Widget Data Schema and Architecture

## Executive Summary

The EngramIQ Widget is a comprehensive solar asset management system designed to transform unstructured operational data into actionable insights. This document outlines the data schema and architecture that powers the widget's ability to process documents, extract maintenance actions, and provide intelligent query capabilities.

## Table of Contents

1. [System Overview](#system-overview)
2. [Data Architecture Principles](#data-architecture-principles)
3. [Core Data Schema](#core-data-schema)
4. [Widget-Specific Data Requirements](#widget-specific-data-requirements)
5. [Data Flow Architecture](#data-flow-architecture)
6. [API Data Contracts](#api-data-contracts)
7. [State Management Strategy](#state-management-strategy)
8. [Performance Considerations](#performance-considerations)
9. [Security Architecture](#security-architecture)

## System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         EngramIQ Widget Architecture                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────┐    ┌──────────────┐    ┌───────────────┐        │
│  │   Widget    │───▶│   REST API   │───▶│  PostgreSQL   │        │
│  │  Frontend   │    │   Backend    │    │   Database    │        │
│  └─────────────┘    └──────────────┘    └───────────────┘        │
│         │                   │                     │                 │
│         │                   │                     │                 │
│         ▼                   ▼                     ▼                 │
│  ┌─────────────┐    ┌──────────────┐    ┌───────────────┐        │
│  │    State    │    │   OpenAI     │    │   pgvector    │        │
│  │ Management  │    │   Services   │    │  Embeddings   │        │
│  └─────────────┘    └──────────────┘    └───────────────┘        │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## Data Architecture Principles

### 1. Domain-Driven Design
- Clear bounded contexts for solar asset management
- Rich domain models representing real-world entities
- Ubiquitous language aligned with solar industry terminology

### 2. Event Sourcing for Timeline
- Immutable event records for complete audit trail
- Temporal data organization for historical analysis
- Support for timeline visualization and filtering

### 3. CQRS Pattern
- Separated read and write models for optimization
- Specialized query models for widget views
- Optimized projections for performance

### 4. Semantic Layer
- Vector embeddings for intelligent search
- Natural language understanding
- Context-aware query processing

## Core Data Schema

### 1. Site Management

```typescript
interface Site {
  id: UUID;
  siteCode: string;           // e.g., "S2367"
  name: string;
  address: string;
  coordinates: {
    latitude: number;
    longitude: number;
  };
  capacity: {
    totalKW: number;
    inverterCount: number;
    panelCount: number;
  };
  metadata: {
    installationDate: Date;
    warrantyExpiry: Date;
    maintenanceSchedule: string;
    [key: string]: any;
  };
  status: 'operational' | 'partial' | 'offline';
  lastUpdated: Date;
}
```

### 2. Component Hierarchy

```typescript
interface Component {
  id: UUID;
  siteId: UUID;
  type: ComponentType;
  identifier: string;         // e.g., "INV001"
  manufacturer: string;
  model: string;
  serialNumber: string;
  status: ComponentStatus;
  specifications: {
    capacity: number;
    voltage: number;
    current: number;
    [key: string]: any;
  };
  location: {
    area: string;
    position: string;
    coordinates?: [number, number];
  };
  relationships: ComponentRelationship[];
  embedding?: number[];       // 1536-dimensional vector
}

type ComponentType = 
  | 'inverter' 
  | 'combiner' 
  | 'panel' 
  | 'transformer' 
  | 'meter' 
  | 'switchgear' 
  | 'monitoring';

type ComponentStatus = 
  | 'operational' 
  | 'fault' 
  | 'maintenance' 
  | 'offline';
```

### 3. Document Processing

```typescript
interface Document {
  id: UUID;
  siteId: UUID;
  title: string;
  type: DocumentType;
  uploadDate: Date;
  content: {
    raw: string;
    processed: string;
    summary?: string;
  };
  metadata: {
    fileHash: string;
    fileSize: number;
    mimeType: string;
    pageCount?: number;
    author?: string;
    createdDate?: Date;
  };
  processing: {
    status: ProcessingStatus;
    embedding?: number[];     // Vector representation
    extractedEntities?: Entity[];
    confidence?: number;
  };
  searchVectors: {
    title: string;            // tsvector for full-text search
    content: string;          // tsvector for full-text search
  };
}

type DocumentType = 
  | 'field_service_report'
  | 'email'
  | 'work_order'
  | 'manual'
  | 'warranty_claim'
  | 'invoice';
```

### 4. Extracted Actions

```typescript
interface ExtractedAction {
  id: UUID;
  documentId: UUID;
  siteId: UUID;
  componentId?: UUID;
  type: ActionType;
  description: string;
  performedDate: Date;
  technicians: string[];
  workOrderNumber?: string;
  measurements?: {
    [parameter: string]: {
      value: number;
      unit: string;
      timestamp: Date;
    };
  };
  faultCodes?: string[];
  recommendations?: string;
  confidenceScore: number;
  status: ActionStatus;
}

type ActionType = 
  | 'maintenance'
  | 'replacement'
  | 'troubleshooting'
  | 'inspection'
  | 'repair'
  | 'calibration'
  | 'testing';
```

### 5. Timeline Events

```typescript
interface TimelineEvent {
  id: UUID;
  siteId: UUID;
  componentId?: UUID;
  documentId?: UUID;
  actionId?: UUID;
  eventType: EventType;
  title: string;
  description: string;
  eventDate: Date;
  priority: EventPriority;
  tags: string[];
  relatedEvents: UUID[];
  metadata: Record<string, any>;
}

type EventType = 
  | 'maintenance_scheduled'
  | 'maintenance_completed'
  | 'fault_detected'
  | 'fault_resolved'
  | 'inspection_completed'
  | 'component_replaced'
  | 'warranty_claim';
```

### 6. Query System

```typescript
interface UserQuery {
  id: UUID;
  userId: UUID;
  siteId: UUID;
  queryText: string;
  queryEmbedding?: number[];
  intent: QueryIntent;
  entities: ExtractedEntity[];
  timestamp: Date;
}

interface QueryResponse {
  queryId: UUID;
  answer: string;
  confidence: number;
  sources: SourceAttribution[];
  processingTime: number;
  noHallucination: boolean;
}

interface SourceAttribution {
  documentId: UUID;
  documentTitle: string;
  relevantExcerpt: string;
  pageNumber?: number;
  confidence: number;
}
```

## Widget-Specific Data Requirements

### 1. Dashboard View Data

```typescript
interface DashboardData {
  site: {
    overview: SiteOverview;
    statusSummary: StatusSummary;
    recentEvents: TimelineEvent[];
    alertCount: number;
  };
  components: {
    totalCount: number;
    byType: Record<ComponentType, number>;
    byStatus: Record<ComponentStatus, number>;
    criticalComponents: Component[];
  };
  documents: {
    totalCount: number;
    recentUploads: Document[];
    processingQueue: number;
  };
  maintenance: {
    upcomingTasks: ExtractedAction[];
    overdueTasks: ExtractedAction[];
    completionRate: number;
  };
}
```

### 2. Component Detail View

```typescript
interface ComponentDetailView {
  component: Component;
  maintenanceHistory: ExtractedAction[];
  relatedDocuments: Document[];
  timelineEvents: TimelineEvent[];
  performanceMetrics: {
    uptime: number;
    efficiency: number;
    faultRate: number;
  };
  relationships: {
    parent?: Component;
    children: Component[];
    connections: Component[];
  };
}
```

### 3. Timeline View

```typescript
interface TimelineView {
  events: TimelineEvent[];
  filters: {
    dateRange: [Date, Date];
    eventTypes: EventType[];
    components: UUID[];
    priority: EventPriority[];
  };
  grouping: 'day' | 'week' | 'month';
  stats: {
    totalEvents: number;
    eventsByType: Record<EventType, number>;
    eventsByComponent: Record<string, number>;
  };
}
```

### 4. Search Results View

```typescript
interface SearchResultsView {
  query: string;
  results: {
    documents: DocumentSearchResult[];
    components: ComponentSearchResult[];
    actions: ActionSearchResult[];
  };
  facets: {
    documentTypes: FacetCount[];
    componentTypes: FacetCount[];
    dateRanges: FacetCount[];
  };
  totalResults: number;
  searchTime: number;
}
```

## Data Flow Architecture

### 1. Document Upload Flow

```
User Upload → API Validation → File Storage → Document Record Creation
                                    ↓
                            Text Extraction
                                    ↓
                            Embedding Generation
                                    ↓
                            Action Extraction → Component Linking
                                    ↓
                            Event Generation → Timeline Update
```

### 2. Query Processing Flow

```
User Query → Intent Analysis → Entity Extraction
                    ↓
            Embedding Generation
                    ↓
            Semantic Search ←→ Full-Text Search
                    ↓
            Source Retrieval
                    ↓
            Response Generation → Source Attribution
                    ↓
            Confidence Scoring → Final Response
```

### 3. Real-Time Update Flow

```
Database Change → Change Event → Event Bus
                        ↓
                WebSocket Server
                        ↓
                Widget Client → State Update → UI Refresh
```

## API Data Contracts

### 1. Document Upload

```typescript
// Request
POST /api/v1/sites/{siteId}/documents
Content-Type: multipart/form-data

{
  file: File;
  documentType: DocumentType;
  metadata?: {
    author?: string;
    date?: string;
    tags?: string[];
  };
}

// Response
{
  document: {
    id: string;
    title: string;
    processingStatus: 'pending';
    uploadedAt: string;
  };
}
```

### 2. Component Query

```typescript
// Request
GET /api/v1/sites/{siteId}/components?
  type={componentType}&
  status={status}&
  page={page}&
  limit={limit}

// Response
{
  components: Component[];
  pagination: {
    total: number;
    page: number;
    limit: number;
    hasMore: boolean;
  };
  aggregations: {
    byType: Record<ComponentType, number>;
    byStatus: Record<ComponentStatus, number>;
  };
}
```

### 3. Natural Language Query

```typescript
// Request
POST /api/v1/sites/{siteId}/queries
{
  queryText: string;
  context?: {
    componentId?: string;
    dateRange?: [string, string];
    previousQueries?: string[];
  };
}

// Response
{
  response: {
    answer: string;
    confidence: number;
    sources: SourceAttribution[];
    suggestedActions?: string[];
    relatedQueries?: string[];
  };
  metadata: {
    processingTime: number;
    queryId: string;
  };
}
```

## State Management Strategy

### 1. Client-Side State Structure

```typescript
interface WidgetState {
  // Core Data
  site: Site | null;
  components: {
    items: Record<UUID, Component>;
    loading: boolean;
    error: Error | null;
  };
  documents: {
    items: Record<UUID, Document>;
    uploadQueue: UploadTask[];
    processingStatus: Record<UUID, ProcessingStatus>;
  };
  
  // UI State
  ui: {
    activeView: ViewType;
    selectedComponent: UUID | null;
    filters: FilterState;
    sort: SortState;
  };
  
  // Query State
  queries: {
    history: UserQuery[];
    activeQuery: QueryState | null;
    suggestions: string[];
  };
  
  // Timeline State
  timeline: {
    events: TimelineEvent[];
    dateRange: [Date, Date];
    grouping: TimeGrouping;
    filters: TimelineFilters;
  };
}
```

### 2. State Update Patterns

```typescript
// Optimistic Updates
const updateComponentStatus = (componentId: UUID, status: ComponentStatus) => {
  // Immediately update UI
  dispatch({ type: 'UPDATE_COMPONENT_STATUS', payload: { componentId, status } });
  
  // API call with rollback on failure
  api.updateComponent(componentId, { status })
    .catch(error => {
      dispatch({ type: 'ROLLBACK_COMPONENT_STATUS', payload: { componentId } });
    });
};

// Incremental Loading
const loadMoreDocuments = async (page: number) => {
  dispatch({ type: 'FETCH_DOCUMENTS_START' });
  
  try {
    const { documents, hasMore } = await api.getDocuments({ page });
    dispatch({ 
      type: 'FETCH_DOCUMENTS_SUCCESS', 
      payload: { documents, page, hasMore } 
    });
  } catch (error) {
    dispatch({ type: 'FETCH_DOCUMENTS_ERROR', payload: error });
  }
};
```

## Performance Considerations

### 1. Data Caching Strategy

```typescript
interface CacheStrategy {
  // In-Memory Cache
  components: LRUCache<UUID, Component>;
  documents: LRUCache<UUID, Document>;
  queryResults: LRUCache<string, QueryResponse>;
  
  // IndexedDB for Offline Support
  offlineData: {
    siteOverview: Site;
    recentDocuments: Document[];
    componentHierarchy: Component[];
  };
  
  // Cache Invalidation
  invalidation: {
    ttl: number;
    eventBased: boolean;
    selective: boolean;
  };
}
```

### 2. Query Optimization

```sql
-- Materialized View for Component Status Summary
CREATE MATERIALIZED VIEW component_status_summary AS
SELECT 
  site_id,
  component_type,
  status,
  COUNT(*) as count,
  MAX(updated_at) as last_updated
FROM site_components
GROUP BY site_id, component_type, status;

-- Partial Index for Active Components
CREATE INDEX idx_active_components 
ON site_components(site_id, component_type) 
WHERE status = 'operational';

-- Composite Index for Timeline Queries
CREATE INDEX idx_timeline_composite 
ON site_events(site_id, event_date DESC, priority) 
WHERE event_date > CURRENT_DATE - INTERVAL '90 days';
```

### 3. Data Pagination

```typescript
interface PaginationStrategy {
  // Cursor-based for Timeline
  timeline: {
    cursor: string;  // Encoded timestamp + ID
    limit: number;
    direction: 'forward' | 'backward';
  };
  
  // Offset-based for Search Results
  search: {
    offset: number;
    limit: number;
    total: number;
  };
  
  // Virtual Scrolling for Large Lists
  virtualScroll: {
    viewportSize: number;
    bufferSize: number;
    itemHeight: number;
  };
}
```

## Security Architecture

### 1. Data Access Control

```typescript
interface AccessControl {
  // Row-Level Security
  siteAccess: {
    userId: UUID;
    siteIds: UUID[];
    role: 'admin' | 'manager' | 'viewer' | 'technician';
  };
  
  // Field-Level Security
  fieldAccess: {
    role: string;
    allowedFields: string[];
    maskedFields: string[];
  };
  
  // Operation-Level Security
  operations: {
    create: string[];
    read: string[];
    update: string[];
    delete: string[];
  };
}
```

### 2. Data Encryption

```typescript
interface EncryptionStrategy {
  // At Rest
  database: {
    algorithm: 'AES-256';
    keyRotation: 'quarterly';
  };
  
  // In Transit
  api: {
    protocol: 'HTTPS';
    minTLSVersion: '1.2';
  };
  
  // Sensitive Fields
  pii: {
    fields: ['email', 'phone', 'address'];
    method: 'field-level-encryption';
  };
}
```

### 3. Audit Trail

```typescript
interface AuditLog {
  id: UUID;
  userId: UUID;
  action: string;
  resourceType: string;
  resourceId: UUID;
  changes: {
    before: Record<string, any>;
    after: Record<string, any>;
  };
  metadata: {
    ipAddress: string;
    userAgent: string;
    sessionId: string;
  };
  timestamp: Date;
}
```

## Implementation Recommendations

### 1. Frontend Framework Integration

```typescript
// React/Next.js Integration
interface WidgetProps {
  siteId: string;
  apiKey: string;
  theme?: ThemeConfig;
  features?: FeatureFlags;
  onError?: (error: Error) => void;
}

// Widget Initialization
const EngramIQWidget: React.FC<WidgetProps> = ({
  siteId,
  apiKey,
  theme,
  features,
  onError
}) => {
  const [state, dispatch] = useReducer(widgetReducer, initialState);
  const queryClient = new QueryClient();
  
  return (
    <QueryClientProvider client={queryClient}>
      <WidgetProvider value={{ state, dispatch }}>
        <ThemeProvider theme={theme}>
          <WidgetContainer />
        </ThemeProvider>
      </WidgetProvider>
    </QueryClientProvider>
  );
};
```

### 2. Data Synchronization

```typescript
// WebSocket Connection for Real-time Updates
const useRealtimeSync = (siteId: string) => {
  useEffect(() => {
    const ws = new WebSocket(`wss://api.engramiq.com/v1/sites/${siteId}/events`);
    
    ws.onmessage = (event) => {
      const update = JSON.parse(event.data);
      
      switch (update.type) {
        case 'DOCUMENT_PROCESSED':
          queryClient.invalidateQueries(['documents', update.documentId]);
          break;
        case 'COMPONENT_STATUS_CHANGED':
          queryClient.setQueryData(['components', update.componentId], update.data);
          break;
        case 'NEW_ACTION_EXTRACTED':
          queryClient.invalidateQueries(['timeline', siteId]);
          break;
      }
    };
    
    return () => ws.close();
  }, [siteId]);
};
```

### 3. Error Handling

```typescript
interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
  retryCount: number;
}

class WidgetErrorBoundary extends React.Component<Props, ErrorBoundaryState> {
  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // Log to error service
    errorReporter.logError({
      error,
      errorInfo,
      context: {
        siteId: this.props.siteId,
        userId: this.props.userId,
        timestamp: new Date(),
      },
    });
    
    // Update state
    this.setState({
      hasError: true,
      error,
      errorInfo,
    });
    
    // Notify parent
    this.props.onError?.(error);
  }
}
```

## Migration Strategy

### 1. Data Migration Plan

```sql
-- Step 1: Create new schema
CREATE SCHEMA widget_v2;

-- Step 2: Migrate with transformation
INSERT INTO widget_v2.components
SELECT 
  id,
  site_id,
  type,
  jsonb_build_object(
    'manufacturer', manufacturer,
    'model', model,
    'capacity', specifications->>'capacity'
  ) as metadata
FROM public.site_components;

-- Step 3: Create migration tracking
CREATE TABLE migration_status (
  table_name VARCHAR(255),
  migrated_count INTEGER,
  total_count INTEGER,
  started_at TIMESTAMP,
  completed_at TIMESTAMP,
  status VARCHAR(50)
);
```

### 2. Backwards Compatibility

```typescript
// API Versioning
const apiClient = {
  v1: {
    getComponents: (siteId: string) => 
      fetch(`/api/v1/sites/${siteId}/components`),
  },
  v2: {
    getComponents: (siteId: string, options?: QueryOptions) =>
      fetch(`/api/v2/sites/${siteId}/components`, {
        body: JSON.stringify(options),
      }),
  },
};

// Feature Detection
const useComponentData = (siteId: string) => {
  const apiVersion = detectAPIVersion();
  
  if (apiVersion >= 2) {
    return useQuery(['components', 'v2', siteId], 
      () => apiClient.v2.getComponents(siteId, { includeRelationships: true })
    );
  }
  
  // Fallback to v1
  return useQuery(['components', 'v1', siteId], 
    () => apiClient.v1.getComponents(siteId)
  );
};
```

## Conclusion

This data architecture provides a robust foundation for the EngramIQ widget, supporting:

1. **Scalability**: Efficient data structures and indexing for large deployments
2. **Intelligence**: AI-powered features through embeddings and semantic search
3. **Real-time**: Event-driven architecture for live updates
4. **Flexibility**: JSONB fields and extensible schemas for future growth
5. **Security**: Comprehensive access control and encryption

The architecture balances performance, functionality, and maintainability while providing a solid foundation for the widget's current and future capabilities.