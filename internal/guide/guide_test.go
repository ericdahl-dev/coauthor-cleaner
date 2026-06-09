package guide

import (
	"strings"
	"testing"

	"github.com/Skeyelab/coauthor-cleaner/internal/detect"
	"github.com/Skeyelab/coauthor-cleaner/internal/git"
)

func TestNextSteps_AmendedPushedCommit(t *testing.T) {
	st := git.RepoState{UpstreamExists: true, HEADOnRemoteTip: true, Upstream: "origin/main"}
	out := NextSteps(st, CleanActions{AmendedCommit: true}, 0)
	if !strings.Contains(out, "force-with-lease") {
		t.Fatal(out)
	}
}

func TestStatusReport_WithFindings(t *testing.T) {
	out := StatusReport(git.RepoState{Branch: "feat"}, []detect.Finding{
		{Source: detect.SourceStagedDiff, FilePath: "a.go", Match: "Generated with Claude"},
	})
	if !strings.Contains(out, "coauthor-cleaner") {
		t.Fatal(out)
	}
}

func TestDoctorReport(t *testing.T) {
	out := DoctorReport(true, true, false, true, ".coauthor-cleaner.yml")
	if !strings.Contains(out, "doctor") || !strings.Contains(out, "hooks") {
		t.Fatal(out)
	}
}

func TestFormatFindingsBrief_Truncates(t *testing.T) {
	long := strings.Repeat("a", 60)
	out := FormatFindingsBrief([]detect.Finding{{FilePath: "x", Match: long}})
	if !strings.Contains(out, "...") {
		t.Fatal(out)
	}
}
