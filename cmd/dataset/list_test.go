package dataset

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func setupTest(t *testing.T, handler http.Handler) (*options.RootOptions, *iostreams.TestStreams) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    srv.URL,
		Format:    output.FormatJSON,
	}

	if err := config.SetKey("default", config.KeyConfig, "test-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })

	return opts, ts
}

func TestList(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/datasets" {
			t.Errorf("path = %q, want /1/datasets", r.URL.Path)
		}
		if r.Header.Get("X-Honeycomb-Team") != "test-key" {
			t.Errorf("missing auth header")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"name":                  "production",
				"slug":                  "production",
				"description":           "Production events",
				"regular_columns_count": 42,
				"last_written_at":       "2025-01-15T10:30:00Z",
				"created_at":            "2024-06-01T00:00:00Z",
			},
			{
				"name":                  "staging",
				"slug":                  "staging",
				"regular_columns_count": nil,
				"last_written_at":       nil,
				"created_at":            "2024-07-01T00:00:00Z",
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []datasetItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].Name != "production" {
		t.Errorf("items[0].Name = %q, want %q", items[0].Name, "production")
	}
	if items[0].Columns == nil || *items[0].Columns != 42 {
		t.Errorf("items[0].Columns = %v, want 42", items[0].Columns)
	}
	if items[1].Columns != nil {
		t.Errorf("items[1].Columns = %v, want nil", items[1].Columns)
	}
	if items[1].LastWritten != nil {
		t.Errorf("items[1].LastWritten = %v, want nil", items[1].LastWritten)
	}
}

func TestList_Empty(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []datasetItem
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
		Format:    output.FormatJSON,
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list"})
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
	cmd.SetArgs([]string{"list"})
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
