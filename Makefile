# News API - Simplified Makefile
# ================================
# Simple, clean commands for dev/test/prod environments
# Port allocation: dev:8081/5433/6380, test:8082/5434/6381, prod:8080/5433/6379

.PHONY: help deps-check dev test prod stop clean logs status health build seed db-tables db-reset db-status docs docs-serve docs-stop docs-local test-status test-local test-docker test-unit test-integration test-e2e test-all test-coverage test-watch test-run test-verbose test-setup test-clean version build-with-version release-tag changelog-update

# Default target
.DEFAULT_GOAL := help

# ==========================================
# HELP & SETUP
# ==========================================

help: ## Show this help message
	@echo "🚀 News API - Simplified Commands"
	@echo "================================="
	@echo ""
	@echo "🛠️  Setup & Dependencies:"
	@echo "  make deps        → Check and install dependencies"
	@echo ""
	@echo "🚀 Environment Commands:"
	@echo "  make dev         → Start development environment"
	@echo "  make dev-down    → Stop development environment"
	@echo "  make test        → Start test environment" 
	@echo "  make test-down   → Stop test environment"
	@echo "  make prod        → Start production environment"
	@echo "  make prod-down   → Stop production environment"
	@echo "  make stop        → Stop all environments"
	@echo ""
	@echo "🗄️  Database Commands:"
	@echo "  make db-tables   → List database tables"
	@echo "  make db-reset    → Reset database (drops all tables)"
	@echo "  make db-status   → Show database connection status"
	@echo ""
	@echo "🎯 Atlas Migration Commands:"
	@echo "  make atlas-status        → Show Atlas migration status"
	@echo "  make atlas-apply         → Apply Atlas migrations"
	@echo "  make atlas-status-docker → Atlas status via Docker"
	@echo "  make atlas-apply-docker  → Apply Atlas migrations via Docker"
	@echo "  make auto-migrate        → Auto GORM → Atlas migration"
	@echo "  make auto-migrate-commit → Auto migration + git commit"
	@echo "  make atlas-sync          → Sync GORM models to Atlas"
	@echo "  make db-mode-auto        → Switch to GORM AutoMigrate mode"
	@echo "  make db-mode-atlas       → Switch to Atlas migration mode"
	@echo ""
	@echo "🌱 Seed Commands:"
	@echo "  make seed        → Seed database (current env)"
	@echo "  make seed-dev    → Seed development database"
	@echo "  make seed-test   → Seed test database"
	@echo "  make seed-prod   → Seed production database"
	@echo ""
	@echo "🧪 Test Commands:"
	@echo "  make test-local      → Run tests locally (no Docker)"
	@echo "  make test-docker     → Run tests in Docker containers"
	@echo "  make test-unit       → Run unit tests"
	@echo "  make test-integration→ Run integration tests"
	@echo "  make test-e2e        → Run E2E tests"
	@echo "  make test-all        → Run all tests"
	@echo "  make test-coverage   → Run tests with coverage"
	@echo "  make test-watch      → Run tests in watch mode"
	@echo ""
	@echo "🔧 Utility Commands:"
	@echo "  make build       → Build images"
	@echo "  make logs        → Show logs"
	@echo "  make status      → Show environment status"
	@echo "  make health      → Check health"
	@echo "  make clean       → Clean Docker resources"
	@echo ""
	@echo "🏷️  Versioning Commands:"
	@echo "  make version           → Show current version info"
	@echo "  make build-with-version → Build with version information"
	@echo "  make release-tag       → Create release tag (VERSION=v1.1.0)"
	@echo "  make changelog-update  → Reminder to update changelog"
	@echo ""
	@echo "🌐 HTTP/2 Commands:"
	@echo "  make http2-dev   → Start HTTP/2 development server (H2C)"
	@echo "  make http2-prod  → Start HTTP/2 production server (HTTPS)"
	@echo "  make http2-test  → Test HTTP/2 connectivity"
	@echo "  make http2-certs → Generate self-signed certificates"
	@echo ""
	@echo "📚 Documentation Commands:"
	@echo "📚 Documentation Commands:"
	@echo "  make docs        → Generate Swagger docs (Docker-based)"
	@echo "  make docs-local  → Generate Swagger docs (local swag CLI)"
	@echo "  make docs-serve  → Serve Swagger UI at http://localhost:8091"
	@echo "  make docs-stop   → Stop Swagger UI container"
	@echo ""
	@echo "💡 Tip: Use 'make <command> ENV=<dev|test|prod>' to specify environment"
	@echo "📝 Note: Database schema uses GORM AutoMigrate - no manual migrations needed"

deps: deps-check ## Check and install dependencies

