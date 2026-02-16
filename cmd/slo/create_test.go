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

	ts.InBuf.WriteString(`{"name":"Availability","target_per_million":999000,"time_period_days":30,"sli":{"alias":"sli.availability"}}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "-f", "-"})
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

func TestCreate_Flags(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["name"] != "Availability" {
			t.Errorf("name = %v, want %q", body["name"], "Availability")
		}
		sli, ok := body["sli"].(map[string]any)
		if !ok {
			t.Fatal("sli is not a map")
		}
		if sli["alias"] != "sli.availability" {
			t.Errorf("sli.alias = %v, want %q", sli["alias"], "sli.availability")
		}
		if body["target_per_million"] != float64(999000) {
			t.Errorf("target_per_million = %v, want 999000", body["target_per_million"])
		}
		if body["time_period_days"] != float64(30) {
			t.Errorf("time_period_days = %v, want 30", body["time_period_days"])
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

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset",
		"--name", "Availability",
		"--sli-alias", "sli.availability",
		"--target", "999000",
		"--time-period", "30",
	})
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
}

func TestCreate_FlagsWithDescription(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["description"] != "Service availability target" {
			t.Errorf("description = %v, want %q", body["description"], "Service availability target")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                 "slo-new",
			"name":               "Availability",
			"description":        "Service availability target",
			"target_per_million": 999000,
			"time_period_days":   30,
			"sli":                map[string]any{"alias": "sli.availability"},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset",
		"--name", "Availability",
		"--sli-alias", "sli.availability",
		"--target", "999000",
		"--time-period", "30",
		"--description", "Service availability target",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestCreate_MissingRequiredFlags(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{
			name: "no flags",
			args: []string{"create", "--dataset", "test-dataset"},
		},
		{
			name: "missing sli-alias",
			args: []string{"create", "--dataset", "test-dataset",
				"--name", "Avail", "--target", "999000", "--time-period", "30"},
		},
		{
			name: "missing target",
			args: []string{"create", "--dataset", "test-dataset",
				"--name", "Avail", "--sli-alias", "sli.avail", "--time-period", "30"},
		},
		{
			name: "missing time-period",
			args: []string{"create", "--dataset", "test-dataset",
				"--name", "Avail", "--sli-alias", "sli.avail", "--target", "999000"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
			cmd := NewCmd(opts)
			cmd.SetArgs(tc.args)
			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected error for missing flags")
			}
			if !strings.Contains(err.Error(), "--file or all of") {
				t.Errorf("error = %q, want missing flags message", err.Error())
			}
		})
	}
}
