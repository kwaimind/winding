package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwaimind/winding/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Show tracked repos",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repos, err := config.Repos()
		if err != nil {
			return err
		}

		if len(repos) == 0 {
			fmt.Println("no repos tracked yet — run `winding add <path>` to add one")
			return nil
		}

		for _, repo := range repos {
			if _, err := os.Stat(filepath.Join(repo, "package.json")); err != nil {
				fmt.Printf("%s  (missing package.json)\n", repo)
			} else {
				fmt.Println(repo)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
