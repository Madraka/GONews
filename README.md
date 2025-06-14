# 📰 News API

> **A modern, production-ready RESTful API for managing news articles with advanced features**

![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Open Source](https://img.shields.io/badge/open%20source-❤️-red.svg)

## 🌟 What's New - Now Open Source!

This project has been migrated to **open source**! 🎉 We welcome contributions from the community. 
**[→ Learn about the migration](./docs/OPEN_SOURCE_MIGRATION.md)**

## ✨ Key Features

### 🚀 Core Capabilities
- **📝 Article Management** - Full CRUD operations with versioning
- **👥 User System** - Authentication, roles, and profiles  
- **🔍 Smart Search** - AI-powered semantic search with OpenAI
- **📱 Real-time** - WebSocket support for live updates
- **🌍 Multi-language** - English, Spanish, Turkish support
- **📊 Analytics** - Comprehensive metrics and tracking

### 🎯 Modern Features
- **🚨 Breaking News** - Priority alerts and notifications
- **� Stories System** - Ephemeral content like Instagram stories
- **🎥 Video Support** - Embedded video content and processing
- **🔴 Live Streams** - Real-time event coverage
- **📱 Mobile-First** - Optimized for mobile applications

### 🔒 Enterprise-Ready
- **🛡️ Security** - JWT authentication, role-based access
- **⚡ Performance** - Redis caching, HTTP/2 support
- **📊 Observability** - Prometheus metrics, Grafana dashboards
- **� DevOps** - Docker, Kubernetes, CI/CD ready

## 🚀 Quick Start

### Prerequisites
- Go 1.24+
- PostgreSQL 15+
- Redis 7+
- Docker (optional)

### Get Running in 2 Minutes

```bash
# 1. Clone the repository
git clone https://github.com/your-username/news-api.git
cd news-api

# 2. Start services with Docker
make dev-up

# 3. Run the API
make dev

# 4. Visit API documentation
open http://localhost:8081/swagger/index.html
```

### Manual Setup

```bash
# 1. Install dependencies
go mod download

# 2. Configure environment
cp .env.example .env  # Edit with your settings

# 3. Start database
make db-up

# 4. Run migrations
make migrate-up

# 5. Start API server
make run
```

## 📚 Documentation

| Resource | Description |
|----------|-------------|
| **[� Documentation Hub](./docs/README.md)** | Complete documentation index |
| **[🏃 Developer Guide](./docs/DEVELOPER_GUIDE.md)** | Setup and development workflow |
| **[🌟 Open Source Guide](./docs/OPEN_SOURCE_MIGRATION.md)** | How to contribute |
| **[📋 API Reference](./docs/api_documentation.html)** | Complete API documentation |

## 🛠️ Development

### Environment Management
```bash
# Development environment (recommended)
make env-dev && make dev-up && make dev

# Testing environment  
make env-test && make test-all

# Production environment
make env-prod && make prod-deploy
```

### Key Commands
```bash
make build          # Build application
make test           # Run tests
make migrate-up     # Run database migrations
make docs           # Generate API documentation
make docker-build   # Build Docker image
```

## 🏗️ Project Structure

```
News/
├── � README.md                 # Project overview (you are here)
├── 📚 docs/                     # Complete documentation
│   ├── README.md               # Documentation index
│   ├── DEVELOPER_GUIDE.md      # Development setup
│   └── api_documentation.html  # API reference
├── 🚀 cmd/                     # Application entry points
├── 🔒 internal/                # Private application code
├── 🧪 tests/                   # Comprehensive test suite
├── 🐳 deployments/            # Docker & Kubernetes configs
├── 📊 monitoring/             # Observability setup
└── �️ scripts/               # Build and utility scripts
```

**[→ See complete project structure in docs](./docs/README.md)**

## 🌍 API Overview

### Core Endpoints
- `GET /api/articles` - List articles with pagination
- `POST /api/articles` - Create new article (auth required)
- `GET /api/search` - Semantic search with AI
- `GET /api/breaking` - Breaking news alerts
- `GET /api/stories` - Stories feed

### Advanced Features
- `GET /api/videos` - Video content management
- `GET /api/live` - Live stream endpoints
- `WebSocket /ws` - Real-time updates
- `GET /metrics` - Prometheus metrics

**[→ Complete API documentation](./docs/api_documentation.html)**

## 🤝 Contributing

We welcome contributions! This is an open source project.

1. **[Read the Contributing Guide](./CONTRIBUTING.md)**
2. **Fork the repository**
3. **Create a feature branch**
4. **Submit a pull request**

### Areas for Contribution
- � Bug fixes and stability improvements
- ✨ New features and enhancements  
- 📚 Documentation improvements
- 🧪 Test coverage expansion
- 🌍 Internationalization (new languages)
- 🎨 UI/UX improvements for API responses

## 📊 Status & Roadmap

### ✅ Production Ready
- Article Management System
- User Authentication & Authorization
- REST API with Swagger Documentation
- Multi-language Support
- Redis Caching & PostgreSQL

### 🚧 Beta Features  
- AI-Powered Semantic Search
- Video Content Support
- Real-time WebSocket Updates
- HTTP/2 Implementation

### 🔮 Planned Features
- GraphQL API
- Advanced Analytics Dashboard  
- Mobile Push Notifications
- Content Recommendation Engine
- Advanced Search Filters

## � Deployment

### Quick Deploy
```bash
# Development
make dev-deploy

# Production
make prod-deploy
```

### Supported Platforms
- 🐳 **Docker** - Complete containerization
- ☸️ **Kubernetes** - Production orchestration  
- 🌊 **Docker Swarm** - Simple clustering
- 🖥️ **Bare Metal** - Traditional deployment

**[→ Deployment guides](./docs/README.md)**

## 📈 Monitoring

- **📊 Metrics**: Prometheus + Grafana dashboards
- **🔍 Tracing**: OpenTelemetry integration
- **📝 Logging**: Structured logging with levels
- **⚡ Performance**: Built-in profiling endpoints

Access monitoring at `http://localhost:3000` (Grafana)

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with ❤️ using Go and modern technologies
- Thanks to all contributors and the open source community
- Powered by OpenAI for semantic search capabilities

---

**⭐ Star this repository if you find it useful!**

