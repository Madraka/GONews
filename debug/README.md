# Debug Tools Directory

This directory contains debug tools for the News API service.

## Files:

- `debug_server.go` - A simple debug server that logs detailed information about incoming requests
- `debug_client.go` - A test client that sends requests to both the main API and debug server

## Usage

You can run these tools using the Makefile targets:

```bash
# Start the debug server
make debug-server

# Run the debug client
make debug-client
```

Or run them directly:

```bash
# Go to debug directory
cd debug

# Start the debug server
go run debug_server.go

# In another terminal window, run the debug client
go run debug_client.go
```

## Debug Server

The debug server runs on port 8090 and logs detailed information about incoming requests. It's useful for debugging issues with request formatting, headers, and payloads.

## Debug Client

The debug client sends test requests to both the News API and debug server, allowing you to see how the requests are processed by both.
