# Deployment Guide

This directory contains deployment configurations for the News API across different environments.

## 📁 Directory Structure

```
deployments/
├── README.md                    # This comprehensive deployment guide
├── dev/                        # Development environment (port 8081)
│   ├── docker-compose-dev.yml
│   └── .env.dev
├── test/                       # Testing environment (port 8082)
│   ├── docker-compose-test.yml
│   └── .env.test
├── prod/                       # Production environment (port 8080)
│   ├── docker-compose-prod.yml
│   └── .env.prod
├── dockerfiles/                # Docker build files
│   ├── Dockerfile.test         # Test service
│   ├── Dockerfile.dev          # Development build
│   └── Dockerfile.prod         # Production build
├── scripts/                    # Deployment and maintenance scripts
│   ├── deploy.sh              # Main deployment script
│   ├── backup.sh              # Database backup script
│   ├── migrate.sh             # Database migration script
│   └── health-check.sh        # Service health check script
└── k8s/                       # Kubernetes manifests (future)
```

## Port Allocation

All environments can run simultaneously without conflicts:

- **Production (prod)**: API:8080, DB:5432, Redis:6379, Jaeger:16686
- **Development (dev)**: API:8081, DB:5434, Redis:6380, Jaeger:16687  
- **Testing (test)**: API:8082, DB:5432, Redis:6379, Jaeger:16688

## Usage

## 🚀 Quick Start

### Development Environment
```bash
# Start development environment (port 8081)
cd deployments/dev
docker-compose -f docker-compose-dev.yml up -d

# Access API at: http://localhost:8081
# Database: localhost:5434
# Redis: localhost:6380
```

### Testing Environment
```bash
# Start testing environment (port 8082)
cd deployments/test
docker-compose -f docker-compose-test.yml up -d

# Access API at: http://localhost:8082
# Database: localhost:5432
# Redis: localhost:6379
```

### Production Environment
```bash
# Start production environment (port 8080)
cd deployments/prod
docker-compose -f docker-compose-prod.yml up -d

# Access API at: http://localhost:8080
# Database: localhost:5432
# Redis: localhost:6379
```

## 🔧 Advanced Deployment Options

### Using Deployment Scripts (Recommended)

```bash
# Development with automatic migrations
./deployments/scripts/deploy.sh dev --migrate

# Testing with rebuild and logs
./deployments/scripts/deploy.sh test --rebuild --logs

# Production with health checks
./deployments/scripts/deploy.sh prod --health-check
```

### Manual Docker Compose

```bash
# Rebuild and start services
docker-compose -f deployments/[env]/docker-compose-[env].yml up --build -d

# View real-time logs
docker-compose -f deployments/[env]/docker-compose-[env].yml logs -f

# Stop all services
docker-compose -f deployments/[env]/docker-compose-[env].yml down

# Remove volumes (complete cleanup)
docker-compose -f deployments/[env]/docker-compose-[env].yml down -v
```

## ⚙️ Environment Configuration

### Environment Files Structure
Each environment has a standardized `.env` file with the following sections:

```bash
# Database Configuration
DB_HOST=...
DB_PORT=...
DB_USER=...
DB_PASSWORD=...
DB_NAME=...

# Redis Configuration  
REDIS_URL=...
REDIS_HOST=...
REDIS_PORT=...

# JWT Configuration
JWT_SECRET=...
JWT_EXPIRY_HOURS=...

# API Configuration
API_PORT=...
API_HOST=...
LOG_LEVEL=...

# AI Integration
OPENAI_API_KEY=...
AI_MODEL=...

# Monitoring & Tracing
JAEGER_ENDPOINT=...
PROMETHEUS_METRICS=...

# Environment Specific
ENVIRONMENT=...
DEBUG=...
```

### Environment Differences

| Setting | Development | Testing | Production |
|---------|-------------|---------|------------|
| **Debug Mode** | `true` | `false` | `false` |
| **Log Level** | `debug` | `info` | `warn` |
| **JWT Expiry** | 24h | 1h | 8h |
| **Database Pool** | 5 | 10 | 25 |
| **Redis TTL** | 300s | 60s | 3600s |
| **AI Features** | enabled | disabled | enabled |
| **Metrics** | basic | detailed | full |

