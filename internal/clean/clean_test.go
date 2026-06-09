package clean

import (
	"strings"
	"testing"

	"github.com/Skeyelab/coauthor-cleaner/internal/config"
	"github.com/Skeyelab/coauthor-cleaner/internal/detect"
	"github.com/Skeyelab/coauthor-cleaner/internal/git"
	"github.com/Skeyelab/coauthor-cleaner/internal/scan"
	"github.com/Skeyelab/coauthor-cleaner/internal/testutil"
)

func TestApply_StagedUnpushedCommit(t *testing.T) {
	dir := testutil.GitRepo(t)
	g := git.Runner{Dir: dir}
	testutil.Write(t, dir, "README.md", "hi\n")
	testutil.Git(t, dir, "add", "README.md")
	testutil.Git(t, dir, "commit", "-m", "init\n\nCo-authored-by: Claude <noreply@anthropic.com>")

	result, err := scan.Run(g, scan.Options{Commit: "HEAD", Config: config.Default()})
	if err != nil {
		t.Fatal(err)
	}
	r, err := Apply(g, Options{Findings: result.Findings})
	if err != nil {
		t.Fatal(err)
	}
	if r.Applied != 1 || !r.Actions.AmendedCommit {
		t.Fatalf("%+v", r)
	}
	if !strings.Contains(r.Summary, "What to do next") {
		t.Fatal(r.Summary)
	}
	msg, _ := g.CommitMessage("HEAD")
	if strings.Contains(msg, "Claude") {
		t.Fatal(msg)
	}
}

func TestApply_BlocksAmendOfPushedCommit(t *testing.T) {
	dir := testutil.GitRepo(t)
	g := git.Runner{Dir: dir}
	testutil.Write(t, dir, "README.md", "hi\n")
	testutil.Git(t, dir, "add", "README.md")
	testutil.Git(t, dir, "commit", "-m", "init\n\nCo-authored-by: Claude <noreply@anthropic.com>")
	testutil.Git(t, dir, "branch", "-M", "main")
	testutil.Git(t, dir, "remote", "add", "origin", dir)
	testutil.Git(t, dir, "push", "-u", "origin", "main")

	findings := []detect.Finding{{
		Source: detect.SourceCommitTrailer, Selected: true, LineNumber: 3,
		Match: "Co-authored-by: Claude <noreply@anthropic.com>", RuleName: "ai-coauthor-trailer",
	}}
	_, err := Apply(g, Options{Findings: findings})
	if err == nil {
		t.Fatal("want error")
	}
}

func TestApply_ForceAmendPushed(t *testing.T) {
	dir := testutil.GitRepo(t)
	g := git.Runner{Dir: dir}
	testutil.Write(t, dir, "README.md", "hi\n")
	testutil.Git(t, dir, "add", "README.md")
	testutil.Git(t, dir, "commit", "-m", "init\n\nCo-authored-by: Claude <noreply@anthropic.com>")
	testutil.Git(t, dir, "branch", "-M", "main")
	testutil.Git(t, dir, "remote", "add", "origin", dir)
	testutil.Git(t, dir, "push", "-u", "origin", "main")

	findings := []detect.Finding{{
		Source: detect.SourceCommitTrailer, Selected: true, LineNumber: 3,
		Match: "Co-authored-by: Claude <noreply@anthropic.com>", RuleName: "ai-coauthor-trailer",
	}}
	r, err := Apply(g, Options{Findings: findings, ForceAmend: true})
	if err != nil {
		t.Fatal(err)
	}
	if r.Applied != 1 {
		t.Fatal(r)
	}
}
