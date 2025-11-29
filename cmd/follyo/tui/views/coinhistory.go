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
	"github.com/guptarohit/asciigraph"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui/components"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui/format"
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
	selectedCoin   string                       // For single coin view
	selectedCoins  map[string]bool              // For multi-select
	compareCoins   []string                     // Ordered list of coins to compare
	coinData       []CoinDataPoint              // Single coin data
	coinDataMap    map[string][]CoinDataPoint   // Multi-coin data
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

func (m CoinHistoryModel) renderDisplay() string {
	// Header
	title := components.RenderTitle(fmt.Sprintf("%s HISTORY", m.selectedCoin))

	// Scroll indicator
	scrollInfo := ""
	if m.viewportReady {
		scrollPercent := m.viewport.ScrollPercent() * 100
		if m.viewport.TotalLineCount() > m.viewport.Height {
			scrollInfo = lipgloss.NewStyle().
				Foreground(tui.MutedColor).
				Render(fmt.Sprintf(" (%.0f%%)", scrollPercent))
		}
	}

	header := lipgloss.JoinHorizontal(lipgloss.Center, title, scrollInfo)

	// Footer with help
	footer := components.RenderHelp([]components.HelpItem{
		{Key: "↑↓", Action: "scroll"},
		{Key: "esc", Action: "back"},
		{Key: "q", Action: "quit"},
	})

	// Viewport with border
	viewportStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.BorderColor).
		Padding(0, 1)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		viewportStyle.Render(m.viewport.View()),
		footer,
	)
}

func (m CoinHistoryModel) renderCompare() string {
	// Header
	coinList := strings.Join(m.compareCoins, ", ")
	title := components.RenderTitle("COMPARE: " + coinList)

	// Scroll indicator
	scrollInfo := ""
	if m.viewportReady {
		scrollPercent := m.viewport.ScrollPercent() * 100
		if m.viewport.TotalLineCount() > m.viewport.Height {
			scrollInfo = lipgloss.NewStyle().
				Foreground(tui.MutedColor).
				Render(fmt.Sprintf(" (%.0f%%)", scrollPercent))
		}
	}

	header := lipgloss.JoinHorizontal(lipgloss.Center, title, scrollInfo)

	// Footer with help
	footer := components.RenderHelp([]components.HelpItem{
		{Key: "↑↓", Action: "scroll"},
		{Key: "esc", Action: "back"},
		{Key: "q", Action: "quit"},
	})

	// Viewport with border
	viewportStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.BorderColor).
		Padding(0, 1)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		viewportStyle.Render(m.viewport.View()),
		footer,
	)
}

func (m CoinHistoryModel) renderCompareContent() string {
	var b strings.Builder

	if len(m.compareCoins) == 0 {
		b.WriteString(components.RenderEmptyState("No coins selected for comparison."))
		return b.String()
	}

	sectionStyle := lipgloss.NewStyle().
		Foreground(tui.AccentColor).
		Bold(true)

	// Section: Price Comparison Chart (normalized)
	b.WriteString(sectionStyle.Render("PRICE COMPARISON (% Change from Start)"))
	b.WriteString("\n\n")
	b.WriteString(m.renderNormalizedPriceChart())
	b.WriteString("\n\n")

	// Section: Holdings Charts (stacked, each with own scale)
	b.WriteString(sectionStyle.Render("HOLDINGS CHARTS"))
	b.WriteString("\n\n")
	b.WriteString(m.renderStackedHoldingsCharts())
	b.WriteString("\n\n")

	// Section: Comparison Summary Table
	b.WriteString(sectionStyle.Render("COMPARISON SUMMARY"))
	b.WriteString("\n\n")
	b.WriteString(m.renderComparisonTable())
	b.WriteString("\n\n")

	// Section: Combined Historical Data
	b.WriteString(sectionStyle.Render("COMBINED HISTORICAL DATA"))
	b.WriteString("\n\n")
	b.WriteString(m.renderCombinedDataTable())

	return b.String()
}

