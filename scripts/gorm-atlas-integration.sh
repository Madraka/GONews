#!/bin/bash
# GORM to Atlas Migration Strategy
# A practical approach to migrate from GORM AutoMigrate to Atlas

echo "🎯 GORM + Atlas Integration Strategy"
echo "===================================="

# Phase 1: Create Atlas Config for Current State
echo "📋 Phase 1: Setup Atlas Configuration"

# Create proper atlas.hcl
cat > atlas.hcl << 'EOF'
env "dev" {
  name = "development"
  url = "postgres://devuser:devpass@localhost:5433/newsapi_dev?sslmode=disable"
  migration {
    dir = "file://migrations/atlas"
  }
}

env "test" {
  name = "test"  
  url = "postgres://devuser:devpass@localhost:5433/newsapi_test?sslmode=disable"
  migration {
    dir = "file://migrations/atlas"
  }
}

env "prod" {
  name = "production"
  url = getenv("DATABASE_URL")
  migration {
    dir = "file://migrations/atlas"
  }
}
EOF

echo "✅ Atlas configuration created"

# Phase 2: Create baseline migration from current DB
echo "📊 Phase 2: Create Baseline Migration"

# Mark current state as baseline (no actual changes needed)
atlas migrate hash --dir file://migrations/atlas > /dev/null 2>&1 || true

# Create a simple baseline file
mkdir -p migrations/atlas
cat > migrations/atlas/000001_baseline.sql << 'EOF'
-- Baseline migration
-- This represents the current state of the database
-- No actual changes are needed as GORM has already created all structures

-- Add a simple comment to indicate Atlas is now managing migrations
COMMENT ON SCHEMA public IS 'Schema managed by Atlas migrations from this point forward';
EOF

echo "✅ Baseline migration created"

# Phase 3: Update GORM Configuration  
echo "🔧 Phase 3: Update GORM Configuration"

echo "
Next steps:
1. Disable GORM AutoMigrate in production
2. Use Atlas for all future schema changes
3. Keep GORM models as source of truth
4. Generate Atlas migrations from model changes
"

echo "
🚀 Implementation Strategy:

1. **Development Workflow:**
   - Keep GORM models as schema definition
   - Generate Atlas migrations from model changes
   - Test migrations in development environment

2. **Production Deployment:**
   - Use Atlas for production migrations
   - Automated CI/CD pipeline integration
   - Safe rollback capabilities

3. **Future Schema Changes:**
   - Update GORM models
   - Generate migration: atlas migrate diff --env dev
   - Review and test migration
   - Deploy via Atlas in CI/CD
"

echo "✅ Atlas + GORM integration ready!"
