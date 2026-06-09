package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const hookMarker = "# coauthor-cleaner"

var hookTypes = []string{"pre-commit", "commit-msg"}

func (r Runner) InstallHooks(binary string) error {
	gitDir, err := r.GitDir()
	if err != nil {
		return err
	}
	hooksDir := gitDir
	if !filepath.IsAbs(hooksDir) {
		root, err := r.RepoRoot()
		if err != nil {
			return err
		}
		hooksDir = filepath.Join(root, gitDir)
	}
	hooksDir = filepath.Join(hooksDir, "hooks")

	for _, name := range hookTypes {
		path := filepath.Join(hooksDir, name)
		script := hookScript(binary, name)
		if err := mergeHook(path, script); err != nil {
			return err
		}
	}
	return nil
}

func hookScript(binary, hookType string) string {
	if hookType == "commit-msg" {
		return fmt.Sprintf(`#!/bin/sh
%s
exec "%s" hook run commit-msg "$1"
`, hookMarker, binary)
	}
	return fmt.Sprintf(`#!/bin/sh
%s
exec "%s" hook run pre-commit
`, hookMarker, binary)
}

func mergeHook(path, script string) error {
	existing, err := os.ReadFile(path)
	if err == nil && containsMarker(string(existing)) {
		return nil // already installed
	}
	if err == nil && len(existing) > 0 {
		script = string(existing) + "\n" + script
	}
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		return err
	}
	return nil
}

func containsMarker(s string) bool {
	return strings.Contains(s, hookMarker)
}
