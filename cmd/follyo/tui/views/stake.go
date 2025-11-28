package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui/format"
	"github.com/pretty-andrechal/follyo/internal/models"
	"github.com/pretty-andrechal/follyo/internal/portfolio"
)

// StakeViewMode represents the current mode of the stake view
type StakeViewMode int

const (
	StakeList StakeViewMode = iota
	StakeAdd
	StakeConfirmDelete
)

// Stake form field indices
const (
	stakeFieldCoin = iota
	stakeFieldAmount
	stakeFieldPlatform
	stakeFieldAPY
	stakeFieldNotes
	stakeFieldCount
)

// StakeModel represents the stake view
type StakeModel struct {
	portfolio       *portfolio.Portfolio
	defaultPlatform string
	stakes          []models.Stake
	cursor          int
	mode            StakeViewMode
	inputs          []textinput.Model
	focusIndex      int
	keys            tui.KeyMap
	width           int
	height          int
	err             error
	statusMsg       string
}

// NewStakeModel creates a new stake view model
func NewStakeModel(p *portfolio.Portfolio, defaultPlatform string) StakeModel {
	inputs := make([]textinput.Model, stakeFieldCount)

	// Coin input
	inputs[stakeFieldCoin] = textinput.New()
	inputs[stakeFieldCoin].Placeholder = "ETH, SOL, ATOM..."
	inputs[stakeFieldCoin].CharLimit = tui.InputCoinCharLimit
	inputs[stakeFieldCoin].Width = tui.InputCoinWidth

	// Amount input
	inputs[stakeFieldAmount] = textinput.New()
	inputs[stakeFieldAmount].Placeholder = "10.0"
	inputs[stakeFieldAmount].CharLimit = tui.InputAmountCharLimit
	inputs[stakeFieldAmount].Width = tui.InputAmountWidth

	// Platform input (required for staking)
	inputs[stakeFieldPlatform] = textinput.New()
	inputs[stakeFieldPlatform].Placeholder = "Lido, Rocket Pool..."
	inputs[stakeFieldPlatform].CharLimit = tui.InputPlatformCharLimit
	inputs[stakeFieldPlatform].Width = tui.InputPlatformWidth
	if defaultPlatform != "" {
		inputs[stakeFieldPlatform].SetValue(defaultPlatform)
	}

	// APY input (optional)
	inputs[stakeFieldAPY] = textinput.New()
	inputs[stakeFieldAPY].Placeholder = "4.5 (optional)"
	inputs[stakeFieldAPY].CharLimit = tui.InputRateCharLimit
	inputs[stakeFieldAPY].Width = tui.InputRateWidth

	// Notes input
	inputs[stakeFieldNotes] = textinput.New()
	inputs[stakeFieldNotes].Placeholder = "Optional notes..."
	inputs[stakeFieldNotes].CharLimit = tui.InputNotesCharLimit
	inputs[stakeFieldNotes].Width = tui.InputNotesWidth

	m := StakeModel{
		portfolio:       p,
		defaultPlatform: defaultPlatform,
		inputs:          inputs,
		keys:            tui.DefaultKeyMap(),
		mode:            StakeList,
	}

	m.loadStakes()
	return m
}

func (m *StakeModel) loadStakes() {
	stakes, err := m.portfolio.ListStakes()
	if err != nil {
		m.err = err
		return
	}
	m.stakes = stakes
}

// Init initializes the stake model
func (m StakeModel) Init() tea.Cmd {
	return nil
}

// StakeAddedMsg is sent when a stake is added
type StakeAddedMsg struct {
	Stake *models.Stake
	Error error
}

// StakeDeletedMsg is sent when a stake is deleted
type StakeDeletedMsg struct {
	ID    string
	Error error
}

// Update handles messages for the stake model
func (m StakeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case StakeAdd:
			return m.handleAddKeys(msg)
		case StakeConfirmDelete:
			return m.handleDeleteConfirmKeys(msg)
		default:
			return m.handleListKeys(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case StakeAddedMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.statusMsg = fmt.Sprintf("Error: %v", msg.Error)
		} else {
			m.statusMsg = fmt.Sprintf("Staked %s on %s!", msg.Stake.Coin, msg.Stake.Platform)
			m.loadStakes()
		}
		m.mode = StakeList

	case StakeDeletedMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.statusMsg = fmt.Sprintf("Error: %v", msg.Error)
		} else {
			m.statusMsg = "Stake removed"
			m.loadStakes()
			if m.cursor >= len(m.stakes) && m.cursor > 0 {
				m.cursor--
			}
		}
	}

	return m, nil
}

func (m StakeModel) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		if m.cursor < len(m.stakes)-1 {
			m.cursor++
		}

	case msg.String() == "a" || msg.String() == "n":
		// Add new stake
		m.mode = StakeAdd
		m.focusIndex = 0
		m.resetForm()
		m.inputs[stakeFieldCoin].Focus()
		m.statusMsg = ""
		return m, textinput.Blink

	case msg.String() == "d" || msg.String() == "x":
		// Delete stake
		if len(m.stakes) > 0 {
			m.mode = StakeConfirmDelete
			m.statusMsg = ""
		}
	}

	return m, nil
}

