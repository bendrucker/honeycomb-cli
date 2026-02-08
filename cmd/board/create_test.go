package board

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCreate_WithFlags(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/boards" {
			t.Errorf("path = %q, want /1/boards", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "My Board" {
			t.Errorf("name = %v, want %q", body["name"], "My Board")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "new-id",
			"name": "My Board",
			"type": "flexible",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--name", "My Board", "--description", "test"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail boardDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "new-id" {
		t.Errorf("ID = %q, want %q", detail.ID, "new-id")
	}
}

func TestCreate_WithFile(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]any
		_ = json.Unmarshal(body, &parsed)
		if parsed["name"] != "File Board" {
			t.Errorf("name = %v, want %q", parsed["name"], "File Board")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "file-id",
			"name": "File Board",
			"type": "flexible",
		})
	}))

	ts.InBuf.WriteString(`{"name":"File Board","type":"flexible"}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail boardDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "file-id" {
		t.Errorf("ID = %q, want %q", detail.ID, "file-id")
	}
}

func TestCreate_MissingInput(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	opts.IOStreams.SetNeverPrompt(true)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing input")
	}
	if !strings.Contains(err.Error(), "--name or --file is required") {
		t.Errorf("error = %q, want missing input message", err.Error())
	}
}
