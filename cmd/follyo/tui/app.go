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
)

// App is the main application model.
type App struct {
	portfolio      *portfolio.Portfolio
	storage        *storage.Storage
	currentView    ViewType
	menuModel      tea.Model
	summaryModel   tea.Model
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

		// Handle based on current view
		switch a.currentView {
		case ViewMenu:
			return a.updateMenu(msg)
		case ViewSummary:
			return a.updateSummary(msg)
		default:
			// For unimplemented views, go back to menu on any key
			a.currentView = ViewMenu
			a.statusMsg = ""
			return a, nil
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Forward to current view
		var cmd tea.Cmd
		switch a.currentView {
		case ViewMenu:
			if a.menuModel != nil {
				a.menuModel, cmd = a.menuModel.Update(msg)
			}
		case ViewSummary:
			if a.summaryModel != nil {
				// Adjust height for status bar (1 line)
				adjustedMsg := tea.WindowSizeMsg{
					Width:  msg.Width,
					Height: msg.Height - 1,
				}
				a.summaryModel, cmd = a.summaryModel.Update(adjustedMsg)
			}
		}
		return a, cmd

	case MenuSelectMsg:
		return a.handleMenuSelect(msg)

	case BackToMenuMsg:
		a.currentView = ViewMenu
		a.statusMsg = ""
		a.summaryModel = nil
		return a, nil

	case errMsg:
		a.err = msg.err
		return a, nil

	case statusMsg:
		a.statusMsg = string(msg)
		return a, nil

	default:
		// Forward other messages to current view
		switch a.currentView {
		case ViewSummary:
			if a.summaryModel != nil {
				var cmd tea.Cmd
				a.summaryModel, cmd = a.summaryModel.Update(msg)
				return a, cmd
			}
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

	if a.menuModel == nil {
		return a, nil
	}

	var cmd tea.Cmd
	a.menuModel, cmd = a.menuModel.Update(msg)
	return a, cmd
}

func (a *App) updateSummary(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.summaryModel == nil {
		return a, nil
	}

	var cmd tea.Cmd
	a.summaryModel, cmd = a.summaryModel.Update(msg)
	return a, cmd
}

func (a *App) handleMenuSelect(msg MenuSelectMsg) (tea.Model, tea.Cmd) {
	switch msg.Action {
	case "summary":
		a.currentView = ViewSummary
		a.statusMsg = ""
		if a.summaryModel != nil {
			// Initialize and send window size so viewport can be set up
			initCmd := a.summaryModel.Init()
			// Send current window size to summary model (adjusted for status bar)
			if a.width > 0 && a.height > 0 {
				sizeMsg := tea.WindowSizeMsg{Width: a.width, Height: a.height - 1}
				a.summaryModel, _ = a.summaryModel.Update(sizeMsg)
			}
			return a, initCmd
		}
	case "buy":
		a.currentView = ViewBuy
		a.statusMsg = "Buy view - Press any key to go back (coming soon)"
	case "sell":
		a.currentView = ViewSell
		a.statusMsg = "Sell view - Press any key to go back (coming soon)"
	case "stake":
		a.currentView = ViewStake
		a.statusMsg = "Stake view - Press any key to go back (coming soon)"
	case "loan":
		a.currentView = ViewLoan
		a.statusMsg = "Loan view - Press any key to go back (coming soon)"
	case "snapshots":
		a.currentView = ViewSnapshots
		a.statusMsg = "Snapshots view - Press any key to go back (coming soon)"
	case "settings":
		a.currentView = ViewSettings
		a.statusMsg = "Settings view - Press any key to go back (coming soon)"
	}
	return a, nil
}

// View renders the application.
func (a *App) View() string {
	if a.quitting {
		return ""
	}

	var content string

	switch a.currentView {
	case ViewMenu:
		if a.menuModel != nil {
			content = a.menuModel.View()
		} else {
			content = "Loading..."
		}
	case ViewSummary:
		if a.summaryModel != nil {
			content = a.summaryModel.View()
		} else {
			content = "Loading summary..."
		}
	default:
		// Placeholder for unimplemented views
		content = a.renderPlaceholder()
	}

	// Before window size is known, just return content without layout
	if a.width == 0 || a.height == 0 {
		return content
	}

	// Add status bar at the bottom
	statusBar := a.renderStatusBar()

	// Summary view handles its own layout, don't wrap it
	if a.currentView == ViewSummary {
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
	switch a.currentView {
	case ViewSummary:
		return "Portfolio Summary"
	case ViewBuy:
		return "Buy / Purchases"
	case ViewSell:
		return "Sell / Sales"
	case ViewStake:
		return "Staking"
	case ViewLoan:
		return "Loans"
	case ViewSnapshots:
		return "Snapshots"
	case ViewSettings:
		return "Settings"
	default:
		return "Follyo"
	}
}

func (a *App) renderStatusBar() string {
	style := lipgloss.NewStyle().
		Foreground(SubtleTextColor).
		Background(SurfaceColor).
		Width(a.width).
		Padding(0, 1)

	left := "FOLLYO"
	right := "↑↓ Navigate | Enter Select | q Quit"

	if a.currentView == ViewSummary {
		right = "r Refresh | esc Back | q Quit"
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
	a.menuModel = m
}

// SetSummaryModel sets the summary model for the app.
func (a *App) SetSummaryModel(m tea.Model) {
	a.summaryModel = m
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
