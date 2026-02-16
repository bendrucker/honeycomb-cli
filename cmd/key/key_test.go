package key

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
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

func TestCreate_File(t *testing.T) {
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

func TestCreate_Flags(t *testing.T) {
	for _, tc := range []struct {
		name            string
		keyType         string
		wantKeyType     string
		wantContentType string
	}{
		{
			name:    "ingest key",
			keyType: "ingest",
		},
		{
			name:    "configuration key",
			keyType: "configuration",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
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

				var body api.ApiKeyCreateRequest
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode request body: %v", err)
				}
				if body.Data.Type != api.ApiKeyCreateRequestDataTypeApiKeys {
					t.Errorf("data.type = %q, want %q", body.Data.Type, api.ApiKeyCreateRequestDataTypeApiKeys)
				}
				if body.Data.Relationships.Environment.Data.Id != "env1" {
					t.Errorf("environment ID = %q, want %q", body.Data.Relationships.Environment.Data.Id, "env1")
				}

				var attrs struct {
					Name    string `json:"name"`
					KeyType string `json:"key_type"`
				}
				raw, _ := body.Data.Attributes.MarshalJSON()
				_ = json.Unmarshal(raw, &attrs)
				if attrs.Name != "My Key" {
					t.Errorf("name = %q, want %q", attrs.Name, "My Key")
				}
				if attrs.KeyType != tc.keyType {
					t.Errorf("key_type = %q, want %q", attrs.KeyType, tc.keyType)
				}

				w.Header().Set("Content-Type", "application/vnd.api+json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{
					"data": {
						"id": "hcxik_01new",
						"type": "api-keys",
						"attributes": {
							"name": "My Key",
							"key_type": "` + tc.keyType + `",
							"disabled": false,
							"secret": "secret123"
						}
					}
				}`))
			}))

			cmd := NewCmd(opts)
			cmd.SetArgs([]string{"--team", "my-team", "create", "--name", "My Key", "--key-type", tc.keyType, "--environment", "env1"})
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}

			var detail keyDetail
			if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
				t.Fatalf("unmarshal output: %v", err)
			}
			if detail.Name != "My Key" {
				t.Errorf("Name = %q, want %q", detail.Name, "My Key")
			}
			if detail.KeyType != tc.keyType {
				t.Errorf("KeyType = %q, want %q", detail.KeyType, tc.keyType)
			}
			if detail.Secret != "secret123" {
				t.Errorf("Secret = %q, want %q", detail.Secret, "secret123")
			}
		})
	}
}

func TestCreate_FlagsMutuallyExclusiveWithFile(t *testing.T) {
	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    "http://localhost",
		Format:    "json",
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--team", "my-team", "create", "--file", "-", "--name", "My Key"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(err.Error(), "if any flags in the group [file name] are set none of the others can be") {
		t.Errorf("error = %q, want mutual exclusion message", err.Error())
	}
}

func TestCreate_NonInteractiveRequiresFlags(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing name",
			args:    []string{"--team", "my-team", "create", "--key-type", "ingest", "--environment", "env1"},
			wantErr: "--name is required",
		},
		{
			name:    "missing key type",
			args:    []string{"--team", "my-team", "create", "--name", "My Key", "--environment", "env1"},
			wantErr: "--key-type is required",
		},
		{
			name:    "missing environment",
			args:    []string{"--team", "my-team", "create", "--name", "My Key", "--key-type", "ingest"},
			wantErr: "--environment is required",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ts := iostreams.Test(t)
			ts.SetNeverPrompt(true)
			opts := &options.RootOptions{
				IOStreams: ts.IOStreams,
				Config:    &config.Config{},
				APIUrl:    "http://localhost",
				Format:    "json",
			}

			if err := config.SetKey("default", config.KeyManagement, "test-key"); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyManagement) })

			cmd := NewCmd(opts)
			cmd.SetArgs(tc.args)
			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected error for missing flag")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestCreate_InvalidKeyType(t *testing.T) {
	ts := iostreams.Test(t)
	ts.SetNeverPrompt(true)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    "http://localhost",
		Format:    "json",
	}

	if err := config.SetKey("default", config.KeyManagement, "test-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyManagement) })

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--team", "my-team", "create", "--name", "My Key", "--key-type", "invalid", "--environment", "env1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid key type")
	}
	if !strings.Contains(err.Error(), "--key-type must be ingest or configuration") {
		t.Errorf("error = %q, want key type validation message", err.Error())
	}
}

func TestUpdate_File(t *testing.T) {
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

func TestUpdate_Flags(t *testing.T) {
	for _, tc := range []struct {
		name         string
		id           string
		args         []string
		wantName     string
		wantDisabled bool
	}{
		{
			name:         "rename ingest key",
			id:           "hcxik_01abc",
			args:         []string{"--name", "Renamed Key"},
			wantName:     "Renamed Key",
			wantDisabled: false,
		},
		{
			name:         "disable ingest key",
			id:           "hcxik_01abc",
			args:         []string{"--disabled"},
			wantName:     "My Key",
			wantDisabled: true,
		},
		{
			name:         "enable ingest key",
			id:           "hcxik_01abc",
			args:         []string{"--enabled"},
			wantName:     "My Key",
			wantDisabled: false,
		},
		{
			name:         "rename configuration key",
			id:           "hcxlk_02def",
			args:         []string{"--name", "Renamed Config Key"},
			wantName:     "Renamed Config Key",
			wantDisabled: false,
		},
		{
			name:         "disable configuration key",
			id:           "hcxlk_02def",
			args:         []string{"--disabled"},
			wantName:     "My Config Key",
			wantDisabled: true,
		},
		{
			name:         "rename and disable",
			id:           "hcxik_01abc",
			args:         []string{"--name", "New Name", "--disabled"},
			wantName:     "New Name",
			wantDisabled: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("method = %q, want PATCH", r.Method)
				}

				if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
					t.Errorf("Content-Type = %q, want application/vnd.api+json", ct)
				}

				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode request body: %v", err)
				}

				data := body["data"].(map[string]any)
				if data["id"] != tc.id {
					t.Errorf("data.id = %v, want %q", data["id"], tc.id)
				}
				if data["type"] != "api-keys" {
					t.Errorf("data.type = %v, want %q", data["type"], "api-keys")
				}

				keyType := "ingest"
				if strings.HasPrefix(tc.id, "hcxlk_") {
					keyType = "configuration"
				}

				w.Header().Set("Content-Type", "application/vnd.api+json")
				_, _ = w.Write([]byte(fmt.Sprintf(`{
					"data": {
						"id": %q,
						"type": "api-keys",
						"attributes": {
							"name": %q,
							"key_type": %q,
							"disabled": %t
						}
					}
				}`, tc.id, tc.wantName, keyType, tc.wantDisabled)))
			}))

			cmd := NewCmd(opts)
			cmd.SetArgs(append([]string{"--team", "my-team", "update", tc.id}, tc.args...))
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}

			var detail keyDetail
			if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
				t.Fatalf("unmarshal output: %v", err)
			}
			if detail.Name != tc.wantName {
				t.Errorf("Name = %q, want %q", detail.Name, tc.wantName)
			}
			if detail.Disabled != tc.wantDisabled {
				t.Errorf("Disabled = %v, want %v", detail.Disabled, tc.wantDisabled)
			}
		})
	}
}

func TestUpdate_NoFlags(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--team", "my-team", "update", "hcxik_01abc"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for no flags")
	}
	if !strings.Contains(err.Error(), "--file, --name, --disabled, or --enabled is required") {
		t.Errorf("error = %q, want required flags message", err.Error())
	}
}

func TestUpdate_MutuallyExclusive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--team", "my-team", "update", "hcxik_01abc", "--disabled", "--enabled"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(err.Error(), "if any flags in the group [disabled enabled] are set none of the others can be") {
		t.Errorf("error = %q, want mutual exclusion message", err.Error())
	}
}

func TestUpdate_UnrecognizedKeyPrefix(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--team", "my-team", "update", "unknown_01abc", "--name", "New Name"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unrecognized key prefix")
	}
	if !strings.Contains(err.Error(), "unrecognized key ID prefix") {
		t.Errorf("error = %q, want unrecognized prefix message", err.Error())
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
