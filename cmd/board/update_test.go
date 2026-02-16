package board

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestUpdate_WithName(t *testing.T) {
	calls := 0
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		calls++
		if calls == 1 {
			// GET current board
			if r.Method != http.MethodGet {
				t.Errorf("call 1: method = %q, want GET", r.Method)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "abc123",
				"name":        "Old Name",
				"description": "Old desc",
				"type":        "flexible",
			})
			return
		}
		// PUT update
		if r.Method != http.MethodPut {
			t.Errorf("call 2: method = %q, want PUT", r.Method)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "New Name" {
			t.Errorf("name = %v, want %q", body["name"], "New Name")
		}
		if body["description"] != "Old desc" {
			t.Errorf("description = %v, want %q", body["description"], "Old desc")
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "abc123",
			"name":        "New Name",
			"description": "Old desc",
			"type":        "flexible",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "abc123", "--name", "New Name"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail boardDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "New Name" {
		t.Errorf("Name = %q, want %q", detail.Name, "New Name")
	}
}

func TestUpdate_MissingFlags(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "abc123"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing flags")
	}
	if !strings.Contains(err.Error(), "--file, --name, or --description is required") {
		t.Errorf("error = %q, want missing flags message", err.Error())
	}
}

func TestUpdate_FileStdinMerge(t *testing.T) {
	calls := 0
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		calls++
		if calls == 1 {
			if r.Method != http.MethodGet {
				t.Errorf("call 1: method = %q, want GET", r.Method)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "abc123",
				"name":        "Old Name",
				"description": "Old desc",
				"type":        "flexible",
			})
			return
		}
		if r.Method != http.MethodPut {
			t.Errorf("call 2: method = %q, want PUT", r.Method)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Merged Name" {
			t.Errorf("name = %v, want %q", body["name"], "Merged Name")
		}
		if body["description"] != "Old desc" {
			t.Errorf("description = %v, want %q", body["description"], "Old desc")
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "abc123",
			"name":        "Merged Name",
			"description": "Old desc",
			"type":        "flexible",
		})
	}))

	ts.InBuf.WriteString(`{"name":"Merged Name"}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "abc123", "--file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail boardDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "Merged Name" {
		t.Errorf("Name = %q, want %q", detail.Name, "Merged Name")
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2 (GET + PUT)", calls)
	}
}

func TestUpdate_FileStdinReplace(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT (no GET for replace)", r.Method)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Replaced" {
			t.Errorf("name = %v, want %q", body["name"], "Replaced")
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "abc123",
			"name": "Replaced",
			"type": "flexible",
		})
	}))

	ts.InBuf.WriteString(`{"name":"Replaced","type":"flexible"}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "abc123", "--file", "-", "--replace"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail boardDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "Replaced" {
		t.Errorf("Name = %q, want %q", detail.Name, "Replaced")
	}
}

