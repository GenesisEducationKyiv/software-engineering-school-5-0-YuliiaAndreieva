# Subscription Service

## Purpose
Subscription management for weather updates

## Responsibilities
- Handle user subscriptions to weather updates
- Confirm email subscriptions
- Allow users to unsubscribe
- Manage subscription tokens

## Core Functions
- Subscription creation with email confirmation
- Token-based subscription confirmation
- Subscription deletion (unsubscribe)
- Email service integration for confirmations

## Technology Stack
- Go with Gin web framework
- Logrus for structured logging
- Ports and Adapters (Hexagonal) architecture

## API Endpoints

### POST /subscribe
Subscribe to weather updates for a city

Request:
```json
{
  "email": "user@example.com",
  "city": "Kyiv",
  "frequency": "daily"
}
```

Response:
```json
{
  "success": true,
  "message": "Subscription successful. Confirmation email sent.",
  "token": "abc123..."
}
```

### GET /confirm/:token
Confirm subscription using token

Response:
```json
{
  "success": true,
  "message": "Subscription confirmed successfully"
}
```

### GET /unsubscribe/:token
Unsubscribe using token

Response:
```json
{
  "success": true,
  "message": "Successfully unsubscribed"
}
```

### GET /health
Health check endpoint

## Environment Variables
- `SERVER_PORT`: Server port (default: 8082)
- `EMAIL_SERVICE_URL`: Email service URL (default: http://localhost:8081)

## Running the Service

### Using Docker
```bash
docker build -t subscription-service .
docker run -d --name subscription-service -p 8082:8082 subscription-service
```

### Using Go
```bash
go mod tidy
go run ./cmd/subscription-service/main.go
```

## Architecture
- **Domain**: Core business entities and rules
- **Ports/In**: Use case interfaces
- **Ports/Out**: Repository and service interfaces
- **Use Cases**: Business logic implementation
- **Adapters/HTTP**: HTTP handlers
- **Adapters/Logger**: Logging implementation
- **Config**: Configuration management 