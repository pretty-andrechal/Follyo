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

// BuyViewMode represents the current mode of the buy view
type BuyViewMode int

const (
	BuyList BuyViewMode = iota
	BuyAdd
	BuyConfirmDelete
)

// Form field indices
const (
	fieldCoin = iota
	fieldAmount
	fieldPrice
	fieldPlatform
	fieldNotes
	fieldCount
)

// BuyModel represents the buy/purchases view
type BuyModel struct {
	portfolio       portfolio.HoldingsManager
	defaultPlatform string
	holdings        []models.Holding
	cursor          int
	mode            BuyViewMode
	inputs          []textinput.Model
	focusIndex      int
	keys            tui.KeyMap
	width           int
	height          int
	err             error
	statusMsg       string
}

// NewBuyModel creates a new buy view model
func NewBuyModel(p portfolio.HoldingsManager, defaultPlatform string) BuyModel {
	fields := []components.FormField{
		{Placeholder: "BTC, ETH, SOL...", CharLimit: tui.InputCoinCharLimit, Width: tui.InputCoinWidth},
		{Placeholder: "0.5", CharLimit: tui.InputAmountCharLimit, Width: tui.InputAmountWidth},
		{Placeholder: "50000.00", CharLimit: tui.InputPriceCharLimit, Width: tui.InputPriceWidth},
		{Placeholder: "Coinbase, Binance...", CharLimit: tui.InputPlatformCharLimit, Width: tui.InputPlatformWidth, DefaultValue: defaultPlatform},
		{Placeholder: "Optional notes...", CharLimit: tui.InputNotesCharLimit, Width: tui.InputNotesWidth},
	}

	m := BuyModel{
		portfolio:       p,
		defaultPlatform: defaultPlatform,
		inputs:          components.BuildFormInputs(fields),
		keys:            tui.DefaultKeyMap(),
		mode:            BuyList,
	}

	m.loadHoldings()
	return m
}

func (m *BuyModel) loadHoldings() {
	holdings, err := m.portfolio.ListHoldings()
	if err != nil {
		m.err = err
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
		switch m.mode {
		case BuyAdd:
			return m.handleAddKeys(msg)
		case BuyConfirmDelete:
			return m.handleDeleteConfirmKeys(msg)
		default:
			return m.handleListKeys(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case HoldingAddedMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.statusMsg = fmt.Sprintf("Error: %v", msg.Error)
		} else {
			m.statusMsg = fmt.Sprintf("Added %s purchase!", msg.Holding.Coin)
			m.loadHoldings()
		}
		m.mode = BuyList

	case HoldingDeletedMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.statusMsg = fmt.Sprintf("Error: %v", msg.Error)
		} else {
			m.statusMsg = "Purchase deleted"
			m.loadHoldings()
			if m.cursor >= len(m.holdings) && m.cursor > 0 {
				m.cursor--
			}
		}
	}

	return m, nil
}

func (m BuyModel) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		return m, func() tea.Msg { return tui.BackToMenuMsg{} }

	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		m.cursor = components.MoveCursorUp(m.cursor, len(m.holdings))

	case key.Matches(msg, m.keys.Down):
		m.cursor = components.MoveCursorDown(m.cursor, len(m.holdings))

	case msg.String() == "a" || msg.String() == "n":
		// Add new purchase
		m.mode = BuyAdd
		m.focusIndex = 0
		m.resetForm()
		return m, components.FocusField(m.inputs, fieldCoin)

	case msg.String() == "d" || msg.String() == "x":
		// Delete purchase
		if len(m.holdings) > 0 {
			m.mode = BuyConfirmDelete
			m.statusMsg = ""
		}
	}

	return m, nil
}

func (m BuyModel) handleAddKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.mode = BuyList
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
		if m.focusIndex == fieldCount-1 || msg.Alt {
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

func (m BuyModel) handleDeleteConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm delete
		if len(m.holdings) > 0 {
			id := m.holdings[m.cursor].ID
			m.mode = BuyList
			return m, m.deleteHolding(id)
		}
	case "n", "N", "escape":
		m.mode = BuyList
		m.statusMsg = ""
	}
	return m, nil
}

