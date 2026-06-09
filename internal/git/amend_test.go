package git

import (
	"strings"
	"testing"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/testutil"
)

func TestStageFile_AndAmend(t *testing.T) {
	dir := testutil.GitRepo(t)
	g := Runner{Dir: dir}
	testutil.Write(t, dir, "a.go", "old\n")
	testutil.Git(t, dir, "add", "a.go")
	testutil.Git(t, dir, "commit", "-m", "init\n\nCo-authored-by: Claude <noreply@anthropic.com>")

	if err := g.StageFile("a.go", "new\n"); err != nil {
		t.Fatal(err)
	}
	staged, err := g.StagedFileContents()
	if err != nil || staged["a.go"] != "new" {
		t.Fatalf("staged=%v err=%v", staged, err)
	}
	if err := g.AmendCommitMessage("init"); err != nil {
		t.Fatal(err)
	}
	msg, err := g.CommitMessage("HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(msg, "Claude") {
		t.Fatal(msg)
	}
}
