package api

import "fmt"

// Decode validates an API response and returns its typed body. It folds the
// status check and the typed-nil guard that every WithResponse call site
// repeats: CheckResponse rejects non-2xx/3xx status codes (parsing the body
// for an error message), and a nil typed pointer means the server returned a
// status the generated client could not decode into the expected type.
//
// The typed body is passed explicitly because the generated *Resp types expose
// it as a concrete field (JSON200, JSON201, ...) that is not reachable through
// a common interface.
func Decode[T any](statusCode int, status string, body []byte, typed *T) (*T, error) {
	if err := CheckResponse(statusCode, body); err != nil {
		return nil, err
	}
	if typed == nil {
		return nil, fmt.Errorf("unexpected response: %s", status)
	}
	return typed, nil
}
