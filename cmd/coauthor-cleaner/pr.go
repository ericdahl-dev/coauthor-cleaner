package main

import (
	"fmt"

	"github.com/Skeyelab/coauthor-cleaner/internal/github"
	"github.com/Skeyelab/coauthor-cleaner/internal/scan"
	"github.com/spf13/cobra"
)

func prCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Scan and clean the current GitHub pull request",
	}
	cmd.AddCommand(prScanCmd())
	cmd.AddCommand(prCleanCmd())
	return cmd
}

func prScanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan the current PR body for AI attribution",
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := gitRunner()
			if err != nil {
				return err
			}
			root, err := g.RepoRoot()
			if err != nil {
				return err
			}
			client := github.Client{Dir: root}
			if !client.Available() {
				return fmt.Errorf("gh CLI not found — install from https://cli.github.com")
			}
			pr, err := client.ViewPR()
			if err != nil {
				return err
			}
			cfg := loadConfig(g)
			findings := github.ScanPR(cfg, pr, flagStrict, flagAggressive)
			return printScanResult(findings)
		},
	}
	addScanFlags(cmd)
	return cmd
}

func prCleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean AI attribution from the current PR body",
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := gitRunner()
			if err != nil {
				return err
			}
			root, err := g.RepoRoot()
			if err != nil {
				return err
			}
			client := github.Client{Dir: root}
			if !client.Available() {
				return fmt.Errorf("gh CLI not found — install from https://cli.github.com")
			}
			pr, err := client.ViewPR()
			if err != nil {
				return err
			}
			cfg := loadConfig(g)
			findings := github.ScanPR(cfg, pr, flagStrict, flagAggressive)
			if len(findings) == 0 {
				fmt.Println("No AI attribution markers found in PR.")
				return nil
			}
			if !flagYes {
				fmt.Println("Use --yes to apply, or run `coauthor-cleaner pr scan` first.")
				fmt.Print(scan.FormatText(findings))
				return nil
			}
			title, body := github.CleanPR(pr, findings)
			if err := client.EditBody(body); err != nil {
				return err
			}
			if title != pr.Title {
				if err := client.EditTitle(title); err != nil {
					return err
				}
			}
			fmt.Printf("Cleaned %d attribution marker(s) from PR.\n", len(findings))
			return nil
		},
	}
	addScanFlags(cmd)
	cmd.Flags().BoolVar(&flagYes, "yes", false, "apply cleanups without confirmation")
	return cmd
}
