# Engramiq Solar Asset Reporting Agent

A production-ready backend system for solar asset managers with complete database migration system and pre-populated data from Supporting-Documents. Process unstructured operational data, extract maintenance actions, and provide natural language querying with advanced AI features.

## Overview

The Engramiq Solar Asset Reporting Agent transforms how solar asset managers handle operational data by:

- **Complete Migration System**: Production-ready database migrations with site S2367 and 46+ SOLECTRIA inverters pre-loaded
- **Real Document Processing**: Field service reports from Sample-Reports.pdf automatically processed and queryable
- **AI-Powered Analysis**: Use GPT-4 to extract maintenance actions and answer natural language queries
- **Professional Compliance**: Source-attributed responses with content filtering and professional tone
- **Timeline Visualization**: Track events and maintenance history across solar assets

## System Status: PRODUCTION READY

### Complete Migration System
- **Database Migrations**: Production-ready migration framework with tracking
- **Site S2367**: Pre-populated with 46+ SOLECTRIA inverters from inverter_nodes.json
- **Field Service Reports**: Sample-Reports.pdf data automatically processed and searchable
- **Zero-Setup Data**: Ready to query inverter 40 arc protection, inverter 31 replacement, and more

### Enhanced Query System (PRD Implementation)
- **No Hallucinations**: All responses based on actual documents with source attribution
- **Professional Behavior**: Content filtering and professional tone enforcement  
- **Complete Traceability**: Full audit trail of information sources with citations
- **Advanced Intent Analysis**: GPT-4 powered query understanding and entity extraction
- **Supporting-Documents Ready**: Queries work with actual inverter and maintenance data

### Production Infrastructure  
- **Docker Integration**: Complete containerized deployment with health checks
- **Database Management**: PostgreSQL 15 with pgvector, Redis caching, PgAdmin interface
- **Migration Tracking**: Idempotent migrations with rollback support
- **API Testing**: Comprehensive test suite with quick_test.sh

## Quick Start

### Prerequisites
- Docker and Docker Compose
- OpenAI API key (for AI features)
- Go 1.24+ (for development)

### Setup with Docker (Recommended)

1. **Clone and Setup**
   ```bash
   git clone <repository-url>
   cd engramiq-backend
   cp .env.example .env
   ```

2. **Configure Environment**
   Edit `.env` and `docker-compose.override.yml`:
   ```env
   # Add your OpenAI API key to .env
   LLM_API_KEY=your-openai-api-key-here
   OPENAI_API_KEY=your-openai-api-key-here
   ```

3. **Start Complete System**
   ```bash
   # Start all services with migrations
   make docker-up
   
   # Database migrations run automatically on startup
   # Site S2367 and SOLECTRIA inverters are populated automatically
   ```

4. **Verify Setup and Test**
   ```bash
   # Check service status
   make status
   
   # Test with pre-loaded data
   ./quick_test.sh
   
   # Or test manually
   curl http://localhost:8080/api/v1/health
   ```

5. **Explore Pre-loaded Data**
   ```bash
   # Check migrated SOLECTRIA inverters
   SITE_ID=$(docker exec -i engramiq-postgres psql -U engramiq -d engramiq -t -c "SELECT id FROM sites WHERE site_code = 'S2367';" | xargs)
   curl "http://localhost:8080/api/v1/sites/$SITE_ID/components" | jq '.components[] | select(.specifications.manufacturer == "SOLECTRIA") | {name, model: .specifications.model, capacity: .specifications.capacity_kw}'
   
   # Test enhanced queries with actual data
   curl -X POST "http://localhost:8080/api/v1/sites/$SITE_ID/queries" \
     -H "Content-Type: application/json" \
     -d '{"query_text": "What was the issue with inverter 40?", "enhanced": true}' | jq .
   ```

### Manual Setup

1. **Start PostgreSQL with pgvector**
   ```bash
   # Using Docker
   docker run -d \
     --name engramiq-postgres \
     -e POSTGRES_USER=engramiq \
     -e POSTGRES_PASSWORD=engramiq_dev_2024 \
     -e POSTGRES_DB=engramiq \
     -p 5432:5432 \
     pgvector/pgvector:pg15
   ```

