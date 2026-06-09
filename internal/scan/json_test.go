package scan

import (
	"encoding/json"
	"testing"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/detect"
)

func TestFormatJSON(t *testing.T) {
	findings := []detect.Finding{
		{
			ID:         "app.go:1:1",
			Source:     detect.SourcePRDiff,
			FilePath:   "app.go",
			LineNumber: 1,
			Match:      "// Generated with Claude Code",
			RuleName:   "claude-generated-with",
			Confidence: detect.ConfidenceHigh,
		},
	}

	out, err := FormatJSON(findings)
	if err != nil {
		t.Fatal(err)
	}

	var report JSONReport
	if err := json.Unmarshal(out, &report); err != nil {
		t.Fatal(err)
	}
	if report.Count != 1 {
		t.Errorf("count = %d", report.Count)
	}
	if report.Findings[0].Source != "pr_diff" {
		t.Errorf("source = %q", report.Findings[0].Source)
	}
}
