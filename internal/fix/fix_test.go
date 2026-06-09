package fix

import (
	"testing"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/detect"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/git"
)

func TestPartition_SkipsPushedCommit(t *testing.T) {
	st := git.RepoState{UpstreamExists: true, HEADOnRemoteTip: true}
	findings := []detect.Finding{
		{Source: detect.SourceStagedDiff, FilePath: "a.go"},
		{Source: detect.SourceCommitTrailer},
	}
	toApply, skipped := partition(findings, st, false)
	if len(toApply) != 1 {
		t.Fatalf("expected 1 safe finding, got %d", len(toApply))
	}
	if skipped != 1 {
		t.Fatalf("skipped = %d", skipped)
	}
}

func TestPartition_ForceIncludesCommit(t *testing.T) {
	st := git.RepoState{UpstreamExists: true, HEADOnRemoteTip: true}
	findings := []detect.Finding{
		{Source: detect.SourceCommitTrailer},
	}
	toApply, skipped := partition(findings, st, true)
	if len(toApply) != 1 || skipped != 0 {
		t.Fatalf("force should include commit: apply=%d skip=%d", len(toApply), skipped)
	}
}
