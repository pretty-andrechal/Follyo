// Package tui provides an interactive terminal user interface for Follyo.
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pretty-andrechal/follyo/internal/portfolio"
	"github.com/pretty-andrechal/follyo/internal/storage"
)

// ViewType represents the current view being displayed.
type ViewType int

const (
	ViewMenu ViewType = iota
	ViewSummary
	ViewBuy
	ViewSell
	ViewStake
	ViewLoan
	ViewSnapshots
	ViewSettings
	ViewTicker
	ViewCoinHistory
)

// App is the main application model.
type App struct {
	portfolio      *portfolio.Portfolio
	storage        *storage.Storage
	currentView    ViewType
	views          *ViewRegistry
	tickerMappings map[string]string
	width          int
	height         int
	err            error
	quitting       bool
	statusMsg      string
}

// NewApp creates a new application instance.
func NewApp(s *storage.Storage, p *portfolio.Portfolio) *App {
	return &App{
		storage:        s,
		portfolio:      p,
		currentView:    ViewMenu,
		views:          NewViewRegistry(),
		tickerMappings: make(map[string]string),
	}
}

// Init initializes the application.
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tea.EnableMouseCellMotion,
	)
}

// Update handles messages for the application.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit handling
		if msg.String() == "ctrl+c" {
			a.quitting = true
			return a, tea.Quit
		}

		// Handle menu specially (needs q to quit)
		if a.currentView == ViewMenu {
			return a.updateMenu(msg)
		}

		// Forward to current view via registry
		if a.views.Has(a.currentView) {
			cmd := a.views.Update(a.currentView, msg)
			return a, cmd
		}

		// For unimplemented views, go back to menu on any key
		a.currentView = ViewMenu
		a.statusMsg = ""
		return a, nil

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

		// Adjust height for status bar for content views
		adjustedMsg := msg
		if a.currentView != ViewMenu {
			adjustedMsg = tea.WindowSizeMsg{
				Width:  msg.Width,
				Height: msg.Height - 1,
			}
		}

		// Forward to current view
		cmd := a.views.Update(a.currentView, adjustedMsg)
		return a, cmd

	case MenuSelectMsg:
		return a.handleMenuSelect(msg)

	case BackToMenuMsg:
		a.currentView = ViewMenu
		a.statusMsg = ""
		// Clear summary model to force refresh on next visit
		a.views.Set(ViewSummary, nil)
		return a, nil

	case errMsg:
		a.err = msg.err
		return a, nil

	case statusMsg:
		a.statusMsg = string(msg)
		return a, nil

	default:
		// Forward other messages to current view
		if a.views.Has(a.currentView) {
			cmd := a.views.Update(a.currentView, msg)
			return a, cmd
		}
	}

	return a, nil
}

type errMsg struct{ err error }
type statusMsg string

func (a *App) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle quit from menu
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "q" {
			a.quitting = true
			return a, tea.Quit
		}
	}

	cmd := a.views.Update(ViewMenu, msg)
	return a, cmd
}

func (a *App) handleMenuSelect(msg MenuSelectMsg) (tea.Model, tea.Cmd) {
	// Look up the view type for this action
	viewType, found := GetViewTypeForAction(msg.Action)
	if !found {
		return a, nil
	}

	// Switch to the selected view
	a.currentView = viewType
	a.statusMsg = ""

	// Initialize the view and send window size
	if a.views.Has(viewType) {
		initCmd := a.views.Init(viewType)
		if a.width > 0 && a.height > 0 {
			sizeMsg := tea.WindowSizeMsg{Width: a.width, Height: a.height - 1}
			a.views.Update(viewType, sizeMsg)
		}
		return a, initCmd
	}

	return a, nil
}