func (m *BuyModel) resetForm() {
	defaults := []string{"", "", "", m.defaultPlatform, ""}
	components.ResetFormInputs(m.inputs, defaults)
	m.statusMsg = ""
}

func (m BuyModel) submitForm() (tea.Model, tea.Cmd) {
	// Validate inputs
	coin := strings.ToUpper(strings.TrimSpace(m.inputs[fieldCoin].Value()))
	if coin == "" {
		m.statusMsg = "Coin is required"
		return m, nil
	}

	amountStr := strings.TrimSpace(m.inputs[fieldAmount].Value())
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		m.statusMsg = "Invalid amount"
		return m, nil
	}

	priceStr := strings.TrimSpace(m.inputs[fieldPrice].Value())
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price < 0 {
		m.statusMsg = "Invalid price"
		return m, nil
	}

	platform := strings.TrimSpace(m.inputs[fieldPlatform].Value())
	notes := strings.TrimSpace(m.inputs[fieldNotes].Value())

	components.BlurAll(m.inputs)
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
	switch m.mode {
	case BuyAdd:
		return m.renderAddForm()
	case BuyConfirmDelete:
		return m.renderDeleteConfirm()
	default:
		return m.renderList()
	}
}

func (m BuyModel) renderList() string {
	var b strings.Builder

	b.WriteString(components.RenderTitle("PURCHASES"))
	b.WriteString("\n\n")

	if len(m.holdings) == 0 {
		b.WriteString(components.RenderEmptyState("No purchases yet. Press 'a' to add one."))
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
		b.WriteString(components.RenderSeparator(tui.SeparatorWidthBuy))
		b.WriteString("\n")

		// List items
		for i, h := range m.holdings {
			b.WriteString(m.renderHoldingRow(i, h))
		}
	}

	// Status message
	if m.statusMsg != "" {
		b.WriteString("\n")
		b.WriteString(components.RenderStatusMessage(m.statusMsg, false))
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(components.RenderHelp(components.ListHelp(len(m.holdings) > 0)))

	return components.RenderBoxDefault(b.String())
}

func (m BuyModel) renderHoldingRow(index int, h models.Holding) string {
	isSelected := index == m.cursor

	// Cursor
	cursor := "  "
	if isSelected {
		cursor = lipgloss.NewStyle().
			Foreground(tui.PrimaryColor).
			Render("> ")
	}

	// Format values
	total := h.Amount * h.PurchasePriceUSD
	date := format.TruncateDate(h.Date)
	platform := format.TruncatePlatformShort(h.Platform)

	// Build row
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

	return cursor + rowStyle.Render(row) + "\n"
}

func (m BuyModel) renderAddForm() string {
	var b strings.Builder

	b.WriteString(components.RenderTitle("ADD PURCHASE"))
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
		{"Coin:", fieldCoin},
		{"Amount:", fieldAmount},
		{"Price ($):", fieldPrice},
		{"Platform:", fieldPlatform},
		{"Notes:", fieldNotes},
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

func (m BuyModel) renderDeleteConfirm() string {
	var b strings.Builder

	b.WriteString(components.RenderErrorTitle("CONFIRM DELETE"))
	b.WriteString("\n\n")

	if m.cursor < len(m.holdings) {
		h := m.holdings[m.cursor]
		total := h.Amount * h.PurchasePriceUSD

		infoStyle := lipgloss.NewStyle().Foreground(tui.TextColor)
		info := fmt.Sprintf("Delete purchase of %.6f %s for %s?",
			h.Amount, h.Coin, format.USDSimple(total))
		b.WriteString(infoStyle.Render(info))
		b.WriteString("\n\n")
	}

	b.WriteString(components.RenderHelp(components.DeleteConfirmHelp()))

	return components.RenderBoxError(b.String())
}

// GetPortfolio returns the portfolio instance
func (m BuyModel) GetPortfolio() portfolio.HoldingsManager {
	return m.portfolio
}
