package fix

import (
	"fmt"
	"os"
	"strings"

	"github.com/Skeyelab/coauthor-cleaner/internal/clean"
	"github.com/Skeyelab/coauthor-cleaner/internal/detect"
	"github.com/Skeyelab/coauthor-cleaner/internal/git"
	"github.com/Skeyelab/coauthor-cleaner/internal/guide"
	"github.com/Skeyelab/coauthor-cleaner/internal/scan"
)

type Options struct {
	CheckOnly  bool
	ForceAmend bool
	Push       bool
	ForcePush  bool
}

type Result struct {
	Findings      []detect.Finding
	Applied       int
	SkippedCommit int
	Pushed        bool
	ForcePushed   bool
	Summary       string
}

func Run(g git.Runner, scanOpts scan.Options, opts Options) (Result, error) {
	scanOpts.Staged = true
	scanOpts.Commit = "HEAD"

	result, err := scan.Run(g, scanOpts)
	if err != nil {
		return Result{}, err
	}
	if len(result.Findings) == 0 {
		st, _ := g.RepoState()
		return Result{
			Summary: "✓ No AI attribution markers found.\n\n" + guide.StatusReport(st, nil),
		}, nil
	}

	st, err := g.RepoState()
	if err != nil {
		return Result{}, err
	}

	toApply, skipped := partition(result.Findings, st, opts.ForceAmend)

	if opts.CheckOnly {
		return Result{
			Findings:      result.Findings,
			SkippedCommit: skipped,
			Summary:       planSummary(st, toApply, skipped, opts),
		}, nil
	}

	if len(toApply) == 0 {
		var b strings.Builder
		b.WriteString(fmt.Sprintf("Found %d marker(s) but none are safe to auto-fix.\n\n", len(result.Findings)))
		b.WriteString("  • Staged files: run coauthor-cleaner fix (always safe)\n")
		if skipped > 0 {
			b.WriteString("  • HEAD commit is on the remote — re-run with:\n\n")
			b.WriteString("      coauthor-cleaner fix --force --push\n\n")
		}
		b.WriteString(guide.StatusReport(st, result.Findings))
		return Result{Findings: result.Findings, SkippedCommit: skipped, Summary: b.String()}, nil
	}

	cleanResult, err := clean.Apply(g, clean.Options{Findings: toApply, ForceAmend: opts.ForceAmend})
	if err != nil {
		return Result{}, err
	}

	out := Result{
		Findings:      result.Findings,
		Applied:       cleanResult.Applied,
		SkippedCommit: skipped,
		Summary:       cleanResult.Summary,
	}

	st, _ = g.RepoState()
	if opts.Push && out.Applied > 0 {
		needForce := cleanResult.Actions.AmendedCommit && (st.AmendingRewritesPushedCommit() || opts.ForceAmend)
		if needForce {
			if !opts.ForcePush {
				out.Summary += "\n⚠ Skipped push — amended a published commit. Re-run with:\n\n    coauthor-cleaner fix --force --force-push\n"
				return out, nil
			}
			if err := g.PushForceWithLease(); err != nil {
				return out, err
			}
			out.ForcePushed = true
			out.Pushed = true
			out.Summary += "\n✓ Pushed with --force-with-lease\n"
		} else {
			if err := g.Push(); err != nil {
				return out, err
			}
			out.Pushed = true
			out.Summary += "\n✓ Pushed to " + st.Upstream + "\n"
		}
	}

	// Re-scan and show final status
	recheck, _ := scan.Run(g, scanOpts)
	st, _ = g.RepoState()
	out.Summary += "\n" + guide.StatusReport(st, recheck.Findings)
	return out, nil
}

func partition(findings []detect.Finding, st git.RepoState, forceAmend bool) (toApply []detect.Finding, skippedCommit int) {
	needsForce := st.AmendingRewritesPushedCommit() && !forceAmend
	for _, f := range findings {
		f.Selected = true
		if isCommitFinding(f) {
			if needsForce {
				skippedCommit++
				continue
			}
		}
		toApply = append(toApply, f)
	}
	return toApply, skippedCommit
}

func isCommitFinding(f detect.Finding) bool {
	return f.Source == detect.SourceCommitMessage || f.Source == detect.SourceCommitTrailer
}

func planSummary(st git.RepoState, toApply []detect.Finding, skipped int, opts Options) string {
	var b strings.Builder
	b.WriteString("Dry run — coauthor-cleaner fix would:\n\n")
	if len(toApply) == 0 {
		b.WriteString("  (nothing safe to apply automatically)\n")
	} else {
		for _, f := range toApply {
			b.WriteString(fmt.Sprintf("  • remove %s (%s)\n", f.Match, f.Source))
		}
	}
	if skipped > 0 {
		b.WriteString(fmt.Sprintf("\n  • skip %d HEAD commit marker(s) — already on remote (needs --force)\n", skipped))
	}
	if opts.Push {
		if opts.ForcePush {
			b.WriteString("\n  • push with --force-with-lease\n")
		} else {
			b.WriteString("\n  • git push\n")
		}
	}
	return b.String()
}

// CleanCommitMsgFile removes selected lines from a commit message file (pre-commit).
func CleanCommitMsgFile(path string, findings []detect.Finding) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	cleaned := detect.CleanText(string(data), findings)
	if cleaned == string(data) {
		return nil
	}
	return os.WriteFile(path, []byte(cleaned), 0644)
}
