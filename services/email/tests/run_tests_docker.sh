#!/bin/bash

set -e

echo "🚀 Starting Email Service Tests in Docker Environment..."
echo "======================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Stop any existing test containers
print_status "Stopping existing test containers..."
docker-compose -f docker-compose.test.yml down -v 2>/dev/null || true

# Start test environment
print_status "Starting test environment with MailHog..."
docker-compose -f docker-compose.test.yml up -d

# Wait for services to be ready
print_status "Waiting for services to be ready..."
sleep 10

# Check if MailHog is running
if ! curl -s http://localhost:8025 > /dev/null; then
    print_error "MailHog is not accessible. Please check if it's running on port 8025."
    exit 1
fi

print_success "MailHog is running on http://localhost:8025"

# Set test environment variables
export SMTP_HOST=localhost
export SMTP_PORT=1025
export SMTP_USER=test@example.com
export SMTP_PASS=testpass
export SERVER_PORT=8081
export SERVER_BASE_URL=http://localhost:8081

print_status "Running unit tests..."
if go test ./tests/unit/... -v; then
    print_success "Unit tests passed!"
else
    print_error "Unit tests failed!"
    exit 1
fi

print_status "Running integration tests..."
if go test ./tests/integration/... -v; then
    print_success "Integration tests passed!"
else
    print_error "Integration tests failed!"
    exit 1
fi

# Run coverage
print_status "Running coverage analysis..."
go test ./tests/... -coverprofile=coverage.out -covermode=atomic
go tool cover -html=coverage.out -o coverage.html

print_success "Coverage report generated: coverage.html"

# Show MailHog UI info
print_status "Test emails can be viewed at: http://localhost:8025"

# Cleanup
print_status "Cleaning up test environment..."
docker-compose -f docker-compose.test.yml down -v

print_success "All tests completed successfully!" 