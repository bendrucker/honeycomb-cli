//go:build integration

package integration

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// Signals are created by Honeycomb as services report data, so the CLI can only
// read and update them. A fresh test environment usually has none.
func TestSignal(t *testing.T) {
	r, err := runErr(t, nil, "signal", "list")
	if err != nil {
		if strings.Contains(err.Error(), "HTTP 403") || strings.Contains(err.Error(), "HTTP 404") {
			t.Skipf("anomaly detection unavailable for this team: %v", err)
		}
		t.Fatalf("listing signals: %v\nstderr: %s", err, r.stderr)
	}

	signals := parseJSON[[]map[string]any](t, r.stdout)
	if len(signals) == 0 {
		t.Skip("no signals in this environment")
	}

	id, ok := signals[0]["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty id, got %v", signals[0]["id"])
	}

	t.Run("get", func(t *testing.T) {
		r := run(t, nil, "signal", "get", id)
		signal := parseJSON[map[string]any](t, r.stdout)
		if signal["id"] != id {
			t.Errorf("expected id %q, got %q", id, signal["id"])
		}
	})

	t.Run("anomalies", func(t *testing.T) {
		end := time.Now().Unix()
		start := end - int64(24*time.Hour/time.Second)
		run(t, nil, "signal", "anomalies", id,
			"--start-time", strconv.FormatInt(start, 10),
			"--end-time", strconv.FormatInt(end, 10))
	})
}
