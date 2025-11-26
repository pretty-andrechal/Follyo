package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/pretty-andrechal/follyo/internal/prices"
	"github.com/spf13/cobra"
)

const defaultRefreshInterval = 2 * time.Minute

var watchCmd = &cobra.Command{
	Use:     "watch",
	Aliases: []string{"w", "live"},
	Short:   "Live dashboard with auto-refresh",
	Long: `Display a live dashboard of your portfolio with auto-updating prices.

The dashboard refreshes every 2 minutes by default, showing:
- Net holdings with current values
- Total portfolio value
- Profit/loss calculations
- Time of last update

Press Ctrl+C to exit.`,
	Run: func(cmd *cobra.Command, args []string) {
		interval, _ := cmd.Flags().GetDuration("interval")
		if interval == 0 {
			interval = defaultRefreshInterval
		}

		runLiveDashboard(interval)
	},
}

func init() {
	watchCmd.Flags().DurationP("interval", "i", defaultRefreshInterval, "Refresh interval (e.g., 1m, 30s, 2m)")
}

// runLiveDashboard runs the live dashboard with periodic refresh
func runLiveDashboard(interval time.Duration) {
	// Set up signal handling for graceful exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initial display
	displayDashboard()

	// Create ticker for periodic refresh
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fmt.Fprintf(osStdout, "\nRefreshing every %s. Press Ctrl+C to exit.\n", interval)

	for {
		select {
		case <-ticker.C:
			displayDashboard()
			fmt.Fprintf(osStdout, "\nRefreshing every %s. Press Ctrl+C to exit.\n", interval)
		case <-sigChan:
			fmt.Fprintln(osStdout, "\n\nExiting dashboard...")
			return
		}
	}
}

