package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwaimind/winding/internal/config"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "Start tracking a repo (must contain package.json)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}

		if _, err := os.Stat(filepath.Join(path, "package.json")); err != nil {
			return fmt.Errorf("no package.json found at %s", path)
		}

		added, err := config.AddRepo(path)
		if err != nil {
			return err
		}
		if !added {
			fmt.Printf("already tracked: %s\n", path)
			return nil
		}

		fmt.Printf("added: %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
