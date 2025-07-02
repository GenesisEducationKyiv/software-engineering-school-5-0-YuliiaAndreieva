#!/bin/bash

echo "Starting Integration Test Environment..."
docker-compose -f docker-compose.test.yml up -d postgres-integration mailhog

echo "Waiting for services to be ready..."
sleep 10

echo "Running Integration Tests..."
go test -v ./tests/integration/... -tags=integration -count=1

echo "Stopping Integration Test Environment..."
docker-compose -f docker-compose.test.yml down

echo "Press any key to continue..."
read -n 1 -s