func (m CoinHistoryModel) renderNormalizedPriceChart() string {
	// Check if we have enough data
	hasEnoughData := false
	for _, coin := range m.compareCoins {
		if len(m.coinDataMap[coin]) >= 2 {
			hasEnoughData = true
			break
		}
	}

	if !hasEnoughData {
		return lipgloss.NewStyle().
			Foreground(tui.SubtleTextColor).
			Render("  (Need at least 2 data points for chart)")
	}

	// Find the maximum number of data points across all coins
	maxPoints := 0
	for _, coin := range m.compareCoins {
		if len(m.coinDataMap[coin]) > maxPoints {
			maxPoints = len(m.coinDataMap[coin])
		}
	}

	// Build normalized series (% change from start, indexed to 100)
	var series [][]float64
	var legends []string

	for _, coin := range m.compareCoins {
		data := m.coinDataMap[coin]
		if len(data) < 2 {
			continue
		}

		normalized := make([]float64, len(data))
		startPrice := data[0].Price
		for i, dp := range data {
			if startPrice != 0 {
				normalized[i] = (dp.Price / startPrice) * 100
			} else {
				normalized[i] = 100
			}
		}

		series = append(series, normalized)
		legends = append(legends, coin)
	}

	if len(series) == 0 {
		return lipgloss.NewStyle().
			Foreground(tui.SubtleTextColor).
			Render("  (Not enough data for comparison)")
	}

	// Calculate chart width
	chartWidth := m.width - 20
	if chartWidth > 60 {
		chartWidth = 60
	}
	if chartWidth < 20 {
		chartWidth = 20
	}

	return asciigraph.PlotMany(series,
		asciigraph.Height(12),
		asciigraph.Width(chartWidth),
		asciigraph.SeriesLegends(legends...),
		asciigraph.Caption("  Price indexed to 100 at start"),
	)
}

func (m CoinHistoryModel) renderStackedHoldingsCharts() string {
	var b strings.Builder

	chartWidth := m.width - 20
	if chartWidth > 60 {
		chartWidth = 60
	}
	if chartWidth < 20 {
		chartWidth = 20
	}

	for i, coin := range m.compareCoins {
		data := m.coinDataMap[coin]

		if i > 0 {
			b.WriteString("\n")
		}

		coinLabel := lipgloss.NewStyle().
			Foreground(tui.TextColor).
			Bold(true).
			Render(coin)
		b.WriteString(coinLabel)
		b.WriteString("\n")

		if len(data) < 2 {
			b.WriteString(lipgloss.NewStyle().
				Foreground(tui.SubtleTextColor).
				Render("  (Need at least 2 data points)"))
			b.WriteString("\n")
			continue
		}

		amounts := make([]float64, len(data))
		for j, dp := range data {
			amounts[j] = dp.Amount
		}

		chart := asciigraph.Plot(amounts,
			asciigraph.Height(5),
			asciigraph.Width(chartWidth),
			asciigraph.LowerBound(0),
		)
		b.WriteString(chart)
		b.WriteString("\n")
	}

	return b.String()
}

