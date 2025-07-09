package weather

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"weather-api/internal/util/jsonutil"
)

func DecodeResponse[T any](resp *http.Response) (*T, []byte, error) {
	var logBuffer bytes.Buffer
	teeReader := io.TeeReader(resp.Body, &logBuffer)

	result, err := jsonutil.Decode[T](teeReader)
	if err != nil {
		msg := "unable to decode response: " + err.Error()
		log.Print(msg)
		return nil, nil, NewProviderError("OpenWeatherMap", 500, msg)
	}

	return &result, logBuffer.Bytes(), nil
}
