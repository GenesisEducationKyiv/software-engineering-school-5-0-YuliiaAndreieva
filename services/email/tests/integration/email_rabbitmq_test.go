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

const (
	emailServiceHealthTimeout  = 10 * time.Second
	emailServiceHealthInterval = 500 * time.Millisecond
)

func TestEmailConsumerRabbitMQFlow(t *testing.T) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	require.NoError(t, err)
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Logf("Failed to close connection: %v", closeErr)
		}
	}()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer func() {
		if closeErr := ch.Close(); closeErr != nil {
			t.Logf("Failed to close channel: %v", closeErr)
		}
	}()

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

	require.Eventually(t, func() bool {
		resp, err := http.Get("http://email-service:8081/health")
		if err != nil {
			return false
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				t.Logf("Failed to close response body: %v", closeErr)
			}
		}()
		return resp.StatusCode == http.StatusOK
	}, emailServiceHealthTimeout, emailServiceHealthInterval, "email-service did not become healthy in time")

	t.Log("✅ Email service is running and processing RabbitMQ messages")
}

func TestEmailServiceHealth(t *testing.T) {
	require.Eventually(t, func() bool {
		resp, err := http.Get("http://email-service:8081/health")
		if err != nil {
			return false
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				t.Logf("Failed to close response body: %v", closeErr)
			}
		}()
		return resp.StatusCode == http.StatusOK
	}, emailServiceHealthTimeout, emailServiceHealthInterval, "email-service did not become healthy in time")

	resp, err := http.Get("http://email-service:8081/health")
	require.NoError(t, err)
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&healthResponse)
	require.NoError(t, err)

	assert.Equal(t, "ok", healthResponse["status"])
	assert.Equal(t, "email", healthResponse["service"])

	t.Log("✅ Email service health check passed")
}
