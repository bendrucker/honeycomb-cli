package dataset

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/bendrucker/honeycomb-cli/internal/output"
)

func TestUpdate(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/datasets/my-dataset" {
			t.Errorf("path = %q, want /1/datasets/my-dataset", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var payload api.DatasetUpdatePayload
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("unmarshal request body: %v", err)
		}
		if payload.Description != "Updated description" {
			t.Errorf("Description = %q, want %q", payload.Description, "Updated description")
		}
		if payload.ExpandJsonDepth != 5 {
			t.Errorf("ExpandJsonDepth = %d, want 5", payload.ExpandJsonDepth)
		}
		if payload.Settings == nil || payload.Settings.DeleteProtected == nil || !*payload.Settings.DeleteProtected {
			t.Errorf("Settings.DeleteProtected = %v, want true", payload.Settings)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":              "my-dataset",
			"slug":              "my-dataset",
			"description":       "Updated description",
			"expand_json_depth": 5,
			"created_at":        "2025-01-15T10:30:00Z",
			"settings": map[string]any{
				"delete_protected": true,
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "my-dataset", "--description", "Updated description", "--expand-json-depth", "5", "--delete-protected"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail datasetDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Description != "Updated description" {
		t.Errorf("Description = %q, want %q", detail.Description, "Updated description")
	}
	if detail.ExpandJsonDepth == nil || *detail.ExpandJsonDepth != 5 {
		t.Errorf("ExpandJsonDepth = %v, want 5", detail.ExpandJsonDepth)
	}
	if !detail.DeleteProtected {
		t.Errorf("DeleteProtected = false, want true")
	}
}

func TestUpdate_NoKey(t *testing.T) {
	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    "http://localhost",
		Format:    output.FormatJSON,
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "my-dataset", "--description", "test"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "no config key configured") {
		t.Errorf("error = %q, want missing key message", err.Error())
	}
}

func TestUpdate_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}
