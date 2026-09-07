// Package cmd implements winding's CLI commands.
package cmd

import (
	"fmt"

	"github.com/kwaimind/winding/internal/config"
	"github.com/kwaimind/winding/internal/gitops"
	"github.com/kwaimind/winding/internal/yarnbump"
	"github.com/spf13/cobra"
)

var gitFlag bool

var rootCmd = &cobra.Command{
	Use:   "winding",
	Short: "Bump every tracked repo to the latest Yarn and install",
	Long: fmt.Sprintf(`winding bumps every tracked repo to the latest Yarn (via Corepack) and
runs an install.

Config file: %s (edit by hand anytime, or use the subcommands below)`, mustConfigPath()),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAll()
	},
}

func init() {
	rootCmd.Flags().BoolVar(&gitFlag, "git", false, "commit each bumped repo's changes to a new branch")
}

// Execute runs the root command, exiting non-zero on failure.
func Execute() error {
	return rootCmd.Execute()
}

func mustConfigPath() string {
	path, err := config.Path()
	if err != nil {
		return "~/.config/winding/config.yaml"
	}
	return path
}

func runAll() error {
	repos, err := config.Repos()
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		fmt.Println("no repos tracked yet — run `winding add <path>` to add one")
		return nil
	}

	failures := 0
	for _, repo := range repos {
		fmt.Printf("==> %s\n", repo)
		result := yarnbump.Bump(repo)
		if result.OK {
			fmt.Printf("    ok: %s\n", result.Message)
			if gitFlag {
				commitResult, err := gitops.CommitChanges(repo)
				switch {
				case err != nil:
					fmt.Printf("    git: FAILED: %v\n", err)
				case commitResult.Committed:
					fmt.Printf("    git: committed to branch %s\n", commitResult.Branch)
				default:
					fmt.Println("    git: nothing to commit")
				}
			}
		} else {
			failures++
			fmt.Printf("    FAILED: %s\n", result.Message)
		}
	}

	fmt.Printf("\n%d/%d repos bumped and installed successfully\n", len(repos)-failures, len(repos))
	if failures > 0 {
		return fmt.Errorf("%d repo(s) failed", failures)
	}
	return nil
}
