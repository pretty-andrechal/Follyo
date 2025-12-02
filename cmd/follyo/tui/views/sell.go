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

// Form field indices for sell form
const (
	sellFieldCoin = iota
	sellFieldAmount
	sellFieldPrice
	sellFieldTotal
	sellFieldDate
	sellFieldPlatform
	sellFieldNotes
)

// SellModel represents the sell/sales view
type SellModel struct {
	portfolio       portfolio.SalesManager
	defaultPlatform string
	sales           []models.Sale
	state           EntityViewState
	config          EntityViewConfig
}

// NewSellModel creates a new sell view model
func NewSellModel(p portfolio.SalesManager, defaultPlatform string) SellModel {
	fields := []FormFieldConfig{
		{Label: "Coin:", Placeholder: "BTC, ETH, SOL...", CharLimit: tui.InputCoinCharLimit, Width: tui.InputCoinWidth},
		{Label: "Amount:", Placeholder: "0.5", CharLimit: tui.InputAmountCharLimit, Width: tui.InputAmountWidth},
		{Label: "Price ($):", Placeholder: "55000.00 (per unit)", CharLimit: tui.InputPriceCharLimit, Width: tui.InputPriceWidth},
		{Label: "Total ($):", Placeholder: "OR enter total amount", CharLimit: tui.InputPriceCharLimit, Width: tui.InputPriceWidth},
		{Label: "Date:", Placeholder: "YYYY-MM-DD (blank=today)", CharLimit: tui.InputDateCharLimit, Width: tui.InputDateWidth},
		{Label: "Platform:", Placeholder: "Coinbase, Binance...", CharLimit: tui.InputPlatformCharLimit, Width: tui.InputPlatformWidth, DefaultValue: defaultPlatform},
		{Label: "Notes:", Placeholder: "Optional notes...", CharLimit: tui.InputNotesCharLimit, Width: tui.InputNotesWidth},
	}

	m := SellModel{
		portfolio:       p,
		defaultPlatform: defaultPlatform,
		state:           NewEntityViewState(fields),
		config: EntityViewConfig{
			Title:          "SALES",
			EntityName:     "sale",
			EmptyMessage:   "No sales yet. Press 'a' to add one.",
			ColumnHeader:   fmt.Sprintf("  %-8s  %-12s  %14s  %14s  %-12s  %s", "Coin", "Amount", "Price", "Total", "Platform", "Date"),
			SeparatorWidth: tui.SeparatorWidthSell,
			FormTitle:      "ADD SALE",
			FormLabels:     []string{"Coin:", "Amount:", "Price ($):", "Total ($):", "Date:", "Platform:", "Notes:"},
			RenderRow:      nil, // Set below
			RenderDeleteInfo: func(item interface{}) string {
				s := item.(models.Sale)
				total := s.Amount * s.SellPriceUSD
				return fmt.Sprintf("Delete sale of %.6f %s for %s?", s.Amount, s.Coin, format.USDSimple(total))
			},
		},
	}

	m.config.RenderRow = func(index, cursor int, item interface{}) string {
		return renderSaleRowWithCursor(index, cursor, item.(models.Sale))
	}

	m.loadSales()
	return m
}

func (m *SellModel) loadSales() {
	sales, err := m.portfolio.ListSales()
	if err != nil {
		m.state.Err = err
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

	case SaleAddedMsg:
		if msg.Error != nil {
			m.state.Err = msg.Error
			m.state.SetStatusMsg(fmt.Sprintf("Error: %v", msg.Error))
		} else {
			m.state.SetStatusMsg(fmt.Sprintf("Added %s sale!", msg.Sale.Coin))
			m.loadSales()
		}
		m.state.Mode = EntityModeList

	case SaleDeletedMsg:
		if msg.Error != nil {
			m.state.Err = msg.Error
			m.state.SetStatusMsg(fmt.Sprintf("Error: %v", msg.Error))
		} else {
			m.state.SetStatusMsg("Sale deleted")
			m.loadSales()
			m.state.AdjustCursorAfterDelete(len(m.sales))
		}
	}

	return m, nil
}

