package marker

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/zalando/go-keyring"
)

func init() {
	keyring.MockInit()
}

func setupTest(t *testing.T, handler http.Handler) (*options.RootOptions, *iostreams.TestStreams) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    srv.URL,
		Format:    "json",
	}

	if err := config.SetKey("default", config.KeyConfig, "test-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })

	return opts, ts
}

func TestList(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/markers/test-dataset" {
			t.Errorf("path = %q, want /1/markers/test-dataset", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":         "abc123",
				"type":       "deploy",
				"message":    "v1.0.0",
				"start_time": 1700000000,
				"end_time":   1700000300,
				"color":      "#ff0000",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z",
			},
			{
				"id":         "def456",
				"type":       "incident",
				"message":    "outage",
				"start_time": 1700001000,
				"created_at": "2024-01-02T00:00:00Z",
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list", "--dataset", "test-dataset"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []markerItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != "abc123" {
		t.Errorf("items[0].ID = %q, want %q", items[0].ID, "abc123")
	}
	if items[0].Type != "deploy" {
		t.Errorf("items[0].Type = %q, want %q", items[0].Type, "deploy")
	}
	if items[0].StartTime == nil || *items[0].StartTime != 1700000000 {
		t.Errorf("items[0].StartTime = %v, want 1700000000", items[0].StartTime)
	}
	if items[1].EndTime != nil {
		t.Errorf("items[1].EndTime = %v, want nil", items[1].EndTime)
	}
}

func TestList_Empty(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list", "--dataset", "test-dataset"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []markerItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

func TestList_NoKey(t *testing.T) {
	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    "http://localhost",
		Format:    "json",
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list", "--dataset", "test-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "no config key configured") {
		t.Errorf("error = %q, want missing key message", err.Error())
	}
}

func TestList_Unauthorized(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unknown API key - check your credentials"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list", "--dataset", "test-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if !strings.Contains(err.Error(), "HTTP 401") {
		t.Errorf("error = %q, want HTTP 401", err.Error())
	}
	if !strings.Contains(err.Error(), "unknown API key") {
		t.Errorf("error = %q, want error message from body", err.Error())
	}
}
