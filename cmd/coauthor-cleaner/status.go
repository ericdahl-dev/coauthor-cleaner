package main

import (
	"fmt"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/guide"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/scan"
	"github.com/spf13/cobra"
)

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show repo health, findings, and recommended next steps",
		Long: `One-stop check: branch sync, staged changes, attribution markers,
and plain-language guidance on what to run next.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := gitRunner()
			if err != nil {
				return err
			}
			opts := scanOptsFromGit(g)
			opts.Staged = true
			opts.Commit = "HEAD"
			result, err := scan.Run(g, opts)
			if err != nil {
				return err
			}
			st, err := g.RepoState()
			if err != nil {
				return err
			}
			fmt.Print(guide.StatusReport(st, result.Findings))
			return nil
		},
	}
}
