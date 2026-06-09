package scan

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/detect"
)

// FormatMarkdownComment renders findings as a GitHub PR comment.
func FormatMarkdownComment(findings []detect.Finding) string {
	if len(findings) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("## Coauthor Cleaner\n\n")
	b.WriteString(fmt.Sprintf("Found **%d** possible AI attribution marker", len(findings)))
	if len(findings) != 1 {
		b.WriteString("s")
	}
	b.WriteString(" in this PR.\n\n")

	bySource := groupBySource(findings)
	order := []detect.SourceType{
		detect.SourcePRDiff,
		detect.SourceCommitTrailer,
		detect.SourceCommitMessage,
		detect.SourcePRBody,
	}
	labels := map[detect.SourceType]string{
		detect.SourcePRDiff:        "PR diff",
		detect.SourceCommitTrailer: "Commit trailers",
		detect.SourceCommitMessage: "Commit messages",
		detect.SourcePRBody:        "PR body",
	}

	for _, src := range order {
		items, ok := bySource[src]
		if !ok {
			continue
		}
		b.WriteString("### ")
		b.WriteString(labels[src])
		b.WriteString("\n\n")
		for _, f := range items {
			loc := f.FilePath
			if f.LineNumber > 0 {
				loc += ":" + strconv.Itoa(f.LineNumber)
			}
			if loc != "" {
				b.WriteString("- `")
				b.WriteString(loc)
				b.WriteString("` — ")
			} else {
				b.WriteString("- ")
			}
			b.WriteString("`")
			b.WriteString(f.Match)
			b.WriteString("` (")
			b.WriteString(f.RuleName)
			b.WriteString(")\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("**Fix locally:**\n```bash\ncoauthor-cleaner review\n```\n")
	b.WriteString("\nOr scan before pushing:\n```bash\ncoauthor-cleaner scan\n```\n")
	return b.String()
}
