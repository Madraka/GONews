# News API Test Suite

A comprehensive test suite built with the modern testify framework. Supports both local and Docker environments.

## 📁 Test Structure

```
tests/
├── unit/                 # Unit tests (fast, isolated)
│   ├── user_test.go     # User model tests
│   ├── article_test.go  # Article model tests
│   └── auth_test.go     # Authentication tests
├── integration/         # Integration tests (API endpoints)
│   └── api_test.go      # API integration tests
├── e2e/                 # End-to-end tests (complete user workflows)
│   └── e2e_test.go      # Complete user journey tests
├── testutil/            # Test utility functions
│   ├── database.go      # Database test helpers
│   ├── http.go          # HTTP test helpers
│   ├── fixtures.go      # Test data generators
│   └── fixtures_loader.go # Test data loader
├── mocks/               # Mock implementations
│   └── mocks.go         # Service mocks
├── fixtures/            # Test data
│   └── test_data.json   # Static test data
├── main.go              # Custom test runner
├── .env.test            # Docker test configuration
├── .env.test.local      # Local test configuration
├── Makefile.docker      # Docker test commands
└── README.md            # This file
```

## 🚀 Quick Start

### Test Environment Options

**Local Testing (Recommended for Development):**
- ✅ Fast execution
- ✅ Uses Docker services for dependencies
- ✅ Easy debugging
- 🔧 **Config:** Uses `.env.test.local`
- 🗄️ **Database:** Docker PostgreSQL (port 5434)
- 🔴 **Redis:** Docker Redis (port 6381)

**Docker Testing (For CI/CD):**
- ✅ Completely isolated
- ✅ Consistent across environments
- ✅ Production-like conditions
- 🔧 **Config:** Uses `.env.test`
- 🗄️ **Database:** Containerized PostgreSQL
- 🔴 **Redis:** Containerized Redis

### Quick Test Commands

```bash
# Run from main directory (News/):

# Run all tests locally (fastest)
make test-local

# Run all tests in Docker containers
make test-docker

# Individual test types
make test-unit          # Unit tests only
make test-integration   # Integration tests only  
make test-e2e          # E2E tests only

# Test utilities
make test-coverage     # Run with coverage report
make test-verbose      # Run with detailed output
make test-watch        # Watch mode (for development)

# Test environment management
make test-status       # Check test container status
make test-setup        # Setup test environment
make test-clean        # Clean test environment
```

3. Run tests:
```bash
make test
```

## 🧪 Test Types

### Unit Tests
Isolated tests for individual components.

```bash
# Run unit tests
make test-unit

# Or directly with Go
go test -v ./tests/unit/...
```

### Integration Tests
Tests for API endpoints and database interactions.

```bash
# Run integration tests
make test-integration

# Or directly with Go
go test -v ./tests/integration/...
```

### End-to-End Tests
Tests for complete user scenarios.

```bash
# Run E2E tests
make test-e2e

# Or directly with Go
go test -v ./tests/e2e/...
```

## 📊 Test Commands

### Basic Commands

```bash
make test              # Run unit tests (default)
make test-all          # Run all tests
make test-coverage     # Run tests with coverage report
make test-race         # Run tests with race detection
make test-bench        # Run benchmark tests
```

### Advanced Commands

```bash
make test-verbose      # Run tests with detailed output
make test-watch        # Run tests in watch mode
make test-run TEST=TestName  # Run specific test
make test-clean        # Clean test artifacts
```

### Database Commands

```bash
make test-db-setup     # Setup test database
make test-db-clean     # Clean test database
make test-full         # Full test cycle (setup + test + cleanup)
```

## 🔧 Test Configuration

### Environment Variables

Test configuration is defined in the `.env.test` file:

```env
# Database
TEST_DB_HOST=localhost
TEST_DB_PORT=5432
TEST_DB_USER=postgres
TEST_DB_PASSWORD=postgres
TEST_DB_NAME=news_test

# API
TEST_API_BASE_URL=http://localhost:8081
TEST_API_TIMEOUT=30s

# JWT
TEST_JWT_SECRET=test-secret-key
TEST_JWT_EXPIRY=1h
```

