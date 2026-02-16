package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
		var jsonapi struct {
			Errors []struct {
				Title  string `json:"title"`
				Detail string `json:"detail"`
			} `json:"errors"`
		}
		if json.Unmarshal(body, &jsonapi) == nil && len(jsonapi.Errors) > 0 {
			messages := make([]string, 0, len(jsonapi.Errors))
			for _, e := range jsonapi.Errors {
				if e.Detail != "" {
					messages = append(messages, e.Detail)
				} else if e.Title != "" {
					messages = append(messages, e.Title)
				}
			}
			apiErr.Message = strings.Join(messages, ", ")
		}
	}

	if apiErr.Message == "" {
		apiErr.Message = http.StatusText(statusCode)
	}

	return apiErr
}
