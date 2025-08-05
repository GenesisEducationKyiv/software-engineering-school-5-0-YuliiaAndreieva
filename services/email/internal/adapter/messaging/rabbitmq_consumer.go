package messaging

import (
	"context"
	"email/internal/core/domain"
	"email/internal/core/ports/in"
	"email/internal/core/ports/out"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
)

type RabbitMQConsumer struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
	queue    string
	useCase  in.SendEmailUseCase
	logger   out.Logger
	baseURL  string
}

func NewRabbitMQConsumer(url, exchange, queue string, useCase in.SendEmailUseCase, logger out.Logger, baseURL string) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ at %s: %w", url, err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		exchange, // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange '%s': %w", exchange, err)
	}

	q, err := ch.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue '%s': %w", queue, err)
	}

	err = ch.QueueBind(
		q.Name,   // queue name
		"",       // routing key
		exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue '%s' to exchange '%s': %w", queue, exchange, err)
	}

	logger.Infof("Successfully connected to RabbitMQ: exchange=%s, queue=%s", exchange, queue)

	return &RabbitMQConsumer{
		conn:     conn,
		channel:  ch,
		exchange: exchange,
		queue:    queue,
		useCase:  useCase,
		logger:   logger,
		baseURL:  baseURL,
	}, nil
}

func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context is required")
	}

	msgs, err := c.channel.Consume(
		c.queue, // queue
		"",      // consumer
		false,   // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming from queue '%s': %w", c.queue, err)
	}

	c.logger.Infof("Started consuming messages from queue: %s", c.queue)

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.logger.Infof("Stopping consumer due to context cancellation: %v", ctx.Err())
				return
			case msg, ok := <-msgs:
				if !ok {
					c.logger.Errorf("Message channel closed unexpectedly")
					return
				}

				if err := c.handleMessage(ctx, msg); err != nil {
					c.logger.Errorf("Failed to handle message (ID: %s): %v", msg.MessageId, err)
					if err := msg.Nack(false, true); err != nil {
						c.logger.Errorf("Failed to nack message: %v", err)
					}
				} else {
					if err := msg.Ack(false); err != nil {
						c.logger.Errorf("Failed to ack message: %v", err)
					}
				}
			}
		}
	}()

	return nil
}

func (c *RabbitMQConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) error {
	if msg.Body == nil || len(msg.Body) == 0 {
		return fmt.Errorf("received empty message")
	}

	var event domain.SubscriptionCreatedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal message body: %w", err)
	}

	c.logger.Infof("Received subscription created event for email: %s, city: %s", event.Email, event.City)

	req := domain.ConfirmationEmailRequest{
		To:               event.Email,
		Subject:          "Confirm your weather subscription",
		City:             event.City,
		ConfirmationLink: fmt.Sprintf("%s/confirm/%s", c.baseURL, event.Token),
	}

	result, err := c.useCase.SendConfirmationEmail(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send confirmation email to %s: %w", event.Email, err)
	}

	if result.Status == domain.StatusFailed {
		return fmt.Errorf("email delivery failed for %s: %s", event.Email, result.Error)
	}

	c.logger.Infof("Successfully sent confirmation email to: %s (EmailID: %s)", event.Email, result.EmailID)
	return nil
}

func (c *RabbitMQConsumer) Close() error {
	var errors []error

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			c.logger.Errorf("Failed to close channel: %v", err)
			errors = append(errors, fmt.Errorf("failed to close channel: %w", err))
		} else {
			c.logger.Infof("Successfully closed RabbitMQ channel")
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			c.logger.Errorf("Failed to close connection: %v", err)
			errors = append(errors, fmt.Errorf("failed to close connection: %w", err))
		} else {
			c.logger.Infof("Successfully closed RabbitMQ connection")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors during close: %v", len(errors), errors)
	}

	return nil
}
