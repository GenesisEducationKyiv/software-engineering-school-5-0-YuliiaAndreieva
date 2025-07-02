#!/bin/bash

set -e

echo "Running All Tests..."

echo "Running Unit Tests..."
./scripts/run-unit-tests.sh

echo "Running Integration Tests..."
./scripts/run-integration-tests.sh

echo "Running E2E Tests..."
./scripts/run-e2e-tests.sh

echo "All tests completed!"

echo "Press any key to continue..."
read -n 1 -s