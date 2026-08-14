package signal

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAnomalies(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/signals/sig-1/historical_anomalies" {
			t.Errorf("path = %q, want /1/signals/sig-1/historical_anomalies", r.URL.Path)
		}
		if got := r.URL.Query().Get("start_time"); got != "1767225600" {
			t.Errorf("start_time = %q, want 1767225600", got)
		}
		if got := r.URL.Query().Get("end_time"); got != "1767312000" {
			t.Errorf("end_time = %q, want 1767312000", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"historical_anomalies": []map[string]any{
				{
					"id":           "anom-1",
					"started_at":   1767229200,
					"ended_at":     1767232800,
					"measurement":  0.42,
					"normal_range": map[string]any{"lower": 0.01, "upper": 0.09},
				},
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"anomalies", "sig-1", "--start-time", "1767225600", "--end-time", "1767312000"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []anomalyItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}
	if items[0].ID != "anom-1" {
		t.Errorf("ID = %q, want %q", items[0].ID, "anom-1")
	}
	if items[0].UpperBound != 0.09 {
		t.Errorf("UpperBound = %g, want 0.09", items[0].UpperBound)
	}
}

func TestAnomalies_RequiresTimeRange(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("unexpected request")
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"anomalies", "sig-1"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing time range")
	}
}

func TestFormatEpoch(t *testing.T) {
	for _, tc := range []struct {
		name     string
		seconds  int
		expected string
	}{
		{name: "epoch", seconds: 1767229200, expected: "2026-01-01T01:00:00Z"},
		{name: "unset", seconds: 0, expected: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatEpoch(tc.seconds); got != tc.expected {
				t.Errorf("formatEpoch(%d) = %q, want %q", tc.seconds, got, tc.expected)
			}
		})
	}
}
