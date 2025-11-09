package cmd

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AND2797/dnb/cmd/internal"
	"slices"
)

func Open(notebook string, config internal.Config) string {
	root_dir := config.NotebookRoot
	contains := slices.Contains(config.Notebooks, notebook)
	if !contains {
		fmt.Println("Notebook doesn't exist")
		return ""
	}

	notebookPath := filepath.Join(root_dir, notebook)
	notebookPath = expandHome(notebookPath)
	todaysFile := getTodaysFile(notebookPath)

	if _, err := os.Stat(todaysFile); os.IsNotExist(err) {
		rollOverPrevious(notebookPath, todaysFile, time.Now())
	} else {
		file, _ := os.Create(todaysFile)
		file.Close()
	}

	return todaysFile
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		usr, _ := user.Current()
		return filepath.Join(usr.HomeDir, path[1:])
	}
	return path
}

func getTodaysFile(notebookPath string) string {
	today := time.Now()
	year := today.Format("2006")
	month := today.Format("01")
	day := today.Format("20060102")

	dirPath := filepath.Join(notebookPath, year, month)
	err := os.MkdirAll(dirPath, 0777)
	if err != nil {
		return ""
	}

	return filepath.Join(dirPath, day+".txt")
}

func rollOverPrevious(notebookPath string, todaysFile string, withRespectTo time.Time) {
	prevFilePath, err := findLatestFile(notebookPath, withRespectTo)
	if err == nil && prevFilePath != "" {
		fmt.Println(fmt.Sprintf("Rolling over from %s", prevFilePath))

		if err := copyFile(prevFilePath, todaysFile); err != nil {
			fmt.Println("Error while copying file:", err)
			file, _ := os.Create(todaysFile)
			file.Close()
		}
		fullDate := withRespectTo.Format("Monday, January 2, 2006")
		err = writeHeader(todaysFile, fullDate)
		if err != nil {
			fmt.Println("Error writing header:", err)
		}
	} else {
		// No previous file found, create empty file
		file, _ := os.Create(todaysFile)
		file.Close()
	}

}

func findLatestFile(basePath string, today time.Time) (string, error) {
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

func writeHeader(filePath string, dateStr string) error {
	f, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Read existing content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Construct header line
	header := fmt.Sprintf("%s \n", dateStr)

	// Write header + original content back to file
	f.Truncate(0)
	f.Seek(0, 0)
	_, err = f.WriteString(header)
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	if err != nil {
		return err
	}

	return nil
}
