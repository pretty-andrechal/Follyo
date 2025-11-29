package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui/components"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui/format"
)

// Table and content rendering methods for CoinHistoryModel

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
