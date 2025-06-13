#!/bin/bash

# Deployment script for News API
# Usage: ./deploy.sh [environment] [options]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT=""
BUILD_ONLY=false
NO_CACHE=false
ROLLBACK=false
BACKUP=true

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
    echo "Usage: $0 [environment] [options]"
    echo ""
    echo "Environments:"
    echo "  dev      - Development environment"
    echo "  test     - Testing environment"
    echo "  prod     - Production environment"
    echo ""
    echo "Options:"
    echo "  --build-only    Build images only, don't deploy"
    echo "  --no-cache      Build without using cache"
    echo "  --no-backup     Skip database backup (prod only)"
    echo "  --rollback      Rollback to previous version"
    echo "  --help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 dev                    # Deploy to development"
    echo "  $0 prod --no-cache        # Deploy to production with fresh build"
    echo "  $0 test --build-only      # Build test images only"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        dev|test|prod)
            ENVIRONMENT="$1"
            shift
            ;;
        --build-only)
            BUILD_ONLY=true
            shift
            ;;
        --no-cache)
            NO_CACHE=true
            shift
            ;;
        --no-backup)
            BACKUP=false
            shift
            ;;
        --rollback)
            ROLLBACK=true
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

# Validate environment
if [[ -z "$ENVIRONMENT" ]]; then
    print_error "Environment is required"
    show_usage
    exit 1
fi

# Set environment-specific variables
case $ENVIRONMENT in
    dev)
        COMPOSE_FILE="deployments/dev/docker-compose-dev.yml"
        ENV_FILE="deployments/dev/.env.dev"
        PROJECT_NAME="news_dev"
        ;;
    test)
        COMPOSE_FILE="deployments/test/docker-compose-test.yml"
        ENV_FILE="deployments/test/.env.test"
        PROJECT_NAME="news_test"
        ;;
    prod)
        COMPOSE_FILE="deployments/prod/docker-compose-prod.yml"
        ENV_FILE="deployments/prod/.env.prod"
        PROJECT_NAME="news_prod"
        
        # Check if production env file exists
        if [[ ! -f "$ENV_FILE" ]]; then
            print_error "Production environment file not found: $ENV_FILE"
            print_warning "Copy .env.prod.example to .env.prod and configure it first"
            exit 1
        fi
        ;;
esac

print_status "Deploying to $ENVIRONMENT environment..."

# Change to project root
cd "$(dirname "$0")/.."

# Check if docker-compose file exists
if [[ ! -f "$COMPOSE_FILE" ]]; then
    print_error "Docker compose file not found: $COMPOSE_FILE"
    exit 1
fi

# Function to perform backup
perform_backup() {
    if [[ "$ENVIRONMENT" == "prod" && "$BACKUP" == true ]]; then
        print_status "Creating database backup..."
        ./scripts/backup.sh
        if [[ $? -eq 0 ]]; then
            print_success "Database backup completed"
        else
            print_warning "Database backup failed, continuing anyway..."
        fi
    fi
}

# Function to build images
build_images() {
    print_status "Building Docker images..."
    
    if [[ "$NO_CACHE" == true ]]; then
        docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" build --no-cache
    else
        docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" build
    fi
    
    if [[ $? -eq 0 ]]; then
        print_success "Images built successfully"
    else
        print_error "Image build failed"
        exit 1
    fi
}

# Function to deploy services
deploy_services() {
    print_status "Deploying services..."
    
    # Load environment variables
    if [[ -f "$ENV_FILE" ]]; then
        export $(cat "$ENV_FILE" | grep -v '#' | xargs)
    fi
    
    # Stop existing services
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" down
    
    # Start services
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" up -d
    
    if [[ $? -eq 0 ]]; then
        print_success "Services deployed successfully"
    else
        print_error "Service deployment failed"
        exit 1
    fi
}

# Function to wait for services to be healthy
wait_for_health() {
    print_status "Waiting for services to be healthy..."
    
    # Wait for database
    max_attempts=30
    attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        if docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" exec -T "${ENVIRONMENT}_db" pg_isready > /dev/null 2>&1; then
            break
        fi
        
        print_status "Waiting for database... (attempt $attempt/$max_attempts)"
        sleep 5
        ((attempt++))
    done
    
    if [[ $attempt -gt $max_attempts ]]; then
        print_error "Database failed to become healthy"
        exit 1
    fi
    
    # Wait for API
    attempt=1
    while [[ $attempt -le $max_attempts ]]; do
        if curl -f "http://localhost:$(docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" port "${ENVIRONMENT}_api" 8080 | cut -d: -f2)/health" > /dev/null 2>&1; then
            break
        fi
        
        print_status "Waiting for API... (attempt $attempt/$max_attempts)"
        sleep 5
        ((attempt++))
    done
    
    if [[ $attempt -gt $max_attempts ]]; then
        print_error "API failed to become healthy"
        exit 1
    fi
    
    print_success "All services are healthy"
}

# Function to run database migrations
run_migrations() {
    print_status "Running database migrations..."
    
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" exec -T "${ENVIRONMENT}_api" ./news-api migrate up
    
    if [[ $? -eq 0 ]]; then
        print_success "Database migrations completed"
    else
        print_error "Database migrations failed"
        exit 1
    fi
}

# Function to show deployment info
show_info() {
    print_success "Deployment completed successfully!"
    echo ""
    print_status "Service URLs:"
    
    API_PORT=$(docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" port "${ENVIRONMENT}_api" 8080 2>/dev/null | cut -d: -f2)
    if [[ -n "$API_PORT" ]]; then
        echo "  API: http://localhost:$API_PORT"
        echo "  Health: http://localhost:$API_PORT/health"
        echo "  Swagger: http://localhost:$API_PORT/swagger/index.html"
    fi
    
    JAEGER_PORT=$(docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" port "${ENVIRONMENT}_jaeger" 16686 2>/dev/null | cut -d: -f2)
    if [[ -n "$JAEGER_PORT" ]]; then
        echo "  Jaeger UI: http://localhost:$JAEGER_PORT"
    fi
    
    echo ""
    print_status "Useful commands:"
    echo "  View logs: docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f"
    echo "  Stop services: docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down"
    echo "  Service status: docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME ps"
}

# Main deployment flow
main() {
    if [[ "$ROLLBACK" == true ]]; then
        print_status "Rolling back deployment..."
        # Implement rollback logic here
        print_warning "Rollback functionality not implemented yet"
        exit 1
    fi
    
    # Perform backup before deployment
    perform_backup
    
    # Build images
    build_images
    
    # Deploy if not build-only
    if [[ "$BUILD_ONLY" == false ]]; then
        deploy_services
        wait_for_health
        run_migrations
        show_info
    else
        print_success "Build completed successfully"
    fi
}

# Run main function
main
