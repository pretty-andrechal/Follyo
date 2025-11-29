package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui/format"
	"github.com/pretty-andrechal/follyo/internal/models"
	"github.com/pretty-andrechal/follyo/internal/portfolio"
)

// Form field indices for stake form
const (
	stakeFieldCoin = iota
	stakeFieldAmount
	stakeFieldPlatform
	stakeFieldAPY
	stakeFieldNotes
)

// StakeModel represents the stake view
type StakeModel struct {
	portfolio       portfolio.StakesManager
	defaultPlatform string
	stakes          []models.Stake
	state           EntityViewState
	config          EntityViewConfig
}

// NewStakeModel creates a new stake view model
func NewStakeModel(p portfolio.StakesManager, defaultPlatform string) StakeModel {
	fields := []FormFieldConfig{
		{Label: "Coin:", Placeholder: "ETH, SOL, ATOM...", CharLimit: tui.InputCoinCharLimit, Width: tui.InputCoinWidth},
		{Label: "Amount:", Placeholder: "10.0", CharLimit: tui.InputAmountCharLimit, Width: tui.InputAmountWidth},
		{Label: "Platform:", Placeholder: "Lido, Rocket Pool...", CharLimit: tui.InputPlatformCharLimit, Width: tui.InputPlatformWidth, DefaultValue: defaultPlatform},
		{Label: "APY (%):", Placeholder: "4.5 (optional)", CharLimit: tui.InputRateCharLimit, Width: tui.InputRateWidth},
		{Label: "Notes:", Placeholder: "Optional notes...", CharLimit: tui.InputNotesCharLimit, Width: tui.InputNotesWidth},
	}

	m := StakeModel{
		portfolio:       p,
		defaultPlatform: defaultPlatform,
		state:           NewEntityViewState(fields),
		config: EntityViewConfig{
			Title:          "STAKING",
			EntityName:     "stake",
			EmptyMessage:   "No staked positions yet. Press 'a' to add one.",
			ColumnHeader:   fmt.Sprintf("  %-8s  %14s  %-20s  %8s  %s", "Coin", "Amount", "Platform", "APY", "Date"),
			SeparatorWidth: tui.SeparatorWidthStake,
			FormTitle:      "ADD STAKE",
			FormLabels:     []string{"Coin:", "Amount:", "Platform:", "APY (%):", "Notes:"},
			RenderRow:      nil, // Set below
			RenderDeleteInfo: func(item interface{}) string {
				s := item.(models.Stake)
				return fmt.Sprintf("Unstake %.6f %s from %s?", s.Amount, s.Coin, s.Platform)
			},
		},
	}

	m.config.RenderRow = func(index, cursor int, item interface{}) string {
		return renderStakeRowWithCursor(index, cursor, item.(models.Stake))
	}

	m.loadStakes()
	return m
}

func (m *StakeModel) loadStakes() {
	stakes, err := m.portfolio.ListStakes()
	if err != nil {
		m.state.Err = err
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
		switch m.state.Mode {
		case EntityModeAdd:
			return m.handleAddKeys(msg)
		case EntityModeConfirmDelete:
			return m.handleDeleteConfirmKeys(msg)
		default:
			return m.handleListKeys(msg)
		}

	case tea.WindowSizeMsg:
		m.state.HandleWindowSize(msg.Width, msg.Height)

	case StakeAddedMsg:
		if msg.Error != nil {
			m.state.Err = msg.Error
			m.state.SetStatusMsg(fmt.Sprintf("Error: %v", msg.Error))
		} else {
			m.state.SetStatusMsg(fmt.Sprintf("Staked %s on %s!", msg.Stake.Coin, msg.Stake.Platform))
			m.loadStakes()
		}
		m.state.Mode = EntityModeList

	case StakeDeletedMsg:
		if msg.Error != nil {
			m.state.Err = msg.Error
			m.state.SetStatusMsg(fmt.Sprintf("Error: %v", msg.Error))
		} else {
			m.state.SetStatusMsg("Stake removed")
			m.loadStakes()
			m.state.AdjustCursorAfterDelete(len(m.stakes))
		}
	}

	return m, nil
}

