//go:build integration

package integration

import (
	"fmt"
	"strings"
	"testing"
)

func TestAPI(t *testing.T) {
	t.Run("GET v1", func(t *testing.T) {
		r := run(t, nil, "api", "/1/datasets")
		datasets := parseJSON[[]map[string]any](t, r.stdout)
		if len(datasets) == 0 {
			t.Error("expected at least one dataset")
		}
	})

	t.Run("GET v2 management key", func(t *testing.T) {
		r := run(t, nil, "api", fmt.Sprintf("/2/teams/%s/environments", team), "--key-type", "management")
		_ = parseJSON[[]map[string]any](t, r.stdout)
	})

	t.Run("jq filter", func(t *testing.T) {
		r := run(t, nil, "api", "/1/datasets", "--jq", ".[0].name")
		output := strings.TrimSpace(string(r.stdout))
		if output == "" {
			t.Error("expected non-empty jq output")
		}
	})

	t.Run("paginate", func(t *testing.T) {
		r := run(t, nil, "api", fmt.Sprintf("/1/columns/%s", dataset), "--paginate")
		_ = parseJSON[[]map[string]any](t, r.stdout)
	})

	t.Run("POST with fields", func(t *testing.T) {
		r := run(t, nil, "api", "-X", "POST", fmt.Sprintf("/1/markers/%s", dataset),
			"-f", "type=api-test",
			"-f", "message=test",
		)
		marker := parseJSON[map[string]any](t, r.stdout)
		markerID, ok := marker["id"].(string)
		if !ok || markerID == "" {
			t.Fatalf("expected non-empty marker id: %s", r.stdout)
		}

		t.Cleanup(func() {
			_, _ = runErr(nil, "api", "-X", "DELETE", fmt.Sprintf("/1/markers/%s/%s", dataset, markerID))
		})
	})
}
