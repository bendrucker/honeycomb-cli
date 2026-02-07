package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIUrl        string              `yaml:"api_url,omitempty"`
	ActiveProfile string              `yaml:"active_profile,omitempty"`
	Profiles      map[string]*Profile `yaml:"profiles,omitempty"`
}

type Profile struct {
	APIUrl string `yaml:"api_url,omitempty"`
}

func DefaultDir() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "honeycomb")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "honeycomb")
}

func DefaultPath() string {
	return filepath.Join(DefaultDir(), "config.yaml")
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
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
