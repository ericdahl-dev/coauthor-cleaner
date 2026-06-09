package detect

import (
	"fmt"
	"strings"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/config"
)

type ScanOpts struct {
	AllowedTrailers []string
}

func ScanLines(text string, source SourceType, filePath string, rules []Rule, opts ...ScanOpts) []Finding {
	var o ScanOpts
	if len(opts) > 0 {
		o = opts[0]
	}
	lines := strings.Split(text, "\n")
	var findings []Finding
	id := 0

	for lineNum, line := range lines {
		trimmed := strings.TrimRight(line, "\r")
		for _, rule := range rules {
			if !ruleApplies(rule, source) {
				continue
			}

			effectiveSource := source
			if rule.Trailer && source == SourceCommitMessage {
				effectiveSource = SourceCommitTrailer
			}

			if !rule.Pattern.MatchString(trimmed) {
				continue
			}
			if rule.Trailer && config.IsAllowedTrailer(trimmed, o.AllowedTrailers) {
				continue
			}

			id++
			findings = append(findings, Finding{
				ID:          fmt.Sprintf("%s:%d:%d", filePath, lineNum+1, id),
				Source:      effectiveSource,
				FilePath:    filePath,
				LineNumber:  lineNum + 1,
				Match:       trimmed,
				Replacement: rule.Replacement,
				Confidence:  rule.Confidence,
				RuleName:    rule.Name,
				Selected:    true,
			})
			break
		}
	}

	return findings
}

func CleanText(text string, findings []Finding) string {
	if len(findings) == 0 {
		return text
	}

	remove := make(map[int]bool)
	for _, f := range findings {
		if f.Selected {
			remove[f.LineNumber] = true
		}
	}

	lines := strings.Split(text, "\n")
	var out []string
	for i, line := range lines {
		if !remove[i+1] {
			out = append(out, line)
		}
	}

	result := strings.Join(out, "\n")
	// preserve trailing newline if input had one
	if strings.HasSuffix(text, "\n") && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
}
