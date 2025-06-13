# OpenTelemetry Distributed Tracing Setup

## Overview

The News API now includes comprehensive OpenTelemetry distributed tracing, providing visibility into request flows, dependencies, and performance characteristics across the entire application stack.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   News API      │───▶│ OpenTelemetry   │───▶│    Jaeger       │
│  (Instrumented) │    │   Collector     │    │  (Trace Store)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │   Prometheus    │    │    Grafana      │
│   (Database)    │    │   (Metrics)     │    │ (Visualization) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Features Implemented

### 1. **Application Instrumentation**
- **HTTP Requests**: Automatic tracing of all incoming HTTP requests
- **Database Operations**: Tracing of GORM database queries with OTLP plugin
- **Cache Operations**: Tracing of Redis cache interactions
- **Service Layer**: Custom spans for business logic operations
- **Error Tracking**: Automatic error recording and span status updates

### 2. **Trace Attributes**
- Request/response metadata (method, status, size)
- Database query details (operation type, table, duration)
- Cache hit/miss ratios and operation types
- User context (when available)
- Custom business logic attributes

### 3. **OpenTelemetry Collector**
- **Receivers**: OTLP (gRPC/HTTP), Prometheus scraping
- **Processors**: Batch processing, memory limiting, resource enhancement
- **Exporters**: Jaeger (traces), Prometheus (metrics), logging (debug)

### 4. **Observability Stack**
- **Jaeger**: Distributed trace visualization and analysis
- **Prometheus**: Metrics collection and alerting
- **Grafana**: Unified dashboards for metrics and traces

## Quick Start

### 1. Start the Full Stack
```bash
./scripts/start_otel_environment.sh
```

### 2. Generate Some Traces
```bash
# Basic API requests
curl -H "X-API-Key: basic_key_123" http://localhost:8080/api/news
curl -H "X-API-Key: basic_key_123" http://localhost:8080/api/news/1

# Create some load
for i in {1..10}; do
  curl -H "X-API-Key: basic_key_123" http://localhost:8080/api/news?page=$i
done
```

### 3. View Traces
- **Jaeger UI**: http://localhost:16686
- **Grafana**: http://localhost:3000 (admin/admin)

## Configuration

### Environment Variables
```bash
# OpenTelemetry Configuration
OTEL_COLLECTOR_URL=http://otel-collector:4318    # Collector endpoint
OTEL_SERVICE_NAME=news-api                       # Service name in traces
OTEL_SERVICE_VERSION=1.0.0                       # Service version
OTEL_RESOURCE_ATTRIBUTES=environment=production  # Additional attributes
```

### Trace Sampling
The application uses `AlwaysSample()` for development. For production, consider:
- `TraceIDRatioBased(0.1)` for 10% sampling
- `ParentBased(TraceIDRatioBased(0.1))` for parent-based sampling

## Monitoring and Alerting

### Key Metrics to Monitor
1. **Trace Volume**: `otel_collector_receiver_accepted_spans_total`
2. **Error Rate**: `otel_collector_processor_batch_send_failed_spans_total`
3. **Latency**: HTTP request duration from traces
4. **Service Dependencies**: Span relationships and call patterns

### Sample Queries
```promql
# Average request latency by endpoint
histogram_quantile(0.95, 
  rate(http_request_duration_seconds_bucket[5m])
)

# Database operation rates
rate(database_operations_total[5m])

# Cache hit ratio
rate(cache_operations_total{result="hit"}[5m]) / 
rate(cache_operations_total[5m])
```

## Troubleshooting

### Common Issues

1. **No Traces Appearing**
   - Check OpenTelemetry Collector logs: `docker-compose logs otel-collector`
   - Verify collector endpoint: `curl http://localhost:8888/metrics`
   - Check application startup logs for tracing initialization

2. **High Memory Usage**
   - Adjust batch processor settings in `otel-collector-config.yaml`
   - Implement sampling for high-volume services
   - Monitor memory limiter processor

3. **Missing Spans**
   - Ensure context propagation in service calls
   - Check for proper span lifecycle management (defer span.End())
   - Verify attribute limits in collector configuration

### Debug Mode
Enable debug logging in the collector:
```yaml
service:
  telemetry:
    logs:
      level: "debug"
```

## Production Considerations

### Security
- Enable TLS for OTLP endpoints
- Implement authentication for collector access
- Sanitize sensitive data in span attributes

### Performance
- Use appropriate sampling strategies
- Configure batch processors for efficiency
- Monitor collector resource usage

### Retention
- Configure Jaeger storage retention policies
- Implement log rotation for collector logs
- Consider long-term trace storage solutions

## Integration with CI/CD

The tracing setup integrates with the existing CI/CD pipeline:
- Kubernetes manifests include OpenTelemetry sidecar
- Docker compose configurations for different environments
- Automated testing includes trace validation

## Contributing

When adding new traced operations:
1. Start spans with descriptive names
2. Add relevant attributes for debugging
3. Properly handle errors and set span status
4. End spans in defer statements
5. Update this documentation

## Resources

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [OTLP Specification](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/otlp.md)
