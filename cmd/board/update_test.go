package board

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestUpdate_WithName(t *testing.T) {
	calls := 0
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		calls++
		if calls == 1 {
			// GET current board
			if r.Method != http.MethodGet {
				t.Errorf("call 1: method = %q, want GET", r.Method)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "abc123",
				"name":        "Old Name",
				"description": "Old desc",
				"type":        "flexible",
			})
			return
		}
		// PUT update
		if r.Method != http.MethodPut {
			t.Errorf("call 2: method = %q, want PUT", r.Method)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "New Name" {
			t.Errorf("name = %v, want %q", body["name"], "New Name")
		}
		if body["description"] != "Old desc" {
			t.Errorf("description = %v, want %q", body["description"], "Old desc")
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "abc123",
			"name":        "New Name",
			"description": "Old desc",
			"type":        "flexible",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "abc123", "--name", "New Name"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail boardDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "New Name" {
		t.Errorf("Name = %q, want %q", detail.Name, "New Name")
	}
}

func TestUpdate_MissingFlags(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "abc123"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing flags")
	}
	if !strings.Contains(err.Error(), "--file, --name, or --description is required") {
		t.Errorf("error = %q, want missing flags message", err.Error())
	}
}

func TestUpdate_FileStdinMerge(t *testing.T) {
	calls := 0
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		calls++
		if calls == 1 {
			if r.Method != http.MethodGet {
				t.Errorf("call 1: method = %q, want GET", r.Method)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "abc123",
				"name":        "Old Name",
				"description": "Old desc",
				"type":        "flexible",
			})
			return
		}
		if r.Method != http.MethodPut {
			t.Errorf("call 2: method = %q, want PUT", r.Method)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Merged Name" {
			t.Errorf("name = %v, want %q", body["name"], "Merged Name")
		}
		if body["description"] != "Old desc" {
			t.Errorf("description = %v, want %q", body["description"], "Old desc")
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "abc123",
			"name":        "Merged Name",
			"description": "Old desc",
			"type":        "flexible",
		})
	}))

	ts.InBuf.WriteString(`{"name":"Merged Name"}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "abc123", "--file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail boardDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "Merged Name" {
		t.Errorf("Name = %q, want %q", detail.Name, "Merged Name")
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2 (GET + PUT)", calls)
	}
}

func TestUpdate_FileStdinReplace(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT (no GET for replace)", r.Method)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Replaced" {
			t.Errorf("name = %v, want %q", body["name"], "Replaced")
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "abc123",
			"name": "Replaced",
			"type": "flexible",
		})
	}))

	ts.InBuf.WriteString(`{"name":"Replaced","type":"flexible"}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "abc123", "--file", "-", "--replace"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail boardDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "Replaced" {
		t.Errorf("Name = %q, want %q", detail.Name, "Replaced")
	}
}

func TestUpdate_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"board not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "missing", "--name", "New"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("error = %q, want HTTP 404", err.Error())
	}
}
