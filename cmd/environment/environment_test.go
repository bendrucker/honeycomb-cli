package environment

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

	ts := iostreams.Test()
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

func jsonapiEnvelope(t *testing.T, w http.ResponseWriter, status int, data any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func TestList(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2/teams/my-team/environments" {
			t.Errorf("path = %q, want /2/teams/my-team/environments", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-mgmt-key" {
			t.Errorf("auth = %q, want Bearer header", r.Header.Get("Authorization"))
		}
		jsonapiEnvelope(t, w, http.StatusOK, map[string]any{
			"data": []map[string]any{
				{
					"id":   "env-1",
					"type": "environments",
					"attributes": map[string]any{
						"name":        "Production",
						"slug":        "production",
						"description": "Prod environment",
						"color":       "blue",
						"settings":    map[string]any{"delete_protected": true},
					},
				},
				{
					"id":   "env-2",
					"type": "environments",
					"attributes": map[string]any{
						"name":     "Staging",
						"slug":     "staging",
						"color":    "green",
						"settings": map[string]any{"delete_protected": false},
					},
				},
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list", "--team", "my-team"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []environmentItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].Name != "Production" {
		t.Errorf("items[0].Name = %q, want %q", items[0].Name, "Production")
	}
	if items[0].ID != "env-1" {
		t.Errorf("items[0].ID = %q, want %q", items[0].ID, "env-1")
	}
	if items[0].Color != "blue" {
		t.Errorf("items[0].Color = %q, want %q", items[0].Color, "blue")
	}
	if items[1].Slug != "staging" {
		t.Errorf("items[1].Slug = %q, want %q", items[1].Slug, "staging")
	}
}

func TestList_Empty(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonapiEnvelope(t, w, http.StatusOK, map[string]any{"data": []any{}})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list", "--team", "my-team"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []environmentItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

func TestList_NoKey(t *testing.T) {
	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    "http://localhost",
		Format:    "json",
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"list", "--team", "my-team"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "no management key configured") {
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
	cmd.SetArgs([]string{"list", "--team", "my-team"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if !strings.Contains(err.Error(), "HTTP 401") {
		t.Errorf("error = %q, want HTTP 401", err.Error())
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
		t.Errorf("error = %q, want required flag error", err.Error())
	}
}

func TestGet(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2/teams/my-team/environments/env-1" {
			t.Errorf("path = %q, want /2/teams/my-team/environments/env-1", r.URL.Path)
		}
		jsonapiEnvelope(t, w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"id":   "env-1",
				"type": "environments",
				"attributes": map[string]any{
					"name":        "Production",
					"slug":        "production",
					"description": "Prod environment",
					"color":       "blue",
					"settings":    map[string]any{"delete_protected": true},
				},
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "env-1", "--team", "my-team"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail environmentDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "env-1" {
		t.Errorf("ID = %q, want %q", detail.ID, "env-1")
	}
	if detail.Name != "Production" {
		t.Errorf("Name = %q, want %q", detail.Name, "Production")
	}
	if detail.Slug != "production" {
		t.Errorf("Slug = %q, want %q", detail.Slug, "production")
	}
	if !detail.DeleteProtected {
		t.Error("DeleteProtected = false, want true")
	}
}

func TestGet_NotFound(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"environment not found"}`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"get", "missing", "--team", "my-team"})
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
	cmd.SetArgs([]string{"get", "--team", "my-team"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}

func TestDelete_WithYes(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2/teams/my-team/environments/env-1" {
			t.Errorf("path = %q, want /2/teams/my-team/environments/env-1", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete", "env-1", "--team", "my-team", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var result map[string]string
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result["id"] != "env-1" {
		t.Errorf("id = %q, want %q", result["id"], "env-1")
	}
}

func TestDelete_NoYesNonInteractive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	opts.IOStreams.SetNeverPrompt(true)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"delete", "env-1", "--team", "my-team"})
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
	cmd.SetArgs([]string{"delete", "--team", "my-team"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}

func TestCreate(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2/teams/my-team/environments" {
			t.Errorf("path = %q, want /2/teams/my-team/environments", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		data, _ := body["data"].(map[string]any)
		attrs, _ := data["attributes"].(map[string]any)
		if attrs["name"] != "Test Env" {
			t.Errorf("name = %q, want %q", attrs["name"], "Test Env")
		}
		jsonapiEnvelope(t, w, http.StatusCreated, map[string]any{
			"data": map[string]any{
				"id":   "env-new",
				"type": "environments",
				"attributes": map[string]any{
					"name":        "Test Env",
					"slug":        "test-env",
					"description": "A test environment",
					"color":       "blue",
					"settings":    map[string]any{"delete_protected": false},
				},
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--team", "my-team", "--name", "Test Env", "--description", "A test environment", "--color", "blue"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail environmentDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.ID != "env-new" {
		t.Errorf("ID = %q, want %q", detail.ID, "env-new")
	}
	if detail.Name != "Test Env" {
		t.Errorf("Name = %q, want %q", detail.Name, "Test Env")
	}
}

func TestCreate_NoNameNonInteractive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	opts.IOStreams.SetNeverPrompt(true)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"create", "--team", "my-team"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --name")
	}
	if !strings.Contains(err.Error(), "--name is required in non-interactive mode") {
		t.Errorf("error = %q, want non-interactive error", err.Error())
	}
}

func TestUpdate(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2/teams/my-team/environments/env-1" {
			t.Errorf("path = %q, want /2/teams/my-team/environments/env-1", r.URL.Path)
		}
		if r.Method != http.MethodPatch {
			t.Errorf("method = %q, want PATCH", r.Method)
		}
		jsonapiEnvelope(t, w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"id":   "env-1",
				"type": "environments",
				"attributes": map[string]any{
					"name":        "Production",
					"slug":        "production",
					"description": "Updated description",
					"color":       "red",
					"settings":    map[string]any{"delete_protected": false},
				},
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "env-1", "--team", "my-team", "--description", "Updated description", "--color", "red"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var detail environmentDetail
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if detail.Description != "Updated description" {
		t.Errorf("Description = %q, want %q", detail.Description, "Updated description")
	}
}

func TestUpdate_NoFlags(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "env-1", "--team", "my-team"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for no flags")
	}
	if !strings.Contains(err.Error(), "--description, --color, or --delete-protected is required") {
		t.Errorf("error = %q, want missing flag error", err.Error())
	}
}
