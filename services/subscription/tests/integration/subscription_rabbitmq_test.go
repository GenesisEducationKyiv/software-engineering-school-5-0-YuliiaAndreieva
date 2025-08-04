package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"subscription/internal/core/domain"
	"testing"
	"time"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionRabbitMQFlow(t *testing.T) {
	conn, err := amqp.Dial("amqp://admin:password@localhost:5672/")
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"subscription_events",
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
		"subscription_events",
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
		"http://localhost:8082/subscribe",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	select {
	case <-messageReceived:
		assert.Equal(t, subscriptionReq.Email, receivedEvent.Email)
		assert.Equal(t, subscriptionReq.City, receivedEvent.City)
		assert.Equal(t, subscriptionReq.Frequency, string(receivedEvent.Frequency))
		assert.False(t, receivedEvent.IsConfirmed)
		assert.NotEmpty(t, receivedEvent.Token)

		t.Logf("✅ Successfully received RabbitMQ event for email: %s", receivedEvent.Email)

	case <-time.After(10 * time.Second):
		t.Fatal("❌ Timeout waiting for RabbitMQ message")
	}
}

func TestRabbitMQConnection(t *testing.T) {
	conn, err := amqp.Dial("amqp://admin:password@localhost:5672/")
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"test_exchange",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	require.NoError(t, err)

	err = ch.ExchangeDelete(
		"test_exchange",
		false,
		false,
	)
	require.NoError(t, err)

	t.Log("✅ RabbitMQ connection test passed")
}
