package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui/format"
	"github.com/pretty-andrechal/follyo/internal/portfolio"
	"github.com/pretty-andrechal/follyo/internal/prices"
)

// SummaryModel represents the portfolio summary view.
type SummaryModel struct {
	portfolio       portfolio.SummaryProvider
	summary         *portfolio.Summary
	livePrices      map[string]float64
	unmappedTickers []string
	isOffline       bool
	loading         bool
	spinner         spinner.Model
	viewport        viewport.Model
	viewportReady   bool
	keys            tui.KeyMap
	width           int
	height          int
	lastUpdated     time.Time
	err             error
	tickerMappings  map[string]string
}

// NewSummaryModel creates a new summary view model.
func NewSummaryModel(p portfolio.SummaryProvider, tickerMappings map[string]string) SummaryModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(tui.PrimaryColor)

	return SummaryModel{
		portfolio:      p,
		spinner:        s,
		keys:           tui.DefaultKeyMap(),
		loading:        true,
		tickerMappings: tickerMappings,
	}
}

// Init initializes the summary model.
func (m SummaryModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadData())
}

// SummaryDataMsg is sent when portfolio data is loaded.
type SummaryDataMsg struct {
	Summary         *portfolio.Summary
	Prices          map[string]float64
	UnmappedTickers []string
	IsOffline       bool
	Error           error
}

func (m SummaryModel) loadData() tea.Cmd {
	return func() tea.Msg {
		// Get portfolio summary
		summary, err := m.portfolio.GetSummary()
		if err != nil {
			return SummaryDataMsg{Error: err}
		}

		// Collect all coins
		allCoins := collectCoins(summary)

		// Fetch prices if there are coins
		var livePrices map[string]float64
		var unmappedTickers []string
		var isOffline bool

		if len(allCoins) > 0 {
			ps := prices.New()

			// Add custom ticker mappings
			for ticker, geckoID := range m.tickerMappings {
				ps.AddCoinMapping(ticker, geckoID)
			}

			unmappedTickers = ps.GetUnmappedTickers(allCoins)

			priceMap, err := ps.GetPrices(allCoins)
			if err != nil {
				isOffline = true
			} else {
				livePrices = priceMap
			}
		}

		return SummaryDataMsg{
			Summary:         &summary,
			Prices:          livePrices,
			UnmappedTickers: unmappedTickers,
			IsOffline:       isOffline,
		}
	}
}

// Update handles messages for the summary model.
func (m SummaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back), key.Matches(msg, m.keys.Quit):
			if msg.String() == "q" {
				return m, tea.Quit
			}
			return m, func() tea.Msg { return tui.BackToMenuMsg{} }

		case key.Matches(msg, m.keys.Refresh):
			m.loading = true
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, m.loadData())
		}

		// Forward other keys to viewport for scrolling
		if m.viewportReady && !m.loading {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Reserve space for header and footer
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

		// Update content if we have data
		if m.summary != nil && !m.loading {
			m.viewport.SetContent(m.renderContent())
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case SummaryDataMsg:
		m.loading = false
		m.lastUpdated = time.Now()
		if msg.Error != nil {
			m.err = msg.Error
		} else {
			m.summary = msg.Summary
			m.livePrices = msg.Prices
			m.unmappedTickers = msg.UnmappedTickers
			m.isOffline = msg.IsOffline

			// Update viewport content
			if m.viewportReady {
				m.viewport.SetContent(m.renderContent())
				m.viewport.GotoTop()
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the summary view.
func (m SummaryModel) View() string {
	if m.loading {
		return m.renderLoading()
	}

	if m.err != nil {
		return m.renderError()
	}

	// If viewport isn't ready yet, render content directly
	if !m.viewportReady {
		return m.renderContent()
	}

	return m.renderWithViewport()
}

func (m SummaryModel) renderLoading() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.PrimaryColor).
		Padding(2, 4)

	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Render("PORTFOLIO SUMMARY")

	loadingText := lipgloss.NewStyle().
		Foreground(tui.SubtleTextColor).
		Render(fmt.Sprintf("\n\n%s Fetching portfolio data and live prices...", m.spinner.View()))

	return boxStyle.Render(title + loadingText)
}

func (m SummaryModel) renderError() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.ErrorColor).
		Padding(2, 4)

	title := lipgloss.NewStyle().
		Foreground(tui.ErrorColor).
		Bold(true).
		Render("Error Loading Portfolio")

	msg := lipgloss.NewStyle().
		Foreground(tui.SubtleTextColor).
		Render(fmt.Sprintf("\n\n%v\n\nPress R to retry or ESC to go back.", m.err))

	return boxStyle.Render(title + msg)
}

