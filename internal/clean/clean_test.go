package clean

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Skeyelab/coauthor-cleaner/internal/detect"
	"github.com/Skeyelab/coauthor-cleaner/internal/git"
)

func TestApply_BlocksAmendOfPushedCommit(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "t@example.com")
	run("config", "user.name", "T")
	readme := filepath.Join(dir, "README.md")
	os.WriteFile(readme, []byte("hi\n"), 0644)
	run("add", "README.md")
	run("commit", "-m", "init\n\nCo-authored-by: Claude <noreply@anthropic.com>")

	g := git.Runner{Dir: dir}
	// simulate pushed: set upstream to same commit
	run("branch", "-M", "main")
	run("remote", "add", "origin", dir) // local remote
	run("push", "-u", "origin", "main")

	findings := []detect.Finding{
		{
			Source: detect.SourceCommitTrailer, Selected: true,
			Match: "Co-authored-by: Claude <noreply@anthropic.com>",
			LineNumber: 3, RuleName: "ai-coauthor-trailer",
		},
	}
	_, err := Apply(g, Options{Findings: findings})
	if err == nil {
		t.Fatal("expected error when amending pushed commit without --force")
	}
}
