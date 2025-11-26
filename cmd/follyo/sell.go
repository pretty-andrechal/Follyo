package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var sellCmd = &cobra.Command{
	Use:     "sell",
	Aliases: []string{"sl"},
	Short:   "Manage coin sales",
}

var sellAddCmd = &cobra.Command{
	Use:   "add COIN AMOUNT [PRICE]",
	Short: "Record a coin sale",
	Long: `Record a coin sale.

COIN: The cryptocurrency symbol (e.g., BTC, ETH)
AMOUNT: Amount of coins sold
PRICE: Sell price per coin in USD (optional if --total is used)

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

		sale, err := p.AddSale(coin, amount, price, platform, notes, date)
		if err != nil {
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
		}
		fmt.Printf("Sold %s %s @ %s (ID: %s)\n", formatAmount(sale.Amount), sale.Coin, formatUSD(sale.SellPriceUSD), sale.ID)
	},
}

var sellListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sales",
	Run: func(cmd *cobra.Command, args []string) {
		sales, err := p.ListSales()
		if err != nil {
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
		}

		if len(sales) == 0 {
			fmt.Fprintln(osStdout, "No sales found.")
			return
		}

		w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tCoin\tAmount\tPrice/Unit\tTotal USD\tPlatform\tDate")
		for _, s := range sales {
			platform := s.Platform
			if platform == "" {
				platform = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				s.ID, s.Coin, formatAmount(s.Amount),
				formatUSD(s.SellPriceUSD), formatUSD(s.TotalValueUSD()),
				platform, s.Date)
		}
		w.Flush()
	},
}

var sellRemoveCmd = &cobra.Command{
	Use:   "remove ID",
	Short: "Remove a sale by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		removed, err := p.RemoveSale(id)
		if err != nil {
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
		}
		if removed {
			fmt.Printf("Removed sale %s\n", id)
		} else {
			fmt.Printf("Sale %s not found\n", id)
		}
	},
}
