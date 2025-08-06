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
	"weather-api/internal/core/ports/out"
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
		t.Logf("Unable to clear MailHog messages: %v", err)
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

	testCases := []struct {
		name          string
		setup         func(t *testing.T) ([]string, string) // returns emails, city
		expectedCount int
		validateEmail func(t *testing.T, message mailHogMessage, email, city string)
		timeout       time.Duration
	}{
		{
			name: "Send Confirmation Email",
			setup: func(t *testing.T) ([]string, string) {
				email := "confirmation@example.com"
				city := "Rivne"

				token, err := ts.SubscribeUseCase.Subscribe(context.Background(), out.SubscribeOptions{
					Email:     email,
					City:      city,
					Frequency: domain.FrequencyDaily,
				})
				require.NoError(t, err)
				assert.NotEmpty(t, token)

				return []string{email}, city
			},
			expectedCount: 1,
			validateEmail: func(t *testing.T, message mailHogMessage, email, city string) {
				assert.Equal(t, "test@example.com", fmt.Sprintf("%s@%s", message.From.Mailbox, message.From.Domain))
				assert.Equal(t, email, fmt.Sprintf("%s@%s", message.To[0].Mailbox, message.To[0].Domain))
				assert.Contains(t, message.Content.Body, "Please click the link below to confirm your subscription")
				assert.Contains(t, message.Content.Body, city)
			},
			timeout: 5 * time.Second,
		},
		{
			name: "Send Weather Update Email",
			setup: func(t *testing.T) ([]string, string) {
				email := "update@example.com"
				city := "Ternopil"

				token, err := ts.SubscribeUseCase.Subscribe(context.Background(), out.SubscribeOptions{
					Email:     email,
					City:      city,
					Frequency: domain.FrequencyDaily,
				})
				require.NoError(t, err)
				err = ts.ConfirmUseCase.ConfirmSubscription(context.Background(), token)
				require.NoError(t, err)

				clearMailHogMessages(t)

				updates, err := ts.WeatherUpdateService.PrepareUpdates(context.Background(), domain.FrequencyDaily)
				require.NoError(t, err)
				require.Len(t, updates, 1)

				err = ts.EmailService.SendUpdates(updates)
				require.NoError(t, err)

				return []string{email}, city
			},
			expectedCount: 1,
			validateEmail: func(t *testing.T, message mailHogMessage, email, city string) {
				assert.Equal(t, "test@example.com", fmt.Sprintf("%s@%s", message.From.Mailbox, message.From.Domain))
				assert.Equal(t, email, fmt.Sprintf("%s@%s", message.To[0].Mailbox, message.To[0].Domain))
				assert.Contains(t, message.Content.Body, city)
				assert.Contains(t, message.Content.Body, "Unsubscribe")
			},
			timeout: 5 * time.Second,
		},
		{
			name: "Multiple Emails in Sequence",
			setup: func(t *testing.T) ([]string, string) {
				emails := []string{"multi1@example.com", "multi2@example.com", "multi3@example.com"}
				city := "Ivano-Frankivsk"

				clearMailHogMessages(t)

				for _, email := range emails {
					token, err := ts.SubscribeUseCase.Subscribe(context.Background(), out.SubscribeOptions{
						Email:     email,
						City:      city,
						Frequency: domain.FrequencyDaily,
					})
					require.NoError(t, err)
					assert.NotEmpty(t, token)
				}

				return emails, city
			},
			expectedCount: 3,
			validateEmail: func(t *testing.T, message mailHogMessage, email, city string) {
				assert.Contains(t, message.Content.Body, "Please click the link below to confirm your subscription")
			},
			timeout: 10 * time.Second,
		},
		{
			name: "Email Headers Validation",
			setup: func(t *testing.T) ([]string, string) {
				email := "headers@example.com"
				city := "Lutsk"

				clearMailHogMessages(t)

				_, err := ts.SubscribeUseCase.Subscribe(context.Background(), out.SubscribeOptions{
					Email:     email,
					City:      city,
					Frequency: domain.FrequencyDaily,
				})
				require.NoError(t, err)

				return []string{email}, city
			},
			expectedCount: 1,
			validateEmail: func(t *testing.T, message mailHogMessage, email, city string) {
				headers := message.Content.Headers
				assert.NotEmpty(t, headers["Subject"])
				assert.NotEmpty(t, headers["From"])
				assert.NotEmpty(t, headers["To"])
				assert.Contains(t, headers["Content-Type"][0], "text/html")
			},
			timeout: 5 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			emails, city := tc.setup(t)

			var messages []mailHogMessage
			require.Eventually(t, func() bool {
				messages = getMailHogMessages(t)
				return len(messages) == tc.expectedCount
			}, tc.timeout, 100*time.Millisecond, fmt.Sprintf("Expected %d email message(s) to arrive", tc.expectedCount))

			for i, message := range messages {
				if i < len(emails) {
					tc.validateEmail(t, message, emails[i], city)
				}
			}

			if tc.name == "Multiple Emails in Sequence" {
				sentEmails := make(map[string]bool)
				for _, message := range messages {
					recipient := fmt.Sprintf("%s@%s", message.To[0].Mailbox, message.To[0].Domain)
					sentEmails[recipient] = true
				}

				for _, email := range emails {
					assert.True(t, sentEmails[email], "Email was not sent to %s", email)
				}
			}
		})
	}
}
