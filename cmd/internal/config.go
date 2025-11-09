package internal

import (
	"log"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	NotebookRoot string   `yaml:"notebook_root"`
	Notebooks    []string `yaml:"notebooks"`
}

func GetConfig() Config {

	var config Config

	usr, err := user.Current()
	if err != nil {
		// TODO: handle this better
		return config
	}

	configPath := filepath.Join(usr.HomeDir, ".dnbconf", "config.yaml")
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Error reading internal file: %v\n", err)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Error parsing internal file: %v\n", err)
	}

	return config
}
