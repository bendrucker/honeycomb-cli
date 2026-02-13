package board

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestViewDelete_WithYes(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/boards/board-1/views/v1" {
			t.Errorf("path = %q, want /1/boards/board-1/views/v1", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "delete", "v1", "--board", "board-1", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var result map[string]string
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result["id"] != "v1" {
		t.Errorf("id = %q, want %q", result["id"], "v1")
	}
}

func TestViewDelete_NoYesNonInteractive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	opts.IOStreams.SetNeverPrompt(true)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "delete", "v1", "--board", "board-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-interactive without --yes")
	}
	if !strings.Contains(err.Error(), "--yes is required in non-interactive mode") {
		t.Errorf("error = %q, want non-interactive error", err.Error())
	}
}

func TestViewDelete_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "delete", "--board", "board-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}
