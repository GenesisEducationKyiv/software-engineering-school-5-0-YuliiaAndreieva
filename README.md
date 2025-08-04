# Weather Services

Microservices architecture for weather notifications with gRPC communication between services.

## Architecture

```
┌─────────────────┐    HTTP    ┌─────────────────┐
│   Gateway       │◄──────────►│   Client        │
└─────────────────┘            └─────────────────┘
         │
         │ HTTP
         ▼
┌─────────────────┐    gRPC    ┌─────────────────┐    gRPC    ┌─────────────────┐
│ Weather-Broadcast│◄──────────►│   Subscription  │◄──────────►│     Email       │
└─────────────────┘            └─────────────────┘            └─────────────────┘
         │                              │
         │ gRPC                         │
         ▼                              │
┌─────────────────┐                     │
│    Weather      │◄────────────────────┘
└─────────────────┘
```

**Services:**
- **Gateway** (8080) - API Gateway
- **Email** (8081/9091) - Email notifications
- **Subscription** (8082/9090) - Subscription management
- **Token** (8083) - JWT tokens
- **Weather** (8084/9092) - Weather data
- **Weather-Broadcast** (8085) - Weather broadcast service

## Build & Run

### Docker
```bash
# Build all services
docker-compose build

# Start
docker-compose up

# Stop
docker-compose down
```

### Local Development
```bash
# Generate proto files
cd proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative subscription/subscription.proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative email/email.proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative weather/weather.proto
go mod tidy

# Run services (separately)
cd services/email && go run cmd/email-service/main.go
cd services/subscription && go run cmd/subscription-service/main.go
cd services/weather && go run cmd/weather-service/main.go
cd services/weather-broadcast && go run cmd/weather-broadcast-service/main.go
```

## Testing

### Generate Mocks
```bash
# Email service
cd services/email
mockgen -source=internal/core/ports/in/email_usecase.go -destination=tests/mocks/email_usecase_mock.go

# Subscription service  
cd services/subscription
mockgen -source=internal/core/ports/in/subscription_usecase.go -destination=tests/mocks/subscription_usecase_mock.go

# Weather service
cd services/weather
mockgen -source=internal/core/ports/in/weather_usecase.go -destination=tests/mocks/weather_usecase_mock.go

# Weather-broadcast service
cd services/weather-broadcast
mockgen -source=internal/core/ports/out/subscription_client.go -destination=tests/mocks/subscription_client_mock.go
mockgen -source=internal/core/ports/out/email_client.go -destination=tests/mocks/email_client_mock.go
mockgen -source=internal/core/ports/out/weather_client.go -destination=tests/mocks/weather_client_mock.go
```

### Run Tests
```bash
# Unit tests
cd services/email && go test ./tests/unit/...
cd services/subscription && go test ./tests/usecase/...
cd services/weather && go test ./tests/unit/...
cd services/weather-broadcast && go test ./tests/usecase/...

# Integration tests
cd services/email && go test ./tests/integration/...
cd services/subscription && go test ./tests/integration/...
cd services/weather && go test ./tests/integration/...

# All service tests
cd services/email && go test ./...
cd services/subscription && go test ./...
cd services/weather && go test ./...
cd services/weather-broadcast && go test ./...
```

### Docker Tests
```bash
# Email service tests
cd services/email/tests
docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit

# Subscription service tests  
cd services/subscription/tests
docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
```

## Technologies

- **Go 1.24** - main language
- **gRPC** - inter-service communication
- **HTTP/REST** - external APIs
- **PostgreSQL** - database
- **Redis** - caching
- **Docker** - containerization
- **Gin** - HTTP framework
- **GORM** - ORM
- **Logrus** - logging
- **Cron** - task scheduling 