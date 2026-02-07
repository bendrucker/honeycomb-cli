package dataset

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bendrucker/honeycomb-cli/internal/api"
)

func TestCreate(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/datasets" {
			t.Errorf("path = %q, want /1/datasets", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var payload api.DatasetCreationPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("unmarshal request body: %v", err)
		}
		if payload.Name != "my-dataset" {
			t.Errorf("Name = %q, want %q", payload.Name, "my-dataset")
		}
		if payload.Description == nil || *payload.Description != "A test dataset" {
			t.Errorf("Description = %v, want %q", payload.Description, "A test dataset")
		}
		if payload.ExpandJsonDepth == nil || *payload.ExpandJsonDepth != 3 {
			t.Errorf("ExpandJsonDepth = %v, want 3", payload.ExpandJsonDepth)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":              "my-dataset",
			"slug":              "my-dataset",
			"description":       "A test dataset",
			"expand_json_depth": 3,
			"created_at":        "2025-01-15T10:30:00Z",
			"settings": map[string]any{
				"delete_protected": false,
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--name", "my-dataset", "--description", "A test dataset", "--expand-json-depth", "3"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail datasetDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "my-dataset" {
		t.Errorf("Name = %q, want %q", detail.Name, "my-dataset")
	}
	if detail.Slug != "my-dataset" {
		t.Errorf("Slug = %q, want %q", detail.Slug, "my-dataset")
	}
	if detail.ExpandJsonDepth == nil || *detail.ExpandJsonDepth != 3 {
		t.Errorf("ExpandJsonDepth = %v, want 3", detail.ExpandJsonDepth)
	}
}

func TestCreate_200(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":       "existing",
			"slug":       "existing",
			"created_at": "2024-06-01T00:00:00Z",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--name", "existing"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail datasetDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "existing" {
		t.Errorf("Name = %q, want %q", detail.Name, "existing")
	}
}

func TestCreate_MissingName(t *testing.T) {
	opts, _ := setupTest(t, http.NotFoundHandler())

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "--name is required") {
		t.Errorf("error = %q, want --name required message", err.Error())
	}
}
