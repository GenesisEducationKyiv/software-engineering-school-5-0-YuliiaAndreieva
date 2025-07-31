package http

import (
	"net/http"
	"time"
	"weather/internal/core/ports/out"
)

type Client struct {
	client *http.Client
}

func NewHTTPClient() out.HTTPDoer {
	return &Client{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (h *Client) Do(req *http.Request) (*http.Response, error) {
	return h.client.Do(req)
}
