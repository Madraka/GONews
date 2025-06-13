#!/bin/bash
# Automatic Atlas Migration Generator
# Converts GORM model changes to Atlas migrations using Docker

set -e

echo "🤖 Automatic Atlas Migration Generator (Docker-based)"
echo "===================================================="

# Environment setup
ENV=${ENV:-dev}
COMPOSE_FILE="deployments/${ENV}/docker-compose-${ENV}.yml"
DATABASE_URL="postgres://devuser:devpass@${ENV}_db:5432/newsapi_${ENV}?sslmode=disable"

# Functions
wait_for_db() {
    echo "⏳ Waiting for database connection..."
    for i in {1..30}; do
        # Use PostgreSQL container to check database readiness
        if docker-compose -f "$COMPOSE_FILE" run --rm --no-deps "${ENV}_db" sh -c "pg_isready -h ${ENV}_db -p 5432 -U devuser" >/dev/null 2>&1; then
            echo "✅ Database is ready!"
            return 0
        fi
        sleep 2
    done
    echo "❌ Database connection timeout"
    exit 1
}

backup_current_schema() {
    echo "💾 Backing up current schema..."
    if [ -f "schema/current_complete.hcl" ]; then
        cp "schema/current_complete.hcl" "schema/backup_$(date +%Y%m%d_%H%M%S).hcl"
    fi
}

run_gorm_migrate() {
    echo "🔧 Running GORM AutoMigrate via Docker..."
    
    # Run GORM AutoMigrate inside the API container
    docker-compose -f "$COMPOSE_FILE" exec -T "${ENV}_api" sh -c "
        export DB_MIGRATION_MODE=auto
        go run cmd/migrate/main.go
    "
    
    echo "✅ GORM AutoMigrate completed"
}

extract_new_schema() {
    echo "📊 Extracting new database schema via Docker..."
    
    # Use Atlas container to extract schema
    docker-compose -f "$COMPOSE_FILE" --profile atlas run --rm "${ENV}_atlas" \
        schema inspect \
            --url "${DATABASE_URL}" \
            --format "{{ hcl . }}" \
        > schema/new_state.hcl
    
    echo "✅ Schema extracted: schema/new_state.hcl"
}

generate_atlas_migration() {
    echo "📝 Generating Atlas migration via Docker..."
    
    # Create migration name
    TIMESTAMP=$(date +%Y%m%d_%H%M%S)
    MIGRATION_NAME="gorm_sync_${TIMESTAMP}"
    
    # Create a clean dev database URL for comparison
    CLEAN_DEV_URL="postgres://devuser:devpass@${ENV}_db:5432/newsapi_${ENV}_clean?sslmode=disable"
    
    # Create temporary clean database for diff generation
    echo "🧹 Creating temporary clean database for diff generation..."
    docker-compose -f "$COMPOSE_FILE" run --rm --no-deps "${ENV}_db" sh -c "
        export PGPASSWORD=devpass
        psql -h ${ENV}_db -U devuser -d newsapi_${ENV} -c 'DROP DATABASE IF EXISTS newsapi_${ENV}_clean;'
        psql -h ${ENV}_db -U devuser -d newsapi_${ENV} -c 'CREATE DATABASE newsapi_${ENV}_clean;'
    " || echo "⚠️  Clean database creation had warnings (likely already exists)"
    
    # Create Atlas migration diff using Docker with clean dev database
    docker-compose -f "$COMPOSE_FILE" --profile atlas run --rm "${ENV}_atlas" \
        migrate diff "$MIGRATION_NAME" \
            --env "docker-${ENV}" \
            --dev-url "${CLEAN_DEV_URL}" \
            --to "file://schema/new_state.hcl"
    
    # Clean up temporary database
    echo "🧹 Cleaning up temporary database..."
    docker-compose -f "$COMPOSE_FILE" run --rm --no-deps "${ENV}_db" sh -c "
        export PGPASSWORD=devpass
        psql -h ${ENV}_db -U devuser -d newsapi_${ENV} -c 'DROP DATABASE IF EXISTS newsapi_${ENV}_clean;'
    " || echo "⚠️  Cleanup had warnings"
    
    echo "✅ Migration created: $MIGRATION_NAME"
}

update_schema_file() {
    echo "🔄 Updating schema file..."
    cp schema/new_state.hcl schema/current_complete.hcl
    echo "✅ schema/current_complete.hcl updated"
}

commit_changes() {
    if [ "$1" = "--commit" ]; then
        echo "📝 Committing to git..."
        
        git add schema/current_complete.hcl
        git add migrations/atlas/
        
        COMMIT_MSG="🤖 Auto-sync: GORM → Atlas migration

Generated: $(date)
Migration: $MIGRATION_NAME

[auto-migration]"
        
        git commit -m "$COMMIT_MSG" || echo "ℹ️  No changes to commit"
        echo "✅ Changes committed"
    else
        echo "ℹ️  Skipping git commit (no --commit flag)"
    fi
}

# Main execution
main() {
    echo "🚀 Starting automatic migration process..."
    
    # Prerequisites check
    command -v docker >/dev/null 2>&1 || { echo "❌ Docker not found"; exit 1; }
    command -v docker-compose >/dev/null 2>&1 || { echo "❌ Docker Compose not found"; exit 1; }
    
    # Check if development environment is running
    if ! docker ps | grep -q "news_${ENV}_db"; then
        echo "❌ ${ENV} database is not running"
        echo "💡 Please run 'make ${ENV}' first"
        exit 1
    fi
    
    # Wait for database
    wait_for_db
    
    # Backup current schema
    backup_current_schema
    
    # Run GORM migration
    run_gorm_migrate
    
    # Extract new schema
    extract_new_schema
    
    # Check if there are differences
    if diff -q schema/current_complete.hcl schema/new_state.hcl >/dev/null 2>&1; then
        echo "ℹ️  No schema changes detected"
        echo "✅ Database is already up to date"
        rm schema/new_state.hcl
        exit 0
    fi
    
    echo "🔍 Schema changes detected!"
    
    # Generate Atlas migration
    generate_atlas_migration
    
    # Update schema file
    update_schema_file
    
    # Clean up
    rm schema/new_state.hcl
    
    # Commit if requested
    commit_changes "$1"
    
    echo "🎉 Automatic migration completed!"
    echo ""
    echo "📋 What was done:"
    echo "  ✅ GORM AutoMigrate executed"
    echo "  ✅ New schema extracted"
    echo "  ✅ Atlas migration created"
    echo "  ✅ Schema file updated"
    echo ""
    echo "💡 What you can do next:"
    echo "  • make atlas-status ENV=${ENV}  → Check migration status"
    echo "  • make atlas-apply ENV=${ENV}   → Apply the migration"
    echo "  • git push                      → Push changes"
}

# Run with all arguments
main "$@"
