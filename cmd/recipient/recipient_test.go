package recipient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/zalando/go-keyring"
)

func init() {
	keyring.MockInit()
}

func setupTest(t *testing.T, handler http.Handler) (*options.RootOptions, *iostreams.TestStreams) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    srv.URL,
		Format:    "json",
	}

	if err := config.SetKey("default", config.KeyConfig, "test-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })

	return opts, ts
}

func TestList(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/recipients" {
			t.Errorf("path = %q, want /1/recipients", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"id": "r1",
				"type": "email",
				"details": {"address": "test@example.com"},
				"created_at": "2025-01-01T00:00:00Z",
				"updated_at": "2025-01-01T00:00:00Z"
			},
			{
				"id": "r2",
				"type": "slack",
				"details": {"channel": "#alerts"},
				"created_at": "2025-01-02T00:00:00Z",
				"updated_at": "2025-01-02T00:00:00Z"
			}
		]`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []recipientItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != "r1" {
		t.Errorf("items[0].ID = %q, want %q", items[0].ID, "r1")
	}
	if items[0].Type != "email" {
		t.Errorf("items[0].Type = %q, want %q", items[0].Type, "email")
	}
	if items[0].Target != "test@example.com" {
		t.Errorf("items[0].Target = %q, want %q", items[0].Target, "test@example.com")
	}
	if items[1].Type != "slack" {
		t.Errorf("items[1].Type = %q, want %q", items[1].Type, "slack")
	}
	if items[1].Target != "#alerts" {
		t.Errorf("items[1].Target = %q, want %q", items[1].Target, "#alerts")
	}
}

func TestList_Empty(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []recipientItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

func TestList_NoKey(t *testing.T) {
	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    "http://localhost",
		Format:    "json",
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "no config key configured") {
		t.Errorf("error = %q, want missing key message", err.Error())
	}
}

func TestList_Unauthorized(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unknown API key - check your credentials"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if !strings.Contains(err.Error(), "HTTP 401") {
		t.Errorf("error = %q, want HTTP 401", err.Error())
	}
	if !strings.Contains(err.Error(), "unknown API key") {
		t.Errorf("error = %q, want error message from body", err.Error())
	}
}

func TestGet(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/recipients/r1" {
			t.Errorf("path = %q, want /1/recipients/r1", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "r1",
			"type": "email",
			"details": {"address": "test@example.com"},
			"created_at": "2025-01-01T00:00:00Z",
			"updated_at": "2025-01-01T00:00:00Z"
		}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "r1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail recipientDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "r1" {
		t.Errorf("ID = %q, want %q", detail.ID, "r1")
	}
	if detail.Type != "email" {
		t.Errorf("Type = %q, want %q", detail.Type, "email")
	}
	if detail.CreatedAt != "2025-01-01T00:00:00Z" {
		t.Errorf("CreatedAt = %q, want %q", detail.CreatedAt, "2025-01-01T00:00:00Z")
	}
}

func TestGet_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"recipient not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "missing"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("error = %q, want HTTP 404", err.Error())
	}
}

func TestGet_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}

func TestDelete_WithYes(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/recipients/r1" {
			t.Errorf("path = %q, want /1/recipients/r1", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete", "r1", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var result map[string]string
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result["id"] != "r1" {
		t.Errorf("id = %q, want %q", result["id"], "r1")
	}
}

func TestDelete_NoYesNonInteractive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	opts.IOStreams.SetNeverPrompt(true)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete", "r1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-interactive without --yes")
	}
	if !strings.Contains(err.Error(), "--yes is required in non-interactive mode") {
		t.Errorf("error = %q, want non-interactive error", err.Error())
	}
}

func TestDelete_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}

func TestCreate(t *testing.T) {
	for _, tc := range []struct {
		name     string
		args     []string
		wantType string
		wantKey  string
		wantVal  string
	}{
		{
			name:     "email",
			args:     []string{"create", "--type", "email", "--target", "test@example.com"},
			wantType: "email",
			wantKey:  "email_address",
			wantVal:  "test@example.com",
		},
		{
			name:     "slack",
			args:     []string{"create", "--type", "slack", "--channel", "#alerts"},
			wantType: "slack",
			wantKey:  "slack_channel",
			wantVal:  "#alerts",
		},
		{
			name:     "pagerduty",
			args:     []string{"create", "--type", "pagerduty", "--integration-key", "abc123"},
			wantType: "pagerduty",
			wantKey:  "pagerduty_integration_key",
			wantVal:  "abc123",
		},
		{
			name:     "webhook",
			args:     []string{"create", "--type", "webhook", "--url", "https://example.com/hook", "--name", "my-hook"},
			wantType: "webhook",
			wantKey:  "webhook_url",
			wantVal:  "https://example.com/hook",
		},
		{
			name:     "msteams workflow",
			args:     []string{"create", "--type", "msteams_workflow", "--url", "https://teams.example.com/hook", "--name", "teams-hook"},
			wantType: "msteams_workflow",
			wantKey:  "webhook_url",
			wantVal:  "https://teams.example.com/hook",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotBody map[string]any
			opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/1/recipients" {
					t.Errorf("path = %q, want /1/recipients", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("method = %q, want POST", r.Method)
				}
				if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
					t.Fatalf("decoding request body: %v", err)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":   "r-new",
					"type": tc.wantType,
				})
			}))

			cmd := NewCmd(opts)
			cmd.SetArgs(tc.args)
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}

			if gotType, _ := gotBody["type"].(string); gotType != tc.wantType {
				t.Errorf("body type = %q, want %q", gotType, tc.wantType)
			}
			details, _ := gotBody["details"].(map[string]any)
			if details == nil {
				t.Fatal("body details is nil")
			}
			if val, _ := details[tc.wantKey].(string); val != tc.wantVal {
				t.Errorf("details[%q] = %q, want %q", tc.wantKey, val, tc.wantVal)
			}

			var detail recipientDetail
			if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
				t.Fatalf("unmarshal output: %v", err)
			}
			if detail.ID != "r-new" {
				t.Errorf("output ID = %q, want %q", detail.ID, "r-new")
			}
		})
	}
}

func TestCreate_FromFile(t *testing.T) {
	var gotBody map[string]any
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "r-file",
			"type": "email",
		})
	}))

	ts.InBuf.WriteString(`{"type":"email","details":{"email_address":"file@example.com"}}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "-f", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if gotType, _ := gotBody["type"].(string); gotType != "email" {
		t.Errorf("body type = %q, want %q", gotType, "email")
	}

	var detail recipientDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "r-file" {
		t.Errorf("output ID = %q, want %q", detail.ID, "r-file")
	}
}