func (m StakeModel) handleAddKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.mode = StakeList
		m.inputs[m.focusIndex].Blur()
		m.statusMsg = ""
		return m, nil

	case tea.KeyTab, tea.KeyShiftTab, tea.KeyDown, tea.KeyUp:
		// Navigate between fields
		if msg.Type == tea.KeyUp || msg.Type == tea.KeyShiftTab {
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = stakeFieldCount - 1
			}
		} else {
			m.focusIndex++
			if m.focusIndex >= stakeFieldCount {
				m.focusIndex = 0
			}
		}

		// Update focus
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focusIndex].Focus()
		return m, textinput.Blink

	case tea.KeyEnter:
		// If on last field or explicitly submitting, try to save
		if m.focusIndex == stakeFieldCount-1 || msg.Alt {
			return m.submitForm()
		}
		// Otherwise move to next field
		m.focusIndex++
		if m.focusIndex >= stakeFieldCount {
			m.focusIndex = 0
		}
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focusIndex].Focus()
		return m, textinput.Blink

	default:
		// Update the focused input
		var cmd tea.Cmd
		m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
		return m, cmd
	}
}

func (m StakeModel) handleDeleteConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm delete
		if len(m.stakes) > 0 {
			id := m.stakes[m.cursor].ID
			m.mode = StakeList
			return m, m.deleteStake(id)
		}
	case "n", "N", "escape":
		m.mode = StakeList
		m.statusMsg = ""
	}
	return m, nil
}

func (m *StakeModel) resetForm() {
	m.inputs[stakeFieldCoin].SetValue("")
	m.inputs[stakeFieldAmount].SetValue("")
	m.inputs[stakeFieldPlatform].SetValue(m.defaultPlatform)
	m.inputs[stakeFieldAPY].SetValue("")
	m.inputs[stakeFieldNotes].SetValue("")
}

func (m StakeModel) submitForm() (tea.Model, tea.Cmd) {
	// Validate inputs
	coin := strings.ToUpper(strings.TrimSpace(m.inputs[stakeFieldCoin].Value()))
	if coin == "" {
		m.statusMsg = "Coin is required"
		return m, nil
	}

	amountStr := strings.TrimSpace(m.inputs[stakeFieldAmount].Value())
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		m.statusMsg = "Invalid amount"
		return m, nil
	}

	platform := strings.TrimSpace(m.inputs[stakeFieldPlatform].Value())
	if platform == "" {
		m.statusMsg = "Platform is required for staking"
		return m, nil
	}

	// APY is optional
	var apy *float64
	apyStr := strings.TrimSpace(m.inputs[stakeFieldAPY].Value())
	if apyStr != "" {
		apyVal, err := strconv.ParseFloat(apyStr, 64)
		if err != nil || apyVal < 0 {
			m.statusMsg = "Invalid APY"
			return m, nil
		}
		apy = &apyVal
	}

	notes := strings.TrimSpace(m.inputs[stakeFieldNotes].Value())

	// Blur all inputs
	for i := range m.inputs {
		m.inputs[i].Blur()
	}

	return m, m.addStake(coin, amount, platform, apy, notes)
}

func (m StakeModel) addStake(coin string, amount float64, platform string, apy *float64, notes string) tea.Cmd {
	return func() tea.Msg {
		stake, err := m.portfolio.AddStake(coin, amount, platform, apy, notes, "")
		if err != nil {
			return StakeAddedMsg{Error: err}
		}
		return StakeAddedMsg{Stake: &stake}
	}
}

func (m StakeModel) deleteStake(id string) tea.Cmd {
	return func() tea.Msg {
		removed, err := m.portfolio.RemoveStake(id)
		if err != nil {
			return StakeDeletedMsg{ID: id, Error: err}
		}
		if !removed {
			return StakeDeletedMsg{ID: id, Error: fmt.Errorf("stake not found")}
		}
		return StakeDeletedMsg{ID: id}
	}
}

// View renders the stake view
func (m StakeModel) View() string {
	switch m.mode {
	case StakeAdd:
		return m.renderAddForm()
	case StakeConfirmDelete:
		return m.renderDeleteConfirm()
	default:
		return m.renderList()
	}
}