func (m CoinHistoryModel) renderComparisonTable() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Foreground(tui.MutedColor).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(tui.SubtleTextColor).Width(14)
	valueStyle := lipgloss.NewStyle().Foreground(tui.TextColor).Width(14)

	// Header row
	header := fmt.Sprintf("%-14s", "Metric")
	for _, coin := range m.compareCoins {
		header += fmt.Sprintf("  %12s", coin)
	}
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(components.RenderSeparator(14 + len(m.compareCoins)*14))
	b.WriteString("\n")

	// Data Points row
	b.WriteString(labelStyle.Render("Data Points"))
	for _, coin := range m.compareCoins {
		count := len(m.coinDataMap[coin])
		b.WriteString(valueStyle.Render(fmt.Sprintf("  %12d", count)))
	}
	b.WriteString("\n")

	// Current Holdings row
	b.WriteString(labelStyle.Render("Holdings"))
	for _, coin := range m.compareCoins {
		data := m.coinDataMap[coin]
		if len(data) > 0 {
			b.WriteString(valueStyle.Render(fmt.Sprintf("  %12.4f", data[len(data)-1].Amount)))
		} else {
			b.WriteString(valueStyle.Render(fmt.Sprintf("  %12s", "-")))
		}
	}
	b.WriteString("\n")

	// Current Price row
	b.WriteString(labelStyle.Render("Current Price"))
	for _, coin := range m.compareCoins {
		data := m.coinDataMap[coin]
		if len(data) > 0 {
			b.WriteString(valueStyle.Render(fmt.Sprintf("  %12s", format.USDSimple(data[len(data)-1].Price))))
		} else {
			b.WriteString(valueStyle.Render(fmt.Sprintf("  %12s", "-")))
		}
	}
	b.WriteString("\n")

	// Current Value row
	b.WriteString(labelStyle.Render("Current Value"))
	for _, coin := range m.compareCoins {
		data := m.coinDataMap[coin]
		if len(data) > 0 {
			b.WriteString(valueStyle.Render(fmt.Sprintf("  %12s", format.USDSimple(data[len(data)-1].Value))))
		} else {
			b.WriteString(valueStyle.Render(fmt.Sprintf("  %12s", "-")))
		}
	}
	b.WriteString("\n")

	// Price Change row
	b.WriteString(labelStyle.Render("Price Change"))
	for _, coin := range m.compareCoins {
		data := m.coinDataMap[coin]
		if len(data) >= 2 {
			first := data[0].Price
			last := data[len(data)-1].Price
			var pct float64
			if first != 0 {
				pct = ((last - first) / first) * 100
			}
			style := valueStyle
			if pct > 0 {
				style = style.Foreground(tui.SuccessColor)
			} else if pct < 0 {
				style = style.Foreground(tui.ErrorColor)
			}
			b.WriteString(style.Render(fmt.Sprintf("  %+11.1f%%", pct)))
		} else {
			b.WriteString(valueStyle.Render(fmt.Sprintf("  %12s", "-")))
		}
	}
	b.WriteString("\n")

	// Total Value row
	b.WriteString("\n")
	totalValue := 0.0
	for _, coin := range m.compareCoins {
		data := m.coinDataMap[coin]
		if len(data) > 0 {
			totalValue += data[len(data)-1].Value
		}
	}
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(tui.TextColor).Render(
		fmt.Sprintf("Total Value: %s", format.USDSimple(totalValue))))

	return b.String()
}

func (m CoinHistoryModel) renderCombinedDataTable() string {
	var b strings.Builder

	// Get all unique timestamps across all coins
	timestampSet := make(map[time.Time]bool)
	for _, coin := range m.compareCoins {
		for _, dp := range m.coinDataMap[coin] {
			timestampSet[dp.Timestamp] = true
		}
	}

	// Sort timestamps (newest first for display)
	timestamps := make([]time.Time, 0, len(timestampSet))
	for ts := range timestampSet {
		timestamps = append(timestamps, ts)
	}
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i].After(timestamps[j])
	})

	// Build lookup maps for quick access
	coinDataByTime := make(map[string]map[time.Time]CoinDataPoint)
	for _, coin := range m.compareCoins {
		coinDataByTime[coin] = make(map[time.Time]CoinDataPoint)
		for _, dp := range m.coinDataMap[coin] {
			coinDataByTime[coin][dp.Timestamp] = dp
		}
	}

	// Header
	headerStyle := lipgloss.NewStyle().Foreground(tui.MutedColor).Bold(true)
	header := fmt.Sprintf("%-17s", "Date")
	for _, coin := range m.compareCoins {
		header += fmt.Sprintf("  %10s  %10s", coin, "Value")
	}
	header += fmt.Sprintf("  %12s", "Total")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	sepWidth := 17 + len(m.compareCoins)*24 + 14
	b.WriteString(components.RenderSeparator(sepWidth))
	b.WriteString("\n")

	// Limit to most recent 20 entries to avoid overwhelming display
	displayCount := len(timestamps)
	if displayCount > 20 {
		displayCount = 20
	}

	for i := 0; i < displayCount; i++ {
		ts := timestamps[i]
		row := fmt.Sprintf("%-17s", ts.Format("2006-01-02 15:04"))

		totalValue := 0.0
		for _, coin := range m.compareCoins {
			if dp, ok := coinDataByTime[coin][ts]; ok {
				row += fmt.Sprintf("  %10.4f  %10s", dp.Amount, format.USDSimple(dp.Value))
				totalValue += dp.Value
			} else {
				row += fmt.Sprintf("  %10s  %10s", "-", "-")
			}
		}
		row += fmt.Sprintf("  %12s", format.USDSimple(totalValue))

		b.WriteString(lipgloss.NewStyle().Foreground(tui.TextColor).Render(row))
		b.WriteString("\n")
	}

	if len(timestamps) > 20 {
		b.WriteString(lipgloss.NewStyle().Foreground(tui.MutedColor).Render(
			fmt.Sprintf("\n  ... and %d more entries", len(timestamps)-20)))
	}

	return b.String()
}

