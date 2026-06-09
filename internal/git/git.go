package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Runner struct {
	Dir string
}

func (r Runner) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if r.Dir != "" {
		cmd.Dir = r.Dir
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(out.String()), nil
}

func (r Runner) StagedDiff() (string, error) {
	return r.run("diff", "--cached")
}

func (r Runner) CommitMessage(ref string) (string, error) {
	if ref == "" {
		ref = "HEAD"
	}
	return r.run("log", "-1", "--format=%B", ref)
}

func (r Runner) StagedFileContents() (map[string]string, error) {
	names, err := r.run("diff", "--cached", "--name-only")
	if err != nil {
		return nil, err
	}
	files := make(map[string]string)
	for _, name := range strings.Split(strings.TrimSpace(names), "\n") {
		if name == "" {
			continue
		}
		content, err := r.run("show", ":"+name)
		if err != nil {
			return nil, err
		}
		files[name] = content
	}
	return files, nil
}

func (r Runner) StageFile(path, content string) error {
	cmd := exec.Command("git", "hash-object", "-w", "--stdin")
	if r.Dir != "" {
		cmd.Dir = r.Dir
	}
	cmd.Stdin = strings.NewReader(content)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return err
	}
	hash := strings.TrimSpace(out.String())
	_, err := r.run("update-index", "--cacheinfo", "100644,"+hash+","+path)
	return err
}

func (r Runner) AmendCommitMessage(message string) error {
	cmd := exec.Command("git", "commit", "--amend", "-m", message)
	if r.Dir != "" {
		cmd.Dir = r.Dir
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit --amend: %w: %s", err, stderr.String())
	}
	return nil
}

func (r Runner) InRepo() bool {
	_, err := r.run("rev-parse", "--git-dir")
	return err == nil
}

func (r Runner) RepoRoot() (string, error) {
	return r.run("rev-parse", "--show-toplevel")
}

func (r Runner) GitDir() (string, error) {
	return r.run("rev-parse", "--git-dir")
}