### Security Configuration

#### Development (Permissive)
- Debug logs enabled
- CORS origins: `*`
- Simple JWT secrets
- Local file storage

#### Testing (Controlled)
- Structured logging only
- CORS origins: test domains
- Rotation JWT secrets
- Memory storage

#### Production (Secure)
- Error logs only
- CORS origins: production domains
- HSM-backed JWT secrets
- Encrypted storage

## 📦 Deployment Scripts

### Available Scripts

#### `deploy.sh`
Main deployment automation script with options:
```bash
./deployments/scripts/deploy.sh <env> [options]

Options:
  --rebuild     Force rebuild of Docker images
  --migrate     Run database migrations
  --seed        Load seed data (dev/test only)
  --logs        Show logs after deployment
  --health      Run health checks after deployment
  --backup      Create backup before deployment (prod only)
```

#### `health-check.sh`
Comprehensive health monitoring:
```bash
./deployments/scripts/health-check.sh <env> [options]

Options:
  --detailed    Show detailed service information
  --json        Output results in JSON format
  --silent      Only return exit codes
  --timeout=30  Custom timeout in seconds
```

#### `backup.sh`
Database backup utility:
```bash
./deployments/scripts/backup.sh <env> [options]

Options:
  --compress    Compress backup files
  --encrypt     Encrypt backup files
  --remote      Upload to remote storage
  --retention=7 Keep backups for N days
```

#### `migrate.sh`
Database migration management:
```bash
./deployments/scripts/migrate.sh <env> [command]

Commands:
  up           Apply all pending migrations
  down N       Rollback N migrations
  force V      Mark migration V as applied
  version      Show current migration version
  status       Show migration status
```

## 🏭 Production Deployment

### Pre-deployment Checklist
- [ ] Environment variables configured
- [ ] SSL certificates installed
- [ ] Database backups created
- [ ] Health check endpoints working
- [ ] Monitoring systems configured
- [ ] Log aggregation setup
- [ ] Security scanning completed
- [ ] Load testing performed

### Production Best Practices

#### Infrastructure
```bash
# Use production-grade database
# - Managed PostgreSQL (AWS RDS, Google Cloud SQL)
# - Read replicas for scaling
# - Automated backups

# Use managed Redis
# - AWS ElastiCache, Google Memorystore
# - Cluster mode for high availability
# - Encryption at rest and in transit

# Container orchestration
# - Kubernetes for production
# - Auto-scaling policies
# - Rolling updates
```

#### Security
```bash
# Secrets management
# - AWS Secrets Manager, HashiCorp Vault
# - Never commit secrets to code
# - Rotate secrets regularly

# Network security
# - VPC/private networks
# - Security groups/firewalls
# - WAF for API protection

# Container security
# - Non-root users
# - Minimal base images
# - Security scanning
```

#### Monitoring
```bash
# Application monitoring
# - Prometheus + Grafana
# - Custom business metrics
# - SLA/SLO tracking

# Infrastructure monitoring
# - CPU, memory, disk usage
# - Network performance
# - Database performance

# Alerting
# - PagerDuty, Slack integration
# - Escalation policies
# - Runbook documentation
```

### Kubernetes Deployment
```bash
# Apply Kubernetes manifests
kubectl apply -f deployments/k8s/

# Create namespace
kubectl create namespace news-api-prod

# Deploy with Helm (if available)
helm install news-api ./charts/news-api --namespace news-api-prod

# Check deployment status
kubectl get pods -n news-api-prod
kubectl get services -n news-api-prod

# View logs
kubectl logs -f deployment/news-api -n news-api-prod
```

## 🔄 CI/CD Integration

### GitHub Actions
```yaml
# Example workflow for automated deployment
name: Deploy
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Deploy to Testing
        run: ./deployments/scripts/deploy.sh test --migrate --health
      - name: Deploy to Production
        if: github.ref == 'refs/heads/main'
        run: ./deployments/scripts/deploy.sh prod --backup --migrate --health
```

