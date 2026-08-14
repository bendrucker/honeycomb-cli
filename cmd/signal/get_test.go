package signal

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/signals/sig-1" {
			t.Errorf("path = %q, want /1/signals/sig-1", r.URL.Path)
		}
		detail := signalJSON("sig-1", "checkout")
		detail["last_anomaly_started_at"] = 1767225600
		detail["last_anomaly_ended_at"] = nil
		detail["currently_anomalous"] = true
		detail["recipients"] = []map[string]any{
			{"id": "r1", "type": "email", "target": "team@example.com", "details": map[string]any{"muted": true}},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(detail)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "sig-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail signalDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "sig-1" {
		t.Errorf("ID = %q, want %q", detail.ID, "sig-1")
	}
	if detail.Sensitivity != "medium" {
		t.Errorf("Sensitivity = %q, want %q", detail.Sensitivity, "medium")
	}
	if !detail.Anomalous {
		t.Error("Anomalous = false, want true")
	}
	if detail.LastAnomalyStartedAt == nil || *detail.LastAnomalyStartedAt != 1767225600 {
		t.Errorf("LastAnomalyStartedAt = %v, want 1767225600", detail.LastAnomalyStartedAt)
	}
	if detail.LastAnomalyEndedAt != nil {
		t.Errorf("LastAnomalyEndedAt = %v, want nil", *detail.LastAnomalyEndedAt)
	}
	if len(detail.Recipients) != 1 {
		t.Fatalf("Recipients len = %d, want 1", len(detail.Recipients))
	}
	if detail.Recipients[0].Target != "team@example.com" {
		t.Errorf("Recipients[0].Target = %q, want %q", detail.Recipients[0].Target, "team@example.com")
	}
	if !detail.Recipients[0].Muted {
		t.Error("Recipients[0].Muted = false, want true")
	}
}

func TestGet_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "nonexistent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("error = %q, want HTTP 404", err.Error())
	}
}

func TestGet_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing arg")
	}
}
