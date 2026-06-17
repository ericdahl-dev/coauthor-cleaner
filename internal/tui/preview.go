package tui

import (
	"fmt"
	"strings"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/detect"
	"github.com/charmbracelet/lipgloss"
)

func PreviewDiff(f detect.Finding) string {
	var b strings.Builder
	b.WriteString(subStyle.Render("Before / after") + "\n\n")

	var diff strings.Builder
	diff.WriteString(lipgloss.NewStyle().Foreground(colorError).Render(fmt.Sprintf("- %s", f.Match)) + "\n")
	if f.Replacement != "" {
		diff.WriteString(lipgloss.NewStyle().Foreground(colorOK).Render(fmt.Sprintf("+ %s", f.Replacement)) + "\n")
	} else {
		diff.WriteString(lipgloss.NewStyle().Foreground(colorOK).Render("+ <removed>") + "\n")
	}
	b.WriteString(codeBoxStyle.Render(diff.String()))
	b.WriteString("\n\n")
	b.WriteString(subStyle.Render("Rule") + " " + titleStyle.Render(f.RuleName) + "\n")
	b.WriteString(subStyle.Render("Confidence") + " " + okStyle.Render(string(f.Confidence)) + "\n")
	return b.String()
}
