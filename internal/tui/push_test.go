package tui

import (
	"testing"

	"github.com/Skeyelab/coauthor-cleaner/internal/git"
	"github.com/Skeyelab/coauthor-cleaner/internal/guide"
)

func TestPushPlanAfterClean_ForceWhenAmendedPublished(t *testing.T) {
	st := git.RepoState{UpstreamExists: true, Upstream: "origin/main"}
	plan := pushPlanAfterClean(st, guide.CleanActions{AmendedCommit: true}, true)
	if !plan.Available || !plan.Force {
		t.Fatalf("plan = %+v", plan)
	}
}

func TestPushPlanAfterClean_StagedOnlyNeedsCommit(t *testing.T) {
	st := git.RepoState{UpstreamExists: true}
	plan := pushPlanAfterClean(st, guide.CleanActions{StagedFiles: true}, false)
	if plan.Available {
		t.Fatal("should not push staged-only without commit")
	}
}
