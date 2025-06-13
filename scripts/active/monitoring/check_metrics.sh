#!/bin/bash

# This script verifies that the metrics system is working
# by checking that Prometheus is scraping metrics from the API
# and that Grafana can access Prometheus

echo "Checking metrics system..."

# Check if the Prometheus service is running
echo "Checking Prometheus service..."
if docker ps | grep -q prometheus; then
    echo "✅ Prometheus service is running"
else
    echo "❌ Prometheus service is not running"
    echo "Try running: docker-compose -f ../docker/docker-compose-dev.yml up -d prometheus"
fi

# Check if the Grafana service is running
echo "Checking Grafana service..."
if docker ps | grep -q grafana; then
    echo "✅ Grafana service is running"
else
    echo "❌ Grafana service is not running"
    echo "Try running: docker-compose -f docker-compose-dev.yml up -d grafana"
fi

# Check if the API is running and exposing metrics
echo "Checking API metrics endpoint..."
if curl -s http://localhost:8080/metrics > /dev/null; then
    echo "✅ API metrics endpoint is accessible"
else
    echo "❌ API metrics endpoint is not accessible"
    echo "Make sure the API is running and exposing metrics at /metrics"
fi

# Check if Prometheus can scrape metrics from the API
echo "Checking if Prometheus is scraping API metrics..."
if curl -s http://localhost:9090/api/v1/targets | grep -q "news_api"; then
    echo "✅ Prometheus is scraping API metrics"
else
    echo "❌ Prometheus is not scraping API metrics"
    echo "Check the prometheus.yml configuration and make sure the API target is correct"
fi

# Check if Grafana can access Prometheus data source
echo "Checking Grafana data source..."
if curl -s -u admin:admin http://localhost:3000/api/datasources/name/prometheus | grep -q "prometheus"; then
    echo "✅ Grafana can access Prometheus data source"
else
    echo "❌ Grafana cannot access Prometheus data source"
    echo "Check the Grafana data source configuration"
fi

echo "Metrics system check complete"