// displayDashboard clears screen and displays the current portfolio state
func displayDashboard() {
	// Clear screen (ANSI escape code)
	fmt.Fprint(osStdout, "\033[H\033[2J")

	summary, err := p.GetSummary()
	if err != nil {
		fmt.Fprintf(osStderr, "Error: %v\n", err)
		return
	}

	// Collect all unique coins
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

	// Fetch live prices
	var livePrices map[string]float64
	var unmappedTickers []string

	if len(allCoins) > 0 {
		ps := prices.New()

		// Load custom mappings
		cfg := loadConfig()
		customMappings := cfg.GetAllTickerMappings()
		for ticker, geckoID := range customMappings {
			ps.AddCoinMapping(ticker, geckoID)
		}

		// Convert to slice
		var coins []string
		for coin := range allCoins {
			coins = append(coins, coin)
		}
		sortStrings(coins)

		// Check for unmapped tickers
		unmappedTickers = ps.GetUnmappedTickers(coins)

		livePrices, err = ps.GetPrices(coins)
		if err != nil {
			fmt.Fprintf(osStderr, "Warning: Could not fetch prices: %v\n", err)
			livePrices = nil
		}
	}

	// Display header with timestamp
	now := time.Now()
	fmt.Fprintln(osStdout, "╔════════════════════════════════════════════════════════════╗")
	fmt.Fprintln(osStdout, "║           FOLLYO - LIVE PORTFOLIO DASHBOARD                ║")
	fmt.Fprintf(osStdout, "║           Last Update: %-35s║\n", now.Format("2006-01-02 15:04:05"))
	fmt.Fprintln(osStdout, "╚════════════════════════════════════════════════════════════╝")

	// Net holdings section (most important for dashboard)
	fmt.Fprintln(osStdout, "\n┌─ NET HOLDINGS ─────────────────────────────────────────────┐")
	var totalCurrentValue float64
	var totalLoanValue float64

	if len(summary.NetByCoin) > 0 {
		w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', tabwriter.AlignRight)
		for _, coin := range sortedKeys(summary.NetByCoin) {
			amount := summary.NetByCoin[coin]
			value := printDashboardCoinLine(w, coin, amount, livePrices)
			if amount > 0 {
				totalCurrentValue += value
			}
		}
		w.Flush()
	} else {
		fmt.Fprintln(osStdout, "  (no holdings)")
	}

	// Calculate loan values separately
	if len(summary.LoansByCoin) > 0 {
		for _, coin := range sortedKeys(summary.LoansByCoin) {
			amount := summary.LoansByCoin[coin]
			if livePrices != nil {
				if price, ok := livePrices[coin]; ok {
					totalLoanValue += amount * price
				}
			}
		}
	}

	// Staked assets summary
	if len(summary.StakesByCoin) > 0 {
		fmt.Fprintln(osStdout, "\n┌─ STAKED ASSETS ────────────────────────────────────────────┐")
		w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', tabwriter.AlignRight)
		for _, coin := range sortedKeys(summary.StakesByCoin) {
			amount := summary.StakesByCoin[coin]
			printDashboardCoinLine(w, coin, amount, livePrices)
		}
		w.Flush()
	}

	// Loans summary
	if len(summary.LoansByCoin) > 0 {
		fmt.Fprintln(osStdout, "\n┌─ OUTSTANDING LOANS ────────────────────────────────────────┐")
		w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', tabwriter.AlignRight)
		for _, coin := range sortedKeys(summary.LoansByCoin) {
			amount := summary.LoansByCoin[coin]
			printDashboardCoinLine(w, coin, amount, livePrices)
		}
		w.Flush()
	}

	// Portfolio totals
	fmt.Fprintln(osStdout, "\n┌─ PORTFOLIO TOTALS ─────────────────────────────────────────┐")

	if livePrices != nil {
		// Calculate holdings value (not net, just holdings)
		var holdingsValue float64
		for _, coin := range sortedKeys(summary.HoldingsByCoin) {
			amount := summary.HoldingsByCoin[coin]
			if price, ok := livePrices[coin]; ok {
				holdingsValue += amount * price
			}
		}

		fmt.Fprintf(osStdout, "  Holdings Value:    %s\n", formatUSD(holdingsValue))
		if totalLoanValue > 0 {
			fmt.Fprintf(osStdout, "  Loans Value:      -%s\n", colorRedText(formatUSD(totalLoanValue)))
		}
		netValue := holdingsValue - totalLoanValue
		fmt.Fprintf(osStdout, "  ────────────────────────\n")
		fmt.Fprintf(osStdout, "  Net Value:         %s\n", colorGreenText(formatUSD(netValue)))

		// Profit/Loss calculation
		profitLoss := netValue - summary.TotalInvestedUSD + summary.TotalSoldUSD
		profitLossPercent := safeDivide(profitLoss, summary.TotalInvestedUSD) * 100
		prefix := ""
		if profitLoss > 0 {
			prefix = "+"
		}
		plText := fmt.Sprintf("%s%s (%.1f%%)", prefix, formatUSD(profitLoss), profitLossPercent)
		fmt.Fprintf(osStdout, "  Profit/Loss:       %s\n", colorByValue(plText, profitLoss))
	}

	fmt.Fprintf(osStdout, "\n  Total Invested:    %s\n", formatUSD(summary.TotalInvestedUSD))
	fmt.Fprintf(osStdout, "  Total Sold:        %s\n", formatUSD(summary.TotalSoldUSD))

	// Show warning for unmapped tickers
	if len(unmappedTickers) > 0 {
		fmt.Fprintln(osStdout, "\n┌─ WARNINGS ─────────────────────────────────────────────────┐")
		fmt.Fprintf(osStdout, "  No price data for: %s\n", strings.Join(unmappedTickers, ", "))
		fmt.Fprintln(osStdout, "  Run 'follyo ticker search <query> <TICKER>' to add mapping")
	}

	fmt.Fprintln(osStdout, "\n└────────────────────────────────────────────────────────────┘")
}

// printDashboardCoinLine prints a coin line for the dashboard and returns the value
func printDashboardCoinLine(w *tabwriter.Writer, coin string, amount float64, livePrices map[string]float64) float64 {
	if livePrices != nil {
		if price, ok := livePrices[coin]; ok {
			value := amount * price
			fmt.Fprintf(w, "  %-8s\t%s\t@ %s\t= %s\t\n",
				coin+":", formatAmountAligned(amount), formatUSD(price), formatUSD(value))
			return value
		}
		fmt.Fprintf(w, "  %-8s\t%s\t@ %s\t= %s\t\n",
			coin+":", formatAmountAligned(amount), "N/A", "N/A")
		return 0
	}
	fmt.Fprintf(w, "  %-8s\t%s\t\n", coin+":", formatAmountAligned(amount))
	return 0
}
