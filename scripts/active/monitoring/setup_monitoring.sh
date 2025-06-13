#!/bin/bash

# Create necessary directories for monitoring
echo "Creating monitoring directories..."
mkdir -p monitoring/grafana/dashboards
mkdir -p monitoring/grafana/provisioning/datasources
mkdir -p monitoring/grafana/provisioning/dashboards

# Make sure the script is executable
chmod +x scripts/load-test.js

# Create directories to store uploads
mkdir -p uploads

echo "Setting up monitoring environment..."

# Start the monitoring stack
echo "Starting services with monitoring enabled..."
docker-compose -f ../docker/docker-compose-dev.yml up -d

# Wait for services to be ready
echo "Waiting for services to start up..."
sleep 10

echo "Setup complete! Services are running with monitoring enabled."
echo ""
echo "Available endpoints:"
echo "- News API: http://localhost:8080"
echo "- Prometheus: http://localhost:9090"
echo "- Grafana: http://localhost:3000 (admin/admin)"
echo ""
echo "To run load tests:"
echo "k6 run scripts/load-test.js"