deps-check: ## Check if required dependencies are installed
	@echo "🔍 Checking dependencies..."
	@command -v docker >/dev/null 2>&1 || { echo "❌ Docker is required but not installed. Please install Docker first."; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "❌ Docker Compose is required but not installed. Please install Docker Compose first."; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "❌ Go is required but not installed. Please install Go first."; exit 1; }
	@echo "✅ All dependencies are installed"
	@echo "📊 Versions:"
	@echo "  Docker: $$(docker --version | cut -d' ' -f3 | cut -d',' -f1)"
	@echo "  Docker Compose: $$(docker-compose --version | cut -d' ' -f3 | cut -d',' -f1)"
	@echo "  Go: $$(go version | cut -d' ' -f3)"

# ==========================================
# ENVIRONMENT CONFIGURATION
# ==========================================

# Environment detection
ENV ?= dev
COMPOSE_FILE_DEV = deployments/dev/docker-compose-dev.yml
COMPOSE_FILE_TEST = deployments/test/docker-compose-test.yml
COMPOSE_FILE_PROD = deployments/prod/docker-compose-prod.yml

# Set compose file based on environment
ifeq ($(ENV),dev)
    COMPOSE_FILE = $(COMPOSE_FILE_DEV)
    API_PORT = 8081
    DB_PORT = 5433
    REDIS_PORT = 6380
    JAEGER_PORT = 16687
endif

ifeq ($(ENV),test)
    COMPOSE_FILE = $(COMPOSE_FILE_TEST)
    API_PORT = 8082
    DB_PORT = 5434
    REDIS_PORT = 6381
    JAEGER_PORT = 16688
endif

ifeq ($(ENV),prod)
    COMPOSE_FILE = $(COMPOSE_FILE_PROD)
    API_PORT = 8080
    DB_PORT = 5433
    REDIS_PORT = 6379
    JAEGER_PORT = 16686
endif

# ==========================================
# ENVIRONMENT COMMANDS
# ==========================================

dev: deps-check ## Start development environment
	@echo "🚀 Starting development environment..."
	@$(MAKE) _stop-others ENV=dev
	@docker-compose -f $(COMPOSE_FILE_DEV) up -d
	@$(MAKE) _wait-for-health ENV=dev
	@$(MAKE) _show-env-info ENV=dev

dev-down: ## Stop development environment
	@echo "🛑 Stopping development environment..."
	@docker-compose -f $(COMPOSE_FILE_DEV) down

test: deps-check ## Start test environment
	@echo "🧪 Starting test environment..."
	@$(MAKE) _stop-others ENV=test
	@docker-compose -f $(COMPOSE_FILE_TEST) up -d
	@$(MAKE) _wait-for-health ENV=test
	@$(MAKE) _show-env-info ENV=test

test-down: ## Stop test environment
	@echo "🛑 Stopping test environment..."
	@docker-compose -f $(COMPOSE_FILE_TEST) down

prod: deps-check ## Start production environment
	@echo "🚀 Starting production environment..."
	@$(MAKE) _stop-others ENV=prod
	@docker-compose -f $(COMPOSE_FILE_PROD) up -d
	@$(MAKE) _wait-for-health ENV=prod
	@$(MAKE) _show-env-info ENV=prod

prod-down: ## Stop production environment
	@echo "🛑 Stopping production environment..."
	@docker-compose -f $(COMPOSE_FILE_PROD) down

stop: ## Stop all environments
	@echo "🛑 Stopping all environments..."
	@docker-compose -f $(COMPOSE_FILE_DEV) down 2>/dev/null || true
	@docker-compose -f $(COMPOSE_FILE_TEST) down 2>/dev/null || true
	@docker-compose -f $(COMPOSE_FILE_PROD) down 2>/dev/null || true
	@echo "✅ All environments stopped"

# ==========================================
# BUILD COMMANDS
# ==========================================

build: deps-check ## Build images for specified environment
	@echo "🔨 Building images for $(ENV) environment..."
	@docker-compose -f $(COMPOSE_FILE) build --parallel
	@echo "✅ Build completed for $(ENV)"

build-no-cache: deps-check ## Build images without cache
	@echo "🔨 Building images for $(ENV) environment (no cache)..."
	@docker-compose -f $(COMPOSE_FILE) build --no-cache --parallel
	@echo "✅ Build completed for $(ENV)"

# ==========================================
# SEED COMMANDS
# ==========================================

seed: seed-$(ENV) ## Seed database for current environment

seed-dev: deps-check ## Seed development database
	@echo "🌱 Seeding development database..."
	@$(MAKE) _ensure-running ENV=dev
	@$(MAKE) _run-seed-command ENV=dev

seed-test: deps-check ## Seed test database
	@echo "🌱 Seeding test database..."
	@$(MAKE) _ensure-running ENV=test
	@$(MAKE) _run-seed-command ENV=test

seed-prod: deps-check ## Seed production database
	@echo "🌱 Seeding production database..."
	@$(MAKE) _ensure-running ENV=prod
	@$(MAKE) _run-seed-command ENV=prod

