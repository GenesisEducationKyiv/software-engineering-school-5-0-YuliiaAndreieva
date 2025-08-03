package helpers

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type MailHogMessage struct {
	ID      string           `json:"ID"`
	From    MailHogAddress   `json:"From"`
	To      []MailHogAddress `json:"To"`
	Content MailHogContent   `json:"Content"`
	Created string           `json:"Created"`
}

type MailHogAddress struct {
	Mailbox string `json:"Mailbox"`
	Domain  string `json:"Domain"`
	Params  string `json:"Params"`
}

type MailHogContent struct {
	Headers map[string][]string `json:"Headers"`
	Body    string              `json:"Body"`
	Size    int                 `json:"Size"`
}

type MailHogResponse struct {
	Total int              `json:"Total"`
	Count int              `json:"Count"`
	Start int              `json:"Start"`
	Items []MailHogMessage `json:"Items"`
}

type MailHogClient struct {
	baseURL string
}

func NewMailHogClient(baseURL string) *MailHogClient {
	return &MailHogClient{
		baseURL: baseURL,
	}
}

func (m *MailHogClient) CheckEmailSent(t *testing.T, recipientEmail, expectedSubject string) bool {
	time.Sleep(2 * time.Second)

	resp, err := http.Get(m.baseURL + "/api/v2/messages")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var mailHogResp MailHogResponse
	err = json.Unmarshal(body, &mailHogResp)
	require.NoError(t, err)

	if len(mailHogResp.Items) == 0 {
		return false
	}

	for _, message := range mailHogResp.Items {
		recipientFound := false
		for _, to := range message.To {
			if to.Mailbox+"@"+to.Domain == recipientEmail {
				recipientFound = true
				break
			}
		}

		if !recipientFound {
			continue
		}

		if subjects, ok := message.Content.Headers["Subject"]; ok {
			for _, subject := range subjects {
				if subject == expectedSubject {
					return true
				}
			}
		}
	}

	return false
}

func (m *MailHogClient) ClearMailHog(t *testing.T) {
	req, err := http.NewRequest("DELETE", m.baseURL+"/api/v1/messages", nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	time.Sleep(1 * time.Second)
}

func (m *MailHogClient) GetEmailContent(t *testing.T, recipientEmail string) *MailHogMessage {
	time.Sleep(3 * time.Second)

	resp, err := http.Get(m.baseURL + "/api/v2/messages")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var mailHogResp MailHogResponse
	err = json.Unmarshal(body, &mailHogResp)
	require.NoError(t, err)

	for _, message := range mailHogResp.Items {
		for _, to := range message.To {
			if to.Mailbox+"@"+to.Domain == recipientEmail {
				return &message
			}
		}
	}

	return nil
}

func (m *MailHogClient) GetEmailsCount(t *testing.T) int {
	resp, err := http.Get(m.baseURL + "/api/v2/messages")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var mailHogResp MailHogResponse
	err = json.Unmarshal(body, &mailHogResp)
	require.NoError(t, err)

	return mailHogResp.Total
}
