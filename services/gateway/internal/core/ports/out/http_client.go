package out

import (
	"net/http"
)

//go:generate mockery --name HTTPClient
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
