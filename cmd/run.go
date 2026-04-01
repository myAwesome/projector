package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"proj/internal/runner"
	"proj/internal/store"
)

var runCmd = &cobra.Command{
	Use:   "run <name>",
	Short: "Start a registered project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		p, err := store.Find(name)
		if errors.Is(err, store.ErrNotFound) {
			return fmt.Errorf("project %q not found", name)
		}
		if err != nil {
			return err
		}

		rs, err := runner.Start(p)
		if errors.Is(err, runner.ErrAlreadyRunning) {
			return fmt.Errorf("project %q is already running", name)
		}
		if err != nil {
			return err
		}

		fmt.Printf("started %q (pid %d)\n", name, rs.PID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
