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

// SellViewMode represents the current mode of the sell view
type SellViewMode int

const (
	SellList SellViewMode = iota
	SellAdd
	SellConfirmDelete
)

// Sell form field indices
const (
	sellFieldCoin = iota
	sellFieldAmount
	sellFieldPrice
	sellFieldPlatform
	sellFieldNotes
	sellFieldCount
)

// SellModel represents the sell/sales view
type SellModel struct {
	portfolio       *portfolio.Portfolio
	defaultPlatform string
	sales           []models.Sale
	cursor          int
	mode            SellViewMode
	inputs          []textinput.Model
	focusIndex      int
	keys            tui.KeyMap
	width           int
	height          int
	err             error
	statusMsg       string
}

// NewSellModel creates a new sell view model
func NewSellModel(p *portfolio.Portfolio, defaultPlatform string) SellModel {
	inputs := make([]textinput.Model, sellFieldCount)

	// Coin input
	inputs[sellFieldCoin] = textinput.New()
	inputs[sellFieldCoin].Placeholder = "BTC, ETH, SOL..."
	inputs[sellFieldCoin].CharLimit = tui.InputCoinCharLimit
	inputs[sellFieldCoin].Width = tui.InputCoinWidth

	// Amount input
	inputs[sellFieldAmount] = textinput.New()
	inputs[sellFieldAmount].Placeholder = "0.5"
	inputs[sellFieldAmount].CharLimit = tui.InputAmountCharLimit
	inputs[sellFieldAmount].Width = tui.InputAmountWidth

	// Price input
	inputs[sellFieldPrice] = textinput.New()
	inputs[sellFieldPrice].Placeholder = "55000.00"
	inputs[sellFieldPrice].CharLimit = tui.InputPriceCharLimit
	inputs[sellFieldPrice].Width = tui.InputPriceWidth

	// Platform input
	inputs[sellFieldPlatform] = textinput.New()
	inputs[sellFieldPlatform].Placeholder = "Coinbase, Binance..."
	inputs[sellFieldPlatform].CharLimit = tui.InputPlatformCharLimit
	inputs[sellFieldPlatform].Width = tui.InputPlatformWidth
	if defaultPlatform != "" {
		inputs[sellFieldPlatform].SetValue(defaultPlatform)
	}

	// Notes input
	inputs[sellFieldNotes] = textinput.New()
	inputs[sellFieldNotes].Placeholder = "Optional notes..."
	inputs[sellFieldNotes].CharLimit = tui.InputNotesCharLimit
	inputs[sellFieldNotes].Width = tui.InputNotesWidth

	m := SellModel{
		portfolio:       p,
		defaultPlatform: defaultPlatform,
		inputs:          inputs,
		keys:            tui.DefaultKeyMap(),
		mode:            SellList,
	}

	m.loadSales()
	return m
}

func (m *SellModel) loadSales() {
	sales, err := m.portfolio.ListSales()
	if err != nil {
		m.err = err
		return
	}
	m.sales = sales
}

// Init initializes the sell model
func (m SellModel) Init() tea.Cmd {
	return nil
}

// SaleAddedMsg is sent when a sale is added
type SaleAddedMsg struct {
	Sale  *models.Sale
	Error error
}

// SaleDeletedMsg is sent when a sale is deleted
type SaleDeletedMsg struct {
	ID    string
	Error error
}

// Update handles messages for the sell model
func (m SellModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case SellAdd:
			return m.handleAddKeys(msg)
		case SellConfirmDelete:
			return m.handleDeleteConfirmKeys(msg)
		default:
			return m.handleListKeys(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case SaleAddedMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.statusMsg = fmt.Sprintf("Error: %v", msg.Error)
		} else {
			m.statusMsg = fmt.Sprintf("Added %s sale!", msg.Sale.Coin)
			m.loadSales()
		}
		m.mode = SellList

	case SaleDeletedMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.statusMsg = fmt.Sprintf("Error: %v", msg.Error)
		} else {
			m.statusMsg = "Sale deleted"
			m.loadSales()
			if m.cursor >= len(m.sales) && m.cursor > 0 {
				m.cursor--
			}
		}
	}

	return m, nil
}

