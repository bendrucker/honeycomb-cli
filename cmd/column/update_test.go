package column

import (
	"encoding/json"
	"io"
	"net/http"
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
