package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
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

func findLatestFileBefore(basePath string, today time.Time) (string, error) {
	var files []string
	var fileDates []time.Time

	re := regexp.MustCompile(`(\d{8})\.txt$`)
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil // skip directories & errors
		}
		matches := re.FindStringSubmatch(info.Name())
		if matches != nil {
			// Parse date from filename
			fileDate, err := time.Parse("20060102", matches[1])
			if err == nil && fileDate.Before(today) {
				files = append(files, path)
				fileDates = append(fileDates, fileDate)
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(fileDates) == 0 {
		return "", nil // no previous file found
	}
	// Find the max (latest) date
	idx := 0
	for i := 1; i < len(fileDates); i++ {
		if fileDates[i].After(fileDates[idx]) {
			idx = i
		}
	}
	return files[idx], nil
}

// copyFile copies the contents of the file named src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy file content from src to dst
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	// Flush file to disk
	err = out.Sync()
	if err != nil {
		return err
	}

	return nil
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
	// If today's file doesn't exist, handle rollover
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Find the latest previous file
		prevFilePath, err := findLatestFileBefore(basePath, today)
		if err == nil && prevFilePath != "" {
			// Copy previous notebook's content
			fmt.Println(fmt.Sprintf("Rolling over from %s", prevFilePath))
			if err := copyFile(prevFilePath, filePath); err != nil {
				fmt.Println("Error copying rollover file:", err)
				// Optionally create empty file if copy fails
				file, _ := os.Create(filePath)
				file.Close()
			}
		} else {
			// No previous file found, create empty notebook
			file, _ := os.Create(filePath)
			file.Close()
		}
	}

	cmd := exec.Command("vim", filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
