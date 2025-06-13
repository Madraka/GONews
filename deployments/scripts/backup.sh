#!/bin/bash

# Database backup script for News API
# Usage: ./backup.sh [environment] [options]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="prod"
BACKUP_DIR="./backups"
RETENTION_DAYS=30
COMPRESS=true

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
    echo "  dev      - Development environment (default: news_dev)"
    echo "  test     - Testing environment (default: news_test)"
    echo "  prod     - Production environment (default: news_prod)"
    echo ""
    echo "Options:"
    echo "  --dir DIR       Backup directory (default: ./backups)"
    echo "  --retention N   Keep backups for N days (default: 30)"
    echo "  --no-compress   Don't compress backup files"
    echo "  --help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 prod                           # Backup production database"
    echo "  $0 dev --dir /backup --retention 7  # Backup dev with custom settings"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        dev|test|prod)
            ENVIRONMENT="$1"
            shift
            ;;
        --dir)
            BACKUP_DIR="$2"
            shift 2
            ;;
        --retention)
            RETENTION_DAYS="$2"
            shift 2
            ;;
        --no-compress)
            COMPRESS=false
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
        COMPOSE_FILE="deployments/environments/dev/docker-compose.yml"
        PROJECT_NAME="news_dev"
        DB_CONTAINER="${ENVIRONMENT}_db"
        DB_USER="newsuser"
        DB_NAME="newsdb"
        ;;
    test)
        COMPOSE_FILE="deployments/environments/test/docker-compose.yml"
        PROJECT_NAME="news_test"
        DB_CONTAINER="${ENVIRONMENT}_db"
        DB_USER="testuser"
        DB_NAME="testdb"
        ;;
    prod)
        COMPOSE_FILE="deployments/environments/prod/docker-compose.yml"
        PROJECT_NAME="news_prod"
        DB_CONTAINER="${ENVIRONMENT}_db"
        # Read from environment file for production
        ENV_FILE="deployments/environments/prod/.env.prod"
        if [[ -f "$ENV_FILE" ]]; then
            DB_USER=$(grep "^DB_USER=" "$ENV_FILE" | cut -d= -f2)
            DB_NAME=$(grep "^DB_NAME=" "$ENV_FILE" | cut -d= -f2)
        else
            print_error "Production environment file not found: $ENV_FILE"
            exit 1
        fi
        ;;
esac

print_status "Creating backup for $ENVIRONMENT environment..."

# Change to project root
cd "$(dirname "$0")/.."

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Generate backup filename with timestamp
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILENAME="${ENVIRONMENT}_${DB_NAME}_${TIMESTAMP}"

if [[ "$COMPRESS" == true ]]; then
    BACKUP_FILE="${BACKUP_DIR}/${BACKUP_FILENAME}.sql.gz"
else
    BACKUP_FILE="${BACKUP_DIR}/${BACKUP_FILENAME}.sql"
fi

# Check if database container is running
if ! docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps "$DB_CONTAINER" | grep -q "Up"; then
    print_error "Database container is not running: $DB_CONTAINER"
    print_status "Please start the services first with: docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d"
    exit 1
fi

# Create the backup
print_status "Creating database dump..."

if [[ "$COMPRESS" == true ]]; then
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" exec -T "$DB_CONTAINER" \
        pg_dump -U "$DB_USER" "$DB_NAME" | gzip > "$BACKUP_FILE"
else
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" exec -T "$DB_CONTAINER" \
        pg_dump -U "$DB_USER" "$DB_NAME" > "$BACKUP_FILE"
fi

if [[ $? -eq 0 ]]; then
    print_success "Database backup created: $BACKUP_FILE"
    
    # Get file size
    if command -v du >/dev/null 2>&1; then
        FILE_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
        print_status "Backup size: $FILE_SIZE"
    fi
else
    print_error "Database backup failed"
    exit 1
fi

# Clean up old backups
if [[ "$RETENTION_DAYS" -gt 0 ]]; then
    print_status "Cleaning up backups older than $RETENTION_DAYS days..."
    
    find "$BACKUP_DIR" -name "${ENVIRONMENT}_${DB_NAME}_*.sql*" -type f -mtime +$RETENTION_DAYS -delete
    
    DELETED_COUNT=$(find "$BACKUP_DIR" -name "${ENVIRONMENT}_${DB_NAME}_*.sql*" -type f -mtime +$RETENTION_DAYS | wc -l)
    if [[ $DELETED_COUNT -gt 0 ]]; then
        print_status "Deleted $DELETED_COUNT old backup files"
    fi
fi

# List recent backups
print_status "Recent backups:"
find "$BACKUP_DIR" -name "${ENVIRONMENT}_${DB_NAME}_*.sql*" -type f -exec ls -lh {} \; | tail -5

print_success "Backup process completed successfully"

# Show restore instructions
echo ""
print_status "To restore this backup:"
if [[ "$COMPRESS" == true ]]; then
    echo "  gunzip -c $BACKUP_FILE | docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec -T $DB_CONTAINER psql -U $DB_USER $DB_NAME"
else
    echo "  cat $BACKUP_FILE | docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec -T $DB_CONTAINER psql -U $DB_USER $DB_NAME"
fi
