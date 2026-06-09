package main

import (
	"os"
	"path/filepath"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/config"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/git"
)

func loadConfig(g git.Runner) config.Config {
	if !g.InRepo() {
		return config.Default()
	}
	root, err := g.RepoRoot()
	if err != nil {
		return config.Default()
	}
	cfg, _, err := config.LoadFromRepo(root)
	if err != nil {
		return config.Default()
	}
	return cfg
}

func executable() string {
	ex, err := os.Executable()
	if err != nil {
		return "coauthor-cleaner"
	}
	abs, err := filepath.Abs(ex)
	if err != nil {
		return ex
	}
	return abs
}
