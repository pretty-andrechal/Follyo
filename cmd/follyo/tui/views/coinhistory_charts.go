package views

import (
	"fmt"
	"math"
	"strings"
	"time"

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

	chart := asciigraph.Plot(prices,
		asciigraph.Height(10),
		asciigraph.Width(chartWidth),
		asciigraph.Caption(fmt.Sprintf("  %s Price (USD)", m.selectedCoin)),
	)

	// Add padding to align with holdings chart
	priceLabelWidth := calculateLabelWidth(prices) + 1 // +1 for label format
	amounts := make([]float64, len(m.coinData))
	for i, dp := range m.coinData {
		amounts[i] = dp.Amount
	}
	amountLabelWidth := calculateLabelWidth(amounts) + 1
	if amountLabelWidth > priceLabelWidth {
		padding := strings.Repeat(" ", amountLabelWidth-priceLabelWidth)
		chart = addPaddingToChart(chart, padding)
	}

	return chart
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

	chart := asciigraph.Plot(amounts,
		asciigraph.Height(8),
		asciigraph.Width(chartWidth),
		asciigraph.LowerBound(0),
		asciigraph.Caption(fmt.Sprintf("  %s Holdings (Amount)", m.selectedCoin)),
	)

	// Add padding to align with price chart
	amountLabelWidth := calculateLabelWidth(amounts) + 1 // +1 for label format
	prices := make([]float64, len(m.coinData))
	for i, dp := range m.coinData {
		prices[i] = dp.Price
	}
	priceLabelWidth := calculateLabelWidth(prices) + 1
	if priceLabelWidth > amountLabelWidth {
		padding := strings.Repeat(" ", priceLabelWidth-amountLabelWidth)
		chart = addPaddingToChart(chart, padding)
	}

	return chart
}

