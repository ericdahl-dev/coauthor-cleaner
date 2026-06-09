package main

import (
	"fmt"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/fix"
	"github.com/spf13/cobra"
)

var (
	flagCheckOnly  bool
	flagPush       bool
	flagForcePush  bool
)

func fixCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fix",
		Short: "Automatically clean safe findings (default workflow)",
		Long: `Scan staged changes and HEAD, then auto-apply safe fixes.

Safe (always):
  • Staged file attribution lines

Safe when HEAD is not on remote yet:
  • HEAD commit message / trailers

Needs --force (rewrites published history):
  • HEAD commit already pushed to remote

Optional:
  --push         git push after cleaning (safe pushes only)
  --force-push   use git push --force-with-lease (with --force --push)
  --check        dry run, show what would happen`,
		RunE: runFix,
	}
	cmd.Flags().BoolVar(&flagCheckOnly, "check", false, "dry run — show planned actions")
	cmd.Flags().BoolVar(&flagPush, "push", false, "push to remote after cleaning")
	cmd.Flags().BoolVar(&flagForcePush, "force-push", false, "push with --force-with-lease after amending a published commit")
	cmd.Flags().BoolVar(&flagForce, "force", false, "allow amending a commit already on remote")
	return cmd
}

func runFix(cmd *cobra.Command, args []string) error {
	g, err := gitRunner()
	if err != nil {
		return err
	}
	result, err := fix.Run(g, scanOptsFromGit(g), fix.Options{
		CheckOnly:  flagCheckOnly,
		ForceAmend: flagForce,
		Push:       flagPush,
		ForcePush:  flagForcePush,
	})
	if err != nil {
		return err
	}
	fmt.Print(result.Summary)
	if !flagCheckOnly && len(result.Findings) > 0 && result.Applied == 0 && result.SkippedCommit > 0 {
		return fmt.Errorf("could not auto-fix %d marker(s) without --force", result.SkippedCommit)
	}
	return nil
}
