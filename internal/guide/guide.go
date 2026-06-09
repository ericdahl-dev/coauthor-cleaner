package guide

import (
	"fmt"
	"strings"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/detect"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/git"
)

type CleanActions struct {
	AmendedCommit bool
	StagedFiles   bool
}

func NextSteps(st git.RepoState, actions CleanActions, findingCount int) string {
	if findingCount > 0 && !actions.AmendedCommit && !actions.StagedFiles {
		return formatFindingSteps(findingCount)
	}

	var b strings.Builder
	b.WriteString("What to do next\n\n")

	if actions.AmendedCommit {
		b.WriteString("  • HEAD commit message was rewritten (new commit hash).\n")
		if st.AmendingRewritesPushedCommit() {
			b.WriteString("\n  ⚠ This commit was already on the remote. Update GitHub with:\n\n")
			b.WriteString("      git push --force-with-lease\n\n")
			b.WriteString("  Only do this if you are sure no one else based work on the old commit.\n")
		} else if st.UpstreamExists && st.Ahead > 0 {
			b.WriteString("\n  • Push your branch:\n\n")
			b.WriteString("      git push\n\n")
		} else if !st.UpstreamExists {
			b.WriteString("\n  • No upstream set. When ready, push with:\n\n")
			b.WriteString("      git push -u origin " + st.Branch + "\n\n")
		} else {
			b.WriteString("\n  • Verify with: git log -1 --format=fuller\n")
			b.WriteString("  • Push when ready: git push\n\n")
		}
	}

	if actions.StagedFiles {
		b.WriteString("  • Staged files were updated. Commit when ready:\n\n")
		b.WriteString("      git commit -m \"your message\"\n\n")
	}

	if !actions.AmendedCommit && !actions.StagedFiles {
		b.WriteString("  • Verify: coauthor-cleaner scan\n")
	}

	b.WriteString("  • Check overall status: coauthor-cleaner status\n")
	return b.String()
}

func formatFindingSteps(count int) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Found %d attribution marker(s). Recommended:\n\n", count))
	b.WriteString("  1. coauthor-cleaner              # interactive TUI (recommended)\n")
	b.WriteString("  2. coauthor-cleaner fix --push   # non-interactive\n\n")
	b.WriteString("Run coauthor-cleaner status for a full repo check.\n")
	return b.String()
}

func StatusReport(st git.RepoState, findings []detect.Finding) string {
	var b strings.Builder
	b.WriteString("Coauthor Cleaner status\n\n")

	b.WriteString(fmt.Sprintf("Branch: %s\n", st.Branch))
	if st.UpstreamExists {
		b.WriteString(fmt.Sprintf("Upstream: %s (ahead %d, behind %d)\n", st.Upstream, st.Ahead, st.Behind))
	} else {
		b.WriteString("Upstream: not set — run: git push -u origin " + st.Branch + "\n")
	}

	if st.HasStagedChanges {
		b.WriteString("Staged changes: yes\n")
	} else {
		b.WriteString("Staged changes: none\n")
	}

	if len(findings) == 0 {
		b.WriteString("\n✓ No AI attribution markers in staged changes or HEAD commit.\n")
		if st.AmendingRewritesPushedCommit() {
			b.WriteString("\nNote: HEAD matches remote. Amending this commit will require git push --force-with-lease.\n")
		}
		return b.String()
	}

	b.WriteString(fmt.Sprintf("\n✗ Found %d attribution marker(s).\n\n", len(findings)))
	b.WriteString(FormatFindingsBrief(findings))
	b.WriteString("\n")
	b.WriteString(formatFindingSteps(len(findings)))
	return b.String()
}

func FormatFindingsBrief(findings []detect.Finding) string {
	var b strings.Builder
	limit := 5
	for i, f := range findings {
		if i >= limit {
			b.WriteString(fmt.Sprintf("  ... and %d more\n", len(findings)-limit))
			break
		}
		loc := string(f.Source)
		if f.FilePath != "" {
			loc = f.FilePath
		}
		b.WriteString(fmt.Sprintf("  • %s — %s\n", loc, truncate(f.Match, 50)))
	}
	return b.String()
}

func DoctorReport(inRepo, ghOK, hooksOK, configOK bool, configPath string) string {
	var b strings.Builder
	b.WriteString("Coauthor Cleaner doctor\n\n")

	check := func(ok bool, label, fix string) {
		if ok {
			b.WriteString("  ✓ " + label + "\n")
		} else {
			b.WriteString("  ✗ " + label + "\n")
			if fix != "" {
				b.WriteString("      → " + fix + "\n")
			}
		}
	}

	check(inRepo, "Inside a git repository", "cd into your project repo")
	check(configOK, "Config file ("+configPath+")", "coauthor-cleaner config init")
	check(hooksOK, "Git hooks installed", "coauthor-cleaner hook install")
	check(ghOK, "GitHub CLI (gh) for pr commands", "brew install gh && gh auth login")

	b.WriteString("\nRecommended workflow:\n")
	b.WriteString("  coauthor-cleaner           → interactive TUI (default)\n")
	b.WriteString("  hook_mode: clean           → auto-fix on git commit\n")
	return b.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
