#!/bin/bash

# Database migration script for Docker environments
# Usage: ./migrate.sh [environment] [command] [options]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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
    echo "Usage: $0 [environment] [command] [options]"
    echo ""
    echo "Environments:"
    echo "  dev      - Development environment"
    echo "  test     - Testing environment"  
    echo "  prod     - Production environment"
    echo ""
    echo "Commands:"
    echo "  up           Apply all pending migrations"
    echo "  down N       Rollback N migrations"
    echo "  force V      Mark migration V as applied (use with caution)"
    echo "  version      Show current migration version"
    echo "  status       Show migration status"
    echo "  reset        Reset database and apply all migrations"
    echo ""
    echo "Options:"
    echo "  --dry-run    Show what would be done without executing"
    echo "  --verbose    Show detailed output"
    echo "  --help       Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 dev up                 # Apply all pending migrations to dev"
    echo "  $0 test down 2            # Rollback 2 migrations in test"
    echo "  $0 prod version           # Show current migration version in prod"
    echo "  $0 dev reset --verbose    # Reset dev database with detailed output"
}

# Default values
ENVIRONMENT=""
COMMAND=""
MIGRATION_COUNT=""
DRY_RUN=false
VERBOSE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        dev|test|prod)
            ENVIRONMENT="$1"
            shift
            ;;
        up|down|force|version|status|reset)
            COMMAND="$1"
            shift
            # For down and force commands, get the next argument as count/version
            if [[ "$COMMAND" == "down" || "$COMMAND" == "force" ]] && [[ $# -gt 0 && "$1" =~ ^[0-9]+$ ]]; then
                MIGRATION_COUNT="$1"
                shift
            fi
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            if [[ "$1" =~ ^[0-9]+$ ]] && [[ -z "$MIGRATION_COUNT" ]]; then
                MIGRATION_COUNT="$1"
            else
                print_error "Unknown option: $1"
                show_usage
                exit 1
            fi
            shift
            ;;
    esac
done

# Validate environment and command
if [[ -z "$ENVIRONMENT" ]]; then
    print_error "Environment is required"
    show_usage
    exit 1
fi

if [[ -z "$COMMAND" ]]; then
    print_error "Command is required"
    show_usage
    exit 1
fi

# Validate migration count for down and force commands
if [[ "$COMMAND" == "down" || "$COMMAND" == "force" ]] && [[ -z "$MIGRATION_COUNT" ]]; then
    print_error "Migration count/version is required for '$COMMAND' command"
    show_usage
    exit 1
fi

# Set up environment-specific variables
case $ENVIRONMENT in
    dev)
        COMPOSE_FILE="docker-compose-dev.yml"
        DB_CONTAINER="news_${ENVIRONMENT}_db"
        API_CONTAINER="news_${ENVIRONMENT}_api"
        DB_NAME="newsapi_dev"
        ;;
    test)
        COMPOSE_FILE="docker-compose-test.yml"
        DB_CONTAINER="news_${ENVIRONMENT}_db"
        API_CONTAINER="news_${ENVIRONMENT}_api"
        DB_NAME="newsapi_test"
        ;;
    prod)
        COMPOSE_FILE="docker-compose-prod.yml"
        DB_CONTAINER="news_${ENVIRONMENT}_db"
        API_CONTAINER="news_${ENVIRONMENT}_api"
        DB_NAME="newsapi_production"
        ;;
    *)
        print_error "Unknown environment: $ENVIRONMENT"
        exit 1
        ;;
esac

# Set script directory and navigate to deployment environment
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_DIR="$SCRIPT_DIR/../$ENVIRONMENT"

if [[ ! -d "$ENV_DIR" ]]; then
    print_error "Environment directory not found: $ENV_DIR"
    exit 1
fi

cd "$ENV_DIR"

# Check if docker-compose file exists
if [[ ! -f "$COMPOSE_FILE" ]]; then
    print_error "Docker compose file not found: $COMPOSE_FILE"
    exit 1
fi

# Function to check if containers are running
check_containers() {
    if ! docker-compose -f "$COMPOSE_FILE" ps | grep -q "$DB_CONTAINER.*Up"; then
        print_error "Database container is not running. Please start the environment first:"
        echo "  cd $ENV_DIR && docker-compose -f $COMPOSE_FILE up -d"
        exit 1
    fi
}

