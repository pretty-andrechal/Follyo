package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/pretty-andrechal/follyo/internal/portfolio"
	"github.com/pretty-andrechal/follyo/internal/storage"
	"github.com/spf13/cobra"
)

var (
	p        *portfolio.Portfolio
	dataPath string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initPortfolio)

	rootCmd.PersistentFlags().StringVar(&dataPath, "data", "", "path to portfolio data file")

	// Add subcommands
	rootCmd.AddCommand(holdingCmd)
	rootCmd.AddCommand(loanCmd)
	rootCmd.AddCommand(saleCmd)
	rootCmd.AddCommand(summaryCmd)

	// Holding subcommands
	holdingCmd.AddCommand(holdingAddCmd)
	holdingCmd.AddCommand(holdingListCmd)
	holdingCmd.AddCommand(holdingRemoveCmd)

	// Loan subcommands
	loanCmd.AddCommand(loanAddCmd)
	loanCmd.AddCommand(loanListCmd)
	loanCmd.AddCommand(loanRemoveCmd)

	// Sale subcommands
	saleCmd.AddCommand(saleAddCmd)
	saleCmd.AddCommand(saleListCmd)
	saleCmd.AddCommand(saleRemoveCmd)

	// Add flags for holding add
	holdingAddCmd.Flags().StringP("platform", "p", "", "Platform where held")
	holdingAddCmd.Flags().StringP("notes", "n", "", "Optional notes")
	holdingAddCmd.Flags().StringP("date", "d", "", "Purchase date (YYYY-MM-DD)")

	// Add flags for loan add
	loanAddCmd.Flags().Float64P("rate", "r", 0, "Annual interest rate (%)")
	loanAddCmd.Flags().StringP("notes", "n", "", "Optional notes")
	loanAddCmd.Flags().StringP("date", "d", "", "Loan date (YYYY-MM-DD)")

	// Add flags for sale add
	saleAddCmd.Flags().StringP("platform", "p", "", "Platform where sold")
	saleAddCmd.Flags().StringP("notes", "n", "", "Optional notes")
	saleAddCmd.Flags().StringP("date", "d", "", "Sale date (YYYY-MM-DD)")
}

func initPortfolio() {
	if dataPath == "" {
		// Use data directory relative to current working directory
		dataPath = filepath.Join("data", "portfolio.json")
	}

	s, err := storage.New(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}
	p = portfolio.New(s)
}

var rootCmd = &cobra.Command{
	Use:   "follyo",
	Short: "Follyo - Personal Crypto Portfolio Tracker",
	Long:  "Track your crypto holdings, sales, and loans across platforms.",
}

// ============ Holdings Commands ============

var holdingCmd = &cobra.Command{
	Use:   "holding",
	Short: "Manage crypto holdings",
}

var holdingAddCmd = &cobra.Command{
	Use:   "add COIN AMOUNT PRICE",
	Short: "Add a coin holding",
	Long: `Add a coin holding.

COIN: The cryptocurrency symbol (e.g., BTC, ETH)
AMOUNT: Amount of coins
PRICE: Purchase price per coin in USD`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		coin := args[0]
		amount := parseFloat(args[1], "amount")
		price := parseFloat(args[2], "price")

		platform, _ := cmd.Flags().GetString("platform")
		notes, _ := cmd.Flags().GetString("notes")
		date, _ := cmd.Flags().GetString("date")

		holding, err := p.AddHolding(coin, amount, price, platform, notes, date)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Added holding: %v %s @ $%.2f (ID: %s)\n", holding.Amount, holding.Coin, holding.PurchasePriceUSD, holding.ID)
	},
}

var holdingListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all holdings",
	Run: func(cmd *cobra.Command, args []string) {
		holdings, err := p.ListHoldings()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(holdings) == 0 {
			fmt.Println("No holdings found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tCoin\tAmount\tPrice/Unit\tTotal USD\tPlatform\tDate")
		for _, h := range holdings {
			platform := h.Platform
			if platform == "" {
				platform = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t$%.2f\t$%.2f\t%s\t%s\n",
				h.ID, h.Coin, formatAmount(h.Amount),
				h.PurchasePriceUSD, h.TotalValueUSD(),
				platform, h.Date)
		}
		w.Flush()
	},
}

var holdingRemoveCmd = &cobra.Command{
	Use:   "remove ID",
	Short: "Remove a holding by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		removed, err := p.RemoveHolding(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if removed {
			fmt.Printf("Removed holding %s\n", id)
		} else {
			fmt.Printf("Holding %s not found\n", id)
		}
	},
}

