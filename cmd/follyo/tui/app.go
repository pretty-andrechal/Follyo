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
	portfolio   *portfolio.Portfolio
	storage     *storage.Storage
	currentView ViewType
	menuModel   tea.Model
	width       int
	height      int
	err         error
	quitting    bool
	statusMsg   string
}

// NewApp creates a new application instance.
func NewApp(s *storage.Storage, p *portfolio.Portfolio) *App {
	return &App{
		storage:     s,
		portfolio:   p,
		currentView: ViewMenu,
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
		default:
			// For unimplemented views, go back to menu on any key
			a.currentView = ViewMenu
			return a, nil
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Forward to current view
		if a.menuModel != nil {
			var cmd tea.Cmd
			a.menuModel, cmd = a.menuModel.Update(msg)
			return a, cmd
		}

	case MenuSelectMsg:
		return a.handleMenuSelect(msg)

	case errMsg:
		a.err = msg.err
		return a, nil

	case statusMsg:
		a.statusMsg = string(msg)
		return a, nil
	}

	return a, nil
}

// MenuSelectMsg is sent when a menu item is selected.
type MenuSelectMsg struct {
	Action string
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

func (a *App) handleMenuSelect(msg MenuSelectMsg) (tea.Model, tea.Cmd) {
	switch msg.Action {
	case "summary":
		a.currentView = ViewSummary
		a.statusMsg = "Summary view - Press ESC to go back (coming soon)"
	case "buy":
		a.currentView = ViewBuy
		a.statusMsg = "Buy view - Press ESC to go back (coming soon)"
	case "sell":
		a.currentView = ViewSell
		a.statusMsg = "Sell view - Press ESC to go back (coming soon)"
	case "stake":
		a.currentView = ViewStake
		a.statusMsg = "Stake view - Press ESC to go back (coming soon)"
	case "loan":
		a.currentView = ViewLoan
		a.statusMsg = "Loan view - Press ESC to go back (coming soon)"
	case "snapshots":
		a.currentView = ViewSnapshots
		a.statusMsg = "Snapshots view - Press ESC to go back (coming soon)"
	case "settings":
		a.currentView = ViewSettings
		a.statusMsg = "Settings view - Press ESC to go back (coming soon)"
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
	default:
		// Placeholder for unimplemented views
		content = a.renderPlaceholder()
	}

	// Add status bar at the bottom
	statusBar := a.renderStatusBar()

	// Calculate available height for content
	statusHeight := lipgloss.Height(statusBar)
	contentHeight := a.height - statusHeight

	// Style content area
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
