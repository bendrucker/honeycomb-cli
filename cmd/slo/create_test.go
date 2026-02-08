package slo

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCreate_File(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/slos/test-dataset" {
			t.Errorf("path = %q, want /1/slos/test-dataset", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["name"] != "Availability" {
			t.Errorf("body.name = %q, want %q", body["name"], "Availability")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                 "slo-new",
			"name":               "Availability",
			"target_per_million": 999000,
			"time_period_days":   30,
			"sli":                map[string]any{"alias": "sli.availability"},
		})
	}))

	ts.InBuf.WriteString("")
	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "--file", "-"})
	ts.InBuf.Reset()
	ts.InBuf.WriteString(`{"name":"Availability","target_per_million":999000,"time_period_days":30,"sli":{"alias":"sli.availability"}}`)

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail sloDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "slo-new" {
		t.Errorf("ID = %q, want %q", detail.ID, "slo-new")
	}
	if detail.Name != "Availability" {
		t.Errorf("Name = %q, want %q", detail.Name, "Availability")
	}
}

func TestCreate_NoFile_NonInteractive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --file")
	}
	if !strings.Contains(err.Error(), "--file is required") {
		t.Errorf("error = %q, want --file required message", err.Error())
	}
}
