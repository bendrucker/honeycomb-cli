package trigger

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/triggers/test-dataset/trigger-1" {
			t.Errorf("path = %q, want /1/triggers/test-dataset/trigger-1", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":           "trigger-1",
			"name":         "High Latency",
			"description":  "P99 latency above threshold",
			"dataset_slug": "test-dataset",
			"disabled":     false,
			"triggered":    true,
			"alert_type":   "on_change",
			"frequency":    900,
			"threshold":    map[string]any{"op": ">", "value": 100, "exceeded_limit": 1},
			"query_id":     "abc-123",
			"recipients": []map[string]any{
				{"id": "r1", "type": "email", "target": "team@example.com"},
			},
			"tags": []map[string]any{
				{"key": "env", "value": "production"},
			},
			"created_at": "2025-01-15T10:30:00Z",
			"updated_at": "2025-06-01T12:00:00Z",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "--dataset", "test-dataset", "trigger-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail triggerDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "trigger-1" {
		t.Errorf("ID = %q, want %q", detail.ID, "trigger-1")
	}
	if detail.Name != "High Latency" {
		t.Errorf("Name = %q, want %q", detail.Name, "High Latency")
	}
	if detail.Frequency != 900 {
		t.Errorf("Frequency = %d, want 900", detail.Frequency)
	}
	if detail.Threshold == nil {
		t.Fatal("Threshold is nil")
	}
	if detail.Threshold.Op != ">" {
		t.Errorf("Threshold.Op = %q, want %q", detail.Threshold.Op, ">")
	}
	if detail.Threshold.Value != 100 {
		t.Errorf("Threshold.Value = %g, want 100", detail.Threshold.Value)
	}
	if detail.Threshold.ExceededLimit != 1 {
		t.Errorf("Threshold.ExceededLimit = %d, want 1", detail.Threshold.ExceededLimit)
	}
	if detail.QueryID != "abc-123" {
		t.Errorf("QueryID = %q, want %q", detail.QueryID, "abc-123")
	}
	if len(detail.Recipients) != 1 {
		t.Fatalf("Recipients len = %d, want 1", len(detail.Recipients))
	}
	if detail.Recipients[0].Target != "team@example.com" {
		t.Errorf("Recipients[0].Target = %q, want %q", detail.Recipients[0].Target, "team@example.com")
	}
	if len(detail.Tags) != 1 {
		t.Fatalf("Tags len = %d, want 1", len(detail.Tags))
	}
	if detail.Tags[0].Key != "env" {
		t.Errorf("Tags[0].Key = %q, want %q", detail.Tags[0].Key, "env")
	}
}

func TestGet_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "--dataset", "test-dataset", "nonexistent"})
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
	cmd.SetArgs([]string{"get", "--dataset", "test-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}
