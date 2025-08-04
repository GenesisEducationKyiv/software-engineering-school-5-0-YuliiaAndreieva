package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"subscription/internal/core/domain"
	"subscription/internal/core/ports/out"

	"github.com/streadway/amqp"
)

type RabbitMQPublisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
	logger   out.Logger
}

func NewRabbitMQPublisher(url, exchange string, logger out.Logger) (*RabbitMQPublisher, error) {
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

	return &RabbitMQPublisher{
		conn:     conn,
		channel:  ch,
		exchange: exchange,
		logger:   logger,
	}, nil
}

func (p *RabbitMQPublisher) PublishSubscriptionCreated(ctx context.Context, subscription domain.Subscription) error {
	event := domain.SubscriptionCreatedEvent{
		Email:       subscription.Email,
		City:        subscription.City,
		Frequency:   subscription.Frequency,
		Token:       subscription.Token,
		IsConfirmed: subscription.IsConfirmed,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.channel.Publish(
		p.exchange, // exchange
		"",         // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Infof("Published subscription created event for email: %s", subscription.Email)
	return nil
}

func (p *RabbitMQPublisher) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			p.logger.Errorf("Failed to close channel: %v", err)
		}
	}
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			p.logger.Errorf("Failed to close connection: %v", err)
		}
	}
	return nil
}
