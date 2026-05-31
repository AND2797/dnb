package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/AND2797/dnb/cmd/internal"
)

const headerLayout = "Monday, January 2, 2006"

var dateFileRe = regexp.MustCompile(`(\d{8})\.txt$`)

// Open resolves the path to today's file for the given notebook, creating it
// (and rolling over the previous day's contents) if it does not yet exist.
func Open(notebook string, config internal.Config) (string, error) {
	if !slices.Contains(config.Notebooks, notebook) {
		return "", fmt.Errorf("notebook %q doesn't exist", notebook)
	}

	notebookPath := expandHome(filepath.Join(config.NotebookRoot, notebook))
	now := time.Now()

	todaysFile, err := getTodaysFile(notebookPath, now)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(todaysFile); os.IsNotExist(err) {
		if err := rollOverPrevious(notebookPath, todaysFile, now); err != nil {
			return "", err
		}
	}

	return todaysFile, nil
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return path
		}
		return filepath.Join(usr.HomeDir, path[1:])
	}
	return path
}

// getTodaysFile returns the YYYY/MM/YYYYMMDD.txt path for the given day,
// ensuring its parent directory exists.
func getTodaysFile(notebookPath string, day time.Time) (string, error) {
	dirPath := filepath.Join(notebookPath, day.Format("2006"), day.Format("01"))
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("creating %s: %w", dirPath, err)
	}
	return filepath.Join(dirPath, day.Format("20060102")+".txt"), nil
}

// rollOverPrevious creates todaysFile. If a previous day's file exists, its
// contents (minus any leading date header) are carried over beneath a fresh
// header for the current day; otherwise an empty file is created.
func rollOverPrevious(notebookPath string, todaysFile string, day time.Time) error {
	prevFilePath, err := findLatestFile(notebookPath, day)
	if err != nil {
		return err
	}

	if prevFilePath == "" {
		// No previous file found, create empty file.
		f, err := os.Create(todaysFile)
		if err != nil {
			return fmt.Errorf("creating %s: %w", todaysFile, err)
		}
		return f.Close()
	}

	fmt.Printf("Rolling over from %s\n", prevFilePath)

	prev, err := os.ReadFile(prevFilePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", prevFilePath, err)
	}

	header := fmt.Sprintf("%s \n", day.Format(headerLayout))
	out := append([]byte(header), stripHeader(prev)...)
	if err := os.WriteFile(todaysFile, out, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", todaysFile, err)
	}
	return nil
}

// stripHeader removes a leading date-header line (as written by rollOverPrevious)
// so headers don't accumulate across rollovers.
func stripHeader(content []byte) []byte {
	line, rest, found := bytes.Cut(content, []byte("\n"))
	if _, err := time.Parse(headerLayout, strings.TrimSpace(string(line))); err != nil {
		return content // first line isn't a header, leave content untouched
	}
	if !found {
		return nil
	}
	return rest
}

// findLatestFile returns the path of the most recent YYYYMMDD.txt file strictly
// before the given day, or "" if none exist.
func findLatestFile(basePath string, before time.Time) (string, error) {
	var latestPath string
	var latestDate time.Time

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil // skip directories & errors
		}
		matches := dateFileRe.FindStringSubmatch(info.Name())
		if matches == nil {
			return nil
		}
		fileDate, err := time.Parse("20060102", matches[1])
		if err != nil || !fileDate.Before(before) {
			return nil
		}
		if latestPath == "" || fileDate.After(latestDate) {
			latestPath = path
			latestDate = fileDate
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return latestPath, nil
}
