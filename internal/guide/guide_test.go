package guide

import (
	"strings"
	"testing"

	"github.com/Skeyelab/coauthor-cleaner/internal/detect"
	"github.com/Skeyelab/coauthor-cleaner/internal/git"
)

func TestNextSteps_AmendedPushedCommit(t *testing.T) {
	st := git.RepoState{
		Branch:          "main",
		Upstream:        "origin/main",
		UpstreamExists:  true,
		HEADOnRemoteTip: true,
	}
	out := NextSteps(st, CleanActions{AmendedCommit: true}, 0)
	if !strings.Contains(out, "force-with-lease") {
		t.Errorf("expected force push guidance: %q", out)
	}
}

func TestStatusReport_WithFindings(t *testing.T) {
	out := StatusReport(git.RepoState{Branch: "feat"}, []detect.Finding{
		{Source: detect.SourceStagedDiff, FilePath: "a.go", Match: "Generated with Claude"},
	})
	if !strings.Contains(out, "review") {
		t.Error("expected review recommendation")
	}
}
