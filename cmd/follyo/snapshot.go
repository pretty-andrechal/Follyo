package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/pretty-andrechal/follyo/internal/models"
	"github.com/pretty-andrechal/follyo/internal/portfolio"
	"github.com/pretty-andrechal/follyo/internal/prices"
	"github.com/pretty-andrechal/follyo/internal/storage"
	"github.com/spf13/cobra"
)

var snapshotStore *storage.SnapshotStore

func initSnapshotStore() *storage.SnapshotStore {
	if snapshotStore != nil {
		return snapshotStore
	}

	snapshotPath := filepath.Join("data", "snapshots.json")
	ss, err := storage.NewSnapshotStore(snapshotPath)
	if err != nil {
		fmt.Fprintf(osStderr, "Error loading snapshots: %v\n", err)
		osExit(1)
	}
	snapshotStore = ss
	return ss
}

var snapshotCmd = &cobra.Command{
	Use:     "snapshot",
	Aliases: []string{"snap"},
	Short:   "Manage portfolio snapshots",
	Long: `Manage historical portfolio value snapshots.

Save point-in-time snapshots of your portfolio value to track performance over time.`,
}

var snapshotSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save a snapshot of current portfolio value",
	Long: `Save a snapshot of current portfolio value.

Fetches live prices and saves the current portfolio state for historical comparison.`,
	Run: func(cmd *cobra.Command, args []string) {
		note, _ := cmd.Flags().GetString("note")

		// Get portfolio summary
		summary, err := p.GetSummary()
		if err != nil {
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
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
			fmt.Fprintln(osStderr, "Error: No holdings to snapshot")
			osExit(1)
		}

		// Fetch prices
		fmt.Fprintln(osStdout, "Fetching live prices...")
		ps := prices.New()

		// Load custom mappings
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
			fmt.Fprintf(osStderr, "Error fetching prices: %v\n", err)
			osExit(1)
		}

		// Create snapshot
		snapshot, err := p.CreateSnapshot(livePrices, note)
		if err != nil {
			fmt.Fprintf(osStderr, "Error creating snapshot: %v\n", err)
			osExit(1)
		}

		// Save snapshot
		ss := initSnapshotStore()
		if err := ss.Add(snapshot); err != nil {
			fmt.Fprintf(osStderr, "Error saving snapshot: %v\n", err)
			osExit(1)
		}

		fmt.Fprintf(osStdout, "Saved snapshot %s\n", snapshot.ID)
		fmt.Fprintf(osStdout, "  Net Value:    %s\n", formatUSD(snapshot.NetValue))
		plText := fmt.Sprintf("%s (%.1f%%)", formatUSDWithSign(snapshot.ProfitLoss), snapshot.ProfitPercent)
		fmt.Fprintf(osStdout, "  Profit/Loss:  %s\n", colorByValue(plText, snapshot.ProfitLoss))
	},
}

var snapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all snapshots",
	Run: func(cmd *cobra.Command, args []string) {
		ss := initSnapshotStore()
		snapshots := ss.List()

		if len(snapshots) == 0 {
			fmt.Fprintln(osStdout, "No snapshots found.")
			fmt.Fprintln(osStdout, "Use 'follyo snapshot save' to create one.")
			return
		}

		w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tDate\tNet Value\tP/L\tNote")
		for _, s := range snapshots {
			dateStr := s.Timestamp.Format("2006-01-02 15:04")
			plText := formatUSDWithSign(s.ProfitLoss)
			note := s.Note
			if len(note) > 20 {
				note = note[:17] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				s.ID, dateStr, formatUSD(s.NetValue), colorByValue(plText, s.ProfitLoss), note)
		}
		w.Flush()
	},
}

