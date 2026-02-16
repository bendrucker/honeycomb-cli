package board

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/bendrucker/honeycomb-cli/internal/api"
)

func TestViewCreate_WithFlags(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/boards/board-1/views" {
			t.Errorf("path = %q, want /1/boards/board-1/views", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}

		var body api.CreateBoardViewRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "Errors" {
			t.Errorf("name = %q, want %q", body.Name, "Errors")
		}
		if len(body.Filters) != 1 {
			t.Fatalf("filters count = %d, want 1", len(body.Filters))
		}
		if body.Filters[0].Column != "status" {
			t.Errorf("filter column = %q, want %q", body.Filters[0].Column, "status")
		}
		if body.Filters[0].Operation != "=" {
			t.Errorf("filter operation = %q, want %q", body.Filters[0].Operation, "=")
		}
		if body.Filters[0].Value != "error" {
			t.Errorf("filter value = %v, want %q", body.Filters[0].Value, "error")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "v1",
			"name": "Errors",
			"filters": []map[string]any{
				{"column": "status", "operation": "=", "value": "error"},
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "create", "--board", "board-1", "--name", "Errors", "--filter", "status:=:error"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail viewDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "v1" {
		t.Errorf("ID = %q, want %q", detail.ID, "v1")
	}
}

func TestViewCreate_MultipleFilters(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body api.CreateBoardViewRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if len(body.Filters) != 2 {
			t.Fatalf("filters count = %d, want 2", len(body.Filters))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "v2",
			"name": "Filtered",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "create", "--board", "board-1", "--name", "Filtered",
		"--filter", "status:=:error", "--filter", "env:=:prod"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestViewCreate_FilterWithoutValue(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body api.CreateBoardViewRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Filters[0].Operation != "exists" {
			t.Errorf("filter operation = %q, want %q", body.Filters[0].Operation, "exists")
		}
		if body.Filters[0].Value != nil {
			t.Errorf("filter value = %v, want nil", body.Filters[0].Value)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "v3",
			"name": "Exists",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "create", "--board", "board-1", "--name", "Exists", "--filter", "status:exists"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestViewCreate_WithFile(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "File View" {
			t.Errorf("name = %v, want %q", body["name"], "File View")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "v4",
			"name": "File View",
		})
	}))

	ts.InBuf.WriteString(`{"name":"File View","filters":[{"column":"env","operation":"=","value":"prod"}]}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "create", "--board", "board-1", "--file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestViewCreate_MissingFilter(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	opts.IOStreams.SetNeverPrompt(true)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "create", "--board", "board-1", "--name", "No Filters"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing filter")
	}
	if !strings.Contains(err.Error(), "at least one --filter is required") {
		t.Errorf("error = %q, want filter required message", err.Error())
	}
}

func TestViewCreate_MissingInput(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	opts.IOStreams.SetNeverPrompt(true)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "create", "--board", "board-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing input")
	}
	if !strings.Contains(err.Error(), "--name or --file is required") {
		t.Errorf("error = %q, want missing input message", err.Error())
	}
}

func TestViewCreate_InvalidFilter(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"view", "create", "--board", "board-1", "--name", "Bad", "--filter", "nocolon"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid filter")
	}
	if !strings.Contains(err.Error(), "invalid filter") {
		t.Errorf("error = %q, want invalid filter message", err.Error())
	}
}

func TestParseViewFilters(t *testing.T) {
	for _, tc := range []struct {
		name      string
		args      []string
		wantCount int
		wantErr   string
	}{
		{
			name:      "single filter with value",
			args:      []string{"status:=:error"},
			wantCount: 1,
		},
		{
			name:      "filter without value",
			args:      []string{"col:exists"},
			wantCount: 1,
		},
		{
			name:      "value with colons",
			args:      []string{"url:=:http://example.com:8080"},
			wantCount: 1,
		},
		{
			name:    "missing operation",
			args:    []string{"nocolon"},
			wantErr: "invalid filter",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			filters, err := parseViewFilters(tc.args)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("error = %q, want %q", err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(filters) != tc.wantCount {
				t.Errorf("filter count = %d, want %d", len(filters), tc.wantCount)
			}
		})
	}
}
