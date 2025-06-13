#!/bin/bash

# Health check script for News API services
# Usage: ./health-check.sh [environment]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="prod"
TIMEOUT=30
DETAILED=false

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [environment] [options]"
    echo ""
    echo "Environments:"
    echo "  dev      - Development environment"
    echo "  test     - Testing environment"
    echo "  prod     - Production environment (default)"
    echo ""
    echo "Options:"
    echo "  --timeout N     Connection timeout in seconds (default: 30)"
    echo "  --detailed      Show detailed health information"
    echo "  --help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 prod                    # Check production health"
    echo "  $0 dev --detailed         # Check development with details"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        dev|test|prod)
            ENVIRONMENT="$1"
            shift
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --detailed)
            DETAILED=true
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

# Set environment-specific variables
case $ENVIRONMENT in
    dev)
        COMPOSE_FILE="deployments/dev/docker-compose-dev.yml"
        PROJECT_NAME="news_dev"
        API_PORT="8081"
        JAEGER_PORT="16687"
        ;;
    test)
        COMPOSE_FILE="deployments/test/docker-compose-test.yml"
        PROJECT_NAME="news_test"
        API_PORT="8082"
        JAEGER_PORT="16688"
        ;;
    prod)
        COMPOSE_FILE="deployments/prod/docker-compose-prod.yml"
        PROJECT_NAME="news_prod"
        API_PORT="8080"
        JAEGER_PORT="16686"
        ;;
esac

print_status "Checking health of $ENVIRONMENT environment..."

# Change to project root
cd "$(dirname "$0")/.."

# Function to check if a service is running
check_container() {
    local container_name="$1"
    local service_name="$2"
    
    if docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps "$container_name" | grep -q "Up"; then
        print_success "$service_name is running"
        return 0
    else
        print_error "$service_name is not running"
        return 1
    fi
}

# Function to check HTTP endpoint
check_http_endpoint() {
    local url="$1"
    local service_name="$2"
    
    if curl -sf --connect-timeout "$TIMEOUT" "$url" > /dev/null; then
        print_success "$service_name is responding"
        return 0
    else
        print_error "$service_name is not responding at $url"
        return 1
    fi
}

# Function to check database connection
check_database() {
    local container_name="${ENVIRONMENT}_db"
    local db_user=""
    local db_name=""
    
    case $ENVIRONMENT in
        dev)
            db_user="newsuser"
            db_name="newsdb"
            ;;
        test)
            db_user="testuser"
            db_name="testdb"
            ;;
        prod)
            ENV_FILE="deployments/environments/prod/.env.prod"
            if [[ -f "$ENV_FILE" ]]; then
                db_user=$(grep "^DB_USER=" "$ENV_FILE" | cut -d= -f2)
                db_name=$(grep "^DB_NAME=" "$ENV_FILE" | cut -d= -f2)
            else
                db_user="postgres"
                db_name="postgres"
            fi
            ;;
    esac
    
    if docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" exec -T "$container_name" \
        pg_isready -U "$db_user" -d "$db_name" > /dev/null 2>&1; then
        print_success "Database is accepting connections"
        return 0
    else
        print_error "Database is not accepting connections"
        return 1
    fi
}

# Function to check Redis connection
check_redis() {
    local container_name="${ENVIRONMENT}_redis"
    
    if docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" exec -T "$container_name" \
        redis-cli ping > /dev/null 2>&1; then
        print_success "Redis is responding"
        return 0
    else
        print_error "Redis is not responding"
        return 1
    fi
}

# Function to get detailed service information
show_detailed_info() {
    echo ""
    print_status "Detailed service information:"
    echo ""
    
    # Container status
    print_status "Container Status:"
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps
    echo ""
    
    # API health endpoint
    local api_health_url="http://localhost:$API_PORT/health"
    print_status "API Health Check:"
    if curl -sf --connect-timeout 5 "$api_health_url" 2>/dev/null; then
        curl -s "$api_health_url" | jq . 2>/dev/null || curl -s "$api_health_url"
    else
        print_warning "Could not retrieve detailed health information"
    fi
    echo ""
    
    # Resource usage
    print_status "Resource Usage:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}" \
        $(docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps -q) 2>/dev/null || \
        print_warning "Could not retrieve resource usage"
    echo ""
    
    # Recent logs (last 10 lines)
    print_status "Recent API Logs:"
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" logs --tail=10 "${ENVIRONMENT}_api" 2>/dev/null || \
        print_warning "Could not retrieve recent logs"
}

# Main health check function
main() {
    local overall_status=0
    
    echo ""
    print_status "=== Container Health ==="
    
    # Check containers
    check_container "${ENVIRONMENT}_api" "API Service" || overall_status=1
    check_container "${ENVIRONMENT}_db" "Database" || overall_status=1
    check_container "${ENVIRONMENT}_redis" "Redis" || overall_status=1
    check_container "${ENVIRONMENT}_jaeger" "Jaeger" || overall_status=1
    
    echo ""
    print_status "=== Service Health ==="
    
    # Check service endpoints
    check_http_endpoint "http://localhost:$API_PORT/health" "API Health Endpoint" || overall_status=1
    check_http_endpoint "http://localhost:$JAEGER_PORT" "Jaeger UI" || overall_status=1
    
    echo ""
    print_status "=== Database & Cache ==="
    
    # Check database and Redis
    check_database || overall_status=1
    check_redis || overall_status=1
    
    # Show detailed information if requested
    if [[ "$DETAILED" == true ]]; then
        show_detailed_info
    fi
    
    echo ""
    if [[ $overall_status -eq 0 ]]; then
        print_success "All services are healthy!"
        echo ""
        print_status "Service URLs:"
        echo "  API: http://localhost:$API_PORT"
        echo "  API Health: http://localhost:$API_PORT/health"
        echo "  API Docs: http://localhost:$API_PORT/swagger/index.html"
        echo "  Jaeger UI: http://localhost:$JAEGER_PORT"
    else
        print_error "Some services are unhealthy!"
        echo ""
        print_status "Troubleshooting commands:"
        echo "  View logs: docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f"
        echo "  Restart services: docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME restart"
        echo "  Check containers: docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME ps"
        exit 1
    fi
}

# Run main function
main
