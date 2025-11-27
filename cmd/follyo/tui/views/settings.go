package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
	"github.com/pretty-andrechal/follyo/internal/config"
)

// SettingType represents the type of a setting
type SettingType int

const (
	SettingToggle SettingType = iota
	SettingText
)

// SettingItem represents a single setting
type SettingItem struct {
	Key         string
	Label       string
	Description string
	Type        SettingType
	BoolValue   bool
	TextValue   string
}

// SettingsModel represents the settings view
type SettingsModel struct {
	config      *config.ConfigStore
	items       []SettingItem
	cursor      int
	editing     bool
	textInput   textinput.Model
	keys        tui.KeyMap
	width       int
	height      int
	err         error
	statusMsg   string
}

// NewSettingsModel creates a new settings view model
func NewSettingsModel(cfg *config.ConfigStore) SettingsModel {
	ti := textinput.New()
	ti.Placeholder = "Enter platform name..."
	ti.CharLimit = 50

	m := SettingsModel{
		config:    cfg,
		keys:      tui.DefaultKeyMap(),
		textInput: ti,
	}

	m.loadSettings()
	return m
}

func (m *SettingsModel) loadSettings() {
	m.items = []SettingItem{
		{
			Key:         "prices",
			Label:       "Live Prices",
			Description: "Fetch live prices from CoinGecko by default",
			Type:        SettingToggle,
			BoolValue:   m.config.GetFetchPrices(),
		},
		{
			Key:         "colors",
			Label:       "Color Output",
			Description: "Use colored output in terminal",
			Type:        SettingToggle,
			BoolValue:   m.config.GetColorOutput(),
		},
		{
			Key:         "platform",
			Label:       "Default Platform",
			Description: "Default platform for new entries",
			Type:        SettingText,
			TextValue:   m.config.GetDefaultPlatform(),
		},
	}
}

// Init initializes the settings model
func (m SettingsModel) Init() tea.Cmd {
	return nil
}

// SettingsSavedMsg is sent when a setting is saved
type SettingsSavedMsg struct {
	Key   string
	Error error
}

// Update handles messages for the settings model
func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If editing text, handle text input
		if m.editing {
			return m.handleEditingKeys(msg)
		}

		switch {
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return tui.BackToMenuMsg{} }

		case key.Matches(msg, m.keys.Quit):
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
			return m.handleSelect()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case SettingsSavedMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.statusMsg = fmt.Sprintf("Error saving %s: %v", msg.Key, msg.Error)
		} else {
			m.statusMsg = fmt.Sprintf("Saved %s", msg.Key)
			m.loadSettings() // Reload to get updated values
		}
	}

	return m, nil
}

func (m SettingsModel) handleEditingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		// Cancel editing
		m.editing = false
		m.textInput.Blur()
		m.statusMsg = "Cancelled"
		return m, nil

	case tea.KeyEnter:
		// Save the value
		m.editing = false
		m.textInput.Blur()
		value := strings.TrimSpace(m.textInput.Value())
		return m, m.saveSetting(m.items[m.cursor].Key, value)

	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

func (m SettingsModel) handleSelect() (tea.Model, tea.Cmd) {
	item := m.items[m.cursor]

	switch item.Type {
	case SettingToggle:
		// Toggle the boolean value
		newValue := !item.BoolValue
		return m, m.saveToggle(item.Key, newValue)

	case SettingText:
		// Enter edit mode
		m.editing = true
		m.textInput.SetValue(item.TextValue)
		m.textInput.Focus()
		m.statusMsg = "Enter to save, Esc to cancel"
		return m, textinput.Blink
	}

	return m, nil
}

func (m SettingsModel) saveToggle(key string, value bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		switch key {
		case "prices":
			err = m.config.SetFetchPrices(value)
		case "colors":
			err = m.config.SetColorOutput(value)
		}
		return SettingsSavedMsg{Key: key, Error: err}
	}
}

func (m SettingsModel) saveSetting(key, value string) tea.Cmd {
	return func() tea.Msg {
		var err error
		switch key {
		case "platform":
			if value == "" {
				err = m.config.ClearDefaultPlatform()
			} else {
				err = m.config.SetDefaultPlatform(value)
			}
		}
		return SettingsSavedMsg{Key: key, Error: err}
	}
}

// View renders the settings view
func (m SettingsModel) View() string {
	var b strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Render("SETTINGS")

	b.WriteString(title)
	b.WriteString("\n\n")

	// Settings list
	for i, item := range m.items {
		b.WriteString(m.renderSettingItem(i, item))
		b.WriteString("\n")
	}

	// Status message
	if m.statusMsg != "" {
		b.WriteString("\n")
		statusStyle := lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Italic(true)
		b.WriteString(statusStyle.Render(m.statusMsg))
	}

	// Help
	b.WriteString("\n\n")
	help := m.renderHelp()
	b.WriteString(help)

	// Wrap in a box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.BorderColor).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

func (m SettingsModel) renderSettingItem(index int, item SettingItem) string {
	isSelected := index == m.cursor

	// Cursor
	cursor := "  "
	if isSelected {
		cursor = lipgloss.NewStyle().
			Foreground(tui.PrimaryColor).
			Render("> ")
	}

	// Label
	labelStyle := lipgloss.NewStyle().
		Width(20).
		Foreground(tui.TextColor)
	if isSelected {
		labelStyle = labelStyle.Bold(true).Foreground(tui.PrimaryColor)
	}
	label := labelStyle.Render(item.Label)

	// Value
	var value string
	switch item.Type {
	case SettingToggle:
		if item.BoolValue {
			value = lipgloss.NewStyle().
				Foreground(tui.SuccessColor).
				Bold(true).
				Render("[ON]")
		} else {
			value = lipgloss.NewStyle().
				Foreground(tui.ErrorColor).
				Render("[OFF]")
		}

	case SettingText:
		if m.editing && isSelected {
			// Show text input
			value = m.textInput.View()
		} else if item.TextValue == "" {
			value = lipgloss.NewStyle().
				Foreground(tui.MutedColor).
				Italic(true).
				Render("(not set)")
		} else {
			value = lipgloss.NewStyle().
				Foreground(tui.AccentColor).
				Render(item.TextValue)
		}
	}

	// Description (shown for selected item)
	var desc string
	if isSelected && !m.editing {
		desc = "\n     " + lipgloss.NewStyle().
			Foreground(tui.SubtleTextColor).
			Italic(true).
			Render(item.Description)
	}

	return cursor + label + value + desc
}

func (m SettingsModel) renderHelp() string {
	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)

	var help string
	if m.editing {
		help = fmt.Sprintf("%s save  %s cancel",
			tui.HelpKeyStyle.Render("enter"),
			tui.HelpKeyStyle.Render("esc"))
	} else {
		help = fmt.Sprintf("%s navigate  %s toggle/edit  %s back  %s quit",
			tui.HelpKeyStyle.Render("↑↓"),
			tui.HelpKeyStyle.Render("enter"),
			tui.HelpKeyStyle.Render("esc"),
			tui.HelpKeyStyle.Render("q"))
	}

	return helpStyle.Render(help)
}

// GetConfig returns the config store
func (m SettingsModel) GetConfig() *config.ConfigStore {
	return m.config
}
