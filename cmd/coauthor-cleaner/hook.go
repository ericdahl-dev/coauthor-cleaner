package main

import (
	"fmt"
	"os"

	"github.com/Skeyelab/coauthor-cleaner/internal/clean"
	"github.com/Skeyelab/coauthor-cleaner/internal/config"
	"github.com/Skeyelab/coauthor-cleaner/internal/fix"
	"github.com/Skeyelab/coauthor-cleaner/internal/detect"
	"github.com/Skeyelab/coauthor-cleaner/internal/git"
	"github.com/Skeyelab/coauthor-cleaner/internal/scan"
	"github.com/spf13/cobra"
)

func hookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Manage git hooks",
	}
	cmd.AddCommand(hookInstallCmd())
	cmd.AddCommand(hookRunCmd())
	return cmd
}

func hookInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install pre-commit and commit-msg hooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := gitRunner()
			if err != nil {
				return err
			}
			if err := g.InstallHooks(executable()); err != nil {
				return err
			}
			fmt.Println("Installed coauthor-cleaner git hooks (pre-commit, commit-msg)")
			return nil
		},
	}
}

func hookRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [hook-type] [args...]",
		Short: "Run a hook handler (called by git hooks)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := gitRunner()
			if err != nil {
				return err
			}
			cfg := loadConfig(g)
			mode := cfg.Behavior.HookMode
			if mode == "" {
				mode = "warn"
			}

			switch args[0] {
			case "pre-commit":
				return runPreCommitHook(g, cfg, mode)
			case "commit-msg":
				msgFile := ""
				if len(args) > 1 {
					msgFile = args[1]
				}
				return runCommitMsgHook(cfg, mode, msgFile)
			default:
				return fmt.Errorf("unknown hook: %s", args[0])
			}
		},
	}
}

func runPreCommitHook(g git.Runner, cfg config.Config, mode string) error {
	result, err := scan.Run(g, scan.Options{Staged: true, Config: cfg})
	if err != nil {
		return err
	}
	return handleHookFindings(g, mode, "", result.Findings, true)
}

func runCommitMsgHook(cfg config.Config, mode, msgFile string) error {
	if msgFile == "" {
		return nil
	}
	data, err := os.ReadFile(msgFile)
	if err != nil {
		return err
	}
	rules := detect.SelectRules(cfg, false, false)
	opts := detect.ScanOpts{AllowedTrailers: cfg.AllowedTrailers}
	findings := detect.ScanLines(string(data), detect.SourceCommitMessage, "", rules, opts)
	return handleHookFindings(git.Runner{}, mode, msgFile, findings, false)
}

func handleHookFindings(g git.Runner, mode, msgFile string, findings []detect.Finding, stagedOnly bool) error {
	if len(findings) == 0 {
		return nil
	}

	switch mode {
	case "clean":
		if msgFile != "" {
			for i := range findings {
				findings[i].Selected = true
			}
			if err := fix.CleanCommitMsgFile(msgFile, findings); err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, "coauthor-cleaner: auto-cleaned commit message")
			return nil
		}
		if stagedOnly && g.InRepo() {
			var staged []detect.Finding
			for _, f := range findings {
				if f.Source == detect.SourceStagedDiff {
					f.Selected = true
					staged = append(staged, f)
				}
			}
			if len(staged) > 0 {
				_, err := clean.Apply(g, clean.Options{Findings: staged})
				if err == nil {
					fmt.Fprintln(os.Stderr, "coauthor-cleaner: auto-cleaned staged files")
				}
				return err
			}
		}
		return nil
	}

	fmt.Fprint(os.Stderr, scan.FormatText(findings))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Run: coauthor-cleaner")
	fmt.Fprintln(os.Stderr, "Or bypass with: git commit --no-verify")

	if mode == "block" {
		return fmt.Errorf("coauthor-cleaner: %d attribution marker(s) found", len(findings))
	}
	return nil
}
