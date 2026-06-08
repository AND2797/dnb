package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/AND2797/dnb/cmd"
	"github.com/AND2797/dnb/cmd/internal"
)

// editor returns the editor command (and any arguments) to launch, honoring
// $EDITOR and falling back to vim. $EDITOR may include flags, e.g. "code -w",
// so it is split into fields rather than treated as a single binary name.
func editor() []string {
	if fields := strings.Fields(os.Getenv("EDITOR")); len(fields) > 0 {
		return fields
	}
	return []string{"vim"}
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
		ed := editor()
		c := exec.Command(ed[0], append(ed[1:], nb)...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
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
