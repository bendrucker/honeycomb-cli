package column

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCalculatedList(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/derived_columns/test" {
			t.Errorf("path = %q, want /1/derived_columns/test", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":          "dc1",
				"alias":       "latency_p99",
				"expression":  "HEATMAP(duration_ms)",
				"description": "P99 latency heatmap",
			},
			{
				"id":         "dc2",
				"alias":      "error_rate",
				"expression": "COUNT(status_code = 500) / COUNT()",
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--dataset", "test", "calculated", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []calculatedItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].Alias != "latency_p99" {
		t.Errorf("items[0].Alias = %q, want %q", items[0].Alias, "latency_p99")
	}
	if items[0].Expression != "HEATMAP(duration_ms)" {
		t.Errorf("items[0].Expression = %q, want %q", items[0].Expression, "HEATMAP(duration_ms)")
	}
	if items[1].Description != "" {
		t.Errorf("items[1].Description = %q, want empty", items[1].Description)
	}
}

func TestCalculatedList_Empty(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--dataset", "test", "calculated", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []calculatedItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

func TestCalculatedGet(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/derived_columns/test/dc1" {
			t.Errorf("path = %q, want /1/derived_columns/test/dc1", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "dc1",
			"alias":       "latency_p99",
			"expression":  "HEATMAP(duration_ms)",
			"description": "P99 latency heatmap",
			"created_at":  "2025-01-01T00:00:00Z",
			"updated_at":  "2025-01-02T00:00:00Z",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--dataset", "test", "calculated", "get", "dc1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail calculatedDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "dc1" {
		t.Errorf("ID = %q, want %q", detail.ID, "dc1")
	}
	if detail.Alias != "latency_p99" {
		t.Errorf("Alias = %q, want %q", detail.Alias, "latency_p99")
	}
	if detail.CreatedAt != "2025-01-01T00:00:00Z" {
		t.Errorf("CreatedAt = %q, want %q", detail.CreatedAt, "2025-01-01T00:00:00Z")
	}
}

func TestCalculatedGet_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--dataset", "test", "calculated", "get", "nonexistent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("error = %q, want HTTP 404", err.Error())
	}
}

func TestCalculatedCreate(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/1/derived_columns/test" {
			t.Errorf("path = %q, want /1/derived_columns/test", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["alias"] != "latency_p99" {
			t.Errorf("alias = %q, want %q", body["alias"], "latency_p99")
		}
		if body["expression"] != "HEATMAP(duration_ms)" {
			t.Errorf("expression = %q, want %q", body["expression"], "HEATMAP(duration_ms)")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         "dc1",
			"alias":      "latency_p99",
			"expression": "HEATMAP(duration_ms)",
			"created_at": "2025-01-01T00:00:00Z",
			"updated_at": "2025-01-01T00:00:00Z",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--dataset", "test", "calculated", "create", "--alias", "latency_p99", "--expression", "HEATMAP(duration_ms)"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail calculatedDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "dc1" {
		t.Errorf("ID = %q, want %q", detail.ID, "dc1")
	}
	if detail.Alias != "latency_p99" {
		t.Errorf("Alias = %q, want %q", detail.Alias, "latency_p99")
	}
}

func TestCalculatedUpdate(t *testing.T) {
	calls := 0
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet:
			calls++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":         "dc1",
				"alias":      "latency_p99",
				"expression": "HEATMAP(duration_ms)",
			})
		case r.Method == http.MethodPut:
			calls++
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			if body["alias"] != "latency_p99_updated" {
				t.Errorf("alias = %q, want %q", body["alias"], "latency_p99_updated")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":         "dc1",
				"alias":      "latency_p99_updated",
				"expression": "HEATMAP(duration_ms)",
				"updated_at": "2025-01-02T00:00:00Z",
			})
		}
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--dataset", "test", "calculated", "update", "dc1", "--alias", "latency_p99_updated"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if calls != 2 {
		t.Errorf("API calls = %d, want 2 (GET + PUT)", calls)
	}

	var detail calculatedDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Alias != "latency_p99_updated" {
		t.Errorf("Alias = %q, want %q", detail.Alias, "latency_p99_updated")
	}
}

func TestCalculatedUpdate_NoFlags(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--dataset", "test", "calculated", "update", "dc1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for no flags")
	}
	if !strings.Contains(err.Error(), "at least one") {
		t.Errorf("error = %q, want 'at least one' message", err.Error())
	}
}

func TestCalculatedDelete_WithYes(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/1/derived_columns/test/dc1" {
			t.Errorf("path = %q, want /1/derived_columns/test/dc1", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--dataset", "test", "calculated", "delete", "dc1", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(ts.ErrBuf.String(), "Calculated field dc1 deleted") {
		t.Errorf("stderr = %q, want deletion message", ts.ErrBuf.String())
	}
}

func TestCalculatedDelete_NoYesNonInteractive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--dataset", "test", "calculated", "delete", "dc1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --yes")
	}
	if !strings.Contains(err.Error(), "--yes is required") {
		t.Errorf("error = %q, want '--yes is required' message", err.Error())
	}
}
