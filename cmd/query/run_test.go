package query

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/bendrucker/honeycomb-cli/internal/api"
)

func TestRun_File(t *testing.T) {
	var getResultCalls atomic.Int32

	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/1/queries/test-dataset":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":         "qry-1",
				"time_range": 3600,
				"breakdowns": []string{"service.name"},
				"calculations": []map[string]any{
					{"op": "COUNT"},
					{"op": "AVG", "column": "duration_ms"},
				},
			})

		case r.Method == http.MethodPost && r.URL.Path == "/1/query_results/test-dataset":
			var body api.CreateQueryResultRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode request body: %v", err)
			}
			if body.QueryId == nil || *body.QueryId != "qry-1" {
				t.Errorf("query_id = %v, want %q", body.QueryId, "qry-1")
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":       "result-1",
				"complete": false,
			})

		case r.Method == http.MethodGet && r.URL.Path == "/1/query_results/test-dataset/result-1":
			n := getResultCalls.Add(1)
			if n < 2 {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":       "result-1",
					"complete": false,
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"complete": true,
				"id":       "result-1",
				"data": map[string]any{
					"results": []map[string]any{
						{"data": map[string]any{"service.name": "api", "COUNT": 1500, "AVG(duration_ms)": 42.5}},
						{"data": map[string]any{"service.name": "web", "COUNT": 800, "AVG(duration_ms)": 120.3}},
					},
				},
				"query": map[string]any{
					"id":         "qry-1",
					"breakdowns": []string{"service.name"},
					"calculations": []map[string]any{
						{"op": "COUNT"},
						{"op": "AVG", "column": "duration_ms"},
					},
				},
				"links": map[string]any{
					"query_url": "https://ui.honeycomb.io/query/result-1",
				},
			})

		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	ts.InBuf.WriteString(`{"time_range": 3600, "breakdowns": ["service.name"], "calculations": [{"op": "COUNT"}, {"op": "AVG", "column": "duration_ms"}]}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"run", "--dataset", "test-dataset", "--file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var result api.QueryResultDetails
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result.Complete == nil || !*result.Complete {
		t.Error("expected complete=true in output")
	}
	if result.Data == nil || result.Data.Results == nil {
		t.Fatal("expected results in output")
	}
	if len(*result.Data.Results) != 2 {
		t.Errorf("got %d results, want 2", len(*result.Data.Results))
	}

	if n := getResultCalls.Load(); n < 2 {
		t.Errorf("expected at least 2 GET result calls, got %d", n)
	}

	if !strings.Contains(ts.ErrBuf.String(), "https://ui.honeycomb.io/query/result-1") {
		t.Errorf("expected query URL in stderr, got %q", ts.ErrBuf.String())
	}
}

func TestRun_Annotation(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/1/query_annotations/test-dataset/ann-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":       "ann-1",
				"name":     "My Query",
				"query_id": "qry-1",
			})

		case r.Method == http.MethodPost && r.URL.Path == "/1/query_results/test-dataset":
			var body api.CreateQueryResultRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode request body: %v", err)
			}
			if body.QueryId == nil || *body.QueryId != "qry-1" {
				t.Errorf("query_id = %v, want %q", body.QueryId, "qry-1")
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":       "result-1",
				"complete": false,
			})

		case r.Method == http.MethodGet && r.URL.Path == "/1/query_results/test-dataset/result-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"complete": true,
				"id":       "result-1",
				"data": map[string]any{
					"results": []map[string]any{
						{"data": map[string]any{"COUNT": 42}},
					},
				},
				"query": map[string]any{
					"id": "qry-1",
					"calculations": []map[string]any{
						{"op": "COUNT"},
					},
				},
			})

		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"run", "--dataset", "test-dataset", "--annotation", "ann-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var result api.QueryResultDetails
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result.Complete == nil || !*result.Complete {
		t.Error("expected complete=true in output")
	}
}

func TestRun_NoInput_NonInteractive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("unexpected API call")
		w.WriteHeader(http.StatusInternalServerError)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"run", "--dataset", "test-dataset"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing input")
	}
	if !strings.Contains(err.Error(), "either --file or --annotation is required") {
		t.Errorf("error = %q, want missing input message", err.Error())
	}
}

func TestRun_MutuallyExclusive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("unexpected API call")
		w.WriteHeader(http.StatusInternalServerError)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"run", "--dataset", "test-dataset", "--file", "query.json", "--annotation", "ann-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
}
