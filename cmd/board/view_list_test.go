package board

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestViewList(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/boards/board-1/views" {
			t.Errorf("path = %q, want /1/boards/board-1/views", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": "v1", "name": "View One"},
			{"id": "v2", "name": "View Two"},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "list", "--board", "board-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []viewItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != "v1" {
		t.Errorf("items[0].ID = %q, want %q", items[0].ID, "v1")
	}
	if items[0].Name != "View One" {
		t.Errorf("items[0].Name = %q, want %q", items[0].Name, "View One")
	}
}

func TestViewList_Empty(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "list", "--board", "board-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []viewItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}
