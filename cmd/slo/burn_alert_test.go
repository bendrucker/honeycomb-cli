package slo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/zalando/go-keyring"
)

func init() {
	keyring.MockInit()
}

func setupBurnAlertTest(t *testing.T, handler http.Handler) (*options.RootOptions, *iostreams.TestStreams) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    srv.URL,
		Format:    "json",
	}

	if err := config.SetKey("default", config.KeyConfig, "test-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })

	return opts, ts
}

func TestBurnAlertList(t *testing.T) {
	opts, ts := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/burn_alerts/my-dataset" {
			t.Errorf("path = %q, want /1/burn_alerts/my-dataset", r.URL.Path)
		}
		if got := r.URL.Query().Get("slo_id"); got != "slo-123" {
			t.Errorf("slo_id = %q, want %q", got, "slo-123")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":         "ba-1",
				"alert_type": "exhaustion_time",
				"slo_id":     "slo-123",
				"created_at": "2024-01-01T00:00:00Z",
			},
			{
				"id":         "ba-2",
				"alert_type": "budget_rate",
				"slo_id":     "slo-123",
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"burn-alert", "list", "--dataset", "my-dataset", "--slo-id", "slo-123"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []burnAlertItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != "ba-1" {
		t.Errorf("items[0].ID = %q, want %q", items[0].ID, "ba-1")
	}
	if items[0].AlertType != "exhaustion_time" {
		t.Errorf("items[0].AlertType = %q, want %q", items[0].AlertType, "exhaustion_time")
	}
	if items[1].AlertType != "budget_rate" {
		t.Errorf("items[1].AlertType = %q, want %q", items[1].AlertType, "budget_rate")
	}
}

func TestBurnAlertList_Empty(t *testing.T) {
	opts, ts := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"burn-alert", "list", "--dataset", "my-dataset", "--slo-id", "slo-123"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []burnAlertItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

func TestBurnAlertGet(t *testing.T) {
	opts, ts := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/burn_alerts/my-dataset/ba-1" {
			t.Errorf("path = %q, want /1/burn_alerts/my-dataset/ba-1", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                 "ba-1",
			"alert_type":         "exhaustion_time",
			"description":        "Alert when budget exhausted",
			"slo_id":             "slo-123",
			"exhaustion_minutes": 240,
			"created_at":         "2024-01-01T00:00:00Z",
			"updated_at":         "2024-01-02T00:00:00Z",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"burn-alert", "get", "ba-1", "--dataset", "my-dataset"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail burnAlertDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "ba-1" {
		t.Errorf("ID = %q, want %q", detail.ID, "ba-1")
	}
	if detail.AlertType != "exhaustion_time" {
		t.Errorf("AlertType = %q, want %q", detail.AlertType, "exhaustion_time")
	}
	if detail.Description != "Alert when budget exhausted" {
		t.Errorf("Description = %q, want %q", detail.Description, "Alert when budget exhausted")
	}
	if detail.ExhaustionMinutes == nil || *detail.ExhaustionMinutes != 240 {
		t.Errorf("ExhaustionMinutes = %v, want 240", detail.ExhaustionMinutes)
	}
}

func TestBurnAlertDelete(t *testing.T) {
	opts, ts := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/1/burn_alerts/my-dataset/ba-1" {
			t.Errorf("path = %q, want /1/burn_alerts/my-dataset/ba-1", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"burn-alert", "delete", "ba-1", "--dataset", "my-dataset", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var result map[string]string
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result["id"] != "ba-1" {
		t.Errorf("id = %q, want %q", result["id"], "ba-1")
	}
}

func TestBurnAlertDelete_RequiresYesNonInteractive(t *testing.T) {
	opts, _ := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("unexpected API call")
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"burn-alert", "delete", "ba-1", "--dataset", "my-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --yes")
	}
	if !strings.Contains(err.Error(), "--yes is required") {
		t.Errorf("error = %q, want --yes required message", err.Error())
	}
}

func TestBurnAlertUpdate_Flags(t *testing.T) {
	for _, tc := range []struct {
		name     string
		args     []string
		getResp  map[string]any
		wantKey  string
		wantVal  any
	}{
		{
			name: "exhaustion minutes",
			args: []string{"burn-alert", "update", "ba-1", "--dataset", "my-dataset", "--exhaustion-minutes", "120"},
			getResp: map[string]any{
				"id":                 "ba-1",
				"alert_type":         "exhaustion_time",
				"exhaustion_minutes": 240,
				"recipients":         []any{map[string]any{"id": "r-1"}},
			},
			wantKey: "exhaustion_minutes",
			wantVal: float64(120),
		},
		{
			name: "budget rate threshold",
			args: []string{"burn-alert", "update", "ba-1", "--dataset", "my-dataset", "--budget-rate-threshold", "50000"},
			getResp: map[string]any{
				"id":         "ba-1",
				"alert_type": "budget_rate",
				"budget_rate_decrease_threshold_per_million": 10000,
				"budget_rate_window_minutes":                 60,
				"recipients":                                 []any{map[string]any{"id": "r-1"}},
			},
			wantKey: "budget_rate_decrease_threshold_per_million",
			wantVal: float64(50000),
		},
		{
			name: "budget rate window",
			args: []string{"burn-alert", "update", "ba-1", "--dataset", "my-dataset", "--budget-rate-window-minutes", "120"},
			getResp: map[string]any{
				"id":         "ba-1",
				"alert_type": "budget_rate",
				"budget_rate_decrease_threshold_per_million": 10000,
				"budget_rate_window_minutes":                 60,
				"recipients":                                 []any{map[string]any{"id": "r-1"}},
			},
			wantKey: "budget_rate_window_minutes",
			wantVal: float64(120),
		},
		{
			name: "description",
			args: []string{"burn-alert", "update", "ba-1", "--dataset", "my-dataset", "--description", "New description"},
			getResp: map[string]any{
				"id":          "ba-1",
				"alert_type":  "exhaustion_time",
				"description": "Old description",
				"recipients":  []any{map[string]any{"id": "r-1"}},
			},
			wantKey: "description",
			wantVal: "New description",
		},
		{
			name: "recipients",
			args: []string{"burn-alert", "update", "ba-1", "--dataset", "my-dataset", "--recipient", "r-2", "--recipient", "r-3"},
			getResp: map[string]any{
				"id":         "ba-1",
				"alert_type": "exhaustion_time",
				"recipients": []any{map[string]any{"id": "r-1"}},
			},
			wantKey: "recipients",
			wantVal: []any{map[string]any{"id": "r-2"}, map[string]any{"id": "r-3"}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotBody map[string]any
			opts, ts := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(tc.getResp)
				case http.MethodPut:
					if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
						t.Fatalf("decoding request body: %v", err)
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(map[string]any{
						"id":         "ba-1",
						"alert_type": tc.getResp["alert_type"],
					})
				default:
					t.Errorf("unexpected method %q", r.Method)
				}
			}))

			cmd := NewCmd(opts)
			cmd.SetArgs(tc.args)
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}

			if gotBody["id"] != nil {
				t.Error("body should not include read-only id field")
			}
			if gotBody["slo_id"] != nil {
				t.Error("body should not include read-only slo_id field")
			}

			gotVal, err := json.Marshal(gotBody[tc.wantKey])
			if err != nil {
				t.Fatalf("marshal got value: %v", err)
			}
			wantVal, err := json.Marshal(tc.wantVal)
			if err != nil {
				t.Fatalf("marshal want value: %v", err)
			}
			if string(gotVal) != string(wantVal) {
				t.Errorf("body[%q] = %s, want %s", tc.wantKey, gotVal, wantVal)
			}

			var detail burnAlertDetail
			if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
				t.Fatalf("unmarshal output: %v", err)
			}
			if detail.ID != "ba-1" {
				t.Errorf("output ID = %q, want %q", detail.ID, "ba-1")
			}
		})
	}
}

