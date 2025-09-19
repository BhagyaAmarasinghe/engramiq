# EngramIQ - Solar Asset Reporting Agent

<div align="center">

![EngramIQ Logo](https://img.shields.io/badge/EngramIQ-Advanced%20Site%20Memory-17c480?style=for-the-badge)

**Transform unstructured operational data into actionable insights for solar asset management**

[![Next.js](https://img.shields.io/badge/Next.js-14.0-black?style=flat-square)](https://nextjs.org/)
[![Go](https://img.shields.io/badge/Go-1.21-00ADD8?style=flat-square)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat-square)](https://postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED?style=flat-square)](https://docker.com/)
[![AI](https://img.shields.io/badge/OpenAI-GPT--4-412991?style=flat-square)](https://openai.com/)

</div>

##  Quick Start

Get EngramIQ running in under 2 minutes:

```bash
# 1. Clone the repository
git clone <repository-url>
cd engramiq

# 2. Set up environment
cp .env.example .env
# Edit .env with your OpenAI API key

# 3. Start everything
make quickstart
```

**That's it!** Visit [http://localhost:3000](http://localhost:3000) to access EngramIQ.

##  Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Next.js       │────▶│   Go API        │────▶│  PostgreSQL     │
│   Frontend      │     │   Backend       │     │  + pgvector     │
│   (Port 3000)   │     │   (Port 8080)   │     │  (Port 5432)    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
        │                         │                         │
        │                         ▼                         │
        │               ┌─────────────────┐                 │
        └──────────────▶│     Redis       │◀────────────────┘
                        │   (Port 6379)   │
                        └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │   OpenAI API    │
                        │   (GPT-4 +      │
                        │   Embeddings)   │
                        └─────────────────┘
```

##  Tech Stack

### Frontend
- **Next.js 14** - React framework with App Router
- **HeroUI** - Modern, accessible component library
- **Tailwind CSS** - Utility-first CSS framework
- **TypeScript** - Type-safe development

### Backend
- **Go** - High-performance API server
- **Fiber** - Express-inspired web framework
- **GORM** - Go ORM with PostgreSQL
- **pgvector** - Vector similarity search

### Database & AI
- **PostgreSQL 15** - Primary database with vector extensions
- **Redis** - Caching and session storage
- **OpenAI GPT-4** - Natural language processing
- **Vector Embeddings** - Semantic document search

### Infrastructure
- **Docker Compose** - Multi-container orchestration
- **Nginx** - Reverse proxy (production)
- **GitHub Actions** - CI/CD pipeline

##  Usage Guide

### 1. Upload Documents
```bash
# Supported formats: PDF, DOCX, emails, meeting transcripts
# Drag and drop files or click to upload
# Documents are automatically processed and made searchable
```

### 2. Ask Questions
```
"What maintenance was performed on inverter INV001 last month?"
"Show me all fault reports from the past week"
"Which components require preventive maintenance?"
```

### 3. Get Sourced Answers
```
 Maintenance performed on INV001:
   - Quarterly inspection on March 15th [Source: Service Report]
   - DC disconnect replacement on March 22nd [Source: Work Order #1234]
   
 Confidence: 94%
 Sources: Field Service Report - March 2024, Work Order #1234
```

##  Deployment Options

### Option 1: Quick Development (Recommended)
```bash
make dev  # Starts all services with Docker Compose
```
-  Database migrations run automatically
-  Sample data pre-loaded (Site S2367 with 36 inverters)
-  All services with hot reloading

### Option 2: Production Build
```bash
make build    # Build all Docker images
make start    # Start in production mode
```

### Option 3: Individual Services
```bash
make frontend-dev  # Frontend only on port 3000
make backend-dev   # Backend only on port 8080
```

### Option 4: Using Nginx Proxy
```bash
# Access everything through port 80
docker-compose up nginx
# Frontend: http://localhost/
# API: http://localhost/api/
```

##  Configuration

### Environment Variables
```bash
# Required
LLM_API_KEY=your-openai-api-key-here

# Optional
JWT_SECRET=your-secret-key
DATABASE_URL=postgres://...
REDIS_URL=redis://...
```

### Makefile Commands
```bash
make help          # Show all available commands
make dev           # Start development environment
make build         # Build for production
make logs          # View service logs
make status        # Check service status
make clean         # Clean all caches and builds
make test          # Run test suite
```

##  Features

###  Document Processing
- Multi-format file upload (PDF, DOCX, emails)
- Automatic text extraction and processing
- AI-powered action extraction
- Component relationship mapping
- Deduplication and versioning

###  Natural Language Queries
- Conversational AI interface
- Source attribution for all answers
- Confidence scoring
- Hallucination prevention
- Professional behavior guards

###  Site Management
- Component status monitoring
- Timeline view of activities
- Performance metrics dashboard
- Alert and notification system

###  Security & Compliance
- JWT authentication ready
- Role-based access control
- Data encryption at rest and in transit
- GDPR compliance features
- Audit logging

##  Security

- **Authentication**: JWT tokens with refresh mechanism
- **Authorization**: Role-based access control
- **Data Protection**: Encryption at rest and in transit
- **Input Validation**: Comprehensive sanitization
- **Rate Limiting**: API endpoint protection

##  Performance

- **Response Time**: <2s for most queries
- **Throughput**: 1000+ concurrent users
- **Storage**: Efficient vector indexing
- **Caching**: Multi-layer caching strategy

##  Docker Support

Full Docker Compose setup includes:
- Frontend (Next.js)
- Backend (Go API)
- PostgreSQL with pgvector
- Redis cache
- Nginx reverse proxy
- PgAdmin for database management

```

Full API documentation available at `http://localhost:8080/swagger/`

## Monitoring

### Health Checks
```bash
curl http://localhost:8080/api/v1/health  # Backend health
curl http://localhost:3000/api/health     # Frontend health
```

### Logging
```bash
make logs              # All services
make logs-backend      # Backend only
make logs-frontend     # Frontend only
make logs-db           # Database only
```

### Database Access
```bash
make db-shell          # PostgreSQL shell
# PgAdmin: http://localhost:5050 (admin@engramiq.dev / admin123)
```

## Troubleshooting

### Common Issues

**Services won't start:**
```bash
make clean && make dev
```

**Database connection failed:**
```bash
docker-compose restart postgres
make db-migrate
```

**Frontend not updating:**
```bash
cd frontend && rm -rf .next node_modules
npm install && npm run dev
```

**Missing API key:**
```bash
# Edit .env file with your OpenAI API key
LLM_API_KEY=sk-...
```