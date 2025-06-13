# Project Organization Guide

This document explains the folder structure of the News API project and how to work with each component.

## Folder Structure

The News API project is organized into the following key directories:

### Core Code

- `cmd/` - Application entry points
  - `api/` - Main API server code
  - `migrate/` - Database migration tool

- `internal/` - Private application code
  - `auth/` - Authentication logic
  - `cache/` - Redis cache integration
  - `database/` - Database connections and ORM setup
  - `handlers/` - HTTP request handlers
  - `models/` - Data models and structures
  - `services/` - Business logic
  - `tests/` - Integration test suite
  - ... and more

### Deployments and Configuration

- `docker/` - Docker configuration
  - `Dockerfile` - Main production container
  - `Dockerfile.dev` - Development container
  - `docker-compose.yml` - Production container setup
  - `docker-compose-dev.yml` - Development container setup
  - `docker-compose-otel.yml` - OpenTelemetry setup

- `kubernetes/` - Kubernetes deployment files
  - `deployment.yaml` - Main deployment configuration
  - `monitoring.yaml` - Monitoring configuration

### Tools and Testing

- `debug/` - Debug tools
  - `debug_client.go` - Debug client for API testing
  - `debug_server.go` - Debug server to intercept and log requests

- `scripts/` - Utility scripts
  - `docker-helper.sh` - Helper for Docker commands
  - `test_*.sh` - Various test scripts
  - ... and more

- `tests/` - API test suite
  - `test_*.go` - Various API test programs
  - `test_observability.sh` - Observability stack test
  - `run_tests.sh` - Script to run all tests

### Database

- `migrations/` - SQL migration files
  - `000001_create_users_table.up.sql`
  - `000001_create_users_table.down.sql`
  - ... and more

### Documentation

- `docs/` - Project documentation
  - API documentation
  - Implementation guides
  - Migration guides
  - Tracing guides

## Using Makefile

We've provided Makefile targets to work with each component:

```
make debug-server      # Run the debug server
make debug-client      # Run the debug client
make test-api          # Run API tests
make test-observability # Run observability tests
make docker-helper     # Run Docker helper script
```

See `make help` for a complete list of available commands.

## Docker Helper

We've created a Docker helper script to make working with Docker easier:

```bash
# Run the Docker helper directly
./scripts/docker-helper.sh

# Or use the Makefile target
make docker-helper
```

Commands:
- `build` - Build production image
- `up-dev` - Start development services
- `down` - Stop production services
- `logs api` - View logs for the API service

## Keeping the Project Organized

When adding new files:
1. Add them to the appropriate folder based on their purpose
2. Update documentation if needed
3. Consider adding Makefile targets for new tools
