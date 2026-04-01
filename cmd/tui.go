package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"proj/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Start interactive TUI",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
		_, err := p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
