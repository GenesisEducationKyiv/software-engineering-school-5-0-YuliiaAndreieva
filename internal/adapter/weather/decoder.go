package weather

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"weather-api/internal/util/jsonutil"
)

func DecodeResponse[T any](resp *http.Response) (*T, []byte, error) {
	var logBuffer bytes.Buffer
	teeReader := io.TeeReader(resp.Body, &logBuffer)

	result, err := jsonutil.Decode[T](teeReader)
	if err != nil {
		return nil, nil, fmt.Errorf("decoding error: %w", err)
	}

	return &result, logBuffer.Bytes(), nil
}
