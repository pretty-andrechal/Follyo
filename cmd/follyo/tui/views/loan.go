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

// Form field indices for loan form
const (
	loanFieldCoin = iota
	loanFieldAmount
	loanFieldPlatform
	loanFieldInterestRate
	loanFieldNotes
)

// LoanModel represents the loan view
type LoanModel struct {
	portfolio       portfolio.LoansManager
	defaultPlatform string
	loans           []models.Loan
	state           EntityViewState
	config          EntityViewConfig
}

// NewLoanModel creates a new loan view model
func NewLoanModel(p portfolio.LoansManager, defaultPlatform string) LoanModel {
	fields := []FormFieldConfig{
		{Label: "Coin:", Placeholder: "USDT, USDC, BTC...", CharLimit: tui.InputCoinCharLimit, Width: tui.InputCoinWidth},
		{Label: "Amount:", Placeholder: "5000.0", CharLimit: tui.InputAmountCharLimit, Width: tui.InputAmountWidth},
		{Label: "Platform:", Placeholder: "Nexo, Celsius...", CharLimit: tui.InputPlatformCharLimit, Width: tui.InputPlatformWidth, DefaultValue: defaultPlatform},
		{Label: "Rate (%):", Placeholder: "6.9 (optional)", CharLimit: tui.InputRateCharLimit, Width: tui.InputRateWidth},
		{Label: "Notes:", Placeholder: "Optional notes...", CharLimit: tui.InputNotesCharLimit, Width: tui.InputNotesWidth},
	}

	m := LoanModel{
		portfolio:       p,
		defaultPlatform: defaultPlatform,
		state:           NewEntityViewState(fields),
		config: EntityViewConfig{
			Title:          "LOANS",
			EntityName:     "loan",
			EmptyMessage:   "No loans yet. Press 'a' to add one.",
			ColumnHeader:   fmt.Sprintf("  %-8s  %14s  %-20s  %8s  %s", "Coin", "Amount", "Platform", "Rate", "Date"),
			SeparatorWidth: tui.SeparatorWidthLoan,
			FormTitle:      "ADD LOAN",
			FormLabels:     []string{"Coin:", "Amount:", "Platform:", "Rate (%):", "Notes:"},
			RenderRow:      nil, // Set below
			RenderDeleteInfo: func(item interface{}) string {
				l := item.(models.Loan)
				return fmt.Sprintf("Remove %.6f %s loan from %s?", l.Amount, l.Coin, l.Platform)
			},
		},
	}

	m.config.RenderRow = func(index, cursor int, item interface{}) string {
		return renderLoanRowWithCursor(index, cursor, item.(models.Loan))
	}

	m.loadLoans()
	return m
}

func (m *LoanModel) loadLoans() {
	loans, err := m.portfolio.ListLoans()
	if err != nil {
		m.state.Err = err
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

	case LoanAddedMsg:
		if msg.Error != nil {
			m.state.Err = msg.Error
			m.state.SetStatusMsg(fmt.Sprintf("Error: %v", msg.Error))
		} else {
			m.state.SetStatusMsg(fmt.Sprintf("Added %s loan on %s!", msg.Loan.Coin, msg.Loan.Platform))
			m.loadLoans()
		}
		m.state.Mode = EntityModeList

	case LoanDeletedMsg:
		if msg.Error != nil {
			m.state.Err = msg.Error
			m.state.SetStatusMsg(fmt.Sprintf("Error: %v", msg.Error))
		} else {
			m.state.SetStatusMsg("Loan removed")
			m.loadLoans()
			m.state.AdjustCursorAfterDelete(len(m.loans))
		}
	}

	return m, nil
}

func (m LoanModel) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.state.Keys.Back):
		return m, func() tea.Msg { return tui.BackToMenuMsg{} }

	case key.Matches(msg, m.state.Keys.Quit):
		return m, tea.Quit

	case m.state.HandleListNavigation(msg, len(m.loans)):
		// Navigation handled

	case msg.String() == "a" || msg.String() == "n":
		defaults := []string{"", "", m.defaultPlatform, "", ""}
		return m, m.state.EnterAddMode(defaults)

	case msg.String() == "d" || msg.String() == "x":
		if len(m.loans) > 0 {
			m.state.EnterDeleteMode()
		}
	}

	return m, nil
}

func (m LoanModel) handleAddKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func (m LoanModel) handleDeleteConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if len(m.loans) > 0 {
			id := m.loans[m.state.Cursor].ID
			m.state.Mode = EntityModeList
			return m, m.deleteLoan(id)
		}
	case "n", "N", "escape":
		m.state.ExitToList()
	}
	return m, nil
}

func (m LoanModel) submitForm() (tea.Model, tea.Cmd) {
	coin := strings.ToUpper(m.state.GetFieldValue(loanFieldCoin))
	if coin == "" {
		m.state.SetStatusMsg("Coin is required")
		return m, nil
	}

	amount, err := strconv.ParseFloat(m.state.GetFieldValue(loanFieldAmount), 64)
	if err != nil || amount <= 0 {
		m.state.SetStatusMsg("Invalid amount")
		return m, nil
	}

	platform := m.state.GetFieldValue(loanFieldPlatform)
	if platform == "" {
		m.state.SetStatusMsg("Platform is required for loans")
		return m, nil
	}

	// Interest rate is optional
	var interestRate *float64
	rateStr := m.state.GetFieldValue(loanFieldInterestRate)
	if rateStr != "" {
		rateVal, err := strconv.ParseFloat(rateStr, 64)
		if err != nil || rateVal < 0 {
			m.state.SetStatusMsg("Invalid interest rate")
			return m, nil
		}
		interestRate = &rateVal
	}

	notes := m.state.GetFieldValue(loanFieldNotes)

	m.state.ExitToList()
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
	switch m.state.Mode {
	case EntityModeAdd:
		return RenderAddForm(m.config, &m.state)
	case EntityModeConfirmDelete:
		if m.state.Cursor < len(m.loans) {
			return RenderDeleteConfirm(m.config, m.loans[m.state.Cursor])
		}
		return RenderDeleteConfirm(m.config, nil)
	default:
		return m.renderList()
	}
}

func (m LoanModel) renderList() string {
	items := make([]interface{}, len(m.loans))
	for i, l := range m.loans {
		items[i] = l
	}
	return RenderListView(m.config, &m.state, items)
}

// renderLoanRowWithCursor renders a single loan row with cursor highlighting
func renderLoanRowWithCursor(index, cursor int, l models.Loan) string {
	isSelected := index == cursor

	cursorStr := "  "
	if isSelected {
		cursorStr = lipgloss.NewStyle().
			Foreground(tui.PrimaryColor).
			Render("> ")
	}

	date := format.TruncateDate(l.Date)
	platform := format.TruncatePlatformLong(l.Platform)

	rateStr := "-"
	if l.InterestRate != nil {
		rateStr = fmt.Sprintf("%.2f%%", *l.InterestRate)
	}

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

	return cursorStr + rowStyle.Render(row) + "\n"
}

// GetPortfolio returns the portfolio instance
func (m LoanModel) GetPortfolio() portfolio.LoansManager {
	return m.portfolio
}
