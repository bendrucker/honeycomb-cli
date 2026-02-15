//go:build integration

package integration

import "testing"

func TestDataset(t *testing.T) {
	name := uniqueName(t)
	var slug string

	t.Cleanup(func() {
		if slug != "" {
			_, _ = runErr(nil, "dataset", "update", slug, "--delete-protected=false")
			_, _ = runErr(nil, "dataset", "delete", slug, "--yes")
		}
	})

	t.Run("create", func(t *testing.T) {
		r := run(t, nil, "dataset", "create", "--name", name)
		ds := parseJSON[struct {
			Slug string `json:"slug"`
			Name string `json:"name"`
		}](t, r.stdout)
		if ds.Name != name {
			t.Errorf("expected name %q, got %q", name, ds.Name)
		}
		if ds.Slug == "" {
			t.Fatalf("expected non-empty slug")
		}
		slug = ds.Slug
	})

	if slug == "" {
		t.Fatal("create failed")
	}

	t.Run("list", func(t *testing.T) {
		r := run(t, nil, "dataset", "list")
		items := parseJSON[[]struct {
			Slug string `json:"slug"`
		}](t, r.stdout)
		found := false
		for _, item := range items {
			if item.Slug == slug {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected dataset %q in list", slug)
		}
	})

	t.Run("get", func(t *testing.T) {
		r := run(t, nil, "dataset", "get", slug)
		ds := parseJSON[struct {
			Name string `json:"name"`
		}](t, r.stdout)
		if ds.Name != name {
			t.Errorf("expected name %q, got %q", name, ds.Name)
		}
	})

	t.Run("update", func(t *testing.T) {
		r := run(t, nil, "dataset", "update", slug, "--description", "updated")
		ds := parseJSON[struct {
			Description string `json:"description"`
		}](t, r.stdout)
		if ds.Description != "updated" {
			t.Errorf("expected description %q, got %q", "updated", ds.Description)
		}
	})

	t.Run("delete", func(t *testing.T) {
		throwawayName := uniqueName(t)
		r := run(t, nil, "dataset", "create", "--name", throwawayName)
		ds := parseJSON[struct {
			Slug string `json:"slug"`
		}](t, r.stdout)
		if ds.Slug == "" {
			t.Fatal("expected non-empty slug for throwaway dataset")
		}
		run(t, nil, "dataset", "update", ds.Slug, "--delete-protected=false")
		run(t, nil, "dataset", "delete", ds.Slug, "--yes")
	})
}
