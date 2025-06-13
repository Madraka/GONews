#!/bin/bash
# Atlas Migration Setup Script
# Run this to complete GORM → Atlas migration

echo "🚀 Starting GORM to Atlas Migration..."

# 1. Extract current schema from database
echo "📊 Extracting current database schema..."
atlas schema inspect \
  --url "postgres://newsuser:newspass@localhost:5432/newsapi_dev?sslmode=disable" \
  --format "{{ hcl . }}" \
  > schema/current_complete.hcl

# 2. Create baseline migration
echo "📝 Creating baseline migration..."
atlas migrate diff initial_baseline \
  --env dev \
  --to file://schema

# 3. Validate current schema
echo "✅ Validating schema..."
atlas schema validate --env dev

# 4. Mark baseline as applied (no-op for existing DB)
echo "🏷️  Marking baseline migration as applied..."
atlas migrate apply \
  --env dev \
  --baseline $(atlas migrate status --env dev | grep "Migration Files" | awk '{print $NF}')

echo "✅ Migration setup complete!"
echo "Now you can safely disable GORM AutoMigrate and use Atlas."
