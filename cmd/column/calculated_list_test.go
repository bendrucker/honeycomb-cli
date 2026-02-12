package column

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCalculatedList(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/derived_columns/my-dataset" {
			t.Errorf("path = %q, want /1/derived_columns/my-dataset", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":          "dc-1",
				"alias":       "req.duration_ms",
				"expression":  "SUB($duration_ms, $overhead_ms)",
				"description": "Net request duration",
			},
			{
				"id":         "dc-2",
				"alias":      "is_error",
				"expression": "GTE($status_code, 400)",
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"calculated", "list", "--dataset", "my-dataset"})
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
	if items[0].Alias != "req.duration_ms" {
		t.Errorf("items[0].Alias = %q, want %q", items[0].Alias, "req.duration_ms")
	}
	if items[0].Expression != "SUB($duration_ms, $overhead_ms)" {
		t.Errorf("items[0].Expression = %q, want %q", items[0].Expression, "SUB($duration_ms, $overhead_ms)")
	}
	if items[0].Description != "Net request duration" {
		t.Errorf("items[0].Description = %q, want %q", items[0].Description, "Net request duration")
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
	cmd.SetArgs([]string{"calculated", "list", "--dataset", "my-dataset"})
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

func TestCalculatedList_Unauthorized(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unknown API key - check your credentials"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"calculated", "list", "--dataset", "my-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if !strings.Contains(err.Error(), "HTTP 401") {
		t.Errorf("error = %q, want HTTP 401", err.Error())
	}
}
