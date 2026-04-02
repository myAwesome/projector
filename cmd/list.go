package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"project/internal/runner"
	"project/internal/store"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		projects, err := store.Load()
		if err != nil {
			return err
		}
		if len(projects) == 0 {
			fmt.Println("no projects registered")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tDESCRIPTION\tSTATUS\tPORTS\tSTARTED\tSCRIPT")
		for _, p := range projects {
			status := "stopped"
			ports := "-"
			started := "-"

			if rs, running := runner.IsRunning(p.Name); running {
				status = fmt.Sprintf("running (pid %d)", rs.PID)
				ports = runner.FormatPorts(runner.Ports(rs.PGID, rs.PID))
				started = rs.StartedAt.Format("15:04:05")
			}

			description := p.Description
			if description == "" {
				description = "-"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", p.Name, description, status, ports, started, p.Script)
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
