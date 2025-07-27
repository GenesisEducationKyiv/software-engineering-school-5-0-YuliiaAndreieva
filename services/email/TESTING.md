# Email Service Testing Guide

## Prerequisites
- Docker installed and running
- Container running on port 8081

## Running the Service

1. Build the Docker image:
```bash
docker build -t weather-api-email .
```

2. Run the container with environment variables:
```bash
docker run -d --name email-service -p 8081:8081 \
  -e SMTP_HOST=smtp.gmail.com \
  -e SMTP_PORT=587 \
  -e SMTP_USER=your-email@gmail.com \
  -e SMTP_PASS=your-app-password \
  -e SERVER_PORT=8081 \
  -e BASE_URL=http://localhost:8081 \
  weather-api-email
```

## Testing Endpoints

### 1. Health Check
```bash
curl http://localhost:8081/health
```

Expected response:
```json
{
  "status": "ok",
  "message": "Email service is running"
}
```

### 2. Send Confirmation Email
```bash
curl -X POST http://localhost:8081/send/confirmation \
  -H "Content-Type: application/json" \
  -d '{
    "to": "user@example.com",
    "subject": "Email Confirmation",
    "name": "John Doe",
    "confirmationLink": "https://example.com/confirm?token=abc123"
  }'
```

### 3. Send Weather Update Email
```bash
curl -X POST http://localhost:8081/send/weather-update \
  -H "Content-Type: application/json" \
  -d '{
    "to": "user@example.com",
    "subject": "Weather Update",
    "name": "Jane Smith",
    "location": "Kyiv",
    "temperature": 25,
    "description": "Sunny",
    "humidity": 60,
    "windSpeed": 10
  }'
```

## Using HTTP File
You can use the `email-service.http` file with VS Code REST Client extension or similar tools to test the endpoints.

## Troubleshooting

1. Check if container is running:
```bash
docker ps
```

2. Check container logs:
```bash
docker logs email-service
```

3. If container exits immediately, check for environment variable errors in logs.

## Environment Variables
Make sure to set proper SMTP credentials:
- `SMTP_HOST`: SMTP server host (default: smtp.gmail.com)
- `SMTP_PORT`: SMTP server port (default: 587)
- `SMTP_USER`: Your email address
- `SMTP_PASS`: Your email password or app password
- `SERVER_PORT`: Service port (default: 8081)
- `BASE_URL`: Service base URL (default: http://localhost:8081) 