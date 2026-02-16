package api

import (
	"errors"
	"testing"
)

func TestCheckResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantNil    bool
		wantCode   int
		wantMsg    string
	}{
		{
			name:       "200 returns nil",
			statusCode: 200,
			body:       `{"ok":true}`,
			wantNil:    true,
		},
		{
			name:       "400 with error field",
			statusCode: 400,
			body:       `{"error":"bad request"}`,
			wantCode:   400,
			wantMsg:    "bad request",
		},
		{
			name:       "401 with error field",
			statusCode: 401,
			body:       `{"error":"unauthorized"}`,
			wantCode:   401,
			wantMsg:    "unauthorized",
		},
		{
			name:       "429 with detail field",
			statusCode: 429,
			body:       `{"detail":"rate limited"}`,
			wantCode:   429,
			wantMsg:    "rate limited",
		},
		{
			name:       "409 with jsonapi errors",
			statusCode: 409,
			body:       `{"errors":[{"title":"Conflict","detail":"environment is delete protected"}]}`,
			wantCode:   409,
			wantMsg:    "environment is delete protected",
		},
		{
			name:       "409 with jsonapi title only",
			statusCode: 409,
			body:       `{"errors":[{"title":"Conflict"}]}`,
			wantCode:   409,
			wantMsg:    "Conflict",
		},
		{
			name:       "422 with multiple jsonapi errors",
			statusCode: 422,
			body:       `{"errors":[{"detail":"name is required"},{"detail":"slug is invalid"}]}`,
			wantCode:   422,
			wantMsg:    "name is required, slug is invalid",
		},
		{
			name:       "standard error takes priority over jsonapi",
			statusCode: 400,
			body:       `{"error":"bad request","errors":[{"detail":"should not appear"}]}`,
			wantCode:   400,
			wantMsg:    "bad request",
		},
		{
			name:       "500 with empty body",
			statusCode: 500,
			body:       "",
			wantCode:   500,
			wantMsg:    "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckResponse(tt.statusCode, []byte(tt.body))
			if tt.wantNil {
				if err != nil {
					t.Fatalf("got error %v, want nil", err)
				}
				return
			}

			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("got %T, want *APIError", err)
			}
			if apiErr.StatusCode != tt.wantCode {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.wantCode)
			}
			if apiErr.Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", apiErr.Message, tt.wantMsg)
			}
		})
	}
}
