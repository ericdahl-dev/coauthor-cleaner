package main

import (
	"fmt"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/detect"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/git"
	"github.com/spf13/cobra"
)

func rulesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rules",
		Short: "List detection rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := gitRunner()
			if err != nil {
				// rules list works outside repo too
				g = git.Runner{}
			}
			cfg := loadConfig(g)
			rules := detect.ListRules(cfg)
			fmt.Println("Active detection rules:")
			for _, r := range rules {
				fmt.Printf("  %s (%s)\n", r.Name, r.Confidence)
			}
			return nil
		},
	}
}