func TestCreate_MissingTypeNonInteractive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	opts.IOStreams.SetNeverPrompt(true)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing type")
	}
	if !strings.Contains(err.Error(), "--type or --file is required") {
		t.Errorf("error = %q, want missing type message", err.Error())
	}
}

func TestCreate_MissingDetailNonInteractive(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "email missing target",
			args:    []string{"create", "--type", "email"},
			wantErr: "--target is required for email",
		},
		{
			name:    "slack missing channel",
			args:    []string{"create", "--type", "slack"},
			wantErr: "--channel is required for slack",
		},
		{
			name:    "pagerduty missing key",
			args:    []string{"create", "--type", "pagerduty"},
			wantErr: "--integration-key is required for pagerduty",
		},
		{
			name:    "webhook missing url",
			args:    []string{"create", "--type", "webhook"},
			wantErr: "--url is required for webhook",
		},
		{
			name:    "webhook missing name",
			args:    []string{"create", "--type", "webhook", "--url", "https://example.com"},
			wantErr: "--name is required for webhook",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
			opts.IOStreams.SetNeverPrompt(true)

			cmd := NewCmd(opts)
			cmd.SetArgs(tc.args)
			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestCreate_FileMutuallyExclusive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "-f", "-", "--type", "email"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(err.Error(), "none of the others can be") {
		t.Errorf("error = %q, want mutually exclusive message", err.Error())
	}
}

func TestTriggers(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/recipients/r1/triggers" {
			t.Errorf("path = %q, want /1/recipients/r1/triggers", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":          "t1",
				"name":        "High Latency",
				"description": "Fires when p99 > 1s",
				"disabled":    false,
				"triggered":   true,
				"alert_type":  "on_change",
				"threshold":   map[string]any{"op": ">", "value": 1000},
			},
			{
				"id":        "t2",
				"name":      "Error Rate",
				"disabled":  false,
				"triggered": false,
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"triggers", "r1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []triggerItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].Name != "High Latency" {
		t.Errorf("items[0].Name = %q, want %q", items[0].Name, "High Latency")
	}
	if !items[0].Triggered {
		t.Errorf("items[0].Triggered = false, want true")
	}
	if items[0].Threshold != "> 1000" {
		t.Errorf("items[0].Threshold = %q, want %q", items[0].Threshold, "> 1000")
	}
	if items[1].Name != "Error Rate" {
		t.Errorf("items[1].Name = %q, want %q", items[1].Name, "Error Rate")
	}
}

func TestTriggers_MissingArg(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"triggers"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}

