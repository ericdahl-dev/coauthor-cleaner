package main

import (
	"fmt"
	"os"

	"github.com/Skeyelab/coauthor-cleaner/internal/scan"
	"github.com/spf13/cobra"
)

var (
	flagCIMode   string
	flagCIJSON   bool
	flagCIBase   string
	flagCIHead   string
	flagCIReport   string
	flagCIComment  string
)

func ciCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ci",
		Short: "Scan a PR range for CI/GitHub Actions (exits 1 in block mode when findings exist)",
		Long: `Scan the diff and commit messages between base and head.

Designed for GitHub Actions pull_request workflows. Use --mode block to fail
the job, or --mode warn to report without failing.`,
		RunE: runCI,
	}
	cmd.Flags().StringVar(&flagCIMode, "mode", envOr("COAUTHOR_CLEANER_MODE", "block"), "block or warn")
	cmd.Flags().BoolVar(&flagCIJSON, "json", false, "output findings as JSON")
	cmd.Flags().StringVar(&flagCIBase, "base", envOr("COAUTHOR_CLEANER_BASE", ""), "base ref (merge base side)")
	cmd.Flags().StringVar(&flagCIHead, "head", envOr("COAUTHOR_CLEANER_HEAD", "HEAD"), "head ref (PR branch tip)")
	cmd.Flags().StringVar(&flagCIReport, "report-file", "", "write JSON report to file")
	cmd.Flags().StringVar(&flagCIComment, "comment-file", "", "write GitHub PR comment markdown to file")
	cmd.Flags().BoolVar(&flagStrict, "strict", false, "only high-confidence patterns")
	cmd.Flags().BoolVar(&flagAggressive, "aggressive", false, "broader pattern matching")
	return cmd
}

func runCI(cmd *cobra.Command, args []string) error {
	if flagCIBase == "" {
		return fmt.Errorf("--base is required (or set COAUTHOR_CLEANER_BASE)")
	}

	g, err := gitRunner()
	if err != nil {
		return err
	}

	opts := scanOptsFromGit(g)
	opts.Base = flagCIBase
	opts.Head = flagCIHead
	result, err := scan.Run(g, opts)
	if err != nil {
		return err
	}

	if flagCIJSON || flagCIReport != "" {
		out, err := scan.FormatJSON(result.Findings)
		if err != nil {
			return err
		}
		if flagCIReport != "" {
			if err := os.WriteFile(flagCIReport, out, 0644); err != nil {
				return err
			}
		}
		if flagCIJSON {
			fmt.Println(string(out))
		}
	} else if flagCIComment == "" {
		fmt.Print(scan.FormatText(result.Findings))
	}

	if flagCIComment != "" && len(result.Findings) > 0 {
		comment := scan.FormatMarkdownComment(result.Findings)
		if err := os.WriteFile(flagCIComment, []byte(comment), 0644); err != nil {
			return err
		}
	}

	if len(result.Findings) == 0 {
		return nil
	}

	if flagCIMode == "block" {
		return fmt.Errorf("found %d AI attribution marker(s)", len(result.Findings))
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