func TestBurnAlertUpdate_PreservesRecipients(t *testing.T) {
	var gotBody map[string]any
	opts, _ := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                 "ba-1",
				"alert_type":         "exhaustion_time",
				"exhaustion_minutes": 240,
				"recipients":         []any{map[string]any{"id": "r-1"}, map[string]any{"id": "r-2"}},
			})
		case http.MethodPut:
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":         "ba-1",
				"alert_type": "exhaustion_time",
			})
		}
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"burn-alert", "update", "ba-1", "--dataset", "my-dataset", "--description", "Updated"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	recipients, ok := gotBody["recipients"].([]any)
	if !ok {
		t.Fatal("recipients not preserved in request body")
	}
	if len(recipients) != 2 {
		t.Errorf("got %d recipients, want 2 (should preserve existing)", len(recipients))
	}
}

func TestBurnAlertUpdate_File(t *testing.T) {
	opts, ts := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         "ba-1",
			"alert_type": "exhaustion_time",
		})
	}))

	ts.InBuf.WriteString(`{"alert_type":"exhaustion_time","exhaustion_minutes":120,"recipients":[{"id":"r-1"}]}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"burn-alert", "update", "ba-1", "--dataset", "my-dataset", "-f", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail burnAlertDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "ba-1" {
		t.Errorf("output ID = %q, want %q", detail.ID, "ba-1")
	}
}

func TestBurnAlertUpdate_NoFlags(t *testing.T) {
	opts, _ := setupBurnAlertTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"burn-alert", "update", "ba-1", "--dataset", "my-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for no flags")
	}
	if !strings.Contains(err.Error(), "--file") {
		t.Errorf("error = %q, want message about required flags", err.Error())
	}
}

func TestBurnAlertUpdate_FileMutuallyExclusive(t *testing.T) {
	opts, _ := setupBurnAlertTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"burn-alert", "update", "ba-1", "--dataset", "my-dataset", "-f", "-", "--description", "test"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(err.Error(), "none of the others can be") {
		t.Errorf("error = %q, want mutually exclusive message", err.Error())
	}
}

func TestBurnAlertList_NoKey(t *testing.T) {
	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    "http://localhost",
		Format:    "json",
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"burn-alert", "list", "--dataset", "my-dataset", "--slo-id", "slo-123"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "no config key configured") {
		t.Errorf("error = %q, want missing key message", err.Error())
	}
}