func TestStripPanelDataset(t *testing.T) {
	for _, tc := range []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no panels",
			input:    `{"name":"Board"}`,
			expected: `{"name":"Board"}`,
		},
		{
			name:     "query panel with dataset",
			input:    `{"name":"Board","panels":[{"type":"query","query_panel":{"query_id":"q1","dataset":"my-dataset"}}]}`,
			expected: `{"name":"Board","panels":[{"type":"query","query_panel":{"query_id":"q1"}}]}`,
		},
		{
			name:     "query panel without dataset",
			input:    `{"name":"Board","panels":[{"type":"query","query_panel":{"query_id":"q1"}}]}`,
			expected: `{"name":"Board","panels":[{"type":"query","query_panel":{"query_id":"q1"}}]}`,
		},
		{
			name:     "slo panel unchanged",
			input:    `{"name":"Board","panels":[{"type":"slo","slo_panel":{"slo_id":"s1"}}]}`,
			expected: `{"name":"Board","panels":[{"type":"slo","slo_panel":{"slo_id":"s1"}}]}`,
		},
		{
			name:     "mixed panels",
			input:    `{"name":"Board","panels":[{"type":"query","query_panel":{"query_id":"q1","dataset":"ds"}},{"type":"slo","slo_panel":{"slo_id":"s1"}}]}`,
			expected: `{"name":"Board","panels":[{"type":"query","query_panel":{"query_id":"q1"}},{"type":"slo","slo_panel":{"slo_id":"s1"}}]}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := stripPanelDataset([]byte(tc.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var gotParsed, expectedParsed any
			if err := json.Unmarshal(got, &gotParsed); err != nil {
				t.Fatalf("unmarshal got: %v", err)
			}
			if err := json.Unmarshal([]byte(tc.expected), &expectedParsed); err != nil {
				t.Fatalf("unmarshal expected: %v", err)
			}

			gotJSON, _ := json.Marshal(gotParsed)
			expectedJSON, _ := json.Marshal(expectedParsed)
			if string(gotJSON) != string(expectedJSON) {
				t.Errorf("got %s, want %s", gotJSON, expectedJSON)
			}
		})
	}
}

func TestUpdate_FileReplaceFillsRequiredFields(t *testing.T) {
	for _, tc := range []struct {
		name         string
		input        string
		expectedName string
		expectedType string
		expectGET    bool
	}{
		{
			name:         "missing name and type",
			input:        `{"description":"new desc"}`,
			expectedName: "Existing Board",
			expectedType: "flexible",
			expectGET:    true,
		},
		{
			name:         "missing type only",
			input:        `{"name":"Custom Name"}`,
			expectedName: "Custom Name",
			expectedType: "flexible",
			expectGET:    true,
		},
		{
			name:         "missing name only",
			input:        `{"type":"flexible","description":"new desc"}`,
			expectedName: "Existing Board",
			expectedType: "flexible",
			expectGET:    true,
		},
		{
			name:         "both present",
			input:        `{"name":"Custom","type":"flexible"}`,
			expectedName: "Custom",
			expectedType: "flexible",
			expectGET:    false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotGET bool
			opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet {
					gotGET = true
					_ = json.NewEncoder(w).Encode(map[string]any{
						"id":   "abc123",
						"name": "Existing Board",
						"type": "flexible",
					})
					return
				}
				var body map[string]any
				_ = json.NewDecoder(r.Body).Decode(&body)
				if body["name"] != tc.expectedName {
					t.Errorf("name = %v, want %q", body["name"], tc.expectedName)
				}
				if body["type"] != tc.expectedType {
					t.Errorf("type = %v, want %q", body["type"], tc.expectedType)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":   "abc123",
					"name": tc.expectedName,
					"type": tc.expectedType,
				})
			}))

			ts.InBuf.WriteString(tc.input)

			cmd := NewCmd(opts)
			cmd.SetArgs([]string{"update", "abc123", "--file", "-", "--replace"})
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}

			if gotGET != tc.expectGET {
				t.Errorf("GET request = %v, want %v", gotGET, tc.expectGET)
			}
		})
	}
}

func TestUpdate_StripsPanelDataset(t *testing.T) {
	calls := 0
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		calls++
		if calls == 1 {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":   "abc123",
				"name": "Board",
				"type": "flexible",
				"panels": []map[string]any{
					{
						"type": "query",
						"query_panel": map[string]any{
							"query_id":            "q1",
							"query_annotation_id": "a1",
							"dataset":             "my-dataset",
						},
					},
				},
			})
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		panels, ok := body["panels"].([]any)
		if !ok || len(panels) != 1 {
			t.Fatalf("expected 1 panel, got %v", body["panels"])
		}
		panel := panels[0].(map[string]any)
		qp := panel["query_panel"].(map[string]any)
		if _, has := qp["dataset"]; has {
			t.Error("dataset should be stripped from query_panel")
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "abc123",
			"name": "New Name",
			"type": "flexible",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "abc123", "--name", "New Name"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_FileStripsPanelDataset(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		panels, ok := body["panels"].([]any)
		if !ok || len(panels) != 1 {
			t.Fatalf("expected 1 panel, got %v", body["panels"])
		}
		panel := panels[0].(map[string]any)
		qp := panel["query_panel"].(map[string]any)
		if _, has := qp["dataset"]; has {
			t.Error("dataset should be stripped from query_panel")
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "abc123",
			"name": "Board",
			"type": "flexible",
		})
	}))

	ts.InBuf.WriteString(`{"name":"Board","type":"flexible","panels":[{"type":"query","query_panel":{"query_id":"q1","query_annotation_id":"a1","dataset":"my-dataset"}}]}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "abc123", "--file", "-", "--replace"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"board not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "missing", "--name", "New"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("error = %q, want HTTP 404", err.Error())
	}
}
