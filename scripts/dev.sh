#!/usr/bin/env bash

# Development script for gitmirror

set -e

# Build the project
echo "Building the project..."
go build -o gitmirror ./cmd/gitmirror

# Format code and README
echo "Formatting source files..."
./scripts/format.sh

# Run tests
echo "Running tests..."
go test ./...

# Start the application
echo "Starting the application..."
./gitmirror --help

# Additional commands can be added here for further development tasks.