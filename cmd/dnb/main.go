package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

func expandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		usr, _ := user.Current()
		return filepath.Join(usr.HomeDir, path[1:])
	}
	return path
}

func readBasePathFromConfig() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(usr.HomeDir, ".dnbconf", "config.txt")
	file, err := os.Open(configPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "nbrootdir=") {
			basepath := strings.TrimSpace(strings.TrimPrefix(line, "nbrootdir="))
			return basepath, nil
		}
	}
	return "", fmt.Errorf("nbrootdir not found in config")
}

func main() {
	basePath, err := readBasePathFromConfig()
	if err != nil {
		fmt.Println("Error reading nbrootdir from config:", err)
		os.Exit(1)
	}

	basePath = expandHome(basePath)

	today := time.Now()
	year := today.Format("2006")
	month := today.Format("01")
	day := today.Format("20060102")

	dirPath := filepath.Join(basePath, year, month)
	os.MkdirAll(dirPath, 0755)

	filePath := filepath.Join(dirPath, day+".txt")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, _ := os.Create(filePath)
		file.Close()
	}

	cmd := exec.Command("vim", filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
