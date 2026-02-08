package board

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestView(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/boards/abc123" {
			t.Errorf("path = %q, want /1/boards/abc123", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                "abc123",
			"name":              "My Board",
			"description":       "A test board",
			"type":              "flexible",
			"layout_generation": "auto",
			"links":             map[string]any{"board_url": "https://ui.honeycomb.io/boards/abc123"},
			"panels": []map[string]any{
				{
					"type":        "query",
					"query_panel": map[string]any{"query_id": "q1", "query_annotation_id": "qa1"},
				},
				{
					"type":       "text",
					"text_panel": map[string]any{"content": "Hello world"},
				},
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "abc123"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail boardDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "abc123" {
		t.Errorf("ID = %q, want %q", detail.ID, "abc123")
	}
	if detail.Name != "My Board" {
		t.Errorf("Name = %q, want %q", detail.Name, "My Board")
	}
	if detail.Panels == nil {
		t.Fatal("Panels is nil")
	}
}

func TestView_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"board not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "missing"})
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
	cmd.SetArgs([]string{"view"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}
