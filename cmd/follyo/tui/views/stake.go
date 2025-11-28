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
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui/components"
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
	portfolio       portfolio.StakesManager
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
func NewStakeModel(p portfolio.StakesManager, defaultPlatform string) StakeModel {
	fields := []components.FormField{
		{Placeholder: "ETH, SOL, ATOM...", CharLimit: tui.InputCoinCharLimit, Width: tui.InputCoinWidth},
		{Placeholder: "10.0", CharLimit: tui.InputAmountCharLimit, Width: tui.InputAmountWidth},
		{Placeholder: "Lido, Rocket Pool...", CharLimit: tui.InputPlatformCharLimit, Width: tui.InputPlatformWidth, DefaultValue: defaultPlatform},
		{Placeholder: "4.5 (optional)", CharLimit: tui.InputRateCharLimit, Width: tui.InputRateWidth},
		{Placeholder: "Optional notes...", CharLimit: tui.InputNotesCharLimit, Width: tui.InputNotesWidth},
	}

	m := StakeModel{
		portfolio:       p,
		defaultPlatform: defaultPlatform,
		inputs:          components.BuildFormInputs(fields),
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
		m.cursor = components.MoveCursorUp(m.cursor, len(m.stakes))

	case key.Matches(msg, m.keys.Down):
		m.cursor = components.MoveCursorDown(m.cursor, len(m.stakes))

	case msg.String() == "a" || msg.String() == "n":
		// Add new stake
		m.mode = StakeAdd
		m.focusIndex = 0
		m.resetForm()
		return m, components.FocusField(m.inputs, stakeFieldCoin)

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
		components.BlurAll(m.inputs)
		m.statusMsg = ""
		return m, nil

	case tea.KeyTab, tea.KeyShiftTab, tea.KeyDown, tea.KeyUp:
		// Navigate between fields
		var cmd tea.Cmd
		if msg.Type == tea.KeyUp || msg.Type == tea.KeyShiftTab {
			m.focusIndex, cmd = components.PrevField(m.inputs, m.focusIndex)
		} else {
			m.focusIndex, cmd = components.NextField(m.inputs, m.focusIndex)
		}
		return m, cmd

	case tea.KeyEnter:
		// If on last field or explicitly submitting, try to save
		if m.focusIndex == stakeFieldCount-1 || msg.Alt {
			return m.submitForm()
		}
		// Otherwise move to next field
		var cmd tea.Cmd
		m.focusIndex, cmd = components.NextField(m.inputs, m.focusIndex)
		return m, cmd

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
	defaults := []string{"", "", m.defaultPlatform, "", ""}
	components.ResetFormInputs(m.inputs, defaults)
	m.statusMsg = ""
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

	components.BlurAll(m.inputs)
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

	b.WriteString(components.RenderTitle("STAKING"))
	b.WriteString("\n\n")

	if len(m.stakes) == 0 {
		b.WriteString(components.RenderEmptyState("No staked positions yet. Press 'a' to add one."))
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
		b.WriteString(components.RenderSeparator(tui.SeparatorWidthStake))
		b.WriteString("\n")

		// List items
		for i, s := range m.stakes {
			b.WriteString(m.renderStakeRow(i, s))
		}
	}

	// Status message
	if m.statusMsg != "" {
		b.WriteString("\n")
		b.WriteString(components.RenderStatusMessage(m.statusMsg, false))
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(components.RenderHelp(components.ListHelp(len(m.stakes) > 0)))

	return components.RenderBoxDefault(b.String())
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

func (m StakeModel) renderAddForm() string {
	var b strings.Builder

	b.WriteString(components.RenderTitle("ADD STAKE"))
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
		b.WriteString(components.RenderStatusMessage(m.statusMsg, true))
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(components.RenderHelp(components.FormHelp()))

	return components.RenderBoxDefault(b.String())
}

func (m StakeModel) renderDeleteConfirm() string {
	var b strings.Builder

	b.WriteString(components.RenderErrorTitle("CONFIRM UNSTAKE"))
	b.WriteString("\n\n")

	if m.cursor < len(m.stakes) {
		s := m.stakes[m.cursor]

		infoStyle := lipgloss.NewStyle().Foreground(tui.TextColor)
		info := fmt.Sprintf("Unstake %.6f %s from %s?",
			s.Amount, s.Coin, s.Platform)
		b.WriteString(infoStyle.Render(info))
		b.WriteString("\n\n")
	}

	b.WriteString(components.RenderHelp(components.DeleteConfirmHelp()))

	return components.RenderBoxError(b.String())
}

// GetPortfolio returns the portfolio instance
func (m StakeModel) GetPortfolio() portfolio.StakesManager {
	return m.portfolio
}
