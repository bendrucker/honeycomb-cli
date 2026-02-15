//go:build integration

package integration

import (
	"testing"
)

func TestEnvironment(t *testing.T) {
	var id string
	name := uniqueName(t)

	t.Cleanup(func() {
		if id != "" {
			_, _ = runErr(nil, "environment", "delete", id, "--team", team, "--yes")
		}
	})

	t.Run("create", func(t *testing.T) {
		r := run(t, nil, "environment", "create",
			"--team", team,
			"--name", name,
			"--description", "integration test environment",
		)
		env := parseJSON[map[string]any](t, r.stdout)
		v, ok := env["id"].(string)
		if !ok || v == "" {
			t.Fatalf("expected non-empty id in response: %s", r.stdout)
		}
		id = v
		if got := env["name"]; got != name {
			t.Errorf("expected name %q, got %q", name, got)
		}
	})

	if id == "" {
		t.Fatal("create failed, cannot continue")
	}

	t.Run("list", func(t *testing.T) {
		r := run(t, nil, "environment", "list", "--team", team)
		envs := parseJSON[[]map[string]any](t, r.stdout)
		found := false
		for _, e := range envs {
			if e["id"] == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("environment %s not found in list", id)
		}
	})

	t.Run("get", func(t *testing.T) {
		r := run(t, nil, "environment", "get", id, "--team", team)
		env := parseJSON[map[string]any](t, r.stdout)
		if got := env["name"]; got != name {
			t.Errorf("expected name %q, got %q", name, got)
		}
	})

	t.Run("update", func(t *testing.T) {
		r := run(t, nil, "environment", "update", id,
			"--team", team,
			"--description", "updated",
		)
		env := parseJSON[map[string]any](t, r.stdout)
		if got := env["description"]; got != "updated" {
			t.Errorf("expected description %q, got %q", "updated", got)
		}
	})

	t.Run("delete", func(t *testing.T) {
		throwawayName := uniqueName(t)
		r := run(t, nil, "environment", "create",
			"--team", team,
			"--name", throwawayName,
		)
		throwaway := parseJSON[map[string]any](t, r.stdout)
		throwawayID, ok := throwaway["id"].(string)
		if !ok || throwawayID == "" {
			t.Fatal("expected non-empty id for throwaway environment")
		}

		run(t, nil, "environment", "delete", throwawayID, "--team", team, "--yes")
	})
}
