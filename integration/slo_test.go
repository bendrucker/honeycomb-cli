//go:build integration

package integration

import (
	"fmt"
	"maps"
	"slices"
	"testing"
	"time"
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
			_, _ = runErr(t, nil, "slo", "delete", sloID, "--dataset", dataset, "--yes")
		}
		if dcID != "" {
			_, _ = runErr(t, nil, "column", "calculated", "delete", dcID, "--dataset", dataset, "--yes")
		}
	})

	t.Run("create", func(t *testing.T) {
		body := toJSON(t, map[string]any{
			"name":               name,
			"sli":                map[string]any{"alias": alias},
			"time_period_days":   30,
			"target_per_million": 999000,
		})
		r := run(t, body, "slo", "create", "--dataset", dataset, "-f", "-")
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

	t.Run("update-from-file", func(t *testing.T) {
		updatedName := name + "-file"
		body := toJSON(t, map[string]any{
			"name":               updatedName,
			"sli":                map[string]any{"alias": alias},
			"time_period_days":   30,
			"target_per_million": 995000,
		})
		path := writeTemp(t, body)
		r := run(t, nil, "slo", "update", sloID, "--dataset", dataset, "-f", path)
		slo := parseJSON[map[string]any](t, r.stdout)
		if got := slo["name"]; got != updatedName {
			t.Errorf("expected name %q, got %q", updatedName, got)
		}
		target, _ := slo["target_per_million"].(float64)
		if target != 995000 {
			t.Errorf("expected target_per_million 995000, got %v", target)
		}
	})

	t.Run("delete", func(t *testing.T) {
		throwawayName := uniqueName(t)
		throwawayAlias := throwawayName + "-sli"

		r := run(t, nil, "column", "calculated", "create",
			"--dataset", dataset,
			"--alias", throwawayAlias,
			"--expression", "BOOL(1)",
		)
		throwawayDC := parseJSON[map[string]any](t, r.stdout)
		throwawayDCID, _ := throwawayDC["id"].(string)
		t.Cleanup(func() {
			if throwawayDCID != "" {
				_, _ = runErr(t, nil, "column", "calculated", "delete", throwawayDCID, "--dataset", dataset, "--yes")
			}
		})

		body := toJSON(t, map[string]any{
			"name":               throwawayName,
			"sli":                map[string]any{"alias": throwawayAlias},
			"time_period_days":   30,
			"target_per_million": 999000,
		})
		r = run(t, body, "slo", "create", "--dataset", dataset, "-f", "-")
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
	var recipientID string
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
	sloBody := toJSON(t, map[string]any{
		"name":               name,
		"sli":                map[string]any{"alias": alias},
		"time_period_days":   30,
		"target_per_million": 999000,
	})
	r = run(t, sloBody, "slo", "create", "--dataset", dataset, "-f", "-")
	slo := parseJSON[map[string]any](t, r.stdout)
	sid, ok := slo["id"].(string)
	if !ok || sid == "" {
		t.Fatalf("expected non-empty SLO id: %s", r.stdout)
	}
	sloID = sid

	// Create a recipient for burn alerts
	recipientBody := toJSON(t, map[string]any{
		"type":    "email",
		"details": map[string]any{"email_address": "burn-alert-test@example.com"},
	})
	r = run(t, recipientBody, "recipient", "create", "-f", "-")
	rec := parseJSON[map[string]any](t, r.stdout)
	rid, ok := rec["id"].(string)
	if !ok || rid == "" {
		t.Fatalf("expected non-empty recipient id: %s", r.stdout)
	}
	recipientID = rid

	recipients := []map[string]any{{"id": recipientID}}
	sloRef := map[string]any{"id": sloID}

	t.Cleanup(func() {
		if burnAlertID != "" {
			_, _ = runErr(t, nil, "slo", "burn-alert", "delete", burnAlertID, "--dataset", dataset, "--yes")
		}
		if sloID != "" {
			_, _ = runErr(t, nil, "slo", "delete", sloID, "--dataset", dataset, "--yes")
		}
		if dcID != "" {
			_, _ = runErr(t, nil, "column", "calculated", "delete", dcID, "--dataset", dataset, "--yes")
		}
		if recipientID != "" {
			_, _ = runErr(t, nil, "recipient", "delete", recipientID, "--yes")
		}
	})

	t.Run("create", func(t *testing.T) {
		body := toJSON(t, map[string]any{
			"alert_type":         "exhaustion_time",
			"exhaustion_minutes": 240,
			"slo":                sloRef,
			"recipients":         recipients,
		})
		r := run(t, body, "slo", "burn-alert", "create", "--dataset", dataset, "-f", "-")
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

	t.Run("update", func(t *testing.T) {
		body := toJSON(t, map[string]any{
			"exhaustion_minutes": 120,
			"recipients":         recipients,
		})
		r := run(t, body, "slo", "burn-alert", "update", burnAlertID, "--dataset", dataset, "-f", "-")
		ba := parseJSON[map[string]any](t, r.stdout)
		mins, _ := ba["exhaustion_minutes"].(float64)
		if mins != 120 {
			t.Errorf("expected exhaustion_minutes 120, got %v", mins)
		}
	})

	t.Run("budget-rate", func(t *testing.T) {
		body := toJSON(t, map[string]any{
			"alert_type":                                 "budget_rate",
			"budget_rate_window_minutes":                 60,
			"budget_rate_decrease_threshold_per_million": 50000,
			"slo":        sloRef,
			"recipients": recipients,
		})
		r := run(t, body, "slo", "burn-alert", "create", "--dataset", dataset, "-f", "-")
		ba := parseJSON[map[string]any](t, r.stdout)
		id, ok := ba["id"].(string)
		if !ok || id == "" {
			t.Fatalf("expected non-empty id in response: %s", r.stdout)
		}

		if got := ba["alert_type"]; got != "budget_rate" {
			t.Errorf("expected alert_type %q, got %q", "budget_rate", got)
		}

		t.Cleanup(func() {
			_, _ = runErr(t, nil, "slo", "burn-alert", "delete", id, "--dataset", dataset, "--yes")
		})
	})

	t.Run("delete", func(t *testing.T) {
		throwawayBody := toJSON(t, map[string]any{
			"alert_type":         "exhaustion_time",
			"exhaustion_minutes": 120,
			"slo":                sloRef,
			"recipients":         recipients,
		})
		r := run(t, throwawayBody, "slo", "burn-alert", "create", "--dataset", dataset, "-f", "-")
		throwaway := parseJSON[map[string]any](t, r.stdout)
		throwawayID, ok := throwaway["id"].(string)
		if !ok || throwawayID == "" {
			t.Fatal("expected non-empty id for throwaway burn alert")
		}

		run(t, nil, "slo", "burn-alert", "delete", throwawayID, "--dataset", dataset, "--yes")
	})
}

func TestSLOHistory(t *testing.T) {
	skipWithoutPro(t)

	name := uniqueName(t)
	alias := name + "-sli"
	var sloID string
	var dcID string

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
	sloBody := toJSON(t, map[string]any{
		"name":               name,
		"sli":                map[string]any{"alias": alias},
		"time_period_days":   30,
		"target_per_million": 999000,
	})
	r = run(t, sloBody, "slo", "create", "--dataset", dataset, "-f", "-")
	slo := parseJSON[map[string]any](t, r.stdout)
	sid, ok := slo["id"].(string)
	if !ok || sid == "" {
		t.Fatalf("expected non-empty SLO id: %s", r.stdout)
	}
	sloID = sid

	t.Cleanup(func() {
		if sloID != "" {
			_, _ = runErr(t, nil, "slo", "delete", sloID, "--dataset", dataset, "--yes")
		}
		if dcID != "" {
			_, _ = runErr(t, nil, "column", "calculated", "delete", dcID, "--dataset", dataset, "--yes")
		}
	})

	now := time.Now()
	start := fmt.Sprintf("%d", now.Add(-24*time.Hour).Unix())
	end := fmt.Sprintf("%d", now.Unix())

	r = run(t, nil, "slo", "history",
		"--dataset", dataset,
		"--slo-id", sloID,
		"--start-time", start,
		"--end-time", end,
	)

	history := parseJSON[map[string][]map[string]any](t, r.stdout)
	entries, ok := history[sloID]
	if !ok {
		t.Fatalf("expected history key for SLO %s, got keys: %v", sloID, slices.Collect(maps.Keys(history)))
	}
	for _, entry := range entries {
		if _, ok := entry["compliance"]; !ok {
			t.Error("expected compliance field in history entry")
		}
		if _, ok := entry["budget_remaining"]; !ok {
			t.Error("expected budget_remaining field in history entry")
		}
	}
}
