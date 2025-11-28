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
)

// App is the main application model.
type App struct {
	portfolio      *portfolio.Portfolio
	storage        *storage.Storage
	currentView    ViewType
	menuModel      tea.Model
	summaryModel   tea.Model
	settingsModel  tea.Model
	snapshotsModel tea.Model
	buyModel       tea.Model
	sellModel      tea.Model
	stakeModel     tea.Model
	loanModel      tea.Model
	tickerModel    tea.Model
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
		case ViewSettings:
			return a.updateSettings(msg)
		case ViewSnapshots:
			return a.updateSnapshots(msg)
		case ViewBuy:
			return a.updateBuy(msg)
		case ViewSell:
			return a.updateSell(msg)
		case ViewStake:
			return a.updateStake(msg)
		case ViewLoan:
			return a.updateLoan(msg)
		case ViewTicker:
			return a.updateTicker(msg)
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
		case ViewSettings:
			if a.settingsModel != nil {
				a.settingsModel, cmd = a.settingsModel.Update(msg)
			}
		case ViewSnapshots:
			if a.snapshotsModel != nil {
				a.snapshotsModel, cmd = a.snapshotsModel.Update(msg)
			}
		case ViewBuy:
			if a.buyModel != nil {
				a.buyModel, cmd = a.buyModel.Update(msg)
			}
		case ViewSell:
			if a.sellModel != nil {
				a.sellModel, cmd = a.sellModel.Update(msg)
			}
		case ViewStake:
			if a.stakeModel != nil {
				a.stakeModel, cmd = a.stakeModel.Update(msg)
			}
		case ViewLoan:
			if a.loanModel != nil {
				a.loanModel, cmd = a.loanModel.Update(msg)
			}
		case ViewTicker:
			if a.tickerModel != nil {
				a.tickerModel, cmd = a.tickerModel.Update(msg)
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
		case ViewSettings:
			if a.settingsModel != nil {
				var cmd tea.Cmd
				a.settingsModel, cmd = a.settingsModel.Update(msg)
				return a, cmd
			}
		case ViewSnapshots:
			if a.snapshotsModel != nil {
				var cmd tea.Cmd
				a.snapshotsModel, cmd = a.snapshotsModel.Update(msg)
				return a, cmd
			}
		case ViewBuy:
			if a.buyModel != nil {
				var cmd tea.Cmd
				a.buyModel, cmd = a.buyModel.Update(msg)
				return a, cmd
			}
		case ViewSell:
			if a.sellModel != nil {
				var cmd tea.Cmd
				a.sellModel, cmd = a.sellModel.Update(msg)
				return a, cmd
			}
		case ViewStake:
			if a.stakeModel != nil {
				var cmd tea.Cmd
				a.stakeModel, cmd = a.stakeModel.Update(msg)
				return a, cmd
			}
		case ViewLoan:
			if a.loanModel != nil {
				var cmd tea.Cmd
				a.loanModel, cmd = a.loanModel.Update(msg)
				return a, cmd
			}
		case ViewTicker:
			if a.tickerModel != nil {
				var cmd tea.Cmd
				a.tickerModel, cmd = a.tickerModel.Update(msg)
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

func (a *App) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.settingsModel == nil {
		return a, nil
	}

	var cmd tea.Cmd
	a.settingsModel, cmd = a.settingsModel.Update(msg)
	return a, cmd
}

func (a *App) updateSnapshots(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.snapshotsModel == nil {
		return a, nil
	}

	var cmd tea.Cmd
	a.snapshotsModel, cmd = a.snapshotsModel.Update(msg)
	return a, cmd
}

func (a *App) updateBuy(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.buyModel == nil {
		return a, nil
	}

	var cmd tea.Cmd
	a.buyModel, cmd = a.buyModel.Update(msg)
	return a, cmd
}

func (a *App) updateSell(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.sellModel == nil {
		return a, nil
	}

	var cmd tea.Cmd
	a.sellModel, cmd = a.sellModel.Update(msg)
	return a, cmd
}

func (a *App) updateStake(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.stakeModel == nil {
		return a, nil
	}

	var cmd tea.Cmd
	a.stakeModel, cmd = a.stakeModel.Update(msg)
	return a, cmd
}

func (a *App) updateLoan(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.loanModel == nil {
		return a, nil
	}

	var cmd tea.Cmd
	a.loanModel, cmd = a.loanModel.Update(msg)
	return a, cmd
}

func (a *App) updateTicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.tickerModel == nil {
		return a, nil
	}

	var cmd tea.Cmd
	a.tickerModel, cmd = a.tickerModel.Update(msg)
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
		a.statusMsg = ""
		if a.buyModel != nil {
			initCmd := a.buyModel.Init()
			if a.width > 0 && a.height > 0 {
				sizeMsg := tea.WindowSizeMsg{Width: a.width, Height: a.height - 1}
				a.buyModel, _ = a.buyModel.Update(sizeMsg)
			}
			return a, initCmd
		}
	case "sell":
		a.currentView = ViewSell
		a.statusMsg = ""
		if a.sellModel != nil {
			initCmd := a.sellModel.Init()
			if a.width > 0 && a.height > 0 {
				sizeMsg := tea.WindowSizeMsg{Width: a.width, Height: a.height - 1}
				a.sellModel, _ = a.sellModel.Update(sizeMsg)
			}
			return a, initCmd
		}
	case "stake":
		a.currentView = ViewStake
		a.statusMsg = ""
		if a.stakeModel != nil {
			initCmd := a.stakeModel.Init()
			if a.width > 0 && a.height > 0 {
				sizeMsg := tea.WindowSizeMsg{Width: a.width, Height: a.height - 1}
				a.stakeModel, _ = a.stakeModel.Update(sizeMsg)
			}
			return a, initCmd
		}
	case "loan":
		a.currentView = ViewLoan
		a.statusMsg = ""
		if a.loanModel != nil {
			initCmd := a.loanModel.Init()
			if a.width > 0 && a.height > 0 {
				sizeMsg := tea.WindowSizeMsg{Width: a.width, Height: a.height - 1}
				a.loanModel, _ = a.loanModel.Update(sizeMsg)
			}
			return a, initCmd
		}
	case "snapshots":
		a.currentView = ViewSnapshots
		a.statusMsg = ""
		if a.snapshotsModel != nil {
			initCmd := a.snapshotsModel.Init()
			if a.width > 0 && a.height > 0 {
				sizeMsg := tea.WindowSizeMsg{Width: a.width, Height: a.height - 1}
				a.snapshotsModel, _ = a.snapshotsModel.Update(sizeMsg)
			}
			return a, initCmd
		}
	case "settings":
		a.currentView = ViewSettings
		a.statusMsg = ""
		if a.settingsModel != nil {
			initCmd := a.settingsModel.Init()
			if a.width > 0 && a.height > 0 {
				sizeMsg := tea.WindowSizeMsg{Width: a.width, Height: a.height - 1}
				a.settingsModel, _ = a.settingsModel.Update(sizeMsg)
			}
			return a, initCmd
		}
	case "ticker":
		a.currentView = ViewTicker
		a.statusMsg = ""
		if a.tickerModel != nil {
			initCmd := a.tickerModel.Init()
			if a.width > 0 && a.height > 0 {
				sizeMsg := tea.WindowSizeMsg{Width: a.width, Height: a.height - 1}
				a.tickerModel, _ = a.tickerModel.Update(sizeMsg)
			}
			return a, initCmd
		}
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
	case ViewSettings:
		if a.settingsModel != nil {
			content = a.settingsModel.View()
		} else {
			content = "Loading settings..."
		}
	case ViewSnapshots:
		if a.snapshotsModel != nil {
			content = a.snapshotsModel.View()
		} else {
			content = "Loading snapshots..."
		}
	case ViewBuy:
		if a.buyModel != nil {
			content = a.buyModel.View()
		} else {
			content = "Loading purchases..."
		}
	case ViewSell:
		if a.sellModel != nil {
			content = a.sellModel.View()
		} else {
			content = "Loading sales..."
		}
	case ViewStake:
		if a.stakeModel != nil {
			content = a.stakeModel.View()
		} else {
			content = "Loading stakes..."
		}
	case ViewLoan:
		if a.loanModel != nil {
			content = a.loanModel.View()
		} else {
			content = "Loading loans..."
		}
	case ViewTicker:
		if a.tickerModel != nil {
			content = a.tickerModel.View()
		} else {
			content = "Loading ticker mappings..."
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

	// These views handle their own layout, don't wrap them
	if a.currentView == ViewSummary || a.currentView == ViewSettings || a.currentView == ViewSnapshots || a.currentView == ViewBuy || a.currentView == ViewSell || a.currentView == ViewStake || a.currentView == ViewLoan || a.currentView == ViewTicker {
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
	case ViewTicker:
		return "Ticker Mappings"
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

	switch a.currentView {
	case ViewSummary:
		right = "r Refresh | esc Back | q Quit"
	case ViewSettings:
		right = "↑↓ Navigate | Enter Toggle/Edit | esc Back | q Quit"
	case ViewSnapshots:
		right = "↑↓ Navigate | n New | d Delete | esc Back | q Quit"
	case ViewBuy:
		right = "↑↓ Navigate | a Add | d Delete | esc Back | q Quit"
	case ViewSell:
		right = "↑↓ Navigate | a Add | d Delete | esc Back | q Quit"
	case ViewStake:
		right = "↑↓ Navigate | a Add | d Unstake | esc Back | q Quit"
	case ViewLoan:
		right = "↑↓ Navigate | a Add | d Repay | esc Back | q Quit"
	case ViewTicker:
		right = "↑↓ Navigate | a Add | s Search | d Delete | v Defaults | esc Back"
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

// SetSettingsModel sets the settings model for the app.
func (a *App) SetSettingsModel(m tea.Model) {
	a.settingsModel = m
}

// SetSnapshotsModel sets the snapshots model for the app.
func (a *App) SetSnapshotsModel(m tea.Model) {
	a.snapshotsModel = m
}

// SetBuyModel sets the buy model for the app.
func (a *App) SetBuyModel(m tea.Model) {
	a.buyModel = m
}

// SetSellModel sets the sell model for the app.
func (a *App) SetSellModel(m tea.Model) {
	a.sellModel = m
}

// SetStakeModel sets the stake model for the app.
func (a *App) SetStakeModel(m tea.Model) {
	a.stakeModel = m
}

// SetLoanModel sets the loan model for the app.
func (a *App) SetLoanModel(m tea.Model) {
	a.loanModel = m
}

// SetTickerModel sets the ticker model for the app.
func (a *App) SetTickerModel(m tea.Model) {
	a.tickerModel = m
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
