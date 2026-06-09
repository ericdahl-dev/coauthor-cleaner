package scan

import (
	"strconv"
	"strings"

	"github.com/Skeyelab/coauthor-cleaner/internal/config"
	"github.com/Skeyelab/coauthor-cleaner/internal/detect"
	"github.com/Skeyelab/coauthor-cleaner/internal/git"
)

type Options struct {
	Staged     bool
	Commit     string
	Base       string
	Head       string
	PRBody     string
	Strict     bool
	Aggressive bool
	Config     config.Config
}

type Result struct {
	Findings []detect.Finding
}

func Run(g git.Runner, opts Options) (Result, error) {
	cfg := opts.Config
	if cfg.Targets == (config.Targets{}) {
		cfg = config.Default()
	}
	rules := detect.SelectRules(cfg, opts.Strict, opts.Aggressive)
	scanOpts := detect.ScanOpts{AllowedTrailers: cfg.AllowedTrailers}

	var findings []detect.Finding

	if opts.Staged && cfg.Targets.StagedDiff {
		files, err := g.StagedFileContents()
		if err != nil {
			return Result{}, err
		}
		for path, content := range files {
			found := detect.ScanLines(content, detect.SourceStagedDiff, path, rules, scanOpts)
			findings = append(findings, found...)
		}
	}

	if opts.Commit != "" && cfg.Targets.CommitMessages {
		msg, err := g.CommitMessage(opts.Commit)
		if err != nil {
			return Result{}, err
		}
		ref := opts.Commit
		if ref == "" {
			ref = "HEAD"
		}
		found := detect.ScanLines(msg, detect.SourceCommitMessage, ref, rules, scanOpts)
		findings = append(findings, found...)
	}

	if opts.PRBody != "" && cfg.Targets.PRBody {
		found := detect.ScanLines(opts.PRBody, detect.SourcePRBody, "pull_request", rules, scanOpts)
		findings = append(findings, found...)
	}

	if opts.Base != "" && opts.Head != "" {
		if cfg.Targets.StagedDiff {
			added, err := g.DiffAddedLines(opts.Base, opts.Head)
			if err != nil {
				return Result{}, err
			}
			for _, line := range added {
				text := line.Line + "\n"
				found := detect.ScanLines(text, detect.SourcePRDiff, line.FilePath, rules, scanOpts)
				findings = append(findings, found...)
			}
		}

		if cfg.Targets.CommitMessages {
			commits, err := g.CommitMessagesInRange(opts.Base, opts.Head)
			if err != nil {
				return Result{}, err
			}
			for sha, msg := range commits {
				found := detect.ScanLines(msg, detect.SourceCommitMessage, sha[:7], rules, scanOpts)
				findings = append(findings, found...)
			}
		}
	}

	return Result{Findings: findings}, nil
}

func FormatText(findings []detect.Finding) string {
	if len(findings) == 0 {
		return "No AI attribution markers found.\n"
	}

	var b strings.Builder
	b.WriteString("Coauthor Cleaner found ")
	if len(findings) == 1 {
		b.WriteString("1 attribution marker\n\n")
	} else {
		b.WriteString(strconv.Itoa(len(findings)) + " attribution markers\n\n")
	}

	bySource := groupBySource(findings)
	order := []detect.SourceType{
		detect.SourcePRDiff,
		detect.SourceStagedDiff,
		detect.SourceCommitTrailer,
		detect.SourceCommitMessage,
		detect.SourcePRBody,
		detect.SourceFileHeader,
	}
	labels := map[detect.SourceType]string{
		detect.SourcePRDiff:        "PR diff",
		detect.SourceStagedDiff:    "staged diff",
		detect.SourceCommitTrailer: "commit trailer",
		detect.SourceCommitMessage: "commit message",
		detect.SourcePRBody:        "PR body",
		detect.SourceFileHeader:    "generated header",
	}

	for _, src := range order {
		items, ok := bySource[src]
		if !ok || len(items) == 0 {
			continue
		}
		b.WriteString("  ")
		b.WriteString(labels[src])
		b.WriteString("\n")
		for _, f := range items {
			loc := f.FilePath
			if f.LineNumber > 0 {
				loc += ":" + strconv.Itoa(f.LineNumber)
			}
			if loc != "" {
				b.WriteString("    ")
				b.WriteString(loc)
				b.WriteString("\n")
			}
			b.WriteString("    ")
			b.WriteString(f.Match)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

func groupBySource(findings []detect.Finding) map[detect.SourceType][]detect.Finding {
	m := make(map[detect.SourceType][]detect.Finding)
	for _, f := range findings {
		m[f.Source] = append(m[f.Source], f)
	}
	return m
}