# ==========================================
# DATABASE UTILITY COMMANDS (GORM AutoMigrate & Atlas)
# ==========================================

db-tables: deps-check ## List database tables for current environment
	@echo "📊 Listing database tables for $(ENV) environment..."
	@$(MAKE) _ensure-running ENV=$(ENV)
	@$(MAKE) _list-tables ENV=$(ENV)

db-migrate: deps-check ## Run database migration with GORM and create performance indexes
	@echo "🔧 Running database migration for $(ENV) environment..."
	@$(MAKE) _ensure-running ENV=$(ENV)
	@echo "📝 Building migration binary..."
	@go build -o bin/migrate cmd/migrate/main.go
	@echo "🚀 Running GORM AutoMigrate with performance indexes..."
	@./bin/migrate
	@echo "✅ Database migration completed successfully!"

# ==========================================
# ATLAS MIGRATION COMMANDS
# ==========================================

atlas-check: ## Check if Atlas CLI is installed
	@echo "🔍 Checking Atlas CLI installation..."
	@command -v atlas >/dev/null 2>&1 || { echo "❌ Atlas CLI not found. Installing..."; $(MAKE) atlas-install; }
	@echo "✅ Atlas CLI is available"

atlas-install: ## Install Atlas CLI
	@echo "📦 Installing Atlas CLI..."
	@curl -sSf https://atlasgo.sh | sh
	@echo "✅ Atlas CLI installed successfully"

atlas-status: atlas-check ## Show Atlas migration status
	@echo "📊 Atlas migration status for $(ENV) environment..."
	@atlas migrate status --env $(ENV) 2>/dev/null || atlas migrate status --url "postgres://devuser:devpass@localhost:5433/newsapi_$(ENV)?sslmode=disable" --dir "file://migrations/atlas"

atlas-apply: atlas-check ## Apply Atlas migrations
	@echo "🚀 Applying Atlas migrations for $(ENV) environment..."
	@$(MAKE) _ensure-running ENV=$(ENV)
	@atlas migrate apply --env $(ENV) 2>/dev/null || atlas migrate apply --url "postgres://devuser:devpass@localhost:5433/newsapi_$(ENV)?sslmode=disable" --dir "file://migrations/atlas"
	@echo "✅ Atlas migrations applied successfully"

atlas-apply-docker: ## Apply Atlas migrations using Docker
	@echo "🐳 Applying Atlas migrations for $(ENV) environment using Docker..."
	@$(MAKE) _ensure-running ENV=$(ENV)
	@docker-compose -f deployments/$(ENV)/docker-compose-$(ENV).yml --profile atlas run --rm $(ENV)_atlas migrate apply --env docker-$(ENV)
	@echo "✅ Atlas migrations applied successfully via Docker"

atlas-status-docker: ## Check Atlas migration status using Docker
	@echo "🐳 Checking Atlas migration status for $(ENV) environment using Docker..."
	@$(MAKE) _ensure-running ENV=$(ENV)
	@docker-compose -f deployments/$(ENV)/docker-compose-$(ENV).yml --profile atlas run --rm $(ENV)_atlas migrate status --env docker-$(ENV)

atlas-diff: atlas-check ## Create new Atlas migration from schema changes
	@echo "📝 Creating new Atlas migration for $(ENV) environment..."
	@atlas migrate diff --env $(ENV)
	@echo "✅ New migration created"

atlas-validate: atlas-check ## Validate Atlas schema
	@echo "✅ Validating Atlas schema..."
	@atlas schema validate --env $(ENV)
	@echo "✅ Schema validation completed"

atlas-validate-docker: ## Validate Atlas schema using Docker
	@echo "🐳 Validating Atlas schema for $(ENV) environment using Docker..."
	@$(MAKE) _ensure-running ENV=$(ENV)
	@docker-compose -f deployments/$(ENV)/docker-compose-$(ENV).yml --profile atlas run --rm $(ENV)_atlas schema validate --env docker-$(ENV)

atlas-hash: atlas-check ## Update Atlas migration hash
	@echo "🔄 Updating Atlas migration hash..."
	@atlas migrate hash --dir file://migrations/atlas
	@echo "✅ Migration hash updated"

atlas-dev-setup: atlas-check ## Setup Atlas for development (create baseline)
	@echo "🛠️ Setting up Atlas for development..."
	@$(MAKE) _ensure-running ENV=dev
	@atlas migrate apply --env dev --baseline $(shell atlas migrate status --env dev | head -1 | awk '{print $$1}') || true
	@echo "✅ Atlas development setup completed"

# Switch migration modes
db-mode-auto: ## Switch to GORM AutoMigrate mode
	@echo "🔄 Switching to GORM AutoMigrate mode..."
	@sed -i.bak 's/DB_MIGRATION_MODE=.*/DB_MIGRATION_MODE=auto/' deployments/$(ENV)/.env.$(ENV)
	@echo "✅ Migration mode set to: auto"
	@echo "💡 Restart the application to apply changes"

