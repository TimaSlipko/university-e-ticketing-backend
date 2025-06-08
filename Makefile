# Makefile for E-Ticketing System

# Variables
APP_NAME=eticketing
BINARY_NAME=bin/$(APP_NAME)
MAIN_PATH=cmd/server/main.go

# Default environment
ENV ?= development

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: setup
setup: ## Install dependencies and setup development environment
	go mod download
	go mod tidy
	@echo "Setup complete. Don't forget to:"
	@echo "1. Copy .env.example to .env"
	@echo "2. Configure your database credentials"
	@echo "3. Create the database: createdb e_ticketing_dev"

.PHONY: run
run: ## Run the application in development mode
	go run $(MAIN_PATH)

.PHONY: build
build: ## Build the application
	@mkdir -p bin
	go build -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary built: $(BINARY_NAME)"

.PHONY: build-linux
build-linux: ## Build the application for Linux
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux $(MAIN_PATH)
	@echo "Linux binary built: $(BINARY_NAME)-linux"

.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: lint
lint: ## Run linter (requires golangci-lint)
	golangci-lint run

.PHONY: fmt
fmt: ## Format code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: tidy
tidy: ## Tidy go modules
	go mod tidy

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

.PHONY: migrate
migrate: ## Run database migrations
	go run $(MAIN_PATH) -migrate

.PHONY: seed
seed: ## Seed database with sample data
	go run $(MAIN_PATH) -seed

.PHONY: dev-db
dev-db: ## Setup development database (migrate + seed)
	@echo "Setting up development database..."
	@$(MAKE) migrate
	@$(MAKE) seed
	@echo "Development database ready!"

.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(APP_NAME):latest -f docker/Dockerfile .

.PHONY: docker-run
docker-run: ## Run application in Docker
	docker-compose -f docker/docker-compose.yml up -d

.PHONY: docker-stop
docker-stop: ## Stop Docker containers
	docker-compose -f docker/docker-compose.yml down

.PHONY: docker-logs
docker-logs: ## View Docker logs
	docker-compose -f docker/docker-compose.yml logs -f

.PHONY: install-tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development tools installed"

.PHONY: check
check: fmt vet lint test ## Run all checks (format, vet, lint, test)

.PHONY: pre-commit
pre-commit: check ## Run pre-commit checks
	@echo "All checks passed! 🎉"

# Development shortcuts
.PHONY: dev
dev: ## Quick development setup
	@$(MAKE) setup
	@$(MAKE) dev-db
	@echo "Development environment ready!"
	@echo "Run 'make run' to start the server"

.PHONY: restart
restart: ## Restart the application
	@pkill -f "$(APP_NAME)" || true
	@$(MAKE) run

# Production deployment
.PHONY: deploy
deploy: ## Deploy to production (Docker)
	@echo "Deploying to production..."
	@$(MAKE) docker-build
	@$(MAKE) docker-run
	@echo "Deployment complete!"

.PHONY: status
status: ## Check application status
	@curl -s http://localhost:8080/health | jq . || echo "Application not running or jq not installed"