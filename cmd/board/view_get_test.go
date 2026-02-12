package board

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestViewGet(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/boards/board-1/views/v1" {
			t.Errorf("path = %q, want /1/boards/board-1/views/v1", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "v1",
			"name": "View One",
			"filters": []map[string]any{
				{"column": "env", "operation": "=", "value": "prod"},
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "get", "v1", "--board", "board-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail viewDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "v1" {
		t.Errorf("ID = %q, want %q", detail.ID, "v1")
	}
	if detail.Name != "View One" {
		t.Errorf("Name = %q, want %q", detail.Name, "View One")
	}
	if detail.Filters == nil {
		t.Fatal("Filters is nil")
	}
}

func TestViewGet_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"view not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "get", "missing", "--board", "board-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("error = %q, want HTTP 404", err.Error())
	}
}

func TestViewGet_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "get", "--board", "board-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}
