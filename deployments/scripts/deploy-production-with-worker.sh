#!/bin/bash

# Production Deployment Script with Worker Integration
# Usage: ./deploy-production-with-worker.sh [options]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="deployments/prod/docker-compose-prod.yml"
PROJECT_NAME="news_prod"
ENV_FILE="deployments/prod/.env.prod"

# Options
BUILD_WORKER_ONLY=false
CHECK_WORKER_HEALTH=true
SCALE_WORKERS=3

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --worker-only     Deploy only the worker service"
    echo "  --scale N         Scale workers to N instances (default: 3)"
    echo "  --no-health       Skip worker health checks"
    echo "  --help           Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                           # Full production deployment with worker"
    echo "  $0 --worker-only             # Deploy only the worker service"
    echo "  $0 --scale 5                 # Deploy with 5 worker instances"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --worker-only)
            BUILD_WORKER_ONLY=true
            shift
            ;;
        --scale)
            SCALE_WORKERS="$2"
            shift 2
            ;;
        --no-health)
            CHECK_WORKER_HEALTH=false
            shift
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Validation
if [[ ! -f "$COMPOSE_FILE" ]]; then
    print_error "Docker compose file not found: $COMPOSE_FILE"
    exit 1
fi

if [[ ! -f "$ENV_FILE" ]]; then
    print_error "Environment file not found: $ENV_FILE"
    exit 1
fi

# Function to build worker image
build_worker() {
    print_status "Building worker Docker image..."
    
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" build prod_worker
    
    if [[ $? -eq 0 ]]; then
        print_success "Worker image built successfully"
    else
        print_error "Worker image build failed"
        exit 1
    fi
}

# Function to deploy full production stack
deploy_full_stack() {
    print_status "Deploying full production stack with worker..."
    
    # Load environment variables
    if [[ -f "$ENV_FILE" ]]; then
        export $(cat "$ENV_FILE" | grep -v '#' | xargs)
    fi
    
    # Build all images
    print_status "Building all production images..."
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" build
    
    # Stop existing services
    print_status "Stopping existing services..."
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" down
    
    # Start infrastructure services first
    print_status "Starting infrastructure services..."
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" up -d prod_db prod_redis prod_elasticsearch
    
    # Wait for infrastructure
    print_status "Waiting for infrastructure to be ready..."
    sleep 30
    
    # Start application services
    print_status "Starting application services..."
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" up -d prod_api
    
    # Start worker service
    print_status "Starting worker service..."
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" up -d prod_worker
    
    # Scale workers if specified
    if [[ "$SCALE_WORKERS" -gt 1 ]]; then
        print_status "Scaling workers to $SCALE_WORKERS instances..."
        docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" up -d --scale prod_worker="$SCALE_WORKERS" prod_worker
    fi
    
    # Start monitoring services
    print_status "Starting monitoring services..."
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" up -d prod_prometheus prod_grafana prod_jaeger
    
    if [[ $? -eq 0 ]]; then
        print_success "Full production stack deployed successfully"
    else
        print_error "Deployment failed"
        exit 1
    fi
}

# Function to deploy worker only
deploy_worker_only() {
    print_status "Deploying worker service only..."
    
    # Load environment variables
    if [[ -f "$ENV_FILE" ]]; then
        export $(cat "$ENV_FILE" | grep -v '#' | xargs)
    fi
    
    # Build worker image
    build_worker
    
    # Stop existing worker
    print_status "Stopping existing worker..."
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" stop prod_worker
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" rm -f prod_worker
    
    # Start worker service
    print_status "Starting worker service..."
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" up -d prod_worker
    
    # Scale workers if specified
    if [[ "$SCALE_WORKERS" -gt 1 ]]; then
        print_status "Scaling workers to $SCALE_WORKERS instances..."
        docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" up -d --scale prod_worker="$SCALE_WORKERS" prod_worker
    fi
    
    if [[ $? -eq 0 ]]; then
        print_success "Worker service deployed successfully"
    else
        print_error "Worker deployment failed"
        exit 1
    fi
}

# Function to check worker health
check_worker_health() {
    if [[ "$CHECK_WORKER_HEALTH" == false ]]; then
        return 0
    fi
    
    print_status "Checking worker health..."
    
    # Wait for workers to start
    sleep 15
    
    # Check if worker containers are running
    WORKER_COUNT=$(docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps -q prod_worker | wc -l)
    if [[ "$WORKER_COUNT" -eq 0 ]]; then
        print_error "No worker containers found"
        return 1
    fi
    
    print_success "Found $WORKER_COUNT worker container(s) running"
    
    # Check worker process health
    HEALTHY_WORKERS=0
    for container in $(docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps -q prod_worker); do
        if docker exec "$container" pgrep -f 'worker' > /dev/null 2>&1; then
            ((HEALTHY_WORKERS++))
        fi
    done
    
    if [[ "$HEALTHY_WORKERS" -eq "$WORKER_COUNT" ]]; then
        print_success "All $HEALTHY_WORKERS worker(s) are healthy"
        return 0
    else
        print_warning "$HEALTHY_WORKERS out of $WORKER_COUNT workers are healthy"
        return 1
    fi
}

# Function to show deployment status
show_deployment_status() {
    print_status "Production deployment status:"
    echo ""
    
    # Show service status
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps
    echo ""
    
    # Show worker-specific information
    print_status "Worker service details:"
    WORKER_CONTAINERS=$(docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps prod_worker)
    echo "$WORKER_CONTAINERS"
    echo ""
    
    # Show recent worker logs
    print_status "Recent worker logs (last 10 lines):"
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" logs --tail=10 prod_worker
    echo ""
    
    # Show useful commands
    print_status "Useful commands:"
    echo "  View worker logs: docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f prod_worker"
    echo "  Scale workers: docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d --scale prod_worker=5 prod_worker"
    echo "  Stop workers: docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME stop prod_worker"
    echo "  Worker stats: docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec prod_worker ps aux"
    echo "  Queue status: docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec prod_redis redis-cli LLEN queue:general"
    echo ""
    
    # Show access URLs
    API_PORT=$(docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" port prod_api 8080 2>/dev/null | cut -d: -f2)
    GRAFANA_PORT=$(docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" port prod_grafana 3000 2>/dev/null | cut -d: -f2)
    
    if [[ -n "$API_PORT" ]]; then
        echo "  API: http://localhost:$API_PORT"
        echo "  Health: http://localhost:$API_PORT/health"
    fi
    
    if [[ -n "$GRAFANA_PORT" ]]; then
        echo "  Grafana: http://localhost:$GRAFANA_PORT"
    fi
}

# Main execution
print_status "Starting production deployment with worker integration..."

if [[ "$BUILD_WORKER_ONLY" == true ]]; then
    deploy_worker_only
else
    deploy_full_stack
fi

# Check worker health
if ! check_worker_health; then
    print_warning "Some workers may not be healthy. Check logs for details."
fi

# Show deployment status
show_deployment_status

print_success "Production deployment completed successfully!"
print_status "Worker system is ready for background job processing."
