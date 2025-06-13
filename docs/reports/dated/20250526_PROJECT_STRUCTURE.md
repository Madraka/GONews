# News API Project Structure

This document outlines the organized file structure for the News API project.

## Directory Structure

```
News/
├── README.md                    # Main project documentation
├── Makefile                     # Build and deployment commands
├── docker-compose.yml           # Main Docker configuration
├── go.mod & go.sum             # Go modules
├── Dockerfile                   # Production Docker build
├── .env                        # Environment variables
├── .gitignore                  # Git ignore rules
│
├── cmd/                        # Application entry points
│   └── api/
│       └── main.go
│
├── internal/                   # Private application code
│   ├── handlers/               # HTTP handlers
│   ├── middleware/             # HTTP middleware
│   ├── models/                 # Data models
│   ├── routes/                 # Route definitions
│   ├── config/                 # Configuration
│   └── utils/                  # Utility functions
│
├── api/                        # API related files
│   └── v1/                     # API version 1
│
├── docs/                       # Documentation
│   ├── api/                    # API documentation
│   │   ├── swagger.json
│   │   ├── swagger.yaml
│   │   └── docs.go
│   ├── guides/                 # Implementation guides
│   ├── reports/                # Test and deployment reports
│   └── migration/              # Migration documentation
│
├── tests/                      # Test files
│   ├── integration/            # Integration tests
│   ├── unit/                   # Unit tests
│   └── scripts/                # Test scripts
│
├── scripts/                    # Build and utility scripts
│   ├── build/                  # Build scripts
│   ├── deploy/                 # Deployment scripts
│   └── test/                   # Test automation scripts
│
├── monitoring/                 # Observability configuration
│   ├── grafana/               # Grafana dashboards
│   ├── prometheus/            # Prometheus configuration
│   └── jaeger/                # Jaeger tracing config
│
├── deployments/               # Deployment configurations
│   ├── docker/                # Docker configurations
│   ├── kubernetes/            # Kubernetes manifests
│   └── dev/                   # Development environment
│
├── migrations/                # Database migrations
│
└── tools/                     # Development tools
    └── vendors/               # Vendor dependencies
```

## File Organization Rules

### Reports and Documentation
- All test reports moved to `docs/reports/`
- API documentation in `docs/api/`
- Implementation guides in `docs/guides/`

### Scripts and Tools
- Test scripts in `tests/scripts/`
- Build scripts in `scripts/build/`
- Utility scripts in `scripts/`

### Configuration Files
- Docker files organized in `deployments/docker/`
- Environment-specific configs in appropriate folders

### Development Files
- Debug and temporary files removed or organized properly
- Clear separation between development and production configurations
