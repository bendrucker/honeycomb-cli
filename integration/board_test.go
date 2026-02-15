//go:build integration

package integration

import "testing"

func TestBoard(t *testing.T) {
	name := uniqueName(t)
	var id string

	t.Cleanup(func() {
		if id != "" {
			_, _ = runErr(nil, "board", "delete", id, "--yes")
		}
	})

	t.Run("create", func(t *testing.T) {
		r := run(t, nil, "board", "create", "--name", name)
		board := parseJSON[struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}](t, r.stdout)
		if board.Name != name {
			t.Errorf("expected name %q, got %q", name, board.Name)
		}
		if board.ID == "" {
			t.Fatalf("expected non-empty id")
		}
		id = board.ID
	})

	if id == "" {
		t.Fatal("create failed")
	}

	t.Run("list", func(t *testing.T) {
		r := run(t, nil, "board", "list")
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
			t.Errorf("expected board %q in list", id)
		}
	})

	t.Run("get", func(t *testing.T) {
		r := run(t, nil, "board", "get", id)
		board := parseJSON[struct {
			Name string `json:"name"`
		}](t, r.stdout)
		if board.Name != name {
			t.Errorf("expected name %q, got %q", name, board.Name)
		}
	})

	t.Run("update", func(t *testing.T) {
		updatedName := "updated-" + name
		r := run(t, nil, "board", "update", id, "--name", updatedName)
		board := parseJSON[struct {
			Name string `json:"name"`
		}](t, r.stdout)
		if board.Name != updatedName {
			t.Errorf("expected name %q, got %q", updatedName, board.Name)
		}
	})

	t.Run("delete", func(t *testing.T) {
		throwawayName := uniqueName(t)
		r := run(t, nil, "board", "create", "--name", throwawayName)
		board := parseJSON[struct {
			ID string `json:"id"`
		}](t, r.stdout)
		if board.ID == "" {
			t.Fatal("expected non-empty id for throwaway board")
		}
		run(t, nil, "board", "delete", board.ID, "--yes")
	})
}

func TestBoardView(t *testing.T) {
	boardName := uniqueName(t)
	r := run(t, nil, "board", "create", "--name", boardName)
	board := parseJSON[struct {
		ID string `json:"id"`
	}](t, r.stdout)
	if board.ID == "" {
		t.Fatal("expected non-empty board id")
	}
	boardID := board.ID

	t.Cleanup(func() {
		_, _ = runErr(nil, "board", "delete", boardID, "--yes")
	})

	viewName := uniqueName(t)
	var viewID string

	t.Cleanup(func() {
		if viewID != "" {
			_, _ = runErr(nil, "board", "view", "delete", viewID, "--board", boardID, "--yes")
		}
	})

	t.Run("create", func(t *testing.T) {
		r := run(t, nil, "board", "view", "create", "--board", boardID, "--name", viewName)
		view := parseJSON[struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}](t, r.stdout)
		if view.Name != viewName {
			t.Errorf("expected name %q, got %q", viewName, view.Name)
		}
		if view.ID == "" {
			t.Fatalf("expected non-empty id")
		}
		viewID = view.ID
	})

	if viewID == "" {
		t.Fatal("create failed")
	}

	t.Run("list", func(t *testing.T) {
		r := run(t, nil, "board", "view", "list", "--board", boardID)
		items := parseJSON[[]struct {
			ID string `json:"id"`
		}](t, r.stdout)
		found := false
		for _, item := range items {
			if item.ID == viewID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected view %q in list", viewID)
		}
	})

	t.Run("get", func(t *testing.T) {
		r := run(t, nil, "board", "view", "get", viewID, "--board", boardID)
		view := parseJSON[struct {
			Name string `json:"name"`
		}](t, r.stdout)
		if view.Name != viewName {
			t.Errorf("expected name %q, got %q", viewName, view.Name)
		}
	})

	t.Run("update", func(t *testing.T) {
		updatedName := "updated-" + viewName
		r := run(t, nil, "board", "view", "update", viewID, "--board", boardID, "--name", updatedName)
		view := parseJSON[struct {
			Name string `json:"name"`
		}](t, r.stdout)
		if view.Name != updatedName {
			t.Errorf("expected name %q, got %q", updatedName, view.Name)
		}
	})

	t.Run("delete", func(t *testing.T) {
		throwawayName := uniqueName(t)
		r := run(t, nil, "board", "view", "create", "--board", boardID, "--name", throwawayName)
		view := parseJSON[struct {
			ID string `json:"id"`
		}](t, r.stdout)
		if view.ID == "" {
			t.Fatal("expected non-empty id for throwaway view")
		}
		run(t, nil, "board", "view", "delete", view.ID, "--board", boardID, "--yes")
	})
}
