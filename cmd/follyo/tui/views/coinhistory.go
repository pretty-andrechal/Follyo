package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui/components"
	"github.com/pretty-andrechal/follyo/internal/storage"
)

// CoinHistoryViewMode represents the current mode of the coin history view
type CoinHistoryViewMode int

const (
	CoinHistoryCoinSelect CoinHistoryViewMode = iota
	CoinHistoryDisplay
	CoinHistoryCompare
)

// CoinDataPoint represents a single historical data point for a coin
type CoinDataPoint struct {
	Timestamp time.Time
	Price     float64
	Amount    float64
	Value     float64
}

// CoinHistoryModel represents the coin history view
type CoinHistoryModel struct {
	store          *storage.SnapshotStore
	availableCoins []string
	selectedCoin   string                     // For single coin view
	selectedCoins  map[string]bool            // For multi-select
	compareCoins   []string                   // Ordered list of coins to compare
	coinData       []CoinDataPoint            // Single coin data
	coinDataMap    map[string][]CoinDataPoint // Multi-coin data
	cursor         int
	mode           CoinHistoryViewMode
	viewport       viewport.Model
	viewportReady  bool
	keys           tui.KeyMap
	width          int
	height         int
	err            error
	statusMsg      string
}

// NewCoinHistoryModel creates a new coin history view model
func NewCoinHistoryModel(store *storage.SnapshotStore) CoinHistoryModel {
	m := CoinHistoryModel{
		store:         store,
		keys:          tui.DefaultKeyMap(),
		mode:          CoinHistoryCoinSelect,
		selectedCoins: make(map[string]bool),
		coinDataMap:   make(map[string][]CoinDataPoint),
	}

	m.extractAvailableCoins()
	return m
}

// extractAvailableCoins scans all snapshots to find unique coins
func (m *CoinHistoryModel) extractAvailableCoins() {
	coinSet := make(map[string]bool)
	for _, snap := range m.store.List() {
		for coin := range snap.CoinValues {
			coinSet[coin] = true
		}
	}

	m.availableCoins = make([]string, 0, len(coinSet))
	for coin := range coinSet {
		m.availableCoins = append(m.availableCoins, coin)
	}
	sort.Strings(m.availableCoins)
}

// loadCoinHistory extracts data points for the selected coin from all snapshots
func (m *CoinHistoryModel) loadCoinHistory(coin string) {
	snapshots := m.store.List() // Returns newest first
	m.coinData = make([]CoinDataPoint, 0)

	// Iterate in reverse for chronological order (oldest first)
	for i := len(snapshots) - 1; i >= 0; i-- {
		snap := snapshots[i]
		if cv, ok := snap.CoinValues[coin]; ok {
			m.coinData = append(m.coinData, CoinDataPoint{
				Timestamp: snap.Timestamp,
				Price:     cv.Price,
				Amount:    cv.Amount,
				Value:     cv.Value,
			})
		}
	}
}

// loadMultiCoinHistory loads history for all selected coins
func (m *CoinHistoryModel) loadMultiCoinHistory() {
	m.coinDataMap = make(map[string][]CoinDataPoint)
	m.compareCoins = make([]string, 0)

	// Get sorted list of selected coins
	for coin := range m.selectedCoins {
		m.compareCoins = append(m.compareCoins, coin)
	}
	sort.Strings(m.compareCoins)

	snapshots := m.store.List() // Returns newest first

	for _, coin := range m.compareCoins {
		data := make([]CoinDataPoint, 0)
		// Iterate in reverse for chronological order (oldest first)
		for i := len(snapshots) - 1; i >= 0; i-- {
			snap := snapshots[i]
			if cv, ok := snap.CoinValues[coin]; ok {
				data = append(data, CoinDataPoint{
					Timestamp: snap.Timestamp,
					Price:     cv.Price,
					Amount:    cv.Amount,
					Value:     cv.Value,
				})
			}
		}
		m.coinDataMap[coin] = data
	}
}

// Init initializes the coin history model
func (m CoinHistoryModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the coin history model
func (m CoinHistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case CoinHistoryDisplay:
			return m.handleDisplayKeys(msg)
		case CoinHistoryCompare:
			return m.handleCompareKeys(msg)
		default:
			return m.handleCoinSelectKeys(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 3
		footerHeight := 3
		verticalMargins := headerHeight + footerHeight

		if !m.viewportReady {
			m.viewport = viewport.New(msg.Width-4, msg.Height-verticalMargins)
			m.viewport.YPosition = headerHeight
			m.viewportReady = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - verticalMargins
		}
	}

	return m, nil
}

func (m CoinHistoryModel) handleCoinSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		// If we have selections, clear them first
		if len(m.selectedCoins) > 0 {
			m.selectedCoins = make(map[string]bool)
			m.statusMsg = ""
			return m, nil
		}
		return m, func() tea.Msg { return tui.BackToMenuMsg{} }

	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		m.cursor = components.MoveCursorUp(m.cursor, len(m.availableCoins))

	case key.Matches(msg, m.keys.Down):
		m.cursor = components.MoveCursorDown(m.cursor, len(m.availableCoins))

	case msg.String() == " ":
		// Toggle selection with space
		if len(m.availableCoins) > 0 {
			coin := m.availableCoins[m.cursor]
			if m.selectedCoins[coin] {
				delete(m.selectedCoins, coin)
			} else {
				m.selectedCoins[coin] = true
			}
			m.updateSelectionStatus()
		}

	case key.Matches(msg, m.keys.Select):
		if len(m.availableCoins) > 0 {
			if len(m.selectedCoins) >= 2 {
				// Compare multiple coins
				m.loadMultiCoinHistory()
				m.mode = CoinHistoryCompare
				m.updateCompareViewport()
			} else if len(m.selectedCoins) == 1 {
				// Single selected coin
				for coin := range m.selectedCoins {
					m.selectedCoin = coin
				}
				m.loadCoinHistory(m.selectedCoin)
				m.mode = CoinHistoryDisplay
				m.updateDisplayViewport()
			} else {
				// No selections, use cursor position
				m.selectedCoin = m.availableCoins[m.cursor]
				m.loadCoinHistory(m.selectedCoin)
				m.mode = CoinHistoryDisplay
				m.updateDisplayViewport()
			}
		}
	}

	return m, nil
}

