# EngramIQ Full Stack Makefile

.PHONY: help dev build start stop clean logs frontend-dev backend-dev docker-up docker-down status

# Default target
.DEFAULT_GOAL := help

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

# Help command
help: ## Show this help message
	@echo "${BLUE}EngramIQ Full Stack Commands${NC}"
	@echo ""
	@echo "${YELLOW}Usage:${NC}"
	@echo "  make ${GREEN}<target>${NC}"
	@echo ""
	@echo "${YELLOW}Targets:${NC}"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  ${GREEN}%-15s${NC} %s\n", $$1, $$2}'

# Development Commands
dev: ## Start full stack in development mode (recommended)
	@echo "${BLUE}Starting EngramIQ in development mode...${NC}"
	@$(MAKE) docker-up
	@echo "${GREEN}✓ EngramIQ is running!${NC}"
	@echo ""
	@echo "${YELLOW}Access:${NC}"
	@echo "  Frontend:  ${GREEN}http://localhost:3000${NC}"
	@echo "  Backend:   ${GREEN}http://localhost:8080${NC}"
	@echo "  PgAdmin:   ${GREEN}http://localhost:5050${NC}"
	@echo ""
	@echo "${YELLOW}Logs:${NC} make logs"
	@echo "${YELLOW}Stop:${NC} make stop"

docker-up: ## Start all services with Docker Compose
	@echo "${BLUE}Starting Docker services...${NC}"
	@if [ ! -f .env ]; then \
		echo "${YELLOW}Creating .env file...${NC}"; \
		echo "LLM_API_KEY=your-openai-api-key-here" > .env; \
		echo "JWT_SECRET=your-super-secret-jwt-key" >> .env; \
		echo "${RED}⚠️  Please update .env with your OpenAI API key${NC}"; \
	fi
	@docker-compose up -d
	@echo "${GREEN}✓ All services started${NC}"

docker-down: ## Stop all Docker services
	@echo "${BLUE}Stopping Docker services...${NC}"
	@docker-compose down
	@echo "${GREEN}✓ All services stopped${NC}"

docker-clean: ## Stop services and remove volumes
	@echo "${RED}Stopping services and removing volumes...${NC}"
	@docker-compose down -v
	@echo "${GREEN}✓ Clean complete${NC}"

# Individual service commands
frontend-dev: ## Start only frontend in dev mode
	@echo "${BLUE}Starting frontend...${NC}"
	@cd frontend && npm install && npm run dev

backend-dev: ## Start only backend in dev mode
	@echo "${BLUE}Starting backend...${NC}"
	@cd backend && go run cmd/api/main.go

frontend-build: ## Build frontend for production
	@echo "${BLUE}Building frontend...${NC}"
	@cd frontend && npm install && npm run build
	@echo "${GREEN}✓ Frontend built${NC}"

backend-build: ## Build backend for production
	@echo "${BLUE}Building backend...${NC}"
	@cd backend && go build -o bin/api ./cmd/api
	@echo "${GREEN}✓ Backend built${NC}"

# Production commands
build: ## Build all services for production
	@echo "${BLUE}Building all services...${NC}"
	@docker-compose build
	@echo "${GREEN}✓ Build complete${NC}"

start: ## Start in production mode
	@echo "${BLUE}Starting in production mode...${NC}"
	@docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
	@echo "${GREEN}✓ Production services started${NC}"

stop: ## Stop all services
	@$(MAKE) docker-down

# Monitoring commands
status: ## Show status of all services
	@echo "${BLUE}Service Status:${NC}"
	@docker-compose ps

logs: ## Show logs from all services
	@docker-compose logs -f

logs-backend: ## Show backend logs
	@docker-compose logs -f backend

logs-frontend: ## Show frontend logs
	@docker-compose logs -f frontend

logs-db: ## Show database logs
	@docker-compose logs -f postgres

# Database commands
db-migrate: ## Run database migrations
	@echo "${BLUE}Running migrations...${NC}"
	@docker-compose exec backend go run ./cmd/migrate
	@echo "${GREEN}✓ Migrations complete${NC}"

db-shell: ## Open PostgreSQL shell
	@docker-compose exec postgres psql -U engramiq -d engramiq

# Utility commands
clean: ## Clean all generated files and caches
	@echo "${BLUE}Cleaning...${NC}"
	@rm -rf frontend/.next frontend/node_modules
	@rm -rf backend/bin backend/tmp
	@echo "${GREEN}✓ Clean complete${NC}"

install: ## Install all dependencies
	@echo "${BLUE}Installing dependencies...${NC}"
	@cd frontend && npm install
	@cd backend && go mod download
	@echo "${GREEN}✓ Dependencies installed${NC}"

test: ## Run all tests
	@echo "${BLUE}Running tests...${NC}"
	@cd backend && go test ./...
	@cd frontend && npm test
	@echo "${GREEN}✓ Tests complete${NC}"

# Quick start for first time users
quickstart: ## First time setup and start
	@echo "${BLUE}Welcome to EngramIQ!${NC}"
	@echo ""
	@echo "${YELLOW}Setting up your environment...${NC}"
	@$(MAKE) install
	@$(MAKE) dev
	@echo ""
	@echo "${GREEN}✓ EngramIQ is ready!${NC}"
	@echo ""
	@echo "${YELLOW}Next steps:${NC}"
	@echo "1. Update ${GREEN}.env${NC} with your OpenAI API key"
	@echo "2. Visit ${GREEN}http://localhost:3000${NC} to access the app"
	@echo "3. Upload documents and start querying!"