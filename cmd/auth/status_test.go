package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/zalando/go-keyring"
)

func init() {
	keyring.MockInit()
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatal(err)
	}
}

func TestAuthStatus_NoKeys(t *testing.T) {
	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		Format:    output.FormatJSON,
	}

	err := runAuthStatus(t.Context(), opts, false)
	if err == nil {
		t.Fatal("expected error for no keys")
	}
	want := `no keys configured for profile "default" (run honeycomb auth login)`
	if err.Error() != want {
		t.Errorf("got error %q, want %q", err.Error(), want)
	}
}

func TestAuthStatus_Offline(t *testing.T) {
	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		Format:    output.FormatJSON,
	}

	if err := config.SetKey("default", config.KeyConfig, "test-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })

	if err := runAuthStatus(t.Context(), opts, true); err != nil {
		t.Fatal(err)
	}

	var statuses []KeyStatus
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &statuses); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("got %d statuses, want 1", len(statuses))
	}
	if statuses[0].Type != "config" {
		t.Errorf("type = %q, want %q", statuses[0].Type, "config")
	}
	if statuses[0].Status != "stored" {
		t.Errorf("status = %q, want %q", statuses[0].Status, "stored")
	}
}

func TestAuthStatus_Verify(t *testing.T) {
	tests := []struct {
		name    string
		keyType config.KeyType
		key     string
		handler http.Handler
		want    KeyStatus
	}{
		{
			name:    "config key valid",
			keyType: config.KeyConfig,
			key:     "test-config-key",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/1/auth" {
					http.NotFound(w, r)
					return
				}
				if r.Header.Get("X-Honeycomb-Team") != "test-config-key" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":             "abc123",
					"type":           "configuration",
					"team":           map[string]string{"name": "My Team", "slug": "my-team"},
					"environment":    map[string]string{"name": "production", "slug": "production"},
					"api_key_access": map[string]bool{"events": true},
				})
			}),
			want: KeyStatus{
				Type:        "config",
				Status:      "valid",
				Team:        "My Team",
				Environment: "production",
				KeyID:       "abc123",
			},
		},
		{
			name:    "ingest key unauthorized",
			keyType: config.KeyIngest,
			key:     "bad-key",
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			}),
			want: KeyStatus{
				Type:   "ingest",
				Status: "invalid",
			},
		},
		{
			name:    "management key valid",
			keyType: config.KeyManagement,
			key:     "mgmt-key",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/2/auth" {
					http.NotFound(w, r)
					return
				}
				if r.Header.Get("Authorization") != "Bearer mgmt-key" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.Header().Set("Content-Type", "application/vnd.api+json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"id":   "mgmt-id",
						"type": "api-keys",
						"attributes": map[string]any{
							"name":     "My Management Key",
							"key_type": "management",
						},
					},
				})
			}),
			want: KeyStatus{
				Type:   "management",
				Status: "valid",
				KeyID:  "mgmt-id",
				Name:   "My Management Key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			t.Cleanup(srv.Close)

			ts := iostreams.Test(t)
			opts := &options.RootOptions{
				IOStreams: ts.IOStreams,
				Config:    &config.Config{},
				APIUrl:    srv.URL,
				Format:    output.FormatJSON,
			}

			if err := config.SetKey("default", tt.keyType, tt.key); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = config.DeleteKey("default", tt.keyType) })

			if err := runAuthStatus(t.Context(), opts, false); err != nil {
				t.Fatal(err)
			}

			var statuses []KeyStatus
			if err := json.Unmarshal(ts.OutBuf.Bytes(), &statuses); err != nil {
				t.Fatalf("unmarshal output: %v", err)
			}
			if len(statuses) != 1 {
				t.Fatalf("got %d statuses, want 1", len(statuses))
			}
			if got := statuses[0]; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAuthStatus_MultipleKeys(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/1/auth":
			w.Header().Set("Content-Type", "application/json")
			writeJSON(t, w, map[string]any{
				"id":             "cfg-id",
				"type":           "configuration",
				"team":           map[string]string{"name": "Team"},
				"environment":    map[string]string{},
				"api_key_access": map[string]bool{},
			})
		case "/2/auth":
			w.Header().Set("Content-Type", "application/vnd.api+json")
			writeJSON(t, w, map[string]any{
				"data": map[string]any{
					"id":   "mgmt-id",
					"type": "api-keys",
					"attributes": map[string]any{
						"name":     "Mgmt Key",
						"key_type": "management",
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    srv.URL,
		Format:    output.FormatJSON,
	}

	if err := config.SetKey("default", config.KeyConfig, "cfg-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })

	if err := config.SetKey("default", config.KeyManagement, "mgmt-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyManagement) })

	if err := runAuthStatus(t.Context(), opts, false); err != nil {
		t.Fatal(err)
	}

	var statuses []KeyStatus
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &statuses); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(statuses) != 2 {
		t.Fatalf("got %d statuses, want 2", len(statuses))
	}
	if statuses[0].Type != "config" {
		t.Errorf("statuses[0].type = %q, want %q", statuses[0].Type, "config")
	}
	if statuses[1].Type != "management" {
		t.Errorf("statuses[1].type = %q, want %q", statuses[1].Type, "management")
	}
}