func (m SellModel) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		if m.cursor < len(m.sales)-1 {
			m.cursor++
		}

	case msg.String() == "a" || msg.String() == "n":
		// Add new sale
		m.mode = SellAdd
		m.focusIndex = 0
		m.resetForm()
		m.inputs[sellFieldCoin].Focus()
		m.statusMsg = ""
		return m, textinput.Blink

	case msg.String() == "d" || msg.String() == "x":
		// Delete sale
		if len(m.sales) > 0 {
			m.mode = SellConfirmDelete
			m.statusMsg = ""
		}
	}

	return m, nil
}

func (m SellModel) handleAddKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.mode = SellList
		m.inputs[m.focusIndex].Blur()
		m.statusMsg = ""
		return m, nil

	case tea.KeyTab, tea.KeyShiftTab, tea.KeyDown, tea.KeyUp:
		// Navigate between fields
		if msg.Type == tea.KeyUp || msg.Type == tea.KeyShiftTab {
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = sellFieldCount - 1
			}
		} else {
			m.focusIndex++
			if m.focusIndex >= sellFieldCount {
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
		if m.focusIndex == sellFieldCount-1 || msg.Alt {
			return m.submitForm()
		}
		// Otherwise move to next field
		m.focusIndex++
		if m.focusIndex >= sellFieldCount {
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

func (m SellModel) handleDeleteConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm delete
		if len(m.sales) > 0 {
			id := m.sales[m.cursor].ID
			m.mode = SellList
			return m, m.deleteSale(id)
		}
	case "n", "N", "escape":
		m.mode = SellList
		m.statusMsg = ""
	}
	return m, nil
}

func (m *SellModel) resetForm() {
	m.inputs[sellFieldCoin].SetValue("")
	m.inputs[sellFieldAmount].SetValue("")
	m.inputs[sellFieldPrice].SetValue("")
	m.inputs[sellFieldPlatform].SetValue(m.defaultPlatform)
	m.inputs[sellFieldNotes].SetValue("")
}

func (m SellModel) submitForm() (tea.Model, tea.Cmd) {
	// Validate inputs
	coin := strings.ToUpper(strings.TrimSpace(m.inputs[sellFieldCoin].Value()))
	if coin == "" {
		m.statusMsg = "Coin is required"
		return m, nil
	}

	amountStr := strings.TrimSpace(m.inputs[sellFieldAmount].Value())
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		m.statusMsg = "Invalid amount"
		return m, nil
	}

	priceStr := strings.TrimSpace(m.inputs[sellFieldPrice].Value())
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price < 0 {
		m.statusMsg = "Invalid price"
		return m, nil
	}

	platform := strings.TrimSpace(m.inputs[sellFieldPlatform].Value())
	notes := strings.TrimSpace(m.inputs[sellFieldNotes].Value())

	// Blur all inputs
	for i := range m.inputs {
		m.inputs[i].Blur()
	}

	return m, m.addSale(coin, amount, price, platform, notes)
}

func (m SellModel) addSale(coin string, amount, price float64, platform, notes string) tea.Cmd {
	return func() tea.Msg {
		sale, err := m.portfolio.AddSale(coin, amount, price, platform, notes, "")
		if err != nil {
			return SaleAddedMsg{Error: err}
		}
		return SaleAddedMsg{Sale: &sale}
	}
}

func (m SellModel) deleteSale(id string) tea.Cmd {
	return func() tea.Msg {
		removed, err := m.portfolio.RemoveSale(id)
		if err != nil {
			return SaleDeletedMsg{ID: id, Error: err}
		}
		if !removed {
			return SaleDeletedMsg{ID: id, Error: fmt.Errorf("sale not found")}
		}
		return SaleDeletedMsg{ID: id}
	}
}

