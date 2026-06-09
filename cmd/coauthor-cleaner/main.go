package main

import (
	"fmt"
	"os"

	"github.com/Skeyelab/coauthor-cleaner/internal/clean"
	"github.com/Skeyelab/coauthor-cleaner/internal/detect"
	"github.com/Skeyelab/coauthor-cleaner/internal/git"
	"github.com/Skeyelab/coauthor-cleaner/internal/scan"
	"github.com/Skeyelab/coauthor-cleaner/internal/tui"
	"github.com/spf13/cobra"
)

var (
	flagStaged       bool
	flagCommit       string
	flagYes          bool
	flagStrict       bool
	flagAggressive   bool
	flagJSON         bool
	flagFailFindings bool
	flagBase         string
	flagHead         string
)

func main() {
	root := &cobra.Command{
		Use:   "coauthor-cleaner",
		Short: "Remove unwanted AI attribution from Git commits and staged changes",
	}

	root.AddCommand(scanCmd())
	root.AddCommand(cleanCmd())
	root.AddCommand(reviewCmd())
	root.AddCommand(ciCmd())
	root.AddCommand(configCmd())
	root.AddCommand(rulesCmd())
	root.AddCommand(hookCmd())
	root.AddCommand(prCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func gitRunner() (git.Runner, error) {
	g := git.Runner{}
	if !g.InRepo() {
		return g, fmt.Errorf("not inside a git repository")
	}
	return g, nil
}

func scanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan for AI attribution markers without making changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := gitRunner()
			if err != nil {
				return err
			}
			opts := scanOptsFromGit(g)
			if !opts.Staged && opts.Commit == "" && opts.Base == "" {
				opts.Staged = true
				opts.Commit = "HEAD"
			}
			result, err := scan.Run(g, opts)
			if err != nil {
				return err
			}
			if err := printScanResult(result.Findings); err != nil {
				return err
			}
			if flagFailFindings && len(result.Findings) > 0 {
				return fmt.Errorf("found %d AI attribution marker(s)", len(result.Findings))
			}
			return nil
		},
	}
	addScanFlags(cmd)
	return cmd
}

func cleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove selected AI attribution markers",
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := gitRunner()
			if err != nil {
				return err
			}
			opts := scanOptsFromGit(g)
			if !opts.Staged && opts.Commit == "" && opts.Base == "" {
				opts.Staged = true
			}
			result, err := scan.Run(g, opts)
			if err != nil {
				return err
			}
			if len(result.Findings) == 0 {
				fmt.Println("No AI attribution markers found.")
				return nil
			}
			if !flagYes {
				fmt.Println("Use --yes to apply cleanups non-interactively, or run `coauthor-cleaner review`.")
				fmt.Print(scan.FormatText(result.Findings))
				return nil
			}
			cleanResult, err := clean.Apply(g, clean.Options{Findings: result.Findings})
			if err != nil {
				return err
			}
			fmt.Print(cleanResult.Summary)
			return nil
		},
	}
	addScanFlags(cmd)
	cmd.Flags().BoolVar(&flagYes, "yes", false, "apply cleanups without confirmation")
	return cmd
}

func reviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "review",
		Short: "Review and clean AI attribution markers interactively",
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
			return tui.Run(g, result.Findings)
		},
	}
}

func scanOptsFromGit(g git.Runner) scan.Options {
	cfg := loadConfig(g)
	head := flagHead
	if head == "" && flagBase != "" {
		head = "HEAD"
	}
	return scan.Options{
		Staged:     flagStaged,
		Commit:     flagCommit,
		Base:       flagBase,
		Head:       head,
		Strict:     flagStrict,
		Aggressive: flagAggressive,
		Config:     cfg,
	}
}

func printScanResult(findings []detect.Finding) error {
	if flagJSON {
		out, err := scan.FormatJSON(findings)
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	}
	fmt.Print(scan.FormatText(findings))
	return nil
}

func addScanFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&flagStaged, "staged", false, "scan staged changes")
	cmd.Flags().StringVar(&flagCommit, "commit", "", "scan commit message (e.g. HEAD)")
	cmd.Flags().StringVar(&flagBase, "base", "", "base ref for PR range scan")
	cmd.Flags().StringVar(&flagHead, "head", "", "head ref for PR range scan (default HEAD)")
	cmd.Flags().BoolVar(&flagJSON, "json", false, "output findings as JSON")
	cmd.Flags().BoolVar(&flagFailFindings, "fail-on-findings", false, "exit 1 when findings exist")
	cmd.Flags().BoolVar(&flagStrict, "strict", false, "only high-confidence patterns")
	cmd.Flags().BoolVar(&flagAggressive, "aggressive", false, "broader pattern matching")
}