func (m SummaryModel) renderWithViewport() string {
	// Header
	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Render("PORTFOLIO SUMMARY")

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
	var footerParts []string

	if !m.lastUpdated.IsZero() {
		timeStr := lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Render(fmt.Sprintf("Updated: %s", m.lastUpdated.Format("15:04:05")))
		footerParts = append(footerParts, timeStr)
	}

	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)
	help := fmt.Sprintf("%s scroll  %s refresh  %s back  %s quit",
		tui.HelpKeyStyle.Render("↑↓"),
		tui.HelpKeyStyle.Render("r"),
		tui.HelpKeyStyle.Render("esc"),
		tui.HelpKeyStyle.Render("q"))
	footerParts = append(footerParts, helpStyle.Render(help))

	footer := lipgloss.JoinHorizontal(lipgloss.Center, strings.Join(footerParts, "  |  "))

	// Combine with border
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

func (m SummaryModel) renderContent() string {
	var b strings.Builder

	// Calculate totals for value display
	var totalCurrentValue, totalLoanValue float64

	// Holdings section
	b.WriteString(m.renderTable("HOLDINGS BY COIN", m.summary.HoldingsByCoin, false, &totalCurrentValue))

	// Staked section
	b.WriteString(m.renderTable("STAKED BY COIN", m.summary.StakesByCoin, false, nil))

	// Available section
	b.WriteString(m.renderTable("AVAILABLE (Holdings - Staked)", m.summary.AvailableByCoin, false, nil))

	// Loans section
	b.WriteString(m.renderTable("LOANS BY COIN", m.summary.LoansByCoin, false, &totalLoanValue))

	// Net holdings section
	b.WriteString(m.renderTable("NET HOLDINGS (Holdings - Loans)", m.summary.NetByCoin, true, nil))

	// Statistics
	b.WriteString(m.renderStats())

	// Value summary
	if m.livePrices != nil && totalCurrentValue > 0 {
		b.WriteString(m.renderValueSummary(totalCurrentValue, totalLoanValue))
	} else if m.isOffline {
		offlineMsg := lipgloss.NewStyle().
			Foreground(tui.WarningColor).
			Render("\n(Offline - prices unavailable)")
		b.WriteString(offlineMsg)
	}

	// Unmapped tickers warning
	if len(m.unmappedTickers) > 0 {
		warning := lipgloss.NewStyle().
			Foreground(tui.WarningColor).
			Render(fmt.Sprintf("\n\nNo CoinGecko mapping for: %s\nRun 'follyo ticker search <query> <TICKER>' to add a mapping", strings.Join(m.unmappedTickers, ", ")))
		b.WriteString(warning)
	}

	return b.String()
}

