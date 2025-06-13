# Contributing to News API

First off, thanks for taking the time to contribute! 🎉

The following is a set of guidelines for contributing to News API. These are mostly guidelines, not rules. Use your best judgment, and feel free to propose changes to this document in a pull request.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Pull Request Process](#pull-request-process)
- [Style Guides](#style-guides)

## Code of Conduct

This project and everyone participating in it is governed by our commitment to creating a welcoming and inclusive environment. Please be respectful and professional in all interactions.

## Getting Started

### Prerequisites

- Go 1.24 or higher
- PostgreSQL 15+
- Redis 7+
- Docker (optional but recommended)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/Madraka/GONews.git
   cd GONews
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   cp deployments/dev/env.dev.example deployments/dev/env.dev
   # Edit the file with your configuration
   ```

4. **Run database migrations**
   ```bash
   make migrate-up
   ```

5. **Seed the database**
   ```bash
   make seed
   ```

6. **Start the development server**
   ```bash
   make dev
   ```

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues as you might find out that the problem has already been reported. When you are creating a bug report, please include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples to demonstrate the steps**
- **Describe the behavior you observed and what behavior you expected**
- **Include screenshots if applicable**
- **Include your environment details** (OS, Go version, etc.)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

- **Use a clear and descriptive title**
- **Provide a step-by-step description of the suggested enhancement**
- **Provide specific examples to demonstrate the enhancement**
- **Describe the current behavior and explain the expected behavior**
- **Explain why this enhancement would be useful**

### Pull Requests

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Make your changes**
4. **Add tests for your changes**
5. **Run the test suite** (`make test`)
6. **Run linting** (`make lint`)
7. **Commit your changes** (`git commit -m 'Add amazing feature'`)
8. **Push to the branch** (`git push origin feature/amazing-feature`)
9. **Open a Pull Request**

## Development Setup

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test package
go test ./internal/handlers/...
```

### Running Lints

```bash
# Run all linters
make lint

# Run specific linter
golangci-lint run
```

### Database Migrations

```bash
# Create new migration
make migration name=add_new_table

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down
```

## Style Guides

### Git Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

### Go Style Guide

- Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Write meaningful comments for exported functions and types
- Keep functions small and focused
- Use meaningful variable and function names

### API Design

- Follow RESTful principles
- Use consistent naming conventions
- Include proper HTTP status codes
- Provide comprehensive error messages
- Document all endpoints with OpenAPI/Swagger

## Project Structure

```
News/
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── handlers/          # HTTP handlers
│   ├── services/          # Business logic
│   ├── models/            # Data models
│   ├── repositories/      # Data access layer
│   └── middleware/        # HTTP middleware
├── tests/                 # Test files
├── docs/                  # Documentation
├── deployments/           # Deployment configurations
└── migrations/            # Database migrations
```

## Questions?

Feel free to open an issue with the question label, or reach out to the maintainers directly.

Thank you for contributing! 🚀
