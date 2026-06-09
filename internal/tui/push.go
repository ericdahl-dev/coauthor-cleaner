package tui

import (
	"github.com/ericdahl-dev/coauthor-cleaner/internal/git"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/guide"
)

type pushPlan struct {
	Available bool
	Force     bool
	Reason    string
}

func pushPlanAfterClean(st git.RepoState, actions guide.CleanActions, amendedPublished bool) pushPlan {
	if actions.StagedFiles && !actions.AmendedCommit {
		return pushPlan{
			Reason: "Staged files cleaned — commit first, then push from your terminal.",
		}
	}
	if !st.UpstreamExists {
		return pushPlan{
			Reason: "No upstream branch. Run: git push -u origin " + st.Branch,
		}
	}
	if actions.AmendedCommit && amendedPublished {
		return pushPlan{Available: true, Force: true, Reason: "Amended a commit that was on the remote."}
	}
	if actions.AmendedCommit || st.Ahead > 0 {
		return pushPlan{Available: true, Reason: "Branch has unpushed commits."}
	}
	return pushPlan{Reason: "Nothing to push."}
}