### GitLab CI
```yaml
# Example .gitlab-ci.yml
stages:
  - test
  - deploy

deploy_test:
  stage: deploy
  script:
    - ./deployments/scripts/deploy.sh test --migrate
  only:
    - develop

deploy_prod:
  stage: deploy
  script:
    - ./deployments/scripts/deploy.sh prod --backup --migrate
  only:
    - main
```

## 🔍 Service Management

### Container Status
```bash
# Check running containers
docker-compose -f deployments/[env]/docker-compose-[env].yml ps

# View resource usage
docker stats

# Inspect specific service
docker-compose -f deployments/[env]/docker-compose-[env].yml exec [env]_api sh
```

### Logs and Debugging
```bash
# View all logs
docker-compose -f deployments/[env]/docker-compose-[env].yml logs

# Follow API logs only
docker-compose -f deployments/[env]/docker-compose-[env].yml logs -f [env]_api

# View database logs
docker-compose -f deployments/[env]/docker-compose-[env].yml logs [env]_postgres

# View Redis logs
docker-compose -f deployments/[env]/docker-compose-[env].yml logs [env]_redis
```

### Database Operations
```bash
# Connect to database
docker-compose -f deployments/[env]/docker-compose-[env].yml exec [env]_postgres psql -U newsapi -d newsapi_[env]

# Run migrations manually
docker-compose -f deployments/[env]/docker-compose-[env].yml exec [env]_api ./migrate -path /app/internal/database/migrations -database "postgres://newsapi:password@[env]_postgres:5432/newsapi_[env]?sslmode=disable" up

# Create database backup
docker-compose -f deployments/[env]/docker-compose-[env].yml exec [env]_postgres pg_dump -U newsapi newsapi_[env] > backup_$(date +%Y%m%d_%H%M%S).sql
```

## 🛠️ Troubleshooting

### Common Issues

#### Port Conflicts
```bash
# Check what's using a port
lsof -i :8080
lsof -i :5432

# Kill process using port
kill -9 $(lsof -t -i:8080)
```

#### Redis Connection Issues
```bash
# Check Redis connectivity
docker-compose -f deployments/test/docker-compose-test.yml exec test_redis redis-cli ping

# Verify Redis URL in environment
docker-compose -f deployments/test/docker-compose-test.yml exec test_api env | grep REDIS
```

#### Database Connection Problems
```bash
# Test database connection
docker-compose -f deployments/test/docker-compose-test.yml exec test_postgres pg_isready -U newsapi

# Check database exists
docker-compose -f deployments/test/docker-compose-test.yml exec test_postgres psql -U newsapi -l
```

#### Migration Issues
```bash
# Check migration status
docker-compose -f deployments/[env]/docker-compose-[env].yml exec [env]_postgres psql -U newsapi -d newsapi_[env] -c "SELECT * FROM schema_migrations;"

# Reset dirty migration
docker-compose -f deployments/[env]/docker-compose-[env].yml exec [env]_postgres psql -U newsapi -d newsapi_[env] -c "UPDATE schema_migrations SET dirty = false WHERE version = [version_number];"

# Clean restart (removes all data)
docker-compose -f deployments/[env]/docker-compose-[env].yml down -v
docker-compose -f deployments/[env]/docker-compose-[env].yml up -d
```

### Health Checks
```bash
# API health check
curl http://localhost:808[0|1|2]/health

# Database health
curl http://localhost:808[0|1|2]/health/db

# Redis health
curl http://localhost:808[0|1|2]/health/redis

# Complete system status
./deployments/scripts/health-check.sh [env] --detailed
```

### Performance Monitoring
```bash
# View container resource usage
docker-compose -f deployments/[env]/docker-compose-[env].yml top

# Monitor logs for errors
docker-compose -f deployments/[env]/docker-compose-[env].yml logs -f | grep -i error

# Check API response times
curl -w "@deployments/scripts/curl-format.txt" -o /dev/null -s http://localhost:8081/api/news
```
