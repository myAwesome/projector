package cmd

import (
	"fmt"
	"text/tabwriter"
	"os"

	"github.com/spf13/cobra"
	"proj/internal/store"
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
		fmt.Fprintln(w, "NAME\tDIR\tSCRIPT")
		for _, p := range projects {
			fmt.Fprintf(w, "%s\t%s\t%s\n", p.Name, p.Dir, p.Script)
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
