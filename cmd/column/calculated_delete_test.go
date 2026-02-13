package column

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCalculatedDelete_Yes(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/1/derived_columns/my-dataset/dc-1" {
			t.Errorf("path = %q, want /1/derived_columns/my-dataset/dc-1", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"calculated", "delete", "--dataset", "my-dataset", "--yes", "dc-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var result map[string]string
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result["id"] != "dc-1" {
		t.Errorf("id = %q, want %q", result["id"], "dc-1")
	}
}

func TestCalculatedDelete_NoYes_NonInteractive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	opts.IOStreams.SetNeverPrompt(true)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"calculated", "delete", "--dataset", "my-dataset", "dc-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --yes")
	}
	if !strings.Contains(err.Error(), "--yes is required in non-interactive mode") {
		t.Errorf("error = %q, want --yes required message", err.Error())
	}
}
