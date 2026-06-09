package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPush_LocalBareRemote(t *testing.T) {
	dir := t.TempDir()
	bare := filepath.Join(dir, "bare.git")
	exec.Command("git", "init", "--bare", bare).Run()

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
	os.WriteFile(filepath.Join(dir, "f"), []byte("a\n"), 0644)
	run("add", "f")
	run("commit", "-m", "init")
	run("branch", "-M", "main")
	run("remote", "add", "origin", bare)
	run("push", "-u", "origin", "main")

	os.WriteFile(filepath.Join(dir, "f"), []byte("b\n"), 0644)
	run("add", "f")
	run("commit", "-m", "second")

	g := Runner{Dir: dir}
	if err := g.Push(); err != nil {
		t.Fatal(err)
	}
}
