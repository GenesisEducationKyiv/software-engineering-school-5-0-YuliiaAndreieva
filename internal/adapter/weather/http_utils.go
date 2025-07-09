package weather

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

func ExecuteRequest(httpClient HTTPDoer, req *http.Request) (*http.Response, error) {
	resp, err := httpClient.Do(req)
	if err != nil {
		msg := fmt.Sprintf("make HTTP request: %v", err)
		log.Print(msg)
		return nil, errors.New(msg)
	}
	return resp, nil
}

func CloseResponse(resp *http.Response) {
	if closeErr := resp.Body.Close(); closeErr != nil {
		log.Printf("close response body: %v", closeErr)
	}
}
