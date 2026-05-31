package internal

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	NotebookRoot string   `yaml:"notebook_root"`
	Notebooks    []string `yaml:"notebooks"`
}

// GetConfig loads the config from ~/.dnbconf/config.yaml.
func GetConfig() (Config, error) {
	usr, err := user.Current()
	if err != nil {
		return Config{}, fmt.Errorf("looking up current user: %w", err)
	}

	configPath := filepath.Join(usr.HomeDir, ".dnbconf", "config.yaml")
	return loadConfig(configPath)
}

// loadConfig reads and parses a config file at the given path.
func loadConfig(configPath string) (Config, error) {
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("reading config %s: %w", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return Config{}, fmt.Errorf("parsing config %s: %w", configPath, err)
	}

	return config, nil
}
