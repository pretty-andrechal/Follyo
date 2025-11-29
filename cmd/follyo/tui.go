package main

import (
	"fmt"
	"os"
	"path/filepath"

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

	// Load config for ticker mappings
	cfg := loadConfig()
	tickerMappings := cfg.GetAllTickerMappings()

	// Create the app
	app := tui.NewApp(s, p)
	app.SetTickerMappings(tickerMappings)

	// Create and set the menu model
	menuModel := views.NewMenuModel()
	app.SetMenuModel(menuModel)

	// Create and set the summary model
	summaryModel := views.NewSummaryModel(p, tickerMappings)
	app.SetSummaryModel(summaryModel)

	// Create and set the settings model
	settingsModel := views.NewSettingsModel(cfg)
	app.SetSettingsModel(settingsModel)

	// Create snapshot store and model
	snapshotPath := filepath.Join(filepath.Dir(dataPath), "snapshots.json")
	snapshotStore, err := storage.NewSnapshotStore(snapshotPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing snapshot store: %v\n", err)
		os.Exit(1)
	}
	snapshotsModel := views.NewSnapshotsModel(snapshotStore, p, tickerMappings)
	app.SetSnapshotsModel(snapshotsModel)

	// Create coin history model
	coinHistoryModel := views.NewCoinHistoryModel(snapshotStore)
	app.SetCoinHistoryModel(coinHistoryModel)

	// Create buy model
	buyModel := views.NewBuyModel(p, cfg.GetDefaultPlatform())
	app.SetBuyModel(buyModel)

	// Create sell model
	sellModel := views.NewSellModel(p, cfg.GetDefaultPlatform())
	app.SetSellModel(sellModel)

	// Create stake model
	stakeModel := views.NewStakeModel(p, cfg.GetDefaultPlatform())
	app.SetStakeModel(stakeModel)

	// Create loan model
	loanModel := views.NewLoanModel(p, cfg.GetDefaultPlatform())
	app.SetLoanModel(loanModel)

	// Create ticker model
	tickerModel := views.NewTickerModel(cfg)
	app.SetTickerModel(tickerModel)

	// Run the TUI program
	program := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
