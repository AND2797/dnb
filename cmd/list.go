package cmd

import (
	"fmt"
	"github.com/AND2797/dnb/cmd/internal"
)

func List(config internal.Config) {
	fmt.Printf("Notebooks:\n")
	for _, v := range config.Notebooks {
		fmt.Printf("- %s\n", v)
	}
}
