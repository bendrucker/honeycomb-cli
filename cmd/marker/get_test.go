package marker

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":         "abc123",
				"type":       "deploy",
				"message":    "v1.0.0",
				"url":        "https://example.com",
				"start_time": 1700000000,
				"end_time":   1700000300,
				"color":      "#ff0000",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z",
			},
			{
				"id":      "def456",
				"type":    "incident",
				"message": "outage",
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "--dataset", "test-dataset", "abc123"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var item markerItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &item); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if item.ID != "abc123" {
		t.Errorf("ID = %q, want %q", item.ID, "abc123")
	}
	if item.Type != "deploy" {
		t.Errorf("Type = %q, want %q", item.Type, "deploy")
	}
	if item.URL != "https://example.com" {
		t.Errorf("URL = %q, want %q", item.URL, "https://example.com")
	}
}

func TestGet_ViewAlias(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": "abc123", "type": "deploy"},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "--dataset", "test-dataset", "abc123"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestGet_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": "abc123", "type": "deploy"},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "--dataset", "test-dataset", "zzz999"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !strings.Contains(err.Error(), `marker "zzz999" not found`) {
		t.Errorf("error = %q, want not found message", err.Error())
	}
}

func TestGet_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "--dataset", "test-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}
