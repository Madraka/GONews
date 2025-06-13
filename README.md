# 📰 News API

A modern, production-ready RESTful API for managing news articles with advanced features including authentication, real-time updates, AI-powered semantic search, monitoring, and comprehensive content management.

![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)

## � Features

### 🚀 Core Features
- **📝 Article Management**: Create, edit, delete, and organize news articles
- **👥 User Management**: Authentication, authorization, and user profiles
- **🔍 Advanced Search**: Full-text search with AI-powered semantic search
- **📱 Real-time Updates**: WebSocket support for live news feeds
- **🌍 Multi-language Support**: Internationalization (English, Spanish, Turkish)
- **📊 Analytics**: Comprehensive metrics and user behavior tracking

### 🤖 AI-Powered Features
- **🧠 Semantic Search**: OpenAI-powered intelligent article search
- **📈 Content Recommendations**: AI-driven article suggestions
- **🔖 Auto-tagging**: Automatic content categorization
- **📊 Sentiment Analysis**: Article sentiment scoring

### 🎯 Modern News Features
- **🚨 Breaking News Banners**: Urgent news alerts and notifications
- **📖 News Stories**: Instagram-style ephemeral content
- **🔴 Live News Streams**: Real-time coverage of ongoing events
- **🎥 Video Integration**: Embedded video content support
- **📱 Mobile-First Design**: Responsive API design for all devices

### 🔒 Security & Performance
- **🛡️ JWT Authentication**: Secure token-based authentication
- **🚦 Rate Limiting**: Intelligent request throttling and usage quotas
- **⚡ Redis Caching**: High-performance caching layer
- **🔐 Role-Based Access**: Granular permission system
- **📊 Monitoring & Tracing**: Comprehensive observability

## 🏗️ Project Architecture

This project follows a clean, organized structure based on Go best practices:

```
News/
├── 🚀 cmd/                    # Application entry points
│   ├── api/                   # Main API server
│   ├── worker/                # Background job processor
│   ├── migrate/               # Database migration tool
│   └── seed/                  # Database seeding utility
├── 🔒 internal/               # Private application code
│   ├── auth/                  # Authentication & authorization
│   ├── handlers/              # HTTP request handlers
│   ├── models/                # Data models and entities
│   ├── services/              # Business logic layer
│   ├── repositories/          # Data access layer
│   ├── middleware/            # HTTP middleware
│   └── config/                # Configuration management
├── 📚 docs/                   # All documentation
│   ├── api/                   # API documentation (Swagger)
│   ├── guides/                # Implementation guides
│   ├── reports/               # Test and deployment reports
│   └── migration/             # Migration documentation
├── 🧪 tests/                  # Comprehensive test suite
│   ├── unit/                  # Unit tests
│   ├── integration/           # Integration tests
│   └── e2e/                   # End-to-end tests
├── 🐳 deployments/           # Docker and Kubernetes configs
│   ├── dev/                   # Development environment
│   ├── test/                  # Testing environment
│   └── prod/                  # Production environment
├── 📊 monitoring/            # Observability configuration
│   ├── grafana/               # Grafana dashboards
│   └── prometheus/            # Prometheus configuration
├── 🔧 scripts/               # Build and utility scripts
└── 📦 bin/                   # Compiled binaries (git-ignored)
```

