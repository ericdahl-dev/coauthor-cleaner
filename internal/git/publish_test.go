package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRepoState_UnpushedCommit(t *testing.T) {
	dir, r := setupTestRepo(t)
	readme := filepath.Join(dir, "README.md")
	os.WriteFile(readme, []byte("hi\n"), 0644)
	exec.Command("git", "-C", dir, "add", "README.md").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "init").Run()

	st, err := r.RepoState()
	if err != nil {
		t.Fatal(err)
	}
	if st.Branch == "" {
		t.Error("expected branch")
	}
	if st.AmendingRewritesPushedCommit() {
		t.Error("no upstream should not imply pushed commit")
	}
}
