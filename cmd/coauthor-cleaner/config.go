package main

import (
	"fmt"
	"os"

	"github.com/Skeyelab/coauthor-cleaner/internal/config"
	"github.com/spf13/cobra"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage coauthor-cleaner configuration",
	}
	cmd.AddCommand(configInitCmd())
	return cmd
}

func configInitCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a .coauthor-cleaner.yml config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := gitRunner()
			if err != nil {
				return err
			}
			root, err := g.RepoRoot()
			if err != nil {
				return err
			}
			path := root + "/" + config.FileName
			if _, err := os.Stat(path); err == nil && !force {
				return fmt.Errorf("%s already exists (use --force to overwrite)", path)
			}
			return os.WriteFile(path, []byte(config.DefaultYAML()), 0644)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing config")
	return cmd
}
