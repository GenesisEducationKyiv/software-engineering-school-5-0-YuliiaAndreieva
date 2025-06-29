//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"testing"
	"time"
	"weather-api/internal/core/domain"
)

type mailHogMessage struct {
	ID   string `json:"ID"`
	From struct {
		Mailbox string `json:"Mailbox"`
		Domain  string `json:"Domain"`
	} `json:"From"`
	To []struct {
		Mailbox string `json:"Mailbox"`
		Domain  string `json:"Domain"`
	} `json:"To"`
	Content struct {
		Headers map[string][]string `json:"Headers"`
		Body    string              `json:"Body"`
	} `json:"Content"`
}

func setupEmailTestServices(t *testing.T) *TestServices {
	testConfig := GetTestConfig()
	os.Setenv("DB_CONN_STR", testConfig.DBConnStr)
	os.Setenv("WEATHER_API_KEY", testConfig.WeatherAPIKey)
	os.Setenv("SMTP_HOST", testConfig.SMTPHost)
	os.Setenv("SMTP_PORT", testConfig.SMTPPort)
	os.Setenv("SMTP_USER", testConfig.SMTPUser)
	os.Setenv("SMTP_PASS", testConfig.SMTPPass)
	os.Setenv("PORT", testConfig.Port)
	os.Setenv("BASE_URL", testConfig.BaseURL)

	return SetupTestServices(t)
}

func checkMailHogAvailable(t *testing.T) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:8025/api/v2/messages")
	if err != nil {
		t.Logf("MailHog is not available: %v", err)
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func getMailHogMessages(t *testing.T) []mailHogMessage {
	if !checkMailHogAvailable(t) {
		return []mailHogMessage{}
	}

	resp, err := http.Get("http://localhost:8025/api/v2/messages")
	require.NoError(t, err)
	defer resp.Body.Close()

	var response struct {
		Items []mailHogMessage `json:"items"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	return response.Items
}

func clearMailHogMessages(t *testing.T) {
	if !checkMailHogAvailable(t) {
		return
	}

	req, err := http.NewRequest("DELETE", "http://localhost:8025/api/v1/messages", nil)
	require.NoError(t, err)

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Failed to clear MailHog messages: %v", err)
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestEmailIntegration_WithMailHog(t *testing.T) {
	if !checkMailHogAvailable(t) {
		t.Skip("MailHog is not available, skipping email integration tests")
	}

	ts := setupEmailTestServices(t)
	defer ts.Cleanup()

	clearMailHogMessages(t)

	t.Run("Send Confirmation Email", func(t *testing.T) {
		ctx := context.Background()
		email := "confirmation@example.com"
		city := "Rivne"

		token, err := ts.SubscriptionService.Subscribe(ctx, email, city, domain.FrequencyDaily)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		time.Sleep(2 * time.Second)

		messages := getMailHogMessages(t)
		require.Len(t, messages, 1)

		message := messages[0]
		assert.Equal(t, "test@example.com", fmt.Sprintf("%s@%s", message.From.Mailbox, message.From.Domain))
		assert.Equal(t, email, fmt.Sprintf("%s@%s", message.To[0].Mailbox, message.To[0].Domain))

		assert.Contains(t, message.Content.Body, "Please click the link below to confirm your subscription")
		assert.Contains(t, message.Content.Body, city)
	})

	t.Run("Send Weather Update Email", func(t *testing.T) {
		ctx := context.Background()
		email := "update@example.com"
		city := "Ternopil"

		token, err := ts.SubscriptionService.Subscribe(ctx, email, city, domain.FrequencyDaily)
		require.NoError(t, err)
		err = ts.SubscriptionService.Confirm(ctx, token)
		require.NoError(t, err)

		clearMailHogMessages(t)

		updates, err := ts.WeatherUpdateService.PrepareUpdates(ctx, domain.FrequencyDaily)
		require.NoError(t, err)
		require.Len(t, updates, 1)

		err = ts.EmailService.SendUpdates(updates)
		require.NoError(t, err)

		time.Sleep(2 * time.Second)

		messages := getMailHogMessages(t)
		require.Len(t, messages, 1)

		message := messages[0]
		assert.Equal(t, "test@example.com", fmt.Sprintf("%s@%s", message.From.Mailbox, message.From.Domain))
		assert.Equal(t, email, fmt.Sprintf("%s@%s", message.To[0].Mailbox, message.To[0].Domain))

		assert.Contains(t, message.Content.Body, city)
		assert.Contains(t, message.Content.Body, "Unsubscribe")
	})

	t.Run("Multiple Emails in Sequence", func(t *testing.T) {
		ctx := context.Background()
		emails := []string{"multi1@example.com", "multi2@example.com", "multi3@example.com"}
		city := "Ivano-Frankivsk"

		clearMailHogMessages(t)

		for _, email := range emails {
			token, err := ts.SubscriptionService.Subscribe(ctx, email, city, domain.FrequencyDaily)
			require.NoError(t, err)
			assert.NotEmpty(t, token)
		}

		time.Sleep(3 * time.Second)

		messages := getMailHogMessages(t)
		assert.Len(t, messages, 3)

		sentEmails := make(map[string]bool)
		for _, message := range messages {
			recipient := fmt.Sprintf("%s@%s", message.To[0].Mailbox, message.To[0].Domain)
			sentEmails[recipient] = true
			assert.Contains(t, message.Content.Body, "Please click the link below to confirm your subscription")
		}

		for _, email := range emails {
			assert.True(t, sentEmails[email], "Email was not sent to %s", email)
		}
	})

	t.Run("Email Headers Validation", func(t *testing.T) {
		ctx := context.Background()
		email := "headers@example.com"
		city := "Lutsk"

		clearMailHogMessages(t)

		_, err := ts.SubscriptionService.Subscribe(ctx, email, city, domain.FrequencyDaily)
		require.NoError(t, err)

		time.Sleep(2 * time.Second)

		messages := getMailHogMessages(t)
		require.Len(t, messages, 1)

		message := messages[0]
		headers := message.Content.Headers

		assert.NotEmpty(t, headers["Subject"])
		assert.NotEmpty(t, headers["From"])
		assert.NotEmpty(t, headers["To"])
		assert.Contains(t, headers["Content-Type"][0], "text/html")
	})
}
