package recipient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:   &config.Config{},
		APIUrl:   srv.URL,
		Format:   "json",
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
			{"id": "abc123", "type": "email", "details": {"email_address": "test@example.com"}},
			{"id": "def456", "type": "slack", "details": {"slack_channel": "#alerts"}}
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
	if items[0].Type != "email" {
		t.Errorf("items[0].Type = %q, want %q", items[0].Type, "email")
	}
	if items[0].Target != "test@example.com" {
		t.Errorf("items[0].Target = %q, want %q", items[0].Target, "test@example.com")
	}
	if items[1].Target != "#alerts" {
		t.Errorf("items[1].Target = %q, want %q", items[1].Target, "#alerts")
	}
}

func TestGet(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/recipients/abc123" {
			t.Errorf("path = %q, want /1/recipients/abc123", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "abc123",
			"type": "email",
			"details": {"email_address": "test@example.com"},
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-02T00:00:00Z"
		}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "abc123"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail recipientDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "abc123" {
		t.Errorf("ID = %q, want %q", detail.ID, "abc123")
	}
	if detail.Type != "email" {
		t.Errorf("Type = %q, want %q", detail.Type, "email")
	}
	if detail.Target != "test@example.com" {
		t.Errorf("Target = %q, want %q", detail.Target, "test@example.com")
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

func TestCreate(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/1/recipients" {
			t.Errorf("path = %q, want /1/recipients", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "new123",
			"type": "email",
			"details": {"email_address": "new@example.com"},
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z"
		}`))
	}))

	tmpFile := filepath.Join(t.TempDir(), "recipient.json")
	if err := os.WriteFile(tmpFile, []byte(`{"type":"email","details":{"email_address":"new@example.com"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--file", tmpFile})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail recipientDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "new123" {
		t.Errorf("ID = %q, want %q", detail.ID, "new123")
	}
	if detail.Target != "new@example.com" {
		t.Errorf("Target = %q, want %q", detail.Target, "new@example.com")
	}
}

func TestUpdate(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT", r.Method)
		}
		if r.URL.Path != "/1/recipients/abc123" {
			t.Errorf("path = %q, want /1/recipients/abc123", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "abc123",
			"type": "email",
			"details": {"email_address": "updated@example.com"},
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-03T00:00:00Z"
		}`))
	}))

	tmpFile := filepath.Join(t.TempDir(), "recipient.json")
	if err := os.WriteFile(tmpFile, []byte(`{"type":"email","details":{"email_address":"updated@example.com"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "abc123", "--file", tmpFile})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail recipientDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Target != "updated@example.com" {
		t.Errorf("Target = %q, want %q", detail.Target, "updated@example.com")
	}
}

func TestDelete_WithYes(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/1/recipients/abc123" {
			t.Errorf("path = %q, want /1/recipients/abc123", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete", "abc123", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(ts.ErrBuf.String(), "Recipient abc123 deleted") {
		t.Errorf("stderr = %q, want delete message", ts.ErrBuf.String())
	}
}

func TestTriggers(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/recipients/abc123/triggers" {
			t.Errorf("path = %q, want /1/recipients/abc123/triggers", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": "t1", "name": "High Latency", "dataset_slug": "production"},
			{"id": "t2", "name": "Error Rate", "dataset_slug": "staging"},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"triggers", "abc123"})
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
	if items[0].Dataset != "production" {
		t.Errorf("items[0].Dataset = %q, want %q", items[0].Dataset, "production")
	}
}
