package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFile = "config.json"

type Config struct {
	APIUrl        string              `json:"api_url,omitempty"`
	MCPUrl        string              `json:"mcp_url,omitempty"`
	ActiveProfile string              `json:"active_profile,omitempty"`
	Profiles      map[string]*Profile `json:"profiles,omitempty"`
}

type Profile struct {
	APIUrl string `json:"api_url,omitempty"`
	MCPUrl string `json:"mcp_url,omitempty"`
	Team   string `json:"team,omitempty"`
}

func DefaultDir() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "honeycomb")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "honeycomb")
}

func DefaultPath() string {
	return filepath.Join(DefaultDir(), configFile)
}

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

func (c *Config) EnsureProfile(name string) *Profile {
	if c.Profiles == nil {
		c.Profiles = make(map[string]*Profile)
	}
	if c.Profiles[name] == nil {
		c.Profiles[name] = &Profile{}
	}
	return c.Profiles[name]
}

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
