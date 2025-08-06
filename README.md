# Weather API Service

A Go-based weather API service that provides weather updates via email subscriptions. Users can subscribe to receive weather updates for their preferred cities either hourly or daily.

### 🚨 Please be aware that site is deployed at https://weatherapi-rpum.onrender.com (currently unavailable) 🚨
## Architecture

The project follows a clean architecture pattern with the following structure:

```
weather-api/
├── cmd/
│   └── server/         # Application entry point
├── internal/
│   ├── adapter/        # External service adapters (email, weather API, database)
│   ├── core/           # Domain models and business logic
│   ├── handler/        # HTTP handlers
│   ├── mocks/          # Mock implementations for testing
│   └── util/           # Utility functions and configuration
├── migrations/         # Database migrations
└── web/               # Frontend static files
```

### Key Components

- **Weather Service**: Fetches weather data from external weather API
- **Email Service**: Handles email notifications using SMTP
- **Subscription Service**: Manages user subscriptions and confirmation
- **Token Service**: Generates subscription tokens
- **PostgreSQL Database**: Stores subscription information

## Prerequisites

- Go 1.24
- PostgreSQL
- SMTP server access (google app generated pass)
- Weather API key (https://www.weatherapi.com)

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
DB_CONN_STR=postgres://username:password@localhost:5432/weather_db?sslmode=disable
WEATHER_API_KEY=your_weather_api_key
BASE_URL=http://localhost:8080
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASS=your_app_specific_password
PORT=8080
```

## Running the Project

1. Start the server and postgres db using docker:
```bash
docker compose up -d
```
Migrations will be automatically applied upon lift.
The server will start on port 8080 by default.

## Running Tests

Run all tests:
```bash
go test ./...
```

Run tests only in core/service directory:
```bash
go test ./internal/core/service/...
```

### Code Quality with golangci-lint

Install golangci-lint:

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
```

Run the linter:
```bash
golangci-lint run -v
```

## API Endpoints

- `GET /api/weather` - Get current weather for a city
- `POST /api/subscribe` - Subscribe to weather updates
- `GET /api/confirm/:token` - Confirm subscription
- `GET /api/unsubscribe/:token` - Unsubscribe from updates

## Subscription Frequencies

The service supports two types of update frequencies:

1. **Hourly Updates**: Sent every hour
2. **Daily Updates**: Sent once daily at 8:00 AM

**Please be aware, that after click button Subscribe - on ui only button changes color and email sent, no alerts**
To modify the update schedule, edit the cron expressions in `cmd/server/main.go`:

```go
cron.AddFunc("0 * * * *", func() { emailService.SendUpdates(context.Background(), domain.FrequencyHourly) })
cron.AddFunc("0 0 * * *", func() { emailService.SendUpdates(context.Background(), domain.FrequencyDaily) })
```

## Example Subscription Request

```json
{
    "email": "user@example.com",
    "city": "Kyiv",
    "frequency": "hourly"
}
```