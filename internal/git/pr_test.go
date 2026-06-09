package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseAddedLines(t *testing.T) {
	diff := `diff --git a/app.go b/app.go
index abc..def 100644
--- a/app.go
+++ b/app.go
@@ -1,2 +1,3 @@
 package main
+// Generated with Claude Code
 func main() {}
`
	lines := parseAddedLines(diff)
	if len(lines) != 1 {
		t.Fatalf("expected 1 added line, got %d", len(lines))
	}
	if lines[0].FilePath != "app.go" {
		t.Errorf("file = %q", lines[0].FilePath)
	}
	if lines[0].Line != "// Generated with Claude Code" {
		t.Errorf("line = %q", lines[0].Line)
	}
}

func TestDiffAddedLines_InRepo(t *testing.T) {
	dir, r := setupTestRepo(t)

	main := filepath.Join(dir, "main.go")
	os.WriteFile(main, []byte("package main\n"), 0644)
	exec.Command("git", "-C", dir, "add", "main.go").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "init").Run()
	base, _ := r.run("rev-parse", "HEAD")

	os.WriteFile(main, []byte("package main\n\n// Generated with Claude Code\n"), 0644)
	exec.Command("git", "-C", dir, "add", "main.go").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "add attribution").Run()
	head, _ := r.run("rev-parse", "HEAD")

	lines, err := r.DiffAddedLines(base, head)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d: %+v", len(lines), lines)
	}
}

func TestCommitMessagesInRange(t *testing.T) {
	dir, r := setupTestRepo(t)
	readme := filepath.Join(dir, "README.md")
	os.WriteFile(readme, []byte("hi\n"), 0644)
	exec.Command("git", "-C", dir, "add", "README.md").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "init").Run()
	base, _ := r.run("rev-parse", "HEAD")

	os.WriteFile(readme, []byte("hi\nmore\n"), 0644)
	exec.Command("git", "-C", dir, "add", "README.md").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "feat: thing\n\nCo-authored-by: Claude <noreply@anthropic.com>").Run()
	head, _ := r.run("rev-parse", "HEAD")

	msgs, err := r.CommitMessagesInRange(base, head)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(msgs))
	}
	for _, msg := range msgs {
		if !findSub(msg, "Co-authored-by: Claude") {
			t.Errorf("missing trailer in %q", msg)
		}
	}
}
