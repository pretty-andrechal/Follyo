// Package views provides the different views/screens for the TUI.
package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
)

// MenuItem represents a single menu item.
type MenuItem struct {
	Title       string
	Description string
	Action      string // Identifier for what action to take
}

// MenuModel represents the main menu view.
type MenuModel struct {
	items    []MenuItem
	cursor   int
	keys     tui.KeyMap
	width    int
	height   int
	quitting bool
}

// NewMenuModel creates a new main menu model.
func NewMenuModel() MenuModel {
	items := []MenuItem{
		{Title: "Portfolio Summary", Description: "View your portfolio overview with live prices", Action: "summary"},
		{Title: "Buy", Description: "Manage purchases", Action: "buy"},
		{Title: "Sell", Description: "Manage sales", Action: "sell"},
		{Title: "Stake", Description: "Manage staking positions", Action: "stake"},
		{Title: "Loan", Description: "Manage loans", Action: "loan"},
		{Title: "Snapshots", Description: "Save and compare portfolio snapshots", Action: "snapshots"},
		{Title: "Coin History", Description: "Track price and holdings for a coin over time", Action: "coinhistory"},
		{Title: "Ticker Mappings", Description: "Map tickers to CoinGecko IDs for price fetching", Action: "ticker"},
		{Title: "Settings", Description: "Configure preferences", Action: "settings"},
	}

	return MenuModel{
		items:  items,
		cursor: 0,
		keys:   tui.DefaultKeyMap(),
	}
}

// Init initializes the menu model.
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the menu model.
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case key.Matches(msg, m.keys.Select):
			// Return selected action - will be handled by parent app model
			return m, func() tea.Msg {
				return tui.MenuSelectMsg{Action: m.items[m.cursor].Action}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the menu.
func (m MenuModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Logo/Header
	logo := `
  ███████╗ ██████╗ ██╗     ██╗  ██╗   ██╗ ██████╗
  ██╔════╝██╔═══██╗██║     ██║  ╚██╗ ██╔╝██╔═══██╗
  █████╗  ██║   ██║██║     ██║   ╚████╔╝ ██║   ██║
  ██╔══╝  ██║   ██║██║     ██║    ╚██╔╝  ██║   ██║
  ██║     ╚██████╔╝███████╗███████╗██║   ╚██████╔╝
  ╚═╝      ╚═════╝ ╚══════╝╚══════╝╚═╝    ╚═════╝ `

	logoStyle := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true)

	b.WriteString(logoStyle.Render(logo))
	b.WriteString("\n\n")

	// Subtitle
	subtitle := lipgloss.NewStyle().
		Foreground(tui.SubtleTextColor).
		Italic(true).
		Render("  Crypto Portfolio Tracker")
	b.WriteString(subtitle)
	b.WriteString("\n\n")

	// Menu items
	for i, item := range m.items {
		cursor := "  "
		style := tui.MenuItemStyle

		if i == m.cursor {
			cursor = "> "
			style = tui.SelectedMenuItemStyle
		}

		// Item line
		itemText := fmt.Sprintf("%s%s", cursor, item.Title)
		b.WriteString(style.Render(itemText))
		b.WriteString("\n")

		// Description (only for selected item)
		if i == m.cursor {
			descStyle := lipgloss.NewStyle().
				Foreground(tui.MutedColor).
				PaddingLeft(4)
			b.WriteString(descStyle.Render(item.Description))
			b.WriteString("\n")
		}
	}

	// Help
	b.WriteString("\n")
	helpStyle := lipgloss.NewStyle().
		Foreground(tui.MutedColor).
		PaddingLeft(2)
	help := fmt.Sprintf("%s navigate  %s select  %s quit",
		tui.HelpKeyStyle.Render("↑↓"),
		tui.HelpKeyStyle.Render("enter"),
		tui.HelpKeyStyle.Render("q"))
	b.WriteString(helpStyle.Render(help))

	// Wrap in a box
	content := b.String()
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.BorderColor).
		Padding(1, 2)

	return boxStyle.Render(content)
}

// SelectedItem returns the currently selected menu item.
func (m MenuModel) SelectedItem() MenuItem {
	return m.items[m.cursor]
}
