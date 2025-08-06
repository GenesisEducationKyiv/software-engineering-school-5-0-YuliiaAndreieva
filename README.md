# Weather Services

Microservices architecture for weather notifications with gRPC communication between services.

## Architecture

```mermaid
graph TB
    Client[Client] <-->|HTTP| Gateway[Gateway]
    Gateway -->|HTTP| WeatherBroadcast[Weather-Broadcast]
    Gateway -->|HTTP| Subscription[Subscription]
    Gateway -->|HTTP| Token[Token]
    Gateway -->|HTTP| Weather[Weather]
    
    WeatherBroadcast -->|gRPC| Subscription
    WeatherBroadcast -->|gRPC| Email[Email]
    WeatherBroadcast -->|gRPC| Weather
    
    Subscription -->|gRPC| Email
    
    subgraph "Services"
        Gateway
        WeatherBroadcast
        Subscription
        Token
        Weather
        Email
    end
```

**Services:**
- **Gateway** (8080) - API Gateway (HTTP)
- **Email** (8081 HTTP / 9091 gRPC) - Email notifications
- **Subscription** (8082 HTTP / 9093 gRPC) - Subscription management
- **Token** (8083) - JWT tokens (HTTP)
- **Weather** (8084 HTTP / 9092 gRPC) - Weather data
- **Weather-Broadcast** (8085) - Weather broadcast service (HTTP)

## Environment Configuration

### Gateway Service (.env)
```bash
PORT=8080
WEATHER_SERVICE_URL=http://weather-service:8084
SUBSCRIPTION_SERVICE_URL=http://subscription-service:8082
HTTP_CLIENT_TIMEOUT=30s
SHUTDOWN_TIMEOUT=5s
LOG_INITIAL=100
LOG_THEREAFTER=100
LOG_TICK=1s
```

### Email Service (.env)
```bash
# Server Configuration
SERVER_PORT=8081
GRPC_PORT=9091
BASE_URL=http://email-service:8081
SUBSCRIPTION_SERVICE_URL=http://subscription-service:8082

# SMTP Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password

# RabbitMQ Configuration
RABBITMQ_URL=amqp://admin:password@rabbitmq:5672/
RABBITMQ_EXCHANGE=subscription_events
RABBITMQ_QUEUE=email_notifications

# Timeout Configuration
SHUTDOWN_TIMEOUT=5s

# Logging Configuration
LOG_INITIAL=100
LOG_THEREAFTER=100
LOG_TICK=1s
```

### Subscription Service (.env)
```bash
# Server Configuration
SERVER_PORT=8082
GRPC_PORT=9093
BASE_URL=http://subscription-service:8082

# Token Service Configuration
TOKEN_SERVICE_URL=http://token-service:8083
TOKEN_EXPIRATION=24h

# Database Configuration
DATABASE_DSN=host=postgres user=postgres password=postgres dbname=subscriptions port=5432 sslmode=disable

# RabbitMQ Configuration
RABBITMQ_URL=amqp://admin:password@rabbitmq:5672/
RABBITMQ_EXCHANGE=subscription_events
RABBITMQ_QUEUE=email_notifications

# Timeout Configuration
HTTP_CLIENT_TIMEOUT=10s
SHUTDOWN_TIMEOUT=5s
DATABASE_RETRY_DELAY=2s
DATABASE_MAX_RETRIES=30

# Logging Configuration
LOG_INITIAL=100
LOG_THEREAFTER=100
LOG_TICK=1s
```

### Token Service (.env)
```bash
# Server Configuration
SERVER_PORT=8083

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here

# Timeout Configuration
SHUTDOWN_TIMEOUT=5s

# Logging Configuration
LOG_INITIAL=100
LOG_THEREAFTER=100
LOG_TICK=1s
```

### Weather Service (.env)
```bash
# Server Configuration
PORT=8084
GRPC_PORT=9092
BASE_URL=http://weather-service:8084

# Weather API Configuration
WEATHER_API_KEY=your-weather-api-key
WEATHER_API_BASE_URL=http://api.weatherapi.com/v1
OPENWEATHERMAP_API_KEY=your-openweathermap-api-key
OPENWEATHERMAP_BASE_URL=https://api.openweathermap.org/data/2.5

# Redis Configuration
REDIS_ADDRESS=redis:6379
REDIS_TTL=30m
REDIS_DIAL_TIMEOUT=5s
REDIS_READ_TIMEOUT=3s
REDIS_WRITE_TIMEOUT=3s
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONNS=5

# Timeout Configuration
HTTP_READ_TIMEOUT=10s
HTTP_WRITE_TIMEOUT=10s
SHUTDOWN_TIMEOUT=5s
HTTP_CLIENT_TIMEOUT=10s

# Logging Configuration
LOG_INITIAL=100
LOG_THEREAFTER=100
LOG_TICK=1s
```

### Weather-Broadcast Service (.env)
```bash
# Server Configuration
PORT=8085

# Service URLs
SUBSCRIPTION_SERVICE_URL=http://subscription-service:8082
WEATHER_SERVICE_URL=http://weather-service:8084
EMAIL_SERVICE_URL=http://email-service:8081

# gRPC URLs
SUBSCRIPTION_GRPC_URL=subscription-service:9093
EMAIL_GRPC_URL=email-service:9091
WEATHER_GRPC_URL=weather-service:9092

# Worker Configuration
WORKER_AMOUNT=10
PAGE_SIZE=100

# Timeout Configuration
HTTP_CLIENT_TIMEOUT=10s
HTTP_READ_TIMEOUT=10s
HTTP_WRITE_TIMEOUT=10s
SHUTDOWN_TIMEOUT=5s

# Logging Configuration
LOG_INITIAL=100
LOG_THEREAFTER=100
LOG_TICK=1s
```

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