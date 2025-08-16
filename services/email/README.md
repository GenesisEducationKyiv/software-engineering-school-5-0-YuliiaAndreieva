# Email Microservice

A standalone email delivery microservice for the Weather API project following **Ports and Adapters** (Hexagonal) architecture.

## Architecture

The microservice follows Clean Architecture principles with the following structure:

```
services/email/
├── cmd/
│   └── email-service/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── core/
│   │   ├── domain/              # Domain models and entities
│   │   ├── ports/
│   │   │   ├── in/              # Input ports (interfaces)
│   │   │   └── out/             # Output ports (interfaces)
│   │   └── usecase/             # Business logic (use cases)
│   └── adapter/
│       ├── email/               # Email adapters (SMTP, templates)
│       └── http/                # HTTP handlers
├── Dockerfile                   # Docker configuration
├── go.mod                       # Go module
├── README.md                    # Documentation
└── env.example                  # Configuration example
```

## Purpose

Email delivery and management for the weather subscription system.

## Responsibilities

- Send confirmation emails
- Send weather update emails
- Handle email delivery failures
- Manage email templates
- Validate email addresses during sending

## Core Functions

- SMTP integration
- Email template rendering
- Delivery status tracking
- Email validation and bounce handling
- Email existence verification during delivery

## Technology Stack

- Go with SMTP integration
- Gin web framework
- Email template engine
- SMTP client
- Ports and Adapters architecture

## API Endpoints

### Health Check
```
GET /health
```
Returns service status.

### Send Confirmation Email
```
POST /send/confirmation
Content-Type: application/json

{
  "to": "user@example.com",
  "city": "Kyiv",
  "token": "confirmation-token"
}
```

### Send Weather Update Email
```
POST /send/weather-update
Content-Type: application/json

{
  "to": "user@example.com",
  "city": "Kyiv",
  "temperature": 20.5,
  "humidity": 60,
  "description": "Sunny",
  "token": "unsubscribe-token"
}
```

## Environment Variables

```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-specific-password
PORT=8081
BASE_URL=http://localhost:8080
```

## Running the Service

### Local Development
```bash
cd services/email
go mod tidy
go run cmd/email-service/main.go
```

### Docker
```bash
docker build -f services/email/Dockerfile -t weather-api-email .
docker run -p 8081:8081 weather-api-email
```

## Architecture Benefits

- **Separation of Concerns**: Business logic is separated from infrastructure
- **Testability**: Easy to mock dependencies and test business logic
- **Flexibility**: Easy to swap implementations (e.g., different email providers)
- **Maintainability**: Clear boundaries between layers
- **Scalability**: Independent deployment and scaling

## Integration

This microservice is designed to be called by the main weather API service for email delivery operations. The main service will make HTTP requests to this microservice's endpoints to send emails.

## Future Enhancements

- RabbitMQ integration for async email processing
- Email delivery status tracking
- Bounce handling
- Email template management
- Rate limiting
- Retry mechanisms
- Multiple email provider support 