2. **Build and Run**
   ```bash
   go mod tidy
   go build -o bin/api ./cmd/api
   ./bin/api
   ```

## API Usage

### Upload and Process Documents

```bash
# Upload a field service report
curl -X POST "http://localhost:8080/api/v1/sites/{siteId}/documents" \
  -F "file=@field_service_report.pdf" \
  -F "document_type=field_service_report"

# Process document for action extraction
curl -X POST "http://localhost:8080/api/v1/documents/{documentId}/process"
```

### Natural Language Queries

```bash
# Enhanced query with source attribution
curl -X POST "http://localhost:8080/api/v1/sites/{siteId}/queries" \
  -H "Content-Type: application/json" \
  -d '{
    "query_text": "What maintenance was performed on inverter INV001 last month?",
    "enhanced": true
  }'
```

**Response Format**:
```json
{
  "answer": "Based on the field service reports, inverter INV001 underwent preventive maintenance on March 15th [Source 1] and had a faulty DC disconnect replaced on March 22nd [Source 2].",
  "confidence_score": 0.92,
  "sources": [
    {
      "document_id": "uuid-here",
      "document_title": "Field Service Report - March 2024",
      "relevant_excerpt": "Performed quarterly maintenance on INV001...",
      "citation": "Field Service Report - March 2024 (2024-03-15), p. 3"
    }
  ],
  "no_hallucination": true,
  "response_type": "maintenance_history",
  "processing_time_ms": 1250
}
```

### Component Management

```bash
# Create site components
curl -X POST "http://localhost:8080/api/v1/sites/{siteId}/components" \
  -H "Content-Type: application/json" \
  -d '{
    "component_type": "inverter",
    "name": "INV001",
    "external_id": "INV001",
    "specifications": {
      "manufacturer": "SolarEdge",
      "model": "SE25K",
      "capacity_kw": 25.0
    }
  }'

# Get component maintenance history
curl "http://localhost:8080/api/v1/components/{componentId}/actions"
```

## Architecture

### Technology Stack
- **Backend**: Go 1.25+ with Fiber web framework
- **Database**: PostgreSQL 15+ with pgvector extension
- **Cache**: Redis 7 for session and query caching
- **AI**: OpenAI GPT-4 and text-embedding-ada-002
- **Infrastructure**: Docker and Docker Compose

### System Architecture
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │───▶│  Fiber Handler  │───▶│  Service Layer  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                       ┌─────────────────┐             ▼
                       │  Content Filter │    ┌─────────────────┐
                       │   & Validation  │◀───│ Repository Layer│
                       └─────────────────┘    └─────────────────┘
                                                        │
┌─────────────────┐    ┌─────────────────┐             ▼
│   OpenAI API    │◀───│   LLM Service   │    ┌─────────────────┐
│     (GPT-4)     │    │  Enhancement    │    │  PostgreSQL +   │
└─────────────────┘    └─────────────────┘    │    pgvector     │
                                              └─────────────────┘
```

### Key Components

- **Domain Layer**: Business entities and rules
- **Repository Layer**: Data access with semantic search
- **Service Layer**: Business logic and AI integration
- **Handler Layer**: HTTP request processing
- **Infrastructure Layer**: Database and external services

## Development

### Available Commands

```bash
# Docker management
make docker-up      # Start all Docker services
make docker-down    # Stop all services
make docker-clean   # Clean volumes and containers

# Development workflow
make dev            # Start database + application
make build          # Build Go application
make run            # Run application
make test           # Run test suite

