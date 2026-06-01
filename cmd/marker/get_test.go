package marker

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/markers/test-dataset" {
			t.Errorf("path = %q, want /1/markers/test-dataset", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":      "abc123",
				"type":    "deploy",
				"message": "v1.0.0",
			},
			{
				"id":      "def456",
				"type":    "incident",
				"message": "outage",
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "--dataset", "test-dataset", "def456"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var item markerItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &item); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if item.ID != "def456" {
		t.Errorf("id = %q, want %q", item.ID, "def456")
	}
	if item.Type != "incident" {
		t.Errorf("type = %q, want %q", item.Type, "incident")
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
	cmd.SetArgs([]string{"get", "--dataset", "test-dataset", "missing"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing marker")
	}
	if !strings.Contains(err.Error(), `marker "missing" not found`) {
		t.Errorf("error = %q, want not found message", err.Error())
	}
}
