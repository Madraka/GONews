# Project Reorganization Summary

## Changes Made

1. **Created Dedicated Folders**
   - Created `/debug` folder for debug-related tools
   - Created `/tests` folder for test scripts and tools
   - Created `/docker` folder for Docker configuration files
   - Enhanced organization of monitoring files in `/monitoring` folder

2. **Moved Files to Appropriate Locations**
   - Moved debug tools (`debug_client.go`, `debug_server.go`) to `/debug` folder
   - Moved test scripts (`test_*.go`, `test_observability.sh`) to `/tests` folder
   - Moved Docker files (`Dockerfile*`, `docker-compose*.yml`) to `/docker` folder
   - Moved `prometheus.yml` configuration file to `/monitoring` folder

3. **Added Documentation**
   - Created README.md for `/debug`, `/tests`, and `/docker` folders
   - Updated main README.md with the new folder structure

4. **Updated Makefile**
   - Added targets to run tools from their new locations:
     - `make debug-server` to run the debug server
     - `make debug-client` to run the debug client
     - `make test-api` to run API tests
     - `make test-observability` to run observability tests

5. **Created Helper Scripts**
   - Added `run_tests.sh` in the `/tests` folder to run each test individually
   - Created `docker-helper.sh` in the `scripts` folder to simplify Docker operations
   - Created symbolic links to maintain backward compatibility

## Benefits

1. **Better Organization**: Code is now more logically organized by purpose
2. **Improved Maintenance**: Easier to find and maintain related files
3. **Cleaner Root Directory**: Reduced clutter in the project root
4. **Documented Purpose**: README files explain the purpose of each folder
5. **Consistent Interface**: Makefile targets provide a consistent way to run tools

## Usage

To work with the reorganized structure:

```bash
# Run debug tools
make debug-server
make debug-client

# Run tests
make test-api
make test-observability

# Or run individual tests directly
cd tests
./run_tests.sh

# Run Docker containers
make docker-build      # Build Docker image using docker/Dockerfile
make docker-build-dev  # Build development Docker image using docker/Dockerfile.dev
make docker-run        # Run production containers using docker/docker-compose.yml
make docker-run-dev    # Run development containers using docker/docker-compose-dev.yml

# Use Docker helper script
make docker-helper     # Shows available Docker helper commands
./scripts/docker-helper.sh build   # Build production image
./scripts/docker-helper.sh up-dev  # Start development environment
```
