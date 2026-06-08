package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Scenario: a well-formed config.yaml with a root and two notebooks.
	// Expectation: fields are parsed into the Config struct in order.
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	yaml := "notebook_root: ~/dnb_notebooks\nnotebooks:\n  - daily\n  - personal\n"
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.NotebookRoot != "~/dnb_notebooks" {
		t.Errorf("NotebookRoot = %q", cfg.NotebookRoot)
	}
	if len(cfg.Notebooks) != 2 || cfg.Notebooks[0] != "daily" || cfg.Notebooks[1] != "personal" {
		t.Errorf("Notebooks = %v", cfg.Notebooks)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	// Scenario: the config file path does not exist.
	// Expectation: loadConfig returns an error rather than a zero Config.
	_, err := loadConfig(filepath.Join(t.TempDir(), "nope.yaml"))
	if err == nil {
		t.Error("expected error for missing config file")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	// Scenario: the config file contains malformed YAML.
	// Expectation: loadConfig surfaces a parse error.
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("notebooks: [unterminated\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(path); err == nil {
		t.Error("expected parse error for invalid yaml")
	}
}
