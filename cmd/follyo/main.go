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
	rootCmd.AddCommand(buyCmd)
	rootCmd.AddCommand(loanCmd)
	rootCmd.AddCommand(sellCmd)
	rootCmd.AddCommand(stakeCmd)
	rootCmd.AddCommand(summaryCmd)

	// Buy subcommands
	buyCmd.AddCommand(buyAddCmd)
	buyCmd.AddCommand(buyListCmd)
	buyCmd.AddCommand(buyRemoveCmd)

	// Loan subcommands
	loanCmd.AddCommand(loanAddCmd)
	loanCmd.AddCommand(loanListCmd)
	loanCmd.AddCommand(loanRemoveCmd)

	// Sell subcommands
	sellCmd.AddCommand(sellAddCmd)
	sellCmd.AddCommand(sellListCmd)
	sellCmd.AddCommand(sellRemoveCmd)

	// Stake subcommands
	stakeCmd.AddCommand(stakeAddCmd)
	stakeCmd.AddCommand(stakeListCmd)
	stakeCmd.AddCommand(stakeRemoveCmd)

	// Add flags for buy add
	buyAddCmd.Flags().StringP("platform", "p", "", "Platform where held")
	buyAddCmd.Flags().StringP("notes", "n", "", "Optional notes")
	buyAddCmd.Flags().StringP("date", "d", "", "Purchase date (YYYY-MM-DD)")

	// Add flags for loan add
	loanAddCmd.Flags().Float64P("rate", "r", 0, "Annual interest rate (%)")
	loanAddCmd.Flags().StringP("notes", "n", "", "Optional notes")
	loanAddCmd.Flags().StringP("date", "d", "", "Loan date (YYYY-MM-DD)")

	// Add flags for sell add
	sellAddCmd.Flags().StringP("platform", "p", "", "Platform where sold")
	sellAddCmd.Flags().StringP("notes", "n", "", "Optional notes")
	sellAddCmd.Flags().StringP("date", "d", "", "Sale date (YYYY-MM-DD)")

	// Add flags for stake add
	stakeAddCmd.Flags().Float64P("apy", "a", 0, "Annual percentage yield (%)")
	stakeAddCmd.Flags().StringP("notes", "n", "", "Optional notes")
	stakeAddCmd.Flags().StringP("date", "d", "", "Stake date (YYYY-MM-DD)")
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

// ============ Buy Commands ============

var buyCmd = &cobra.Command{
	Use:   "buy",
	Short: "Manage coin purchases",
}

var buyAddCmd = &cobra.Command{
	Use:   "add COIN AMOUNT PRICE",
	Short: "Record a coin purchase",
	Long: `Record a coin purchase.

COIN: The cryptocurrency symbol (e.g., BTC, ETH)
AMOUNT: Amount of coins bought
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
		fmt.Printf("Bought %s %s @ %s (ID: %s)\n", formatAmount(holding.Amount), holding.Coin, formatUSD(holding.PurchasePriceUSD), holding.ID)
	},
}

var buyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all purchases",
	Run: func(cmd *cobra.Command, args []string) {
		holdings, err := p.ListHoldings()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(holdings) == 0 {
			fmt.Println("No purchases found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
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
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if removed {
			fmt.Printf("Removed purchase %s\n", id)
		} else {
			fmt.Printf("Purchase %s not found\n", id)
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

// ============ Sell Commands ============

var sellCmd = &cobra.Command{
	Use:   "sell",
	Short: "Manage coin sales",
}

var sellAddCmd = &cobra.Command{
	Use:   "add COIN AMOUNT PRICE",
	Short: "Record a coin sale",
	Long: `Record a coin sale.

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
		fmt.Printf("Sold %s %s @ %s (ID: %s)\n", formatAmount(sale.Amount), sale.Coin, formatUSD(sale.SellPriceUSD), sale.ID)
	},
}

var sellListCmd = &cobra.Command{
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

// ============ Stake Commands ============

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
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(stakes) == 0 {
			fmt.Println("No stakes found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
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
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if removed {
			fmt.Printf("Removed stake %s (unstaked)\n", id)
		} else {
			fmt.Printf("Stake %s not found\n", id)
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

		// Staked by coin
		fmt.Println("\nSTAKED BY COIN:")
		if len(summary.StakesByCoin) > 0 {
			for _, coin := range sortedKeys(summary.StakesByCoin) {
				fmt.Printf("  %s: %s\n", coin, formatAmount(summary.StakesByCoin[coin]))
			}
		} else {
			fmt.Println("  (none)")
		}

		// Available by coin (holdings - sales - staked)
		fmt.Println("\nAVAILABLE BY COIN (Holdings - Sales - Staked):")
		if len(summary.AvailableByCoin) > 0 {
			for _, coin := range sortedKeys(summary.AvailableByCoin) {
				fmt.Printf("  %s: %s\n", coin, formatAmount(summary.AvailableByCoin[coin]))
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
		fmt.Printf("Total Stakes: %d\n", summary.TotalStakesCount)
		fmt.Printf("Total Loans: %d\n", summary.TotalLoansCount)
		fmt.Printf("Total Invested: %s\n", formatUSD(summary.TotalInvestedUSD))
		fmt.Printf("Total Sold: %s\n", formatUSD(summary.TotalSoldUSD))
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

// addCommas adds thousand separators to a numeric string
func addCommas(s string) string {
	// Split into integer and decimal parts
	parts := strings.Split(s, ".")
	intPart := parts[0]

	// Handle negative numbers
	negative := false
	if strings.HasPrefix(intPart, "-") {
		negative = true
		intPart = intPart[1:]
	}

	// Add commas to integer part
	n := len(intPart)
	if n <= 3 {
		if negative {
			intPart = "-" + intPart
		}
		if len(parts) > 1 {
			return intPart + "." + parts[1]
		}
		return intPart
	}

	// Build result with commas
	var result strings.Builder
	remainder := n % 3
	if remainder > 0 {
		result.WriteString(intPart[:remainder])
		if n > remainder {
			result.WriteString(",")
		}
	}
	for i := remainder; i < n; i += 3 {
		result.WriteString(intPart[i : i+3])
		if i+3 < n {
			result.WriteString(",")
		}
	}

	finalInt := result.String()
	if negative {
		finalInt = "-" + finalInt
	}

	if len(parts) > 1 {
		return finalInt + "." + parts[1]
	}
	return finalInt
}

func formatAmount(amount float64) string {
	s := fmt.Sprintf("%.8f", amount)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return addCommas(s)
}

func formatUSD(amount float64) string {
	s := fmt.Sprintf("%.2f", amount)
	return "$" + addCommas(s)
}

func sortedKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
