package cmd

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"project/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:          "project",
	Short:        "Interactive TUI for local web app projects",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
		_, err := p.Run()
		return err
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
