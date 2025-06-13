# Monitoring Configuration

This directory contains configuration files for the monitoring stack used in the News API project.

## Contents

- `prometheus.yml` - Prometheus configuration for metrics collection
- `prometheus-otel.yml` - Prometheus configuration for OpenTelemetry integration
- `otel-collector-config.yaml` - OpenTelemetry collector configuration
- `grafana/` - Grafana dashboard and configuration files

## Usage

These configuration files are used by the Docker Compose setup. You can start the monitoring stack using:

```bash
# Start monitoring services
make metrics-up

# Check metrics setup
make check-metrics

# Start OpenTelemetry stack
./scripts/docker-helper.sh up-otel
```

## Dashboards

Grafana dashboards are available at `http://localhost:3000` after starting the monitoring stack.

Default credentials:
- Username: admin
- Password: admin

## Prometheus

Prometheus is available at `http://localhost:9090` after starting the monitoring stack.

## References

For more information about monitoring and observability in this project, see:
- [OBSERVABILITY_GUIDE.md](/Users/madraka/News/OBSERVABILITY_GUIDE.md)