db-mode-atlas: ## Switch to Atlas migration mode
	@echo "🎯 Switching to Atlas migration mode..."
	@sed -i.bak 's/DB_MIGRATION_MODE=.*/DB_MIGRATION_MODE=atlas/' deployments/$(ENV)/.env.$(ENV)
	@echo "✅ Migration mode set to: atlas"
	@echo "💡 Run 'make atlas-apply' to apply migrations"
	@echo "💡 Restart the application to apply changes"

db-reset: deps-check ## Reset database and let GORM AutoMigrate recreate schema
	@echo "⚠️  Resetting database for $(ENV) environment..."
	@echo "📝 Note: GORM AutoMigrate will recreate the schema when the API starts"
	@read -p "Are you sure you want to reset the $(ENV) database? [y/N] " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		$(MAKE) _ensure-running ENV=$(ENV); \
		$(MAKE) _reset-database ENV=$(ENV); \
		echo "✅ Database reset complete. Restart the API to recreate schema with GORM AutoMigrate."; \
	else \
		echo "❌ Database reset cancelled"; \
	fi

db-status: deps-check ## Show database connection status
	@echo "🔍 Checking database status for $(ENV) environment..."
	@$(MAKE) _ensure-running ENV=$(ENV)
	@$(MAKE) _check-db-connection ENV=$(ENV)

# ==========================================
# TEST COMMANDS
# ==========================================

# Test with our custom test runner (local)
test-local: deps-check ## Run tests locally using custom test runner
	@echo "🧪 Running tests locally with custom test runner..."
	@cd tests && go run main.go all

test-docker: deps-check ## Run tests in Docker containers
	@echo "🐳 Running tests in Docker containers..."
	@$(MAKE) _ensure-running ENV=test
	@echo "🧪 Running tests against Docker services..."
	@cd tests && go run main.go all

# Individual test types - LOCAL
test-unit: deps-check ## Run unit tests locally
	@echo "🧪 Running unit tests locally..."
	@cd tests && go run main.go unit

test-integration: deps-check ## Run integration tests locally
	@echo "🔗 Running integration tests locally..."
	@$(MAKE) _ensure-running ENV=test
	@cd tests && go run main.go integration

test-e2e: deps-check ## Run E2E tests locally
	@echo "🎯 Running E2E tests locally..."
	@$(MAKE) _ensure-running ENV=test
	@cd tests && go run main.go e2e

test-all: deps-check ## Run all tests locally
	@echo "🚀 Running all tests locally..."
	@$(MAKE) _ensure-running ENV=test
	@cd tests && go run main.go all

# Docker-based individual test types
test-unit-docker: deps-check ## Run unit tests in Docker
	@echo "🐳 Running unit tests in Docker..."
	@$(MAKE) _ensure-running ENV=test
	@cd tests && $(MAKE) -f Makefile.docker docker-test-unit

test-integration-docker: deps-check ## Run integration tests in Docker
	@echo "🐳 Running integration tests in Docker..."
	@$(MAKE) _ensure-running ENV=test
	@cd tests && $(MAKE) -f Makefile.docker docker-test-integration

test-e2e-docker: deps-check ## Run E2E tests in Docker
	@echo "🐳 Running E2E tests in Docker..."
	@$(MAKE) _ensure-running ENV=test
	@cd tests && $(MAKE) -f Makefile.docker docker-test-e2e

# Test coverage and utilities
test-coverage: deps-check ## Run tests with coverage report
	@echo "📊 Running tests with coverage..."
	@go test -coverprofile=coverage.out ./tests/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

test-watch: deps-check ## Run tests in watch mode
	@echo "👀 Running tests in watch mode..."
	@if command -v fswatch >/dev/null 2>&1; then \
		echo "Using fswatch for file monitoring..."; \
		while true; do \
			$(MAKE) test-unit; \
			echo "Waiting for changes..."; \
			fswatch -1 internal/ tests/ --exclude='.*\.git.*' | head -1; \
		done; \
	elif command -v inotifywait >/dev/null 2>&1; then \
		echo "Using inotifywait for file monitoring..."; \
		while true; do \
			$(MAKE) test-unit; \
			echo "Waiting for changes..."; \
			inotifywait -r -e modify internal/ tests/ --exclude='.*\.git.*'; \
		done; \
	else \
		echo "❌ File watcher not found. Please install fswatch (macOS) or inotify-tools (Linux)"; \
		echo "macOS: brew install fswatch"; \
		echo "Ubuntu: sudo apt-get install inotify-tools"; \
		exit 1; \
	fi

# Test specific test by name
test-run: deps-check ## Run specific test (usage: make test-run TEST=TestName)
	@if [ -z "$(TEST)" ]; then \
		echo "❌ Usage: make test-run TEST=TestName"; \
		exit 1; \
	fi
	@echo "🧪 Running test: $(TEST)"
	@go test -v -run $(TEST) ./tests/...

