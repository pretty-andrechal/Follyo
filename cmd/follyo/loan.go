package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var loanCmd = &cobra.Command{
	Use:   "loan",
	Short: "Manage crypto loans",
}

var loanAddCmd = &cobra.Command{
	Use:   "add COIN AMOUNT PLATFORM",
	Short: "Add a loan",
	Long: `Add a loan.

COIN: The cryptocurrency symbol (e.g., BTC, USDT)
AMOUNT: Amount borrowed
PLATFORM: Platform where loan is held (e.g., Nexo, Celsius)`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		coin := args[0]
		amount := parseFloat(args[1], "amount")
		platform := args[2]

		rate, _ := cmd.Flags().GetFloat64("rate")
		var ratePtr *float64
		if rate != 0 {
			ratePtr = &rate
		}
		notes, _ := cmd.Flags().GetString("notes")
		date, _ := cmd.Flags().GetString("date")

		loan, err := p.AddLoan(coin, amount, platform, ratePtr, notes, date)
		if err != nil {
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
		}
		fmt.Printf("Added loan: %v %s on %s (ID: %s)\n", loan.Amount, loan.Coin, loan.Platform, loan.ID)
	},
}

var loanListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all loans",
	Run: func(cmd *cobra.Command, args []string) {
		loans, err := p.ListLoans()
		if err != nil {
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
		}

		if len(loans) == 0 {
			fmt.Fprintln(osStdout, "No loans found.")
			return
		}

		w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tCoin\tAmount\tPlatform\tRate\tDate")
		for _, l := range loans {
			rate := "-"
			if l.InterestRate != nil {
				rate = fmt.Sprintf("%.1f%%", *l.InterestRate)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				l.ID, l.Coin, formatAmount(l.Amount),
				l.Platform, rate, l.Date)
		}
		w.Flush()
	},
}

var loanRemoveCmd = &cobra.Command{
	Use:   "remove ID",
	Short: "Remove a loan by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		removed, err := p.RemoveLoan(id)
		if err != nil {
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
		}
		if removed {
			fmt.Printf("Removed loan %s\n", id)
		} else {
			fmt.Printf("Loan %s not found\n", id)
		}
	},
}