func (m SummaryModel) renderTable(title string, data map[string]float64, showPrefix bool, accumulator *float64) string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(tui.AccentColor).
		Bold(true)

	b.WriteString(headerStyle.Render(title))
	b.WriteString("\n")

	if len(data) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			PaddingLeft(2)
		b.WriteString(emptyStyle.Render("(none)"))
		b.WriteString("\n\n")
		return b.String()
	}

	// Sort coins
	coins := make([]string, 0, len(data))
	for coin := range data {
		coins = append(coins, coin)
	}
	sort.Strings(coins)

	// Column widths for alignment
	const (
		coinWidth   = 8
		amountWidth = 14
		priceWidth  = 12
		valueWidth  = 14
	)

	// Table header
	if m.livePrices != nil {
		headerRow := fmt.Sprintf("  %-*s  %*s  %*s  %*s",
			coinWidth, "Coin",
			amountWidth, "Amount",
			priceWidth, "Price",
			valueWidth, "Value")
		b.WriteString(lipgloss.NewStyle().Foreground(tui.MutedColor).Render(headerRow))
		b.WriteString("\n")

		// Separator
		sepLen := 2 + coinWidth + 2 + amountWidth + 2 + priceWidth + 2 + valueWidth
		b.WriteString(lipgloss.NewStyle().Foreground(tui.BorderColor).Render(strings.Repeat("─", sepLen)))
		b.WriteString("\n")
	}

	// Data rows
	for _, coin := range coins {
		amount := data[coin]
		b.WriteString(m.renderTableRow(coin, amount, showPrefix, accumulator, coinWidth, amountWidth, priceWidth, valueWidth))
	}
	b.WriteString("\n")

	return b.String()
}

func (m SummaryModel) renderTableRow(coin string, amount float64, showPrefix bool, accumulator *float64, coinWidth, amountWidth, priceWidth, valueWidth int) string {
	coinStyle := lipgloss.NewStyle().Foreground(tui.TextColor)

	// Format amount with prefix and proper alignment
	prefix := ""
	if showPrefix && amount > 0 {
		prefix = "+"
	} else if showPrefix && amount < 0 {
		prefix = "" // negative sign is already included
	}

	amountStr := fmt.Sprintf("%s%.*f", prefix, decimalPlaces(amount), amount)

	if m.livePrices != nil {
		if price, ok := m.livePrices[coin]; ok {
			value := amount * price
			if accumulator != nil {
				*accumulator += value
			}

			// Color the value based on sign
			valueStyle := lipgloss.NewStyle().Foreground(tui.TextColor)
			if showPrefix {
				if value > 0 {
					valueStyle = valueStyle.Foreground(tui.SuccessColor)
				} else if value < 0 {
					valueStyle = valueStyle.Foreground(tui.ErrorColor)
				}
			}

			valuePrefix := ""
			if showPrefix && value > 0 {
				valuePrefix = "+"
			}

			return fmt.Sprintf("  %s  %s  %s  %s\n",
				coinStyle.Width(coinWidth).Render(coin),
				lipgloss.NewStyle().Foreground(tui.TextColor).Width(amountWidth).Align(lipgloss.Right).Render(amountStr),
				lipgloss.NewStyle().Foreground(tui.SubtleTextColor).Width(priceWidth).Align(lipgloss.Right).Render(format.USDSimple(price)),
				valueStyle.Width(valueWidth).Align(lipgloss.Right).Render(valuePrefix+format.USDSimple(value)))
		}
		// No price available
		return fmt.Sprintf("  %s  %s  %s  %s\n",
			coinStyle.Width(coinWidth).Render(coin),
			lipgloss.NewStyle().Foreground(tui.TextColor).Width(amountWidth).Align(lipgloss.Right).Render(amountStr),
			lipgloss.NewStyle().Foreground(tui.MutedColor).Width(priceWidth).Align(lipgloss.Right).Render("N/A"),
			lipgloss.NewStyle().Foreground(tui.MutedColor).Width(valueWidth).Align(lipgloss.Right).Render("N/A"))
	}

	return fmt.Sprintf("  %s  %s\n",
		coinStyle.Width(coinWidth).Render(coin),
		lipgloss.NewStyle().Foreground(tui.TextColor).Width(amountWidth).Align(lipgloss.Right).Render(amountStr))
}