# Verbose test output
test-verbose: deps-check ## Run tests with verbose output and no cache
	@echo "🧪 Running tests with verbose output..."
	@go test -v -count=1 ./tests/...

# Test setup
test-setup: deps-check ## Setup test environment and dependencies
	@echo "🛠️  Setting up test environment..."
	@mkdir -p tests/tmp/test_uploads
	@mkdir -p logs
	@touch tests/.env.test
	@echo "✅ Test environment setup complete"

# Clean test artifacts
test-clean: ## Clean test artifacts and temporary files
	@echo "🧹 Cleaning test artifacts..."
	@rm -f coverage.out coverage.html
	@rm -rf tests/tmp/
	@rm -f *.test
	@echo "✅ Test artifacts cleaned"

# ==========================================
# DOCUMENTATION COMMANDS
# ==========================================

docs: ## Generate API documentation using Docker
	@echo "📚 Generating Swagger documentation with Docker..."
	@docker run --rm -v "$(PWD):/app" -w /app golang:1.24-alpine sh -c '\
		apk add --no-cache git && \
		go install github.com/swaggo/swag/cmd/swag@latest && \
		swag init -g cmd/api/main.go -o cmd/api/docs'
	@echo "✅ Swagger documentation generated at: cmd/api/docs/"
	@echo "🌐 Accessible at: http://localhost:$(API_PORT)/swagger/index.html"
	@if [ -f "cmd/api/docs/swagger.json" ]; then \
		echo "✅ swagger.json generated successfully"; \
	else \
		echo "❌ Failed to generate swagger.json"; \
	fi

docs-serve: ## Serve Swagger UI in Docker container
	@echo "🌐 Starting Swagger UI server..."
	@if [ ! -f "cmd/api/docs/swagger.json" ]; then \
		echo "⚠️  No swagger.json found. Generating documentation first..."; \
		$(MAKE) docs; \
	fi
	@docker run --rm -d \
		--name swagger-ui \
		-p 8091:8080 \
		-v "$(PWD)/cmd/api/docs:/usr/share/nginx/html/docs" \
		-e SWAGGER_JSON=/docs/swagger.json \
		swaggerapi/swagger-ui
	@echo "✅ Swagger UI available at: http://localhost:8091"
	@echo "💡 To stop: docker stop swagger-ui"

docs-stop: ## Stop Swagger UI container
	@echo "🛑 Stopping Swagger UI server..."
	@docker stop swagger-ui 2>/dev/null || echo "Swagger UI container not running"
	@echo "✅ Swagger UI stopped"

docs-local: ## Generate API documentation locally (requires swag CLI)
	@echo "📚 Generating Swagger documentation locally..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api/main.go -o cmd/api/docs && \
		echo "✅ Swagger documentation generated at: cmd/api/docs/"; \
	else \
		echo "❌ swag CLI not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
		echo "💡 Or use: make docs (Docker-based)"; \
	fi

# ==========================================
# UTILITY COMMANDS
# ==========================================

logs: ## Show logs for current environment
	@echo "📋 Showing logs for $(ENV) environment..."
	@docker-compose -f $(COMPOSE_FILE) logs -f

status: ## Show environment status
	@echo "📊 Environment Status ($(ENV)):"
	@echo "=============================="
	@docker-compose -f $(COMPOSE_FILE) ps

health: ## Check environment health
	@$(MAKE) _check-health ENV=$(ENV)

clean: ## Clean Docker resources
	@echo "🧹 Cleaning Docker resources..."
	@docker system prune -f
	@docker volume prune -f
	@echo "✅ Docker cleanup completed"

test-status: ## Show test environment status
	@echo "📊 Test Environment Status:"
	@echo "=========================="
	@if docker ps | grep -q "news_test_api"; then \
		echo "✅ Test API Container: Running"; \
	else \
		echo "❌ Test API Container: Not running"; \
	fi
	@if docker ps | grep -q "news_test_db"; then \
		echo "✅ Test Database Container: Running"; \
	else \
		echo "❌ Test Database Container: Not running"; \
	fi
	@if docker ps | grep -q "news_test_redis"; then \
		echo "✅ Test Redis Container: Running"; \
	else \
		echo "❌ Test Redis Container: Not running"; \
	fi
	@echo ""
	@echo "🔗 Test URLs:"
	@echo "  API: http://localhost:8082"
	@echo "  Database: localhost:5434"
	@echo "  Redis: localhost:6381"
	@echo "  Jaeger: http://localhost:16688"

# ==========================================
# INTERNAL HELPER FUNCTIONS
# ==========================================

