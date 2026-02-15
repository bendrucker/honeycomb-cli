//go:build integration

package integration

import (
	"fmt"
	"testing"
)

func TestSLO(t *testing.T) {
	skipWithoutPro(t)

	name := uniqueName(t)
	alias := name + "-sli"
	var sloID string
	var dcID string

	// Create a derived column for the SLI
	r := run(t, nil, "column", "calculated", "create",
		"--dataset", dataset,
		"--alias", alias,
		"--expression", "BOOL(1)",
	)
	dc := parseJSON[map[string]any](t, r.stdout)
	v, ok := dc["id"].(string)
	if !ok || v == "" {
		t.Fatalf("expected non-empty derived column id: %s", r.stdout)
	}
	dcID = v

	t.Cleanup(func() {
		if sloID != "" {
			_, _ = runErr(nil, "slo", "delete", sloID, "--dataset", dataset, "--yes")
		}
		if dcID != "" {
			_, _ = runErr(nil, "column", "calculated", "delete", dcID, "--dataset", dataset, "--yes")
		}
	})

	t.Run("create", func(t *testing.T) {
		body := fmt.Sprintf(`{"name":"%s","sli":{"alias":"%s"},"time_period_days":30,"target_per_million":999000}`, name, alias)
		r := run(t, []byte(body), "slo", "create", "--dataset", dataset, "-f", "-")
		slo := parseJSON[map[string]any](t, r.stdout)
		id, ok := slo["id"].(string)
		if !ok || id == "" {
			t.Fatalf("expected non-empty id in response: %s", r.stdout)
		}
		sloID = id
	})

	if sloID == "" {
		t.Fatal("create failed, cannot continue")
	}

	t.Run("list", func(t *testing.T) {
		r := run(t, nil, "slo", "list", "--dataset", dataset)
		slos := parseJSON[[]map[string]any](t, r.stdout)
		found := false
		for _, s := range slos {
			if s["id"] == sloID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SLO %s not found in list", sloID)
		}
	})

	t.Run("get", func(t *testing.T) {
		r := run(t, nil, "slo", "get", sloID, "--dataset", dataset)
		slo := parseJSON[map[string]any](t, r.stdout)
		if got := slo["name"]; got != name {
			t.Errorf("expected name %q, got %q", name, got)
		}
	})

	t.Run("update", func(t *testing.T) {
		updatedName := name + "-upd"
		r := run(t, nil, "slo", "update", sloID, "--dataset", dataset, "--name", updatedName)
		slo := parseJSON[map[string]any](t, r.stdout)
		if got := slo["name"]; got != updatedName {
			t.Errorf("expected name %q, got %q", updatedName, got)
		}
	})

	t.Run("delete", func(t *testing.T) {
		throwawayName := uniqueName(t)
		body := fmt.Sprintf(`{"name":"%s","sli":{"alias":"%s"},"time_period_days":30,"target_per_million":999000}`, throwawayName, alias)
		r := run(t, []byte(body), "slo", "create", "--dataset", dataset, "-f", "-")
		throwaway := parseJSON[map[string]any](t, r.stdout)
		throwawayID, ok := throwaway["id"].(string)
		if !ok || throwawayID == "" {
			t.Fatal("expected non-empty id for throwaway SLO")
		}

		run(t, nil, "slo", "delete", throwawayID, "--dataset", dataset, "--yes")
	})
}

func TestBurnAlert(t *testing.T) {
	skipWithoutPro(t)

	name := uniqueName(t)
	alias := name + "-sli"
	var dcID string
	var sloID string
	var burnAlertID string

	// Create derived column for SLI
	r := run(t, nil, "column", "calculated", "create",
		"--dataset", dataset,
		"--alias", alias,
		"--expression", "BOOL(1)",
	)
	dc := parseJSON[map[string]any](t, r.stdout)
	v, ok := dc["id"].(string)
	if !ok || v == "" {
		t.Fatalf("expected non-empty derived column id: %s", r.stdout)
	}
	dcID = v

	// Create SLO
	sloBody := fmt.Sprintf(`{"name":"%s","sli":{"alias":"%s"},"time_period_days":30,"target_per_million":999000}`, name, alias)
	r = run(t, []byte(sloBody), "slo", "create", "--dataset", dataset, "-f", "-")
	slo := parseJSON[map[string]any](t, r.stdout)
	sid, ok := slo["id"].(string)
	if !ok || sid == "" {
		t.Fatalf("expected non-empty SLO id: %s", r.stdout)
	}
	sloID = sid

	t.Cleanup(func() {
		if burnAlertID != "" {
			_, _ = runErr(nil, "slo", "burn-alert", "delete", burnAlertID, "--dataset", dataset, "--yes")
		}
		if sloID != "" {
			_, _ = runErr(nil, "slo", "delete", sloID, "--dataset", dataset, "--yes")
		}
		if dcID != "" {
			_, _ = runErr(nil, "column", "calculated", "delete", dcID, "--dataset", dataset, "--yes")
		}
	})

	t.Run("create", func(t *testing.T) {
		body := fmt.Sprintf(`{"alert_type":"exhaustion_time","exhaustion_minutes":240,"slo_id":"%s"}`, sloID)
		r := run(t, []byte(body), "slo", "burn-alert", "create", "--dataset", dataset, "-f", "-")
		ba := parseJSON[map[string]any](t, r.stdout)
		id, ok := ba["id"].(string)
		if !ok || id == "" {
			t.Fatalf("expected non-empty id in response: %s", r.stdout)
		}
		burnAlertID = id
	})

	if burnAlertID == "" {
		t.Fatal("create failed, cannot continue")
	}

	t.Run("list", func(t *testing.T) {
		r := run(t, nil, "slo", "burn-alert", "list", "--dataset", dataset, "--slo-id", sloID)
		alerts := parseJSON[[]map[string]any](t, r.stdout)
		found := false
		for _, a := range alerts {
			if a["id"] == burnAlertID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("burn alert %s not found in list", burnAlertID)
		}
	})

	t.Run("get", func(t *testing.T) {
		r := run(t, nil, "slo", "burn-alert", "get", burnAlertID, "--dataset", dataset)
		ba := parseJSON[map[string]any](t, r.stdout)
		if got := ba["alert_type"]; got != "exhaustion_time" {
			t.Errorf("expected alert_type %q, got %q", "exhaustion_time", got)
		}
	})

	t.Run("delete", func(t *testing.T) {
		throwawayBody := fmt.Sprintf(`{"alert_type":"exhaustion_time","exhaustion_minutes":120,"slo_id":"%s"}`, sloID)
		r := run(t, []byte(throwawayBody), "slo", "burn-alert", "create", "--dataset", dataset, "-f", "-")
		throwaway := parseJSON[map[string]any](t, r.stdout)
		throwawayID, ok := throwaway["id"].(string)
		if !ok || throwawayID == "" {
			t.Fatal("expected non-empty id for throwaway burn alert")
		}

		run(t, nil, "slo", "burn-alert", "delete", throwawayID, "--dataset", dataset, "--yes")
	})
}