// View renders the application.
func (a *App) View() string {
	if a.quitting {
		return ""
	}

	// Get content from registry (handles loading text automatically)
	content := a.views.View(a.currentView)
	if !a.views.Has(a.currentView) && a.currentView != ViewMenu {
		// Placeholder for unimplemented views
		content = a.renderPlaceholder()
	}

	// Before window size is known, just return content without layout
	if a.width == 0 || a.height == 0 {
		return content
	}

	// Add status bar at the bottom
	statusBar := a.renderStatusBar()

	// Content views handle their own layout, don't wrap them
	if IsContentView(a.currentView) {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			statusBar,
		)
	}

	// Calculate available height for content
	statusHeight := lipgloss.Height(statusBar)
	contentHeight := a.height - statusHeight
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Style content area (for menu and other views)
	contentStyle := lipgloss.NewStyle().
		Width(a.width).
		Height(contentHeight).
		Align(lipgloss.Center, lipgloss.Center)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		contentStyle.Render(content),
		statusBar,
	)
}

func (a *App) renderPlaceholder() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(2, 4)

	title := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Render(a.getViewTitle())

	msg := lipgloss.NewStyle().
		Foreground(SubtleTextColor).
		Render("\n\nThis view is coming soon!\n\nPress any key to go back.")

	return boxStyle.Render(title + msg)
}

func (a *App) getViewTitle() string {
	config := GetViewConfig(a.currentView)
	if config.Title != "" {
		return config.Title
	}
	return "Follyo"
}

func (a *App) renderStatusBar() string {
	style := lipgloss.NewStyle().
		Foreground(SubtleTextColor).
		Background(SurfaceColor).
		Width(a.width).
		Padding(0, 1)

	left := "FOLLYO"

	// Get help text from view config
	config := GetViewConfig(a.currentView)
	right := config.HelpText
	if right == "" {
		right = "↑↓ Navigate | Enter Select | q Quit"
	}

	if a.statusMsg != "" {
		left = a.statusMsg
	}

	if a.err != nil {
		left = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Render(fmt.Sprintf("Error: %v", a.err))
	}

	// Calculate spacing
	spacing := a.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if spacing < 1 {
		spacing = 1
	}

	content := left + lipgloss.NewStyle().Width(spacing).Render("") + right

	return style.Render(content)
}

// SetMenuModel sets the menu model for the app.
func (a *App) SetMenuModel(m tea.Model) {
	a.views.Set(ViewMenu, m)
}

// SetSummaryModel sets the summary model for the app.
func (a *App) SetSummaryModel(m tea.Model) {
	a.views.Set(ViewSummary, m)
}

// SetSettingsModel sets the settings model for the app.
func (a *App) SetSettingsModel(m tea.Model) {
	a.views.Set(ViewSettings, m)
}

// SetSnapshotsModel sets the snapshots model for the app.
func (a *App) SetSnapshotsModel(m tea.Model) {
	a.views.Set(ViewSnapshots, m)
}

// SetBuyModel sets the buy model for the app.
func (a *App) SetBuyModel(m tea.Model) {
	a.views.Set(ViewBuy, m)
}

// SetSellModel sets the sell model for the app.
func (a *App) SetSellModel(m tea.Model) {
	a.views.Set(ViewSell, m)
}

// SetStakeModel sets the stake model for the app.
func (a *App) SetStakeModel(m tea.Model) {
	a.views.Set(ViewStake, m)
}

// SetLoanModel sets the loan model for the app.
func (a *App) SetLoanModel(m tea.Model) {
	a.views.Set(ViewLoan, m)
}

// SetTickerModel sets the ticker model for the app.
func (a *App) SetTickerModel(m tea.Model) {
	a.views.Set(ViewTicker, m)
}

// SetCoinHistoryModel sets the coin history model for the app.
func (a *App) SetCoinHistoryModel(m tea.Model) {
	a.views.Set(ViewCoinHistory, m)
}

// SetTickerMappings sets the custom ticker mappings for price fetching.
func (a *App) SetTickerMappings(mappings map[string]string) {
	a.tickerMappings = mappings
}

// GetPortfolio returns the portfolio instance.
func (a *App) GetPortfolio() *portfolio.Portfolio {
	return a.portfolio
}

// GetTickerMappings returns the custom ticker mappings.
func (a *App) GetTickerMappings() map[string]string {
	return a.tickerMappings
}
