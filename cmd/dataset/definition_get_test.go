package dataset

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/bendrucker/honeycomb-cli/internal/output"
)

func TestDefinitionGet(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/dataset_definitions/my-dataset" {
			t.Errorf("path = %q, want /1/dataset_definitions/my-dataset", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"duration_ms": map[string]any{
				"name":        "duration_ms",
				"column_type": "column",
			},
			"trace_id": map[string]any{
				"name":        "trace.trace_id",
				"column_type": "column",
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"definition", "get", "my-dataset"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var defs api.DatasetDefinitions
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &defs); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if defs.DurationMs == nil {
		t.Fatal("DurationMs is nil")
	}
	if defs.DurationMs.Name != "duration_ms" {
		t.Errorf("DurationMs.Name = %q, want %q", defs.DurationMs.Name, "duration_ms")
	}
	if defs.TraceId == nil {
		t.Fatal("TraceId is nil")
	}
	if defs.TraceId.Name != "trace.trace_id" {
		t.Errorf("TraceId.Name = %q, want %q", defs.TraceId.Name, "trace.trace_id")
	}
}

func TestDefinitionGet_NoKey(t *testing.T) {
	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    "http://localhost",
		Format:    output.FormatJSON,
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"definition", "get", "my-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "no config key configured") {
		t.Errorf("error = %q, want missing key message", err.Error())
	}
}

func TestDefinitionGet_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"definition", "get"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}
