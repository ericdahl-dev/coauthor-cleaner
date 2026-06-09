package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Skeyelab/coauthor-cleaner/internal/clean"
	"github.com/Skeyelab/coauthor-cleaner/internal/config"
	"github.com/Skeyelab/coauthor-cleaner/internal/git"
	"github.com/Skeyelab/coauthor-cleaner/internal/scan"
)

func setupRepo(t *testing.T) (string, git.Runner) {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test User")
	return dir, git.Runner{Dir: dir}
}

func TestE2E_ScanAndCleanStaged(t *testing.T) {
	dir, g := setupRepo(t)
	path := filepath.Join(dir, "app.go")
	content := "package main\n\n// Generated with Claude Code\nfunc main() {}\n"
	os.WriteFile(path, []byte(content), 0644)
	exec.Command("git", "-C", dir, "add", "app.go").Run()

	cfg := config.Default()
	result, err := scan.Run(g, scan.Options{Staged: true, Config: cfg})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(result.Findings))
	}

	cleanResult, err := clean.Apply(g, clean.Options{Findings: result.Findings})
	if err != nil {
		t.Fatal(err)
	}
	if cleanResult.Applied != 1 {
		t.Fatalf("applied = %d", cleanResult.Applied)
	}

	files, err := g.StagedFileContents()
	if err != nil {
		t.Fatal(err)
	}
	staged := files["app.go"]
	if staged == content {
		t.Error("staged content should be cleaned")
	}
	if contains(staged, "Generated with Claude") {
		t.Errorf("attribution still present: %q", staged)
	}

	recheck, err := scan.Run(g, scan.Options{Staged: true, Config: cfg})
	if err != nil {
		t.Fatal(err)
	}
	if len(recheck.Findings) != 0 {
		t.Fatalf("expected clean staged, got %d findings", len(recheck.Findings))
	}
}

func TestE2E_CommitMessageClean(t *testing.T) {
	dir, g := setupRepo(t)
	readme := filepath.Join(dir, "README.md")
	os.WriteFile(readme, []byte("hi\n"), 0644)
	exec.Command("git", "-C", dir, "add", "README.md").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "init\n\nCo-authored-by: Claude <noreply@anthropic.com>").Run()

	cfg := config.Default()
	result, err := scan.Run(g, scan.Options{Commit: "HEAD", Config: cfg})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(result.Findings))
	}

	_, err = clean.Apply(g, clean.Options{Findings: result.Findings})
	if err != nil {
		t.Fatal(err)
	}

	msg, err := g.CommitMessage("HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if contains(msg, "Co-authored-by: Claude") {
		t.Errorf("trailer still present: %q", msg)
	}
}

func TestE2E_ConfigAllowedTrailer(t *testing.T) {
	dir, g := setupRepo(t)
	cfg := config.Default()
	cfg.AllowedTrailers = append(cfg.AllowedTrailers, "Co-authored-by: Eric Dahl")

	readme := filepath.Join(dir, "README.md")
	os.WriteFile(readme, []byte("x\n"), 0644)
	exec.Command("git", "-C", dir, "add", "README.md").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "init\n\nCo-authored-by: Eric Dahl <e@example.com>").Run()

	result, err := scan.Run(g, scan.Options{Commit: "HEAD", Config: cfg})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Findings) != 0 {
		t.Fatalf("human coauthor should be allowed, got %d findings", len(result.Findings))
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || findSub(s, sub))
}

func findSub(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
