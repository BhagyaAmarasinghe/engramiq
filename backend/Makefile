# Engramiq Backend Makefile

.PHONY: help build run test clean docker-up docker-down docker-logs db-migrate db-reset docker-build docker-rebuild

# Default target
help:
	@echo "Available commands:"
	@echo ""
	@echo "Local Development:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  dev          - Start database services and run application locally"
	@echo ""
	@echo "Docker Full Stack:"
	@echo "  docker-build - Build and start all services (backend, database, redis)"
	@echo "  docker-rebuild - Force rebuild all services"
	@echo "  docker-up    - Start all Docker services"
	@echo "  docker-down  - Stop Docker services"
	@echo "  docker-logs  - View Docker service logs"
	@echo "  docker-clean - Clean Docker resources"
	@echo ""
	@echo "Database:"
	@echo "  db-migrate   - Run database migrations"
	@echo "  db-reset     - Reset database"
	@echo ""
	@echo "Utils:"
	@echo "  setup-env    - Setup .env file from template"
	@echo "  status       - Show service status"
	@echo "  health       - Check service health"

# Application commands
build:
	@echo "Building application..."
	go build -o bin/api ./cmd/api

run: build
	@echo "Starting application..."
	./bin/api

test:
	@echo "Running tests..."
	go test -v ./internal/...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go mod tidy

# Docker commands - Full Stack
docker-build:
	@echo "Building and starting all Docker services..."
	docker-compose up -d --build
	@echo "Waiting for services to be ready..."
	@sleep 15
	@echo "Services started!"
	@echo "  Backend API: http://localhost:8080"
	@echo "  PostgreSQL: localhost:5432"
	@echo "  Redis: localhost:6379"
	@echo "  PgAdmin: http://localhost:5050 (admin@engramiq.dev / admin123)"

docker-rebuild:
	@echo "Force rebuilding all Docker services..."
	docker-compose build --no-cache
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 15
	@echo "Services rebuilt and started!"

docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Services started!"
	@echo "  Backend API: http://localhost:8080"
	@echo "  PostgreSQL: localhost:5432"
	@echo "  Redis: localhost:6379"
	@echo "  PgAdmin: http://localhost:5050 (admin@engramiq.dev / admin123)"

docker-down:
	@echo "Stopping Docker services..."
	docker-compose down

docker-logs:
	@echo "Viewing Docker logs..."
	docker-compose logs -f

docker-clean:
	@echo "Cleaning Docker resources..."
	docker-compose down -v
	docker system prune -f

# Database commands
db-migrate: docker-up
	@echo "Running database migrations..."
	@sleep 3
	go run ./cmd/migrate

db-reset: docker-down
	@echo "Resetting database..."
	docker-compose down -v
	docker-compose up -d postgres
	@sleep 5
	go run ./cmd/migrate

# Development workflow
dev: docker-up
	@echo "Starting development environment..."
	@sleep 5
	@echo "Database ready. Starting application..."
	@make run

# Health checks
check-db:
	@echo "Checking database connection..."
	@docker exec engramiq-postgres pg_isready -U engramiq -d engramiq || echo "Database not ready"

check-redis:
	@echo "Checking Redis connection..."
	@docker exec engramiq-redis redis-cli ping || echo "Redis not ready"

# Utils
setup-env:
	@echo "Setting up environment configuration..."
	@if [ ! -f .env ]; then \
		echo "Copying .env.docker.example to .env..."; \
		cp .env.docker.example .env; \
		echo ""; \
		echo ".env file created!"; \
		echo "Please edit .env and add your OpenAI API key to LLM_API_KEY"; \
		echo ""; \
	else \
		echo ".env file already exists"; \
	fi

status:
	@echo "Service Status:"
	@docker-compose ps

health:
	@echo "Checking service health..."
	@echo "Backend API:"
	@curl -s http://localhost:8080/api/v1/health | jq . 2>/dev/null || echo "Backend service not responding"
	@echo "Database:"
	@docker exec engramiq-postgres pg_isready -U engramiq -d engramiq 2>/dev/null && echo "Database ready" || echo "Database not ready"
	@echo "Redis:"
	@docker exec engramiq-redis redis-cli ping 2>/dev/null | grep -q "PONG" && echo "Redis ready" || echo "Redis not ready"