// addPaddingToChart adds leading padding to each line of a chart string
func addPaddingToChart(chart, padding string) string {
	lines := strings.Split(chart, "\n")
	for i, line := range lines {
		lines[i] = padding + line
	}
	return strings.Join(lines, "\n")
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

// calculateAlignedOffset calculates the offset needed for asciigraph to align
// both price and holdings charts with consistent Y-axis positioning
func (m CoinHistoryModel) calculateAlignedOffset() int {
	if len(m.coinData) == 0 {
		return 3 // default asciigraph offset
	}

	// Calculate label width needed for prices
	prices := make([]float64, len(m.coinData))
	for i, dp := range m.coinData {
		prices[i] = dp.Price
	}
	priceLabelWidth := calculateLabelWidth(prices)

	// Calculate label width needed for holdings amounts
	amounts := make([]float64, len(m.coinData))
	for i, dp := range m.coinData {
		amounts[i] = dp.Amount
	}
	amountLabelWidth := calculateLabelWidth(amounts)

	// Use the maximum of both to ensure alignment
	maxLabelWidth := priceLabelWidth
	if amountLabelWidth > maxLabelWidth {
		maxLabelWidth = amountLabelWidth
	}

	// Return offset that accommodates the largest labels
	// asciigraph places labels at position (offset - len(label)), so offset must be >= maxLabelWidth + 1
	return maxLabelWidth + 1
}

// calculateLabelWidth calculates the width of the largest Y-axis label for a given data series
// This mimics asciigraph's internal label formatting logic
func calculateLabelWidth(values []float64) int {
	if len(values) == 0 {
		return 4 // default minimum
	}

	// Find min and max
	minVal, maxVal := values[0], values[0]
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	// asciigraph uses precision=2 by default, but adjusts based on magnitude
	precision := 2
	logMax := math.Log10(math.Max(math.Abs(maxVal), math.Abs(minVal)))
	if minVal == 0 && maxVal == 0 {
		logMax = -1
	}
	if logMax < 0 {
		if math.Mod(logMax, 1) != 0 {
			precision += int(math.Abs(logMax))
		} else {
			precision += int(math.Abs(logMax) - 1)
		}
	} else if logMax > 2 {
		precision = 0
	}

	// Calculate the width of the largest formatted number
	maxNumLen := len(fmt.Sprintf("%0.*f", precision, maxVal))
	minNumLen := len(fmt.Sprintf("%0.*f", precision, minVal))
	if minNumLen > maxNumLen {
		return minNumLen
	}
	return maxNumLen
}

// calculateYAxisWidth calculates the visual position of the Y-axis separator (┤)
// This is used for x-axis alignment and accounts for the padding added to align charts
func (m CoinHistoryModel) calculateYAxisWidth() int {
	if len(m.coinData) == 0 {
		return 7 // fallback default
	}

	// Get label widths for both charts
	prices := make([]float64, len(m.coinData))
	amounts := make([]float64, len(m.coinData))
	for i, dp := range m.coinData {
		prices[i] = dp.Price
		amounts[i] = dp.Amount
	}

	priceLabelWidth := calculateLabelWidth(prices) + 1   // +1 for sprintf format
	amountLabelWidth := calculateLabelWidth(amounts) + 1 // +1 for sprintf format

	// The chart with larger labels determines the y-axis width
	maxLabelWidth := priceLabelWidth
	if amountLabelWidth > maxLabelWidth {
		maxLabelWidth = amountLabelWidth
	}

	// asciigraph layout: [label][space cell][┤ cell][data...]
	// The separator ┤ is at position: label (maxLabelWidth chars) + space (1) = maxLabelWidth + 1
	// X-axis should align with the separator position
	return maxLabelWidth + 1
}

// renderXAxis renders a shared x-axis with date labels for the charts
// It uses adaptive downsampling to show key dates without cluttering
func (m CoinHistoryModel) renderXAxis() string {
	if len(m.coinData) < 2 {
		return ""
	}

	chartWidth := m.calculateChartWidth()

	// Calculate y-axis label width to match asciigraph's rendering
	// asciigraph uses: offset (3) + maxWidth (length of largest formatted number) + 1
	yAxisWidth := m.calculateYAxisWidth()

	// Calculate how many labels can fit
	// Each label needs ~7 chars minimum ("Jan 02" + space)
	labelWidth := 7
	maxLabels := chartWidth / labelWidth
	if maxLabels < 3 {
		maxLabels = 3 // Always show at least first, middle, last
	}
	if maxLabels > 10 {
		maxLabels = 10 // Cap at 10 to avoid clutter
	}

	// Select which data points to show labels for
	dataLen := len(m.coinData)
	labelIndices := selectLabelIndices(dataLen, maxLabels)

	// Find indices where holdings amount changed (for highlighting)
	holdingsChangeIndices := m.findHoldingsChangeIndices()

	// Build the tick marks line and labels line
	var ticksBuilder strings.Builder
	var labelsBuilder strings.Builder

	// Left padding to align with chart
	ticksBuilder.WriteString(strings.Repeat(" ", yAxisWidth))
	labelsBuilder.WriteString(strings.Repeat(" ", yAxisWidth))

	// Track visual position separately (ANSI codes don't count toward visual width)
	visualPos := 0

	for i := 0; i < dataLen; i++ {
		// Calculate x position for this data point
		xPos := 0
		if dataLen > 1 {
			xPos = (i * (chartWidth - 1)) / (dataLen - 1)
		}

		// Check if this index should have a label
		shouldLabel := false
		for _, idx := range labelIndices {
			if idx == i {
				shouldLabel = true
				break
			}
		}

		// Check if this is a holdings change point
		isHoldingsChange := false
		for _, idx := range holdingsChangeIndices {
			if idx == i {
				isHoldingsChange = true
				break
			}
		}

		if shouldLabel {
			// Format the date label
			date := m.coinData[i].Timestamp
			label := formatDateLabel(date)

			// Add highlighting for key dates
			isFirst := i == 0
			isLast := i == dataLen-1

			// Calculate label start position (center label on xPos)
			labelStart := xPos - len(label)/2
			if labelStart < visualPos {
				labelStart = visualPos
			}
			if labelStart+len(label) > chartWidth {
				labelStart = chartWidth - len(label)
			}
			if labelStart < 0 {
				labelStart = 0
			}

			// Add spacing to reach label position
			if labelStart > visualPos {
				labelsBuilder.WriteString(strings.Repeat(" ", labelStart-visualPos))
				visualPos = labelStart
			}

			// Style the label
			style := lipgloss.NewStyle().Foreground(tui.SubtleTextColor)
			if isFirst || isLast {
				style = style.Foreground(tui.AccentColor).Bold(true)
			} else if isHoldingsChange {
				style = style.Foreground(tui.WarningColor)
			}

			labelsBuilder.WriteString(style.Render(label))
			visualPos += len(label)
		}
	}

	// Build tick marks line
	for i := 0; i < chartWidth; i++ {
		// Check if any label index maps to this position
		hasTick := false
		for _, idx := range labelIndices {
			xPos := 0
			if dataLen > 1 {
				xPos = (idx * (chartWidth - 1)) / (dataLen - 1)
			}
			if xPos == i {
				hasTick = true
				break
			}
		}

		if hasTick {
			ticksBuilder.WriteString("│")
		} else {
			ticksBuilder.WriteString("─")
		}
	}

	return ticksBuilder.String() + "\n" + labelsBuilder.String()
}

// selectLabelIndices selects which data point indices should have labels
// Always includes first and last, with evenly distributed points in between
func selectLabelIndices(dataLen, maxLabels int) []int {
	if dataLen <= maxLabels {
		// Show all labels if we have room
		indices := make([]int, dataLen)
		for i := range indices {
			indices[i] = i
		}
		return indices
	}

	// Always include first and last
	indices := make([]int, 0, maxLabels)
	indices = append(indices, 0)

	// Add evenly spaced middle points
	if maxLabels > 2 {
		step := float64(dataLen-1) / float64(maxLabels-1)
		for i := 1; i < maxLabels-1; i++ {
			idx := int(math.Round(float64(i) * step))
			if idx > 0 && idx < dataLen-1 {
				indices = append(indices, idx)
			}
		}
	}

	indices = append(indices, dataLen-1)
	return indices
}

// formatDateLabel formats a timestamp for x-axis display
func formatDateLabel(t time.Time) string {
	now := time.Now()
	daysDiff := int(now.Sub(t).Hours() / 24)

	if daysDiff == 0 {
		return "Today"
	} else if daysDiff == 1 {
		return "Yest"
	} else if t.Year() == now.Year() {
		return t.Format("Jan 2")
	}
	return t.Format("Jan 02")
}

// findHoldingsChangeIndices returns indices where the holdings amount changed
func (m CoinHistoryModel) findHoldingsChangeIndices() []int {
	if len(m.coinData) < 2 {
		return nil
	}

	var indices []int
	prevAmount := m.coinData[0].Amount

	for i := 1; i < len(m.coinData); i++ {
		if m.coinData[i].Amount != prevAmount {
			indices = append(indices, i)
			prevAmount = m.coinData[i].Amount
		}
	}

	return indices
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