var snapshotShowCmd = &cobra.Command{
	Use:   "show ID",
	Short: "Show details of a snapshot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ss := initSnapshotStore()
		snapshot, found := ss.Get(args[0])
		if !found {
			fmt.Fprintf(osStderr, "Snapshot %s not found\n", args[0])
			osExit(1)
		}

		fmt.Fprintf(osStdout, "Snapshot: %s\n", snapshot.ID)
		fmt.Fprintf(osStdout, "Date:     %s\n", snapshot.Timestamp.Format("2006-01-02 15:04:05"))
		if snapshot.Note != "" {
			fmt.Fprintf(osStdout, "Note:     %s\n", snapshot.Note)
		}

		fmt.Fprintln(osStdout, "\nCOIN VALUES:")
		if len(snapshot.CoinValues) > 0 {
			w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', tabwriter.AlignRight)

			// Sort coins
			var coins []string
			for coin := range snapshot.CoinValues {
				coins = append(coins, coin)
			}
			sort.Strings(coins)

			for _, coin := range coins {
				cv := snapshot.CoinValues[coin]
				fmt.Fprintf(w, "  %-8s\t%s\t@ %s\t= %s\t\n",
					coin+":", formatAmountAligned(cv.Amount), formatUSD(cv.Price), formatUSD(cv.Value))
			}
			w.Flush()
		}

		fmt.Fprintln(osStdout, "\n---------------------------")
		fmt.Fprintf(osStdout, "Holdings Value: %s\n", formatUSD(snapshot.HoldingsValue))
		if snapshot.LoansValue > 0 {
			fmt.Fprintf(osStdout, "Loans Value:   -%s\n", colorRedText(formatUSD(snapshot.LoansValue)))
		}
		fmt.Fprintf(osStdout, "Net Value:      %s\n", formatUSD(snapshot.NetValue))
		fmt.Fprintf(osStdout, "Total Invested: %s\n", formatUSD(snapshot.TotalInvested))
		fmt.Fprintf(osStdout, "Total Sold:     %s\n", formatUSD(snapshot.TotalSold))

		plText := fmt.Sprintf("%s (%.1f%%)", formatUSDWithSign(snapshot.ProfitLoss), snapshot.ProfitPercent)
		fmt.Fprintf(osStdout, "Profit/Loss:    %s\n", colorByValue(plText, snapshot.ProfitLoss))
	},
}

var snapshotCompareCmd = &cobra.Command{
	Use:   "compare ID1 [ID2]",
	Short: "Compare two snapshots (or snapshot to current)",
	Long: `Compare two snapshots, or compare a snapshot to current portfolio.

If only one ID is provided, compares that snapshot to the current portfolio state.
If two IDs are provided, compares the two snapshots.`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		ss := initSnapshotStore()

		// Get the first (older) snapshot
		older, found := ss.Get(args[0])
		if !found {
			fmt.Fprintf(osStderr, "Snapshot %s not found\n", args[0])
			osExit(1)
		}

		var newer *models.Snapshot

		if len(args) == 2 {
			// Compare two snapshots
			newer, found = ss.Get(args[1])
			if !found {
				fmt.Fprintf(osStderr, "Snapshot %s not found\n", args[1])
				osExit(1)
			}
		} else {
			// Compare to current - fetch live prices
			summary, err := p.GetSummary()
			if err != nil {
				fmt.Fprintf(osStderr, "Error: %v\n", err)
				osExit(1)
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
				fmt.Fprintln(osStderr, "Error: No current holdings to compare")
				osExit(1)
			}

			fmt.Fprintln(osStdout, "Fetching live prices...")
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
				fmt.Fprintf(osStderr, "Error fetching prices: %v\n", err)
				osExit(1)
			}

			currentSnapshot, err := p.CreateSnapshot(livePrices, "")
			if err != nil {
				fmt.Fprintf(osStderr, "Error: %v\n", err)
				osExit(1)
			}
			newer = &currentSnapshot
		}

		// Ensure older is actually older
		if older.Timestamp.After(newer.Timestamp) {
			older, newer = newer, older
		}

		comparison := portfolio.CompareSnapshots(older, newer)

		// Display comparison
		olderLabel := fmt.Sprintf("%s (%s)", older.ID, older.Timestamp.Format("2006-01-02"))
		newerLabel := "Current"
		if len(args) == 2 {
			newerLabel = fmt.Sprintf("%s (%s)", newer.ID, newer.Timestamp.Format("2006-01-02"))
		}

		fmt.Fprintf(osStdout, "\nComparing %s to %s:\n\n", olderLabel, newerLabel)

		w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "\tThen\tNow\tChange\n")
		fmt.Fprintf(w, "Net Value:\t%s\t%s\t%s\n",
			formatUSD(older.NetValue),
			formatUSD(newer.NetValue),
			colorByValue(formatChangeWithPercent(comparison.NetValueChange, comparison.NetValuePercent), comparison.NetValueChange))

		fmt.Fprintf(w, "Profit/Loss:\t%s\t%s\t%s\n",
			formatUSDWithSign(older.ProfitLoss),
			formatUSDWithSign(newer.ProfitLoss),
			colorByValue(formatUSDWithSign(comparison.ProfitLossChange), comparison.ProfitLossChange))
		w.Flush()

		// Show coin changes
		if len(comparison.CoinChanges) > 0 {
			fmt.Fprintln(osStdout, "\nCOIN CHANGES:")
			w = tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', 0)

			// Sort coins
			var coins []string
			for coin := range comparison.CoinChanges {
				coins = append(coins, coin)
			}
			sort.Strings(coins)

			for _, coin := range coins {
				cc := comparison.CoinChanges[coin]
				if cc.OldValue == 0 && cc.NewValue == 0 {
					continue // Skip coins with no value in either snapshot
				}
				priceChange := ""
				if cc.OldPrice > 0 && cc.NewPrice > 0 {
					pricePct := ((cc.NewPrice - cc.OldPrice) / cc.OldPrice) * 100
					priceChange = fmt.Sprintf("(%.1f%%)", pricePct)
				}
				fmt.Fprintf(w, "  %s:\t%s @ %s\t→\t%s @ %s\t%s\t%s\n",
					coin,
					formatAmount(cc.OldAmount), formatUSD(cc.OldPrice),
					formatAmount(cc.NewAmount), formatUSD(cc.NewPrice),
					colorByValue(formatUSDWithSign(cc.ValueChange), cc.ValueChange),
					priceChange)
			}
			w.Flush()
		}
	},
}

