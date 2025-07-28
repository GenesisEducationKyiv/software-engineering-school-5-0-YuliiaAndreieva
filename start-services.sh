#!/bin/bash

echo "Starting all weather services..."

# Build all services
echo "Building all services..."
docker-compose build

# Start all services
echo "Starting all services..."
docker-compose up -d

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 10

# Check service status
echo "Checking service status..."
docker-compose ps

echo "All services are running!"
echo ""
echo "Service URLs:"
echo "Email Service: http://localhost:8081"
echo "Subscription Service: http://localhost:8082"
echo "Token Service: http://localhost:8083"
echo "Weather Service: http://localhost:8084"
echo "Weather Broadcast Service: http://localhost:8085"
echo "Redis: localhost:6379"
echo ""
echo "To stop all services: docker-compose down"
echo "To view logs: docker-compose logs -f"

echo "Press any key to continue..."
read -n 1 -s