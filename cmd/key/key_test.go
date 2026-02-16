package key

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

	if err := config.SetKey("default", config.KeyManagement, "test-mgmt-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyManagement) })

	return opts, ts
}

func TestList(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2/teams/my-team/api-keys" {
			t.Errorf("path = %q, want /2/teams/my-team/api-keys", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-mgmt-key" {
			t.Errorf("Authorization = %q, want Bearer test-mgmt-key", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{
			"data": [
				{
					"id": "hcxik_01abc",
					"type": "api-keys",
					"attributes": {
						"name": "My Ingest Key",
						"key_type": "ingest",
						"disabled": false
					}
				},
				{
					"id": "hcxlk_02def",
					"type": "api-keys",
					"attributes": {
						"name": "My Config Key",
						"key_type": "configuration",
						"disabled": true
					}
				}
			]
		}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--team", "my-team", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []keyItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != "hcxik_01abc" {
		t.Errorf("items[0].ID = %q, want %q", items[0].ID, "hcxik_01abc")
	}
	if items[0].Name != "My Ingest Key" {
		t.Errorf("items[0].Name = %q, want %q", items[0].Name, "My Ingest Key")
	}
	if items[0].KeyType != "ingest" {
		t.Errorf("items[0].KeyType = %q, want %q", items[0].KeyType, "ingest")
	}
	if items[0].Disabled != false {
		t.Errorf("items[0].Disabled = %v, want false", items[0].Disabled)
	}
	if items[1].Disabled != true {
		t.Errorf("items[1].Disabled = %v, want true", items[1].Disabled)
	}
}

func TestList_WithTypeFilter(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("filter[type]"); got != "ingest" {
			t.Errorf("filter[type] = %q, want %q", got, "ingest")
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data": []}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--team", "my-team", "list", "--type", "ingest"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestList_Empty(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data": []}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--team", "my-team", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []keyItem
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
	cmd.SetArgs([]string{"--team", "my-team", "list"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "no management key configured") {
		t.Errorf("error = %q, want missing key message", err.Error())
	}
}

func TestGet(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2/teams/my-team/api-keys/hcxik_01abc" {
			t.Errorf("path = %q, want /2/teams/my-team/api-keys/hcxik_01abc", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{
			"data": {
				"id": "hcxik_01abc",
				"type": "api-keys",
				"attributes": {
					"name": "My Ingest Key",
					"key_type": "ingest",
					"disabled": false
				}
			}
		}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--team", "my-team", "get", "hcxik_01abc"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail keyDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "hcxik_01abc" {
		t.Errorf("ID = %q, want %q", detail.ID, "hcxik_01abc")
	}
	if detail.Name != "My Ingest Key" {
		t.Errorf("Name = %q, want %q", detail.Name, "My Ingest Key")
	}
	if detail.KeyType != "ingest" {
		t.Errorf("KeyType = %q, want %q", detail.KeyType, "ingest")
	}
}

func TestCreate(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/2/teams/my-team/api-keys" {
			t.Errorf("path = %q, want /2/teams/my-team/api-keys", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want application/vnd.api+json", ct)
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"data": {
				"id": "hcxik_01new",
				"type": "api-keys",
				"attributes": {
					"name": "New Key",
					"key_type": "ingest",
					"disabled": false,
					"secret": "hcxik_01new_secret_value"
				},
				"links": {"self": "/2/teams/my-team/api-keys/hcxik_01new"},
				"relationships": {
					"environment": {"data": {"id": "env1", "type": "environments"}}
				}
			}
		}`))
	}))

	ts.InBuf.WriteString(`{"data":{"type":"api-keys","attributes":{"name":"New Key","key_type":"ingest"},"relationships":{"environment":{"data":{"id":"env1","type":"environments"}}}}}`)
	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--team", "my-team", "create", "--file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail keyDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "hcxik_01new" {
		t.Errorf("ID = %q, want %q", detail.ID, "hcxik_01new")
	}
	if detail.Secret != "hcxik_01new_secret_value" {
		t.Errorf("Secret = %q, want %q", detail.Secret, "hcxik_01new_secret_value")
	}
	if detail.Name != "New Key" {
		t.Errorf("Name = %q, want %q", detail.Name, "New Key")
	}

	errOutput := ts.ErrBuf.String()
	if !strings.Contains(errOutput, "Save this secret now") {
		t.Errorf("stderr = %q, want secret warning", errOutput)
	}
}

func TestUpdate(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %q, want PATCH", r.Method)
		}
		if r.URL.Path != "/2/teams/my-team/api-keys/hcxik_01abc" {
			t.Errorf("path = %q, want /2/teams/my-team/api-keys/hcxik_01abc", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{
			"data": {
				"id": "hcxik_01abc",
				"type": "api-keys",
				"attributes": {
					"name": "Updated Key",
					"key_type": "ingest",
					"disabled": false
				}
			}
		}`))
	}))

	ts.InBuf.WriteString(`{"data":{"type":"api-keys","attributes":{"name":"Updated Key"},"id":"hcxik_01abc"}}`)
	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--team", "my-team", "update", "hcxik_01abc", "--file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail keyDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Name != "Updated Key" {
		t.Errorf("Name = %q, want %q", detail.Name, "Updated Key")
	}
}

func TestDelete_WithYes(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/2/teams/my-team/api-keys/hcxik_01abc" {
			t.Errorf("path = %q, want /2/teams/my-team/api-keys/hcxik_01abc", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--team", "my-team", "delete", "hcxik_01abc", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var result map[string]string
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result["id"] != "hcxik_01abc" {
		t.Errorf("id = %q, want %q", result["id"], "hcxik_01abc")
	}
}

func TestDelete_NoYesNonInteractive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	opts.IOStreams.SetNeverPrompt(true)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--team", "my-team", "delete", "hcxik_01abc"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --yes")
	}
	if !strings.Contains(err.Error(), "--yes is required") {
		t.Errorf("error = %q, want --yes required message", err.Error())
	}
}

func TestList_Unauthorized(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unknown API key - check your credentials"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--team", "my-team", "list"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if !strings.Contains(err.Error(), "HTTP 401") {
		t.Errorf("error = %q, want HTTP 401", err.Error())
	}
}

func TestList_TeamFromConfig(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2/teams/config-team/api-keys" {
			t.Errorf("path = %q, want /2/teams/config-team/api-keys", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data": []}`))
	}))

	opts.Config = &config.Config{
		Profiles: map[string]*config.Profile{
			"default": {Team: "config-team"},
		},
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := ts.OutBuf.String()
	if output == "" {
		t.Fatal("expected output")
	}
}

func TestList_MissingTeam(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --team")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("error = %q, want required error", err.Error())
	}
}
