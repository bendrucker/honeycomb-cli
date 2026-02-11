package dataset

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestUpdate(t *testing.T) {
	var gotBody map[string]any

	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/datasets/production" {
			t.Errorf("path = %q, want /1/datasets/production", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name":              "production",
				"slug":              "production",
				"description":       "Original description",
				"expand_json_depth": 2,
				"created_at":        "2024-06-01T00:00:00Z",
				"settings":          map[string]any{"delete_protected": true},
			})
			return
		}

		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT", r.Method)
		}

		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":        "production",
			"slug":        "production",
			"description": "Updated description",
			"created_at":  "2024-06-01T00:00:00Z",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "production", "--description", "Updated description"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail datasetDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Description != "Updated description" {
		t.Errorf("description = %q, want %q", detail.Description, "Updated description")
	}

	if gotBody["description"] != "Updated description" {
		t.Errorf("request body description = %v, want %q", gotBody["description"], "Updated description")
	}
	if gotBody["expand_json_depth"] != float64(2) {
		t.Errorf("request body expand_json_depth = %v, want 2 (preserved from GET)", gotBody["expand_json_depth"])
	}
}

func TestUpdate_NoFlags(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "production"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for no flags")
	}
	if !strings.Contains(err.Error(), "at least one of") {
		t.Errorf("error = %q, want flag requirement error", err.Error())
	}
}

func TestUpdate_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"dataset not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "nonexistent", "--description", "test"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("error = %q, want HTTP 404", err.Error())
	}
}
