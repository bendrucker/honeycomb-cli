package query

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestView(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/1/query_annotations/test-dataset/ann-1":
			_, _ = w.Write([]byte(`{
				"id": "ann-1",
				"name": "Latency Query",
				"description": "P99 latency query",
				"query_id": "q-abc",
				"source": "query",
				"created_at": "2025-01-15T10:30:00Z",
				"updated_at": "2025-02-01T12:00:00Z"
			}`))
		case "/1/queries/test-dataset/q-abc":
			_, _ = w.Write([]byte(`{
				"id": "q-abc",
				"time_range": 7200,
				"breakdowns": ["service.name"],
				"calculations": [{"op": "P99", "column": "duration_ms"}],
				"filters": [{"column": "status", "op": "=", "value": "error"}]
			}`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "--dataset", "test-dataset", "ann-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var result annotationWithQuery
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result.ID != "ann-1" {
		t.Errorf("ID = %q, want %q", result.ID, "ann-1")
	}
	if result.Name != "Latency Query" {
		t.Errorf("Name = %q, want %q", result.Name, "Latency Query")
	}
	if result.QueryID != "q-abc" {
		t.Errorf("QueryID = %q, want %q", result.QueryID, "q-abc")
	}
	if result.Query == nil {
		t.Fatal("Query is nil")
	}
	if result.Query.TimeRange == nil || *result.Query.TimeRange != 7200 {
		t.Errorf("Query.TimeRange = %v, want 7200", result.Query.TimeRange)
	}
}

func TestView_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"Query Annotation not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "--dataset", "test-dataset", "nonexistent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("error = %q, want HTTP 404", err.Error())
	}
}

func TestView_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "--dataset", "test-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}