func (m StakeModel) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.state.Keys.Back):
		return m, func() tea.Msg { return tui.BackToMenuMsg{} }

	case key.Matches(msg, m.state.Keys.Quit):
		return m, tea.Quit

	case m.state.HandleListNavigation(msg, len(m.stakes)):
		// Navigation handled

	case msg.String() == "a" || msg.String() == "n":
		defaults := []string{"", "", m.defaultPlatform, "", ""}
		return m, m.state.EnterAddMode(defaults)

	case msg.String() == "d" || msg.String() == "x":
		if len(m.stakes) > 0 {
			m.state.EnterDeleteMode()
		}
	}

	return m, nil
}

func (m StakeModel) handleAddKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if handled, cmd := m.state.HandleFormNavigation(msg); handled {
		return m, cmd
	}

	switch msg.Type {
	case tea.KeyEnter:
		if m.state.IsLastField() || msg.Alt {
			return m.submitForm()
		}
		return m, m.state.MoveToNextField()

	default:
		return m, m.state.UpdateFocusedInput(msg)
	}
}

func (m StakeModel) handleDeleteConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if len(m.stakes) > 0 {
			id := m.stakes[m.state.Cursor].ID
			m.state.Mode = EntityModeList
			return m, m.deleteStake(id)
		}
	case "n", "N", "escape":
		m.state.ExitToList()
	}
	return m, nil
}

func (m StakeModel) submitForm() (tea.Model, tea.Cmd) {
	coin := strings.ToUpper(m.state.GetFieldValue(stakeFieldCoin))
	if coin == "" {
		m.state.SetStatusMsg("Coin is required")
		return m, nil
	}

	amount, err := strconv.ParseFloat(m.state.GetFieldValue(stakeFieldAmount), 64)
	if err != nil || amount <= 0 {
		m.state.SetStatusMsg("Invalid amount")
		return m, nil
	}

	platform := m.state.GetFieldValue(stakeFieldPlatform)
	if platform == "" {
		m.state.SetStatusMsg("Platform is required for staking")
		return m, nil
	}

	// APY is optional
	var apy *float64
	apyStr := m.state.GetFieldValue(stakeFieldAPY)
	if apyStr != "" {
		apyVal, err := strconv.ParseFloat(apyStr, 64)
		if err != nil || apyVal < 0 {
			m.state.SetStatusMsg("Invalid APY")
			return m, nil
		}
		apy = &apyVal
	}

	notes := m.state.GetFieldValue(stakeFieldNotes)

	m.state.ExitToList()
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
	switch m.state.Mode {
	case EntityModeAdd:
		return RenderAddForm(m.config, &m.state)
	case EntityModeConfirmDelete:
		if m.state.Cursor < len(m.stakes) {
			return RenderDeleteConfirm(m.config, m.stakes[m.state.Cursor])
		}
		return RenderDeleteConfirm(m.config, nil)
	default:
		return m.renderList()
	}
}

func (m StakeModel) renderList() string {
	items := make([]interface{}, len(m.stakes))
	for i, s := range m.stakes {
		items[i] = s
	}
	return RenderListView(m.config, &m.state, items)
}

// renderStakeRowWithCursor renders a single stake row with cursor highlighting
func renderStakeRowWithCursor(index, cursor int, s models.Stake) string {
	isSelected := index == cursor

	cursorStr := "  "
	if isSelected {
		cursorStr = lipgloss.NewStyle().
			Foreground(tui.PrimaryColor).
			Render("> ")
	}

	date := format.TruncateDate(s.Date)
	platform := format.TruncatePlatformLong(s.Platform)

	apyStr := "-"
	if s.APY != nil {
		apyStr = fmt.Sprintf("%.2f%%", *s.APY)
	}

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

	return cursorStr + rowStyle.Render(row) + "\n"
}

// GetPortfolio returns the portfolio instance
func (m StakeModel) GetPortfolio() portfolio.StakesManager {
	return m.portfolio
}
