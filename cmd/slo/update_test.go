package slo

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestUpdate_Flags(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/slos/test-dataset/slo-1" {
			t.Errorf("path = %q, want /1/slos/test-dataset/slo-1", r.URL.Path)
		}

		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                 "slo-1",
				"name":               "Availability",
				"target_per_million": 999000,
				"time_period_days":   30,
				"sli":                map[string]any{"alias": "sli.availability"},
			})
		case http.MethodPut:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			if body["name"] != "Updated SLO" {
				t.Errorf("body.name = %q, want %q", body["name"], "Updated SLO")
			}
			// target should be preserved from GET
			if body["target_per_million"] != float64(999000) {
				t.Errorf("body.target_per_million = %v, want 999000", body["target_per_million"])
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                 "slo-1",
				"name":               "Updated SLO",
				"target_per_million": 999000,
				"time_period_days":   30,
				"sli":                map[string]any{"alias": "sli.availability"},
			})
		default:
			t.Errorf("unexpected method %q", r.Method)
		}
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "test-dataset", "slo-1", "--name", "Updated SLO"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail sloDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "Updated SLO" {
		t.Errorf("Name = %q, want %q", detail.Name, "Updated SLO")
	}
}

func TestUpdate_File(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                 "slo-1",
			"name":               "From File",
			"target_per_million": 995000,
			"time_period_days":   14,
			"sli":                map[string]any{"alias": "sli.latency"},
		})
	}))

	ts.InBuf.WriteString(`{"name":"From File","target_per_million":995000,"time_period_days":14,"sli":{"alias":"sli.latency"}}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "test-dataset", "slo-1", "--file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail sloDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "From File" {
		t.Errorf("Name = %q, want %q", detail.Name, "From File")
	}
}

func TestUpdate_NoFlags(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "test-dataset", "slo-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for no flags")
	}
	if !strings.Contains(err.Error(), "--file") {
		t.Errorf("error = %q, want message about required flags", err.Error())
	}
}
