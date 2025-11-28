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

// LoanViewMode represents the current mode of the loan view
type LoanViewMode int

const (
	LoanList LoanViewMode = iota
	LoanAdd
	LoanConfirmDelete
)

// Loan form field indices
const (
	loanFieldCoin = iota
	loanFieldAmount
	loanFieldPlatform
	loanFieldInterestRate
	loanFieldNotes
	loanFieldCount
)

// LoanModel represents the loan view
type LoanModel struct {
	portfolio       portfolio.LoansManager
	defaultPlatform string
	loans           []models.Loan
	cursor          int
	mode            LoanViewMode
	inputs          []textinput.Model
	focusIndex      int
	keys            tui.KeyMap
	width           int
	height          int
	err             error
	statusMsg       string
}

// NewLoanModel creates a new loan view model
func NewLoanModel(p portfolio.LoansManager, defaultPlatform string) LoanModel {
	fields := []components.FormField{
		{Placeholder: "USDT, USDC, BTC...", CharLimit: tui.InputCoinCharLimit, Width: tui.InputCoinWidth},
		{Placeholder: "5000.0", CharLimit: tui.InputAmountCharLimit, Width: tui.InputAmountWidth},
		{Placeholder: "Nexo, Celsius...", CharLimit: tui.InputPlatformCharLimit, Width: tui.InputPlatformWidth, DefaultValue: defaultPlatform},
		{Placeholder: "6.9 (optional)", CharLimit: tui.InputRateCharLimit, Width: tui.InputRateWidth},
		{Placeholder: "Optional notes...", CharLimit: tui.InputNotesCharLimit, Width: tui.InputNotesWidth},
	}

	m := LoanModel{
		portfolio:       p,
		defaultPlatform: defaultPlatform,
		inputs:          components.BuildFormInputs(fields),
		keys:            tui.DefaultKeyMap(),
		mode:            LoanList,
	}

	m.loadLoans()
	return m
}

func (m *LoanModel) loadLoans() {
	loans, err := m.portfolio.ListLoans()
	if err != nil {
		m.err = err
		return
	}
	m.loans = loans
}

// Init initializes the loan model
func (m LoanModel) Init() tea.Cmd {
	return nil
}

// LoanAddedMsg is sent when a loan is added
type LoanAddedMsg struct {
	Loan  *models.Loan
	Error error
}

// LoanDeletedMsg is sent when a loan is deleted
type LoanDeletedMsg struct {
	ID    string
	Error error
}

// Update handles messages for the loan model
func (m LoanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case LoanAdd:
			return m.handleAddKeys(msg)
		case LoanConfirmDelete:
			return m.handleDeleteConfirmKeys(msg)
		default:
			return m.handleListKeys(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case LoanAddedMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.statusMsg = fmt.Sprintf("Error: %v", msg.Error)
		} else {
			m.statusMsg = fmt.Sprintf("Added %s loan on %s!", msg.Loan.Coin, msg.Loan.Platform)
			m.loadLoans()
		}
		m.mode = LoanList

	case LoanDeletedMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.statusMsg = fmt.Sprintf("Error: %v", msg.Error)
		} else {
			m.statusMsg = "Loan removed"
			m.loadLoans()
			if m.cursor >= len(m.loans) && m.cursor > 0 {
				m.cursor--
			}
		}
	}

	return m, nil
}

func (m LoanModel) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		return m, func() tea.Msg { return tui.BackToMenuMsg{} }

	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		m.cursor = components.MoveCursorUp(m.cursor, len(m.loans))

	case key.Matches(msg, m.keys.Down):
		m.cursor = components.MoveCursorDown(m.cursor, len(m.loans))

	case msg.String() == "a" || msg.String() == "n":
		// Add new loan
		m.mode = LoanAdd
		m.focusIndex = 0
		m.resetForm()
		return m, components.FocusField(m.inputs, loanFieldCoin)

	case msg.String() == "d" || msg.String() == "x":
		// Delete loan
		if len(m.loans) > 0 {
			m.mode = LoanConfirmDelete
			m.statusMsg = ""
		}
	}

	return m, nil
}

func (m LoanModel) handleAddKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.mode = LoanList
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
		if m.focusIndex == loanFieldCount-1 || msg.Alt {
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

func (m LoanModel) handleDeleteConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm delete
		if len(m.loans) > 0 {
			id := m.loans[m.cursor].ID
			m.mode = LoanList
			return m, m.deleteLoan(id)
		}
	case "n", "N", "escape":
		m.mode = LoanList
		m.statusMsg = ""
	}
	return m, nil
}

func (m *LoanModel) resetForm() {
	defaults := []string{"", "", m.defaultPlatform, "", ""}
	components.ResetFormInputs(m.inputs, defaults)
	m.statusMsg = ""
}

