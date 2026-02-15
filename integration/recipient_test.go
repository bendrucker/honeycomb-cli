//go:build integration

package integration

import (
	"testing"
)

func TestRecipient(t *testing.T) {
	var id string

	createJSON := []byte(`{"type":"email","target":"integration-test@example.com"}`)

	t.Cleanup(func() {
		if id != "" {
			_, _ = runErr(nil, "recipient", "delete", id, "--yes")
		}
	})

	t.Run("create", func(t *testing.T) {
		r := run(t, createJSON, "recipient", "create", "-f", "-")
		rec := parseJSON[map[string]any](t, r.stdout)
		v, ok := rec["id"].(string)
		if !ok || v == "" {
			t.Fatalf("expected non-empty id, got %v", rec["id"])
		}
		id = v
	})

	if id == "" {
		t.Fatal("create failed")
	}

	t.Run("list", func(t *testing.T) {
		r := run(t, nil, "recipient", "list")
		recipients := parseJSON[[]map[string]any](t, r.stdout)
		found := false
		for _, rec := range recipients {
			if rec["id"] == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("recipient %s not found in list", id)
		}
	})

	t.Run("get", func(t *testing.T) {
		r := run(t, nil, "recipient", "get", id)
		rec := parseJSON[map[string]any](t, r.stdout)
		if rec["id"] != id {
			t.Errorf("expected id %q, got %q", id, rec["id"])
		}
	})

	t.Run("update", func(t *testing.T) {
		updateJSON := []byte(`{"type":"email","target":"integration-test-updated@example.com"}`)
		r := run(t, updateJSON, "recipient", "update", id, "-f", "-")
		rec := parseJSON[map[string]any](t, r.stdout)
		if rec["id"] != id {
			t.Errorf("expected id %q, got %q", id, rec["id"])
		}
	})

	t.Run("delete", func(t *testing.T) {
		throwawayJSON := []byte(`{"type":"email","target":"integration-test-throwaway@example.com"}`)
		r := run(t, throwawayJSON, "recipient", "create", "-f", "-")
		throwaway := parseJSON[map[string]any](t, r.stdout)
		throwawayID, ok := throwaway["id"].(string)
		if !ok || throwawayID == "" {
			t.Fatalf("expected non-empty id for throwaway recipient")
		}

		run(t, nil, "recipient", "delete", throwawayID, "--yes")
	})
}
