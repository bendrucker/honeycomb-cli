package recipient

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
		if r.URL.Path != "/1/recipients" {
			t.Errorf("path = %q, want /1/recipients", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"id": "r1",
				"type": "email",
				"details": {"address": "test@example.com"},
				"created_at": "2025-01-01T00:00:00Z",
				"updated_at": "2025-01-01T00:00:00Z"
			},
			{
				"id": "r2",
				"type": "slack",
				"details": {"channel": "#alerts"},
				"created_at": "2025-01-02T00:00:00Z",
				"updated_at": "2025-01-02T00:00:00Z"
			}
		]`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []recipientItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != "r1" {
		t.Errorf("items[0].ID = %q, want %q", items[0].ID, "r1")
	}
	if items[0].Type != "email" {
		t.Errorf("items[0].Type = %q, want %q", items[0].Type, "email")
	}
	if items[0].Target != "test@example.com" {
		t.Errorf("items[0].Target = %q, want %q", items[0].Target, "test@example.com")
	}
	if items[1].Type != "slack" {
		t.Errorf("items[1].Type = %q, want %q", items[1].Type, "slack")
	}
	if items[1].Target != "#alerts" {
		t.Errorf("items[1].Target = %q, want %q", items[1].Target, "#alerts")
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

	var items []recipientItem
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

func TestGet(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/recipients/r1" {
			t.Errorf("path = %q, want /1/recipients/r1", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "r1",
			"type": "email",
			"details": {"address": "test@example.com"},
			"created_at": "2025-01-01T00:00:00Z",
			"updated_at": "2025-01-01T00:00:00Z"
		}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "r1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail recipientDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "r1" {
		t.Errorf("ID = %q, want %q", detail.ID, "r1")
	}
	if detail.Type != "email" {
		t.Errorf("Type = %q, want %q", detail.Type, "email")
	}
	if detail.CreatedAt != "2025-01-01T00:00:00Z" {
		t.Errorf("CreatedAt = %q, want %q", detail.CreatedAt, "2025-01-01T00:00:00Z")
	}
}

func TestGet_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"recipient not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "missing"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("error = %q, want HTTP 404", err.Error())
	}
}

func TestGet_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}

func TestDelete_WithYes(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/recipients/r1" {
			t.Errorf("path = %q, want /1/recipients/r1", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete", "r1", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var result map[string]string
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result["id"] != "r1" {
		t.Errorf("id = %q, want %q", result["id"], "r1")
	}
}

func TestDelete_NoYesNonInteractive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	opts.IOStreams.SetNeverPrompt(true)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete", "r1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-interactive without --yes")
	}
	if !strings.Contains(err.Error(), "--yes is required in non-interactive mode") {
		t.Errorf("error = %q, want non-interactive error", err.Error())
	}
}

func TestDelete_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}

func TestTriggers(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/recipients/r1/triggers" {
			t.Errorf("path = %q, want /1/recipients/r1/triggers", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":          "t1",
				"name":        "High Latency",
				"description": "Fires when p99 > 1s",
				"disabled":    false,
				"triggered":   true,
				"alert_type":  "on_change",
				"threshold":   map[string]any{"op": ">", "value": 1000},
			},
			{
				"id":        "t2",
				"name":      "Error Rate",
				"disabled":  false,
				"triggered": false,
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"triggers", "r1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []triggerItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].Name != "High Latency" {
		t.Errorf("items[0].Name = %q, want %q", items[0].Name, "High Latency")
	}
	if !items[0].Triggered {
		t.Errorf("items[0].Triggered = false, want true")
	}
	if items[0].Threshold != "> 1000" {
		t.Errorf("items[0].Threshold = %q, want %q", items[0].Threshold, "> 1000")
	}
	if items[1].Name != "Error Rate" {
		t.Errorf("items[1].Name = %q, want %q", items[1].Name, "Error Rate")
	}
}

func TestTriggers_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"triggers"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}
