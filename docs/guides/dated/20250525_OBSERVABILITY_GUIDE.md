# News Aggregation Service - Observability Guide

## Overview

This news aggregation service includes a comprehensive observability stack with distributed tracing, metrics collection, and monitoring dashboards. The system is built with OpenTelemetry for standardized observability.

## Architecture

### Services
- **Go REST API**: Backend service with JWT authentication, CRUD operations, file uploads
- **PostgreSQL**: Primary database with user and news management
- **Redis**: Caching layer for performance optimization
- **React Frontend**: Modern UI with responsive design

### Observability Stack
- **OpenTelemetry**: Distributed tracing and metrics collection
- **Jaeger**: Distributed tracing UI and storage
- **Prometheus**: Metrics collection and storage
- **Grafana**: Visualization and dashboards

## Service URLs

| Service | URL | Description |
|---------|-----|-------------|
| API | http://localhost:8080 | REST API endpoints |
| Frontend | http://localhost:3000 | React application |
| Jaeger UI | http://localhost:16686 | Distributed tracing |
| Prometheus | http://localhost:9090 | Metrics and monitoring |
| Grafana | http://localhost:3000 | Dashboards (admin/admin) |

## API Testing

### Authentication
The API uses two-layer authentication:
1. **API Key Authentication**: Required for all endpoints (except health/metrics)
2. **JWT Authentication**: Required for protected endpoints

### Available API Keys
```bash
# Basic Tier (1 req/sec, 10K requests/day)
X-API-Key: api_key_basic_1234

# Pro Tier (5 req/sec, 50K requests/day)
X-API-Key: api_key_pro_5678

# Enterprise Tier (20 req/sec, 200K requests/day)
X-API-Key: api_key_enterprise_9012
```

### Test Commands

#### 1. Health Check
```bash
curl -X GET http://localhost:8080/health
```

#### 2. User Registration
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -H "X-API-Key: api_key_basic_1234" \
  -d '{"username":"testuser","password":"password123","role":"user"}'
```

#### 3. User Login
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -H "X-API-Key: api_key_basic_1234" \
  -d '{"username":"testuser","password":"password123"}'
```

#### 4. Get News Articles
```bash
curl -X GET http://localhost:8080/api/news \
  -H "X-API-Key: api_key_basic_1234"
```

#### 5. Get Specific Article
```bash
curl -X GET http://localhost:8080/api/news/1 \
  -H "X-API-Key: api_key_basic_1234"
```

#### 6. Pro Tier Analytics
```bash
curl -X GET http://localhost:8080/api/analytics \
  -H "X-API-Key: api_key_pro_5678"
```

#### 7. Enterprise Export
```bash
curl -X GET http://localhost:8080/api/export \
  -H "X-API-Key: api_key_enterprise_9012"
```

## Observability Features

### 1. Distributed Tracing
- **Implementation**: OpenTelemetry with Jaeger backend
- **Coverage**: HTTP requests, database queries, Redis operations
- **Features**: 
  - Request correlation across services
  - Performance bottleneck identification
  - Error tracking and debugging
  - Detailed span attributes for troubleshooting
  - Automated error recording and context propagation
  - Tracing middleware for all HTTP requests
- **Enhancements**:
  - Helper functions for consistent span creation
  - Trace ID added to response headers (`X-Trace-ID`)
  - Enhanced context propagation
  - Support for structured attributes
  - Trace correlation with logs and metrics

### 2. Metrics Collection
Custom metrics include:
- `news_api_http_requests_total`: Request counters by method, path, status
- `news_api_http_request_duration_seconds`: Request duration histograms
- `news_api_db_queries_total`: Database query counters
- `news_api_cache_operations_total`: Redis operation counters

### 3. Logging
- **Format**: Structured JSON logging
- **Fields**: Timestamp, level, message, request_id, method, path, status, latency
- **Integration**: Correlated with traces via request IDs

## Current Status

### ✅ Working Components
1. **API Service**: All endpoints functional with proper authentication
2. **Database**: PostgreSQL with proper schema and data
3. **Caching**: Redis integration working
4. **Metrics**: Custom metrics being exported at `/metrics`
5. **Prometheus**: Scraping metrics successfully
6. **Health Checks**: All services reporting healthy
7. **API Tiers**: Rate limiting and tier-based access control working

### ⚠️ Known Issues
1. **Tracing Export**: OpenTelemetry trace export issues have been resolved
   - Metrics are working perfectly
   - Traces are now properly appearing in Jaeger UI
   - API functionality is not affected

### 🔧 Troubleshooting Steps Taken
1. Fixed OpenTelemetry environment variable configuration
2. Updated OTLP endpoint from full path to base URL
3. Verified Jaeger OTLP collector is enabled
4. Confirmed metrics pipeline is working

## Deployment

### Start All Services
```bash
cd /Users/madraka/News
docker-compose up -d
```

### Check Service Status
```bash
docker-compose ps
```

### View Logs
```bash
# All services
docker-compose logs

# Specific service
docker-compose logs api
docker-compose logs db
docker-compose logs redis
```

### Stop Services
```bash
docker-compose down
```

## Performance Testing

### Load Testing
Use the provided API keys to test different tier limits:

```bash
# Test basic tier rate limiting
for i in {1..10}; do
  curl -X GET http://localhost:8080/api/news \
    -H "X-API-Key: api_key_basic_1234"
  sleep 1
done
```

### Monitor Performance
1. **Prometheus**: Query metrics at http://localhost:9090
2. **Grafana**: View dashboards at http://localhost:3000
3. **Jaeger**: Search traces at http://localhost:16686

## Distributed Tracing

### Verifying Tracing Setup

Run the verification script to ensure tracing is properly configured:

```bash
./scripts/verify_tracing.sh
```

This script:
1. Checks that Jaeger is accessible
2. Verifies the News API is running
3. Generates test traffic to produce traces
4. Confirms traces are being received by Jaeger

### Analyzing Traces

1. Open Jaeger UI at http://localhost:16686
2. Select "news-service" from the Service dropdown
3. Click "Find Traces" to see recent request traces
4. Examine individual traces to understand:
   - Request flow through services
   - Performance bottlenecks
   - Error paths and failure points
   - Dependency relationships

### Troubleshooting Common Issues

- **No traces appearing**: Check OTEL_EXPORTER_OTLP_TRACES_ENDPOINT environment variable
- **Missing spans**: Verify tracing is initialized in the service
- **Broken trace context**: Ensure context propagation between services
- **Too many/few spans**: Adjust instrumentation level in code

## Next Steps

1. **Grafana Dashboards**: Import pre-configured dashboards for the application
2. **Alerting**: Set up Prometheus alerting rules
3. **Load Testing**: Comprehensive performance testing
4. **Security**: Enable TLS for production deployment
5. **Custom Trace Analysis**: Create custom trace analysis reports

## Security Considerations

- API keys are for development/demo purposes only
- JWT tokens include CSRF protection
- Rate limiting prevents abuse
- Health checks don't require authentication
- Metrics endpoint is protected

## Support

For issues or questions:
1. Check service logs: `docker-compose logs [service]`
2. Verify service health: `curl http://localhost:8080/health`
3. Check metrics: `curl http://localhost:8080/metrics`
