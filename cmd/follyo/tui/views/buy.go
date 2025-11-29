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

// Form field indices for buy form
const (
	buyFieldCoin = iota
	buyFieldAmount
	buyFieldPrice
	buyFieldPlatform
	buyFieldNotes
)

// BuyModel represents the buy/purchases view
type BuyModel struct {
	portfolio       portfolio.HoldingsManager
	defaultPlatform string
	holdings        []models.Holding
	state           EntityViewState
	config          EntityViewConfig
}

// NewBuyModel creates a new buy view model
func NewBuyModel(p portfolio.HoldingsManager, defaultPlatform string) BuyModel {
	fields := []FormFieldConfig{
		{Label: "Coin:", Placeholder: "BTC, ETH, SOL...", CharLimit: tui.InputCoinCharLimit, Width: tui.InputCoinWidth},
		{Label: "Amount:", Placeholder: "0.5", CharLimit: tui.InputAmountCharLimit, Width: tui.InputAmountWidth},
		{Label: "Price ($):", Placeholder: "50000.00", CharLimit: tui.InputPriceCharLimit, Width: tui.InputPriceWidth},
		{Label: "Platform:", Placeholder: "Coinbase, Binance...", CharLimit: tui.InputPlatformCharLimit, Width: tui.InputPlatformWidth, DefaultValue: defaultPlatform},
		{Label: "Notes:", Placeholder: "Optional notes...", CharLimit: tui.InputNotesCharLimit, Width: tui.InputNotesWidth},
	}

	m := BuyModel{
		portfolio:       p,
		defaultPlatform: defaultPlatform,
		state:           NewEntityViewState(fields),
		config: EntityViewConfig{
			Title:          "PURCHASES",
			EntityName:     "purchase",
			EmptyMessage:   "No purchases yet. Press 'a' to add one.",
			ColumnHeader:   fmt.Sprintf("  %-8s  %-12s  %14s  %14s  %-12s  %s", "Coin", "Amount", "Price", "Total", "Platform", "Date"),
			SeparatorWidth: tui.SeparatorWidthBuy,
			FormTitle:      "ADD PURCHASE",
			FormLabels:     []string{"Coin:", "Amount:", "Price ($):", "Platform:", "Notes:"},
			RenderRow:      nil, // Set below
			RenderDeleteInfo: func(item interface{}) string {
				h := item.(models.Holding)
				total := h.Amount * h.PurchasePriceUSD
				return fmt.Sprintf("Delete purchase of %.6f %s for %s?", h.Amount, h.Coin, format.USDSimple(total))
			},
		},
	}

	// Set render row callback
	m.config.RenderRow = func(index, cursor int, item interface{}) string {
		return renderHoldingRowWithCursor(index, cursor, item.(models.Holding))
	}

	m.loadHoldings()
	return m
}

func (m *BuyModel) loadHoldings() {
	holdings, err := m.portfolio.ListHoldings()
	if err != nil {
		m.state.Err = err
		return
	}
	m.holdings = holdings
}

// Init initializes the buy model
func (m BuyModel) Init() tea.Cmd {
	return nil
}

// HoldingAddedMsg is sent when a holding is added
type HoldingAddedMsg struct {
	Holding *models.Holding
	Error   error
}

// HoldingDeletedMsg is sent when a holding is deleted
type HoldingDeletedMsg struct {
	ID    string
	Error error
}

// Update handles messages for the buy model
func (m BuyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case HoldingAddedMsg:
		if msg.Error != nil {
			m.state.Err = msg.Error
			m.state.SetStatusMsg(fmt.Sprintf("Error: %v", msg.Error))
		} else {
			m.state.SetStatusMsg(fmt.Sprintf("Added %s purchase!", msg.Holding.Coin))
			m.loadHoldings()
		}
		m.state.Mode = EntityModeList

	case HoldingDeletedMsg:
		if msg.Error != nil {
			m.state.Err = msg.Error
			m.state.SetStatusMsg(fmt.Sprintf("Error: %v", msg.Error))
		} else {
			m.state.SetStatusMsg("Purchase deleted")
			m.loadHoldings()
			m.state.AdjustCursorAfterDelete(len(m.holdings))
		}
	}

	return m, nil
}

