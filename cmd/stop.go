package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"proj/internal/runner"
)

var stopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a running project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		err := runner.Stop(name)
		if errors.Is(err, runner.ErrNotRunning) {
			return fmt.Errorf("project %q is not running", name)
		}
		if err != nil {
			return err
		}

		fmt.Printf("stopped %q\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