func (m CoinHistoryModel) renderDisplayContent() string {
	var b strings.Builder

	if len(m.coinData) == 0 {
		b.WriteString(components.RenderEmptyState("No history data available for this coin."))
		return b.String()
	}

	// Section: Price Chart
	sectionStyle := lipgloss.NewStyle().
		Foreground(tui.AccentColor).
		Bold(true)

	b.WriteString(sectionStyle.Render("PRICE CHART"))
	b.WriteString("\n\n")

	// Render ASCII chart
	chart := m.renderPriceChart()
	b.WriteString(chart)
	b.WriteString("\n\n")

	// Section: Holdings Chart
	b.WriteString(sectionStyle.Render("HOLDINGS CHART"))
	b.WriteString("\n\n")
	holdingsChart := m.renderHoldingsChart()
	b.WriteString(holdingsChart)
	b.WriteString("\n\n")

	// Section: Data Table
	b.WriteString(sectionStyle.Render("HISTORICAL DATA"))
	b.WriteString("\n\n")

	// Table header
	headerStyle := lipgloss.NewStyle().Foreground(tui.MutedColor).Bold(true)
	header := fmt.Sprintf("  %-19s  %14s  %14s  %14s",
		"Date", "Price", "Amount", "Value")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// Separator
	b.WriteString(components.RenderSeparator(70))
	b.WriteString("\n")

	// Table rows (newest first for display)
	for i := len(m.coinData) - 1; i >= 0; i-- {
		dp := m.coinData[i]
		row := fmt.Sprintf("  %-19s  %14s  %14.6f  %14s",
			dp.Timestamp.Format("2006-01-02 15:04"),
			format.USDSimple(dp.Price),
			dp.Amount,
			format.USDSimple(dp.Value))
		b.WriteString(lipgloss.NewStyle().Foreground(tui.TextColor).Render(row))
		b.WriteString("\n")
	}

	// Summary statistics
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("SUMMARY"))
	b.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().
		Foreground(tui.SubtleTextColor).
		Width(18)
	valueStyle := lipgloss.NewStyle().
		Foreground(tui.TextColor)

	// Calculate stats
	minPrice, maxPrice, avgPrice := m.calculatePriceStats()
	priceChange, priceChangePercent := m.calculatePriceChange()

	summaryItems := []struct {
		label string
		value string
		style lipgloss.Style
	}{
		{"Data Points:", fmt.Sprintf("%d snapshots", len(m.coinData)), valueStyle},
		{"Min Price:", format.USDSimple(minPrice), valueStyle},
		{"Max Price:", format.USDSimple(maxPrice), valueStyle},
		{"Avg Price:", format.USDSimple(avgPrice), valueStyle},
	}

	for _, item := range summaryItems {
		b.WriteString(labelStyle.Render(item.label))
		b.WriteString(item.style.Render(item.value))
		b.WriteString("\n")
	}

	// Price change with color
	priceChangeStyle := valueStyle
	priceChangeText := format.ProfitLoss(priceChange, priceChangePercent)
	if priceChange > 0 {
		priceChangeStyle = valueStyle.Foreground(tui.SuccessColor)
	} else if priceChange < 0 {
		priceChangeStyle = valueStyle.Foreground(tui.ErrorColor)
	}
	b.WriteString(labelStyle.Render("Price Change:"))
	b.WriteString(priceChangeStyle.Render(priceChangeText))
	b.WriteString("\n")

	return b.String()
}

