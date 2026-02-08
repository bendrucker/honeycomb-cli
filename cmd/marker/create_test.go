package marker

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestCreate(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/1/markers/test-dataset" {
			t.Errorf("path = %q, want /1/markers/test-dataset", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal request: %v", err)
		}
		if req["type"] != "deploy" {
			t.Errorf("type = %v, want deploy", req["type"])
		}
		if req["message"] != "v2.0.0" {
			t.Errorf("message = %v, want v2.0.0", req["message"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         "new123",
			"type":       "deploy",
			"message":    "v2.0.0",
			"start_time": 1700000000,
			"created_at": "2024-01-01T00:00:00Z",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "--type", "deploy", "--message", "v2.0.0", "--start-time", "1700000000"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var item markerItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &item); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if item.ID != "new123" {
		t.Errorf("ID = %q, want %q", item.ID, "new123")
	}
	if item.Type != "deploy" {
		t.Errorf("Type = %q, want %q", item.Type, "deploy")
	}
}

func TestCreate_DefaultStartTime(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal request: %v", err)
		}
		st, ok := req["start_time"].(float64)
		if !ok || st == 0 {
			t.Errorf("start_time should be set to current time, got %v", req["start_time"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         "new456",
			"start_time": int(st),
			"created_at": "2024-01-01T00:00:00Z",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--dataset", "test-dataset", "--type", "deploy"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
