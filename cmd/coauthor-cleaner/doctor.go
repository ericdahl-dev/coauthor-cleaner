package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/config"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/github"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/git"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/guide"
	"github.com/spf13/cobra"
)

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check that coauthor-cleaner is set up correctly",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := git.Runner{}
			inRepo := g.InRepo()

			configPath := "(none)"
			configOK := false
			hooksOK := false
			if inRepo {
				root, err := g.RepoRoot()
				if err == nil {
					configPath = filepath.Join(root, config.FileName)
					_, err := os.Stat(configPath)
					configOK = err == nil
				}
				hooksOK = g.HooksInstalled()
			}

			ghOK := github.Client{}.Available()
			fmt.Print(guide.DoctorReport(inRepo, ghOK, hooksOK, configOK, configPath))
			return nil
		},
	}
}