_stop-others: ## Stop other environments (internal)
	@if [ "$(ENV)" != "dev" ]; then docker-compose -f $(COMPOSE_FILE_DEV) down 2>/dev/null || true; fi
	@if [ "$(ENV)" != "test" ]; then docker-compose -f $(COMPOSE_FILE_TEST) down 2>/dev/null || true; fi
	@if [ "$(ENV)" != "prod" ]; then docker-compose -f $(COMPOSE_FILE_PROD) down 2>/dev/null || true; fi

_wait-for-health: ## Wait for environment to be healthy (internal)
	@echo "⏳ Waiting for $(ENV) environment to be ready..."
	@sleep 10
	@$(MAKE) _check-health ENV=$(ENV) || true

_show-env-info: ## Show environment information (internal)
	@echo ""
	@echo "✅ $(ENV) environment is running!"
	@echo "  📊 API: http://localhost:$(API_PORT)"
	@echo "  🗄️  Database: localhost:$(DB_PORT)"
	@echo "  📦 Redis: localhost:$(REDIS_PORT)"
	@echo "  🔍 Jaeger: http://localhost:$(JAEGER_PORT)"

_ensure-running: ## Ensure environment is running (internal)
	@if ! docker-compose -f $(COMPOSE_FILE) ps | grep -q "Up"; then \
		echo "❌ $(ENV) environment is not running"; \
		echo "💡 Start it with: make $(ENV)"; \
		exit 1; \
	fi

_get-api-service: ## Get API service name for environment (internal)
	@if [ "$(ENV)" = "dev" ]; then echo "dev_api"; \
	elif [ "$(ENV)" = "test" ]; then echo "test_api"; \
	elif [ "$(ENV)" = "prod" ]; then echo "prod_api"; \
	else echo "api"; fi

_get-db-service: ## Get database service name for environment (internal)
	@if [ "$(ENV)" = "dev" ]; then echo "dev_db"; \
	elif [ "$(ENV)" = "test" ]; then echo "test_db"; \
	elif [ "$(ENV)" = "prod" ]; then echo "prod_db"; \
	else echo "db"; fi

_run-seed-command: ## Run seed command (internal)
	@echo "🌱 Starting seed process for $(ENV) environment..."
	@ENVIRONMENT=$(ENV) ./scripts/seeds/seed_database.sh

# Database utility commands (replacing migration helpers)
_list-tables: ## List database tables (internal)
	@echo "📊 Database tables in $(ENV) environment:"
	@$(MAKE) _ensure-running ENV=$(ENV)
	@DB_SERVICE=$$($(MAKE) _get-db-service ENV=$(ENV)); \
	if [ "$(ENV)" = "dev" ]; then \
		docker exec news_$${DB_SERVICE} psql -U devuser -d newsapi_dev -c "\dt" | grep -E "^ public" | awk '{print $$3}' | sort; \
	elif [ "$(ENV)" = "test" ]; then \
		docker exec news_$${DB_SERVICE} psql -U testuser -d newsapi_test -c "\dt" | grep -E "^ public" | awk '{print $$3}' | sort; \
	elif [ "$(ENV)" = "prod" ]; then \
		docker exec news_$${DB_SERVICE} psql -U produser -d newsdb_prod -c "\dt" | grep -E "^ public" | awk '{print $$3}' | sort; \
	fi

_reset-database: ## Reset database (internal) - Note: GORM AutoMigrate will recreate schema
	@DB_SERVICE=$$($(MAKE) _get-db-service ENV=$(ENV)); \
	if [ "$(ENV)" = "dev" ]; then \
		docker exec news_$${DB_SERVICE} psql -U devuser -d newsapi_dev -c "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO public;"; \
	elif [ "$(ENV)" = "test" ]; then \
		docker exec news_$${DB_SERVICE} psql -U testuser -d newsapi_test -c "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO public;"; \
	elif [ "$(ENV)" = "prod" ]; then \
		docker exec news_$${DB_SERVICE} psql -U produser -d newsdb_prod -c "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO public;"; \
	fi

