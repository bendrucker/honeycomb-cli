package slo

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/slos/test-dataset/slo-1" {
			t.Errorf("path = %q, want /1/slos/test-dataset/slo-1", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "slo-1",
			"name": "Availability",
			"description": "Service availability SLO",
			"target_per_million": 999000,
			"time_period_days": 30,
			"sli": {"alias": "sli.availability"},
			"created_at": "2025-01-15T10:30:00Z",
			"updated_at": "2025-02-01T12:00:00Z"
		}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "--dataset", "test-dataset", "slo-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail sloDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "slo-1" {
		t.Errorf("ID = %q, want %q", detail.ID, "slo-1")
	}
	if detail.Name != "Availability" {
		t.Errorf("Name = %q, want %q", detail.Name, "Availability")
	}
	if detail.TargetPerMillion != 999000 {
		t.Errorf("TargetPerMillion = %d, want 999000", detail.TargetPerMillion)
	}
	if detail.SLIAlias != "sli.availability" {
		t.Errorf("SLIAlias = %q, want %q", detail.SLIAlias, "sli.availability")
	}
}

func TestGet_ViewAlias(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "slo-1",
			"name": "Availability",
			"target_per_million": 999000,
			"time_period_days": 30,
			"sli": {"alias": "sli.availability"}
		}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "--dataset", "test-dataset", "slo-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestGet_Detailed(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("detailed") != "true" {
			t.Errorf("detailed param = %q, want %q", r.URL.Query().Get("detailed"), "true")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "slo-1",
			"name": "Availability",
			"target_per_million": 999000,
			"time_period_days": 30,
			"sli": {"alias": "sli.availability"},
			"compliance": 99.95,
			"budget_remaining": 0.05
		}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "--dataset", "test-dataset", "--detailed", "slo-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail sloDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Compliance == nil || *detail.Compliance != 99.95 {
		t.Errorf("Compliance = %v, want 99.95", detail.Compliance)
	}
	if detail.BudgetRemaining == nil || *detail.BudgetRemaining != 0.05 {
		t.Errorf("BudgetRemaining = %v, want 0.05", detail.BudgetRemaining)
	}
}

func TestGet_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"SLO not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "--dataset", "test-dataset", "nonexistent"})
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
	cmd.SetArgs([]string{"get", "--dataset", "test-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}
