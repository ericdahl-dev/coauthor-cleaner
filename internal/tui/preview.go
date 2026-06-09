package tui

import (
	"fmt"
	"strings"

	"github.com/Skeyelab/coauthor-cleaner/internal/detect"
)

func PreviewDiff(f detect.Finding) string {
	var b strings.Builder
	b.WriteString("Before / after\n\n")
	b.WriteString(fmt.Sprintf("  - %s\n", f.Match))
	if f.Replacement != "" {
		b.WriteString(fmt.Sprintf("  + %s\n", f.Replacement))
	} else {
		b.WriteString("  +\n")
	}
	b.WriteString(fmt.Sprintf("\nRule: %s (%s confidence)\n", f.RuleName, f.Confidence))
	return b.String()
}
