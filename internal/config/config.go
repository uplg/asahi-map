// Package config handles application configuration loading and management.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Layout         string `yaml:"layout"`
	LogLevel       string `yaml:"log_level"`
	KeyboardDevice string `yaml:"keyboard_device"`
	ConfigDir      string `yaml:"-"`
}

func DefaultConfig() *Config {
	return &Config{
		Layout:         "azerty-mac",
		LogLevel:       "info",
		KeyboardDevice: "auto",
	}
}

// Load reads configuration from the specified path or default locations.
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	// Search paths in order of priority
	searchPaths := []string{}

	if configPath != "" {
		searchPaths = append(searchPaths, configPath)
	}

	// User config directory (use SUDO_USER if running as root via sudo)
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		searchPaths = append(searchPaths, filepath.Join("/home", sudoUser, ".config", "asahi-map", "config.yaml"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths, filepath.Join(home, ".config", "asahi-map", "config.yaml"))
	}

	// Executable directory (for portable usage)
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		searchPaths = append(searchPaths, filepath.Join(exeDir, "configs", "config.yaml"))
	}

	// System config directory
	searchPaths = append(searchPaths, "/etc/asahi-map/config.yaml")

	var loadedPath string
	for _, path := range searchPaths {
		if data, err := os.ReadFile(path); err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("parsing config %s: %w", path, err)
			}
			loadedPath = path
			break
		}
	}

	// Set config directory based on loaded file or default
	if loadedPath != "" {
		cfg.ConfigDir = filepath.Dir(loadedPath)
	} else {
		// Fallback: use executable directory
		if exe, err := os.Executable(); err == nil {
			cfg.ConfigDir = filepath.Join(filepath.Dir(exe), "configs")
		} else if home, err := os.UserHomeDir(); err == nil {
			cfg.ConfigDir = filepath.Join(home, ".config", "asahi-map")
		} else {
			cfg.ConfigDir = "/etc/asahi-map"
		}
	}

	return cfg, nil
}

func (c *Config) LayoutPath(layoutName string) string {
	return filepath.Join(c.ConfigDir, "layouts", layoutName+".yaml")
}

func (c *Config) AvailableLayouts() ([]string, error) {
	layoutDir := filepath.Join(c.ConfigDir, "layouts")
	entries, err := os.ReadDir(layoutDir)
	if err != nil {
		return nil, fmt.Errorf("reading layouts directory: %w", err)
	}

	var layouts []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
			name := entry.Name()
			layouts = append(layouts, name[:len(name)-5])
		}
	}

	return layouts, nil
}

func (c *Config) Save() error {
	configPath := filepath.Join(c.ConfigDir, "config.yaml")

	if err := os.MkdirAll(c.ConfigDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}
