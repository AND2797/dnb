package main

import (
	"slices"
	"testing"

	"github.com/AND2797/dnb/cmd/internal"
)

func TestParseArgErrors(t *testing.T) {
	// Scenario: parse is called with malformed or unrecognised argument lists.
	// Expectation: every case returns a non-nil error instead of silently succeeding.
	cfg := internal.Config{Notebooks: []string{"daily"}}

	tests := []struct {
		name string
		args []string
	}{
		{"no args", nil},
		{"open without notebook", []string{"open"}},
		{"unknown command", []string{"frobnicate"}},
		{"open unknown notebook", []string{"open", "missing"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parse(tt.args, cfg); err == nil {
				t.Errorf("parse(%v) = nil, want error", tt.args)
			}
		})
	}
}

func TestEditor(t *testing.T) {
	tests := []struct {
		name   string
		editor string
		want   []string
	}{
		// Scenario: $EDITOR is unset/empty.
		// Expectation: falls back to the default editor, vim.
		{"empty falls back to vim", "", []string{"vim"}},
		// Scenario: $EDITOR is only whitespace.
		// Expectation: still falls back to vim rather than an empty command.
		{"whitespace falls back to vim", "   ", []string{"vim"}},
		// Scenario: $EDITOR is a bare command with no flags.
		// Expectation: returns that single command.
		{"bare command", "nano", []string{"nano"}},
		// Scenario: $EDITOR includes flags, e.g. "code -w".
		// Expectation: command is split into binary plus its arguments.
		{"command with flags", "code -w", []string{"code", "-w"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("EDITOR", tt.editor)
			if got := editor(); !slices.Equal(got, tt.want) {
				t.Errorf("editor() = %v, want %v", got, tt.want)
			}
		})
	}
}
