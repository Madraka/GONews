#!/bin/sh
# Atlas Migration Entrypoint Script
# Handles Atlas migration operations in Docker container

set -e

echo "🎯 Atlas Migration Container Starting..."

# Wait for database to be ready
wait_for_db() {
    local db_host=$1
    local db_port=$2
    local max_attempts=30
    local attempt=1
    
    echo "⏳ Waiting for database at $db_host:$db_port..."
    
    while [ $attempt -le $max_attempts ]; do
        if pg_isready -h "$db_host" -p "$db_port" >/dev/null 2>&1; then
            echo "✅ Database is ready!"
            return 0
        fi
        
        echo "🔄 Attempt $attempt/$max_attempts - Database not ready, waiting..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "❌ Database connection timeout after $max_attempts attempts"
    exit 1
}

# Extract database connection info from environment
extract_db_info() {
    if [ -n "$DATABASE_URL" ]; then
        # Parse DATABASE_URL format: postgres://user:pass@host:port/db
        DB_HOST=$(echo "$DATABASE_URL" | sed -E 's|^[^:]+://[^@]+@([^:/]+).*|\1|')
        DB_PORT=$(echo "$DATABASE_URL" | sed -E 's|^[^:]+://[^@]+@[^:]+:([0-9]+).*|\1|')
        
        # Default port if not specified
        if [ "$DB_PORT" = "$DATABASE_URL" ]; then
            DB_PORT=5432
        fi
    else
        # Use individual environment variables
        DB_HOST=${DB_HOST:-localhost}
        DB_PORT=${DB_PORT:-5432}
    fi
}

# Main execution
main() {
    local command=${1:-status}
    
    echo "🎯 Atlas Command: $command"
    echo "📊 Environment: ${ATLAS_ENV:-dev}"
    
    # Extract database connection info
    extract_db_info
    
    # Wait for database to be ready
    wait_for_db "$DB_HOST" "$DB_PORT"
    
    # Execute Atlas command
    case "$command" in
        "status")
            echo "📊 Checking migration status..."
            atlas migrate status --env "${ATLAS_ENV:-dev}"
            ;;
        "apply")
            echo "🚀 Applying migrations..."
            atlas migrate apply --env "${ATLAS_ENV:-dev}"
            ;;
        "diff")
            echo "📝 Creating migration diff..."
            atlas migrate diff --env "${ATLAS_ENV:-dev}"
            ;;
        "validate")
            echo "✅ Validating schema..."
            atlas schema validate --env "${ATLAS_ENV:-dev}"
            ;;
        "hash")
            echo "🔄 Updating migration hash..."
            atlas migrate hash --dir file://migrations/atlas
            ;;
        *)
            echo "🎯 Running custom Atlas command: $*"
            atlas "$@"
            ;;
    esac
    
    echo "✅ Atlas operation completed!"
}

# Run main function with all arguments
main "$@"
