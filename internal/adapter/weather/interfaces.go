package weather

import (
	"net/http"
)

type ProviderLogger interface {
	Log(providerName string, responseBody []byte)
}

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}
