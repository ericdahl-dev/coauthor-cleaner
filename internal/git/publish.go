package git

import (
	"strconv"
	"strings"
)

// RepoState describes branch sync and working tree status.
type RepoState struct {
	Branch           string
	Upstream         string
	UpstreamExists   bool
	Ahead            int
	Behind           int
	HEADOnRemoteTip  bool // amending will rewrite a commit already on the remote
	HasStagedChanges bool
	HasUnstaged      bool
}

func (r Runner) RepoState() (RepoState, error) {
	st := RepoState{}

	branch, err := r.run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return st, err
	}
	st.Branch = branch

	upstream, err := r.run("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	if err == nil && upstream != "" {
		st.Upstream = upstream
		st.UpstreamExists = true

		ahead, _ := r.run("rev-list", "--count", upstream+"..HEAD")
		behind, _ := r.run("rev-list", "--count", "HEAD.."+upstream)
		st.Ahead, _ = strconv.Atoi(ahead)
		st.Behind, _ = strconv.Atoi(behind)
		st.HEADOnRemoteTip = st.Ahead == 0 && st.Behind == 0
	}

	staged, _ := r.run("diff", "--cached", "--name-only")
	st.HasStagedChanges = strings.TrimSpace(staged) != ""

	unstaged, _ := r.run("diff", "--name-only")
	st.HasUnstaged = strings.TrimSpace(unstaged) != ""

	return st, nil
}

// AmendingRewritesPushedCommit reports whether amending HEAD would change a commit
// already on the remote tracking branch.
func (st RepoState) AmendingRewritesPushedCommit() bool {
	return st.UpstreamExists && st.HEADOnRemoteTip
}
