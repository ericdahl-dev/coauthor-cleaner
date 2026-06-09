package clean

import (
	"fmt"
	"strings"

	"github.com/Skeyelab/coauthor-cleaner/internal/detect"
	"github.com/Skeyelab/coauthor-cleaner/internal/git"
	"github.com/Skeyelab/coauthor-cleaner/internal/guide"
)

type Options struct {
	Findings   []detect.Finding
	ForceAmend bool
}

type Result struct {
	Applied int
	Actions guide.CleanActions
	Summary string
}

func Apply(g git.Runner, opts Options) (Result, error) {
	selected := filterSelected(opts.Findings)
	if len(selected) == 0 {
		return Result{Summary: "No changes selected.\n"}, nil
	}

	applied := 0
	actions := guide.CleanActions{}
	byFile := groupByFile(selected)
	byCommit := groupByCommit(selected)

	if len(byCommit) > 0 {
		st, err := g.RepoState()
		if err != nil {
			return Result{}, err
		}
		if st.AmendingRewritesPushedCommit() && !opts.ForceAmend {
			return Result{}, fmt.Errorf("HEAD commit is already on %s — amending rewrites published history\n\nUse --force if you intend to run: git push --force-with-lease\nOr run: coauthor-cleaner status", st.Upstream)
		}
	}

	for path, fileFindings := range byFile {
		files, err := g.StagedFileContents()
		if err != nil {
			return Result{}, err
		}
		content, ok := files[path]
		if !ok {
			continue
		}
		cleaned := detect.CleanText(content, fileFindings)
		if cleaned == content {
			continue
		}
		if err := g.StageFile(path, cleaned); err != nil {
			return Result{}, fmt.Errorf("stage %s: %w", path, err)
		}
		applied += countSelected(fileFindings)
		actions.StagedFiles = true
	}

	for ref, commitFindings := range byCommit {
		msg, err := g.CommitMessage(ref)
		if err != nil {
			return Result{}, err
		}
		cleaned := detect.CleanText(msg, commitFindings)
		if cleaned == msg {
			continue
		}
		if err := g.AmendCommitMessage(strings.TrimSpace(cleaned)); err != nil {
			return Result{}, fmt.Errorf("amend commit: %w", err)
		}
		applied += countSelected(commitFindings)
		actions.AmendedCommit = true
	}

	st, _ := g.RepoState()
	summary := formatSummary(applied, selected)
	summary += "\n" + guide.NextSteps(st, actions, 0)
	return Result{Applied: applied, Actions: actions, Summary: summary}, nil
}

func filterSelected(findings []detect.Finding) []detect.Finding {
	var out []detect.Finding
	for _, f := range findings {
		if f.Selected {
			out = append(out, f)
		}
	}
	return out
}

func groupByFile(findings []detect.Finding) map[string][]detect.Finding {
	m := make(map[string][]detect.Finding)
	for _, f := range findings {
		if f.Source != detect.SourceStagedDiff && f.Source != detect.SourceFileHeader {
			continue
		}
		if f.FilePath == "" {
			continue
		}
		m[f.FilePath] = append(m[f.FilePath], f)
	}
	return m
}

func groupByCommit(findings []detect.Finding) map[string][]detect.Finding {
	m := make(map[string][]detect.Finding)
	for _, f := range findings {
		if f.Source != detect.SourceCommitTrailer && f.Source != detect.SourceCommitMessage {
			continue
		}
		ref := f.FilePath
		if ref == "" {
			ref = "HEAD"
		}
		m[ref] = append(m[ref], f)
	}
	return m
}

func countSelected(findings []detect.Finding) int {
	n := 0
	for _, f := range findings {
		if f.Selected {
			n++
		}
	}
	return n
}

func formatSummary(applied int, findings []detect.Finding) string {
	var b strings.Builder
	if applied == 1 {
		b.WriteString("Cleaned 1 attribution marker\n\n")
	} else {
		b.WriteString(fmt.Sprintf("Cleaned %d attribution markers\n\n", applied))
	}

	rules := make(map[string]int)
	for _, f := range findings {
		if f.Selected {
			rules[f.RuleName]++
		}
	}
	for name, count := range rules {
		b.WriteString(fmt.Sprintf("  removed %d %s\n", count, name))
	}
	b.WriteString("\nNo human coauthors were changed.\n")
	return b.String()
}
