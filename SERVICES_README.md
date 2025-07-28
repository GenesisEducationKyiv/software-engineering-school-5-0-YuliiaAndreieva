# Weather Services - Microservices Architecture

## Services Overview

This project contains 5 microservices:

1. **Email Service** (Port: 8081) - Handles email sending
2. **Subscription Service** (Port: 8082) - Manages user subscriptions
3. **Token Service** (Port: 8083) - Handles JWT token generation/validation
4. **Weather Service** (Port: 8084) - Provides weather data
5. **Weather Broadcast Service** (Port: 8085) - Broadcasts weather updates

## Quick Start

### Prerequisites
- Docker and Docker Compose installed
- WeatherAPI.com API key

### Setup Environment Variables

Create `.env` files for each service with required variables:

#### Email Service (`services/email/.env`)
```
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASS=your_app_password
PORT=8081
```

#### Subscription Service (`services/subscription/.env`)
```
PORT=8082
EMAIL_SERVICE_URL=http://email-service:8081
TOKEN_SERVICE_URL=http://token-service:8083
```

#### Token Service (`services/token/.env`)
```
PORT=8083
JWT_SECRET=your_jwt_secret_here
```

#### Weather Service (`services/weather/.env`)
```
WEATHER_API_KEY=your_weather_api_key_here
WEATHER_API_BASE_URL=http://api.weatherapi.com/v1
REDIS_ADDRESS=redis:6379
PORT=8084
```

#### Weather Broadcast Service (`services/weather-broadcast/.env`)
```
SUBSCRIPTION_SERVICE_URL=http://subscription-service:8082
WEATHER_SERVICE_URL=http://weather-service:8084
EMAIL_SERVICE_URL=http://email-service:8081
PORT=8085
WORKER_AMOUNT=10
PAGE_SIZE=100
HTTP_CLIENT_TIMEOUT=10s
HTTP_READ_TIMEOUT=10s
HTTP_WRITE_TIMEOUT=10s
```

### Start All Services

#### Option 1: Using Docker Compose (Recommended)
```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

#### Option 2: Using Scripts
```bash
# Linux/Mac
./start-services.sh

# Windows PowerShell
.\start-services.ps1
```

#### Option 3: Manual Commands
```bash
# Build all services
docker-compose build

# Start all services
docker-compose up -d

# Check status
docker-compose ps
```

## Service URLs

Once all services are running:

- **Email Service**: http://localhost:8081
- **Subscription Service**: http://localhost:8082
- **Token Service**: http://localhost:8083
- **Weather Service**: http://localhost:8084
- **Weather Broadcast Service**: http://localhost:8085
- **Redis**: localhost:6379

## Health Checks

Test if services are running:
```bash
curl http://localhost:8081/health  # Email Service
curl http://localhost:8082/health  # Subscription Service
curl http://localhost:8083/health  # Token Service
curl http://localhost:8084/health  # Weather Service
curl http://localhost:8085/health  # Weather Broadcast Service
```

## Individual Service Management

### Build individual service
```bash
docker-compose build email-service
docker-compose build subscription-service
docker-compose build token-service
docker-compose build weather-service
docker-compose build weather-broadcast-service
```

### Start individual service
```bash
docker-compose up -d email-service
docker-compose up -d subscription-service
docker-compose up -d token-service
docker-compose up -d weather-service
docker-compose up -d weather-broadcast-service
```

### View logs for specific service
```bash
docker-compose logs -f email-service
docker-compose logs -f subscription-service
docker-compose logs -f token-service
docker-compose logs -f weather-service
docker-compose logs -f weather-broadcast-service
```

## Troubleshooting

### Check if all containers are running
```bash
docker-compose ps
```

### View all logs
```bash
docker-compose logs -f
```

### Restart all services
```bash
docker-compose restart
```

### Clean up everything
```bash
docker-compose down -v --remove-orphans
docker system prune -f
```

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Email Service │    │ Subscription     │    │  Token Service  │
│   (Port: 8081)  │    │ Service          │    │  (Port: 8083)   │
└─────────────────┘    │ (Port: 8082)     │    └─────────────────┘
         ▲             └──────────────────┘             ▲
         │                      ▲                       │
         │                      │                       │
         │              ┌─────────────────┐             │
         │              │ Weather Service │             │
         │              │ (Port: 8084)    │             │
         │              └─────────────────┘             │
         │                      ▲                       │
         │                      │                       │
         └──────────────────────┼───────────────────────┘
                                │
                    ┌─────────────────────┐
                    │ Weather Broadcast   │
                    │ Service             │
                    │ (Port: 8085)       │
                    └─────────────────────┘
``` 