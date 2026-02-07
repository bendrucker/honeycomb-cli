package column

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestCreate(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/columns/my-dataset" {
			t.Errorf("path = %q, want /1/columns/my-dataset", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		_ = json.Unmarshal(body, &req)
		if req["key_name"] != "duration_ms" {
			t.Errorf("key_name = %v, want duration_ms", req["key_name"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "abc123",
			"key_name":    "duration_ms",
			"type":        "float",
			"description": "Request duration",
			"hidden":      false,
			"created_at":  "2025-01-15T10:30:00Z",
			"updated_at":  "2025-01-15T10:30:00Z",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "my-dataset", "--key-name", "duration_ms", "--type", "float", "--description", "Request duration"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail columnDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.KeyName != "duration_ms" {
		t.Errorf("KeyName = %q, want %q", detail.KeyName, "duration_ms")
	}
	if detail.ID != "abc123" {
		t.Errorf("ID = %q, want %q", detail.ID, "abc123")
	}
}

func TestCreate_MissingKeyName_NonInteractive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	opts.IOStreams.SetNeverPrompt(true)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "my-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing key-name")
	}
	if err.Error() != "--key-name is required in non-interactive mode" {
		t.Errorf("error = %q, want missing key-name message", err.Error())
	}
}
