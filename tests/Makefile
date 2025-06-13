# Test Runner Makefile
# This Makefile provides convenient commands for running tests

.PHONY: test test-unit test-integration test-e2e test-all test-coverage test-race test-bench test-clean test-setup

# Default test command
test: test-unit

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	@cd .. && go test -v ./tests/unit/...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	@cd .. && go test -v ./tests/integration/...

# Run e2e tests
test-e2e:
	@echo "Running E2E tests..."
	@cd .. && go test -v ./tests/e2e/...

# Run all tests
test-all:
	@echo "Running all tests..."
	@cd .. && go test -v ./tests/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@cd .. && go test -coverprofile=coverage.out ./tests/...
	@cd .. && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	@cd .. && go test -race -v ./tests/...

# Run benchmarks
test-bench:
	@echo "Running benchmarks..."
	@cd .. && go test -bench=. -v ./tests/...

# Clean test artifacts
test-clean:
	@echo "Cleaning test artifacts..."
	@cd .. && rm -f coverage.out coverage.html
	@rm -rf tmp/

# Setup test environment
test-setup:
	@echo "Setting up test environment..."
	@mkdir -p tmp/test_uploads
	@mkdir -p ../logs
	@echo "Test environment setup complete"

# Run tests in watch mode (requires fswatch or inotifywait)
test-watch:
	@echo "Running tests in watch mode..."
	@while true; do \
		make test-unit; \
		echo "Waiting for changes..."; \
		fswatch -1 ../internal/ ../tests/; \
	done

# Run specific test
test-run:
	@if [ -z "$(TEST)" ]; then \
		echo "Usage: make test-run TEST=TestName"; \
		exit 1; \
	fi
	@echo "Running test: $(TEST)"
	@cd .. && go test -v -run $(TEST) ./tests/...

# Run tests with verbose output and no cache
test-verbose:
	@echo "Running tests with verbose output..."
	@cd .. && go test -v -count=1 ./tests/...

# Generate test mocks (if using mockgen)
test-mocks:
	@echo "Generating test mocks..."
	@echo "Note: Install mockgen first: go install github.com/golang/mock/mockgen@latest"
	# Add mockgen commands here as needed

# Database setup for testing
test-db-setup:
	@echo "Setting up test database..."
	@createdb news_test || echo "Database might already exist"
	@cd .. && go run cmd/migrate/main.go -database=$(TEST_DB_DSN) -path=./migrations up

# Database cleanup for testing
test-db-clean:
	@echo "Cleaning test database..."
	@dropdb news_test || echo "Database might not exist"

# Full test cycle (setup, run, cleanup)
test-full:
	@make test-setup
	@make test-db-setup
	@make test-all
	@make test-db-clean

# Help target
help:
	@echo "Available test commands:"
	@echo "  test              - Run unit tests (default)"
	@echo "  test-unit         - Run unit tests"
	@echo "  test-integration  - Run integration tests"
	@echo "  test-e2e          - Run end-to-end tests"
	@echo "  test-all          - Run all tests"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  test-race         - Run tests with race detection"
	@echo "  test-bench        - Run benchmarks"
	@echo "  test-clean        - Clean test artifacts"
	@echo "  test-setup        - Setup test environment"
	@echo "  test-watch        - Run tests in watch mode"
	@echo "  test-run TEST=X   - Run specific test"
	@echo "  test-verbose      - Run tests with verbose output"
	@echo "  test-db-setup     - Setup test database"
	@echo "  test-db-clean     - Clean test database"
	@echo "  test-full         - Full test cycle"
	@echo "  help              - Show this help"
