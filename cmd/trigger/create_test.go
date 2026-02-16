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

func TestCreate(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/triggers/test-dataset" {
			t.Errorf("path = %q, want /1/triggers/test-dataset", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var parsed map[string]any
		if err := json.Unmarshal(body, &parsed); err != nil {
			t.Fatalf("unmarshal request body: %v", err)
		}
		if parsed["name"] != "High Latency" {
			t.Errorf("name = %v, want %q", parsed["name"], "High Latency")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":           "trigger-new",
			"name":         "High Latency",
			"description":  "P99 above 100ms",
			"dataset_slug": "test-dataset",
			"disabled":     false,
			"triggered":    false,
			"alert_type":   "on_change",
			"frequency":    900,
			"threshold":    map[string]any{"op": ">", "value": 100},
			"created_at":   "2025-06-01T12:00:00Z",
			"updated_at":   "2025-06-01T12:00:00Z",
		})
	}))

	dir := t.TempDir()
	file := filepath.Join(dir, "trigger.json")
	input := `{"name":"High Latency","description":"P99 above 100ms","threshold":{"op":">","value":100},"frequency":900}`
	if err := os.WriteFile(file, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "-f", file})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail triggerDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "trigger-new" {
		t.Errorf("ID = %q, want %q", detail.ID, "trigger-new")
	}
	if detail.Name != "High Latency" {
		t.Errorf("Name = %q, want %q", detail.Name, "High Latency")
	}
}

func TestCreate_Stdin(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "trigger-new",
			"name": "From Stdin",
		})
	}))

	ts.InBuf.WriteString(`{"name":"From Stdin","threshold":{"op":">","value":100},"frequency":900}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "-f", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestCreate_NameOverride(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]any
		_ = json.Unmarshal(body, &parsed)
		if parsed["name"] != "Override Name" {
			t.Errorf("name = %v, want %q", parsed["name"], "Override Name")
		}
		if parsed["frequency"] != float64(900) {
			t.Errorf("frequency = %v, want 900", parsed["frequency"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "trigger-new",
			"name": "Override Name",
		})
	}))

	dir := t.TempDir()
	file := filepath.Join(dir, "trigger.json")
	if err := os.WriteFile(file, []byte(`{"name":"File Name","frequency":900}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "-f", file, "--name", "Override Name"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestCreate_DisabledFlag(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]any
		_ = json.Unmarshal(body, &parsed)
		if parsed["disabled"] != true {
			t.Errorf("disabled = %v, want true", parsed["disabled"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":       "trigger-new",
			"name":     "Test",
			"disabled": true,
		})
	}))

	dir := t.TempDir()
	file := filepath.Join(dir, "trigger.json")
	if err := os.WriteFile(file, []byte(`{"name":"Test"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "-f", file, "--disabled"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestCreate_EnabledFlag(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]any
		_ = json.Unmarshal(body, &parsed)
		if parsed["disabled"] != false {
			t.Errorf("disabled = %v, want false", parsed["disabled"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":       "trigger-new",
			"name":     "Test",
			"disabled": false,
		})
	}))

	dir := t.TempDir()
	file := filepath.Join(dir, "trigger.json")
	if err := os.WriteFile(file, []byte(`{"name":"Test","disabled":true}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "-f", file, "--enabled"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestCreate_NoFile(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --file")
	}
	if !strings.Contains(err.Error(), "--file is required") {
		t.Errorf("error = %q, want --file required message", err.Error())
	}
}
