//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestQueryRun(t *testing.T) {
	skipWithoutEnterprise(t)

	queryJSON := `{"calculations":[{"op":"COUNT"}],"time_range":60}`
	r := run(t, []byte(queryJSON), "query", "run", "--dataset", dataset, "-f", "-")

	var result map[string]any
	if err := json.Unmarshal(r.stdout, &result); err != nil {
		t.Fatalf("parsing query run response: %v\nstdout: %s", err, r.stdout)
	}
	if _, ok := result["data"]; !ok {
		t.Errorf("expected response to contain data field, got keys: %v", mapKeys(result))
	}
}

func TestSavedQuery(t *testing.T) {
	skipWithoutEnterprise(t)

	name := uniqueName(t)
	var annotationID string

	// Create a query spec via the api command to get a query ID
	queryJSON := `{"calculations":[{"op":"COUNT"}],"time_range":60}`
	qr := run(t, []byte(queryJSON), "api", "-X", "POST",
		fmt.Sprintf("/1/queries/%s", dataset),
		"--input", "-",
	)
	querySpec := parseJSON[map[string]any](t, qr.stdout)
	qid, ok := querySpec["id"].(string)
	if !ok || qid == "" {
		t.Fatalf("expected non-empty query id: %s", qr.stdout)
	}

	t.Cleanup(func() {
		if annotationID != "" {
			_, _ = runErr(nil, "query", "delete", annotationID, "--dataset", dataset, "--yes")
		}
	})

	t.Run("create", func(t *testing.T) {
		body := fmt.Sprintf(`{"name":"%s","query_id":"%s"}`, name, qid)
		r := run(t, []byte(body), "query", "create", "--dataset", dataset, "-f", "-")
		annotation := parseJSON[map[string]any](t, r.stdout)
		v, ok := annotation["id"].(string)
		if !ok || v == "" {
			t.Fatalf("expected non-empty id in response: %s", r.stdout)
		}
		annotationID = v
	})

	if annotationID == "" {
		t.Fatal("create failed, cannot continue")
	}

	t.Run("list", func(t *testing.T) {
		r := run(t, nil, "query", "list", "--dataset", dataset)
		annotations := parseJSON[[]map[string]any](t, r.stdout)
		found := false
		for _, a := range annotations {
			if a["id"] == annotationID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("annotation %s not found in list", annotationID)
		}
	})

	t.Run("view", func(t *testing.T) {
		r := run(t, nil, "query", "view", annotationID, "--dataset", dataset)
		annotation := parseJSON[map[string]any](t, r.stdout)
		if got := annotation["name"]; got != name {
			t.Errorf("expected name %q, got %q", name, got)
		}
	})

	t.Run("update", func(t *testing.T) {
		updatedName := name + "-upd"
		r := run(t, nil, "query", "update", annotationID, "--dataset", dataset, "--name", updatedName)
		annotation := parseJSON[map[string]any](t, r.stdout)
		if got := annotation["name"]; got != updatedName {
			t.Errorf("expected name %q, got %q", updatedName, got)
		}
	})

	t.Run("delete", func(t *testing.T) {
		// Create a throwaway annotation for delete testing
		body := fmt.Sprintf(`{"name":"%s-del","query_id":"%s"}`, name, qid)
		r := run(t, []byte(body), "query", "create", "--dataset", dataset, "-f", "-")
		throwaway := parseJSON[map[string]any](t, r.stdout)
		throwawayID, ok := throwaway["id"].(string)
		if !ok || throwawayID == "" {
			t.Fatal("expected non-empty id for throwaway annotation")
		}

		run(t, nil, "query", "delete", throwawayID, "--dataset", dataset, "--yes")
	})
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
