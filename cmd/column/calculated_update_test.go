package column

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCalculatedUpdate(t *testing.T) {
	getCalls := 0
	putCalls := 0
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			getCalls++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "dc-1",
				"alias":       "req.duration_ms",
				"expression":  "SUB($duration_ms, $overhead_ms)",
				"description": "New description",
				"created_at":  "2024-06-01T00:00:00Z",
				"updated_at":  "2025-01-15T10:30:00Z",
			})
		case http.MethodPut:
			putCalls++
			// The PUT response omits created_at/updated_at, mirroring the real API.
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "dc-1",
				"alias":       "req.duration_ms",
				"expression":  "SUB($duration_ms, $overhead_ms)",
				"description": "New description",
			})
		}
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"calculated", "update", "--dataset", "my-dataset", "dc-1", "--description", "New description"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if putCalls != 1 {
		t.Errorf("PUT calls = %d, want 1", putCalls)
	}
	// One GET to read the current column, one GET to re-fetch after the update.
	if getCalls != 2 {
		t.Errorf("GET calls = %d, want 2", getCalls)
	}

	var detail calculatedDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Description != "New description" {
		t.Errorf("Description = %q, want %q", detail.Description, "New description")
	}
	if detail.CreatedAt != "2024-06-01T00:00:00Z" {
		t.Errorf("CreatedAt = %q, want %q (not zero time)", detail.CreatedAt, "2024-06-01T00:00:00Z")
	}
	if detail.UpdatedAt != "2025-01-15T10:30:00Z" {
		t.Errorf("UpdatedAt = %q, want %q (not zero time)", detail.UpdatedAt, "2025-01-15T10:30:00Z")
	}
}

func TestCalculatedUpdate_NoFlags(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"calculated", "update", "--dataset", "my-dataset", "dc-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no flags provided")
	}
}
