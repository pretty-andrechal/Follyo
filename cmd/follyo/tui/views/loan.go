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
	inputs := make([]textinput.Model, loanFieldCount)

	// Coin input
	inputs[loanFieldCoin] = textinput.New()
	inputs[loanFieldCoin].Placeholder = "USDT, USDC, BTC..."
	inputs[loanFieldCoin].CharLimit = tui.InputCoinCharLimit
	inputs[loanFieldCoin].Width = tui.InputCoinWidth

	// Amount input
	inputs[loanFieldAmount] = textinput.New()
	inputs[loanFieldAmount].Placeholder = "5000.0"
	inputs[loanFieldAmount].CharLimit = tui.InputAmountCharLimit
	inputs[loanFieldAmount].Width = tui.InputAmountWidth

	// Platform input (required for loans)
	inputs[loanFieldPlatform] = textinput.New()
	inputs[loanFieldPlatform].Placeholder = "Nexo, Celsius..."
	inputs[loanFieldPlatform].CharLimit = tui.InputPlatformCharLimit
	inputs[loanFieldPlatform].Width = tui.InputPlatformWidth
	if defaultPlatform != "" {
		inputs[loanFieldPlatform].SetValue(defaultPlatform)
	}

	// Interest rate input (optional)
	inputs[loanFieldInterestRate] = textinput.New()
	inputs[loanFieldInterestRate].Placeholder = "6.9 (optional)"
	inputs[loanFieldInterestRate].CharLimit = tui.InputRateCharLimit
	inputs[loanFieldInterestRate].Width = tui.InputRateWidth

	// Notes input
	inputs[loanFieldNotes] = textinput.New()
	inputs[loanFieldNotes].Placeholder = "Optional notes..."
	inputs[loanFieldNotes].CharLimit = tui.InputNotesCharLimit
	inputs[loanFieldNotes].Width = tui.InputNotesWidth

	m := LoanModel{
		portfolio:       p,
		defaultPlatform: defaultPlatform,
		inputs:          inputs,
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
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.loans)-1 {
			m.cursor++
		}

	case msg.String() == "a" || msg.String() == "n":
		// Add new loan
		m.mode = LoanAdd
		m.focusIndex = 0
		m.resetForm()
		m.inputs[loanFieldCoin].Focus()
		m.statusMsg = ""
		return m, textinput.Blink

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
		m.inputs[m.focusIndex].Blur()
		m.statusMsg = ""
		return m, nil

	case tea.KeyTab, tea.KeyShiftTab, tea.KeyDown, tea.KeyUp:
		// Navigate between fields
		if msg.Type == tea.KeyUp || msg.Type == tea.KeyShiftTab {
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = loanFieldCount - 1
			}
		} else {
			m.focusIndex++
			if m.focusIndex >= loanFieldCount {
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
		if m.focusIndex == loanFieldCount-1 || msg.Alt {
			return m.submitForm()
		}
		// Otherwise move to next field
		m.focusIndex++
		if m.focusIndex >= loanFieldCount {
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
	m.inputs[loanFieldCoin].SetValue("")
	m.inputs[loanFieldAmount].SetValue("")
	m.inputs[loanFieldPlatform].SetValue(m.defaultPlatform)
	m.inputs[loanFieldInterestRate].SetValue("")
	m.inputs[loanFieldNotes].SetValue("")
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

	// Blur all inputs
	for i := range m.inputs {
		m.inputs[i].Blur()
	}

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

	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Render("LOANS")

	b.WriteString(title)
	b.WriteString("\n\n")

	if len(m.loans) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Italic(true)
		b.WriteString(emptyStyle.Render("  No loans yet. Press 'a' to add one."))
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
		sepStyle := lipgloss.NewStyle().Foreground(tui.BorderColor)
		b.WriteString(sepStyle.Render(strings.Repeat("─", tui.SeparatorWidthLoan)))
		b.WriteString("\n")

		// List items
		for i, l := range m.loans {
			b.WriteString(m.renderLoanRow(i, l))
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

func (m LoanModel) renderListHelp() string {
	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)

	var help string
	if len(m.loans) > 0 {
		help = fmt.Sprintf("%s navigate  %s add  %s repay  %s back  %s quit",
			tui.HelpKeyStyle.Render("↑↓"),
			tui.HelpKeyStyle.Render("a"),
			tui.HelpKeyStyle.Render("d"),
			tui.HelpKeyStyle.Render("esc"),
			tui.HelpKeyStyle.Render("q"))
	} else {
		help = fmt.Sprintf("%s add loan  %s back  %s quit",
			tui.HelpKeyStyle.Render("a"),
			tui.HelpKeyStyle.Render("esc"),
			tui.HelpKeyStyle.Render("q"))
	}

	return helpStyle.Render(help)
}

func (m LoanModel) renderAddForm() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Render("ADD LOAN")

	b.WriteString(title)
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

func (m LoanModel) renderDeleteConfirm() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(tui.ErrorColor).
		Bold(true).
		Padding(0, 2).
		Render("CONFIRM REPAY/REMOVE")

	b.WriteString(title)
	b.WriteString("\n\n")

	if m.cursor < len(m.loans) {
		l := m.loans[m.cursor]

		infoStyle := lipgloss.NewStyle().Foreground(tui.TextColor)
		info := fmt.Sprintf("Remove %.6f %s loan from %s?",
			l.Amount, l.Coin, l.Platform)
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
func (m LoanModel) GetPortfolio() portfolio.LoansManager {
	return m.portfolio
}
