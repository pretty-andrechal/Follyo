package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ViewConfig holds configuration for a view.
type ViewConfig struct {
	Title       string
	LoadingText string
	HelpText    string
	Action      string // Menu action that triggers this view
}

// viewConfigs maps ViewType to its configuration.
var viewConfigs = map[ViewType]ViewConfig{
	ViewMenu: {
		Title:       "Follyo",
		LoadingText: "Loading...",
		HelpText:    "↑↓ Navigate | Enter Select | q Quit",
	},
	ViewSummary: {
		Title:       "Portfolio Summary",
		LoadingText: "Loading summary...",
		HelpText:    "r Refresh | esc Back | q Quit",
		Action:      "summary",
	},
	ViewBuy: {
		Title:       "Buy / Purchases",
		LoadingText: "Loading purchases...",
		HelpText:    "↑↓ Navigate | a Add | d Delete | esc Back | q Quit",
		Action:      "buy",
	},
	ViewSell: {
		Title:       "Sell / Sales",
		LoadingText: "Loading sales...",
		HelpText:    "↑↓ Navigate | a Add | d Delete | esc Back | q Quit",
		Action:      "sell",
	},
	ViewStake: {
		Title:       "Staking",
		LoadingText: "Loading stakes...",
		HelpText:    "↑↓ Navigate | a Add | d Unstake | esc Back | q Quit",
		Action:      "stake",
	},
	ViewLoan: {
		Title:       "Loans",
		LoadingText: "Loading loans...",
		HelpText:    "↑↓ Navigate | a Add | d Repay | esc Back | q Quit",
		Action:      "loan",
	},
	ViewSnapshots: {
		Title:       "Snapshots",
		LoadingText: "Loading snapshots...",
		HelpText:    "↑↓ Navigate | n New | d Delete | esc Back | q Quit",
		Action:      "snapshots",
	},
	ViewSettings: {
		Title:       "Settings",
		LoadingText: "Loading settings...",
		HelpText:    "↑↓ Navigate | Enter Toggle/Edit | esc Back | q Quit",
		Action:      "settings",
	},
	ViewTicker: {
		Title:       "Ticker Mappings",
		LoadingText: "Loading ticker mappings...",
		HelpText:    "↑↓ Navigate | a Add | s Search | d Delete | v Defaults | esc Back",
		Action:      "ticker",
	},
}

// ViewRegistry manages view models by ViewType.
type ViewRegistry struct {
	views map[ViewType]tea.Model
}

// NewViewRegistry creates a new view registry.
func NewViewRegistry() *ViewRegistry {
	return &ViewRegistry{
		views: make(map[ViewType]tea.Model),
	}
}

// Set stores a view model.
func (r *ViewRegistry) Set(viewType ViewType, model tea.Model) {
	r.views[viewType] = model
}

// Get retrieves a view model.
func (r *ViewRegistry) Get(viewType ViewType) tea.Model {
	return r.views[viewType]
}

// Has checks if a view model exists.
func (r *ViewRegistry) Has(viewType ViewType) bool {
	return r.views[viewType] != nil
}

// Update forwards a message to a view and stores the result.
func (r *ViewRegistry) Update(viewType ViewType, msg tea.Msg) tea.Cmd {
	model := r.views[viewType]
	if model == nil {
		return nil
	}
	var cmd tea.Cmd
	r.views[viewType], cmd = model.Update(msg)
	return cmd
}

// View renders a view model's content.
func (r *ViewRegistry) View(viewType ViewType) string {
	model := r.views[viewType]
	if model == nil {
		config := viewConfigs[viewType]
		if config.LoadingText != "" {
			return config.LoadingText
		}
		return "Loading..."
	}
	return model.View()
}

// Init calls Init on a view model.
func (r *ViewRegistry) Init(viewType ViewType) tea.Cmd {
	model := r.views[viewType]
	if model == nil {
		return nil
	}
	return model.Init()
}

// GetConfig returns the configuration for a view type.
func GetViewConfig(viewType ViewType) ViewConfig {
	return viewConfigs[viewType]
}

// GetViewTypeForAction returns the ViewType for a menu action.
func GetViewTypeForAction(action string) (ViewType, bool) {
	for vt, config := range viewConfigs {
		if config.Action == action {
			return vt, true
		}
	}
	return ViewMenu, false
}

// AllContentViews returns all view types that handle their own layout.
// These views don't need the wrapper styling.
var AllContentViews = []ViewType{
	ViewSummary,
	ViewSettings,
	ViewSnapshots,
	ViewBuy,
	ViewSell,
	ViewStake,
	ViewLoan,
	ViewTicker,
}

// IsContentView checks if a view handles its own layout.
func IsContentView(viewType ViewType) bool {
	for _, vt := range AllContentViews {
		if vt == viewType {
			return true
		}
	}
	return false
}
