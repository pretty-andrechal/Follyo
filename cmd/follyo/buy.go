package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var buyCmd = &cobra.Command{
	Use:     "buy",
	Aliases: []string{"b"},
	Short:   "Manage coin purchases",
}

var buyAddCmd = &cobra.Command{
	Use:   "add COIN AMOUNT [PRICE]",
	Short: "Record a coin purchase",
	Long: `Record a coin purchase.

COIN: The cryptocurrency symbol (e.g., BTC, ETH)
AMOUNT: Amount of coins bought
PRICE: Purchase price per coin in USD (optional if --total is used)

Use either PRICE argument or --total flag, not both.`,
	Args: cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		coin := args[0]
		amount := parseFloat(args[1], "amount")
		total, _ := cmd.Flags().GetFloat64("total")
		price := parsePriceFromArgs(args, total, amount)

		platform, _ := cmd.Flags().GetString("platform")
		platform = getPlatformWithDefault(platform)
		notes, _ := cmd.Flags().GetString("notes")
		date, _ := cmd.Flags().GetString("date")

		holding, err := p.AddHolding(coin, amount, price, platform, notes, date)
		if err != nil {
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
		}
		fmt.Printf("Bought %s %s @ %s (ID: %s)\n", formatAmount(holding.Amount), holding.Coin, formatUSD(holding.PurchasePriceUSD), holding.ID)
	},
}

var buyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all purchases",
	Run: makeListRun(ListConfig{
		EmptyMessage: "No purchases found.",
		Headers:      []string{"ID", "Coin", "Amount", "Price/Unit", "Total USD", "Platform", "Date"},
		FetchAndRender: func(t *Table) (int, error) {
			holdings, err := p.ListHoldings()
			if err != nil {
				return 0, err
			}
			for _, h := range holdings {
				platform := h.Platform
				if platform == "" {
					platform = "-"
				}
				t.Row(h.ID, h.Coin, formatAmount(h.Amount),
					formatUSD(h.PurchasePriceUSD), formatUSD(h.TotalValueUSD()),
					platform, h.Date)
			}
			return len(holdings), nil
		},
	}),
}

var buyRemoveCmd = &cobra.Command{
	Use:   "remove ID",
	Short: "Remove a purchase by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleRemoveByID(args[0], "purchase", p.RemoveHolding)
	},
}
