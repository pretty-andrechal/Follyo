package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/guptarohit/asciigraph"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
)

// Chart rendering methods for CoinHistoryModel

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

	chartWidth := m.calculateChartWidth()

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

	chartWidth := m.calculateChartWidth()

	return asciigraph.Plot(amounts,
		asciigraph.Height(8),
		asciigraph.Width(chartWidth),
		asciigraph.LowerBound(0),
		asciigraph.Caption(fmt.Sprintf("  %s Holdings (Amount)", m.selectedCoin)),
	)
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

	// Build normalized series (% change from start, indexed to 100)
	var series [][]float64
	var legends []string
	maxLen := 0

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
		if len(normalized) > maxLen {
			maxLen = len(normalized)
		}
	}

	if len(series) == 0 {
		return lipgloss.NewStyle().
			Foreground(tui.SubtleTextColor).
			Render("  (Not enough data for comparison)")
	}

	// Pad shorter series to match the longest series length
	// asciigraph.PlotMany with SeriesLegends requires equal-length series
	for i := range series {
		if len(series[i]) < maxLen {
			// Pad with the last value (carry forward)
			lastVal := series[i][len(series[i])-1]
			for len(series[i]) < maxLen {
				series[i] = append(series[i], lastVal)
			}
		}
	}

	chartWidth := m.calculateChartWidth()

	// Define colors for each series (asciigraph requires SeriesColors when using SeriesLegends)
	colors := []asciigraph.AnsiColor{
		asciigraph.Green,
		asciigraph.Blue,
		asciigraph.Yellow,
		asciigraph.Magenta,
		asciigraph.Cyan,
		asciigraph.Red,
		asciigraph.White,
	}
	// Create color slice matching series length
	seriesColors := make([]asciigraph.AnsiColor, len(series))
	for i := range series {
		seriesColors[i] = colors[i%len(colors)]
	}

	return asciigraph.PlotMany(series,
		asciigraph.Height(12),
		asciigraph.Width(chartWidth),
		asciigraph.SeriesColors(seriesColors...),
		asciigraph.SeriesLegends(legends...),
		asciigraph.Caption("  Price indexed to 100 at start"),
	)
}

func (m CoinHistoryModel) renderStackedHoldingsCharts() string {
	var b strings.Builder

	chartWidth := m.calculateChartWidth()

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

// calculateChartWidth returns the appropriate chart width based on terminal width
func (m CoinHistoryModel) calculateChartWidth() int {
	chartWidth := m.width - 20
	if chartWidth > 60 {
		chartWidth = 60
	}
	if chartWidth < 20 {
		chartWidth = 20
	}
	return chartWidth
}

// hasVaryingAmounts checks if holdings amounts vary across data points
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

// calculatePriceStats returns min, max, and average price from coin data
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

// calculatePriceChange returns the absolute and percentage price change
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

