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
}

func NewRabbitMQConsumer(url, exchange, queue string, useCase in.SendEmailUseCase, logger out.Logger) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
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
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
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
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.QueueBind(
		q.Name,   // queue name
		"",       // routing key
		exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &RabbitMQConsumer{
		conn:     conn,
		channel:  ch,
		exchange: exchange,
		queue:    queue,
		useCase:  useCase,
		logger:   logger,
	}, nil
}

func (c *RabbitMQConsumer) Start(ctx context.Context) error {
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
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	c.logger.Infof("Started consuming messages from queue: %s", c.queue)

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("Stopping consumer due to context cancellation")
				return
			case msg := <-msgs:
				if err := c.handleMessage(ctx, msg); err != nil {
					c.logger.Errorf("Failed to handle message: %v", err)
					msg.Nack(false, true)
				} else {
					msg.Ack(false)
				}
			}
		}
	}()

	return nil
}

func (c *RabbitMQConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) error {
	var event domain.SubscriptionCreatedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	c.logger.Infof("Received subscription created event for email: %s", event.Email)

	req := domain.ConfirmationEmailRequest{
		To:               event.Email,
		Subject:          "Confirm your weather subscription",
		City:             event.City,
		ConfirmationLink: fmt.Sprintf("%s/confirm/%s", "http://localhost:8082", event.Token),
	}

	if err := c.useCase.SendConfirmationEmail(ctx, req); err != nil {
		return fmt.Errorf("failed to send confirmation email: %w", err)
	}

	c.logger.Infof("Successfully sent confirmation email to: %s", event.Email)
	return nil
}

func (c *RabbitMQConsumer) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			c.logger.Errorf("Failed to close channel: %v", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			c.logger.Errorf("Failed to close connection: %v", err)
		}
	}
	return nil
}
