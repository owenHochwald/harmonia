#!/bin/bash

# Test database setup script for Harmonia

set -e

echo "🧪 Setting up test environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Start test database
echo "🐘 Starting test PostgreSQL database..."
docker-compose -f docker-compose.test.yml up -d

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
max_attempts=30
attempt=0

while ! docker exec harmonia-test pg_isready -U postgres > /dev/null 2>&1; do
    attempt=$((attempt + 1))
    if [ $attempt -ge $max_attempts ]; then
        echo "❌ Database failed to start after $max_attempts attempts"
        docker-compose -f docker-compose.test.yml logs postgres-test
        exit 1
    fi
    echo "Waiting... (attempt $attempt/$max_attempts)"
    sleep 1
done

echo "✅ Test database is ready!"

# Run tests
if [ "$1" = "run" ]; then
    echo "🏃 Running tests..."
    go test ./internal/repo -v
fi

echo "🎉 Test setup complete!"
echo ""
echo "Commands:"
echo "  Start test DB:     docker-compose -f docker-compose.test.yml up -d"
echo "  Stop test DB:      docker-compose -f docker-compose.test.yml down"
echo "  Run tests:         go test ./internal/repo -v"
echo "  Run with setup:    ./scripts/test-setup.sh run"