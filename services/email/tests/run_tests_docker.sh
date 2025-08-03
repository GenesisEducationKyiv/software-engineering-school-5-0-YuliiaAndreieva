#!/bin/bash

echo "🚀 Starting Email Service Tests..."
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

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Start test environment
print_status "Starting MailHog for email testing..."
docker-compose -f tests/docker-compose.test.yml up -d

# Wait for services to be ready
print_status "Waiting for services to be ready..."
sleep 5

print_status "Running unit tests..."
go test ./tests/unit/... -v -count=1

print_status "Running integration tests..."
go test ./tests/integration/... -v -count=1

# Show MailHog UI info
print_status "Test emails can be viewed at: http://localhost:8025"

print_success "All tests completed!"

# Wait for user input before closing
echo ""
echo "======================================================"
echo "Press any key to close this window..."
read -n 1 -s 