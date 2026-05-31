package main

import (
	"testing"

	"github.com/AND2797/dnb/cmd/internal"
)

func TestParseArgErrors(t *testing.T) {
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

func TestEditorDefault(t *testing.T) {
	t.Setenv("EDITOR", "")
	if got := editor(); got != "vim" {
		t.Errorf("editor() = %q, want vim", got)
	}
	t.Setenv("EDITOR", "nano")
	if got := editor(); got != "nano" {
		t.Errorf("editor() = %q, want nano", got)
	}
}
