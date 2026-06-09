package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallHooks(t *testing.T) {
	dir, r := setupTestRepo(t)
	if err := r.InstallHooks("/usr/local/bin/coauthor-cleaner"); err != nil {
		t.Fatal(err)
	}

	gitDir, _ := r.GitDir()
	hookPath := filepath.Join(dir, gitDir, "hooks", "pre-commit")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	if !containsMarker(string(data)) {
		t.Error("hook missing coauthor-cleaner marker")
	}
}
