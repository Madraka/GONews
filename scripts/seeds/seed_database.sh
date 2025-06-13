#!/bin/bash

# Database seeding script for Docker environments
# Usage: ./seed_database.sh [environment]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[SEED]${NC} $1"
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

# Get environment from ENV variable or argument
ENVIRONMENT=${ENVIRONMENT:-${1:-dev}}

print_status "Starting database seeding for $ENVIRONMENT environment..."

# Set environment-specific variables
case $ENVIRONMENT in
    dev)
        COMPOSE_FILE="deployments/dev/docker-compose-dev.yml"
        API_CONTAINER="news_dev_api"
        DB_CONTAINER="news_dev_db"
        DB_NAME="newsapi_dev"
        DB_USER="devuser"
        ;;
    test)
        COMPOSE_FILE="deployments/test/docker-compose-test.yml"
        API_CONTAINER="news_test_api"
        DB_CONTAINER="news_test_db"
        DB_NAME="newsapi_test"
        DB_USER="testuser"
        ;;
    prod)
        COMPOSE_FILE="deployments/prod/docker-compose-prod.yml"
        API_CONTAINER="news_prod_api"
        DB_CONTAINER="news_prod_db"
        DB_NAME="newsdb_prod"
        DB_USER="produser"
        ;;
    *)
        print_error "Unknown environment: $ENVIRONMENT"
        exit 1
        ;;
esac

# Check if containers are running
if ! docker ps | grep -q "$API_CONTAINER"; then
    print_error "API container ($API_CONTAINER) is not running"
    print_warning "Please start the environment first with: make $ENVIRONMENT"
    exit 1
fi

if ! docker ps | grep -q "$DB_CONTAINER"; then
    print_error "Database container ($DB_CONTAINER) is not running"
    print_warning "Please start the environment first with: make $ENVIRONMENT"
    exit 1
fi

# Wait for database to be ready
print_status "Checking database connectivity..."
for i in {1..30}; do
    if docker exec "$DB_CONTAINER" pg_isready -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1; then
        print_success "Database is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        print_error "Database failed to become ready"
        exit 1
    fi
    print_status "Waiting for database... (attempt $i/30)"
    sleep 2
done

# Run GORM AutoMigrate and Seeding
print_status "Running GORM AutoMigrate and Database Seeding..."

# Run seeding through Docker exec using Go
print_status "Building and running seeding through Go..."

docker exec "$API_CONTAINER" sh -c "
    cd /app && \
    go run -mod=mod cmd/seed/main.go --env=$ENVIRONMENT
"

if [ $? -eq 0 ]; then
    print_success "Database seeding completed successfully!"
else
    print_error "Database seeding failed"
    exit 1
fi

# Show seeding results
print_status "Checking seeded data..."
TABLE_COUNT=$(docker exec "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public';" | tr -d ' ')
print_success "Total tables created: $TABLE_COUNT"

# Show specific counts for key tables
print_status "Data summary:"
for table in users articles categories tags pages page_templates; do
    if docker exec "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c "\d $table" >/dev/null 2>&1; then
        COUNT=$(docker exec "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT count(*) FROM $table;" 2>/dev/null | tr -d ' ' || echo "0")
        printf "  %-15s: %s records\n" "$table" "$COUNT"
    fi
done

print_success "🌱 Database seeding for $ENVIRONMENT environment completed!"
