# 🚀 GONews Developer Quick Start Guide

Welcome to the GONews project! This guide will help you get started with the newly organized codebase.

## 📁 Project Structure Overview

```
GONews/
├── 📄 README.md                    # You are here!
├── 🛠️ Makefile                     # Build commands
├── 🐳 docker-compose.yml           # Main Docker setup
├── 📊 docs/                        # All documentation
├── 🧪 tests/                       # Tests and test scripts
├── 🚀 deployments/                 # Deployment configurations
├── 📈 monitoring/                   # Observability setup
├── 🏗️ cmd/                         # Application entry points
├── 🔒 internal/                    # Private application code
├── 📦 bin/                         # Compiled binaries (auto-generated)
└── 🛠️ scripts/                     # Utility scripts
```

## 🏃‍♂️ Quick Start Commands

### 1. Environment Setup
```bash
# Clone and enter the project
cd /Users/madraka/GONews

# Copy environment file
cp .env.example .env  # Edit with your settings

# Install dependencies
go mod download
```

### 2. Database Setup
```bash
# Start database and services
make db-up

# Run migrations
make migrate-up

# Verify setup
make migrate-status
```

### 3. Development Modes

#### Local Development
```bash
# Build and run locally
make build && make run

# Or use hot reload with Air
air
```

#### Docker Development
```bash
# Start all services (API + Database + Monitoring)
make dev-all-up

# View logs
docker-compose logs -f api

# Stop all services
make dev-all-down
```

### 4. Testing
```bash
# Run unit tests
make test

# Run integration tests
make integration-test

# Test API endpoints
./tests/scripts/quick_test.sh

# Test observability stack
./tests/scripts/test_observability.sh
```

## 📍 Key File Locations

### Need to find something? Here's where to look:

| What you need | Where to find it |
|---------------|------------------|
| 🔧 **API Handlers** | `internal/handlers/` |
| 🛣️ **Routes** | `internal/routes/` |
| 🗃️ **Database Models** | `internal/models/` |
| 🔐 **Middleware** | `internal/middleware/` |
| 📋 **API Documentation** | `docs/api/swagger.json` |
| 🧪 **Test Scripts** | `tests/scripts/` |
| 🐳 **Docker Configs** | `deployments/docker/` |
| 📊 **Monitoring** | `monitoring/` |
| 🔧 **Build Scripts** | `scripts/build/` |

### Quick Navigation Commands
```bash
# Jump to key directories
cd internal/handlers     # API handlers
cd docs/api             # API documentation  
cd tests/scripts        # Test scripts
cd deployments/dev      # Development configs
cd monitoring/          # Observability configs
```

## 🔨 Common Development Tasks

### Adding a New API Endpoint
1. **Create Handler**: Add to `internal/handlers/`
2. **Add Route**: Register in `internal/routes/`
3. **Add Tests**: Create test in `tests/`
4. **Update Docs**: Run `swag init` to update Swagger

### Working with Database
```bash
# Create new migration
make migrate-create name=add_new_table

# Apply migrations
make migrate-up

# Rollback migration
make migrate-down

# Check migration status
make migrate-status
```

### Testing Your Changes
```bash
# Quick API test
./tests/scripts/quick_test.sh

# Verify handlers work
./tests/scripts/verify_handlers.sh

# Test with authentication
./tests/scripts/test_api_authenticated.sh
```

### Monitoring and Observability
```bash
# Start monitoring stack
make metrics-up

# View metrics: http://localhost:9090 (Prometheus)
# View dashboards: http://localhost:3000 (Grafana)
# API metrics: http://localhost:8080/metrics
```

## 📖 Documentation Guide

### Where to Find Information

| Topic | Location | Description |
|-------|----------|-------------|
| **API Reference** | `docs/api/` | Swagger documentation, API guides |
| **Setup Guides** | `docs/guides/` | Implementation and configuration guides |
| **Test Results** | `docs/reports/` | Test reports and deployment status |
| **Database** | `docs/migration/` | Migration guides and database docs |

### Key Documentation Files
- 📋 **API Docs**: `docs/api/swagger.json` - Complete API specification
- 🔍 **Observability**: `docs/guides/OBSERVABILITY_GUIDE.md` - Monitoring setup
- 🔐 **Security**: `docs/guides/enhanced_security_guide.md` - Security best practices
- 🤖 **AI Integration**: `docs/guides/ai_integration_guide.md` - AI features guide

## 🐛 Debugging and Troubleshooting

### Health Checks
```bash
# Check API health
curl http://localhost:8080/health

# Check all container status
docker-compose ps

# View container logs
docker-compose logs api
```

### Debug Tools
```bash
# Run debug server (in debug/ directory)
make debug-server

# Run debug client
make debug-client

# Check metrics endpoint
curl http://localhost:8080/metrics
```

### Common Issues
1. **Port conflicts**: Check if ports 8080, 5432, 6379 are available
2. **Database connection**: Ensure PostgreSQL is running (`make db-up`)
3. **Redis connection**: Ensure Redis is running
4. **Missing migrations**: Run `make migrate-up`

## 🎯 Best Practices

### Code Organization
- ✅ Place handlers in `internal/handlers/`
- ✅ Keep business logic in `internal/services/`
- ✅ Put tests near the code they test
- ✅ Update documentation when adding features

### Git Workflow
- ✅ Binaries are auto-ignored (in `bin/`)
- ✅ Use meaningful commit messages
- ✅ Test before committing
- ✅ Keep `.env` files out of git

### Development Environment
- ✅ Use `make dev-all-up` for full development stack
- ✅ Use `make test` before pushing changes
- ✅ Check `make lint` for code quality
- ✅ Update Swagger docs with `swag init`

## 🆘 Getting Help

### Resources
- 📖 **Full Documentation**: Browse `docs/` directory
- 🧪 **Test Examples**: Check `tests/scripts/` for examples
- 🔧 **Build Commands**: Run `make help` for all available commands
- 📊 **Monitoring**: Access Grafana at http://localhost:3000

### Quick Reference
```bash
# See all available make commands
make help

# Check project status
make status

# Run comprehensive tests
make test-all

# View API documentation
open http://localhost:8080/swagger/index.html
```

---

🎉 **Happy coding!** The project is now well-organized and ready for development. All documentation is in English and properly categorized for easy access.
