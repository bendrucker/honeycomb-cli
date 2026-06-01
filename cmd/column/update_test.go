package column

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdate(t *testing.T) {
	calls := 0
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			calls++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "abc123",
				"key_name":    "duration_ms",
				"type":        "float",
				"description": "Old description",
				"hidden":      false,
			})
		case http.MethodPut:
			calls++
			body, _ := io.ReadAll(r.Body)
			var req map[string]any
			_ = json.Unmarshal(body, &req)
			if req["description"] != "New description" {
				t.Errorf("description = %v, want New description", req["description"])
			}
			if req["hidden"] != true {
				t.Errorf("hidden = %v, want true", req["hidden"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "abc123",
				"key_name":    "duration_ms",
				"type":        "float",
				"description": "New description",
				"hidden":      true,
				"updated_at":  "2025-01-16T00:00:00Z",
			})
		}
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "my-dataset", "abc123", "--description", "New description", "--hidden"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if calls != 2 {
		t.Errorf("API calls = %d, want 2 (GET + PUT)", calls)
	}

	var detail columnDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Description != "New description" {
		t.Errorf("Description = %q, want %q", detail.Description, "New description")
	}
	if !detail.Hidden {
		t.Error("Hidden = false, want true")
	}
}

func TestUpdate_NoFlags(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("unexpected API call")
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "my-dataset", "abc123"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no update flags provided")
	}
}

func TestUpdate_WithFile(t *testing.T) {
	var getCount int
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			getCount++
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			var req map[string]any
			_ = json.Unmarshal(body, &req)
			if req["description"] != "From File" {
				t.Errorf("description = %v, want From File", req["description"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "abc123",
				"key_name":    "duration_ms",
				"type":        "float",
				"description": "From File",
				"hidden":      false,
			})
		}
	}))

	dir := t.TempDir()
	file := filepath.Join(dir, "column.json")
	input := `{"id":"abc123","key_name":"duration_ms","type":"float","description":"From File"}`
	if err := os.WriteFile(file, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "my-dataset", "abc123", "--file", file})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if getCount != 0 {
		t.Errorf("GET count = %d, want 0 (file mode should skip GET)", getCount)
	}

	var detail columnDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Description != "From File" {
		t.Errorf("Description = %q, want %q", detail.Description, "From File")
	}
}

func TestUpdate_WithStdin(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			t.Error("file mode should skip GET")
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		_ = json.Unmarshal(body, &req)
		if req["description"] != "From Stdin" {
			t.Errorf("description = %v, want From Stdin", req["description"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "abc123",
			"key_name":    "duration_ms",
			"type":        "float",
			"description": "From Stdin",
		})
	}))

	ts.InBuf.WriteString(`{"id":"abc123","key_name":"duration_ms","type":"float","description":"From Stdin"}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "my-dataset", "abc123", "--file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail columnDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Description != "From Stdin" {
		t.Errorf("Description = %q, want %q", detail.Description, "From Stdin")
	}
}

func TestUpdate_FileAndFlagsMutuallyExclusive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("API should not be called when flags conflict")
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "my-dataset", "abc123", "--file", "-", "--description", "x"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(err.Error(), "if any flags in the group") {
		t.Errorf("error = %q, want mutual exclusion message", err.Error())
	}
}

func TestUpdate_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"column not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "my-dataset", "bad-id", "--description", "x"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 404")
	}
}