// ============ Loan Commands ============

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
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(loans) == 0 {
			fmt.Println("No loans found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
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
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if removed {
			fmt.Printf("Removed loan %s\n", id)
		} else {
			fmt.Printf("Loan %s not found\n", id)
		}
	},
}

// ============ Sale Commands ============

var saleCmd = &cobra.Command{
	Use:   "sale",
	Short: "Manage crypto sales",
}

var saleAddCmd = &cobra.Command{
	Use:   "add COIN AMOUNT PRICE",
	Short: "Add a coin sale",
	Long: `Add a coin sale.

COIN: The cryptocurrency symbol (e.g., BTC, ETH)
AMOUNT: Amount of coins sold
PRICE: Sell price per coin in USD`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		coin := args[0]
		amount := parseFloat(args[1], "amount")
		price := parseFloat(args[2], "price")

		platform, _ := cmd.Flags().GetString("platform")
		notes, _ := cmd.Flags().GetString("notes")
		date, _ := cmd.Flags().GetString("date")

		sale, err := p.AddSale(coin, amount, price, platform, notes, date)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Added sale: %v %s @ $%.2f (ID: %s)\n", sale.Amount, sale.Coin, sale.SellPriceUSD, sale.ID)
	},
}

var saleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sales",
	Run: func(cmd *cobra.Command, args []string) {
		sales, err := p.ListSales()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(sales) == 0 {
			fmt.Println("No sales found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tCoin\tAmount\tPrice/Unit\tTotal USD\tPlatform\tDate")
		for _, s := range sales {
			platform := s.Platform
			if platform == "" {
				platform = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t$%.2f\t$%.2f\t%s\t%s\n",
				s.ID, s.Coin, formatAmount(s.Amount),
				s.SellPriceUSD, s.TotalValueUSD(),
				platform, s.Date)
		}
		w.Flush()
	},
}

var saleRemoveCmd = &cobra.Command{
	Use:   "remove ID",
	Short: "Remove a sale by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		removed, err := p.RemoveSale(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if removed {
			fmt.Printf("Removed sale %s\n", id)
		} else {
			fmt.Printf("Sale %s not found\n", id)
		}
	},
}

// ============ Summary Command ============

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show portfolio summary",
	Run: func(cmd *cobra.Command, args []string) {
		summary, err := p.GetSummary()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n=== PORTFOLIO SUMMARY ===")

		// Holdings by coin
		fmt.Println("\nHOLDINGS BY COIN:")
		if len(summary.HoldingsByCoin) > 0 {
			for _, coin := range sortedKeys(summary.HoldingsByCoin) {
				fmt.Printf("  %s: %s\n", coin, formatAmount(summary.HoldingsByCoin[coin]))
			}
		} else {
			fmt.Println("  (none)")
		}

		// Sales by coin
		fmt.Println("\nSALES BY COIN:")
		if len(summary.SalesByCoin) > 0 {
			for _, coin := range sortedKeys(summary.SalesByCoin) {
				fmt.Printf("  %s: %s\n", coin, formatAmount(summary.SalesByCoin[coin]))
			}
		} else {
			fmt.Println("  (none)")
		}

		// Loans by coin
		fmt.Println("\nLOANS BY COIN:")
		if len(summary.LoansByCoin) > 0 {
			for _, coin := range sortedKeys(summary.LoansByCoin) {
				fmt.Printf("  %s: %s\n", coin, formatAmount(summary.LoansByCoin[coin]))
			}
		} else {
			fmt.Println("  (none)")
		}

		// Net holdings
		fmt.Println("\nNET HOLDINGS (Holdings - Sales - Loans):")
		if len(summary.NetByCoin) > 0 {
			for _, coin := range sortedKeys(summary.NetByCoin) {
				amount := summary.NetByCoin[coin]
				prefix := ""
				if amount > 0 {
					prefix = "+"
				}
				fmt.Printf("  %s: %s%s\n", coin, prefix, formatAmount(amount))
			}
		} else {
			fmt.Println("  (none)")
		}

		fmt.Println("\n---------------------------")
		fmt.Printf("Total Holdings: %d\n", summary.TotalHoldingsCount)
		fmt.Printf("Total Sales: %d\n", summary.TotalSalesCount)
		fmt.Printf("Total Loans: %d\n", summary.TotalLoansCount)
		fmt.Printf("Total Invested: $%.2f\n", summary.TotalInvestedUSD)
		fmt.Printf("Total Sold: $%.2f\n", summary.TotalSoldUSD)
		fmt.Println()
	},
}

// Helper functions

func parseFloat(s, name string) float64 {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid %s: %s\n", name, s)
		os.Exit(1)
	}
	return f
}

func formatAmount(amount float64) string {
	s := fmt.Sprintf("%.8f", amount)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

func sortedKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
