package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
	"github.com/pretty-andrechal/follyo/internal/portfolio"
	"github.com/pretty-andrechal/follyo/internal/prices"
)

// SummaryModel represents the portfolio summary view.
type SummaryModel struct {
	portfolio       *portfolio.Portfolio
	summary         *portfolio.Summary
	livePrices      map[string]float64
	unmappedTickers []string
	isOffline       bool
	loading         bool
	spinner         spinner.Model
	keys            tui.KeyMap
	width           int
	height          int
	lastUpdated     time.Time
	err             error
	tickerMappings  map[string]string
}

// NewSummaryModel creates a new summary view model.
func NewSummaryModel(p *portfolio.Portfolio, tickerMappings map[string]string) SummaryModel {
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

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
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
		}
	}

	return m, nil
}

// View renders the summary view.
func (m SummaryModel) View() string {
	if m.loading {
		return m.renderLoading()
	}

	if m.err != nil {
		return m.renderError()
	}

	return m.renderSummary()
}

func (m SummaryModel) renderLoading() string {
	style := lipgloss.NewStyle().
		Foreground(tui.SubtleTextColor).
		Padding(2, 4)

	return style.Render(fmt.Sprintf("%s Fetching portfolio data and live prices...", m.spinner.View()))
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

func (m SummaryModel) renderSummary() string {
	var b strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Render("PORTFOLIO SUMMARY")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Calculate totals for value display
	var totalCurrentValue, totalLoanValue float64

	// Holdings section
	b.WriteString(m.renderSection("HOLDINGS BY COIN", m.summary.HoldingsByCoin, false, &totalCurrentValue))

	// Staked section
	b.WriteString(m.renderSection("STAKED BY COIN", m.summary.StakesByCoin, false, nil))

	// Available section
	b.WriteString(m.renderSection("AVAILABLE (Holdings - Staked)", m.summary.AvailableByCoin, false, nil))

	// Loans section
	b.WriteString(m.renderSection("LOANS BY COIN", m.summary.LoansByCoin, false, &totalLoanValue))

	// Net holdings section
	b.WriteString(m.renderSection("NET HOLDINGS (Holdings - Loans)", m.summary.NetByCoin, true, nil))

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
			Render(fmt.Sprintf("\n\nNo CoinGecko mapping for: %s", strings.Join(m.unmappedTickers, ", ")))
		b.WriteString(warning)
	}

	// Footer
	b.WriteString(m.renderFooter())

	// Wrap in a box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.BorderColor).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

func (m SummaryModel) renderSection(title string, data map[string]float64, showPrefix bool, accumulator *float64) string {
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

	for _, coin := range coins {
		amount := data[coin]
		b.WriteString(m.renderCoinLine(coin, amount, showPrefix, accumulator))
	}
	b.WriteString("\n")

	return b.String()
}

func (m SummaryModel) renderCoinLine(coin string, amount float64, showPrefix bool, accumulator *float64) string {
	coinStyle := lipgloss.NewStyle().
		Foreground(tui.TextColor).
		Width(10)

	amountStyle := lipgloss.NewStyle().
		Foreground(tui.TextColor)

	prefix := ""
	if showPrefix && amount > 0 {
		prefix = "+"
	}

	amountStr := fmt.Sprintf("%s%.4f", prefix, amount)

	if m.livePrices != nil {
		if price, ok := m.livePrices[coin]; ok {
			value := amount * price
			if accumulator != nil {
				*accumulator += value
			}

			priceStyle := lipgloss.NewStyle().
				Foreground(tui.SubtleTextColor)
			valueStyle := lipgloss.NewStyle().
				Foreground(tui.SuccessColor)

			return fmt.Sprintf("  %s %s @ %s = %s\n",
				coinStyle.Render(coin+":"),
				amountStyle.Render(amountStr),
				priceStyle.Render(formatUSD(price)),
				valueStyle.Render(formatUSD(value)))
		}
		// No price available
		return fmt.Sprintf("  %s %s @ %s\n",
			coinStyle.Render(coin+":"),
			amountStyle.Render(amountStr),
			lipgloss.NewStyle().Foreground(tui.MutedColor).Render("N/A"))
	}

	return fmt.Sprintf("  %s %s\n",
		coinStyle.Render(coin+":"),
		amountStyle.Render(amountStr))
}

func (m SummaryModel) renderStats() string {
	var b strings.Builder

	divider := lipgloss.NewStyle().
		Foreground(tui.BorderColor).
		Render("─────────────────────────────")
	b.WriteString(divider)
	b.WriteString("\n")

	labelStyle := lipgloss.NewStyle().
		Foreground(tui.SubtleTextColor).
		Width(16)
	valueStyle := lipgloss.NewStyle().
		Foreground(tui.TextColor)

	stats := []struct {
		label string
		value string
	}{
		{"Holdings:", fmt.Sprintf("%d", m.summary.TotalHoldingsCount)},
		{"Sales:", fmt.Sprintf("%d", m.summary.TotalSalesCount)},
		{"Stakes:", fmt.Sprintf("%d", m.summary.TotalStakesCount)},
		{"Loans:", fmt.Sprintf("%d", m.summary.TotalLoansCount)},
		{"Total Invested:", formatUSD(m.summary.TotalInvestedUSD)},
		{"Total Sold:", formatUSD(m.summary.TotalSoldUSD)},
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
		Render("─────────────────────────────")
	b.WriteString("\n")
	b.WriteString(divider)
	b.WriteString("\n")

	labelStyle := lipgloss.NewStyle().
		Foreground(tui.SubtleTextColor).
		Width(16)

	b.WriteString(labelStyle.Render("Holdings Value:"))
	b.WriteString(lipgloss.NewStyle().Foreground(tui.SuccessColor).Render(formatUSD(totalCurrentValue)))
	b.WriteString("\n")

	if totalLoanValue > 0 {
		b.WriteString(labelStyle.Render("Loans Value:"))
		b.WriteString(lipgloss.NewStyle().Foreground(tui.ErrorColor).Render("-" + formatUSD(totalLoanValue)))
		b.WriteString("\n")
	}

	netValue := totalCurrentValue - totalLoanValue
	b.WriteString(labelStyle.Render("Net Value:"))
	b.WriteString(lipgloss.NewStyle().Foreground(tui.TextColor).Bold(true).Render(formatUSD(netValue)))
	b.WriteString("\n")

	// Profit/Loss calculation
	profitLoss := netValue - m.summary.TotalInvestedUSD + m.summary.TotalSoldUSD
	profitLossPercent := safeDivide(profitLoss, m.summary.TotalInvestedUSD) * 100

	prefix := ""
	if profitLoss > 0 {
		prefix = "+"
	}
	plText := fmt.Sprintf("%s%s (%.1f%%)", prefix, formatUSD(profitLoss), profitLossPercent)

	b.WriteString(labelStyle.Render("Profit/Loss:"))
	plStyle := lipgloss.NewStyle()
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

func (m SummaryModel) renderFooter() string {
	var b strings.Builder
	b.WriteString("\n\n")

	timeStyle := lipgloss.NewStyle().
		Foreground(tui.MutedColor)

	if !m.lastUpdated.IsZero() {
		b.WriteString(timeStyle.Render(fmt.Sprintf("Last updated: %s", m.lastUpdated.Format("2006-01-02 15:04:05"))))
		b.WriteString("\n")
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(tui.MutedColor)
	help := fmt.Sprintf("%s refresh  %s back  %s quit",
		tui.HelpKeyStyle.Render("r"),
		tui.HelpKeyStyle.Render("esc"),
		tui.HelpKeyStyle.Render("q"))
	b.WriteString(helpStyle.Render(help))

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

func formatUSD(amount float64) string {
	return fmt.Sprintf("$%.2f", amount)
}

func safeDivide(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}
