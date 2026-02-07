package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}

func CheckResponse(statusCode int, body []byte) error {
	if statusCode >= 200 && statusCode < 400 {
		return nil
	}

	apiErr := &APIError{StatusCode: statusCode}

	var parsed struct {
		Error  *string `json:"error"`
		Detail *string `json:"detail"`
	}
	if json.Unmarshal(body, &parsed) == nil {
		switch {
		case parsed.Error != nil:
			apiErr.Message = *parsed.Error
		case parsed.Detail != nil:
			apiErr.Message = *parsed.Detail
		}
	}

	if apiErr.Message == "" {
		apiErr.Message = http.StatusText(statusCode)
	}

	return apiErr
}