# Database operations
make check-db       # Verify database connectivity
make check-redis    # Verify Redis connectivity
make status         # Show all service status
```

### Environment Configuration

**Database Settings**:
```env
DATABASE_URL=postgres://engramiq:engramiq_dev_2024@localhost:5432/engramiq?sslmode=disable
```

**OpenAI Configuration**:
```env
LLM_API_KEY=your-openai-api-key-here
LLM_MODEL=gpt-4
LLM_PROVIDER=openai
```

**Server Configuration**:
```env
PORT=8080
ENVIRONMENT=development
CORS_ORIGINS=http://localhost:3000,http://localhost:3001
```

### Database Access

**Services**:
- **Application Database**: `localhost:5432`
- **Redis Cache**: `localhost:6379`
- **PgAdmin Interface**: `http://localhost:5050`
  - Email: `admin@engramiq.dev`
  - Password: `admin123`

**Direct Database Access**:
```bash
# Connect to PostgreSQL
docker exec -it engramiq-postgres psql -U engramiq -d engramiq

# Check extensions
SELECT extname, extversion FROM pg_extension WHERE extname IN ('vector', 'uuid-ossp');

# View tables
\dt
```

## API Documentation

### Core Endpoints

#### Document Management
```
POST   /api/v1/sites/{siteId}/documents          # Upload document
GET    /api/v1/sites/{siteId}/documents          # List documents
GET    /api/v1/documents/{id}                    # Get document details
DELETE /api/v1/documents/{id}                    # Delete document
POST   /api/v1/documents/{id}/process            # Process with AI
GET    /api/v1/sites/{siteId}/documents/search   # Search documents
```

#### Enhanced Query System
```
POST   /api/v1/sites/{siteId}/queries            # Submit enhanced query
GET    /api/v1/queries/{id}                      # Get query results
GET    /api/v1/queries/history                   # Query history
GET    /api/v1/sites/{siteId}/queries/similar    # Find similar queries
GET    /api/v1/sites/{siteId}/analytics/queries  # Query analytics
```

#### Component Management
```
POST   /api/v1/sites/{siteId}/components         # Create component
GET    /api/v1/sites/{siteId}/components         # List components
GET    /api/v1/components/{id}                   # Get component
PUT    /api/v1/components/{id}                   # Update component
DELETE /api/v1/components/{id}                   # Delete component
POST   /api/v1/sites/{siteId}/components/bulk    # Bulk operations
```

#### Action and Timeline
```
GET    /api/v1/sites/{siteId}/timeline           # Site event timeline
GET    /api/v1/sites/{siteId}/actions            # List actions
GET    /api/v1/actions/{id}                      # Action details
GET    /api/v1/components/{id}/actions           # Component actions
```

## Features Deep Dive

### PRD Implementation: Enhanced Query System

The system implements advanced Product Requirements Document (PRD) specifications:

**No Hallucinations**: 
- All responses are generated using only retrieved source documents
- RAG (Retrieval-Augmented Generation) pattern ensures factual accuracy
- Source validation prevents unsupported claims

**Professional Behavior**:
- Content filtering blocks inappropriate or off-topic queries
- Professional tone enforcement in all responses
- Sensitive information sanitization

**Complete Traceability**:
- Full audit trail of information sources
- Formatted citations with page numbers and sections
- Query analytics and performance tracking

### AI-Powered Document Processing

**Automatic Action Extraction**:
```go
// Example extracted action
{
  "action_type": "maintenance",
  "description": "Replaced faulty DC disconnect switch on inverter INV001",
  "component_type": "inverter",
  "component_id": "INV001",
  "technician_names": ["John Smith", "Mike Johnson"],
  "work_order_number": "WO-2024-0315",
  "action_date": "2024-03-15T10:30:00Z",
  "confidence_score": 0.95
}
```

**Semantic Search**:
- Vector embeddings for intelligent document similarity
- Context-aware search across all content types
- Relevance scoring and ranking

### Professional Query Processing Pipeline

```
User Query → Content Validation → Intent Analysis → Source Retrieval → 
Response Generation → Source Validation → Professional Tone → Final Response
```

Each step ensures quality, accuracy, and professional standards.

## Security

### Input Validation
- Query content filtering and validation
- Professional topic enforcement
- SQL injection prevention
- XSS protection

### Authentication (Ready for Implementation)
- JWT token-based authentication
- Refresh token mechanism
- Site-based access control
- Role-based permissions

