package dataset

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bendrucker/honeycomb-cli/internal/api"
)

func TestDefinitionUpdate(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/dataset_definitions/my-dataset" {
			t.Errorf("path = %q, want /1/dataset_definitions/my-dataset", r.URL.Path)
		}
		if r.Method != http.MethodPatch {
			t.Errorf("method = %q, want PATCH", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("unmarshal request body: %v", err)
		}
		dm, ok := payload["duration_ms"]
		if !ok {
			t.Fatal("expected duration_ms in payload")
		}
		dmMap := dm.(map[string]any)
		if dmMap["name"] != "duration_ms" {
			t.Errorf("duration_ms.name = %q, want %q", dmMap["name"], "duration_ms")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"duration_ms": map[string]any{
				"name":        "duration_ms",
				"column_type": "column",
			},
		})
	}))

	ts.InBuf.WriteString(`{"duration_ms": {"name": "duration_ms", "column_type": "column"}}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"definition", "update", "my-dataset", "--file", "-"})
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
}

func TestDefinitionUpdate_MissingFile(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"definition", "update", "my-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --file")
	}
	if !strings.Contains(err.Error(), "required flag") {
		t.Errorf("error = %q, want required flag message", err.Error())
	}
}

func TestDefinitionUpdate_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"definition", "update", "--file", "-"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}
