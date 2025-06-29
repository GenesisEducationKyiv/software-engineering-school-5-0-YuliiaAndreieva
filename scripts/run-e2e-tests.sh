#!/bin/bash

echo "Starting E2E Test Environment..."
docker-compose -f docker-compose.test.yml up -d postgres-e2e mailhog

echo "Waiting for services to be ready..."
sleep 10

echo "Installing Playwright dependencies..."
cd tests
npm install
npx playwright install

echo "Running E2E Tests..."
npm run test:e2e

echo "Stopping E2E Test Environment..."
cd ..
docker-compose -f docker-compose.test.yml down

echo "Press any key to continue..."
read -n 1 -s