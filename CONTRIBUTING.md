# 🤝 Contributing to GONews

First off, **thank you** for considering contributing to GONews! 🎉 

As an **open source project**, we welcome contributions from developers around the world. Whether you're fixing a bug, adding a feature, improving documentation, or helping with translations, every contribution makes a difference.

**GitHub Repository**: https://github.com/Madraka/GONews

## 📋 Table of Contents

- [🌟 Quick Start for Contributors](#-quick-start-for-contributors)
- [🤝 How Can I Contribute?](#-how-can-i-contribute)
- [🛠️ Development Setup](#️-development-setup)
- [📝 Pull Request Process](#-pull-request-process)
- [📏 Code Style Guidelines](#-code-style-guidelines)
- [🧪 Testing Guidelines](#-testing-guidelines)
- [📚 Documentation Guidelines](#-documentation-guidelines)

## 🌟 Quick Start for Contributors

### New to Open Source?
- Check our [Good First Issues](https://github.com/Madraka/GONews/labels/good%20first%20issue)
- Read our [Developer Guide](./docs/DEVELOPER_GUIDE.md)
- Join our [Discussions](https://github.com/Madraka/GONews/discussions)

### Experienced Contributor?
- Look at [Help Wanted Issues](https://github.com/Madraka/GONews/labels/help%20wanted)
- Check our [Roadmap](./README.md#-roadmap)
- Review our [Architecture Guide](./docs/README.md#architecture--implementation)

## 🤝 How Can I Contribute?

### 🐛 Bug Reports
Found a bug? We want to know about it!

**Before submitting:**
- Check if the bug has already been reported
- Try to reproduce it with the latest version
- Gather system information (OS, Go version, etc.)

**When reporting:**
- Use a clear, descriptive title
- Describe steps to reproduce
- Include error messages and logs
- Mention your environment details

### ✨ Feature Requests
Have an idea for a new feature?

**Before suggesting:**
- Check if it's already been suggested
- Consider if it fits the project's scope
- Think about backward compatibility

**When suggesting:**
- Explain the problem you're trying to solve
- Describe your proposed solution
- Consider alternative approaches
- Think about implementation complexity

### 🔧 Code Contributions
Ready to write some code?

**Types We Welcome:**
- 🐛 **Bug fixes** - Always appreciated!
- ✨ **New features** - Discuss in an issue first
- ⚡ **Performance improvements** - Benchmarks welcome
- 🧪 **Test coverage** - Help us reach 90%+
- 📚 **Documentation** - Code comments, guides, examples
- 🌍 **Internationalization** - New language support

### 📚 Documentation Contributions
Documentation is as important as code!

**Areas needing help:**
- API documentation improvements
- Tutorial creation
- Code examples
- Translation to other languages
- Developer guides and best practices

## 🛠️ Development Setup

### Prerequisites
- **Go 1.24+** - [Download](https://golang.org/dl/)
- **PostgreSQL 15+** - [Download](https://postgresql.org/download/)
- **Redis 7+** - [Download](https://redis.io/download/)
- **Docker** (recommended) - [Download](https://docker.com/get-started/)
- **Make** - Usually pre-installed

### Quick Setup
```bash
# 1. Fork and clone
git clone https://github.com/YOUR-USERNAME/GONews.git
cd GONews

# 2. Add upstream remote
git remote add upstream https://github.com/Madraka/GONews.git

# 3. Install dependencies
go mod download

# 4. Start development environment
make dev-setup
make dev

# 5. Verify setup
make test
```

### Development Workflow
```bash
# Daily workflow
make dev-up        # Start services
make dev           # Start API with hot reload
make test-watch    # Run tests on file changes

# Before committing
make lint          # Check code style
make test          # Run all tests
make docs          # Update documentation
```

## 📝 Pull Request Process

### 1. Preparation
- **Create an issue first** (unless it's a tiny fix)
- **Discuss the approach** with maintainers
- **Check for related work** - avoid duplicate efforts

### 2. Development
```bash
# Create feature branch
git checkout -b feature/amazing-feature

# Make your changes
# Write tests
# Update documentation

# Test everything
make test-all
make build-all
```

### 3. Before Submitting
**Quality Checklist:**
- [ ] ✅ All tests pass
- [ ] ✅ Code follows style guidelines
- [ ] ✅ New features have tests
- [ ] ✅ Documentation updated
- [ ] ✅ No breaking changes (or clearly documented)
- [ ] ✅ Commit messages are clear

**Performance Checklist:**
- [ ] ⚡ No unnecessary database queries
- [ ] ⚡ Efficient algorithms used
- [ ] ⚡ Memory usage considered
- [ ] ⚡ Benchmarks included for performance changes

### 4. Submission
```bash
# Push your branch
git push origin feature/amazing-feature

# Open pull request with:
# - Clear title and description
# - Reference to related issue
# - Screenshots if UI changes
# - Breaking changes highlighted
```

### 5. Review Process
- **Automated checks** run first
- **Code review** by maintainers
- **Address feedback** promptly
- **Final approval** and merge

## 📏 Code Style Guidelines

### Go Code Style
```go
// ✅ Good: Clear naming and structure
func GetArticlesByCategory(ctx context.Context, categoryID uint) ([]models.Article, error) {
    // Implementation
}

// ❌ Bad: Unclear naming
func GetArtsByCat(id uint) []models.Article {
    // Implementation
}
```

**Rules:**
- Follow standard Go conventions (`gofmt`, `golint`)
- Use meaningful variable names
- Write descriptive comments for public functions
- Keep functions small and focused
- Handle errors appropriately

### API Design
```json
// ✅ Good: Consistent structure
{
  "data": [...],
  "meta": {
    "total": 100,
    "page": 1,
    "limit": 20
  }
}

// ❌ Bad: Inconsistent structure  
{
  "articles": [...],
  "count": 100
}
```

### Database
- Use meaningful table and column names
- Include proper indexes
- Write migration scripts
- Document schema changes

## 🧪 Testing Guidelines

### Test Types
```bash
# Unit tests (fast, isolated)
make test-unit

# Integration tests (slower, with database)
make test-integration  

# End-to-end tests (full stack)
make test-e2e

# All tests
make test-all
```

### Writing Tests
```go
func TestArticleService_Create(t *testing.T) {
    // Setup
    service := setupTestArticleService(t)
    
    // Test cases
    tests := []struct {
        name    string
        input   CreateArticleRequest
        want    *Article
        wantErr bool
    }{
        {
            name: "valid article",
            input: CreateArticleRequest{
                Title:   "Test Article",
                Content: "Test content",
            },
            want: &Article{
                Title:   "Test Article",
                Content: "Test content",
            },
            wantErr: false,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Test Coverage
- **Target**: 85%+ code coverage
- **Required**: All new features must have tests
- **Public APIs**: 100% test coverage required

## 📚 Documentation Guidelines

### Code Documentation
```go
// GetArticlesByCategory retrieves articles filtered by category.
// It returns a paginated list of articles and handles caching automatically.
//
// Parameters:
//   - ctx: Request context with timeout and cancellation
//   - categoryID: The ID of the category to filter by
//   - opts: Pagination and sorting options
//
// Returns:
//   - Articles matching the category
//   - Error if database query fails or category doesn't exist
func GetArticlesByCategory(ctx context.Context, categoryID uint, opts PaginationOptions) ([]Article, error) {
    // Implementation
}
```

### API Documentation
- Use clear, descriptive endpoint names
- Include request/response examples
- Document error responses
- Specify authentication requirements

### README Updates
- Keep examples current
- Test all code snippets
- Update version information
- Add new features to feature list

## 🚀 Release Process

### Versioning
We use [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes

### Changelog
All changes must be documented in `CHANGELOG.md`:
```markdown
## [1.2.0] - 2025-06-14

### Added
- New semantic search endpoint
- User preference settings

### Changed  
- Improved error messages
- Updated dependencies

### Fixed
- Fixed authentication bug
- Resolved memory leak
```

## 🏆 Recognition

### Contributors
All contributors are recognized in:
- `CONTRIBUTORS.md` file
- Release notes
- Social media announcements
- Special badges for significant contributions

### Types of Recognition
- 🥇 **Core Contributors** - Major features or ongoing maintenance
- 🥈 **Regular Contributors** - Multiple contributions over time  
- 🥉 **First-time Contributors** - Welcome to the community!
- 🏅 **Special Recognition** - Documentation, testing, community building

## ❓ Questions?

### Getting Help
- 📖 **Documentation**: Start with [docs/README.md](./docs/README.md)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/your-username/news-api/discussions)
- 🐛 **Issues**: [GitHub Issues](https://github.com/your-username/news-api/issues)
- 📧 **Email**: For sensitive issues only

### Community
- Be respectful and inclusive
- Help newcomers get started
- Share knowledge and experience
- Celebrate successes together

---

## 🙏 Thank You!

Every contribution makes News API better for everyone. Whether you're:

- 🐛 **Fixing bugs**
- ✨ **Adding features** 
- 📚 **Improving docs**
- 🧪 **Writing tests**
- 🌍 **Translating content**
- 💡 **Sharing ideas**

**You're making a difference!** 🎉

---

> 💡 **New to open source?** Don't worry! Everyone started somewhere. We're here to help you make your first contribution. Start small, ask questions, and learn as you go!
   # Start all services including database
   make docker-dev
   ```
   
   **Or run manually:**
   
   ```bash
   # Run database migrations
   make migrate-up
   
   # Seed the database
   make seed
   
   # Start the development server
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
GONews/
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

## ❓ Questions?

### Getting Help
- 📖 **Documentation**: Start with [docs/README.md](./docs/README.md)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/Madraka/GONews/discussions)
- 🐛 **Issues**: [GitHub Issues](https://github.com/Madraka/GONews/issues)
- 📧 **Direct Contact**: For sensitive issues only

---

**🏠 Repository**: https://github.com/Madraka/GONews  
**🌟 Star the project** if you find it useful!  
**📢 Share with others** who might be interested!

Thank you for contributing! 🚀
