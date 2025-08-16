# Token Service

## Purpose
JWT token generation and validation service

## Responsibilities
- Generate secure JWT tokens
- Validate JWT tokens
- Manage token claims and expiration

## Core Functions
- JWT token generation with custom claims
- Token validation with signature verification
- Configurable token expiration
- Secure token signing with HMAC-SHA256

## Technology Stack
- Go with Gin web framework
- JWT library (golang-jwt/jwt/v5)
- Logrus for structured logging
- Ports and Adapters (Hexagonal) architecture

## API Endpoints

### POST /generate
Generate a new JWT token

Request:
```json
{
  "subject": "user123",
  "expires_in": "24h",
  "claims": {
    "email": "user@example.com",
    "role": "user"
  }
}
```

Response:
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Token generated successfully"
}
```

### POST /validate
Validate an existing JWT token

Request:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Response:
```json
{
  "success": true,
  "valid": true,
  "claims": {
    "sub": "user123",
    "exp": 1640995200,
    "iat": 1640908800,
    "jti": "abc123",
    "email": "user@example.com",
    "role": "user"
  },
  "message": "Token is valid"
}
```

### GET /health
Health check endpoint

## Environment Variables
- `SERVER_PORT`: Server port (default: 8083)
- `JWT_SECRET`: Secret key for JWT signing (default: your-secret-key)

## Running the Service

### Using Docker
```bash
docker build -t token-service .
docker run -d --name token-service -p 8083:8083 -e JWT_SECRET=your-secret-key token-service
```

### Using Go
```bash
go mod tidy
go run ./cmd/token-service/main.go
```

## Architecture
- **Domain**: Core business entities and rules
- **Ports/In**: Use case interfaces
- **Ports/Out**: Service interfaces
- **Use Cases**: Business logic implementation
- **Adapters/HTTP**: HTTP handlers
- **Adapters/Logger**: Logging implementation
- **Config**: Configuration management 