var snapshotRemoveCmd = &cobra.Command{
	Use:   "remove ID",
	Short: "Remove a snapshot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ss := initSnapshotStore()
		removed, err := ss.Remove(args[0])
		if err != nil {
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
		}
		if removed {
			fmt.Fprintf(osStdout, "Removed snapshot %s\n", args[0])
		} else {
			fmt.Fprintf(osStdout, "Snapshot %s not found\n", args[0])
		}
	},
}

var snapshotDailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Save a daily snapshot with today's date as note",
	Long: `Save a snapshot with today's date (YYYY-MM-DD) as the note.

This is a convenience command for creating daily snapshots with consistent naming.
Equivalent to: follyo snapshot save --note "2024-01-15"`,
	Run: func(cmd *cobra.Command, args []string) {
		note := time.Now().Format("2006-01-02")

		// Get portfolio summary
		summary, err := p.GetSummary()
		if err != nil {
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
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
			fmt.Fprintln(osStderr, "Error: No holdings to snapshot")
			osExit(1)
		}

		// Fetch prices
		fmt.Fprintln(osStdout, "Fetching live prices...")
		ps := prices.New()

		// Load custom mappings
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
			fmt.Fprintf(osStderr, "Error fetching prices: %v\n", err)
			osExit(1)
		}

		// Create snapshot
		snapshot, err := p.CreateSnapshot(livePrices, note)
		if err != nil {
			fmt.Fprintf(osStderr, "Error creating snapshot: %v\n", err)
			osExit(1)
		}

		// Save snapshot
		ss := initSnapshotStore()
		if err := ss.Add(snapshot); err != nil {
			fmt.Fprintf(osStderr, "Error saving snapshot: %v\n", err)
			osExit(1)
		}

		fmt.Fprintf(osStdout, "Saved daily snapshot %s (%s)\n", snapshot.ID, note)
		fmt.Fprintf(osStdout, "  Net Value:    %s\n", formatUSD(snapshot.NetValue))
		plText := fmt.Sprintf("%s (%.1f%%)", formatUSDWithSign(snapshot.ProfitLoss), snapshot.ProfitPercent)
		fmt.Fprintf(osStdout, "  Profit/Loss:  %s\n", colorByValue(plText, snapshot.ProfitLoss))
	},
}

// formatUSDWithSign formats USD with explicit + sign for positive values
func formatUSDWithSign(amount float64) string {
	if amount >= 0 {
		return "+" + formatUSD(amount)
	}
	return formatUSD(amount)
}

// formatChangeWithPercent formats a change with percentage
func formatChangeWithPercent(change, percent float64) string {
	sign := ""
	if change >= 0 {
		sign = "+"
	}
	return fmt.Sprintf("%s%s (%.1f%%)", sign, formatUSD(change), percent)
}
