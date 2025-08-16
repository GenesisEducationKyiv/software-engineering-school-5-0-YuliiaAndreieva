# RabbitMQ Usage in Weather Service Project

## Overview

This project uses RabbitMQ as a message broker for asynchronous communication between services. The main use case is the subscription confirmation flow where the Subscription Service publishes events that the Email Service consumes.

## Architecture

```
Subscription Service → RabbitMQ → Email Service
```

### Flow Description

1. **Subscription Service** receives a subscription request
2. **Subscription Service** publishes a `SubscriptionCreatedEvent` to RabbitMQ
3. **Email Service** consumes the event and sends a confirmation email
4. This creates an asynchronous, decoupled flow between services

## Configuration

### RabbitMQ Setup

RabbitMQ is configured in `docker-compose.yml`:

```yaml
rabbitmq:
  image: rabbitmq:3-management
  container_name: rabbitmq
  ports:
    - "5672:5672"      # AMQP protocol
    - "15672:15672"    # Management UI
  environment:
    - RABBITMQ_DEFAULT_USER=admin
    - RABBITMQ_DEFAULT_PASS=password
```

### Service Configuration

Each service that uses RabbitMQ has configuration in their `.env` files:

```env
# RabbitMQ Configuration
RABBITMQ_URL=amqp://admin:password@rabbitmq:5672/
RABBITMQ_EXCHANGE=subscription_events
RABBITMQ_QUEUE=email_notifications
```

## Implementation

### Publisher (Subscription Service)

Located in `services/subscription/internal/adapter/messaging/rabbitmq_publisher.go`:

```go
type RabbitMQPublisher struct {
    conn    *amqp.Connection
    channel *amqp.Channel
    logger  out.Logger
}

func (p *RabbitMQPublisher) PublishSubscriptionCreated(ctx context.Context, subscription domain.Subscription) error {
    // Publishes SubscriptionCreatedEvent to RabbitMQ
}
```

### Consumer (Email Service)

Located in `services/email/internal/adapter/messaging/rabbitmq_consumer.go`:

```go
type RabbitMQConsumer struct {
    conn    *amqp.Connection
    channel *amqp.Channel
    logger  out.Logger
    emailService in.SendEmailUseCase
}

func (c *RabbitMQConsumer) Start() {
    // Consumes messages and sends confirmation emails
}
```

## Event Structure

```go
type SubscriptionCreatedEvent struct {
    Email        string           `json:"email"`
    City         string           `json:"city"`
    Frequency    domain.Frequency `json:"frequency"`
    Token        string           `json:"token"`
    IsConfirmed  bool             `json:"is_confirmed"`
}
```

## Testing

### Integration Tests

- `services/subscription/tests/integration/subscription_rabbitmq_test.go` - Tests publishing events
- `services/email/tests/integration/email_rabbitmq_test.go` - Tests consuming events

### Running Tests

```bash
# Start RabbitMQ
docker-compose up -d rabbitmq

# Run integration tests
cd services/subscription && go test ./tests/integration/... -v
cd services/email && go test ./tests/integration/... -v
```

## Benefits

1. **Decoupling**: Services don't need to know about each other directly
2. **Reliability**: Messages are persisted and can be retried
3. **Scalability**: Multiple consumers can process messages in parallel
4. **Fault Tolerance**: If Email Service is down, messages queue up

## Monitoring

Access RabbitMQ Management UI at `http://localhost:15672`:
- Username: `admin`
- Password: `password`

Monitor queues, exchanges, and message flow in real-time. 