func (m LoanModel) submitForm() (tea.Model, tea.Cmd) {
	// Validate inputs
	coin := strings.ToUpper(strings.TrimSpace(m.inputs[loanFieldCoin].Value()))
	if coin == "" {
		m.statusMsg = "Coin is required"
		return m, nil
	}

	amountStr := strings.TrimSpace(m.inputs[loanFieldAmount].Value())
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		m.statusMsg = "Invalid amount"
		return m, nil
	}

	platform := strings.TrimSpace(m.inputs[loanFieldPlatform].Value())
	if platform == "" {
		m.statusMsg = "Platform is required for loans"
		return m, nil
	}

	// Interest rate is optional
	var interestRate *float64
	rateStr := strings.TrimSpace(m.inputs[loanFieldInterestRate].Value())
	if rateStr != "" {
		rateVal, err := strconv.ParseFloat(rateStr, 64)
		if err != nil || rateVal < 0 {
			m.statusMsg = "Invalid interest rate"
			return m, nil
		}
		interestRate = &rateVal
	}

	notes := strings.TrimSpace(m.inputs[loanFieldNotes].Value())

	components.BlurAll(m.inputs)
	return m, m.addLoan(coin, amount, platform, interestRate, notes)
}

func (m LoanModel) addLoan(coin string, amount float64, platform string, interestRate *float64, notes string) tea.Cmd {
	return func() tea.Msg {
		loan, err := m.portfolio.AddLoan(coin, amount, platform, interestRate, notes, "")
		if err != nil {
			return LoanAddedMsg{Error: err}
		}
		return LoanAddedMsg{Loan: &loan}
	}
}

func (m LoanModel) deleteLoan(id string) tea.Cmd {
	return func() tea.Msg {
		removed, err := m.portfolio.RemoveLoan(id)
		if err != nil {
			return LoanDeletedMsg{ID: id, Error: err}
		}
		if !removed {
			return LoanDeletedMsg{ID: id, Error: fmt.Errorf("loan not found")}
		}
		return LoanDeletedMsg{ID: id}
	}
}

// View renders the loan view
func (m LoanModel) View() string {
	switch m.mode {
	case LoanAdd:
		return m.renderAddForm()
	case LoanConfirmDelete:
		return m.renderDeleteConfirm()
	default:
		return m.renderList()
	}
}

func (m LoanModel) renderList() string {
	var b strings.Builder

	b.WriteString(components.RenderTitle("LOANS"))
	b.WriteString("\n\n")

	if len(m.loans) == 0 {
		b.WriteString(components.RenderEmptyState("No loans yet. Press 'a' to add one."))
		b.WriteString("\n")
	} else {
		// Column headers
		headerStyle := lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Bold(true)
		header := fmt.Sprintf("  %-8s  %14s  %-20s  %8s  %s",
			"Coin", "Amount", "Platform", "Rate", "Date")
		b.WriteString(headerStyle.Render(header))
		b.WriteString("\n")

		// Separator
		b.WriteString(components.RenderSeparator(tui.SeparatorWidthLoan))
		b.WriteString("\n")

		// List items
		for i, l := range m.loans {
			b.WriteString(m.renderLoanRow(i, l))
		}
	}

	// Status message
	if m.statusMsg != "" {
		b.WriteString("\n")
		b.WriteString(components.RenderStatusMessage(m.statusMsg, false))
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(components.RenderHelp(components.ListHelp(len(m.loans) > 0)))

	return components.RenderBoxDefault(b.String())
}

func (m LoanModel) renderLoanRow(index int, l models.Loan) string {
	isSelected := index == m.cursor

	// Cursor
	cursor := "  "
	if isSelected {
		cursor = lipgloss.NewStyle().
			Foreground(tui.PrimaryColor).
			Render("> ")
	}

	// Format values
	date := format.TruncateDate(l.Date)
	platform := format.TruncatePlatformLong(l.Platform)

	rateStr := "-"
	if l.InterestRate != nil {
		rateStr = fmt.Sprintf("%.2f%%", *l.InterestRate)
	}

	// Build row
	rowStyle := lipgloss.NewStyle().Foreground(tui.TextColor)
	if isSelected {
		rowStyle = rowStyle.Bold(true).Foreground(tui.PrimaryColor)
	}

	row := fmt.Sprintf("%-8s  %14.6f  %-20s  %8s  %s",
		l.Coin,
		l.Amount,
		platform,
		rateStr,
		date)

	return cursor + rowStyle.Render(row) + "\n"
}

func (m LoanModel) renderAddForm() string {
	var b strings.Builder

	b.WriteString(components.RenderTitle("ADD LOAN"))
	b.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().
		Foreground(tui.SubtleTextColor).
		Width(14)

	focusedLabelStyle := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Width(14)

	fields := []struct {
		label string
		index int
	}{
		{"Coin:", loanFieldCoin},
		{"Amount:", loanFieldAmount},
		{"Platform:", loanFieldPlatform},
		{"Rate (%):", loanFieldInterestRate},
		{"Notes:", loanFieldNotes},
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

func (m LoanModel) renderDeleteConfirm() string {
	var b strings.Builder

	b.WriteString(components.RenderErrorTitle("CONFIRM REPAY/REMOVE"))
	b.WriteString("\n\n")

	if m.cursor < len(m.loans) {
		l := m.loans[m.cursor]

		infoStyle := lipgloss.NewStyle().Foreground(tui.TextColor)
		info := fmt.Sprintf("Remove %.6f %s loan from %s?",
			l.Amount, l.Coin, l.Platform)
		b.WriteString(infoStyle.Render(info))
		b.WriteString("\n\n")
	}

	b.WriteString(components.RenderHelp(components.DeleteConfirmHelp()))

	return components.RenderBoxError(b.String())
}

// GetPortfolio returns the portfolio instance
func (m LoanModel) GetPortfolio() portfolio.LoansManager {
	return m.portfolio
}
