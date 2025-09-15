# Docker Setup for Engramiq Backend

Complete Docker containerization of the Engramiq Solar Asset Reporting Agent backend with PostgreSQL and Redis.

## Quick Start

### 1. Setup Environment
```bash
# Copy environment template
make setup-env

# Edit .env file and add your OpenAI API key
vim .env
# Set: LLM_API_KEY=your-openai-api-key-here
```

### 2. Start All Services
```bash
# Build and start everything (backend, database, redis)
make docker-build

# Or just start existing images
make docker-up
```

### 3. Verify Everything is Working
```bash
# Check service health
make health

# Check service status
make status
```

**Backend API**: http://localhost:8080
**PgAdmin**: http://localhost:5050 (admin@engramiq.dev / admin123)

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│                 │    │                 │    │                 │
│   Frontend UI   │───▶│  Backend API    │───▶│   PostgreSQL    │
│                 │    │   (Go Fiber)    │    │   + pgvector    │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                          
                              ▼                          
                       ┌─────────────────┐               
                       │                 │               
                       │     Redis       │               
                       │    (Cache)      │               
                       │                 │               
                       └─────────────────┘               
```

## Services

### Backend API (`backend`)
- **Image**: Built from local Dockerfile
- **Port**: 8080
- **Health Check**: `GET /api/v1/health`
- **Features**: Document processing, LLM integration, query system

### PostgreSQL (`postgres`)  
- **Image**: `pgvector/pgvector:pg15`
- **Port**: 5432
- **Extensions**: pgvector, uuid-ossp
- **Database**: `engramiq`

### Redis (`redis`)
- **Image**: `redis:7-alpine`  
- **Port**: 6379
- **Purpose**: Caching, session storage

### PgAdmin (`pgadmin`)
- **Image**: `dpage/pgadmin4:latest`
- **Port**: 5050
- **Login**: admin@engramiq.dev / admin123

## Development Workflows

### Option 1: Full Docker Stack
Everything runs in containers (recommended for production-like testing):

```bash
# Build and start all services
make docker-build

# View logs
make docker-logs

# Restart backend only
docker-compose restart backend
```

### Option 2: Hybrid Development
Database in Docker, backend running locally (recommended for active development):

```bash
# Start only database services
docker-compose up -d postgres redis pgadmin

# Run backend locally with hot reload
make build && make run
```

## Configuration

### Environment Variables

**Required:**
- `LLM_API_KEY`: OpenAI API key for document processing

**Optional:**
- `JWT_SECRET`: JWT signing secret (auto-generated if not provided)
- `ENVIRONMENT`: development/production (default: production)
- `CORS_ORIGINS`: Allowed frontend origins
- `LOG_LEVEL`: debug/info/warn/error

### Docker Compose Files

- `docker-compose.yml`: Main production configuration
- `docker-compose.override.yml`: Development overrides (auto-applied)
- `.env`: Environment variables (created from `.env.docker.example`)

## API Endpoints

Once running, the backend exposes:

```
# Health Check
GET    /api/v1/health

# Document Management  
POST   /api/v1/sites/{siteId}/documents
GET    /api/v1/sites/{siteId}/documents
GET    /api/v1/documents/{id}
DELETE /api/v1/documents/{id}

# Enhanced Query System
POST   /api/v1/sites/{siteId}/queries
GET    /api/v1/queries/history

# Component Management
GET    /api/v1/sites/{siteId}/components
GET    /api/v1/components/{id}/actions

# Timeline & Actions
GET    /api/v1/sites/{siteId}/timeline
GET    /api/v1/sites/{siteId}/actions
```

## Data Persistence

All data is persisted in Docker volumes:

- `backend_uploads`: Uploaded documents and files
- `postgres_data`: Database tables and indexes
- `redis_data`: Cache and session data  
- `pgadmin_data`: PgAdmin configuration

```bash
# View volumes
docker volume ls | grep engramiq

# Backup database
docker exec engramiq-postgres pg_dump -U engramiq engramiq > backup.sql

# Clean all data (destructive)
make docker-clean
```

## Monitoring & Debugging

### Health Checks
```bash
# All services
make health

# Individual services
curl http://localhost:8080/api/v1/health
docker exec engramiq-postgres pg_isready -U engramiq -d engramiq
docker exec engramiq-redis redis-cli ping
```

### Logs
```bash
# All services
make docker-logs

# Individual services
docker logs engramiq-backend -f
docker logs engramiq-postgres -f
docker logs engramiq-redis -f
```

### Database Access
```bash
# Via PgAdmin: http://localhost:5050
# Direct access:
docker exec -it engramiq-postgres psql -U engramiq -d engramiq
```

## Makefile Commands

```bash
# Setup
make setup-env      # Create .env from template
make help          # Show all commands

# Docker Full Stack
make docker-build   # Build and start all services
make docker-rebuild # Force rebuild all services  
make docker-up      # Start all services
make docker-down    # Stop all services
make docker-clean   # Stop and remove volumes

# Development
make dev           # Start DB services, run backend locally
make build         # Build Go application
make run           # Run application locally

# Utilities  
make status        # Show service status
make health        # Check service health
make docker-logs   # View all service logs
```

## Troubleshooting

### Common Issues

**Backend won't start:**
```bash
# Check if OpenAI API key is set
docker exec engramiq-backend env | grep LLM_API_KEY

# Check backend logs  
docker logs engramiq-backend
```

**Database connection errors:**
```bash
# Verify database is ready
make health

# Check database logs
docker logs engramiq-postgres

# Test direct connection
docker exec -it engramiq-postgres psql -U engramiq -d engramiq -c "SELECT 1;"
```

**Port conflicts:**
```bash
# Check what's using ports
lsof -i :8080
lsof -i :5432
lsof -i :6379

# Stop conflicting services or change ports in docker-compose.yml
```

### Performance

**Slow document processing:**
- Check OpenAI API rate limits
- Verify sufficient memory allocation
- Monitor Docker resource usage: `docker stats`

**Database performance:**
- Increase shared_buffers in PostgreSQL config
- Monitor query performance in PgAdmin
- Check index usage

## Production Deployment

For production deployment:

1. **Use external database** (managed PostgreSQL + Redis)
2. **Set production environment variables**
3. **Use Docker secrets** for API keys
4. **Enable SSL/TLS**
5. **Set up monitoring** (health checks, metrics)
6. **Configure backup strategy**

Example production docker-compose overrides:
```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  backend:
    environment:
      ENVIRONMENT: production
      DATABASE_URL: ${EXTERNAL_DATABASE_URL}
      REDIS_URL: ${EXTERNAL_REDIS_URL}
    deploy:
      replicas: 3
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M
```

## Integration with Frontend

The containerized backend can be integrated with any frontend:

### React/Vue/Angular
```javascript
const API_BASE = 'http://localhost:8080/api/v1';

// Upload document
const uploadDocument = async (siteId, file) => {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('document_type', 'field_service_report');
  
  return fetch(`${API_BASE}/sites/${siteId}/documents`, {
    method: 'POST',
    body: formData
  });
};

// Query site memory
const queryServer = async (siteId, question) => {
  return fetch(`${API_BASE}/sites/${siteId}/queries`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      query_text: question,
      enhanced: true
    })
  });
};
```

The Docker setup provides a production-ready backend that can be easily deployed and scaled.