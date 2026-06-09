package scan

import (
	"strings"
	"testing"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/detect"
)

func TestFormatText(t *testing.T) {
	out := FormatText([]detect.Finding{{
		Source: detect.SourceStagedDiff, FilePath: "a.go", LineNumber: 1,
		Match: "// Generated with Claude Code",
	}})
	if !strings.Contains(out, "1 attribution marker") || !strings.Contains(out, "a.go") {
		t.Fatal(out)
	}
}

func TestFormatText_Empty(t *testing.T) {
	if FormatText(nil) != "No AI attribution markers found.\n" {
		t.Fatal()
	}
}
