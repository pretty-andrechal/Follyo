package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui/views"
	"github.com/pretty-andrechal/follyo/internal/storage"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive terminal UI",
	Long:  "Launch an interactive terminal user interface for managing your portfolio with keyboard navigation and live updates.",
	Run:   runTUI,
}

func runTUI(cmd *cobra.Command, args []string) {
	// Get storage from the initialized portfolio
	s, err := storage.New(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	// Create the app
	app := tui.NewApp(s, p)

	// Create and set the menu model
	menuModel := views.NewMenuModel()
	app.SetMenuModel(menuModel)

	// Run the TUI program
	program := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
