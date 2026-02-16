//go:build integration

package integration

import (
	"testing"
)

func TestTrigger(t *testing.T) {
	var id string
	name := uniqueName(t)

	triggerBody := toJSON(t, map[string]any{
		"name": name,
		"query": map[string]any{
			"calculations": []map[string]any{{"op": "COUNT"}},
			"time_range":   300,
		},
		"threshold": map[string]any{"op": ">", "value": 100},
		"frequency": 300,
	})

	t.Cleanup(func() {
		if id != "" {
			_, _ = runErr(t, nil, "trigger", "delete", id, "--dataset", dataset, "--yes")
		}
	})

	t.Run("create", func(t *testing.T) {
		path := writeTemp(t, triggerBody)
		r := run(t, nil, "trigger", "create", "--dataset", dataset, "--file", path)
		tr := parseJSON[map[string]any](t, r.stdout)
		v, ok := tr["id"].(string)
		if !ok || v == "" {
			t.Fatalf("expected non-empty id, got %v", tr["id"])
		}
		id = v
	})

	if id == "" {
		t.Fatal("create failed")
	}

	t.Run("list", func(t *testing.T) {
		r := run(t, nil, "trigger", "list", "--dataset", dataset)
		triggers := parseJSON[[]map[string]any](t, r.stdout)
		found := false
		for _, tr := range triggers {
			if tr["id"] == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("trigger %s not found in list", id)
		}
	})

	t.Run("get", func(t *testing.T) {
		r := run(t, nil, "trigger", "get", id, "--dataset", dataset)
		tr := parseJSON[map[string]any](t, r.stdout)
		if tr["id"] != id {
			t.Errorf("expected id %q, got %q", id, tr["id"])
		}
		if tr["name"] != name {
			t.Errorf("expected name %q, got %q", name, tr["name"])
		}
	})

	t.Run("update", func(t *testing.T) {
		updatedName := name + "-upd"
		updatedBody := toJSON(t, map[string]any{
			"name": updatedName,
			"query": map[string]any{
				"calculations": []map[string]any{{"op": "COUNT"}},
				"time_range":   300,
			},
			"threshold": map[string]any{"op": ">", "value": 200},
			"frequency": 300,
		})
		path := writeTemp(t, updatedBody)
		r := run(t, nil, "trigger", "update", id, "--dataset", dataset, "--file", path)
		tr := parseJSON[map[string]any](t, r.stdout)
		if tr["name"] != updatedName {
			t.Errorf("expected name %q, got %q", updatedName, tr["name"])
		}
	})

	t.Run("delete", func(t *testing.T) {
		throwawayName := name + "-del"
		throwawayBody := toJSON(t, map[string]any{
			"name": throwawayName,
			"query": map[string]any{
				"calculations": []map[string]any{{"op": "COUNT"}},
				"time_range":   300,
			},
			"threshold": map[string]any{"op": ">", "value": 100},
			"frequency": 300,
		})
		path := writeTemp(t, throwawayBody)
		r := run(t, nil, "trigger", "create", "--dataset", dataset, "--file", path)
		throwaway := parseJSON[map[string]any](t, r.stdout)
		throwawayID, ok := throwaway["id"].(string)
		if !ok || throwawayID == "" {
			t.Fatalf("expected non-empty id for throwaway trigger")
		}

		run(t, nil, "trigger", "delete", throwawayID, "--dataset", dataset, "--yes")
	})
}