func (m CoinHistoryModel) renderPriceChart() string {
	if len(m.coinData) < 2 {
		return lipgloss.NewStyle().
			Foreground(tui.SubtleTextColor).
			Render("  (Need at least 2 data points for chart)")
	}

	prices := make([]float64, len(m.coinData))
	for i, dp := range m.coinData {
		prices[i] = dp.Price
	}

	// Calculate chart width (leave room for axis labels)
	chartWidth := m.width - 20
	if chartWidth > 60 {
		chartWidth = 60
	}
	if chartWidth < 20 {
		chartWidth = 20
	}

	return asciigraph.Plot(prices,
		asciigraph.Height(10),
		asciigraph.Width(chartWidth),
		asciigraph.Caption(fmt.Sprintf("  %s Price (USD)", m.selectedCoin)),
	)
}

func (m CoinHistoryModel) renderHoldingsChart() string {
	if len(m.coinData) < 2 {
		return lipgloss.NewStyle().
			Foreground(tui.SubtleTextColor).
			Render("  (Need at least 2 data points for chart)")
	}

	amounts := make([]float64, len(m.coinData))
	for i, dp := range m.coinData {
		amounts[i] = dp.Amount
	}

	// Calculate chart width
	chartWidth := m.width - 20
	if chartWidth > 60 {
		chartWidth = 60
	}
	if chartWidth < 20 {
		chartWidth = 20
	}

	return asciigraph.Plot(amounts,
		asciigraph.Height(8),
		asciigraph.Width(chartWidth),
		asciigraph.LowerBound(0),
		asciigraph.Caption(fmt.Sprintf("  %s Holdings (Amount)", m.selectedCoin)),
	)
}

func (m CoinHistoryModel) hasVaryingAmounts() bool {
	if len(m.coinData) < 2 {
		return false
	}
	firstAmount := m.coinData[0].Amount
	for _, dp := range m.coinData[1:] {
		if dp.Amount != firstAmount {
			return true
		}
	}
	return false
}

func (m CoinHistoryModel) calculatePriceStats() (min, max, avg float64) {
	if len(m.coinData) == 0 {
		return 0, 0, 0
	}

	min = m.coinData[0].Price
	max = m.coinData[0].Price
	sum := 0.0

	for _, dp := range m.coinData {
		if dp.Price < min {
			min = dp.Price
		}
		if dp.Price > max {
			max = dp.Price
		}
		sum += dp.Price
	}

	avg = sum / float64(len(m.coinData))
	return min, max, avg
}

func (m CoinHistoryModel) calculatePriceChange() (change, percent float64) {
	if len(m.coinData) < 2 {
		return 0, 0
	}

	first := m.coinData[0].Price                // Oldest
	last := m.coinData[len(m.coinData)-1].Price // Newest
	change = last - first
	if first != 0 {
		percent = (change / first) * 100
	}
	return change, percent
}

// GetStore returns the snapshot store
func (m CoinHistoryModel) GetStore() *storage.SnapshotStore {
	return m.store
}

// GetSelectedCoins returns the currently selected coins (for testing)
func (m CoinHistoryModel) GetSelectedCoins() map[string]bool {
	return m.selectedCoins
}