### Test Data

Test data is stored in JSON format in `fixtures/test_data.json`:

- Test users
- Test articles
- Test categories
- API endpoints
- Error scenarios
- Performance benchmarks

## 📝 Test Writing Guide

### Unit Test Example

```go
package unit

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "news/tests/testutil"
)

type UserTestSuite struct {
    suite.Suite
    testData *testutil.TestData
}

func (suite *UserTestSuite) SetupSuite() {
    suite.testData = testutil.NewTestData()
}

func (suite *UserTestSuite) TestUser_Creation() {
    user := suite.testData.CreateTestUser()
    assert.NotEmpty(suite.T(), user.Username)
    assert.NotEmpty(suite.T(), user.Email)
}

func TestUserTestSuite(t *testing.T) {
    suite.Run(t, new(UserTestSuite))
}
```

### Integration Test Example

```go
package integration

import (
    "testing"
    "net/http"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "news/tests/testutil"
)

type APITestSuite struct {
    suite.Suite
    server   *testutil.TestServer
    testDB   *testutil.TestDB
}

func (suite *APITestSuite) SetupSuite() {
    suite.testDB = testutil.SetupTestDB(suite.T())
    // router := setupRouter(suite.testDB.DB)
    // suite.server = testutil.NewTestServer(router)
}

func (suite *APITestSuite) TestAPI_Health() {
    resp := suite.server.GET(suite.T(), "/health")
    defer resp.Body.Close()
    assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func TestAPITestSuite(t *testing.T) {
    suite.Run(t, new(APITestSuite))
}
```

## 🎯 Mock Usage

Mocks are created using testify/mock:

```go
// Create mock
mockUserRepo := new(mocks.MockUserRepository)

// Define expectations
mockUserRepo.On("GetByID", 1).Return(user, nil)

// Use mock
result, err := mockUserRepo.GetByID(1)

// Verification
mockUserRepo.AssertExpectations(t)
```

## 📈 Coverage Report

To generate a coverage report:

```bash
make test-coverage
```

This command creates a `coverage.html` file that you can open in your browser.

## 🔍 Test Best Practices

### 1. Use Test Suites
- Create separate test suites for each module
- Use SetupSuite/TearDownSuite hooks
- Share test data within the suite

### 2. Test Data Management
- Store test data in fixtures
- Ensure clean database state for each test
- Use TestData helpers

### 3. Mock Usage
- Use mocks for external dependencies
- Verify mock expectations
- Use real dependencies in integration tests

### 4. Test Naming
- Use descriptive test names
- Specify the test scenario in the name
- Prefer BDD format (Given_When_Then)

### 5. Assertions
- Use testify/assert and testify/require
- Write meaningful error messages
- Explicitly check the values to be tested

## 🚨 Troubleshooting

### Common Issues

1. **Database Connection Error**
   ```bash
   # Make sure PostgreSQL is running
   brew services start postgresql
   
   # Create test database
   createdb news_test
   ```

2. **Environment Variables**
   ```bash
   # Make sure .env.test file exists
   cp .env.test.example .env.test
   ```

3. **Port Conflicts**
   ```bash
   # Make sure test API uses a different port
   export TEST_API_PORT=8082
   ```

## 🤝 Contributing

1. When adding new tests, add them to the relevant category (unit/integration/e2e)
2. Keep test data fixtures up to date
3. Create mocks according to interfaces
4. Update test documentation

## 📚 References

- [Testify Documentation](https://github.com/stretchr/testify)
- [Go Testing Package](https://golang.org/pkg/testing/)
- [Test Driven Development](https://en.wikipedia.org/wiki/Test-driven_development)
- [Integration Testing Best Practices](https://martinfowler.com/articles/integration-tests.html)

---

**Note**: This test suite is specifically designed for the News API project. Updates can be made according to project requirements.