func TestUpdate_Flags(t *testing.T) {
	for _, tc := range []struct {
		name       string
		args       []string
		getResp    map[string]any
		wantKey    string
		wantVal    string
		wantType   string
		preserveID bool
	}{
		{
			name: "email target",
			args: []string{"update", "r1", "--target", "new@example.com"},
			getResp: map[string]any{
				"id":   "r1",
				"type": "email",
				"details": map[string]any{
					"email_address": "old@example.com",
				},
			},
			wantKey:  "email_address",
			wantVal:  "new@example.com",
			wantType: "email",
		},
		{
			name: "slack channel",
			args: []string{"update", "r1", "--channel", "#new-alerts"},
			getResp: map[string]any{
				"id":   "r1",
				"type": "slack",
				"details": map[string]any{
					"slack_channel": "#old-alerts",
				},
			},
			wantKey:  "slack_channel",
			wantVal:  "#new-alerts",
			wantType: "slack",
		},
		{
			name: "pagerduty integration key",
			args: []string{"update", "r1", "--integration-key", "new-key"},
			getResp: map[string]any{
				"id":   "r1",
				"type": "pagerduty",
				"details": map[string]any{
					"pagerduty_integration_key":  "old-key",
					"pagerduty_integration_name": "My PD",
				},
			},
			wantKey:  "pagerduty_integration_key",
			wantVal:  "new-key",
			wantType: "pagerduty",
		},
		{
			name: "pagerduty name",
			args: []string{"update", "r1", "--name", "New PD Name"},
			getResp: map[string]any{
				"id":   "r1",
				"type": "pagerduty",
				"details": map[string]any{
					"pagerduty_integration_key":  "key-123",
					"pagerduty_integration_name": "Old PD Name",
				},
			},
			wantKey:  "pagerduty_integration_name",
			wantVal:  "New PD Name",
			wantType: "pagerduty",
		},
		{
			name: "webhook url",
			args: []string{"update", "r1", "--url", "https://new.example.com/hook"},
			getResp: map[string]any{
				"id":   "r1",
				"type": "webhook",
				"details": map[string]any{
					"webhook_url":  "https://old.example.com/hook",
					"webhook_name": "my-hook",
				},
			},
			wantKey:  "webhook_url",
			wantVal:  "https://new.example.com/hook",
			wantType: "webhook",
		},
		{
			name: "webhook name",
			args: []string{"update", "r1", "--name", "new-hook"},
			getResp: map[string]any{
				"id":   "r1",
				"type": "webhook",
				"details": map[string]any{
					"webhook_url":  "https://example.com/hook",
					"webhook_name": "old-hook",
				},
			},
			wantKey:  "webhook_name",
			wantVal:  "new-hook",
			wantType: "webhook",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotBody map[string]any
			opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(tc.getResp)
				case http.MethodPut:
					if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
						t.Fatalf("decoding request body: %v", err)
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(map[string]any{
						"id":   "r1",
						"type": tc.wantType,
					})
				default:
					t.Errorf("unexpected method %q", r.Method)
				}
			}))

			cmd := NewCmd(opts)
			cmd.SetArgs(tc.args)
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}

			if gotBody["id"] != nil {
				t.Error("body should not include read-only id field")
			}
			if gotType, _ := gotBody["type"].(string); gotType != tc.wantType {
				t.Errorf("body type = %q, want %q", gotType, tc.wantType)
			}
			details, _ := gotBody["details"].(map[string]any)
			if details == nil {
				t.Fatal("body details is nil")
			}
			if val, _ := details[tc.wantKey].(string); val != tc.wantVal {
				t.Errorf("details[%q] = %q, want %q", tc.wantKey, val, tc.wantVal)
			}

			var detail recipientDetail
			if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
				t.Fatalf("unmarshal output: %v", err)
			}
			if detail.ID != "r1" {
				t.Errorf("output ID = %q, want %q", detail.ID, "r1")
			}
		})
	}
}

func TestUpdate_File(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "r1",
			"type": "email",
		})
	}))

	ts.InBuf.WriteString(`{"type":"email","details":{"email_address":"file@example.com"}}`)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "r1", "-f", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail recipientDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "r1" {
		t.Errorf("output ID = %q, want %q", detail.ID, "r1")
	}
}

func TestUpdate_NoFlags(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "r1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for no flags")
	}
	if !strings.Contains(err.Error(), "--file") {
		t.Errorf("error = %q, want message about required flags", err.Error())
	}
}

func TestUpdate_FileMutuallyExclusive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "r1", "-f", "-", "--target", "test@example.com"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(err.Error(), "none of the others can be") {
		t.Errorf("error = %q, want mutually exclusive message", err.Error())
	}
}
