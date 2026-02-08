package marker

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestUpdate(t *testing.T) {
	callCount := 0
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++

		if callCount == 1 {
			// List markers (GET)
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{
					"id":         "abc123",
					"type":       "deploy",
					"message":    "v1.0.0",
					"start_time": 1700000000,
					"color":      "#ff0000",
					"created_at": "2024-01-01T00:00:00Z",
				},
			})
			return
		}

		// Update marker (PUT)
		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal request: %v", err)
		}
		if req["message"] != "v2.0.0" {
			t.Errorf("message = %v, want v2.0.0", req["message"])
		}
		// Unchanged fields should be preserved
		if req["type"] != "deploy" {
			t.Errorf("type = %v, want deploy (preserved)", req["type"])
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         "abc123",
			"type":       "deploy",
			"message":    "v2.0.0",
			"start_time": 1700000000,
			"color":      "#ff0000",
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-02T00:00:00Z",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "test-dataset", "abc123", "--message", "v2.0.0"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var item markerItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &item); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if item.Message != "v2.0.0" {
		t.Errorf("Message = %q, want %q", item.Message, "v2.0.0")
	}
	if item.Type != "deploy" {
		t.Errorf("Type = %q, want %q (preserved)", item.Type, "deploy")
	}
}

func TestUpdate_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": "abc123", "type": "deploy"},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "test-dataset", "zzz999", "--message", "new"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for not found")
	}
}
