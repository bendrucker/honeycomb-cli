package query

import (
	"net/http"
	"strings"
	"testing"
)

func TestDelete_Yes(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/query_annotations/test-dataset/ann-1" {
			t.Errorf("path = %q, want /1/query_annotations/test-dataset/ann-1", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete", "--dataset", "test-dataset", "--yes", "ann-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(ts.ErrBuf.String(), "Query annotation ann-1 deleted") {
		t.Errorf("stderr = %q, want deletion confirmation", ts.ErrBuf.String())
	}
}

func TestDelete_NoYes_NonInteractive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete", "--dataset", "test-dataset", "ann-1"})
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
		_, _ = w.Write([]byte(`{"error":"Query Annotation not found"}`))
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
