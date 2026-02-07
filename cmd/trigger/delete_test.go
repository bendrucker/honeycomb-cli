package trigger

import (
	"net/http"
	"strings"
	"testing"
)

func TestDelete_WithYes(t *testing.T) {
	var deleted bool
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			if r.URL.Path != "/1/triggers/test-dataset/trigger-1" {
				t.Errorf("path = %q, want /1/triggers/test-dataset/trigger-1", r.URL.Path)
			}
			deleted = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete", "--dataset", "test-dataset", "--yes", "trigger-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if !deleted {
		t.Error("DELETE was not called")
	}
}

func TestDelete_NoYesNonInteractive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	opts.NoInteractive = true
	opts.IOStreams.SetNeverPrompt(true)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete", "--dataset", "test-dataset", "trigger-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --yes")
	}
	if !strings.Contains(err.Error(), "--yes is required") {
		t.Errorf("error = %q, want --yes required message", err.Error())
	}
}

func TestDelete_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete", "--dataset", "test-dataset", "--yes", "nonexistent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("error = %q, want HTTP 404", err.Error())
	}
}