func (m SummaryModel) renderStats() string {
	var b strings.Builder

	divider := lipgloss.NewStyle().
		Foreground(tui.BorderColor).
		Render(strings.Repeat("─", 40))
	b.WriteString(divider)
	b.WriteString("\n")

	labelStyle := lipgloss.NewStyle().
		Foreground(tui.SubtleTextColor).
		Width(18)
	valueStyle := lipgloss.NewStyle().
		Foreground(tui.TextColor).
		Width(14).
		Align(lipgloss.Right)

	stats := []struct {
		label string
		value string
	}{
		{"Holdings:", fmt.Sprintf("%d", m.summary.TotalHoldingsCount)},
		{"Sales:", fmt.Sprintf("%d", m.summary.TotalSalesCount)},
		{"Stakes:", fmt.Sprintf("%d", m.summary.TotalStakesCount)},
		{"Loans:", fmt.Sprintf("%d", m.summary.TotalLoansCount)},
		{"Total Invested:", format.USDSimple(m.summary.TotalInvestedUSD)},
		{"Total Sold:", format.USDSimple(m.summary.TotalSoldUSD)},
	}

	for _, stat := range stats {
		b.WriteString(labelStyle.Render(stat.label))
		b.WriteString(valueStyle.Render(stat.value))
		b.WriteString("\n")
	}

	return b.String()
}

func (m SummaryModel) renderValueSummary(totalCurrentValue, totalLoanValue float64) string {
	var b strings.Builder

	divider := lipgloss.NewStyle().
		Foreground(tui.BorderColor).
		Render(strings.Repeat("─", 40))
	b.WriteString("\n")
	b.WriteString(divider)
	b.WriteString("\n")

	labelStyle := lipgloss.NewStyle().
		Foreground(tui.SubtleTextColor).
		Width(18)
	valueWidth := 14

	b.WriteString(labelStyle.Render("Holdings Value:"))
	b.WriteString(lipgloss.NewStyle().Foreground(tui.SuccessColor).Width(valueWidth).Align(lipgloss.Right).Render(format.USDSimple(totalCurrentValue)))
	b.WriteString("\n")

	if totalLoanValue > 0 {
		b.WriteString(labelStyle.Render("Loans Value:"))
		b.WriteString(lipgloss.NewStyle().Foreground(tui.ErrorColor).Width(valueWidth).Align(lipgloss.Right).Render("-"+format.USDSimple(totalLoanValue)))
		b.WriteString("\n")
	}

	netValue := totalCurrentValue - totalLoanValue
	b.WriteString(labelStyle.Render("Net Value:"))
	b.WriteString(lipgloss.NewStyle().Foreground(tui.TextColor).Bold(true).Width(valueWidth).Align(lipgloss.Right).Render(format.USDSimple(netValue)))
	b.WriteString("\n")

	// Profit/Loss calculation
	profitLoss := netValue - m.summary.TotalInvestedUSD + m.summary.TotalSoldUSD
	profitLossPercent := format.SafeDivide(profitLoss, m.summary.TotalInvestedUSD) * 100

	plText := format.ProfitLoss(profitLoss, profitLossPercent)

	b.WriteString(labelStyle.Render("Profit/Loss:"))
	plStyle := lipgloss.NewStyle().Width(valueWidth + 10).Align(lipgloss.Right)
	if profitLoss > 0 {
		plStyle = plStyle.Foreground(tui.SuccessColor)
	} else if profitLoss < 0 {
		plStyle = plStyle.Foreground(tui.ErrorColor)
	} else {
		plStyle = plStyle.Foreground(tui.TextColor)
	}
	b.WriteString(plStyle.Render(plText))

	return b.String()
}

// Helper functions

func collectCoins(summary portfolio.Summary) []string {
	allCoins := make(map[string]bool)
	for coin := range summary.HoldingsByCoin {
		allCoins[coin] = true
	}
	for coin := range summary.StakesByCoin {
		allCoins[coin] = true
	}
	for coin := range summary.LoansByCoin {
		allCoins[coin] = true
	}
	for coin := range summary.NetByCoin {
		allCoins[coin] = true
	}

	coins := make([]string, 0, len(allCoins))
	for coin := range allCoins {
		coins = append(coins, coin)
	}
	sort.Strings(coins)
	return coins
}

// decimalPlaces returns appropriate decimal places based on the value
func decimalPlaces(amount float64) int {
	absAmount := amount
	if absAmount < 0 {
		absAmount = -absAmount
	}

	if absAmount >= 1000 {
		return 2
	} else if absAmount >= 1 {
		return 4
	} else if absAmount >= 0.0001 {
		return 6
	}
	return 8
}
