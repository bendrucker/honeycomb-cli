//go:build integration

package integration

import (
	"encoding/json"
	"testing"
)

func keyCreateBody(t *testing.T, name, keyType, envID string) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"data": map[string]any{
			"type": "api-keys",
			"attributes": map[string]any{
				"name":     name,
				"key_type": keyType,
			},
			"relationships": map[string]any{
				"environment": map[string]any{
					"data": map[string]any{"id": envID, "type": "environments"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("marshaling key create body: %v", err)
	}
	return body
}

func keyUpdateBody(t *testing.T, id, name string) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"data": map[string]any{
			"type":       "api-keys",
			"id":         id,
			"attributes": map[string]any{"name": name},
		},
	})
	if err != nil {
		t.Fatalf("marshaling key update body: %v", err)
	}
	return body
}

func TestKey(t *testing.T) {
	var id string
	name := uniqueName(t)

	t.Cleanup(func() {
		if id != "" {
			_, _ = runErr(nil, "key", "delete", id, "--team", team, "--yes")
		}
	})

	t.Run("create", func(t *testing.T) {
		r := run(t, keyCreateBody(t, name, "ingest", environment), "key", "create", "--team", team, "-f", "-")
		key := parseJSON[map[string]any](t, r.stdout)
		v, ok := key["id"].(string)
		if !ok || v == "" {
			t.Fatalf("expected non-empty id in response: %s", r.stdout)
		}
		id = v
		if got := key["name"]; got != name {
			t.Errorf("expected name %q, got %q", name, got)
		}
	})

	if id == "" {
		t.Fatal("create failed, cannot continue")
	}

	t.Run("list", func(t *testing.T) {
		r := run(t, nil, "key", "list", "--team", team)
		keys := parseJSON[[]map[string]any](t, r.stdout)
		found := false
		for _, k := range keys {
			if k["id"] == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("key %s not found in list", id)
		}
	})

	t.Run("get", func(t *testing.T) {
		r := run(t, nil, "key", "get", id, "--team", team)
		key := parseJSON[map[string]any](t, r.stdout)
		if got := key["name"]; got != name {
			t.Errorf("expected name %q, got %q", name, got)
		}
	})

	t.Run("update", func(t *testing.T) {
		newName := name + "-upd"
		r := run(t, keyUpdateBody(t, id, newName), "key", "update", id, "--team", team, "-f", "-")
		key := parseJSON[map[string]any](t, r.stdout)
		if got := key["name"]; got != newName {
			t.Errorf("expected name %q, got %q", newName, got)
		}
		name = newName
	})

	t.Run("delete", func(t *testing.T) {
		throwawayName := uniqueName(t)
		r := run(t, keyCreateBody(t, throwawayName, "ingest", environment), "key", "create", "--team", team, "-f", "-")
		throwaway := parseJSON[map[string]any](t, r.stdout)
		throwawayID, ok := throwaway["id"].(string)
		if !ok || throwawayID == "" {
			t.Fatal("expected non-empty id for throwaway key")
		}

		run(t, nil, "key", "delete", throwawayID, "--team", team, "--yes")
	})
}