_check-db-connection: ## Check database connection status (internal)
	@DB_SERVICE=$$($(MAKE) _get-db-service ENV=$(ENV)); \
	if [ "$(ENV)" = "dev" ]; then \
		if docker exec news_$${DB_SERVICE} pg_isready -U devuser -d newsapi_dev 2>/dev/null; then \
			echo "  ✅ Database connection: OK"; \
			echo "  📊 Database info:"; \
			docker exec news_$${DB_SERVICE} psql -U devuser -d newsapi_dev -c "SELECT version();" | head -3; \
			echo "  📋 Total tables: $$(docker exec news_$${DB_SERVICE} psql -U devuser -d newsapi_dev -c '\dt' 2>/dev/null | grep -c '^ public' || echo '0')"; \
		else \
			echo "  ❌ Database connection: FAILED"; \
		fi; \
	elif [ "$(ENV)" = "test" ]; then \
		if docker exec news_$${DB_SERVICE} pg_isready -U testuser -d newsapi_test 2>/dev/null; then \
			echo "  ✅ Database connection: OK"; \
			echo "  📊 Database info:"; \
			docker exec news_$${DB_SERVICE} psql -U testuser -d newsapi_test -c "SELECT version();" | head -3; \
			echo "  📋 Total tables: $$(docker exec news_$${DB_SERVICE} psql -U testuser -d newsapi_test -c '\dt' 2>/dev/null | grep -c '^ public' || echo '0')"; \
		else \
			echo "  ❌ Database connection: FAILED"; \
		fi; \
	elif [ "$(ENV)" = "prod" ]; then \
		if docker exec news_$${DB_SERVICE} pg_isready -U produser -d newsdb_prod 2>/dev/null; then \
			echo "  ✅ Database connection: OK"; \
			echo "  📊 Database info:"; \
			docker exec news_$${DB_SERVICE} psql -U produser -d newsdb_prod -c "SELECT version();" | head -3; \
			echo "  📋 Total tables: $$(docker exec news_$${DB_SERVICE} psql -U produser -d newsdb_prod -c '\dt' 2>/dev/null | grep -c '^ public' || echo '0')"; \
		else \
			echo "  ❌ Database connection: FAILED"; \
		fi; \
	fi

_check-health: ## Check environment health (internal)
	@echo "🔍 Health check for $(ENV) environment:"
	@API_URL="http://localhost:$(API_PORT)"; \
	if command -v curl >/dev/null 2>&1; then \
		if curl -s "$$API_URL/health" >/dev/null 2>&1; then \
			echo "  ✅ API is healthy"; \
		else \
			echo "  ❌ API is not responding"; \
		fi; \
	else \
		echo "  ⚠️  curl not available, skipping API health check"; \
	fi
	@DB_SERVICE=$$($(MAKE) _get-db-service ENV=$(ENV)); \
	if [ "$(ENV)" = "dev" ]; then \
		if docker exec news_$${DB_SERVICE} pg_isready -U devuser -d newsapi_dev 2>/dev/null; then \
			echo "  ✅ Database is healthy"; \
		else \
			echo "  ❌ Database is not responding"; \
		fi; \
	elif [ "$(ENV)" = "test" ]; then \
		if docker exec news_$${DB_SERVICE} pg_isready -U testuser -d newsapi_test 2>/dev/null; then \
			echo "  ✅ Database is healthy"; \
		else \
			echo "  ❌ Database is not responding"; \
		fi; \
	elif [ "$(ENV)" = "prod" ]; then \
		if docker exec news_$${DB_SERVICE} pg_isready -U produser -d newsdb_prod 2>/dev/null; then \
			echo "  ✅ Database is healthy"; \
		else \
			echo "  ❌ Database is not responding"; \
		fi; \
	fi

# ==========================================
# HTTP/2 COMMANDS
# ==========================================

http2-dev: ## Start HTTP/2 development server with H2C
	@echo "🌐 Starting HTTP/2 Development Server (H2C)"
	@echo "==========================================="
	@./scripts/start-dev-http2.sh

http2-prod: ## Start HTTP/2 production server with HTTPS
	@echo "🌐 Starting HTTP/2 Production Server (HTTPS)"
	@echo "============================================"
	@./scripts/start-prod-http2.sh

http2-test: ## Test HTTP/2 connectivity
	@echo "🧪 Testing HTTP/2 Connectivity"
	@echo "=============================="
	@if [ -z "$(URL)" ]; then \
		echo "Usage: make http2-test URL=http://localhost:8080"; \
		echo "Examples:"; \
		echo "  make http2-test URL=http://localhost:8080   # Test H2C"; \
		echo "  make http2-test URL=https://localhost:8443  # Test HTTPS/2"; \
	else \
		go run ./cmd/http2-test/main.go $(URL); \
	fi

http2-prod-test: ## Test HTTP/2 production deployment with load testing
	@echo "🧪 Testing HTTP/2 Production Deployment"
	@echo "======================================="
	@if [ -z "$(URL)" ]; then \
		echo "Usage: make http2-prod-test URL=https://localhost:8443"; \
		echo "Testing with default URL: https://localhost:8443"; \
		./scripts/test-http2-production.sh; \
	else \
		./scripts/test-http2-production.sh $(URL); \
	fi

http2-certs: ## Generate self-signed certificates for HTTPS/2
	@echo "🔐 Generating HTTP/2 Compatible TLS Certificates"
	@echo "=============================================="
	@./scripts/generate-http2-certs.sh

http2-prod-certs: ## Generate production-grade TLS certificates for HTTPS/2
	@echo "🔐 Generating Production TLS Certificates for HTTP/2"
	@echo "=================================================="
	@./scripts/generate-prod-certs.sh