func (m StakeModel) renderList() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Render("STAKING")

	b.WriteString(title)
	b.WriteString("\n\n")

	if len(m.stakes) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Italic(true)
		b.WriteString(emptyStyle.Render("  No staked positions yet. Press 'a' to add one."))
		b.WriteString("\n")
	} else {
		// Column headers
		headerStyle := lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Bold(true)
		header := fmt.Sprintf("  %-8s  %14s  %-20s  %8s  %s",
			"Coin", "Amount", "Platform", "APY", "Date")
		b.WriteString(headerStyle.Render(header))
		b.WriteString("\n")

		// Separator
		sepStyle := lipgloss.NewStyle().Foreground(tui.BorderColor)
		b.WriteString(sepStyle.Render(strings.Repeat("─", tui.SeparatorWidthStake)))
		b.WriteString("\n")

		// List items
		for i, s := range m.stakes {
			b.WriteString(m.renderStakeRow(i, s))
		}
	}

	// Status message
	if m.statusMsg != "" {
		b.WriteString("\n")
		statusStyle := lipgloss.NewStyle().
			Foreground(tui.AccentColor).
			Italic(true)
		b.WriteString(statusStyle.Render(m.statusMsg))
	}

	// Help
	b.WriteString("\n\n")
	help := m.renderListHelp()
	b.WriteString(help)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.BorderColor).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

func (m StakeModel) renderStakeRow(index int, s models.Stake) string {
	isSelected := index == m.cursor

	// Cursor
	cursor := "  "
	if isSelected {
		cursor = lipgloss.NewStyle().
			Foreground(tui.PrimaryColor).
			Render("> ")
	}

	// Format values
	date := format.TruncateDate(s.Date)
	platform := format.TruncatePlatformLong(s.Platform)

	apyStr := "-"
	if s.APY != nil {
		apyStr = fmt.Sprintf("%.2f%%", *s.APY)
	}

	// Build row
	rowStyle := lipgloss.NewStyle().Foreground(tui.TextColor)
	if isSelected {
		rowStyle = rowStyle.Bold(true).Foreground(tui.PrimaryColor)
	}

	row := fmt.Sprintf("%-8s  %14.6f  %-20s  %8s  %s",
		s.Coin,
		s.Amount,
		platform,
		apyStr,
		date)

	return cursor + rowStyle.Render(row) + "\n"
}

func (m StakeModel) renderListHelp() string {
	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)

	var help string
	if len(m.stakes) > 0 {
		help = fmt.Sprintf("%s navigate  %s add  %s unstake  %s back  %s quit",
			tui.HelpKeyStyle.Render("↑↓"),
			tui.HelpKeyStyle.Render("a"),
			tui.HelpKeyStyle.Render("d"),
			tui.HelpKeyStyle.Render("esc"),
			tui.HelpKeyStyle.Render("q"))
	} else {
		help = fmt.Sprintf("%s add stake  %s back  %s quit",
			tui.HelpKeyStyle.Render("a"),
			tui.HelpKeyStyle.Render("esc"),
			tui.HelpKeyStyle.Render("q"))
	}

	return helpStyle.Render(help)
}

func (m StakeModel) renderAddForm() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Render("ADD STAKE")

	b.WriteString(title)
	b.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().
		Foreground(tui.SubtleTextColor).
		Width(12)

	focusedLabelStyle := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Width(12)

	fields := []struct {
		label string
		index int
	}{
		{"Coin:", stakeFieldCoin},
		{"Amount:", stakeFieldAmount},
		{"Platform:", stakeFieldPlatform},
		{"APY (%):", stakeFieldAPY},
		{"Notes:", stakeFieldNotes},
	}

	for _, f := range fields {
		ls := labelStyle
		if m.focusIndex == f.index {
			ls = focusedLabelStyle
		}
		b.WriteString(ls.Render(f.label))
		b.WriteString(m.inputs[f.index].View())
		b.WriteString("\n")
	}

	// Status message (for errors)
	if m.statusMsg != "" {
		b.WriteString("\n")
		errorStyle := lipgloss.NewStyle().
			Foreground(tui.ErrorColor).
			Italic(true)
		b.WriteString(errorStyle.Render(m.statusMsg))
	}

	// Help
	b.WriteString("\n\n")
	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)
	help := fmt.Sprintf("%s navigate  %s save  %s cancel",
		tui.HelpKeyStyle.Render("tab/↑↓"),
		tui.HelpKeyStyle.Render("enter"),
		tui.HelpKeyStyle.Render("esc"))
	b.WriteString(helpStyle.Render(help))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.BorderColor).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

func (m StakeModel) renderDeleteConfirm() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(tui.ErrorColor).
		Bold(true).
		Padding(0, 2).
		Render("CONFIRM UNSTAKE")

	b.WriteString(title)
	b.WriteString("\n\n")

	if m.cursor < len(m.stakes) {
		s := m.stakes[m.cursor]

		infoStyle := lipgloss.NewStyle().Foreground(tui.TextColor)
		info := fmt.Sprintf("Unstake %.6f %s from %s?",
			s.Amount, s.Coin, s.Platform)
		b.WriteString(infoStyle.Render(info))
		b.WriteString("\n\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)
	help := fmt.Sprintf("%s confirm  %s cancel",
		tui.HelpKeyStyle.Render("y"),
		tui.HelpKeyStyle.Render("n/esc"))
	b.WriteString(helpStyle.Render(help))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.ErrorColor).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

// GetPortfolio returns the portfolio instance
func (m StakeModel) GetPortfolio() *portfolio.Portfolio {
	return m.portfolio
}
