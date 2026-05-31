package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/AND2797/dnb/cmd"
	"github.com/AND2797/dnb/cmd/internal"
)

// editor returns the editor to launch, honoring $EDITOR and falling back to vim.
func editor() string {
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	return "vim"
}

func parse(args []string, config internal.Config) error {
	// TODO: just using a basic parser for now as there aren't many commands.
	// If required I might look into spf13/cobra but it's not required for now

	if len(args) == 0 {
		return fmt.Errorf("usage: dnb <list|open <notebook>>")
	}

	switch args[0] {
	case "list":
		cmd.List(config)
		return nil
	case "open":
		if len(args) < 2 {
			return fmt.Errorf("usage: dnb open <notebook>")
		}
		nb, err := cmd.Open(args[1], config)
		if err != nil {
			return err
		}
		ed := exec.Command(editor(), nb)
		ed.Stdin = os.Stdin
		ed.Stdout = os.Stdout
		ed.Stderr = os.Stderr
		return ed.Run()
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func main() {
	config, err := internal.GetConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	if err := parse(os.Args[1:], config); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
