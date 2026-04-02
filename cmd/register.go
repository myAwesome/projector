package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"project/internal/store"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new project",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		dir, _ := cmd.Flags().GetString("dir")
		script, _ := cmd.Flags().GetString("script")
		description, _ := cmd.Flags().GetString("description")

		if name == "" || dir == "" || script == "" {
			return fmt.Errorf("--name, --dir, and --script are required")
		}

		absDir, err := filepath.Abs(dir)
		if err != nil {
			return err
		}

		err = store.Add(store.Project{
			Name:        name,
			Description: description,
			Dir:         absDir,
			Script:      script,
		})
		if errors.Is(err, store.ErrExists) {
			return fmt.Errorf("project %q already exists", name)
		}
		if err != nil {
			return err
		}

		fmt.Printf("registered %q (%s)\n", name, absDir)
		return nil
	},
}

func init() {
	registerCmd.Flags().String("name", "", "project name")
	registerCmd.Flags().String("dir", "", "project directory")
	registerCmd.Flags().String("script", "", "launch script or command")
	registerCmd.Flags().StringP("description", "d", "", "short project description")
	rootCmd.AddCommand(registerCmd)
}
