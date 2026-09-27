package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the CLI configuration.
type Config struct {
	URL      string // Gitea server URL, e.g. https://gitea.example.com
	Username string // Username
	Token    string // API token (optional; obtained from git credential first)
}

const (
	configDirName = ".gitea-cli"
	configName    = "config"
	configType    = "yaml"
)

// Load loads the configuration, preferring the config file first and then
// environment variables.
func Load() (*Config, error) {
	v := viper.New()

	// Config file location: ~/.gitea-cli/config.yaml
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	configDir := filepath.Join(home, configDirName)

	v.SetConfigName(configName)
	v.SetConfigType(configType)
	v.AddConfigPath(configDir)

	// Bind environment variables (higher priority than config file)
	v.SetEnvPrefix("GITEA")
	v.BindEnv("url")
	v.BindEnv("username")
	v.BindEnv("token")

	// Set defaults
	v.SetDefault("url", "")
	v.SetDefault("username", "")

	// Read the config file (missing file is not an error)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	return &Config{
		URL:      v.GetString("url"),
		Username: v.GetString("username"),
		Token:    v.GetString("token"),
	}, nil
}

// Save writes the configuration to ~/.gitea-cli/config.yaml.
func Save(c *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	configDir := filepath.Join(home, configDirName)
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	v := viper.New()
	v.Set("url", c.URL)
	v.Set("username", c.Username)
	if c.Token != "" {
		v.Set("token", c.Token)
	}

	path := filepath.Join(configDir, configName+"."+configType)
	if err := v.WriteConfigAs(path); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	// The config file may contain a token; restrict permissions
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("failed to set config file permissions: %w", err)
	}
	return nil
}