### Data Protection
- Sensitive information sanitization
- Secure configuration management
- Encrypted connections to external services

## Performance

### Database Optimization
- Connection pooling with configurable limits
- Prepared statement caching
- Efficient indexing for semantic search
- Query optimization for large datasets

### Caching Strategy
- Redis-based query result caching
- Embedding caching to reduce AI costs
- Session and metadata caching

### Scalability
- Stateless service design
- Horizontal scaling ready
- Load balancer compatible
- Microservices architecture

## Troubleshooting

### Common Issues

**Database Connection Issues**:
```bash
# Check if PostgreSQL is running
make check-db

# View database logs
docker logs engramiq-postgres

# Restart database
make docker-down && make docker-up
```

**Application Startup Issues**:
```bash
# Check application logs
docker logs engramiq-backend

# Verify environment configuration
cat .env

# Test API health
curl http://localhost:8080/api/v1/health
```

**OpenAI API Issues**:
- Verify API key in `.env` file
- Check API quota and rate limits
- Monitor API usage in OpenAI dashboard

**FIXED: Automatic Processing Issues**:
- **Previous Issue**: Database schema mismatch preventing document processing
- **Resolution**: Removed invalid `extracted_entities` column reference
- **Current Status**: Documents now process automatically with embeddings
- **File Fixed**: `internal/service/document_service.go:164`

### Development Tips

**Database Debugging**:
```sql
-- Check vector extension
SELECT * FROM pg_extension WHERE extname = 'vector';

-- View document processing status
SELECT processing_status, COUNT(*) FROM documents GROUP BY processing_status;

-- Check embedding similarity
SELECT id, title, embedding <-> '[0.1,0.2,...]'::vector as distance 
FROM documents ORDER BY distance LIMIT 5;
```

**Performance Monitoring**:
- Use structured logging for debugging
- Monitor response times through application logs
- Track OpenAI API usage and costs

## Production Deployment

### Environment Preparation
- Configure production database with proper security
- Set up Redis cluster for high availability
- Configure load balancer and SSL certificates
- Set up monitoring and logging infrastructure

### Security Checklist
- [ ] Environment variables secured
- [ ] Database access restricted
- [ ] API rate limiting configured
- [ ] CORS properly configured
- [ ] Security headers enabled
- [ ] SSL/TLS certificates installed

### Monitoring Setup
- [ ] Application performance monitoring
- [ ] Database performance tracking
- [ ] Error tracking and alerting
- [ ] Usage analytics and reporting

## Testing Results

### Migration System Verification

The system is pre-loaded with Supporting-Documents data and ready for immediate testing:

```bash
# Run comprehensive test suite
./quick_test.sh

# Test results include:
# - 46+ SOLECTRIA inverters from inverter_nodes.json migration
# - Enhanced queries about inverter 40 arc protection issue
# - Enhanced queries about inverter 31 replacement work
# - Content filtering blocking off-topic queries
# - Source attribution with proper citations
```

**Expected Test Results**:
- **Migration Status**: 20240915000001 successfully applied
- **Site S2367**: Combined Site with 46+ SOLECTRIA inverters
- **Query: "What was the issue with inverter 40?"**: Returns "Arc Protect error" with source citation
- **Query: "What work was performed on inverter 31?"**: Returns "inverter replacement with PVI14TL" with details
- **Query: "What is the weather?"**: Properly rejected as off-topic
- **Content Filtering**: Professional behavioral guards operational

### Production Readiness Checklist

- [x] **Database Migration System**: Complete with tracking and rollback support
- [x] **Supporting-Documents Integration**: Site S2367 and inverters pre-loaded  
- [x] **Enhanced Query System**: Working with actual field service report data
- [x] **Content Filtering**: Professional behavior guards operational
- [x] **Docker Integration**: Full containerized deployment
- [x] **API Testing**: Comprehensive test suite passes
- [x] **Documentation**: Complete API guides and troubleshooting
- [x] **Error Handling**: Robust error recovery and logging

---
