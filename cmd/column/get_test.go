package column

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/columns/my-dataset/abc123" {
			t.Errorf("path = %q, want /1/columns/my-dataset/abc123", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":           "abc123",
			"key_name":     "duration_ms",
			"type":         "float",
			"description":  "Request duration",
			"hidden":       false,
			"last_written": "2025-01-15T10:30:00Z",
			"created_at":   "2024-06-01T00:00:00Z",
			"updated_at":   "2025-01-15T10:30:00Z",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "--dataset", "my-dataset", "abc123"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail columnDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "abc123" {
		t.Errorf("ID = %q, want %q", detail.ID, "abc123")
	}
	if detail.KeyName != "duration_ms" {
		t.Errorf("KeyName = %q, want %q", detail.KeyName, "duration_ms")
	}
	if detail.CreatedAt != "2024-06-01T00:00:00Z" {
		t.Errorf("CreatedAt = %q, want %q", detail.CreatedAt, "2024-06-01T00:00:00Z")
	}
}

func TestGet_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"column not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "--dataset", "my-dataset", "bad-id"})
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
	cmd.SetArgs([]string{"get", "--dataset", "my-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}
