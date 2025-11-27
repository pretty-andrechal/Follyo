package main

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var summaryCmd = &cobra.Command{
	Use:     "summary",
	Aliases: []string{"sum", "s"},
	Short:   "Show portfolio summary",
	Long: `Show portfolio summary with holdings, stakes, loans, and totals.

Live prices are fetched by default from CoinGecko (configurable via 'follyo config').
Use --no-prices to disable price fetching for this invocation.`,
	Run: func(cmd *cobra.Command, args []string) {
		summary, err := p.GetSummary()
		if err != nil {
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
		}

		// Check if --no-prices was explicitly set
		noPricesFlag, _ := cmd.Flags().GetBool("no-prices")

		// Determine whether to show prices: flag overrides config
		var showPrices bool
		if noPricesFlag {
			showPrices = false
		} else {
			// Use config preference
			cfg := loadConfig()
			showPrices = cfg.GetFetchPrices()
		}

		// Fetch live prices unless disabled
		var livePrices map[string]float64
		var unmappedTickers []string
		var isOffline bool
		if showPrices {
			// Collect all unique coins from all sections
			coins := collectAllCoins(summary.HoldingsByCoin, summary.StakesByCoin, summary.LoansByCoin, summary.NetByCoin)

			if len(coins) > 0 {
				fmt.Fprintln(osStdout, "Fetching live prices...")
				priceResult := fetchPricesForCoins(coins)
				livePrices = priceResult.Prices
				unmappedTickers = priceResult.UnmappedTickers
				isOffline = priceResult.IsOffline

				if priceResult.Error != nil {
					fmt.Fprintf(osStderr, "Warning: Could not fetch prices (offline mode): %v\n", priceResult.Error)
				}
			}
		}

		fmt.Fprintln(osStdout, "\n=== PORTFOLIO SUMMARY ===")

		// Holdings by coin (current holdings = purchases - sales)
		fmt.Fprintln(osStdout, "\nHOLDINGS BY COIN:")
		var totalCurrentValue float64
		if len(summary.HoldingsByCoin) > 0 {
			w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', tabwriter.AlignRight)
			for _, coin := range sortedKeys(summary.HoldingsByCoin) {
				amount := summary.HoldingsByCoin[coin]
				value := printCoinLine(w, coin, amount, livePrices, false)
				totalCurrentValue += value
			}
			w.Flush()
		} else {
			fmt.Fprintln(osStdout, "  (none)")
		}

		// Staked by coin
		fmt.Fprintln(osStdout, "\nSTAKED BY COIN:")
		if len(summary.StakesByCoin) > 0 {
			w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', tabwriter.AlignRight)
			for _, coin := range sortedKeys(summary.StakesByCoin) {
				amount := summary.StakesByCoin[coin]
				printCoinLine(w, coin, amount, livePrices, false)
			}
			w.Flush()
		} else {
			fmt.Fprintln(osStdout, "  (none)")
		}

		// Available by coin (holdings - staked)
		fmt.Fprintln(osStdout, "\nAVAILABLE BY COIN (Holdings - Staked):")
		if len(summary.AvailableByCoin) > 0 {
			w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', tabwriter.AlignRight)
			for _, coin := range sortedKeys(summary.AvailableByCoin) {
				amount := summary.AvailableByCoin[coin]
				printCoinLine(w, coin, amount, livePrices, false)
			}
			w.Flush()
		} else {
			fmt.Fprintln(osStdout, "  (none)")
		}

		// Loans by coin
		fmt.Fprintln(osStdout, "\nLOANS BY COIN:")
		var totalLoanValue float64
		if len(summary.LoansByCoin) > 0 {
			w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', tabwriter.AlignRight)
			for _, coin := range sortedKeys(summary.LoansByCoin) {
				amount := summary.LoansByCoin[coin]
				value := printCoinLine(w, coin, amount, livePrices, false)
				totalLoanValue += value
			}
			w.Flush()
		} else {
			fmt.Fprintln(osStdout, "  (none)")
		}

		// Net holdings (holdings - loans)
		fmt.Fprintln(osStdout, "\nNET HOLDINGS (Holdings - Loans):")
		if len(summary.NetByCoin) > 0 {
			w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', tabwriter.AlignRight)
			for _, coin := range sortedKeys(summary.NetByCoin) {
				amount := summary.NetByCoin[coin]
				printCoinLine(w, coin, amount, livePrices, true)
			}
			w.Flush()
		} else {
			fmt.Fprintln(osStdout, "  (none)")
		}

		fmt.Fprintln(osStdout, "\n---------------------------")
		fmt.Fprintf(osStdout, "Total Holdings: %d\n", summary.TotalHoldingsCount)
		fmt.Fprintf(osStdout, "Total Sales: %d\n", summary.TotalSalesCount)
		fmt.Fprintf(osStdout, "Total Stakes: %d\n", summary.TotalStakesCount)
		fmt.Fprintf(osStdout, "Total Loans: %d\n", summary.TotalLoansCount)
		fmt.Fprintf(osStdout, "Total Invested: %s\n", formatUSD(summary.TotalInvestedUSD))
		fmt.Fprintf(osStdout, "Total Sold: %s\n", formatUSD(summary.TotalSoldUSD))

		// Show value summary if prices were fetched
		if livePrices != nil && totalCurrentValue > 0 {
			fmt.Fprintln(osStdout, "\n---------------------------")
			fmt.Fprintf(osStdout, "Holdings Value: %s\n", formatUSD(totalCurrentValue))
			if totalLoanValue > 0 {
				fmt.Fprintf(osStdout, "Loans Value:   -%s\n", colorRedText(formatUSD(totalLoanValue)))
			}
			netValue := totalCurrentValue - totalLoanValue
			fmt.Fprintf(osStdout, "Net Value:      %s\n", formatUSD(netValue))
			profitLoss := netValue - summary.TotalInvestedUSD + summary.TotalSoldUSD
			profitLossPercent := safeDivide(profitLoss, summary.TotalInvestedUSD) * 100
			prefix := ""
			if profitLoss > 0 {
				prefix = "+"
			}
			plText := fmt.Sprintf("%s%s (%.1f%%)", prefix, formatUSD(profitLoss), profitLossPercent)
			fmt.Fprintf(osStdout, "Profit/Loss:    %s\n", colorByValue(plText, profitLoss))
		} else if isOffline && showPrices {
			fmt.Fprintln(osStdout, "\n---------------------------")
			fmt.Fprintln(osStdout, colorize("(Offline - prices unavailable)", colorYellow))
		}

		// Show warning for unmapped tickers
		if len(unmappedTickers) > 0 {
			fmt.Fprintln(osStdout, "\n---------------------------")
			fmt.Fprintf(osStdout, "Note: No CoinGecko mapping for: %s\n", strings.Join(unmappedTickers, ", "))
			fmt.Fprintln(osStdout, "Run 'follyo ticker search <query> <TICKER>' to add a mapping")
		}

		fmt.Fprintln(osStdout)
	},
}
