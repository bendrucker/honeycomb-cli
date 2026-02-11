package query

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestUpdate_Flags(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/query_annotations/test-dataset/ann-1" {
			t.Errorf("path = %q, want /1/query_annotations/test-dataset/ann-1", r.URL.Path)
		}

		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":       "ann-1",
				"name":     "Latency Query",
				"query_id": "q-abc",
			})
		case http.MethodPut:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			if body["name"] != "Updated Query" {
				t.Errorf("body.name = %q, want %q", body["name"], "Updated Query")
			}
			if body["query_id"] != "q-abc" {
				t.Errorf("body.query_id = %q, want %q (should be preserved)", body["query_id"], "q-abc")
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":       "ann-1",
				"name":     "Updated Query",
				"query_id": "q-abc",
			})
		default:
			t.Errorf("unexpected method %q", r.Method)
		}
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "test-dataset", "ann-1", "--name", "Updated Query"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail annotationDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "Updated Query" {
		t.Errorf("Name = %q, want %q", detail.Name, "Updated Query")
	}
}

func TestUpdate_File(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":       "ann-1",
			"name":     "From File",
			"query_id": "q-abc",
		})
	}))

	ts.InBuf.WriteString(`{"name":"From File","query_id":"q-abc"}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "--dataset", "test-dataset", "ann-1", "--file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail annotationDetail
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
	cmd.SetArgs([]string{"update", "--dataset", "test-dataset", "ann-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for no flags")
	}
	if !strings.Contains(err.Error(), "--file") {
		t.Errorf("error = %q, want message about required flags", err.Error())
	}
}
