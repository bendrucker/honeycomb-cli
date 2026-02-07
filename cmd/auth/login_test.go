package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
)

func TestAuthLogin_ConfigKey_Valid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/auth" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("X-Honeycomb-Team") != "myid:mysecret" {
			w.WriteHeader(http.StatusUnauthorized)
			writeJSON(t, w, map[string]string{"error": "unauthorized"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSON(t, w, map[string]any{
			"id":          "myid",
			"type":        "configuration",
			"team":        map[string]string{"name": "My Team", "slug": "my-team"},
			"environment": map[string]string{"name": "production", "slug": "production"},
			"api_key_access": map[string]bool{
				"events": true,
			},
		})
	}))
	t.Cleanup(srv.Close)

	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    srv.URL,
		Format:    "json",
	}

	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })

	err := runAuthLogin(t.Context(), opts, "config", "myid", "mysecret", true)
	if err != nil {
		t.Fatal(err)
	}

	var result loginResult
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result.Type != "config" {
		t.Errorf("type = %q, want %q", result.Type, "config")
	}
	if result.Team != "My Team" {
		t.Errorf("team = %q, want %q", result.Team, "My Team")
	}
	if result.Environment != "production" {
		t.Errorf("environment = %q, want %q", result.Environment, "production")
	}
	if !result.Verified {
		t.Error("verified = false, want true")
	}

	stored, err := config.GetKey("default", config.KeyConfig)
	if err != nil {
		t.Fatal(err)
	}
	if stored != "myid:mysecret" {
		t.Errorf("stored key = %q, want %q", stored, "myid:mysecret")
	}
}

func TestAuthLogin_ConfigKey_Invalid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(t, w, map[string]string{"error": "unauthorized"})
	}))
	t.Cleanup(srv.Close)

	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    srv.URL,
		Format:    "json",
	}

	err := runAuthLogin(t.Context(), opts, "config", "badid", "badsecret", true)
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
	if err.Error() != "invalid key" {
		t.Errorf("got error %q, want %q", err.Error(), "invalid key")
	}

	_, err = config.GetKey("default", config.KeyConfig)
	if err == nil {
		t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })
		t.Error("key should not be stored after failed verification")
	}
}

func TestAuthLogin_ManagementKey_Valid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2/auth" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer mgmtid:mgmtsecret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		writeJSON(t, w, map[string]any{
			"data": map[string]any{
				"id":   "mgmt-id",
				"type": "api-keys",
				"attributes": map[string]any{
					"name":     "My Management Key",
					"key_type": "management",
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    srv.URL,
		Format:    "json",
	}

	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyManagement) })

	err := runAuthLogin(t.Context(), opts, "management", "mgmtid", "mgmtsecret", true)
	if err != nil {
		t.Fatal(err)
	}

	var result loginResult
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result.Type != "management" {
		t.Errorf("type = %q, want %q", result.Type, "management")
	}
	if result.Name != "My Management Key" {
		t.Errorf("name = %q, want %q", result.Name, "My Management Key")
	}
	if !result.Verified {
		t.Error("verified = false, want true")
	}

	stored, err := config.GetKey("default", config.KeyManagement)
	if err != nil {
		t.Fatal(err)
	}
	if stored != "mgmtid:mgmtsecret" {
		t.Errorf("stored key = %q, want %q", stored, "mgmtid:mgmtsecret")
	}
}

func TestAuthLogin_NoVerify(t *testing.T) {
	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		Format:    "json",
	}

	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyIngest) })

	err := runAuthLogin(t.Context(), opts, "ingest", "myid", "mysecret", false)
	if err != nil {
		t.Fatal(err)
	}

	var result loginResult
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result.Type != "ingest" {
		t.Errorf("type = %q, want %q", result.Type, "ingest")
	}
	if result.Verified {
		t.Error("verified = true, want false")
	}

	stored, err := config.GetKey("default", config.KeyIngest)
	if err != nil {
		t.Fatal(err)
	}
	if stored != "myid:mysecret" {
		t.Errorf("stored key = %q, want %q", stored, "myid:mysecret")
	}
}

func TestAuthLogin_MissingKeyType_NonInteractive(t *testing.T) {
	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		Format:    "json",
	}

	err := runAuthLogin(t.Context(), opts, "", "myid", "mysecret", false)
	if err == nil {
		t.Fatal("expected error for missing key type")
	}
	want := "--key-type is required in non-interactive mode"
	if err.Error() != want {
		t.Errorf("got error %q, want %q", err.Error(), want)
	}
}

func TestAuthLogin_StdinSecret(t *testing.T) {
	ts := iostreams.Test()
	ts.InBuf.WriteString("stdin-secret\n")

	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		Format:    "json",
	}

	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })

	err := runAuthLogin(t.Context(), opts, "config", "myid", "", false)
	if err != nil {
		t.Fatal(err)
	}

	stored, err := config.GetKey("default", config.KeyConfig)
	if err != nil {
		t.Fatal(err)
	}
	if stored != "myid:stdin-secret" {
		t.Errorf("stored key = %q, want %q", stored, "myid:stdin-secret")
	}
}
