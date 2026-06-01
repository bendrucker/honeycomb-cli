package marker

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
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/1/markers/test-dataset" {
			t.Errorf("path = %q, want /1/markers/test-dataset", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal request: %v", err)
		}
		if req["type"] != "deploy" {
			t.Errorf("type = %v, want deploy", req["type"])
		}
		if req["message"] != "v2.0.0" {
			t.Errorf("message = %v, want v2.0.0", req["message"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         "new123",
			"type":       "deploy",
			"message":    "v2.0.0",
			"start_time": 1700000000,
			"created_at": "2024-01-01T00:00:00Z",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "--type", "deploy", "--message", "v2.0.0", "--start-time", "1700000000"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var item markerItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &item); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if item.ID != "new123" {
		t.Errorf("ID = %q, want %q", item.ID, "new123")
	}
	if item.Type != "deploy" {
		t.Errorf("Type = %q, want %q", item.Type, "deploy")
	}
}

func TestCreate_RequiredFlags_NonInteractive(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing type",
			args:    []string{"create", "--dataset", "test-dataset", "--message", "v2.0.0"},
			wantErr: "--type is required in non-interactive mode",
		},
		{
			name:    "missing message",
			args:    []string{"create", "--dataset", "test-dataset", "--type", "deploy"},
			wantErr: "--message is required in non-interactive mode",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Error("API should not be called when required flags are missing")
			}))
			opts.IOStreams.SetNeverPrompt(true)

			cmd := NewCmd(opts)
			cmd.SetArgs(tc.args)
			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected error for missing required flag")
			}
			if err.Error() != tc.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestCreate_WithFile(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		_ = json.Unmarshal(body, &req)
		if req["type"] != "deploy" {
			t.Errorf("type = %v, want deploy", req["type"])
		}
		if req["message"] != "from file" {
			t.Errorf("message = %v, want from file", req["message"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         "file123",
			"type":       "deploy",
			"message":    "from file",
			"start_time": 1700000000,
			"created_at": "2024-01-01T00:00:00Z",
		})
	}))

	dir := t.TempDir()
	file := filepath.Join(dir, "marker.json")
	if err := os.WriteFile(file, []byte(`{"type":"deploy","message":"from file","start_time":1700000000}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "--file", file})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var item markerItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &item); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if item.ID != "file123" {
		t.Errorf("ID = %q, want %q", item.ID, "file123")
	}
}

func TestCreate_WithStdin(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		_ = json.Unmarshal(body, &req)
		if req["message"] != "from stdin" {
			t.Errorf("message = %v, want from stdin", req["message"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         "stdin123",
			"type":       "deploy",
			"message":    "from stdin",
			"start_time": 1700000000,
			"created_at": "2024-01-01T00:00:00Z",
		})
	}))

	ts.InBuf.WriteString(`{"type":"deploy","message":"from stdin","start_time":1700000000}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "--file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var item markerItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &item); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if item.ID != "stdin123" {
		t.Errorf("ID = %q, want %q", item.ID, "stdin123")
	}
}

func TestCreate_FileAndFlagsMutuallyExclusive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("API should not be called when flags conflict")
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "--file", "-", "--type", "deploy"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(err.Error(), "if any flags in the group") {
		t.Errorf("error = %q, want mutual exclusion message", err.Error())
	}
}

func TestCreate_DefaultStartTime(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal request: %v", err)
		}
		st, ok := req["start_time"].(float64)
		if !ok || st == 0 {
			t.Errorf("start_time should be set to current time, got %v", req["start_time"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         "new456",
			"start_time": int(st),
			"created_at": "2024-01-01T00:00:00Z",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "--type", "deploy", "--message", "v2.0.0"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