func (m *CoinHistoryModel) updateSelectionStatus() {
	count := len(m.selectedCoins)
	if count == 0 {
		m.statusMsg = ""
	} else if count == 1 {
		m.statusMsg = "1 coin selected (select more to compare)"
	} else {
		m.statusMsg = fmt.Sprintf("%d coins selected - press enter to compare", count)
	}
}

func (m CoinHistoryModel) handleDisplayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.mode = CoinHistoryCoinSelect
		m.selectedCoin = ""
		m.coinData = nil
		return m, nil

	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	default:
		// Forward to viewport for scrolling
		if m.viewportReady {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m CoinHistoryModel) handleCompareKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.mode = CoinHistoryCoinSelect
		m.compareCoins = nil
		m.coinDataMap = make(map[string][]CoinDataPoint)
		return m, nil

	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	default:
		// Forward to viewport for scrolling
		if m.viewportReady {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *CoinHistoryModel) updateDisplayViewport() {
	if !m.viewportReady {
		return
	}

	m.viewport.SetContent(m.renderDisplayContent())
	m.viewport.GotoTop()
}

func (m *CoinHistoryModel) updateCompareViewport() {
	if !m.viewportReady {
		return
	}

	m.viewport.SetContent(m.renderCompareContent())
	m.viewport.GotoTop()
}

// View renders the coin history view
func (m CoinHistoryModel) View() string {
	switch m.mode {
	case CoinHistoryDisplay:
		return m.renderDisplay()
	case CoinHistoryCompare:
		return m.renderCompare()
	default:
		return m.renderCoinSelect()
	}
}

func (m CoinHistoryModel) renderCoinSelect() string {
	var b strings.Builder

	b.WriteString(components.RenderTitle("COIN HISTORY"))
	b.WriteString("\n\n")

	if len(m.availableCoins) == 0 {
		b.WriteString(components.RenderEmptyState("No coins found in snapshots.\nCreate some snapshots first to track coin history."))
		b.WriteString("\n")
	} else {
		// Instructions
		labelStyle := lipgloss.NewStyle().Foreground(tui.SubtleTextColor)
		b.WriteString(labelStyle.Render("Select coins to view history (space to toggle, enter to view):"))
		b.WriteString("\n\n")

		// List coins
		for i, coin := range m.availableCoins {
			isCursor := i == m.cursor
			isChecked := m.selectedCoins[coin]

			// Checkbox
			checkbox := "[ ] "
			if isChecked {
				checkbox = lipgloss.NewStyle().Foreground(tui.SuccessColor).Render("[✓] ")
			}

			// Cursor and style
			cursor := "  "
			style := lipgloss.NewStyle().Foreground(tui.TextColor)
			if isCursor {
				cursor = lipgloss.NewStyle().Foreground(tui.PrimaryColor).Render("> ")
				style = style.Bold(true).Foreground(tui.PrimaryColor)
			}

			// Show how many data points we have for this coin
			count := m.countDataPoints(coin)
			countText := lipgloss.NewStyle().
				Foreground(tui.MutedColor).
				Render(fmt.Sprintf(" (%d snapshots)", count))

			b.WriteString(cursor + checkbox + style.Render(coin) + countText + "\n")
		}
	}

	// Status message
	if m.statusMsg != "" {
		b.WriteString("\n")
		b.WriteString(components.RenderStatusMessage(m.statusMsg, false))
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(components.RenderHelp(m.coinSelectHelp()))

	return components.RenderBoxDefault(b.String())
}

func (m CoinHistoryModel) countDataPoints(coin string) int {
	count := 0
	for _, snap := range m.store.List() {
		if _, ok := snap.CoinValues[coin]; ok {
			count++
		}
	}
	return count
}

func (m CoinHistoryModel) coinSelectHelp() []components.HelpItem {
	if len(m.availableCoins) > 0 {
		items := []components.HelpItem{
			{Key: "↑↓", Action: "navigate"},
			{Key: "space", Action: "toggle"},
			{Key: "enter", Action: "view/compare"},
		}
		if len(m.selectedCoins) > 0 {
			items = append(items, components.HelpItem{Key: "esc", Action: "clear"})
		} else {
			items = append(items, components.HelpItem{Key: "esc", Action: "back"})
		}
		items = append(items, components.HelpItem{Key: "q", Action: "quit"})
		return items
	}
	return []components.HelpItem{
		{Key: "esc", Action: "back"},
		{Key: "q", Action: "quit"},
	}
}

// GetStore returns the snapshot store
func (m CoinHistoryModel) GetStore() *storage.SnapshotStore {
	return m.store
}

// GetSelectedCoins returns the currently selected coins (for testing)
func (m CoinHistoryModel) GetSelectedCoins() map[string]bool {
	return m.selectedCoins
}
