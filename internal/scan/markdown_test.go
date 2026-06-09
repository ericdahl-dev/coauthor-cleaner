package scan

import (
	"strings"
	"testing"

	"github.com/Skeyelab/coauthor-cleaner/internal/detect"
)

func TestFormatMarkdownComment(t *testing.T) {
	comment := FormatMarkdownComment([]detect.Finding{
		{
			Source:   detect.SourcePRDiff,
			FilePath: "app.go",
			Match:    "// Generated with Claude Code",
			RuleName: "claude-generated-with",
		},
	})

	if !strings.Contains(comment, "## Coauthor Cleaner") {
		t.Error("missing header")
	}
	if !strings.Contains(comment, "app.go") {
		t.Error("missing file path")
	}
	if !strings.Contains(comment, "coauthor-cleaner review") {
		t.Error("missing fix instructions")
	}
}

func TestFormatMarkdownComment_Empty(t *testing.T) {
	if FormatMarkdownComment(nil) != "" {
		t.Error("expected empty comment")
	}
}
