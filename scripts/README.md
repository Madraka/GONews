# Scripts Directory

This directory contains organized scripts for database operations, seeding, testing, and development tasks.

## 📁 Directory Structure

```
scripts/
├── README.md                    # This file
├── active/                      # Currently used scripts
│   ├── seeding/                # Database seeding scripts
│   ├── testing/                # Testing and validation scripts
│   ├── monitoring/             # Monitoring and metrics scripts
│   └── utilities/              # General utility scripts
├── archive/                     # Archived/deprecated scripts
│   └── 20250529-scripts/       # Scripts archived on 2025-05-29
├── build/                       # Build-related scripts
├── migration/                   # Database migration utilities
├── performance/                 # Performance testing scripts
├── seeds/                       # Organized seed data
└── utilities/                   # Helper utilities
```

## 🚀 Usage

### Database Seeding
```bash
# Auto-detect environment and seed (via Makefile)
make seed-db

# Environment-specific seeding
make seed-db-dev     # Development database
make seed-db-prod    # Production database

# Direct script usage
ENVIRONMENT=dev ./active/seeding/seed_database.sh
ENVIRONMENT=prod ./active/seeding/seed_database.sh
```

### Database Migrations
```bash
# Run migrations (via Makefile)
make test-migrate              # Test environment migrations
make test-migrate-status       # Check migration status
make test-migrate-reset        # Reset and re-run migrations

# Direct migration usage
cd migration && ./migrate_safe.sh up
cd migration && ./health_check.sh
```

### Testing
```bash
# Run tests
./active/testing/test_migration_system.sh
./active/testing/test_model_alignment.sh
./active/testing/validate_tracing.sh
```

## 📂 Key Directories

### `active/`
Contains currently used, production-ready scripts:

- **seeding/**: `seed_database.sh`, `seed_database_new.sh`
- **testing/**: `test_*.sh`, `validate_*.sh`, `verify_*.sh`
- **monitoring/**: `check_metrics.sh`, `setup_monitoring.sh`
- **utilities/**: `cleanup.sh`, `docker-helper.sh`, `start_*.sh`

### `seeds/`
Organized seed data structure:
- `01_core/`: Users, categories, tags, settings
- `02_content/`: Articles, news content
- `03_real_time/`: Breaking news, live streams
- `04_interactions/`: Comments, votes, bookmarks
- `05_relationships/`: Junction tables
- `master_seed.sql`: Main seed orchestrator

### `migration/`
Database migration utilities:
- `migrate_safe.sh`: Safe migration with rollback
- `health_check.sh`: Database health verification

### `performance/`
Performance testing scripts:
- `load-test-*.js`: Various load testing scenarios
- `README.md`: Performance testing documentation

## 🗄️ Archive

The `archive/20250529-scripts/` directory contains 26 archived files including:
- Old migration scripts (`migrate.sh`, `migrate_new.sh`)
- Legacy SQL files (`*.sql`)
- Deprecated utility scripts
- Performance testing scripts moved to `performance/`

## 🔧 Environment Support

- **Development**: `newsapi_dev` database on port 5434
- **Testing**: `newsapi_test` database on port 8082  
- **Production**: `newsapi` database on port 5432

## 🔗 Makefile Integration

All active scripts are integrated with the project Makefile:

```bash
make lightweight-test          # Start test environment
make test-migrate             # Run migrations on test DB
make test-migrate-status      # Check migration status
make env-dev                  # Switch to development config
make env-test                 # Switch to testing config
make clean                    # Clean up Docker resources
```

## 📝 Notes

- All scripts support environment detection via `ENVIRONMENT` variable
- Scripts are organized by function for better maintainability
- Archive contains historical scripts for reference
- Active scripts are production-ready and regularly maintained
