package main

import (
	"github.com/ericdahl-dev/coauthor-cleaner/internal/tui"
	"github.com/spf13/cobra"
)

func runTUI(cmd *cobra.Command, args []string) error {
	g, err := gitRunner()
	if err != nil {
		return err
	}
	return tui.Run(g, scanOptsFromGit(g))
}
