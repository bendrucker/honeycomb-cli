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
	"github.com/zalando/go-keyring"
)

func init() {
	keyring.MockInit()
}

func setupBurnAlertTest(t *testing.T, handler http.Handler) (*options.RootOptions, *iostreams.TestStreams) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	ts := iostreams.Test()
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

func TestBurnAlertList(t *testing.T) {
	opts, ts := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/burn_alerts/my-dataset" {
			t.Errorf("path = %q, want /1/burn_alerts/my-dataset", r.URL.Path)
		}
		if got := r.URL.Query().Get("slo_id"); got != "slo-123" {
			t.Errorf("slo_id = %q, want %q", got, "slo-123")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":         "ba-1",
				"alert_type": "exhaustion_time",
				"slo_id":     "slo-123",
				"created_at": "2024-01-01T00:00:00Z",
			},
			{
				"id":         "ba-2",
				"alert_type": "budget_rate",
				"slo_id":     "slo-123",
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"burn-alert", "list", "--dataset", "my-dataset", "--slo-id", "slo-123"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []burnAlertItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != "ba-1" {
		t.Errorf("items[0].ID = %q, want %q", items[0].ID, "ba-1")
	}
	if items[0].AlertType != "exhaustion_time" {
		t.Errorf("items[0].AlertType = %q, want %q", items[0].AlertType, "exhaustion_time")
	}
	if items[1].AlertType != "budget_rate" {
		t.Errorf("items[1].AlertType = %q, want %q", items[1].AlertType, "budget_rate")
	}
}

func TestBurnAlertList_Empty(t *testing.T) {
	opts, ts := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"burn-alert", "list", "--dataset", "my-dataset", "--slo-id", "slo-123"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []burnAlertItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

func TestBurnAlertGet(t *testing.T) {
	opts, ts := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/burn_alerts/my-dataset/ba-1" {
			t.Errorf("path = %q, want /1/burn_alerts/my-dataset/ba-1", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                 "ba-1",
			"alert_type":         "exhaustion_time",
			"description":        "Alert when budget exhausted",
			"slo_id":             "slo-123",
			"exhaustion_minutes": 240,
			"created_at":         "2024-01-01T00:00:00Z",
			"updated_at":         "2024-01-02T00:00:00Z",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"burn-alert", "get", "ba-1", "--dataset", "my-dataset"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail burnAlertDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "ba-1" {
		t.Errorf("ID = %q, want %q", detail.ID, "ba-1")
	}
	if detail.AlertType != "exhaustion_time" {
		t.Errorf("AlertType = %q, want %q", detail.AlertType, "exhaustion_time")
	}
	if detail.Description != "Alert when budget exhausted" {
		t.Errorf("Description = %q, want %q", detail.Description, "Alert when budget exhausted")
	}
	if detail.ExhaustionMinutes == nil || *detail.ExhaustionMinutes != 240 {
		t.Errorf("ExhaustionMinutes = %v, want 240", detail.ExhaustionMinutes)
	}
}

func TestBurnAlertDelete(t *testing.T) {
	opts, ts := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/1/burn_alerts/my-dataset/ba-1" {
			t.Errorf("path = %q, want /1/burn_alerts/my-dataset/ba-1", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"burn-alert", "delete", "ba-1", "--dataset", "my-dataset", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var result map[string]string
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result["id"] != "ba-1" {
		t.Errorf("id = %q, want %q", result["id"], "ba-1")
	}
}

func TestBurnAlertDelete_RequiresYesNonInteractive(t *testing.T) {
	opts, _ := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("unexpected API call")
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"burn-alert", "delete", "ba-1", "--dataset", "my-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --yes")
	}
	if !strings.Contains(err.Error(), "--yes is required") {
		t.Errorf("error = %q, want --yes required message", err.Error())
	}
}

func TestBurnAlertList_NoKey(t *testing.T) {
	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    "http://localhost",
		Format:    "json",
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"burn-alert", "list", "--dataset", "my-dataset", "--slo-id", "slo-123"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "no config key configured") {
		t.Errorf("error = %q, want missing key message", err.Error())
	}
}