// View renders the sell view
func (m SellModel) View() string {
	switch m.mode {
	case SellAdd:
		return m.renderAddForm()
	case SellConfirmDelete:
		return m.renderDeleteConfirm()
	default:
		return m.renderList()
	}
}

func (m SellModel) renderList() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Render("SALES")

	b.WriteString(title)
	b.WriteString("\n\n")

	if len(m.sales) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Italic(true)
		b.WriteString(emptyStyle.Render("  No sales yet. Press 'a' to add one."))
		b.WriteString("\n")
	} else {
		// Column headers
		headerStyle := lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Bold(true)
		header := fmt.Sprintf("  %-8s  %-12s  %14s  %14s  %-12s  %s",
			"Coin", "Amount", "Price", "Total", "Platform", "Date")
		b.WriteString(headerStyle.Render(header))
		b.WriteString("\n")

		// Separator
		sepStyle := lipgloss.NewStyle().Foreground(tui.BorderColor)
		b.WriteString(sepStyle.Render(strings.Repeat("─", tui.SeparatorWidthSell)))
		b.WriteString("\n")

		// List items
		for i, s := range m.sales {
			b.WriteString(m.renderSaleRow(i, s))
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

func (m SellModel) renderSaleRow(index int, s models.Sale) string {
	isSelected := index == m.cursor

	// Cursor
	cursor := "  "
	if isSelected {
		cursor = lipgloss.NewStyle().
			Foreground(tui.PrimaryColor).
			Render("> ")
	}

	// Format values
	total := s.Amount * s.SellPriceUSD
	date := format.TruncateDate(s.Date)
	platform := format.TruncatePlatformShort(s.Platform)

	// Build row
	rowStyle := lipgloss.NewStyle().Foreground(tui.TextColor)
	if isSelected {
		rowStyle = rowStyle.Bold(true).Foreground(tui.PrimaryColor)
	}

	row := fmt.Sprintf("%-8s  %12.6f  %14s  %14s  %-12s  %s",
		s.Coin,
		s.Amount,
		format.USDSimple(s.SellPriceUSD),
		format.USDSimple(total),
		platform,
		date)

	return cursor + rowStyle.Render(row) + "\n"
}

func (m SellModel) renderListHelp() string {
	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)

	var help string
	if len(m.sales) > 0 {
		help = fmt.Sprintf("%s navigate  %s add  %s delete  %s back  %s quit",
			tui.HelpKeyStyle.Render("↑↓"),
			tui.HelpKeyStyle.Render("a"),
			tui.HelpKeyStyle.Render("d"),
			tui.HelpKeyStyle.Render("esc"),
			tui.HelpKeyStyle.Render("q"))
	} else {
		help = fmt.Sprintf("%s add sale  %s back  %s quit",
			tui.HelpKeyStyle.Render("a"),
			tui.HelpKeyStyle.Render("esc"),
			tui.HelpKeyStyle.Render("q"))
	}

	return helpStyle.Render(help)
}

func (m SellModel) renderAddForm() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Render("ADD SALE")

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
		{"Coin:", sellFieldCoin},
		{"Amount:", sellFieldAmount},
		{"Price ($):", sellFieldPrice},
		{"Platform:", sellFieldPlatform},
		{"Notes:", sellFieldNotes},
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

func (m SellModel) renderDeleteConfirm() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(tui.ErrorColor).
		Bold(true).
		Padding(0, 2).
		Render("CONFIRM DELETE")

	b.WriteString(title)
	b.WriteString("\n\n")

	if m.cursor < len(m.sales) {
		s := m.sales[m.cursor]
		total := s.Amount * s.SellPriceUSD

		infoStyle := lipgloss.NewStyle().Foreground(tui.TextColor)
		info := fmt.Sprintf("Delete sale of %.6f %s for %s?",
			s.Amount, s.Coin, format.USDSimple(total))
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
func (m SellModel) GetPortfolio() *portfolio.Portfolio {
	return m.portfolio
}
