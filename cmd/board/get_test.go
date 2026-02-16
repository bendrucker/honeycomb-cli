package board

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
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
			"preset_filters": []map[string]any{
				{"column": "service.name", "alias": "Service"},
				{"column": "env", "alias": "Environment"},
			},
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
	cmd.SetArgs([]string{"get", "abc123"})
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
	if detail.PresetFilters == nil {
		t.Fatal("PresetFilters is nil")
	}

	var filters []map[string]string
	if err := json.Unmarshal(detail.PresetFilters, &filters); err != nil {
		t.Fatalf("unmarshal preset_filters: %v", err)
	}
	if len(filters) != 2 {
		t.Fatalf("got %d preset_filters, want 2", len(filters))
	}
	if filters[0]["column"] != "service.name" {
		t.Errorf("filters[0].column = %q, want %q", filters[0]["column"], "service.name")
	}
	if filters[0]["alias"] != "Service" {
		t.Errorf("filters[0].alias = %q, want %q", filters[0]["alias"], "Service")
	}
}

func TestGet_NoPresetFilters(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "abc123",
			"name": "My Board",
			"type": "flexible",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "abc123"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail boardDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.PresetFilters != nil {
		t.Errorf("PresetFilters = %s, want nil", detail.PresetFilters)
	}
}

func TestGet_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"board not found"}`))
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

func TestGet_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}
