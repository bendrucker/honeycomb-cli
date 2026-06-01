package slo

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/bendrucker/honeycomb-cli/internal/api"
)

func TestHistory(t *testing.T) {
	opts, ts := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/1/reporting/slos/historical" {
			t.Errorf("path = %q, want /1/reporting/slos/historical", r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading request body: %v", err)
		}
		var req api.SLOHistoryRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshalling request body: %v", err)
		}
		if req.StartTime != 1700000000 {
			t.Errorf("start_time = %d, want 1700000000", req.StartTime)
		}
		if req.EndTime != 1700100000 {
			t.Errorf("end_time = %d, want 1700100000", req.EndTime)
		}
		if len(req.Ids) != 2 {
			t.Fatalf("ids length = %d, want 2", len(req.Ids))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"slo-1": []map[string]any{
				{"timestamp": 1700000000, "compliance": 99.5, "budget_remaining": 0.8},
				{"timestamp": 1700050000, "compliance": 99.2, "budget_remaining": 0.6},
			},
			"slo-2": []map[string]any{
				{"timestamp": 1700000000, "compliance": 98.0, "budget_remaining": 0.3},
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{
		"history",
		"slo-1",
		"--dataset", "ignored",
		"--slo-id", "slo-2",
		"--start-time", "1700000000",
		"--end-time", "1700100000",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var result api.SLOHistoryResponse
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("got %d SLOs, want 2", len(result))
	}
	if len(result["slo-1"]) != 2 {
		t.Errorf("slo-1 entries = %d, want 2", len(result["slo-1"]))
	}
	if len(result["slo-2"]) != 1 {
		t.Errorf("slo-2 entries = %d, want 1", len(result["slo-2"]))
	}
	if result["slo-1"][0].Compliance == nil || *result["slo-1"][0].Compliance != 99.5 {
		t.Errorf("slo-1[0].Compliance = %v, want 99.5", result["slo-1"][0].Compliance)
	}
}

func TestHistory_MergeSLOIDs(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		flagIDs []string
		want    []string
	}{
		{name: "positional only", args: []string{"a", "b"}, want: []string{"a", "b"}},
		{name: "flags only", flagIDs: []string{"a", "b"}, want: []string{"a", "b"}},
		{name: "merged", args: []string{"a"}, flagIDs: []string{"b"}, want: []string{"a", "b"}},
		{name: "dedup across sources", args: []string{"a", "b"}, flagIDs: []string{"b", "c"}, want: []string{"a", "b", "c"}},
		{name: "empty", want: nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeSLOIDs(tc.args, tc.flagIDs)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			}
		})
	}
}

func TestHistory_PositionalOnly(t *testing.T) {
	opts, ts := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading request body: %v", err)
		}
		var req api.SLOHistoryRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshalling request body: %v", err)
		}
		if len(req.Ids) != 1 {
			t.Fatalf("ids length = %d, want 1", len(req.Ids))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"slo-1": []map[string]any{{"timestamp": 1700000000, "compliance": 99.5, "budget_remaining": 0.8}},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{
		"history",
		"slo-1",
		"--dataset", "ignored",
		"--start-time", "1700000000",
		"--end-time", "1700100000",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var result api.SLOHistoryResponse
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("got %d SLOs, want 1", len(result))
	}
}

func TestHistory_NoSLOID(t *testing.T) {
	opts, _ := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("unexpected API call")
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{
		"history",
		"--dataset", "ds",
		"--start-time", "1700000000",
		"--end-time", "1700100000",
	})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no SLO ID is provided")
	}
}

func TestHistory_MissingFlags(t *testing.T) {
	opts, _ := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("unexpected API call")
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"history", "--dataset", "ds"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing required flags")
	}
}

func TestHistory_Unauthorized(t *testing.T) {
	opts, _ := setupBurnAlertTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unknown API key - check your credentials"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{
		"history",
		"--dataset", "ds",
		"--slo-id", "slo-1",
		"--start-time", "1700000000",
		"--end-time", "1700100000",
	})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 401")
	}
}
