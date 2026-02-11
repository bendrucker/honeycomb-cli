package dataset

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/datasets/production" {
			t.Errorf("path = %q, want /1/datasets/production", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":                  "production",
			"slug":                  "production",
			"description":           "Production events",
			"expand_json_depth":     2,
			"regular_columns_count": 42,
			"last_written_at":       "2025-01-15T10:30:00Z",
			"created_at":            "2024-06-01T00:00:00Z",
			"settings": map[string]any{
				"delete_protected": true,
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "production"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail datasetDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "production" {
		t.Errorf("Name = %q, want %q", detail.Name, "production")
	}
	if detail.Slug != "production" {
		t.Errorf("Slug = %q, want %q", detail.Slug, "production")
	}
	if detail.Description != "Production events" {
		t.Errorf("Description = %q, want %q", detail.Description, "Production events")
	}
	if detail.ExpandJsonDepth == nil || *detail.ExpandJsonDepth != 2 {
		t.Errorf("ExpandJsonDepth = %v, want 2", detail.ExpandJsonDepth)
	}
	if detail.Columns == nil || *detail.Columns != 42 {
		t.Errorf("Columns = %v, want 42", detail.Columns)
	}
	if detail.LastWritten == nil || *detail.LastWritten != "2025-01-15T10:30:00Z" {
		t.Errorf("LastWritten = %v, want %q", detail.LastWritten, "2025-01-15T10:30:00Z")
	}
	if !detail.DeleteProtected {
		t.Error("DeleteProtected = false, want true")
	}
}

func TestGet_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"dataset not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "nonexistent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("error = %q, want HTTP 404", err.Error())
	}
}

func TestGet_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.NotFoundHandler())

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Errorf("error = %q, want missing arg message", err.Error())
	}
}
