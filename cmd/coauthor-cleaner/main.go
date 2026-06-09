package main

import (
	"fmt"
	"os"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/clean"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/detect"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/git"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/scan"
	"github.com/spf13/cobra"
)

var (
	flagStaged       bool
	flagCommit       string
	flagYes          bool
	flagForce        bool
	flagStrict       bool
	flagAggressive   bool
	flagJSON         bool
	flagFailFindings bool
	flagBase         string
	flagHead         string
	flagFile         string
	flagDir          string
)

func main() {
	root := &cobra.Command{
		Use:   "coauthor-cleaner",
		Short: "Remove unwanted AI attribution from Git commits and staged changes",
		Long: `Coauthor Cleaner finds and removes AI attribution markers in git repos.

Default — interactive TUI (scan, review, clean, push):
  coauthor-cleaner

Non-interactive automation:
  coauthor-cleaner fix --push
  coauthor-cleaner fix --force --force-push`,
		RunE: runTUI,
	}

	root.AddCommand(fixCmd())
	root.AddCommand(statusCmd())
	root.AddCommand(doctorCmd())
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
		Long: `Scan git state (default), or arbitrary paths without a git repo:

  coauthor-cleaner scan --staged --commit HEAD
  coauthor-cleaner scan --file README.md
  coauthor-cleaner scan --dir ./src`,
		RunE: runScan,
	}
	addScanFlags(cmd)
	cmd.Flags().StringVar(&flagFile, "file", "", "scan a single file (no git repo required)")
	cmd.Flags().StringVar(&flagDir, "dir", "", "scan a directory recursively (no git repo required)")
	return cmd
}

func runScan(cmd *cobra.Command, args []string) error {
	if flagFile != "" || flagDir != "" {
		path := flagFile
		if flagDir != "" {
			path = flagDir
		}
		g := git.Runner{}
		cfg := loadConfig(g)
		findings, err := scan.ScanPath(path, cfg, flagStrict, flagAggressive)
		if err != nil {
			return err
		}
		if err := printScanResult(findings); err != nil {
			return err
		}
		if flagFailFindings && len(findings) > 0 {
			return fmt.Errorf("found %d AI attribution marker(s)", len(findings))
		}
		return nil
	}

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
	if len(result.Findings) > 0 {
		fmt.Println("Tip: run coauthor-cleaner to fix in the TUI, or coauthor-cleaner status for guidance.")
	}
	if flagFailFindings && len(result.Findings) > 0 {
		return fmt.Errorf("found %d AI attribution marker(s)", len(result.Findings))
	}
	return nil
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
			cleanResult, err := clean.Apply(g, clean.Options{Findings: result.Findings, ForceAmend: flagForce})
			if err != nil {
				return err
			}
			fmt.Print(cleanResult.Summary)
			return nil
		},
	}
	addScanFlags(cmd)
	cmd.Flags().BoolVar(&flagYes, "yes", false, "apply cleanups without confirmation")
	cmd.Flags().BoolVar(&flagForce, "force", false, "allow amending a commit already pushed to remote")
	return cmd
}

func reviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "review",
		Aliases: []string{"tui"},
		Short:   "Interactive TUI (same as running coauthor-cleaner with no args)",
		RunE:    runTUI,
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
