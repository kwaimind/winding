package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/kwaimind/winding/internal/config"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <path>",
	Aliases: []string{"rm"},
	Short:   "Stop tracking a repo",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}

		removed, err := config.RemoveRepo(path)
		if err != nil {
			return err
		}
		if !removed {
			fmt.Printf("not tracked: %s\n", path)
			return nil
		}

		fmt.Printf("removed: %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
