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
const autoSnapshotHour = 8 // Take auto-snapshot at 8am

// watchAutoSnapshotTaken tracks if auto-snapshot was taken in this watch session
var watchAutoSnapshotTaken bool

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

	// Reset auto-snapshot tracking for this watch session
	watchAutoSnapshotTaken = false

	// Track next refresh time
	nextRefresh := time.Now().Add(interval)

	// Initial display
	displayDashboard()
	checkAndTakeAutoSnapshot() // Check on initial load
	printStatusLine(interval, nextRefresh)

	// Create ticker for countdown updates (every second)
	countdownTicker := time.NewTicker(1 * time.Second)
	defer countdownTicker.Stop()

	for {
		select {
		case <-countdownTicker.C:
			remaining := time.Until(nextRefresh)
			if remaining <= 0 {
				// Time to refresh
				displayDashboard()
				checkAndTakeAutoSnapshot() // Check on each refresh
				nextRefresh = time.Now().Add(interval)
				printStatusLine(interval, nextRefresh)
			} else {
				// Update countdown only
				updateCountdown(interval, remaining)
			}
		case <-sigChan:
			fmt.Fprintln(osStdout, "\n\nExiting dashboard...")
			return
		}
	}
}

// printStatusLine prints the status line with countdown
func printStatusLine(interval time.Duration, nextRefresh time.Time) {
	remaining := time.Until(nextRefresh)
	if remaining < 0 {
		remaining = 0
	}
	fmt.Fprintf(osStdout, "\nNext refresh in %s (every %s). Press Ctrl+C to exit.",
		formatDuration(remaining), interval)
}

// updateCountdown updates only the countdown portion of the status line
func updateCountdown(interval time.Duration, remaining time.Duration) {
	// Move cursor to beginning of line and clear it
	fmt.Fprint(osStdout, "\r\033[K")
	fmt.Fprintf(osStdout, "Next refresh in %s (every %s). Press Ctrl+C to exit.",
		formatDuration(remaining), interval)
}

// formatDuration formats a duration as a human-readable countdown
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	d = d.Round(time.Second)

	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60

	if minutes > 0 {
		return fmt.Sprintf("%dm %02ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// displayDashboard clears screen and displays the current portfolio state
func displayDashboard() {
	// Clear screen (ANSI escape code)
	fmt.Fprint(osStdout, "\033[H\033[2J")

	// Reload portfolio data from disk to pick up any changes
	// made by other processes (e.g., buys/sells from CLI)
	if err := p.Reload(); err != nil {
		fmt.Fprintf(osStderr, "Warning: Could not reload portfolio: %v\n", err)
		// Continue with cached data
	}

	summary, err := p.GetSummary()
	if err != nil {
		fmt.Fprintf(osStderr, "Error: %v\n", err)
		return
	}

	// Collect all unique coins and fetch prices
	coins := collectAllCoins(summary.HoldingsByCoin, summary.StakesByCoin, summary.LoansByCoin, summary.NetByCoin)

	var livePrices map[string]float64
	var unmappedTickers []string
	var isOffline bool

	if len(coins) > 0 {
		priceResult := fetchPricesForCoins(coins)
		livePrices = priceResult.Prices
		unmappedTickers = priceResult.UnmappedTickers
		isOffline = priceResult.IsOffline

		if priceResult.Error != nil {
			fmt.Fprintf(osStderr, "Warning: Could not fetch prices (offline mode): %v\n", priceResult.Error)
		}
	}

	// Display header with timestamp
	now := time.Now()
	fmt.Fprintln(osStdout, "╔════════════════════════════════════════════════════════════╗")
	fmt.Fprintln(osStdout, "║           FOLLYO - LIVE PORTFOLIO DASHBOARD                ║")
	if isOffline {
		fmt.Fprintf(osStdout, "║           Last Update: %-27s %s║\n", now.Format("2006-01-02 15:04:05"), colorize("(OFFLINE)", colorYellow))
	} else {
		fmt.Fprintf(osStdout, "║           Last Update: %-35s║\n", now.Format("2006-01-02 15:04:05"))
	}
	fmt.Fprintln(osStdout, "╚════════════════════════════════════════════════════════════╝")

	// Net holdings section (most important for dashboard)
	fmt.Fprintln(osStdout, "\n┌─ NET HOLDINGS ─────────────────────────────────────────────┐")
	var totalCurrentValue float64
	var totalLoanValue float64

	if len(summary.NetByCoin) > 0 {
		w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', tabwriter.AlignRight)
		// Sort by USD value (highest first) instead of alphabetically
		for _, coin := range sortByUSDValue(summary.NetByCoin, livePrices) {
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

// checkAndTakeAutoSnapshot checks if it's time for an auto-snapshot and takes one if needed
func checkAndTakeAutoSnapshot() {
	// Skip if already taken this session
	if watchAutoSnapshotTaken {
		return
	}

	now := time.Now()

	// Only take snapshot at or after the configured hour
	if now.Hour() < autoSnapshotHour {
		return
	}

	// Check if snapshot already exists for today
	ss := initSnapshotStore()
	if ss.HasSnapshotForToday() {
		watchAutoSnapshotTaken = true // Don't check again this session
		return
	}

	// Take the auto-snapshot
	if err := takeWatchAutoSnapshot(); err != nil {
		fmt.Fprintf(osStderr, "\nWarning: Auto-snapshot failed: %v\n", err)
	} else {
		fmt.Fprintf(osStdout, "\n  [Auto-snapshot saved for %s]\n", now.Format("2006-01-02"))
	}

	watchAutoSnapshotTaken = true
}

// takeWatchAutoSnapshot takes an automatic daily snapshot
func takeWatchAutoSnapshot() error {
	// Get portfolio summary
	summary, err := p.GetSummary()
	if err != nil {
		return fmt.Errorf("getting summary: %w", err)
	}

	// Collect all coins
	allCoins := make(map[string]bool)
	for coin := range summary.HoldingsByCoin {
		allCoins[coin] = true
	}
	for coin := range summary.LoansByCoin {
		allCoins[coin] = true
	}

	if len(allCoins) == 0 {
		return fmt.Errorf("no holdings to snapshot")
	}

	// Fetch prices
	ps := prices.New()
	cfg := loadConfig()
	customMappings := cfg.GetAllTickerMappings()
	for ticker, geckoID := range customMappings {
		ps.AddCoinMapping(ticker, geckoID)
	}

	var coins []string
	for coin := range allCoins {
		coins = append(coins, coin)
	}

	livePrices, err := ps.GetPrices(coins)
	if err != nil {
		return fmt.Errorf("fetching prices: %w", err)
	}

	// Create snapshot with auto-generated note
	note := fmt.Sprintf("Auto-snapshot %s", time.Now().Format("2006-01-02"))
	snapshot, err := p.CreateSnapshot(livePrices, note)
	if err != nil {
		return fmt.Errorf("creating snapshot: %w", err)
	}

	// Save snapshot
	ss := initSnapshotStore()
	if err := ss.Add(snapshot); err != nil {
		return fmt.Errorf("saving snapshot: %w", err)
	}

	return nil
}
