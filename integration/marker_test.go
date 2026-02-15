//go:build integration

package integration

import (
	"fmt"
	"testing"
)

func TestMarker(t *testing.T) {
	var id string

	t.Cleanup(func() {
		if id != "" {
			_, _ = runErr(nil, "marker", "delete", id, "--dataset", dataset, "--yes")
		}
	})

	t.Run("create", func(t *testing.T) {
		r := run(t, nil, "marker", "create", "--dataset", dataset, "--type", "deploy", "--message", "test")
		m := parseJSON[map[string]any](t, r.stdout)
		v, ok := m["id"].(string)
		if !ok || v == "" {
			t.Fatalf("expected non-empty id, got %v", m["id"])
		}
		id = v
	})

	if id == "" {
		t.Fatal("create failed")
	}

	t.Run("list", func(t *testing.T) {
		r := run(t, nil, "marker", "list", "--dataset", dataset)
		markers := parseJSON[[]map[string]any](t, r.stdout)
		found := false
		for _, m := range markers {
			if m["id"] == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("marker %s not found in list", id)
		}
	})

	t.Run("update", func(t *testing.T) {
		r := run(t, nil, "marker", "update", id, "--dataset", dataset, "--message", "updated")
		m := parseJSON[map[string]any](t, r.stdout)
		if m["message"] != "updated" {
			t.Errorf("expected message %q, got %q", "updated", m["message"])
		}
	})

	t.Run("delete", func(t *testing.T) {
		// Create a throwaway marker to delete
		r := run(t, nil, "marker", "create", "--dataset", dataset, "--type", "deploy", "--message", "throwaway")
		throwaway := parseJSON[map[string]any](t, r.stdout)
		throwawayID, ok := throwaway["id"].(string)
		if !ok || throwawayID == "" {
			t.Fatalf("expected non-empty id for throwaway marker")
		}

		run(t, nil, "marker", "delete", throwawayID, "--dataset", dataset, "--yes")
	})
}

func TestMarkerSetting(t *testing.T) {
	var id string
	settingType := uniqueName(t)

	t.Cleanup(func() {
		if id != "" {
			_, _ = runErr(nil, "marker", "setting", "delete", id, "--dataset", dataset, "--yes")
		}
	})

	t.Run("create", func(t *testing.T) {
		r := run(t, nil, "marker", "setting", "create", "--dataset", dataset, "--type", settingType, "--color", "#FF0000")
		s := parseJSON[map[string]any](t, r.stdout)
		v, ok := s["id"].(string)
		if !ok || v == "" {
			t.Fatalf("expected non-empty id, got %v", s["id"])
		}
		id = v
	})

	if id == "" {
		t.Fatal("create failed")
	}

	t.Run("list", func(t *testing.T) {
		r := run(t, nil, "marker", "setting", "list", "--dataset", dataset)
		settings := parseJSON[[]map[string]any](t, r.stdout)
		found := false
		for _, s := range settings {
			if s["id"] == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("marker setting %s not found in list", id)
		}
	})

	t.Run("update", func(t *testing.T) {
		r := run(t, nil, "marker", "setting", "update", id, "--dataset", dataset, "--color", "#00FF00")
		s := parseJSON[map[string]any](t, r.stdout)
		if s["color"] != "#00FF00" {
			t.Errorf("expected color %q, got %q", "#00FF00", s["color"])
		}
	})

	t.Run("delete", func(t *testing.T) {
		throwawayType := fmt.Sprintf("%s-del", settingType)
		if len(throwawayType) > 40 {
			throwawayType = throwawayType[:40]
		}
		r := run(t, nil, "marker", "setting", "create", "--dataset", dataset, "--type", throwawayType, "--color", "#AABBCC")
		throwaway := parseJSON[map[string]any](t, r.stdout)
		throwawayID, ok := throwaway["id"].(string)
		if !ok || throwawayID == "" {
			t.Fatalf("expected non-empty id for throwaway marker setting")
		}

		run(t, nil, "marker", "setting", "delete", throwawayID, "--dataset", dataset, "--yes")
	})
}
