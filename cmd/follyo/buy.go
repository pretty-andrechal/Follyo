package main

import (
	"fmt"
	"text/tabwriter"

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
		var price float64

		if len(args) == 3 && total > 0 {
			fmt.Fprintln(osStderr, "Error: specify either PRICE argument or --total flag, not both")
			osExit(1)
		}

		if len(args) == 3 {
			price = parseFloat(args[2], "price")
		} else if total > 0 {
			price = total / amount
		} else {
			fmt.Fprintln(osStderr, "Error: specify either PRICE argument or --total flag")
			osExit(1)
		}

		platform, _ := cmd.Flags().GetString("platform")
		// Use default platform if not specified
		if platform == "" {
			cfg := loadConfig()
			platform = cfg.GetDefaultPlatform()
		}
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
	Run: func(cmd *cobra.Command, args []string) {
		holdings, err := p.ListHoldings()
		if err != nil {
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
		}

		if len(holdings) == 0 {
			fmt.Fprintln(osStdout, "No purchases found.")
			return
		}

		w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tCoin\tAmount\tPrice/Unit\tTotal USD\tPlatform\tDate")
		for _, h := range holdings {
			platform := h.Platform
			if platform == "" {
				platform = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				h.ID, h.Coin, formatAmount(h.Amount),
				formatUSD(h.PurchasePriceUSD), formatUSD(h.TotalValueUSD()),
				platform, h.Date)
		}
		w.Flush()
	},
}

var buyRemoveCmd = &cobra.Command{
	Use:   "remove ID",
	Short: "Remove a purchase by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		removed, err := p.RemoveHolding(id)
		if err != nil {
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
		}
		if removed {
			fmt.Printf("Removed purchase %s\n", id)
		} else {
			fmt.Printf("Purchase %s not found\n", id)
		}
	},
}
