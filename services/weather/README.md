# Weather Service

## Purpose
External weather data integration and caching

## Responsibilities
- Fetch weather data from external APIs
- Cache weather results to reduce API calls
- Provide weather data to other services
- Validate city names against external APIs

## Core Functions
- Weather data retrieval from WeatherAPI.com
- Redis caching for weather results (30 min TTL)
- City validation against external API
- Chain of responsibility pattern for multiple providers
- Graceful fallback between providers

## Technology Stack
- Go with Gin web framework
- Redis for caching
- WeatherAPI.com integration
- Logrus for structured logging
- Ports and Adapters (Hexagonal) architecture

## API Endpoints

### POST /weather
Get weather data for a city

Request:
```json
{
  "city": "Kyiv"
}
```

Response:
```json
{
  "success": true,
  "weather": {
    "city": "Kyiv",
    "temperature": 15.5,
    "humidity": 65,
    "description": "Partly cloudy",
    "wind_speed": 12.3,
    "timestamp": "2025-07-27T18:30:00Z"
  },
  "message": "Weather data retrieved successfully"
}
```

### POST /validate-city
Validate if a city exists

Request:
```json
{
  "city": "Kyiv"
}
```

Response:
```json
{
  "success": true,
  "valid": true,
  "message": "City is valid"
}
```

### GET /health
Health check endpoint

## Environment Variables
- `SERVER_PORT`: Server port (default: 8084)
- `WEATHER_API_KEY`: WeatherAPI.com API key (required)
- `REDIS_ADDR`: Redis server address (default: localhost:6379)
- `REDIS_PASSWORD`: Redis password (optional)

## Running the Service

### Using Docker
```bash
docker build -t weather-service .
docker run -d --name weather-service -p 8084:8084 \
  -e WEATHER_API_KEY=your-api-key \
  -e REDIS_ADDR=redis:6379 \
  weather-service
```

### Using Go
```bash
go mod tidy
go run ./cmd/weather-service/main.go
```

## Architecture
- **Domain**: Core business entities and rules
- **Ports/In**: Use case interfaces
- **Ports/Out**: Service interfaces (WeatherProvider, Cache, Logger)
- **Use Cases**: Business logic implementation
- **Adapters/HTTP**: HTTP handlers and client
- **Adapters/Weather**: Weather provider implementations
- **Adapters/Cache**: Redis cache implementation
- **Adapters/Logger**: Logging implementation
- **Config**: Configuration management

## Features
- **Caching**: Redis cache with 30-minute TTL
- **Provider Chain**: Fallback between multiple weather providers
- **City Validation**: Validate city names against external APIs
- **Structured Logging**: Detailed logging for debugging
- **Graceful Shutdown**: Proper cleanup on shutdown 