### 📖 Quick Links
- 🚀 **New to the project?** → [Developer Quick Start Guide](./docs/DEVELOPER_GUIDE.md)
- 🏗️ **Project Organization** → [Organization Guide](./docs/PROJECT_ORGANIZATION_GUIDE.md)
- � **API Documentation** → [Swagger UI](http://localhost:8080/swagger/index.html)
- 📊 **Monitoring** → [Grafana Dashboard](http://localhost:3000)

## 🌍 Environment Configuration

The News API supports three distinct environments with complete isolation and specialized configurations:

### 📋 Environment Overview

| Environment | Purpose | API Port | DB Port | Redis Port | Database Name | Status |
|-------------|---------|----------|---------|------------|---------------|---------|
| **🔧 Development** | Local development | 8081 | 5434 | 6380 | `newsapi_development` | Debug enabled |
| **🧪 Testing** | Automated testing | 8082 | 5432 | 6379 | `newsapi_test` | Optimized for CI/CD |
| **🚀 Production** | Live deployment | 8080 | 5432 | 6379 | `newsapi_production` | Security hardened |

### 🔧 Quick Environment Management

Switch between environments effortlessly:

```bash
# 🔧 Development Environment
make env-dev              # Switch to development
make dev-up              # Start dev services (PostgreSQL + Redis)
make dev                 # Start API with hot reload

# 🧪 Testing Environment  
make env-test            # Switch to testing
make test-local          # Run all tests locally
make test-docker         # Run tests in containers

# 🚀 Production Environment
make env-prod            # Switch to production
make prod-deploy         # Deploy to production

# 📊 Environment Status
make env-show            # Show current environment
make status              # Show all services status
```

### 🛠️ Environment-Specific Features

<details>
<summary>🔧 <strong>Development Environment</strong> (Recommended for coding)</summary>

**Port Configuration:**
- 🔌 **API Port:** 8081 (avoids conflicts)
- 🗄️ **Database Port:** 5434 (isolated PostgreSQL)
- 🔴 **Redis Port:** 6380 (isolated Redis)

**Features:**
- ✅ Debug mode enabled
- ✅ Hot reload support with Air
- ✅ CORS enabled for frontend development
- ✅ Comprehensive logging (debug level)
- ✅ Monitoring and tracing enabled
- ✅ Swagger UI available
- ✅ Live configuration reloading
- ✅ Development seed data
</details>

<details>
<summary>🧪 <strong>Testing Environment</strong> (Optimized for CI/CD)</summary>

**Port Configuration:**
- � **API Port:** 8082 (dedicated test port)
- 🗄️ **Database Port:** 5432 (containerized)
- 🔴 **Redis Port:** 6379 (containerized)

**Features:**
- ✅ Isolated test database
- ✅ Short JWT token duration (5m)
- ✅ Deterministic AI responses (temp=0.0)
- ✅ Automated test data cleanup
- ✅ Parallel test execution
- ❌ Debug features disabled
- ❌ External API calls mocked
</details>

<details>
<summary>🚀 <strong>Production Environment</strong> (Battle-tested)</summary>

**Port Configuration:**
- 🔌 **API Port:** 8080 (standard)
- 🗄️ **Database Port:** 5432 (production PostgreSQL)
- 🔴 **Redis Port:** 6379 (production Redis)

**Features:**
- ✅ SSL/TLS enforcement
- ✅ Security headers enabled
- ✅ Rate limiting enforced
- ✅ Production monitoring
- ✅ Error tracking (Sentry)
- ✅ Performance optimization
- ✅ Automated backups
- ❌ Debug endpoints disabled
</details>

### � Configuration Management

**Configuration Files Structure:**
```
deployments/
├── dev/
│   ├── .env.dev           # Development settings
│   └── docker-compose-dev.yml
├── test/
│   ├── .env.test          # Testing settings
│   └── docker-compose-test.yml
└── prod/
    ├── .env.prod          # Production settings (⚠️ SECRETS)
    └── docker-compose-prod.yml
```

**Key Configuration Categories:**

<details>
<summary>🔧 <strong>Server Configuration</strong></summary>

```bash
# Core server settings
PORT=8080                    # API server port
ENVIRONMENT=production       # Environment name
LOG_LEVEL=info              # Logging level (debug/info/warn/error)
DEBUG_MODE=false            # Enable debug features
CORS_ENABLED=true           # Enable CORS for frontend
```
</details>

<details>
<summary>🗄️ <strong>Database Configuration</strong></summary>

```bash
# PostgreSQL settings
DATABASE_URL=postgresql://user:pass@host:port/db?sslmode=require
DB_HOST=localhost           # Database host
DB_PORT=5432               # Database port
DB_USER=newsapi            # Database username
DB_PASSWORD=secure_pass    # Database password
DB_NAME=newsapi_production # Database name
DB_SSL_MODE=require        # SSL mode (disable/require)
DB_MAX_CONNECTIONS=25      # Connection pool size
```
</details>

<details>
<summary>🔴 <strong>Redis Configuration</strong></summary>

```bash
# Redis cache settings
REDIS_URL=redis://localhost:6379/0
REDIS_PASSWORD=redis_pass   # Redis password
REDIS_DB=0                 # Redis database number
ACCESS_TOKEN_DURATION=15m   # JWT access token duration
REFRESH_TOKEN_DURATION=7d   # JWT refresh token duration
```
</details>

<details>
<summary>🤖 <strong>AI Configuration</strong></summary>

```bash
# OpenAI integration
OPENAI_API_KEY=sk-...      # OpenAI API key
OPENAI_MODEL=gpt-4         # AI model to use
OPENAI_MAX_TOKENS=1000     # Maximum tokens per request
OPENAI_TEMPERATURE=0.7     # AI response randomness (0.0-1.0)
AI_ENABLED=true            # Enable AI features
```
</details>

### 🔒 Security & Secrets Management

**Environment Security Levels:**

| Environment | Security Level | Secrets Handling | Commit Policy |
|-------------|---------------|------------------|---------------|
| **Development** | 🟢 Low | Dummy values | ✅ Safe to commit |
| **Testing** | 🟡 Medium | Test-only values | ✅ Safe to commit |
| **Production** | 🔴 High | Real secrets | ❌ **NEVER COMMIT** |

**⚠️ Important Security Notes:**
- Production `.env.prod` contains real secrets - store in secure vault
- Use environment variables or Docker secrets in production
- Rotate secrets regularly (monthly recommended)
- Monitor for secret leaks in logs and traces

### 🛡️ Port Conflict Prevention

Smart port allocation prevents conflicts when running multiple environments:

```bash
# Development (can run alongside others)
API: 8081, DB: 5434, Redis: 6380, Monitoring: 9091

# Testing (containerized isolation)  
API: 8082, DB: 5432*, Redis: 6379* (*containerized)

# Production (standard ports)
API: 8080, DB: 5432, Redis: 6379, Monitoring: 9090
```

This design allows you to:
- ✅ Run development services locally while testing in containers
- ✅ Debug production issues without stopping live services
- ✅ Perform A/B testing with multiple environments
- ✅ Safely develop new features without interference

## 🚀 Quick Start Guide

### 📋 Prerequisites

Ensure you have the following installed:

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| **Go** | 1.24+ | Main language | [Download](https://golang.org/dl/) |
| **PostgreSQL** | 15+ | Primary database | [Download](https://postgresql.org/download/) |
| **Redis** | 7+ | Caching & sessions | [Download](https://redis.io/download/) |
| **Docker** | 20+ | Containerization | [Download](https://docker.com/get-started/) |
| **Make** | 4+ | Build automation | Usually pre-installed |

**Optional but Recommended:**
- **kubectl** - Kubernetes deployment
- **Air** - Hot reload for development
- **Node.js** - Frontend development and load testing

### 🏃‍♂️ 5-Minute Setup

Get the News API running in under 5 minutes:

```bash
# 1️⃣ Clone the repository
git clone https://github.com/yourusername/news-api.git
cd news-api

# 2️⃣ Install Go dependencies
go mod download

# 3️⃣ Start development environment
make dev-setup           # Setup environment and start services
make migrate-up          # Apply database migrations
make seed-db            # Load sample data

# 4️⃣ Start the API server
make dev                # Start with hot reload
# OR
make build && make run  # Build and run normally

# 5️⃣ Verify everything works
curl http://localhost:8081/health
```

**🎉 Success!** Your API is now running at:
- **API Server**: http://localhost:8081
- **Swagger Docs**: http://localhost:8081/swagger/index.html
- **Health Check**: http://localhost:8081/health

### � Development Workflow

**Daily Development Commands:**
```bash
# Start your development day
make dev-up              # Start PostgreSQL + Redis
make dev                 # Start API with hot reload

# During development
make test-local          # Run tests
make docs               # Update API documentation
make lint               # Run code linting

# End of day
make dev-down           # Stop development services
```

### 🛠️ Advanced Setup Options

<details>
<summary>🐳 <strong>Docker-Only Setup</strong> (Recommended for beginners)</summary>

```bash
# Start everything with Docker
make docker-dev-up      # Start all services in containers
make docker-logs        # View logs
make docker-shell       # Access container shell

# Useful for:
# - Quick setup without installing dependencies
# - Testing in production-like environment
# - Avoiding version conflicts
```
</details>

<details>
<summary>🔧 <strong>Manual Setup</strong> (For advanced users)</summary>

```bash
# 1. Start PostgreSQL manually
sudo systemctl start postgresql
createdb newsapi_development

# 2. Start Redis manually
sudo systemctl start redis
redis-cli ping

# 3. Configure environment
cp deployments/dev/.env.dev .env
# Edit .env with your settings

# 4. Run migrations and seeds
make migrate-up
make seed-db

# 5. Start API
go run ./cmd/api/main.go
```
</details>

<details>
<summary>🌐 <strong>Full Stack Setup</strong> (API + Frontend + Monitoring)</summary>

```bash
# Start comprehensive development environment
make full-dev-up        # API + Database + Redis + Monitoring
make monitoring-up      # Grafana + Prometheus + Jaeger

# Access all services:
# - API: http://localhost:8081
# - Grafana: http://localhost:3000
# - Prometheus: http://localhost:9090
# - Jaeger: http://localhost:16686
```
</details>

### 🔍 Verification Steps

After setup, verify everything is working:

```bash
# 1. Check API health
curl http://localhost:8081/health
# Expected: {"status": "ok", "timestamp": "..."}

# 2. Test authentication
curl -X POST http://localhost:8081/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}'

# 3. Fetch some news articles
curl http://localhost:8081/api/news
# Expected: Array of news articles

# 4. Test AI search (if configured)
curl "http://localhost:8081/api/search/semantic?q=technology"
```

### 🚨 Troubleshooting

**Common Issues and Solutions:**

<details>
<summary>❌ <strong>Port Already in Use</strong></summary>

```bash
# Check what's using the port
lsof -i :8081

# Kill the process
kill -9 <PID>

# Or use a different port
PORT=8082 make dev
```
</details>

<details>
<summary>❌ <strong>Database Connection Failed</strong></summary>

```bash
# Check if PostgreSQL is running
make db-status

# Restart database
make db-restart

# Check logs
make db-logs
```
</details>

<details>
<summary>❌ <strong>Redis Connection Failed</strong></summary>

```bash
# Check Redis status
redis-cli ping

# Restart Redis
make redis-restart

# Use different Redis instance
REDIS_URL=redis://localhost:6380 make dev
```
</details>

<details>
<summary>❌ <strong>Go Module Issues</strong></summary>

```bash
# Clean and reinstall modules
go clean -modcache
go mod download
go mod tidy

# If you're behind a proxy
go env -w GOPROXY=direct
go env -w GOSUMDB=off
```
</details>

### 📚 Next Steps

Once you have the API running:

1. **📖 Read the Documentation**
   - [Developer Guide](./docs/DEVELOPER_GUIDE.md)
   - [API Documentation](http://localhost:8081/swagger/index.html)

2. **🧪 Explore the Features**
   - Try the semantic search
   - Test the authentication system
   - Create some articles

3. **🔧 Customize the Setup**
   - Configure AI integration
   - Set up monitoring
   - Add your own middleware

4. **🚀 Deploy to Production**
   - [Deployment Guide](./deployments/README.md)
   - [Kubernetes Setup](./deployments/k8s/README.md)

## 🗄️ Database Management

### 🚀 Quick Database Setup

**Automated Setup (Recommended):**
```bash
make dev-up              # Start PostgreSQL + Redis containers
make migrate-up          # Apply all migrations
make seed-db            # Load sample data
```

**Manual Database Setup:**
```bash
# Start PostgreSQL container
docker run -d \
  --name news-postgres \
  -e POSTGRES_USER=devuser \
  -e POSTGRES_PASSWORD=devpass123 \
  -e POSTGRES_DB=newsapi_development \
  -p 5434:5432 \
  postgres:15-alpine

# Start Redis container  
docker run -d \
  --name news-redis \
  -p 6380:6379 \
  redis:7-alpine
```

### 🔄 Migration Management

The News API includes a powerful, automated database migration system with Atlas integration:

**Quick Migration Commands:**
```bash
# 📊 Check current status
make migrate-status      # View migration status
make db-status          # Check database connectivity

# ⬆️ Apply migrations
make migrate-up         # Apply all pending migrations
make migrate-up-one     # Apply next migration only

# ⬇️ Rollback migrations  
make migrate-down       # Rollback last migration
make migrate-down-to VERSION=20240101  # Rollback to specific version

# 🆕 Create new migrations
make migrate-create NAME=add_user_profiles
make migrate-create NAME=update_article_schema
```

**Advanced Migration Features:**
```bash
# 🔍 Migration validation
make migrate-validate   # Validate migration files
make migrate-dry-run    # Test migrations without applying

# �️ Database utilities
make db-reset          # Reset database (⚠️ DATA LOSS)
make db-backup         # Create database backup
make db-restore FILE=backup.sql  # Restore from backup

# 📈 Schema management
make schema-generate   # Generate current schema
make schema-diff       # Show schema differences
```

### 🌱 Database Seeding System

Comprehensive database seeding with realistic sample data:

**Quick Seeding:**
```bash
make seed-db           # Auto-detect environment and seed
make seed-db-dev       # Seed development database
make seed-db-prod      # Seed production database (⚠️ use carefully)
```

**Seeding Categories:**
```
scripts/seeds/
├── 01_core/           # 👥 Users, roles, permissions
├── 02_content/        # 📰 Articles, categories, tags  
├── 03_system/         # ⚙️ Settings, configurations
├── 04_interactions/   # 👍 Votes, bookmarks, follows
└── 05_relationships/  # 🔗 Article-category mappings
```

**Sample Data Included:**
- **👥 13 Users**: Admin, editors, regular users with Turkish names
- **📂 12 Categories**: Technology, Sports, Economy, Health, etc.
- **🏷️ 28 Tags**: Various content tags for classification
- **📰 19 Articles**: Complete Turkish news articles with metadata
- **⚙️ System Settings**: Site configuration and feature toggles
- **🧭 Navigation Menus**: Site structure and routing
- **💝 User Interactions**: Sample votes, bookmarks, follows
- **🔗 Relationships**: Article-category and tag associations

**Environment-Specific Seeding:**
- **Development**: Full dataset with debug information
- **Testing**: Minimal dataset for fast test execution
- **Production**: Essential data only (admin user, basic categories)

### 🔧 Database Tools & Utilities

**Connection Management:**
```bash
# 🔌 Connection testing
make db-ping           # Test database connectivity
make redis-ping        # Test Redis connectivity

# 📊 Database inspection
make db-shell          # Connect to database shell
make db-stats          # Show database statistics
make db-size           # Show database size info
```

**Maintenance Commands:**
```bash
# 🧹 Cleanup operations
make db-vacuum         # Optimize database performance
make db-reindex        # Rebuild database indexes
make cache-clear       # Clear Redis cache

# 📈 Performance monitoring
make db-slow-queries   # Show slow queries
make db-connections    # Show active connections
make cache-stats       # Show Redis statistics
```

### 🏗️ Schema Management with Atlas

The project uses Atlas for advanced schema management:

**Atlas Commands:**
```bash
# 📋 Schema inspection
make atlas-inspect     # Inspect current schema
make atlas-schema      # Generate schema file

# 🔄 Migration planning
make atlas-plan        # Plan migration strategy
make atlas-apply       # Apply planned migrations

# 🔍 Schema validation
make atlas-validate    # Validate schema consistency
make atlas-diff        # Compare schemas
```

**Atlas Configuration:**
```hcl
# atlas.hcl
env "dev" {
  src = "file://schema/current.hcl"
  url = "postgres://user:pass@localhost:5434/newsapi_dev?sslmode=disable"
  migration {
    dir = "file://migrations/atlas"
  }
}
```

### 🚨 Backup & Recovery

**Automated Backups:**
```bash
# 💾 Create backups
make backup-create     # Create timestamped backup
make backup-schedule   # Setup automated backups

# 📥 Restore operations
make backup-list       # List available backups
make backup-restore BACKUP=backup_20240613.sql
```

**Backup Storage:**
```
backups/
├── daily/             # Daily automated backups
├── manual/            # Manual backup files
└── migration/         # Pre-migration backups
```

### 🔍 Database Monitoring

**Performance Metrics:**
- Connection pool usage
- Query execution times
- Cache hit ratios
- Transaction statistics

**Monitoring Commands:**
```bash
make db-monitor        # Real-time database monitoring
make db-explain QUERY="SELECT * FROM articles"  # Query analysis
make cache-monitor     # Redis cache monitoring
```

Access monitoring dashboards:
- **Grafana**: http://localhost:3000 (Database Dashboard)
- **pgAdmin**: http://localhost:5050 (Database Administration)
- **Redis Insight**: http://localhost:8001 (Redis Management)

## 🚀 Deployment & Production

> **📋 For comprehensive deployment instructions, see [deployments/README.md](deployments/README.md)**

### 🎯 Deployment Options

The News API supports multiple deployment strategies to fit different needs and scales:

| Deployment Type | Use Case | Complexity | Scalability | Recommended For |
|----------------|----------|------------|-------------|-----------------|
| **🐳 Docker Compose** | Single server | Low | Medium | Small to medium projects |
| **☸️ Kubernetes** | Multi-server | High | High | Enterprise & high-traffic |
| **☁️ Cloud Native** | Managed services | Medium | Very High | Production workloads |
| **🔧 Bare Metal** | Direct installation | Medium | Low | Development & testing |

### 🐳 Docker Deployment (Recommended)

**Quick Deployment:**
```bash
# 🔧 Development Environment (port 8081)
./deployments/scripts/deploy.sh dev --migrate --monitor

# 🧪 Testing Environment (port 8082)
./deployments/scripts/deploy.sh test --migrate --test

# 🚀 Production Environment (port 8080)
./deployments/scripts/deploy.sh prod --backup --migrate --monitor --health
```

**Environment-Specific Deployment:**

<details>
<summary>🔧 <strong>Development Deployment</strong></summary>

```bash
cd deployments/dev
docker-compose -f docker-compose-dev.yml up -d

# Services included:
# - API server with hot reload
# - PostgreSQL with development data
# - Redis for caching
# - Grafana for monitoring
# - Jaeger for tracing

# Access points:
# - API: http://localhost:8081
# - Docs: http://localhost:8081/swagger/index.html  
# - Grafana: http://localhost:3000
# - Jaeger: http://localhost:16686
```
</details>

<details>
<summary>🧪 <strong>Testing Deployment</strong></summary>

```bash
cd deployments/test
docker-compose -f docker-compose-test.yml up -d

# Optimized for CI/CD with:
# - Minimal resource usage
# - Fast startup times
# - Automated test execution
# - Ephemeral data storage

# Access point:
# - API: http://localhost:8082
```
</details>

<details>
<summary>🚀 <strong>Production Deployment</strong></summary>

```bash
cd deployments/prod
docker-compose -f docker-compose-prod.yml up -d

# Production features:
# - SSL/TLS termination
# - Security hardening
# - Performance optimization
# - Health monitoring
# - Automated backups
# - Log aggregation

# Access point:
# - API: https://localhost:8080
```
</details>

### ☸️ Kubernetes Deployment

**Prerequisites:**
- Kubernetes cluster (1.24+)
- kubectl configured
- Helm 3.0+ (optional)

**Quick Kubernetes Setup:**
```bash
# 📦 Deploy using kubectl
kubectl apply -f deployments/k8s/

# 🎡 Deploy using Helm (recommended)
helm install news-api deployments/helm/news-api/

# 📊 Monitor deployment
kubectl get pods -l app=news-api
kubectl logs -f deployment/news-api
```

**Available Environments:**
- **Development:** Port 8081, debug mode, live reload
- **Testing:** Port 8082, optimized for CI/CD  
- **Production:** Port 8080, security hardened, monitoring enabled

For advanced deployment options, troubleshooting, Kubernetes setup, and production best practices, see the [comprehensive deployment guide](deployments/README.md).

## 🚦 Rate Limiting & API Protection

The News API implements intelligent rate limiting to ensure fair usage, prevent abuse, and control operational costs. Our multi-tier rate limiting system adapts to user authentication levels and endpoint types.

### 🎯 Rate Limiting Overview

**Rate Limiting Strategy:**
- 🔄 **Token Bucket Algorithm**: Smooth traffic flow with burst capability
- 👤 **User-Based Limits**: Different limits for different user types
- 🔍 **Endpoint-Specific Limits**: AI endpoints have special cost controls
- 🌍 **Global Protection**: System-wide limits for service protection
- 📊 **Sliding Window**: Accurate usage tracking over time

### 🔍 AI-Powered Search Rate Limits

AI-powered semantic search endpoints have specialized rate limiting to control OpenAI API costs:

| User Type | AI Search Limit | Local Search Limit | Reset Period | Cost Control |
|-----------|-----------------|-------------------|--------------|--------------|
| **🚫 Unauthenticated** | 5 requests/day | 50 requests/day | Daily at midnight | ✅ Strict |
| **👤 Authenticated** | 50 requests/day | 500 requests/day | Daily at midnight | ✅ Generous |
| **👑 Premium** | 200 requests/day | Unlimited | Daily at midnight | ✅ Extended |
| **🌍 Global Limit** | 10,000 requests/day | - | System protection | ✅ Emergency |

**AI Search Endpoints:**
- `GET /api/search/semantic` - Public semantic search (5 AI requests/day)
- `GET /api/v1/search` - Authenticated search (50 AI requests/day)
- `POST /api/v1/search/advanced` - Advanced AI search (premium only)

**Smart Fallback System:**
- ✅ **AI Quota Exceeded**: Automatically falls back to local search
- ✅ **No Service Interruption**: Users always get results
- ✅ **Transparent**: Response metadata indicates fallback reason
- ✅ **Cost Protection**: Prevents unexpected AI API charges

### 📈 Standard API Rate Limits

Regular API endpoints use sophisticated token bucket rate limiting:

| Endpoint Category | Rate Limit | Burst Limit | Window | User Type |
|------------------|------------|-------------|---------|-----------|
| **🌐 Public endpoints** | 5 req/sec | 10 requests | 1 minute | All users |
| **🔐 Authenticated endpoints** | 10 req/sec | 20 requests | 1 minute | JWT required |
| **👑 Admin endpoints** | 15 req/sec | 30 requests | 1 minute | Admin role |
| **🔍 Search endpoints** | 3 req/sec | 10 requests | 1 minute | Special limits |
| **📤 Upload endpoints** | 1 req/sec | 3 requests | 5 minutes | File uploads |

**Advanced Rate Limiting Features:**
- ✅ **Sliding Window**: Accurate usage tracking
- ✅ **Burst Allowance**: Handle traffic spikes gracefully
- ✅ **Path-Specific Limits**: Different limits for different endpoints
- ✅ **IP + User Tracking**: Prevent abuse from multiple sources
- ✅ **Exponential Backoff**: Automatic retry delay calculation

### 🔑 Rate Limit Response Headers

All API responses include comprehensive rate limit information:

```bash
# Standard rate limit headers
X-RateLimit-Limit: 50              # Total requests allowed
X-RateLimit-Remaining: 45          # Requests remaining
X-RateLimit-Reset: 1640995200      # Reset timestamp
X-RateLimit-Window: 60             # Window duration (seconds)

# When rate limited
HTTP/1.1 429 Too Many Requests
Retry-After: 60                    # Seconds to wait
X-RateLimit-Scope: "user"          # Limit scope (user/ip/global)
X-RateLimit-Type: "api"            # Limit type (api/ai/upload)
```

### 🔍 Rate Limit Monitoring

Monitor your API usage with dedicated monitoring endpoints:

**Check Current Limits:**
```bash
# Check your current rate limit status
curl "http://localhost:8081/api/limits"

# Check semantic search limits (public)
curl "http://localhost:8081/api/search/limits"

# Check authenticated user limits
curl -H "Authorization: Bearer $JWT_TOKEN" \
     "http://localhost:8081/api/v1/limits"
```

**Example Response:**
```json
{
  "rate_limits": {
    "api_requests": {
      "limit": 600,
      "remaining": 545,
      "reset_time": "2025-06-03T15:30:00Z",
      "window_seconds": 3600
    },
    "ai_searches": {
      "limit": 50,
      "remaining": 45,
      "used": 5,
      "reset_time": "2025-06-04T00:00:00Z"
    }
  },
  "user_type": "authenticated",
  "current_time": "2025-06-03T14:30:00Z"
}
```

### 🎛️ Rate Limiting Behavior

**When AI Search Limits Are Exceeded:**
- 🔄 **Graceful Fallback**: Automatically switches to local search
- 📊 **Transparent Response**: Includes fallback reason in metadata
- ✅ **No Service Interruption**: Users always receive results
- 📈 **Performance Maintained**: Local search is often faster

**When Standard Rate Limits Are Exceeded:**
- 🚫 **HTTP 429 Response**: "Too Many Requests" status
- ⏰ **Retry-After Header**: Indicates when to retry
- 🔄 **Exponential Backoff**: Recommended retry strategy
- 📊 **Usage Guidance**: Headers show current usage status

### 🔧 Rate Limiting Implementation

**Backend Storage:**
- **🏠 Development**: In-memory rate limiting (single instance)
- **🚀 Production**: Redis-based distributed rate limiting
- **🔄 Auto-Fallback**: Gracefully falls back to memory if Redis unavailable
- **📊 Persistence**: Rate limit data survives service restarts

**Authentication-Based Limits:**
- **🔐 IP-Based**: 5 login attempts per 10 minutes
- **👤 Username-Based**: 3 failed attempts per 10 minutes  
- **🔒 Account Lockout**: 30 minutes after threshold exceeded
- **📧 Email-Based**: Prevent email enumeration attacks

### 📊 Rate Limiting Metrics

Comprehensive metrics are exposed via Prometheus:

```bash
# View rate limiting metrics
curl http://localhost:8081/metrics | grep rate_limit

# Key metrics available:
# - news_api_rate_limit_exceeded_total
# - news_api_rate_limit_requests_total  
# - news_api_semantic_search_requests_total
# - news_api_semantic_search_ai_usage_total
# - news_api_rate_limit_fallback_total
```

**Grafana Dashboard:**
- 📊 **Rate Limit Usage**: Real-time usage across all endpoints
- 🚨 **Abuse Detection**: Unusual traffic patterns
- 💰 **Cost Monitoring**: AI API usage and cost projections
- 📈 **Performance Impact**: Rate limiting effect on performance

### 💡 Best Practices for API Consumers

**1. 🔍 Always Check Rate Limit Status:**
```bash
# Check limits before making bulk requests
curl "http://localhost:8081/api/limits"
```

**2. 🔄 Handle Rate Limit Responses Gracefully:**
```javascript
async function apiRequest(url, options) {
  const response = await fetch(url, options);
  
  if (response.status === 429) {
    const retryAfter = response.headers.get('retry-after');
    console.log(`Rate limited. Retrying in ${retryAfter} seconds`);
    
    // Exponential backoff with jitter
    const delay = Math.min(retryAfter * 1000, 60000) + Math.random() * 1000;
    await new Promise(resolve => setTimeout(resolve, delay));
    
    return apiRequest(url, options); // Retry
  }
  
  return response;
}
```

**3. 👤 Authenticate for Higher Limits:**
```bash
# Register for higher rate limits
curl -X POST "http://localhost:8081/api/auth/register" \
     -H "Content-Type: application/json" \
     -d '{"username":"user","email":"user@example.com","password":"securepass123"}'
```

**4. 📊 Monitor Your Usage:**
```bash
# Regularly check your usage patterns
curl -H "Authorization: Bearer $JWT_TOKEN" \
     "http://localhost:8081/api/v1/usage/summary"
```

**5. 💾 Cache Results When Possible:**
```javascript
// Cache AI search results to avoid repeated queries
const cacheKey = `search_${query}`;
let results = cache.get(cacheKey);

if (!results) {
  results = await apiClient.search(query);
  cache.set(cacheKey, results, 3600); // Cache for 1 hour
}
```

### 🔒 Security Features

**Abuse Prevention:**
- 🛡️ **DDoS Protection**: Multi-layer rate limiting
- 🚨 **Anomaly Detection**: Unusual traffic pattern alerts
- 📊 **Behavior Analysis**: User behavior tracking
- 🔒 **Automatic Blocking**: Suspicious IP/user blocking
- 📋 **Audit Logging**: Complete request audit trail

**Cost Protection:**
- 💰 **AI Cost Monitoring**: Real-time cost tracking
- 🚨 **Budget Alerts**: Automatic notifications at thresholds
- 🔄 **Fallback Mechanisms**: Graceful degradation
- 📊 **Usage Analytics**: Detailed cost analysis and reporting
## 📊 Monitoring & Observability

The News API includes comprehensive monitoring and observability features with real-time metrics, distributed tracing, and centralized logging.

### 🎯 Observability Stack

| Component | Purpose | Port | Credentials | Status |
|-----------|---------|------|-------------|---------|
| **📈 Prometheus** | Metrics collection | 9090 | None | Production ready |
| **📊 Grafana** | Dashboards & visualization | 3000 | admin/admin | Production ready |
| **🔍 Jaeger** | Distributed tracing | 16686 | None | Production ready |
| **📋 ELK Stack** | Log aggregation | 5601 | elastic/changeme | Optional |
| **🚨 AlertManager** | Alert management | 9093 | None | Production ready |

### 🚀 Quick Start Monitoring

**Start Complete Monitoring Stack:**
```bash
# Start all monitoring services
make monitoring-up

# Start individual components
make prometheus-up    # Metrics only
make grafana-up      # Dashboards only  
make jaeger-up       # Tracing only
make logging-up      # ELK stack

# Check status
make monitoring-status
```

**Access Monitoring Services:**
- 📊 **Grafana Dashboard**: http://localhost:3000 (admin/admin)
- 📈 **Prometheus Metrics**: http://localhost:9090
- 🔍 **Jaeger Tracing**: http://localhost:16686
- 📋 **Kibana Logs**: http://localhost:5601
- 🚨 **AlertManager**: http://localhost:9093

### 📈 Metrics & Prometheus

**API Metrics Available:**
```bash
# View all metrics
curl http://localhost:8081/metrics

# Key metric categories:
# HTTP metrics
http_requests_total              # Total HTTP requests
http_request_duration_seconds    # Request duration histograms
http_requests_in_flight         # Active requests

# Business metrics  
news_articles_total             # Total articles in system
news_searches_total             # Search requests
news_ai_requests_total          # AI-powered requests

# System metrics
go_goroutines                   # Active goroutines
go_memstats_*                  # Memory statistics
process_*                      # Process metrics

# Rate limiting metrics
rate_limit_exceeded_total       # Rate limit violations
rate_limit_requests_total       # Rate-limited requests
```

**Custom Metrics Examples:**
```go
// Business metrics in your code
var (
    articlesCreated = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "news_articles_created_total",
            Help: "Total number of articles created",
        },
        []string{"category", "author"},
    )
    
    searchDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "news_search_duration_seconds", 
            Help: "Search request duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"search_type"},
    )
)
```

### 📊 Grafana Dashboards

**Pre-built Dashboards:**

<details>
<summary>📊 <strong>News API Overview Dashboard</strong></summary>

**Panels Include:**
- 📈 Request rate (req/sec)
- ⏱️ Response time percentiles (P50, P90, P99)
- 🚨 Error rate percentage
- 🔍 Search performance metrics
- 💾 Database query performance
- 🔄 Cache hit/miss ratios
- 👥 Active users
- 📰 Content statistics

**Access:** http://localhost:3000/d/news-api-overview
</details>

<details>
<summary>🗄️ <strong>Database Performance Dashboard</strong></summary>

**Panels Include:**
- � Connection pool usage
- ⏱️ Query execution times
- 🚨 Slow query analysis
- 💾 Database size metrics
- 🔒 Lock contention
- 📊 Transaction statistics

**Access:** http://localhost:3000/d/database-performance
</details>

<details>
<summary>🤖 <strong>AI Services Dashboard</strong></summary>

**Panels Include:**
- 🧠 AI request volume
- 💰 OpenAI API costs
- ⏱️ AI response times
- 🔄 Fallback rates
- 📊 Search accuracy metrics
- 🚨 AI service health

**Access:** http://localhost:3000/d/ai-services
</details>

**Import Custom Dashboards:**
```bash
# Import dashboard from file
curl -X POST http://admin:admin@localhost:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @monitoring/grafana/dashboards/custom-dashboard.json
```

### � Distributed Tracing with Jaeger

**Tracing Features:**
- ✅ **Request Flow Visualization**: See complete request journey
- ✅ **Performance Bottlenecks**: Identify slow components
- ✅ **Error Tracking**: Trace errors across services
- ✅ **Dependency Mapping**: Understand service relationships
- ✅ **Custom Spans**: Add business-specific tracing

**Trace Your Requests:**
```bash
# Make a request with tracing
curl -H "X-Trace-ID: custom-trace-123" \
     "http://localhost:8081/api/news"

# View in Jaeger UI
# 1. Go to http://localhost:16686
# 2. Select "news-service" 
# 3. Search for trace ID "custom-trace-123"
```

**Custom Tracing in Code:**
```go
import "go.opentelemetry.io/otel"

func SearchArticles(ctx context.Context, query string) {
    tracer := otel.Tracer("news-service")
    ctx, span := tracer.Start(ctx, "search-articles")
    defer span.End()
    
    // Add custom attributes
    span.SetAttributes(
        attribute.String("search.query", query),
        attribute.String("search.type", "semantic"),
    )
    
    // Your business logic here
    results := performSearch(ctx, query)
    
    span.SetAttributes(
        attribute.Int("search.results_count", len(results)),
    )
}
```

### 📋 Centralized Logging

**Log Aggregation Setup:**
```bash
# Start ELK stack
make logging-up

# View logs in Kibana
# 1. Go to http://localhost:5601
# 2. Create index pattern: news-api-*
# 3. Explore logs in Discover tab
```

**Structured Logging Example:**
```go
import "github.com/sirupsen/logrus"

// Structured logging with fields
logger.WithFields(logrus.Fields{
    "user_id":    userID,
    "article_id": articleID,
    "action":     "article_view",
    "duration":   duration.Milliseconds(),
}).Info("Article viewed")
```

**Log Levels and Categories:**
```bash
# Application logs
level=info msg="Article created" article_id=123 user_id=456
level=error msg="Database connection failed" error="connection timeout"

# Access logs  
method=GET path="/api/news" status=200 duration=123ms user_agent="curl/7.68.0"

# Audit logs
level=info msg="User login" user_id=456 ip_address="192.168.1.1" success=true

# Performance logs
level=info msg="Slow query detected" query="SELECT * FROM articles" duration=2.5s
```

### 🚨 Alerting & Notifications

**Alerting Rules:**
```yaml
# monitoring/prometheus/alerts.yml
groups:
  - name: news-api-alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          
      - alert: DatabaseConnectionLow
        expr: database_connections_available < 5
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Database connection pool running low"
```

**Notification Channels:**
- 📧 **Email**: SMTP integration for critical alerts
- 💬 **Slack**: Webhook integration for team notifications
- 📱 **PagerDuty**: On-call escalation for production issues
- 📞 **Webhook**: Custom integrations

### � Health Monitoring

**Health Check Endpoints:**
```bash
# Basic health check
curl http://localhost:8081/health
# Response: {"status": "ok", "timestamp": "2025-06-03T14:30:00Z"}

# Detailed health check
curl http://localhost:8081/health/detailed
# Response includes database, redis, external services status

# Component-specific health
curl http://localhost:8081/health/database
curl http://localhost:8081/health/redis
curl http://localhost:8081/health/ai-services
```

**Health Monitoring Features:**
- ✅ **Deep Health Checks**: Test all critical dependencies
- ✅ **Circuit Breaker**: Prevent cascade failures
- ✅ **Graceful Degradation**: Maintain service when components fail
- ✅ **Automatic Recovery**: Self-healing capabilities
- ✅ **Health Metrics**: Export health status to monitoring

### 📊 Performance Monitoring

**Performance Metrics:**
```bash
# Response time analysis
curl http://localhost:8081/metrics | grep http_request_duration

# Memory usage monitoring  
curl http://localhost:8081/metrics | grep go_memstats

# Database performance
curl http://localhost:8081/metrics | grep database_query

# Cache performance
curl http://localhost:8081/metrics | grep cache_
```

**Performance Dashboards:**
- 📈 **Response Time Trends**: Track API performance over time
- 💾 **Memory Usage**: Monitor memory consumption patterns
- 🗄️ **Database Performance**: Query performance and optimization
- 🔄 **Cache Efficiency**: Cache hit rates and optimization
- 🌐 **Network I/O**: Network usage and bandwidth

### 🔍 Troubleshooting Guide

**Common Monitoring Issues:**

<details>
<summary>❌ <strong>Grafana Dashboard Not Loading</strong></summary>

```bash
# Check Grafana status
docker ps | grep grafana

# Check Grafana logs
docker logs grafana

# Restart Grafana
make grafana-restart

# Reset Grafana data
make grafana-reset
```
</details>

<details>
<summary>❌ <strong>Missing Metrics in Prometheus</strong></summary>

```bash
# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Verify metrics endpoint
curl http://localhost:8081/metrics

# Check Prometheus configuration
cat monitoring/prometheus/prometheus.yml

# Reload Prometheus config
curl -X POST http://localhost:9090/-/reload
```
</details>

<details>
<summary>❌ <strong>Traces Not Appearing in Jaeger</strong></summary>

```bash
# Check Jaeger services
docker ps | grep jaeger

# Verify tracing configuration
curl http://localhost:8081/health/tracing

# Check trace sampling
# Increase sampling rate for testing
export JAEGER_SAMPLER_PARAM=1.0
```
</details>

### 🛠️ Custom Monitoring Setup

**Add Custom Metrics:**
```go
// Create custom metric
var customMetric = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "news_custom_metric",
        Help: "Custom business metric",
    },
    []string{"label1", "label2"},
)

// Register metric
prometheus.MustRegister(customMetric)

// Update metric value
customMetric.WithLabelValues("value1", "value2").Set(42)
```

**Custom Grafana Panels:**
```json
{
  "title": "Custom Business Metric",
  "type": "graph",
  "targets": [
    {
      "expr": "news_custom_metric",
      "legendFormat": "{{label1}} - {{label2}}"
    }
  ]
}
```

### 📈 Monitoring Best Practices

**1. 🎯 Monitor What Matters:**
- User-facing metrics (response time, availability)
- Business metrics (articles created, searches performed)
- System health (CPU, memory, disk)
- Security metrics (failed logins, suspicious activity)

**2. 🚨 Set Up Meaningful Alerts:**
- Alert on symptoms, not causes
- Use appropriate thresholds
- Avoid alert fatigue
- Test alert delivery channels

**3. 📊 Create Actionable Dashboards:**
- Focus on key metrics
- Use consistent time ranges
- Add context and annotations
- Make dashboards team-accessible

**4. 🔍 Practice Monitoring Hygiene:**
- Regular dashboard reviews
- Clean up unused metrics
- Update alert thresholds
- Document monitoring procedures

1. Grafana will be available at `http://localhost:3000`
   - Default username: `admin`
   - Default password: `admin`

2. Open the "News API Dashboard" to see metrics including:
   - Request rates and durations
   - Error rates
   - Cache hit/miss ratios
   - Rate limiting events
   - Database operation performance

### Distributed Tracing

This application includes OpenTelemetry distributed tracing integrated with Jaeger:

1. Start the OpenTelemetry environment:
   ```
   ./scripts/start_otel_environment.sh
   ```

2. Verify tracing is working properly:
   ```
   ./scripts/verify_tracing.sh
   ```

3. Access the Jaeger UI at `http://localhost:16686`
   - Search for traces by service name: "news-service"
   - View detailed request flows and performance metrics
   - Analyze span details for debugging

4. See [Tracing Best Practices](./docs/tracing_best_practices.md) for guidelines on instrumenting code

### Testing

Run unit tests:
```
go test ./internal/...
```

Run integration tests:
```
go test ./internal/tests/...
```

Run load tests:
```
node ./scripts/load-test.js
```

### Database Migrations

The News API includes an automated database migration system:

1. Run migrations automatically on startup (default behavior)
   - Set `AUTO_MIGRATE=false` to disable

2. Run migrations manually:
   ```bash
   # Apply all pending migrations
   ./scripts/migrate.sh up
   
   # Roll back the last migration
   ./scripts/migrate.sh down 1
   
   # Check migration status
   ./scripts/migrate.sh status
   
   # Create a new migration
   ./scripts/migrate.sh create new_feature
   ```

3. Use Make commands:
   ```bash
   # Apply pending migrations
   make migrate-up
   
   # Create a new migration
   make migrate-create NAME=your_feature_name
   ```

4. For detailed information, see [Migration Guide](./docs/migrations.md)

### Database Seeding

The News API includes a comprehensive database seeding system that populates the database with sample data for development and testing. The seeding system is organized into categories and supports multiple environments.

#### Seeding Structure

```
scripts/seeds/
├── 01_core/           # Core data (users, categories, tags)
├── 02_content/        # Content data (articles, media)
├── 03_system/         # System data (settings, menus)
├── 04_interactions/   # User interactions (votes, bookmarks, follows)
└── 05_relationships/  # Relationship data (article-category mappings)
```

#### Quick Start

## 🎯 Modern News Features

The News API includes cutting-edge features designed for modern news consumption and engagement patterns.

### 🚨 Breaking News Banners

Dynamic, attention-grabbing banners for urgent news that require immediate visibility across the platform.

**✨ Key Features:**
- ⏰ **Time-Controlled Visibility**: Set start and end times for automatic display control
- 🎯 **Priority Levels**: Multiple priority levels for handling multiple breaking news
- 🎨 **Custom Styling**: Configurable colors, sizes, and animations
- 🔗 **Article Integration**: Direct linking to full news articles
- 👀 **Visibility Controls**: Fine-grained control over when and where banners appear
- 📊 **Analytics**: Track banner engagement and click-through rates

**🔌 API Endpoints:**
```bash
# Public endpoints
GET    /api/breaking-news              # Get all active breaking news banners
GET    /api/breaking-news/:id          # Get specific banner details

# Admin endpoints (authentication required)
POST   /admin/breaking-news            # Create new breaking news banner
PUT    /admin/breaking-news/:id        # Update existing banner
DELETE /admin/breaking-news/:id        # Remove banner
PATCH  /admin/breaking-news/:id/toggle # Toggle banner visibility
```

**📝 Example Usage:**
```bash
# Get active breaking news
curl "http://localhost:8081/api/breaking-news"

# Create breaking news (admin)
curl -X POST "http://localhost:8081/admin/breaking-news" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "URGENT: Major Development in Technology Sector",
    "description": "Breaking developments affecting global markets",
    "article_id": 123,
    "priority": "high",
    "background_color": "#ff0000",
    "text_color": "#ffffff",
    "start_time": "2025-06-03T10:00:00Z",
    "end_time": "2025-06-03T18:00:00Z",
    "is_active": true
  }'
```

### 📖 News Stories

Instagram/Facebook-style ephemeral stories providing a visual, engaging way to highlight trending content and behind-the-scenes updates.

**✨ Key Features:**
- ⏱️ **Time-Based Expiration**: Stories automatically expire after set duration
- 👀 **View Tracking**: Track which users have viewed each story
- 🎨 **Rich Media Support**: Images, videos, and interactive elements
- 📱 **Mobile-First Design**: Optimized for mobile consumption
- 🔄 **Story Chains**: Multiple stories grouped together
- 📊 **Engagement Analytics**: Track story performance and engagement

**🔌 API Endpoints:**
```bash
# Public endpoints
GET    /api/news-stories               # Get all active stories
GET    /api/news-stories/:id           # Get specific story details

# Authenticated endpoints
GET    /api/news-stories/unviewed      # Get stories user hasn't viewed
POST   /api/news-stories/:id/view      # Mark story as viewed

# Admin endpoints
POST   /admin/news-stories             # Create new story
PUT    /admin/news-stories/:id         # Update story
DELETE /admin/news-stories/:id         # Delete story
POST   /admin/news-stories/:id/media   # Add media to story
```

**📝 Example Usage:**
```bash
# Get unviewed stories for authenticated user
curl -H "Authorization: Bearer $USER_TOKEN" \
     "http://localhost:8081/api/news-stories/unviewed"

# Mark story as viewed
curl -X POST -H "Authorization: Bearer $USER_TOKEN" \
     "http://localhost:8081/api/news-stories/123/view"

# Create new story (admin)
curl -X POST "http://localhost:8081/admin/news-stories" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Behind the Scenes: Newsroom Update",
    "content": "Quick update from our editorial team",
    "media_url": "https://example.com/story-image.jpg",
    "media_type": "image",
    "duration": 86400,
    "article_id": 456,
    "background_color": "#1a1a1a",
    "text_color": "#ffffff"
  }'
```

### 🔴 Live News Streams

Real-time news coverage for ongoing events with continuous updates, perfect for breaking news, elections, sports events, and developing stories.

**✨ Key Features:**
- 🔄 **Real-Time Updates**: Continuous stream of updates as events unfold
- 📊 **Update Importance Levels**: Critical, important, and standard updates
- 🎛️ **Stream Lifecycle Management**: Draft → Live → Ended states
- ⭐ **Highlighted Streams**: Promote important live coverage
- 👥 **Multi-Author Support**: Multiple reporters can contribute updates
- 📈 **Live Analytics**: Real-time viewer counts and engagement metrics
- 🔔 **Push Notifications**: Alert subscribers to important updates

**🔌 API Endpoints:**
```bash
# Public endpoints
GET    /api/live-news                  # Get all active live streams
GET    /api/live-news/:id              # Get specific stream with updates
GET    /api/live-news/:id/updates      # Get paginated updates for stream

# Authenticated endpoints  
POST   /api/live-news/:id/follow       # Follow a live stream
DELETE /api/live-news/:id/follow       # Unfollow a live stream

# Admin endpoints
POST   /admin/live-news                # Create new live stream
PUT    /admin/live-news/:id            # Update stream details
DELETE /admin/live-news/:id            # Delete stream
POST   /admin/live-news/:id/updates    # Add update to stream
PUT    /admin/live-news/:id/updates/:update_id  # Edit stream update
PATCH  /admin/live-news/:id/status     # Change stream status
```

**🎛️ Stream States:**
- **📝 Draft**: Preparation phase, not visible to public
- **🔴 Live**: Active stream with real-time updates
- **⏹️ Ended**: Completed stream, archived for reference
- **⭐ Highlighted**: Featured prominently on homepage

**📝 Example Usage:**
```bash
# Get active live streams
curl "http://localhost:8081/api/live-news"

# Get specific live stream with recent updates
curl "http://localhost:8081/api/live-news/789?limit=20"

# Create live stream (admin)
curl -X POST "http://localhost:8081/admin/live-news" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Election Results 2025 - Live Coverage",
    "description": "Real-time coverage of election results and analysis",
    "category": "politics",
    "status": "live",
    "is_highlighted": true,
    "tags": ["election", "politics", "breaking"]
  }'

# Add update to live stream
curl -X POST "http://localhost:8081/admin/live-news/789/updates" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "First results coming in from major cities showing early trends",
    "importance": "important",
    "author_name": "Political Correspondent",
    "media_url": "https://example.com/election-map.jpg"
  }'
```

### 🎥 Video Integration

Comprehensive video content management with support for multiple sources and advanced playback features.

**✨ Key Features:**
- 📺 **Multiple Video Sources**: YouTube, Vimeo, direct uploads, live streams
- 🎬 **Video Metadata**: Titles, descriptions, thumbnails, durations
- 📊 **Playback Analytics**: View counts, engagement metrics, watch time
- 🔄 **Automatic Transcoding**: Multiple quality levels and formats
- 📱 **Responsive Players**: Adaptive streaming for all devices
- 🎯 **Video SEO**: Structured data for search engine optimization

**🔌 API Endpoints:**
```bash
# Public endpoints
GET    /api/videos                     # Get all public videos
GET    /api/videos/:id                 # Get specific video details
GET    /api/videos/trending            # Get trending videos

# Admin endpoints
POST   /admin/videos                   # Upload/create new video
PUT    /admin/videos/:id               # Update video details
DELETE /admin/videos/:id               # Delete video
POST   /admin/videos/:id/thumbnail     # Upload custom thumbnail
```

### 📊 Content Analytics & Insights

Advanced analytics for understanding reader behavior and content performance.

**📈 Analytics Features:**
- 👀 **Article Views**: Detailed view tracking with time-based analysis
- 📊 **User Engagement**: Comments, shares, bookmarks, reading time
- 🔍 **Search Analytics**: Popular search terms and result effectiveness
- 📱 **Device Analytics**: Desktop vs mobile usage patterns
- 🌍 **Geographic Insights**: Content performance by region
- ⏰ **Time-Based Analysis**: Peak reading times and seasonal trends

**🔌 Analytics Endpoints:**
```bash
# Public analytics
GET    /api/analytics/popular          # Most popular articles
GET    /api/analytics/trending         # Trending topics

# Admin analytics (authentication required)
GET    /admin/analytics/overview       # Dashboard overview
GET    /admin/analytics/articles/:id   # Article-specific analytics
GET    /admin/analytics/users          # User behavior analysis
GET    /admin/analytics/search         # Search analytics
```

### 🏷️ Advanced Tagging & Categorization

Intelligent content organization with AI-powered auto-tagging and dynamic categorization.

**✨ Features:**
- 🤖 **AI Auto-Tagging**: Automatic tag suggestions based on content analysis
- 🏗️ **Hierarchical Categories**: Multi-level category structures
- 🔍 **Tag-Based Search**: Enhanced search using content tags
- 📊 **Tag Analytics**: Track tag performance and popularity
- 🎯 **Content Recommendations**: AI-driven related content suggestions
- 🔄 **Dynamic Categorization**: Categories adapt based on content trends

### 🔔 Push Notifications & Alerts

Real-time notification system for keeping users engaged and informed.

**🔔 Notification Types:**
- 🚨 **Breaking News Alerts**: Immediate notifications for urgent news
- 📰 **Personalized Updates**: Content based on user preferences
- 🔴 **Live Stream Notifications**: Updates from followed live streams
- 📊 **Weekly Summaries**: Personalized weekly news digest
- 🎯 **Category Alerts**: Notifications for specific news categories

**🔌 Notification Endpoints:**
```bash
# User notification management
GET    /api/notifications              # Get user notifications
POST   /api/notifications/subscribe    # Subscribe to notifications
DELETE /api/notifications/subscribe    # Unsubscribe
PUT    /api/notifications/preferences  # Update notification preferences

# Admin notification management
POST   /admin/notifications/broadcast  # Send broadcast notification
POST   /admin/notifications/targeted   # Send targeted notifications
```

### 📱 Progressive Web App (PWA) Features

Modern web app capabilities for enhanced user experience.

**🚀 PWA Features:**
- 📱 **App-Like Experience**: Native app feel in web browser
- 🔄 **Offline Reading**: Cache articles for offline access
- 🔔 **Push Notifications**: Browser-based notifications
- 📥 **Add to Home Screen**: Install as app on mobile devices
- ⚡ **Fast Loading**: Service worker optimization
- 🔄 **Background Sync**: Sync data when connection returns

### 🌐 Internationalization (i18n)

Multi-language support for global news distribution.

**🗣️ Supported Languages:**
- 🇺🇸 **English**: Primary language with full feature support
- 🇪🇸 **Spanish (Español)**: Complete translation and localization
- 🇹🇷 **Turkish (Türkçe)**: Full Turkish language support
- 🔄 **Easy Extension**: Framework for adding new languages

**🔌 i18n Endpoints:**
```bash
# Language management
GET    /api/languages                  # Get supported languages
GET    /api/content/:id/translations   # Get article translations
POST   /admin/content/:id/translate    # Create translation

# Localized content
GET    /api/news?lang=tr              # Get Turkish news
GET    /api/categories?lang=es         # Get Spanish categories
```

These modern features make the News API a comprehensive solution for contemporary news platforms, supporting everything from traditional article publishing to modern social media-style content consumption patterns.

- `GET /api/breaking-news` - Get all active breaking news banners
- `POST /admin/breaking-news` - Create a new breaking news banner (Admin only)
- `PUT /admin/breaking-news/:id` - Update a breaking news banner (Admin only)
- `DELETE /admin/breaking-news/:id` - Delete a breaking news banner (Admin only)

**Key Features:**

- Time-controlled visibility (start/end time)
- Priority levels for multiple banners
- Custom colors and styling
- Article linking
- Visibility controls

### News Stories

Instagram/Facebook-style news stories provide a visual, ephemeral way to highlight content.

**API Endpoints:**

- `GET /api/news-stories` - Get all active news stories
- `GET /api/news-stories/:id` - Get a specific story
- `GET /api/news-stories/unviewed` - Get stories the user hasn't viewed yet (authenticated)
- `POST /admin/news-stories` - Create a new story (Admin only)
- `PUT /admin/news-stories/:id` - Update a story (Admin only)
- `DELETE /admin/news-stories/:id` - Delete a story (Admin only)

**Key Features:**

- Short-format visual content
- Time-based expiration
- User view tracking
- Custom styling options
- Article linking

### Live News Streams

Real-time news coverage for ongoing events with continuous updates.

**API Endpoints:**

- `GET /api/live-news` - Get all active live news streams
- `GET /api/live-news/:id` - Get a specific live stream with its updates
- `GET /api/live-news/:id/updates` - Get paginated updates for a live stream
- `POST /admin/live-news` - Create a new live stream (Admin only)
- `PUT /admin/live-news/:id` - Update a live stream (Admin only)
- `DELETE /admin/live-news/:id` - Delete a live stream (Admin only)
- `POST /admin/live-news/:id/updates` - Add an update to a live stream (Admin only)

## 📚 Comprehensive API Documentation

### 🎯 Documentation Overview

The News API provides multiple layers of documentation to support developers at every level:

| Documentation Type | Purpose | Audience | Access |
|-------------------|---------|----------|---------|
| **🔧 Interactive Swagger UI** | Live API testing | Developers | http://localhost:8081/swagger/ |
| **📖 Developer Guides** | Implementation help | All developers | [docs/](./docs/) |
| **🏗️ Architecture Docs** | System design | Senior developers | [docs/guides/](./docs/guides/) |
| **� Quick Start** | Fast setup | New developers | This README |

### 🔄 Interactive API Documentation (Swagger UI)

**Environment-Specific Access Points:**
- **🔧 Development**: http://localhost:8081/swagger/index.html
- **🧪 Testing**: http://localhost:8082/swagger/index.html  
- **🚀 Production**: http://localhost:8080/swagger/index.html

**Swagger UI Features:**
- ✅ **Interactive Testing**: Test endpoints directly from browser
- ✅ **Authentication Support**: JWT token authentication built-in
- ✅ **Request/Response Examples**: Complete examples for all endpoints
- ✅ **Schema Validation**: Real-time request validation
- ✅ **Multi-Environment**: Switch between dev/test/prod environments
- ✅ **Export Options**: Download OpenAPI spec for other tools

### 🛠️ Documentation Generation & Updates

**Automated Documentation Workflow:**
```bash
# 🐳 Generate/update documentation (Docker-based)
make docs                    # Generate docs using Docker
make docs-local             # Generate docs locally (requires swag CLI)
make docs-validate          # Validate OpenAPI specification
make docs-lint              # Check documentation quality

# 🚀 Serve documentation standalone
make docs-serve             # Start Swagger UI container
# ✅ Available at: http://localhost:8091
make docs-stop              # Stop documentation container
```

**Docker-Based Documentation Benefits:**
- 🔄 **Zero Dependencies**: No need to install swag CLI locally
- 🚀 **Consistent Generation**: Same results across all environments
- � **Isolated Service**: Documentation server runs independently
- 📝 **Hot Reload**: Regenerate and refresh for live updates
- 🐳 **CI/CD Ready**: Perfect for automated documentation builds

### 📁 Documentation Structure & Organization

**Source Documentation (Tracked in Git):**
```
cmd/api/docs/
├── docs.go              # Generated OpenAPI Go definitions
├── swagger.json         # OpenAPI specification (JSON)
└── swagger.yaml         # OpenAPI specification (YAML)
```

**Documentation Categories:**
```
docs/
├── 🚀 api/              # API-specific documentation
│   ├── authentication.md
│   ├── rate-limiting.md
│   └── examples/
├── 📚 guides/           # Developer implementation guides
│   ├── getting-started.md
│   ├── advanced-features.md
│   └── troubleshooting.md
├── 🏗️ architecture/     # System design documentation
│   ├── system-overview.md
│   ├── database-design.md
│   └── security-model.md
├── 🚀 deployment/       # Deployment and operations
│   ├── docker-setup.md
│   ├── kubernetes.md
│   └── monitoring.md
└── 📊 reports/          # Test and performance reports
    ├── test-coverage.md
    ├── performance.md
    └── security-audit.md
```

### 🎯 API Endpoint Categories

**📰 Content Management Endpoints:**
```bash
# Articles
GET    /api/news                    # List articles
POST   /api/articles               # Create article (auth)
GET    /api/articles/:id           # Get specific article
PUT    /api/articles/:id           # Update article (auth)
DELETE /api/articles/:id           # Delete article (admin)

# Categories & Tags
GET    /api/categories             # List categories
GET    /api/tags                   # List tags
POST   /admin/categories           # Create category (admin)
POST   /admin/tags                 # Create tag (admin)
```

**🔍 Search & Discovery Endpoints:**
```bash
# Search
GET    /api/search                 # Basic text search
GET    /api/search/semantic        # AI-powered semantic search
GET    /api/search/advanced        # Advanced search with filters

# Recommendations
GET    /api/recommendations        # Personalized recommendations
GET    /api/trending               # Trending articles
GET    /api/popular                # Popular content
```

**👤 User Management Endpoints:**
```bash
# Authentication
POST   /api/auth/register          # User registration
POST   /api/auth/login             # User login
POST   /api/auth/refresh           # Token refresh
POST   /api/auth/logout            # User logout

# User Profile
GET    /api/profile                # Get user profile (auth)
PUT    /api/profile                # Update profile (auth)
GET    /api/bookmarks              # User bookmarks (auth)
POST   /api/articles/:id/bookmark  # Bookmark article (auth)
```

**⚡ Real-Time Features Endpoints:**
```bash
# Breaking News
GET    /api/breaking-news          # Active breaking news
POST   /admin/breaking-news        # Create breaking news (admin)

# Live News Streams  
GET    /api/live-news              # Active live streams
GET    /api/live-news/:id/updates  # Stream updates
POST   /admin/live-news            # Create live stream (admin)

# News Stories
GET    /api/news-stories           # Active stories
GET    /api/news-stories/unviewed  # Unviewed stories (auth)
POST   /admin/news-stories         # Create story (admin)
```

**📊 Analytics & Monitoring Endpoints:**
```bash
# System Health
GET    /health                     # Basic health check
GET    /health/detailed            # Detailed system health
GET    /metrics                    # Prometheus metrics

# Rate Limiting
GET    /api/limits                 # Current rate limits
GET    /api/search/limits          # Search-specific limits

# Analytics (Admin)
GET    /admin/analytics/overview   # System overview
GET    /admin/analytics/content    # Content performance
GET    /admin/analytics/users      # User behavior
```

### 🔐 Authentication & Authorization Documentation

**JWT Token Usage:**
```bash
# Get authentication token
curl -X POST "http://localhost:8081/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'

# Use token in requests
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     "http://localhost:8081/api/profile"
```

**Permission Levels:**
- **🌐 Public**: No authentication required
- **👤 Authenticated**: Valid JWT token required
- **✏️ Author**: Can edit own content
- **🛡️ Admin**: Full system access

### 📝 Request/Response Examples

**Article Creation Example:**
```bash
# Create new article (authenticated)
curl -X POST "http://localhost:8081/api/articles" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Revolutionary AI Technology Breakthrough",
    "content": "Detailed article content here...",
    "summary": "Brief summary of the breakthrough",
    "category_id": 1,
    "tags": ["AI", "Technology", "Innovation"],
    "featured_image": "https://example.com/image.jpg",
    "status": "published"
  }'

# Expected Response (201 Created)
{
  "id": 123,
  "title": "Revolutionary AI Technology Breakthrough",
  "slug": "revolutionary-ai-technology-breakthrough",
  "author": {
    "id": 456,
    "name": "John Doe",
    "email": "john@example.com"
  },
  "created_at": "2025-06-03T14:30:00Z",
  "updated_at": "2025-06-03T14:30:00Z",
  "view_count": 0,
  "status": "published"
}
```

**Semantic Search Example:**
```bash
# AI-powered semantic search
curl "http://localhost:8081/api/search/semantic?q=artificial%20intelligence%20breakthrough&limit=10"

# Expected Response
{
  "results": [
    {
      "id": 123,
      "title": "Revolutionary AI Technology Breakthrough",
      "relevance_score": 0.95,
      "summary": "Brief summary...",
      "category": "Technology",
      "published_at": "2025-06-03T14:30:00Z"
    }
  ],
  "metadata": {
    "total_results": 1,
    "search_type": "semantic",
    "ai_model_used": "gpt-4",
    "response_time_ms": 234,
    "rate_limit_remaining": 49
  }
}
```

### 🔧 SDK & Client Libraries

**Official Client Libraries:**
- **JavaScript/TypeScript**: `npm install @news-api/js-client`
- **Python**: `pip install news-api-client`
- **Go**: `go get github.com/news-api/go-client`
- **PHP**: `composer require news-api/php-client`

**Community Libraries:**
- **Ruby**: `gem install news-api-ruby`
- **Java**: Available via Maven Central
- **C#**: Available via NuGet
- **Swift**: Available via Swift Package Manager

### 📖 Developer Resources

**Essential Developer Links:**
- � **Quick Start Guide**: [docs/DEVELOPER_GUIDE.md](./docs/DEVELOPER_GUIDE.md)
- 🏗️ **Architecture Overview**: [docs/guides/ARCHITECTURE.md](./docs/guides/ARCHITECTURE.md)
- 🔒 **Security Guidelines**: [docs/guides/SECURITY.md](./docs/guides/SECURITY.md)
- 🧪 **Testing Guide**: [docs/guides/TESTING.md](./docs/guides/TESTING.md)
- 🚀 **Deployment Guide**: [deployments/README.md](./deployments/README.md)
- 🎯 **Best Practices**: [docs/guides/BEST_PRACTICES.md](./docs/guides/BEST_PRACTICES.md)

**Advanced Topics:**
- 🤖 **AI Integration**: [docs/guides/AI_INTEGRATION.md](./docs/guides/AI_INTEGRATION.md)
- 📊 **Performance Optimization**: [docs/guides/PERFORMANCE.md](./docs/guides/PERFORMANCE.md)
- 🔄 **Migration Guide**: [docs/migration/README.md](./docs/migration/README.md)
- 🌐 **Internationalization**: [docs/guides/I18N.md](./docs/guides/I18N.md)

### 💡 Documentation Best Practices

**For API Consumers:**
1. 📖 **Start with Swagger UI** for interactive exploration
2. 🔐 **Understand Authentication** before making requests
3. 🚦 **Check Rate Limits** to avoid throttling
4. 📊 **Monitor Usage** through analytics endpoints
5. 🔄 **Handle Errors Gracefully** with proper error codes

**For Contributors:**
1. 📝 **Update Documentation** with code changes
2. ✅ **Validate OpenAPI Spec** before committing
3. 🧪 **Test Examples** to ensure they work
4. � **Follow Documentation Standards** for consistency
5. 🔄 **Keep Documentation Current** with regular reviews

## 🤝 Contributing & Community

### 🎯 How to Contribute

We welcome contributions from developers of all skill levels! Here's how you can help improve the News API:

**🐛 Bug Reports & Feature Requests:**
- 📝 Use GitHub Issues for bug reports
- 🎯 Use feature request templates
- 🔍 Search existing issues before creating new ones
- 📊 Provide detailed reproduction steps

**💻 Code Contributions:**
- 🍴 Fork the repository
- 🌿 Create feature branches
- ✅ Write tests for new features
- 📝 Update documentation
- 🔄 Submit pull requests

**📚 Documentation Improvements:**
- 📖 Fix typos and improve clarity
- 🎯 Add examples and use cases
- 🌐 Translate documentation
- 📊 Update outdated information

### 🛠️ Development Guidelines

**Code Quality Standards:**
- ✅ Follow Go best practices and conventions
- 🧪 Maintain test coverage above 80%
- 📝 Write clear, self-documenting code
- 🔍 Use linters and code formatters
- 📊 Profile performance-critical code

**Git Workflow:**
```bash
# 1. Fork and clone the repository
git clone https://github.com/yourusername/news-api.git
cd news-api

# 2. Create a feature branch
git checkout -b feature/your-feature-name

# 3. Make changes and commit
git add .
git commit -m "feat: add new feature description"

# 4. Push and create pull request
git push origin feature/your-feature-name
```

**Commit Message Convention:**
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `style:` - Code style changes
- `refactor:` - Code refactoring
- `test:` - Test additions or changes
- `chore:` - Build process or auxiliary tool changes

### 🌟 Community & Support

**Get Help & Connect:**
- � **Discord**: Join our developer community
- 📧 **Email**: developer-support@news-api.com
- 🐛 **GitHub Issues**: Bug reports and feature requests
- 📚 **Stack Overflow**: Tag questions with `news-api`
- 📖 **Wiki**: Community-maintained documentation

**Stay Updated:**
- 📰 **Release Notes**: Follow GitHub releases
- 📝 **Blog**: Technical articles and updates
- 🐦 **Twitter**: @NewsAPIProject for announcements
- 📧 **Newsletter**: Monthly developer updates

### 📜 License & Legal

**Open Source License:**
This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

**What this means:**
- ✅ **Commercial Use**: Use in commercial projects
- ✅ **Modification**: Modify the source code
- ✅ **Distribution**: Distribute the software
- ✅ **Private Use**: Use privately
- ❌ **Liability**: No warranty or liability
- ❌ **Trademark Use**: No trademark rights granted

**Third-Party Licenses:**
- All dependencies maintain their respective licenses
- See `go.mod` for complete dependency list
- License compatibility verified for all components

---

## 🎉 Thank You!

Thank you for choosing the News API! Whether you're building a news website, mobile app, or integrating news functionality into your existing platform, we're excited to see what you create.

**🚀 Ready to get started?** 
1. 📖 Review the [Quick Start Guide](#-quick-start-guide)
2. 🔧 Set up your [development environment](#-development-workflow)
3. 📚 Explore the [API documentation](#-comprehensive-api-documentation)
4. 🎯 Try the [modern news features](#-modern-news-features)

**💡 Questions or need help?**
- 📚 Check our comprehensive documentation
- 💬 Join the developer community
- 🐛 Report issues on GitHub
- 📧 Contact our support team

**🤝 Want to contribute?**
- 🍴 Fork the repository
- 🌟 Star the project if you find it useful
- 📝 Improve documentation
- 🐛 Report bugs and suggest features

Built with ❤️ by the News API Team | © 2025 | [MIT License](LICENSE)
- 📁 `cmd/api/docs/` - Source documentation (tracked in git)
- 📁 `docs/` - Docker-generated duplicates (ignored in git)
- 🎯 Single source of truth maintained in `cmd/api/docs/`

**Important**: Swagger docs are generated in `cmd/api/docs/` and served directly by the API. Manual editing of these files is not recommended as they will be overwritten.

### 📖 Additional Documentation

- **API Reference**: See generated Swagger documentation above
- **Developer Guide**: [docs/DEVELOPER_GUIDE.md](./docs/DEVELOPER_GUIDE.md)
- **Project Organization**: [docs/PROJECT_ORGANIZATION_GUIDE.md](./docs/PROJECT_ORGANIZATION_GUIDE.md)

## License

This project is licensed under the MIT License - see the LICENSE file for details.
