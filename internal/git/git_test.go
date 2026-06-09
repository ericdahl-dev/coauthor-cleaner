package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupTestRepo(t *testing.T) (string, Runner) {
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

	return dir, Runner{Dir: dir}
}

func TestStagedDiff_FindsAttribution(t *testing.T) {
	dir, r := setupTestRepo(t)
	path := filepath.Join(dir, "app.go")
	if err := os.WriteFile(path, []byte("package main\n\n// Generated with Claude Code\n"), 0644); err != nil {
		t.Fatal(err)
	}
	exec.Command("git", "-C", dir, "add", "app.go").Run()

	diff, err := r.StagedDiff()
	if err != nil {
		t.Fatal(err)
	}
	if !contains(diff, "Generated with Claude Code") {
		t.Errorf("staged diff missing attribution: %q", diff)
	}
}

func TestCommitMessage_HEAD(t *testing.T) {
	dir, r := setupTestRepo(t)
	path := filepath.Join(dir, "README.md")
	os.WriteFile(path, []byte("hello\n"), 0644)
	exec.Command("git", "-C", dir, "add", "README.md").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "init\n\nCo-authored-by: Claude <noreply@anthropic.com>").Run()

	msg, err := r.CommitMessage("HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if !contains(msg, "Co-authored-by: Claude") {
		t.Errorf("commit message = %q", msg)
	}
}

func TestStagedFileContents(t *testing.T) {
	dir, r := setupTestRepo(t)
	path := filepath.Join(dir, "foo.rb")
	os.WriteFile(path, []byte("# Generated with ChatGPT\nclass Foo\nend\n"), 0644)
	exec.Command("git", "-C", dir, "add", "foo.rb").Run()

	files, err := r.StagedFileContents()
	if err != nil {
		t.Fatal(err)
	}
	if files["foo.rb"] == "" {
		t.Error("expected staged file content")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || findSub(s, sub))
}

func findSub(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
