package trigger

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdate_WithFlags(t *testing.T) {
	var getCount int
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			getCount++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "trigger-1",
				"name":        "Old Name",
				"description": "Old description",
				"disabled":    false,
				"triggered":   false,
				"alert_type":  "on_change",
				"frequency":   900,
				"threshold":   map[string]any{"op": ">", "value": 100},
			})
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			var parsed map[string]any
			_ = json.Unmarshal(body, &parsed)
			if parsed["name"] != "New Name" {
				t.Errorf("name = %v, want %q", parsed["name"], "New Name")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "trigger-1",
				"name":        "New Name",
				"description": "Old description",
				"disabled":    false,
				"triggered":   false,
				"alert_type":  "on_change",
				"frequency":   900,
				"threshold":   map[string]any{"op": ">", "value": 100},
				"updated_at":  "2025-06-02T12:00:00Z",
			})
		}
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "test-dataset", "trigger-1", "--name", "New Name"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if getCount != 1 {
		t.Errorf("GET count = %d, want 1", getCount)
	}

	var detail triggerDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "New Name" {
		t.Errorf("Name = %q, want %q", detail.Name, "New Name")
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
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":         "trigger-1",
				"name":       "From File",
				"disabled":   false,
				"triggered":  false,
				"alert_type": "on_change",
				"frequency":  600,
				"threshold":  map[string]any{"op": ">=", "value": 50},
				"updated_at": "2025-06-02T12:00:00Z",
			})
		}
	}))

	dir := t.TempDir()
	file := filepath.Join(dir, "trigger.json")
	input := `{"name":"From File","frequency":600,"threshold":{"op":">=","value":50}}`
	if err := os.WriteFile(file, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "test-dataset", "trigger-1", "--file", file})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if getCount != 0 {
		t.Errorf("GET count = %d, want 0 (file mode should skip GET)", getCount)
	}

	var detail triggerDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "From File" {
		t.Errorf("Name = %q, want %q", detail.Name, "From File")
	}
}

func TestUpdate_DisabledFlag(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":       "trigger-1",
				"name":     "Test",
				"disabled": false,
			})
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			var parsed map[string]any
			_ = json.Unmarshal(body, &parsed)
			if parsed["disabled"] != true {
				t.Errorf("disabled = %v, want true", parsed["disabled"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":       "trigger-1",
				"name":     "Test",
				"disabled": true,
			})
		}
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "test-dataset", "trigger-1", "--disabled"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_EnabledFlag(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":       "trigger-1",
				"name":     "Test",
				"disabled": true,
			})
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			var parsed map[string]any
			_ = json.Unmarshal(body, &parsed)
			if parsed["disabled"] != false {
				t.Errorf("disabled = %v, want false", parsed["disabled"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":       "trigger-1",
				"name":     "Test",
				"disabled": false,
			})
		}
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "test-dataset", "trigger-1", "--enabled"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_NoFlags(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "test-dataset", "trigger-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for no flags")
	}
	if !strings.Contains(err.Error(), "provide --file or at least one of") {
		t.Errorf("error = %q, want helpful message", err.Error())
	}
}
