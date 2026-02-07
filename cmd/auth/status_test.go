package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
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
	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		Format:    "json",
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
	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		Format:    "json",
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

func TestAuthStatus_ConfigKey_Valid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/auth" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("X-Honeycomb-Team") != "test-config-key" {
			w.WriteHeader(http.StatusUnauthorized)
			writeJSON(t, w, map[string]string{"error": "unauthorized"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSON(t, w, map[string]any{
			"id":          "abc123",
			"type":        "configuration",
			"team":        map[string]string{"name": "My Team", "slug": "my-team"},
			"environment": map[string]string{"name": "production", "slug": "production"},
			"api_key_access": map[string]bool{
				"events": true,
			},
		})
	}))
	t.Cleanup(srv.Close)

	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    srv.URL,
		Format:    "json",
	}

	if err := config.SetKey("default", config.KeyConfig, "test-config-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })

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
	s := statuses[0]
	if s.Type != "config" {
		t.Errorf("type = %q, want %q", s.Type, "config")
	}
	if s.Status != "valid" {
		t.Errorf("status = %q, want %q", s.Status, "valid")
	}
	if s.Team != "My Team" {
		t.Errorf("team = %q, want %q", s.Team, "My Team")
	}
	if s.Environment != "production" {
		t.Errorf("environment = %q, want %q", s.Environment, "production")
	}
	if s.KeyID != "abc123" {
		t.Errorf("key_id = %q, want %q", s.KeyID, "abc123")
	}
}

func TestAuthStatus_IngestKey_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(t, w, map[string]string{"error": "unauthorized"})
	}))
	t.Cleanup(srv.Close)

	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    srv.URL,
		Format:    "json",
	}

	if err := config.SetKey("default", config.KeyIngest, "bad-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyIngest) })

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
	if statuses[0].Status != "invalid" {
		t.Errorf("status = %q, want %q", statuses[0].Status, "invalid")
	}
}

func TestAuthStatus_ManagementKey_Valid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2/auth" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer mgmt-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		writeJSON(t, w, map[string]any{
			"data": map[string]any{
				"id":   "mgmt-id",
				"type": "api-keys",
				"attributes": map[string]any{
					"name":     "My Management Key",
					"key_type": "management",
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    srv.URL,
		Format:    "json",
	}

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
	if len(statuses) != 1 {
		t.Fatalf("got %d statuses, want 1", len(statuses))
	}
	s := statuses[0]
	if s.Type != "management" {
		t.Errorf("type = %q, want %q", s.Type, "management")
	}
	if s.Status != "valid" {
		t.Errorf("status = %q, want %q", s.Status, "valid")
	}
	if s.KeyID != "mgmt-id" {
		t.Errorf("key_id = %q, want %q", s.KeyID, "mgmt-id")
	}
	if s.Name != "My Management Key" {
		t.Errorf("name = %q, want %q", s.Name, "My Management Key")
	}
}

func TestAuthStatus_MultipleKeys(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/1/auth":
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

	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    srv.URL,
		Format:    "json",
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
