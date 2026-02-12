package dataset

import (
	"net/http"
	"strings"
	"testing"
)

func TestDelete_WithYes(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/datasets/my-dataset" {
			t.Errorf("path = %q, want /1/datasets/my-dataset", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete", "my-dataset", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(ts.ErrBuf.String(), "Dataset my-dataset deleted") {
		t.Errorf("stderr = %q, want deletion confirmation", ts.ErrBuf.String())
	}
}

func TestDelete_NoYesNonInteractive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	opts.IOStreams.SetNeverPrompt(true)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete", "my-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-interactive without --yes")
	}
	if !strings.Contains(err.Error(), "--yes is required in non-interactive mode") {
		t.Errorf("error = %q, want non-interactive error", err.Error())
	}
}

func TestDelete_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}
