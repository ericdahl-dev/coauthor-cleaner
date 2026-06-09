package main

import (
	"github.com/Skeyelab/coauthor-cleaner/internal/tui"
	"github.com/spf13/cobra"
)

func runTUI(cmd *cobra.Command, args []string) error {
	g, err := gitRunner()
	if err != nil {
		return err
	}
	return tui.Run(g, scanOptsFromGit(g))
}
