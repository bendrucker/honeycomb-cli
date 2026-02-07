package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissing(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ActiveProfile != "" {
		t.Fatal("expected empty active profile")
	}
}

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	cfg := &Config{
		APIUrl:        "https://api.honeycomb.io",
		ActiveProfile: "default",
		Profiles: map[string]*Profile{
			"default": {APIUrl: "https://api.honeycomb.io"},
		},
	}

	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ActiveProfile != "default" {
		t.Fatalf("expected active profile 'default', got %q", loaded.ActiveProfile)
	}
	if loaded.Profiles["default"].APIUrl != "https://api.honeycomb.io" {
		t.Fatalf("unexpected profile API URL: %q", loaded.Profiles["default"].APIUrl)
	}
}

func TestDefaultDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg")
	if got := DefaultDir(); got != "/tmp/xdg/honeycomb" {
		t.Fatalf("expected /tmp/xdg/honeycomb, got %q", got)
	}

	t.Setenv("XDG_CONFIG_HOME", "")
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "honeycomb")
	if got := DefaultDir(); got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}
