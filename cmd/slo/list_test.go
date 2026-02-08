package slo

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

	ts := iostreams.Test()
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
		if r.URL.Path != "/1/slos/test-dataset" {
			t.Errorf("path = %q, want /1/slos/test-dataset", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":                 "slo-1",
				"name":               "Availability",
				"target_per_million": 999000,
				"time_period_days":   30,
				"sli":                map[string]any{"alias": "sli.availability"},
				"description":        "Service availability SLO",
			},
			{
				"id":                 "slo-2",
				"name":               "Latency",
				"target_per_million": 990000,
				"time_period_days":   7,
				"sli":                map[string]any{"alias": "sli.latency"},
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list", "--dataset", "test-dataset"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []sloItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].Name != "Availability" {
		t.Errorf("items[0].Name = %q, want %q", items[0].Name, "Availability")
	}
	if items[0].TargetPerMillion != 999000 {
		t.Errorf("items[0].TargetPerMillion = %d, want 999000", items[0].TargetPerMillion)
	}
	if items[0].SLIAlias != "sli.availability" {
		t.Errorf("items[0].SLIAlias = %q, want %q", items[0].SLIAlias, "sli.availability")
	}
	if items[1].Description != "" {
		t.Errorf("items[1].Description = %q, want empty", items[1].Description)
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

	var items []sloItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

func TestList_NoKey(t *testing.T) {
	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    "http://localhost",
		Format:    output.FormatJSON,
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
