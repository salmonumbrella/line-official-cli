package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the CLI configuration loaded from config file.
type Config struct {
	// Account is the default account name to use
	Account string `yaml:"account,omitempty"`
	// Output is the default output format (text, json, or table)
	Output string `yaml:"output,omitempty"`
	// Debug enables debug output by default
	Debug bool `yaml:"debug,omitempty"`

	// path stores where this config was loaded from (not serialized)
	path string `yaml:"-"`
}

// ConfigPath returns the path where this config was loaded from.
// Returns empty string if config was not loaded from a file.
func (c *Config) ConfigPath() string {
	return c.path
}

// configPaths returns the list of paths to check for config file,
// in order of priority (first match wins).
func configPaths() ([]string, error) {
	var paths []string

	// XDG config dir (highest priority for config file)
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		paths = append(paths, filepath.Join(dir, AppName, "config.yaml"))
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// ~/.config/line-cli/config.yaml (standard XDG location)
	paths = append(paths, filepath.Join(home, ".config", AppName, "config.yaml"))

	// ~/.line-cli.yaml (simple fallback)
	paths = append(paths, filepath.Join(home, ".line-cli.yaml"))

	return paths, nil
}

// Load reads configuration from the first config file found.
// Returns an empty Config if no config file exists.
// Returns an error only if the file exists but cannot be parsed.
func Load() (*Config, error) {
	paths, err := configPaths()
	if err != nil {
		return &Config{}, nil
	}

	for _, path := range paths {
		cfg, err := loadFromPath(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		cfg.path = path
		return cfg, nil
	}

	// No config file found - return empty config
	return &Config{}, nil
}

// loadFromPath loads config from a specific path.
func loadFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// DefaultConfigPath returns the recommended config file path for display purposes.
func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", AppName, "config.yaml"), nil
}

// ExampleConfig returns an example config file content.
func ExampleConfig() string {
	return `# LINE CLI configuration
# Place this file at ~/.config/line-cli/config.yaml or ~/.line-cli.yaml

# Default account name (can be overridden with --account or LINE_ACCOUNT)
# account: my-account

# Default output format: text, json, or table (can be overridden with --output or LINE_OUTPUT)
# output: text

# Enable debug output by default (can be overridden with --debug)
# debug: false
`
}