# Function to execute migration command
execute_migration() {
    local migration_cmd="$1"
    local description="$2"
    
    print_status "$description"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_warning "DRY RUN: Would execute: $migration_cmd"
        return
    fi
    
    if [[ "$VERBOSE" == "true" ]]; then
        print_status "Executing: $migration_cmd"
    fi
    
    if docker-compose -f "$COMPOSE_FILE" exec -T "$API_CONTAINER" sh -c "$migration_cmd"; then
        print_success "$description completed successfully"
    else
        print_error "$description failed"
        exit 1
    fi
}

# Function to get database connection string
get_db_connection() {
    echo "postgres://newsapi:password@${DB_CONTAINER}:5432/${DB_NAME}?sslmode=disable"
}

# Main execution
print_status "Starting migration operation: $COMMAND on $ENVIRONMENT environment"

# Check if containers are running
check_containers

DB_URL=$(get_db_connection)
MIGRATE_PATH="/app/internal/database/migrations"

case $COMMAND in
    up)
        execute_migration "./migrate -path $MIGRATE_PATH -database \"$DB_URL\" up" "Applying all pending migrations"
        ;;
    down)
        if [[ -z "$MIGRATION_COUNT" ]]; then
            print_error "Number of migrations to rollback is required"
            exit 1
        fi
        execute_migration "./migrate -path $MIGRATE_PATH -database \"$DB_URL\" down $MIGRATION_COUNT" "Rolling back $MIGRATION_COUNT migration(s)"
        ;;
    force)
        if [[ -z "$MIGRATION_COUNT" ]]; then
            print_error "Migration version is required"
            exit 1
        fi
        print_warning "Forcing migration version $MIGRATION_COUNT. This should only be used to fix dirty migration state."
        execute_migration "./migrate -path $MIGRATE_PATH -database \"$DB_URL\" force $MIGRATION_COUNT" "Forcing migration version $MIGRATION_COUNT"
        ;;
    version)
        print_status "Getting current migration version"
        if [[ "$DRY_RUN" == "false" ]]; then
            docker-compose -f "$COMPOSE_FILE" exec -T "$API_CONTAINER" sh -c "./migrate -path $MIGRATE_PATH -database \"$DB_URL\" version"
        else
            print_warning "DRY RUN: Would check migration version"
        fi
        ;;
    status)
        print_status "Checking migration status"
        if [[ "$DRY_RUN" == "false" ]]; then
            # Check current version and available migrations
            echo "Current migration version:"
            docker-compose -f "$COMPOSE_FILE" exec -T "$API_CONTAINER" sh -c "./migrate -path $MIGRATE_PATH -database \"$DB_URL\" version"
            echo ""
            echo "Available migration files:"
            docker-compose -f "$COMPOSE_FILE" exec -T "$API_CONTAINER" sh -c "ls -la $MIGRATE_PATH/*.sql"
            echo ""
            echo "Migration status in database:"
            docker-compose -f "$COMPOSE_FILE" exec -T "$DB_CONTAINER" psql -U newsapi -d "$DB_NAME" -c "SELECT version, dirty FROM schema_migrations ORDER BY version;"
        else
            print_warning "DRY RUN: Would check migration status"
        fi
        ;;
    reset)
        print_warning "This will completely reset the database and apply all migrations!"
        if [[ "$DRY_RUN" == "false" ]]; then
            read -p "Are you sure you want to continue? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                print_status "Operation cancelled"
                exit 0
            fi
        fi
        
        # Drop and recreate database
        execute_migration "psql -h $DB_CONTAINER -U newsapi -d postgres -c \"DROP DATABASE IF EXISTS $DB_NAME;\"" "Dropping database"
        execute_migration "psql -h $DB_CONTAINER -U newsapi -d postgres -c \"CREATE DATABASE $DB_NAME;\"" "Creating database"
        execute_migration "./migrate -path $MIGRATE_PATH -database \"$DB_URL\" up" "Applying all migrations"
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac

print_success "Migration operation completed successfully!"
