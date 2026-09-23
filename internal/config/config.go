// Package config manages the CLI's local configuration file and profiles.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFile = "config.json"

// Config holds the CLI's persisted settings, including its registered profiles.
type Config struct {
	APIUrl        string              `json:"api_url,omitempty"`
	MCPUrl        string              `json:"mcp_url,omitempty"`
	ActiveProfile string              `json:"active_profile,omitempty"`
	Profiles      map[string]*Profile `json:"profiles,omitempty"`
}

// Profile holds the connection settings for a single named profile.
type Profile struct {
	APIUrl string `json:"api_url,omitempty"`
	MCPUrl string `json:"mcp_url,omitempty"`
	Team   string `json:"team,omitempty"`
}

// DefaultDir returns the directory containing the config file, honoring XDG_CONFIG_HOME.
func DefaultDir() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "honeycomb")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "honeycomb")
}

// DefaultPath returns the full path to the config file.
func DefaultPath() string {
	return filepath.Join(DefaultDir(), configFile)
}

// Load reads the config file at path, returning an empty Config if the file does not exist.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// EnsureProfile returns the named profile, creating it (and the Profiles map) if it does not exist.
func (c *Config) EnsureProfile(name string) *Profile {
	if c.Profiles == nil {
		c.Profiles = make(map[string]*Profile)
	}
	if c.Profiles[name] == nil {
		c.Profiles[name] = &Profile{}
	}
	return c.Profiles[name]
}

// Save writes c to path as indented JSON, creating parent directories as needed.
func (c *Config) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