func (m SellModel) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.state.Keys.Back):
		return m, func() tea.Msg { return tui.BackToMenuMsg{} }

	case key.Matches(msg, m.state.Keys.Quit):
		return m, tea.Quit

	case m.state.HandleListNavigation(msg, len(m.sales)):
		// Navigation handled

	case msg.String() == "a" || msg.String() == "n":
		defaults := []string{"", "", "", "", "", m.defaultPlatform, ""}
		return m, m.state.EnterAddMode(defaults)

	case msg.String() == "d" || msg.String() == "x":
		if len(m.sales) > 0 {
			m.state.EnterDeleteMode()
		}
	}

	return m, nil
}

func (m SellModel) handleAddKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func (m SellModel) handleDeleteConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if len(m.sales) > 0 {
			id := m.sales[m.state.Cursor].ID
			m.state.Mode = EntityModeList
			return m, m.deleteSale(id)
		}
	case "n", "N", "escape":
		m.state.ExitToList()
	}
	return m, nil
}

func (m SellModel) submitForm() (tea.Model, tea.Cmd) {
	coin := strings.ToUpper(m.state.GetFieldValue(sellFieldCoin))
	if coin == "" {
		m.state.SetStatusMsg("Coin is required")
		return m, nil
	}

	amount, err := strconv.ParseFloat(m.state.GetFieldValue(sellFieldAmount), 64)
	if err != nil || amount <= 0 {
		m.state.SetStatusMsg("Invalid amount")
		return m, nil
	}

	// Get price per unit and total fields
	priceStr := m.state.GetFieldValue(sellFieldPrice)
	totalStr := m.state.GetFieldValue(sellFieldTotal)

	var price float64

	// If total is provided, calculate price from total
	if totalStr != "" {
		total, err := strconv.ParseFloat(totalStr, 64)
		if err != nil || total < 0 {
			m.state.SetStatusMsg("Invalid total amount")
			return m, nil
		}
		// Calculate per-unit price from total
		price = total / amount
	} else if priceStr != "" {
		// Use per-unit price directly
		var err error
		price, err = strconv.ParseFloat(priceStr, 64)
		if err != nil || price < 0 {
			m.state.SetStatusMsg("Invalid price")
			return m, nil
		}
	} else {
		m.state.SetStatusMsg("Enter either Price or Total")
		return m, nil
	}

	date := m.state.GetFieldValue(sellFieldDate)
	platform := m.state.GetFieldValue(sellFieldPlatform)
	notes := m.state.GetFieldValue(sellFieldNotes)

	m.state.ExitToList()
	return m, m.addSale(coin, amount, price, platform, notes, date)
}

func (m SellModel) addSale(coin string, amount, price float64, platform, notes, date string) tea.Cmd {
	return func() tea.Msg {
		sale, err := m.portfolio.AddSale(coin, amount, price, platform, notes, date)
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
	switch m.state.Mode {
	case EntityModeAdd:
		return RenderAddForm(m.config, &m.state)
	case EntityModeConfirmDelete:
		if m.state.Cursor < len(m.sales) {
			return RenderDeleteConfirm(m.config, m.sales[m.state.Cursor])
		}
		return RenderDeleteConfirm(m.config, nil)
	default:
		return m.renderList()
	}
}

func (m SellModel) renderList() string {
	items := make([]interface{}, len(m.sales))
	for i, s := range m.sales {
		items[i] = s
	}
	return RenderListView(m.config, &m.state, items)
}

// renderSaleRowWithCursor renders a single sale row with cursor highlighting
func renderSaleRowWithCursor(index, cursor int, s models.Sale) string {
	isSelected := index == cursor

	cursorStr := "  "
	if isSelected {
		cursorStr = lipgloss.NewStyle().
			Foreground(tui.PrimaryColor).
			Render("> ")
	}

	total := s.Amount * s.SellPriceUSD
	date := format.TruncateDate(s.Date)
	platform := format.TruncatePlatformShort(s.Platform)

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

	return cursorStr + rowStyle.Render(row) + "\n"
}

// GetPortfolio returns the portfolio instance
func (m SellModel) GetPortfolio() portfolio.SalesManager {
	return m.portfolio
}
