package auth

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
)

func TestAuthLogout_SingleKeyType(t *testing.T) {
	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:   &config.Config{},
		Format:   "json",
	}

	if err := config.SetKey("default", config.KeyConfig, "test-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })

	if err := runAuthLogout(opts, "config"); err != nil {
		t.Fatal(err)
	}

	var results []logoutResult
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Type != "config" {
		t.Errorf("type = %q, want %q", results[0].Type, "config")
	}

	if _, err := config.GetKey("default", config.KeyConfig); err == nil {
		t.Error("expected key to be deleted")
	}
}

func TestAuthLogout_AllKeys(t *testing.T) {
	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:   &config.Config{},
		Format:   "json",
	}

	if err := config.SetKey("default", config.KeyConfig, "cfg-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })
	if err := config.SetKey("default", config.KeyIngest, "ingest-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyIngest) })

	if err := runAuthLogout(opts, ""); err != nil {
		t.Fatal(err)
	}

	var results []logoutResult
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Type != "config" {
		t.Errorf("results[0].type = %q, want %q", results[0].Type, "config")
	}
	if results[1].Type != "ingest" {
		t.Errorf("results[1].type = %q, want %q", results[1].Type, "ingest")
	}
}

func TestAuthLogout_NoKeys(t *testing.T) {
	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:   &config.Config{},
		Format:   "json",
	}

	err := runAuthLogout(opts, "")
	if err == nil {
		t.Fatal("expected error for no keys")
	}
	want := `no keys configured for profile "default"`
	if err.Error() != want {
		t.Errorf("got error %q, want %q", err.Error(), want)
	}
}

func TestAuthLogout_KeyTypeNotFound(t *testing.T) {
	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:   &config.Config{},
		Format:   "json",
	}

	if err := config.SetKey("default", config.KeyConfig, "cfg-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })

	err := runAuthLogout(opts, "ingest")
	if err == nil {
		t.Fatal("expected error for key type not found")
	}
	want := `no keys configured for profile "default"`
	if err.Error() != want {
		t.Errorf("got error %q, want %q", err.Error(), want)
	}
}

func TestAuthLogout_InvalidKeyType(t *testing.T) {
	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:   &config.Config{},
		Format:   "json",
	}

	err := runAuthLogout(opts, "bogus")
	if err == nil {
		t.Fatal("expected error for invalid key type")
	}
	if !strings.Contains(err.Error(), `invalid key type "bogus"`) {
		t.Errorf("got error %q, want it to contain %q", err.Error(), `invalid key type "bogus"`)
	}
}

func TestAuthLogout_NonDefaultProfile(t *testing.T) {
	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:   &config.Config{},
		Format:   "json",
		Profile:  "staging",
	}

	if err := config.SetKey("staging", config.KeyConfig, "staging-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("staging", config.KeyConfig) })

	if err := config.SetKey("default", config.KeyConfig, "default-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })

	if err := runAuthLogout(opts, ""); err != nil {
		t.Fatal(err)
	}

	if _, err := config.GetKey("staging", config.KeyConfig); err == nil {
		t.Error("expected staging key to be deleted")
	}

	if _, err := config.GetKey("default", config.KeyConfig); err != nil {
		t.Error("expected default key to remain")
	}
}
