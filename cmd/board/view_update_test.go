package board

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/bendrucker/honeycomb-cli/internal/api"
)

func TestViewUpdate_NameOnly(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":   "v1",
				"name": "Old Name",
				"filters": []map[string]any{
					{"column": "env", "operation": "=", "value": "prod"},
				},
			})
			return
		}

		var body api.UpdateBoardViewRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "New Name" {
			t.Errorf("name = %q, want %q", body.Name, "New Name")
		}
		if len(body.Filters) != 1 || body.Filters[0].Column != "env" {
			t.Errorf("expected existing filters to be preserved, got %v", body.Filters)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "v1",
			"name": "New Name",
			"filters": []map[string]any{
				{"column": "env", "operation": "=", "value": "prod"},
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "update", "v1", "--board", "board-1", "--name", "New Name"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail viewDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "New Name" {
		t.Errorf("Name = %q, want %q", detail.Name, "New Name")
	}
}

func TestViewUpdate_FilterOnly(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":   "v1",
				"name": "My View",
				"filters": []map[string]any{
					{"column": "env", "operation": "=", "value": "prod"},
				},
			})
			return
		}

		var body api.UpdateBoardViewRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "My View" {
			t.Errorf("name = %q, want %q (should preserve current name)", body.Name, "My View")
		}
		if len(body.Filters) != 1 {
			t.Fatalf("filters count = %d, want 1", len(body.Filters))
		}
		if body.Filters[0].Column != "status" {
			t.Errorf("filter column = %q, want %q", body.Filters[0].Column, "status")
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "v1",
			"name": "My View",
			"filters": []map[string]any{
				{"column": "status", "operation": "=", "value": "error"},
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "update", "v1", "--board", "board-1", "--filter", "status:=:error"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestViewUpdate_NameAndFilter(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":   "v1",
				"name": "Old",
				"filters": []map[string]any{
					{"column": "env", "operation": "=", "value": "prod"},
				},
			})
			return
		}

		var body api.UpdateBoardViewRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "Updated" {
			t.Errorf("name = %q, want %q", body.Name, "Updated")
		}
		if len(body.Filters) != 1 || body.Filters[0].Column != "status" {
			t.Errorf("filters = %v, want single status filter", body.Filters)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "v1",
			"name": "Updated",
			"filters": []map[string]any{
				{"column": "status", "operation": "=", "value": "error"},
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "update", "v1", "--board", "board-1", "--name", "Updated", "--filter", "status:=:error"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestViewUpdate_MissingFlags(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "update", "v1", "--board", "board-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing flags")
	}
	if !strings.Contains(err.Error(), "--file, --name, or --filter is required") {
		t.Errorf("error = %q, want missing flags message", err.Error())
	}
}

func TestViewUpdate_WithFile(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "File View" {
			t.Errorf("name = %v, want %q", body["name"], "File View")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "v1",
			"name": "File View",
		})
	}))

	ts.InBuf.WriteString(`{"name":"File View","filters":[{"column":"env","operation":"=","value":"prod"}]}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "update", "v1", "--board", "board-1", "--file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
