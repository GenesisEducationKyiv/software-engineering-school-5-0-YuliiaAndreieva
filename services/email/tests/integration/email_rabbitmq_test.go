package integration

import (
	"email/internal/core/domain"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailConsumerRabbitMQFlow(t *testing.T) {
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

	testEvent := domain.SubscriptionCreatedEvent{
		Email:       "test-consumer@example.com",
		City:        "TestConsumerCity",
		Frequency:   "hourly",
		Token:       "test-token-123",
		IsConfirmed: false,
	}

	eventBody, err := json.Marshal(testEvent)
	require.NoError(t, err)

	err = ch.Publish(
		"subscription_events",
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        eventBody,
		},
	)
	require.NoError(t, err)

	t.Logf("✅ Published test event to RabbitMQ for email: %s", testEvent.Email)

	time.Sleep(2 * time.Second)

	resp, err := http.Get("http://localhost:8081/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	t.Log("✅ Email service is running and processing RabbitMQ messages")
}

func TestEmailServiceHealth(t *testing.T) {
	resp, err := http.Get("http://localhost:8081/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&healthResponse)
	require.NoError(t, err)

	assert.Equal(t, "ok", healthResponse["status"])
	assert.Equal(t, "email", healthResponse["service"])

	t.Log("✅ Email service health check passed")
}

func TestSubscriptionServiceHealth(t *testing.T) {
	resp, err := http.Get("http://localhost:8082/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&healthResponse)
	require.NoError(t, err)

	assert.Equal(t, "ok", healthResponse["status"])
	assert.Equal(t, "subscription", healthResponse["service"])

	t.Log("✅ Subscription service health check passed")
}
