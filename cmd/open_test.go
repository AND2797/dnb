package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AND2797/dnb/cmd/internal"
)

func mustParse(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("20060102", s)
	if err != nil {
		t.Fatalf("parsing date %q: %v", s, err)
	}
	return d
}

func TestGetTodaysFile(t *testing.T) {
	// Scenario: resolve today's file for a given day under an empty root.
	// Expectation: returns the YYYY/MM/YYYYMMDD.txt path and creates its parent dir.
	root := t.TempDir()
	day := mustParse(t, "20260531")

	got, err := getTodaysFile(root, day)
	if err != nil {
		t.Fatalf("getTodaysFile: %v", err)
	}

	want := filepath.Join(root, "2026", "05", "20260531.txt")
	if got != want {
		t.Errorf("path = %q, want %q", got, want)
	}

	// Parent directory should have been created.
	if _, err := os.Stat(filepath.Dir(got)); err != nil {
		t.Errorf("parent dir not created: %v", err)
	}
}

func TestFindLatestFile(t *testing.T) {
	// Scenario: several dated files (and a non-date file) spread across month dirs.
	// Expectation: returns the most recent file strictly before the given day.
	root := t.TempDir()
	// Spread files across different month directories.
	for _, name := range []string{
		"2026/04/20260410.txt",
		"2026/05/20260501.txt",
		"2026/05/20260529.txt",
		"2026/05/not-a-date.txt",
	} {
		p := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("latest before today", func(t *testing.T) {
		// Scenario: look back from 20260531.
		// Expectation: the newest earlier file, 20260529.txt, is chosen.
		got, err := findLatestFile(root, mustParse(t, "20260531"))
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(root, "2026", "05", "20260529.txt")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("excludes the day itself and future", func(t *testing.T) {
		// Scenario: look back from 20260529, which has its own file.
		// Expectation: that day is excluded (not strictly before), so 20260501.txt wins.
		got, err := findLatestFile(root, mustParse(t, "20260529"))
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(root, "2026", "05", "20260501.txt")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("none found", func(t *testing.T) {
		// Scenario: look back from a day earlier than every file on disk.
		// Expectation: returns an empty path with no error.
		got, err := findLatestFile(root, mustParse(t, "20260101"))
		if err != nil {
			t.Fatal(err)
		}
		if got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("missing base path", func(t *testing.T) {
		// Scenario: the notebook directory does not exist yet.
		// Expectation: returns an empty path with no error (treated as "no files").
		got, err := findLatestFile(filepath.Join(root, "nope"), mustParse(t, "20260531"))
		if err != nil {
			t.Fatalf("expected nil error for missing dir, got %v", err)
		}
		if got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}

func TestStripHeader(t *testing.T) {
	// Scenario: content that may or may not begin with a date-header line.
	// Expectation: a leading header is removed; anything else is returned unchanged.
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "strips leading date header",
			in:   "Sunday, May 31, 2026 \n- todo one\n- todo two\n",
			want: "- todo one\n- todo two\n",
		},
		{
			name: "leaves non-header content untouched",
			in:   "- just a note\nmore\n",
			want: "- just a note\nmore\n",
		},
		{
			name: "header only, no newline",
			in:   "Sunday, May 31, 2026 ",
			want: "",
		},
		{
			name: "empty content",
			in:   "",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(stripHeader([]byte(tt.in))); got != tt.want {
				t.Errorf("stripHeader(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRollOverPrevious(t *testing.T) {
	t.Run("no previous file creates empty file", func(t *testing.T) {
		// Scenario: roll over when no earlier notebook file exists.
		// Expectation: today's file is created empty.
		root := t.TempDir()
		day := mustParse(t, "20260531")
		todays, _ := getTodaysFile(root, day)

		if err := rollOverPrevious(root, todays, day); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(todays)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Errorf("expected empty file, got %q", got)
		}
	})

	t.Run("carries over content under a fresh single header", func(t *testing.T) {
		// Scenario: a previous day's file (with its own header) exists.
		// Expectation: its body is carried over beneath one fresh header for today.
		root := t.TempDir()

		// Seed a previous day's file that already has its own header.
		prevDay := mustParse(t, "20260530")
		prev, _ := getTodaysFile(root, prevDay)
		prevContent := "Saturday, May 30, 2026 \n- carry me over\n"
		if err := os.WriteFile(prev, []byte(prevContent), 0644); err != nil {
			t.Fatal(err)
		}

		day := mustParse(t, "20260531")
		todays, _ := getTodaysFile(root, day)
		if err := rollOverPrevious(root, todays, day); err != nil {
			t.Fatal(err)
		}

		got, err := os.ReadFile(todays)
		if err != nil {
			t.Fatal(err)
		}
		want := "Sunday, May 31, 2026 \n- carry me over\n"
		if string(got) != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestOpen(t *testing.T) {
	root := t.TempDir()
	config := internal.Config{
		NotebookRoot: root,
		Notebooks:    []string{"daily"},
	}

	t.Run("unknown notebook returns error", func(t *testing.T) {
		// Scenario: open a notebook not listed in the config.
		// Expectation: returns an error.
		if _, err := Open("missing", config); err == nil {
			t.Error("expected error for unknown notebook")
		}
	})

	t.Run("known notebook returns todays file under notebook dir", func(t *testing.T) {
		// Scenario: open a notebook that exists in the config.
		// Expectation: today's file is created and lives under the notebook directory.
		path, err := Open("daily", config)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(path); err != nil {
			t.Errorf("file not created: %v", err)
		}
		wantPrefix := filepath.Join(root, "daily")
		if rel, err := filepath.Rel(wantPrefix, path); err != nil || filepath.IsAbs(rel) {
			t.Errorf("path %q not under notebook dir %q", path, wantPrefix)
		}
	})
}
