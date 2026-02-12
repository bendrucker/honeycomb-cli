package column

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCalculatedGet(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/derived_columns/my-dataset/dc-1" {
			t.Errorf("path = %q, want /1/derived_columns/my-dataset/dc-1", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "dc-1",
			"alias":       "req.duration_ms",
			"expression":  "SUB($duration_ms, $overhead_ms)",
			"description": "Net request duration",
			"created_at":  "2024-06-01T00:00:00Z",
			"updated_at":  "2025-01-15T10:30:00Z",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"calculated", "get", "--dataset", "my-dataset", "dc-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail calculatedDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "dc-1" {
		t.Errorf("ID = %q, want %q", detail.ID, "dc-1")
	}
	if detail.Alias != "req.duration_ms" {
		t.Errorf("Alias = %q, want %q", detail.Alias, "req.duration_ms")
	}
	if detail.Expression != "SUB($duration_ms, $overhead_ms)" {
		t.Errorf("Expression = %q, want %q", detail.Expression, "SUB($duration_ms, $overhead_ms)")
	}
	if detail.CreatedAt != "2024-06-01T00:00:00Z" {
		t.Errorf("CreatedAt = %q, want %q", detail.CreatedAt, "2024-06-01T00:00:00Z")
	}
}

func TestCalculatedGet_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"calculated column not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"calculated", "get", "--dataset", "my-dataset", "bad-id"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("error = %q, want HTTP 404", err.Error())
	}
}

func TestCalculatedGet_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"calculated", "get", "--dataset", "my-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}
