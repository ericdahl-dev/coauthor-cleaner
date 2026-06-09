package scan

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/git"
)

func TestRun_PRRange(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")

	main := filepath.Join(dir, "main.go")
	os.WriteFile(main, []byte("package main\n"), 0644)
	run("add", "main.go")
	run("commit", "-m", "init")

	g := git.Runner{Dir: dir}
	base := rev(t, dir)
	os.WriteFile(main, []byte("package main\n\n// Generated with Claude Code\n"), 0644)
	run("add", "main.go")
	run("commit", "-m", "add ai\n\nCo-authored-by: Claude <noreply@anthropic.com>")
	head := rev(t, dir)

	result, err := Run(g, Options{Base: base, Head: head})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Findings) < 2 {
		t.Fatalf("expected at least 2 findings, got %d: %+v", len(result.Findings), result.Findings)
	}
}

func rev(t *testing.T, dir string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return string(out[:len(out)-1]) // trim newline
}