http2-benchmark: ## Run HTTP/2 vs HTTP/1.1 performance benchmark
	@echo "📊 Running HTTP/2 vs HTTP/1.1 Benchmark"
	@echo "======================================="
	@if [ -z "$(URL)" ]; then \
		echo "Usage: make http2-benchmark URL=http://localhost:8080"; \
	else \
		echo "Starting benchmark against $(URL)..."; \
		go run ./cmd/http2-test/main.go $(URL); \
		echo ""; \
		echo "🚀 For more detailed benchmarks, use tools like:"; \
		echo "  • h2load: h2load -n1000 -c10 $(URL)"; \
		echo "  • wrk: wrk -t12 -c400 -d30s $(URL)"; \
	fi

http2-status: ## Check HTTP/2 server status and protocol support
	@echo "📊 HTTP/2 Server Status"
	@echo "======================"
	@echo "Development Server (H2C):"
	@curl -s -I http://localhost:8080/health 2>/dev/null | head -1 || echo "  ❌ Not running"
	@echo ""
	@echo "Production Server (HTTPS):"
	@curl -s -I -k https://localhost:8443/health 2>/dev/null | head -1 || echo "  ❌ Not running"
	@echo ""
	@echo "Protocol Support Check:"
	@echo "  • curl HTTP/2 support: $$(curl --version | grep -q HTTP2 && echo '✅ Available' || echo '❌ Not available')"
	@echo "  • OpenSSL version: $$(openssl version | cut -d' ' -f2)"

.PHONY: http2-dev http2-prod http2-test http2-certs http2-benchmark http2-status

# ==========================================
# ATLAS TEST & DEVELOPMENT COMMANDS
# ==========================================

atlas-test-dev: ## Test Atlas in development environment
	@echo "🧪 Testing Atlas in development environment..."
	@$(MAKE) _ensure-running ENV=dev
	@echo "📊 Current migration status:"
	@$(MAKE) atlas-status-docker ENV=dev || true
	@echo "🚀 Applying migrations:"
	@$(MAKE) atlas-apply-docker ENV=dev || true
	@echo "📊 Final migration status:"
	@$(MAKE) atlas-status-docker ENV=dev || true
	@echo "✅ Atlas test completed!"

atlas-demo: ## Demo Atlas migration workflow
	@echo "🎬 Atlas Migration Demo"
	@echo "======================"
	@echo "This will demonstrate the Atlas migration workflow:"
	@echo "1. Check current status"
	@echo "2. Apply migrations"
	@echo "3. Verify results"
	@echo ""
	@read -p "Continue with demo? [y/N] " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		$(MAKE) atlas-test-dev ENV=dev; \
	else \
		echo "❌ Demo cancelled"; \
	fi

# ==========================================
# AUTOMATIC ATLAS MIGRATION COMMANDS
# ==========================================

auto-migrate: ## Automatic GORM → Atlas migration (Docker-based)
	@echo "🤖 Starting automatic GORM → Atlas migration (Docker-based)..."
	@$(MAKE) _ensure-running ENV=$(ENV)
	@ENV=$(ENV) ./scripts/auto-atlas-migration.sh

auto-migrate-commit: ## Automatic migration + git commit (Docker-based)
	@echo "🤖 Automatic migration + commit (Docker-based)..."
	@$(MAKE) _ensure-running ENV=$(ENV)
	@ENV=$(ENV) ./scripts/auto-atlas-migration.sh --commit

atlas-sync: ## Sync GORM models to Atlas (Docker-based)
	@echo "🔄 Syncing GORM models to Atlas (Docker-based)..."
	@$(MAKE) _ensure-running ENV=$(ENV)
	@ENV=$(ENV) ./scripts/auto-atlas-migration.sh
	@echo "💡 Migration created. Apply with 'make atlas-apply-docker ENV=$(ENV)'"

atlas-workflow-test: ## Test GitHub workflow locally (Docker-based)
	@echo "🧪 Testing GitHub workflow locally (Docker-based)..."
	@echo "1. Testing schema changes..."
	@$(MAKE) atlas-validate-docker ENV=$(ENV)

# ==========================================
# VERSIONING
# ==========================================

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")

version: ## Show current version
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

build-with-version: ## Build with version information
	@echo "🔨 Building with version info..."
	go build -ldflags "-X news/internal/version.Version=$(VERSION) -X news/internal/version.BuildTime=$(BUILD_TIME) -X news/internal/version.GitCommit=$(GIT_COMMIT)" -o bin/news-api cmd/api/main.go

release-tag: ## Create a new release tag (usage: make release-tag VERSION=v1.1.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION is required. Usage: make release-tag VERSION=v1.1.0"; \
		exit 1; \
	fi
	@echo "🏷️  Creating release tag $(VERSION)"
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "✅ Tag $(VERSION) created and pushed"

changelog-update: ## Update changelog for new version
	@echo "📝 Please update CHANGELOG.md with new version changes"
	@echo "   Add new section for version $(VERSION)"
