package query

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCreate_File(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/query_annotations/test-dataset" {
			t.Errorf("path = %q, want /1/query_annotations/test-dataset", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["name"] != "Latency Query" {
			t.Errorf("body.name = %q, want %q", body["name"], "Latency Query")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":       "ann-new",
			"name":     "Latency Query",
			"query_id": "q-abc",
		})
	}))

	ts.InBuf.WriteString(`{"name":"Latency Query","query_id":"q-abc"}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "--file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail annotationDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "ann-new" {
		t.Errorf("ID = %q, want %q", detail.ID, "ann-new")
	}
	if detail.Name != "Latency Query" {
		t.Errorf("Name = %q, want %q", detail.Name, "Latency Query")
	}
}

func TestCreate_Flags(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["name"] != "My Query" {
			t.Errorf("body.name = %q, want %q", body["name"], "My Query")
		}
		if body["query_id"] != "q-123" {
			t.Errorf("body.query_id = %q, want %q", body["query_id"], "q-123")
		}
		if body["description"] != "A description" {
			t.Errorf("body.description = %q, want %q", body["description"], "A description")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "ann-new",
			"name":        "My Query",
			"query_id":    "q-123",
			"description": "A description",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "--name", "My Query", "--query-id", "q-123", "--description", "A description"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail annotationDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "My Query" {
		t.Errorf("Name = %q, want %q", detail.Name, "My Query")
	}
}

func TestCreate_Flags_MissingQueryID(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "--name", "My Query"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --query-id")
	}
	if !strings.Contains(err.Error(), "--query-id is required") {
		t.Errorf("error = %q, want --query-id required message", err.Error())
	}
}

func TestCreate_Flags_MissingName(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "--query-id", "q-123"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --name")
	}
	if !strings.Contains(err.Error(), "--name is required") {
		t.Errorf("error = %q, want --name required message", err.Error())
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