func (m BuyModel) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.state.Keys.Back):
		return m, func() tea.Msg { return tui.BackToMenuMsg{} }

	case key.Matches(msg, m.state.Keys.Quit):
		return m, tea.Quit

	case m.state.HandleListNavigation(msg, len(m.holdings)):
		// Navigation handled

	case msg.String() == "a" || msg.String() == "n":
		defaults := []string{"", "", "", m.defaultPlatform, ""}
		return m, m.state.EnterAddMode(defaults)

	case msg.String() == "d" || msg.String() == "x":
		if len(m.holdings) > 0 {
			m.state.EnterDeleteMode()
		}
	}

	return m, nil
}

func (m BuyModel) handleAddKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func (m BuyModel) handleDeleteConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if len(m.holdings) > 0 {
			id := m.holdings[m.state.Cursor].ID
			m.state.Mode = EntityModeList
			return m, m.deleteHolding(id)
		}
	case "n", "N", "escape":
		m.state.ExitToList()
	}
	return m, nil
}

func (m BuyModel) submitForm() (tea.Model, tea.Cmd) {
	coin := strings.ToUpper(m.state.GetFieldValue(buyFieldCoin))
	if coin == "" {
		m.state.SetStatusMsg("Coin is required")
		return m, nil
	}

	amount, err := strconv.ParseFloat(m.state.GetFieldValue(buyFieldAmount), 64)
	if err != nil || amount <= 0 {
		m.state.SetStatusMsg("Invalid amount")
		return m, nil
	}

	price, err := strconv.ParseFloat(m.state.GetFieldValue(buyFieldPrice), 64)
	if err != nil || price < 0 {
		m.state.SetStatusMsg("Invalid price")
		return m, nil
	}

	platform := m.state.GetFieldValue(buyFieldPlatform)
	notes := m.state.GetFieldValue(buyFieldNotes)

	m.state.ExitToList()
	return m, m.addHolding(coin, amount, price, platform, notes)
}

func (m BuyModel) addHolding(coin string, amount, price float64, platform, notes string) tea.Cmd {
	return func() tea.Msg {
		holding, err := m.portfolio.AddHolding(coin, amount, price, platform, notes, "")
		if err != nil {
			return HoldingAddedMsg{Error: err}
		}
		return HoldingAddedMsg{Holding: &holding}
	}
}

func (m BuyModel) deleteHolding(id string) tea.Cmd {
	return func() tea.Msg {
		removed, err := m.portfolio.RemoveHolding(id)
		if err != nil {
			return HoldingDeletedMsg{ID: id, Error: err}
		}
		if !removed {
			return HoldingDeletedMsg{ID: id, Error: fmt.Errorf("holding not found")}
		}
		return HoldingDeletedMsg{ID: id}
	}
}

// View renders the buy view
func (m BuyModel) View() string {
	switch m.state.Mode {
	case EntityModeAdd:
		return RenderAddForm(m.config, &m.state)
	case EntityModeConfirmDelete:
		if m.state.Cursor < len(m.holdings) {
			return RenderDeleteConfirm(m.config, m.holdings[m.state.Cursor])
		}
		return RenderDeleteConfirm(m.config, nil)
	default:
		return m.renderList()
	}
}

func (m BuyModel) renderList() string {
	items := make([]interface{}, len(m.holdings))
	for i, h := range m.holdings {
		items[i] = h
	}
	return RenderListView(m.config, &m.state, items)
}

// renderHoldingRowWithCursor renders a single holding row with cursor highlighting
func renderHoldingRowWithCursor(index, cursor int, h models.Holding) string {
	isSelected := index == cursor

	cursorStr := "  "
	if isSelected {
		cursorStr = lipgloss.NewStyle().
			Foreground(tui.PrimaryColor).
			Render("> ")
	}

	total := h.Amount * h.PurchasePriceUSD
	date := format.TruncateDate(h.Date)
	platform := format.TruncatePlatformShort(h.Platform)

	rowStyle := lipgloss.NewStyle().Foreground(tui.TextColor)
	if isSelected {
		rowStyle = rowStyle.Bold(true).Foreground(tui.PrimaryColor)
	}

	row := fmt.Sprintf("%-8s  %12.6f  %14s  %14s  %-12s  %s",
		h.Coin,
		h.Amount,
		format.USDSimple(h.PurchasePriceUSD),
		format.USDSimple(total),
		platform,
		date)

	return cursorStr + rowStyle.Render(row) + "\n"
}

// GetPortfolio returns the portfolio instance
func (m BuyModel) GetPortfolio() portfolio.HoldingsManager {
	return m.portfolio
}
