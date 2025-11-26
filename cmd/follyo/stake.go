package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var stakeCmd = &cobra.Command{
	Use:   "stake",
	Short: "Manage staked crypto",
}

var stakeAddCmd = &cobra.Command{
	Use:   "add COIN AMOUNT PLATFORM",
	Short: "Stake crypto on a platform",
	Long: `Stake crypto on a platform.

COIN: The cryptocurrency symbol (e.g., ETH, SOL)
AMOUNT: Amount to stake
PLATFORM: Platform where staking (e.g., Lido, Coinbase)

Note: You can only stake coins you own (holdings - sales - already staked).`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		coin := args[0]
		amount := parseFloat(args[1], "amount")
		platform := args[2]

		apy, _ := cmd.Flags().GetFloat64("apy")
		var apyPtr *float64
		if apy != 0 {
			apyPtr = &apy
		}
		notes, _ := cmd.Flags().GetString("notes")
		date, _ := cmd.Flags().GetString("date")

		stake, err := p.AddStake(coin, amount, platform, apyPtr, notes, date)
		if err != nil {
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
		}
		fmt.Printf("Staked %v %s on %s (ID: %s)\n", stake.Amount, stake.Coin, stake.Platform, stake.ID)
	},
}

var stakeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all staked crypto",
	Run: func(cmd *cobra.Command, args []string) {
		stakes, err := p.ListStakes()
		if err != nil {
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
		}

		if len(stakes) == 0 {
			fmt.Fprintln(osStdout, "No stakes found.")
			return
		}

		w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tCoin\tAmount\tPlatform\tAPY\tDate")
		for _, st := range stakes {
			apy := "-"
			if st.APY != nil {
				apy = fmt.Sprintf("%.1f%%", *st.APY)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				st.ID, st.Coin, formatAmount(st.Amount),
				st.Platform, apy, st.Date)
		}
		w.Flush()
	},
}

var stakeRemoveCmd = &cobra.Command{
	Use:   "remove ID",
	Short: "Remove a stake by ID (unstake)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		removed, err := p.RemoveStake(id)
		if err != nil {
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
		}
		if removed {
			fmt.Printf("Removed stake %s (unstaked)\n", id)
		} else {
			fmt.Printf("Stake %s not found\n", id)
		}
	},
}
