package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"subscription/internal/config"
	"subscription/internal/core/domain"
	"testing"
	"time"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionRabbitMQFlow(t *testing.T) {
	cfg, err := config.LoadTestConfig()
	require.NoError(t, err)

	conn, err := amqp.Dial(cfg.RabbitMQ.URL)
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	err = ch.ExchangeDeclare(
		cfg.RabbitMQ.Exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	require.NoError(t, err)

	q, err := ch.QueueDeclare(
		"test_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	require.NoError(t, err)

	err = ch.QueueBind(
		q.Name,
		"",
		cfg.RabbitMQ.Exchange,
		false,
		nil,
	)
	require.NoError(t, err)

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	require.NoError(t, err)

	messageReceived := make(chan bool, 1)
	var receivedEvent domain.SubscriptionCreatedEvent

	go func() {
		for msg := range msgs {
			var event domain.SubscriptionCreatedEvent
			err := json.Unmarshal(msg.Body, &event)
			if err == nil {
				receivedEvent = event
				messageReceived <- true
			}
			msg.Ack(false)
		}
	}()

	subscriptionReq := domain.SubscriptionRequest{
		Email:     "test-rabbitmq@example.com",
		City:      "TestCity",
		Frequency: "daily",
	}

	reqBody, err := json.Marshal(subscriptionReq)
	require.NoError(t, err)

	resp, err := http.Post(
		cfg.Server.BaseURL+"/subscribe",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	select {
	case <-messageReceived:
		assert.Equal(t, "test-rabbitmq@example.com", receivedEvent.Email)
		assert.Equal(t, "TestCity", receivedEvent.City)
		assert.Equal(t, domain.Frequency("daily"), receivedEvent.Frequency)
		t.Log("✅ RabbitMQ message received successfully")
	case <-time.After(10 * time.Second):
		t.Fatal("❌ Timeout waiting for RabbitMQ message")
	}
}

func TestRabbitMQConnection(t *testing.T) {
	cfg, err := config.LoadTestConfig()
	require.NoError(t, err)

	conn, err := amqp.Dial(cfg.RabbitMQ.URL)
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	t.Log("✅ RabbitMQ connection successful")
}
