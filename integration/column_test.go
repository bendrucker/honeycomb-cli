//go:build integration

package integration

import "testing"

func TestColumn(t *testing.T) {
	keyName := uniqueName(t)
	var id string

	t.Cleanup(func() {
		if id != "" {
			_, _ = runErr(nil, "column", "delete", id, "--dataset", dataset, "--yes")
		}
	})

	t.Run("create", func(t *testing.T) {
		r := run(t, nil, "column", "create", "--dataset", dataset, "--key-name", keyName, "--type", "string")
		col := parseJSON[struct {
			ID      string `json:"id"`
			KeyName string `json:"key_name"`
			Type    string `json:"type"`
		}](t, r.stdout)
		if col.KeyName != keyName {
			t.Errorf("expected key_name %q, got %q", keyName, col.KeyName)
		}
		if col.ID == "" {
			t.Fatalf("expected non-empty id")
		}
		id = col.ID
	})

	if id == "" {
		t.Fatal("create failed")
	}

	t.Run("list", func(t *testing.T) {
		r := run(t, nil, "column", "list", "--dataset", dataset)
		items := parseJSON[[]struct {
			ID string `json:"id"`
		}](t, r.stdout)
		found := false
		for _, item := range items {
			if item.ID == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected column %q in list", id)
		}
	})

	t.Run("get", func(t *testing.T) {
		r := run(t, nil, "column", "get", id, "--dataset", dataset)
		col := parseJSON[struct {
			KeyName string `json:"key_name"`
		}](t, r.stdout)
		if col.KeyName != keyName {
			t.Errorf("expected key_name %q, got %q", keyName, col.KeyName)
		}
	})

	t.Run("update", func(t *testing.T) {
		r := run(t, nil, "column", "update", id, "--dataset", dataset, "--description", "updated")
		col := parseJSON[struct {
			Description string `json:"description"`
		}](t, r.stdout)
		if col.Description != "updated" {
			t.Errorf("expected description %q, got %q", "updated", col.Description)
		}
	})

	t.Run("delete", func(t *testing.T) {
		throwawayName := uniqueName(t)
		r := run(t, nil, "column", "create", "--dataset", dataset, "--key-name", throwawayName, "--type", "string")
		col := parseJSON[struct {
			ID string `json:"id"`
		}](t, r.stdout)
		if col.ID == "" {
			t.Fatal("expected non-empty id for throwaway column")
		}
		run(t, nil, "column", "delete", col.ID, "--dataset", dataset, "--yes")
	})
}

func TestCalculatedColumn(t *testing.T) {
	alias := uniqueName(t)
	var id string

	t.Cleanup(func() {
		if id != "" {
			_, _ = runErr(nil, "column", "calculated", "delete", id, "--dataset", dataset, "--yes")
		}
	})

	t.Run("create", func(t *testing.T) {
		r := run(t, nil, "column", "calculated", "create", "--dataset", dataset, "--alias", alias, "--expression", "BOOL(1)")
		col := parseJSON[struct {
			ID    string `json:"id"`
			Alias string `json:"alias"`
		}](t, r.stdout)
		if col.Alias != alias {
			t.Errorf("expected alias %q, got %q", alias, col.Alias)
		}
		if col.ID == "" {
			t.Fatalf("expected non-empty id")
		}
		id = col.ID
	})

	if id == "" {
		t.Fatal("create failed")
	}

	t.Run("list", func(t *testing.T) {
		r := run(t, nil, "column", "calculated", "list", "--dataset", dataset)
		items := parseJSON[[]struct {
			ID string `json:"id"`
		}](t, r.stdout)
		found := false
		for _, item := range items {
			if item.ID == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected calculated column %q in list", id)
		}
	})

	t.Run("get", func(t *testing.T) {
		r := run(t, nil, "column", "calculated", "get", id, "--dataset", dataset)
		col := parseJSON[struct {
			Alias string `json:"alias"`
		}](t, r.stdout)
		if col.Alias != alias {
			t.Errorf("expected alias %q, got %q", alias, col.Alias)
		}
	})

	t.Run("update", func(t *testing.T) {
		r := run(t, nil, "column", "calculated", "update", id, "--dataset", dataset, "--description", "updated")
		col := parseJSON[struct {
			Description string `json:"description"`
		}](t, r.stdout)
		if col.Description != "updated" {
			t.Errorf("expected description %q, got %q", "updated", col.Description)
		}
	})

	t.Run("delete", func(t *testing.T) {
		throwawayAlias := uniqueName(t)
		r := run(t, nil, "column", "calculated", "create", "--dataset", dataset, "--alias", throwawayAlias, "--expression", "BOOL(1)")
		col := parseJSON[struct {
			ID string `json:"id"`
		}](t, r.stdout)
		if col.ID == "" {
			t.Fatal("expected non-empty id for throwaway calculated column")
		}
		run(t, nil, "column", "calculated", "delete", col.ID, "--dataset", dataset, "--yes")
	})
}
