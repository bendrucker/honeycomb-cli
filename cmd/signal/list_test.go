package signal

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func signalJSON(id, service string) map[string]any {
	return map[string]any{
		"id":                  id,
		"service_name":        service,
		"dataset_slug":        "test-dataset",
		"environment_slug":    "test-env",
		"measured_signal":     "error_rate",
		"status":              "active",
		"enabled":             true,
		"sensitivity":         "medium",
		"currently_anomalous": false,
		"created_at":          "2026-01-15T10:30:00Z",
		"updated_at":          "2026-06-01T12:00:00Z",
	}
}

func TestList(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/signals" {
			t.Errorf("path = %q, want /1/signals", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"signals": []map[string]any{signalJSON("sig-1", "checkout")},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []signalItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}
	if items[0].ID != "sig-1" {
		t.Errorf("ID = %q, want %q", items[0].ID, "sig-1")
	}
	if items[0].ServiceName != "checkout" {
		t.Errorf("ServiceName = %q, want %q", items[0].ServiceName, "checkout")
	}
	if items[0].Status != "active" {
		t.Errorf("Status = %q, want %q", items[0].Status, "active")
	}
}

func TestList_Filters(t *testing.T) {
	for _, tc := range []struct {
		name     string
		args     []string
		expected map[string]string
	}{
		{
			name:     "service",
			args:     []string{"--service", "checkout"},
			expected: map[string]string{"service_name": "checkout"},
		},
		{
			name:     "dataset",
			args:     []string{"--dataset", "prod"},
			expected: map[string]string{"dataset_slug": "prod"},
		},
		{
			name:     "measured signal",
			args:     []string{"--measured-signal", "presence"},
			expected: map[string]string{"measured_signal": "presence"},
		},
		{
			name:     "status",
			args:     []string{"--status", "off"},
			expected: map[string]string{"status": "off"},
		},
		{
			name:     "anomalous",
			args:     []string{"--anomalous"},
			expected: map[string]string{"currently_anomalous": "true"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for param, want := range tc.expected {
					if got := r.URL.Query().Get(param); got != want {
						t.Errorf("query %s = %q, want %q", param, got, want)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{"signals": []map[string]any{}})
			}))

			cmd := NewCmd(opts)
			cmd.SetArgs(append([]string{"list"}, tc.args...))
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestList_InvalidFilter(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list", "--measured-signal", "latency"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid measured signal")
	}
	if !strings.Contains(err.Error(), "measured-signal") {
		t.Errorf("error = %q, want mention of measured-signal", err.Error())
	}
}

func TestList_Paginates(t *testing.T) {
	var cursors []string
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cursor := r.URL.Query().Get("page[after]")
		cursors = append(cursors, cursor)

		page := map[string]any{"signals": []map[string]any{signalJSON("sig-"+cursor, "checkout")}}
		if cursor == "" {
			page["signals"] = []map[string]any{signalJSON("sig-1", "checkout")}
			page["links"] = map[string]any{"next": "https://api.honeycomb.io/1/signals?page%5Bafter%5D=cursor-2"}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(page)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if len(cursors) != 2 {
		t.Fatalf("requests = %d, want 2", len(cursors))
	}
	if cursors[1] != "cursor-2" {
		t.Errorf("second request cursor = %q, want %q", cursors[1], "cursor-2")
	}

	var items []signalItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("items len = %d, want 2", len(items))
	}
}
