package clean

import (
	"fmt"
	"strings"

	"github.com/Skeyelab/coauthor-cleaner/internal/detect"
	"github.com/Skeyelab/coauthor-cleaner/internal/git"
)

type Options struct {
	Findings []detect.Finding
}

type Result struct {
	Applied int
	Summary string
}

func Apply(g git.Runner, opts Options) (Result, error) {
	selected := filterSelected(opts.Findings)
	if len(selected) == 0 {
		return Result{Summary: "No changes selected.\n"}, nil
	}

	applied := 0
	byFile := groupByFile(selected)
	byCommit := groupByCommit(selected)

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
	}

	summary := formatSummary(applied, selected)
	return Result{Applied: applied, Summary: summary}, nil
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
