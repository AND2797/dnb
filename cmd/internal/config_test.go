package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
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
	_, err := loadConfig(filepath.Join(t.TempDir(), "nope.yaml"))
	if err == nil {
		t.Error("expected error for missing config file")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("notebooks: [unterminated\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(path); err == nil {
		t.Error("expected parse error for invalid yaml")
	}
}
