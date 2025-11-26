package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/pretty-andrechal/follyo/internal/config"
	"github.com/pretty-andrechal/follyo/internal/portfolio"
	"github.com/pretty-andrechal/follyo/internal/prices"
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
	rootCmd.AddCommand(tickerCmd)

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

	// Ticker subcommands
	tickerCmd.AddCommand(tickerMapCmd)
	tickerCmd.AddCommand(tickerUnmapCmd)
	tickerCmd.AddCommand(tickerListCmd)
	tickerCmd.AddCommand(tickerSearchCmd)

	// Add flags for ticker list
	tickerListCmd.Flags().BoolP("all", "a", false, "Show all default mappings")

	// Add flags for buy add
	buyAddCmd.Flags().StringP("platform", "p", "", "Platform where held")
	buyAddCmd.Flags().StringP("notes", "n", "", "Optional notes")
	buyAddCmd.Flags().StringP("date", "d", "", "Purchase date (YYYY-MM-DD)")
	buyAddCmd.Flags().Float64P("total", "t", 0, "Total purchase cost in USD (alternative to per-unit price)")

	// Add flags for loan add
	loanAddCmd.Flags().Float64P("rate", "r", 0, "Annual interest rate (%)")
	loanAddCmd.Flags().StringP("notes", "n", "", "Optional notes")
	loanAddCmd.Flags().StringP("date", "d", "", "Loan date (YYYY-MM-DD)")

	// Add flags for sell add
	sellAddCmd.Flags().StringP("platform", "p", "", "Platform where sold")
	sellAddCmd.Flags().StringP("notes", "n", "", "Optional notes")
	sellAddCmd.Flags().StringP("date", "d", "", "Sale date (YYYY-MM-DD)")
	sellAddCmd.Flags().Float64P("total", "t", 0, "Total sale amount in USD (alternative to per-unit price)")

	// Add flags for stake add
	stakeAddCmd.Flags().Float64P("apy", "a", 0, "Annual percentage yield (%)")
	stakeAddCmd.Flags().StringP("notes", "n", "", "Optional notes")
	stakeAddCmd.Flags().StringP("date", "d", "", "Stake date (YYYY-MM-DD)")

	// Add flags for summary
	summaryCmd.Flags().BoolP("prices", "p", false, "Fetch and display live prices from CoinGecko")
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
			fmt.Fprintln(os.Stderr, "Error: specify either PRICE argument or --total flag, not both")
			os.Exit(1)
		}

		if len(args) == 3 {
			price = parseFloat(args[2], "price")
		} else if total > 0 {
			price = total / amount
		} else {
			fmt.Fprintln(os.Stderr, "Error: specify either PRICE argument or --total flag")
			os.Exit(1)
		}

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
			fmt.Fprintln(os.Stderr, "Error: specify either PRICE argument or --total flag, not both")
			os.Exit(1)
		}

		if len(args) == 3 {
			price = parseFloat(args[2], "price")
		} else if total > 0 {
			price = total / amount
		} else {
			fmt.Fprintln(os.Stderr, "Error: specify either PRICE argument or --total flag")
			os.Exit(1)
		}

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
	Long: `Show portfolio summary with holdings, stakes, loans, and totals.

Use --prices to fetch live prices from CoinGecko and display current values.`,
	Run: func(cmd *cobra.Command, args []string) {
		summary, err := p.GetSummary()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		showPrices, _ := cmd.Flags().GetBool("prices")

		// Fetch live prices if requested
		var livePrices map[string]float64
		var unmappedTickers []string
		if showPrices && len(summary.HoldingsByCoin) > 0 {
			fmt.Println("Fetching live prices...")
			ps := prices.New()

			// Load custom mappings
			cfg := loadConfig()
			customMappings := cfg.GetAllTickerMappings()
			for ticker, geckoID := range customMappings {
				ps.AddCoinMapping(ticker, geckoID)
			}

			coins := sortedKeys(summary.HoldingsByCoin)

			// Check for unmapped tickers
			unmappedTickers = ps.GetUnmappedTickers(coins)

			livePrices, err = ps.GetPrices(coins)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not fetch prices: %v\n", err)
				livePrices = nil
			}
		}

		fmt.Println("\n=== PORTFOLIO SUMMARY ===")

		// Holdings by coin (current holdings = purchases - sales)
		fmt.Println("\nHOLDINGS BY COIN:")
		var totalCurrentValue float64
		if len(summary.HoldingsByCoin) > 0 {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
			for _, coin := range sortedKeys(summary.HoldingsByCoin) {
				amount := summary.HoldingsByCoin[coin]
				if livePrices != nil {
					if price, ok := livePrices[coin]; ok {
						value := amount * price
						totalCurrentValue += value
						fmt.Fprintf(w, "  %-8s\t%s\t@ %s\t= %s\t\n",
							coin+":", formatAmountAligned(amount), formatUSD(price), formatUSD(value))
					} else {
						fmt.Fprintf(w, "  %-8s\t%s\t@ %s\t= %s\t\n",
							coin+":", formatAmountAligned(amount), "N/A", "N/A")
					}
				} else {
					fmt.Fprintf(w, "  %-8s\t%s\t\n", coin+":", formatAmountAligned(amount))
				}
			}
			w.Flush()
		} else {
			fmt.Println("  (none)")
		}

		// Staked by coin
		fmt.Println("\nSTAKED BY COIN:")
		if len(summary.StakesByCoin) > 0 {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
			for _, coin := range sortedKeys(summary.StakesByCoin) {
				fmt.Fprintf(w, "  %-8s\t%s\t\n", coin+":", formatAmountAligned(summary.StakesByCoin[coin]))
			}
			w.Flush()
		} else {
			fmt.Println("  (none)")
		}

		// Available by coin (holdings - staked)
		fmt.Println("\nAVAILABLE BY COIN (Holdings - Staked):")
		if len(summary.AvailableByCoin) > 0 {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
			for _, coin := range sortedKeys(summary.AvailableByCoin) {
				fmt.Fprintf(w, "  %-8s\t%s\t\n", coin+":", formatAmountAligned(summary.AvailableByCoin[coin]))
			}
			w.Flush()
		} else {
			fmt.Println("  (none)")
		}

		// Loans by coin
		fmt.Println("\nLOANS BY COIN:")
		if len(summary.LoansByCoin) > 0 {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
			for _, coin := range sortedKeys(summary.LoansByCoin) {
				fmt.Fprintf(w, "  %-8s\t%s\t\n", coin+":", formatAmountAligned(summary.LoansByCoin[coin]))
			}
			w.Flush()
		} else {
			fmt.Println("  (none)")
		}

		// Net holdings (holdings - loans)
		fmt.Println("\nNET HOLDINGS (Holdings - Loans):")
		if len(summary.NetByCoin) > 0 {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
			for _, coin := range sortedKeys(summary.NetByCoin) {
				amount := summary.NetByCoin[coin]
				prefix := ""
				if amount > 0 {
					prefix = "+"
				}
				fmt.Fprintf(w, "  %-8s\t%s%s\t\n", coin+":", prefix, formatAmountAligned(amount))
			}
			w.Flush()
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

		// Show value summary if prices were fetched
		if livePrices != nil && totalCurrentValue > 0 {
			fmt.Println("\n---------------------------")
			fmt.Printf("Current Value: %s\n", formatUSD(totalCurrentValue))
			profitLoss := totalCurrentValue - summary.TotalInvestedUSD + summary.TotalSoldUSD
			profitLossPercent := (profitLoss / summary.TotalInvestedUSD) * 100
			prefix := ""
			if profitLoss > 0 {
				prefix = "+"
			}
			fmt.Printf("Profit/Loss: %s%s (%.1f%%)\n", prefix, formatUSD(profitLoss), profitLossPercent)
		}

		// Show warning for unmapped tickers
		if len(unmappedTickers) > 0 {
			fmt.Println("\n---------------------------")
			fmt.Printf("Note: No CoinGecko mapping for: %s\n", strings.Join(unmappedTickers, ", "))
			fmt.Println("Run 'follyo ticker search <query> <TICKER>' to add a mapping")
		}

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

// formatAmountAligned formats amount with exactly 4 decimal places for decimal alignment
// Keeps trailing zeros to ensure decimal points line up
func formatAmountAligned(amount float64) string {
	s := fmt.Sprintf("%.4f", amount)
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

// ============ Ticker Commands ============

var tickerCmd = &cobra.Command{
	Use:   "ticker",
	Short: "Manage ticker to CoinGecko ID mappings",
	Long: `Manage ticker to CoinGecko ID mappings for live price fetching.

Use these commands to map your portfolio tickers to CoinGecko IDs,
which enables accurate price lookups with 'follyo summary --prices'.`,
}

var tickerMapCmd = &cobra.Command{
	Use:   "map TICKER COINGECKO_ID",
	Short: "Map a ticker to a CoinGecko ID",
	Long: `Map a ticker symbol to a CoinGecko ID.

Example: follyo ticker map MUTE mute-io

This creates a custom mapping that overrides any default mapping.
Use 'follyo ticker search' to find the correct CoinGecko ID.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ticker := strings.ToUpper(args[0])
		geckoID := args[1]

		cfg := loadConfig()
		if err := cfg.SetTickerMapping(ticker, geckoID); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Mapped %s -> %s\n", ticker, geckoID)
	},
}

var tickerUnmapCmd = &cobra.Command{
	Use:   "unmap TICKER",
	Short: "Remove a custom ticker mapping",
	Long: `Remove a custom ticker mapping.

If a default mapping exists for this ticker, it will be used instead.
If no default exists, the ticker will show N/A for prices.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ticker := strings.ToUpper(args[0])

		cfg := loadConfig()

		if !cfg.HasTickerMapping(ticker) {
			fmt.Printf("No custom mapping exists for %s\n", ticker)
			return
		}

		if err := cfg.RemoveTickerMapping(ticker); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Check if there's a default
		defaults := prices.GetDefaultMappings()
		if defaultID, ok := defaults[ticker]; ok {
			fmt.Printf("Removed custom mapping for %s (will use default: %s)\n", ticker, defaultID)
		} else {
			fmt.Printf("Removed mapping for %s\n", ticker)
		}
	},
}

var tickerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ticker mappings",
	Long:  `List all ticker mappings (both default and custom).`,
	Run: func(cmd *cobra.Command, args []string) {
		showAll, _ := cmd.Flags().GetBool("all")

		cfg := loadConfig()
		customMappings := cfg.GetAllTickerMappings()
		defaultMappings := prices.GetDefaultMappings()

		// Merge mappings (custom overrides default)
		allMappings := make(map[string]struct {
			geckoID  string
			isCustom bool
		})

		for ticker, geckoID := range defaultMappings {
			allMappings[ticker] = struct {
				geckoID  string
				isCustom bool
			}{geckoID, false}
		}
		for ticker, geckoID := range customMappings {
			allMappings[ticker] = struct {
				geckoID  string
				isCustom bool
			}{geckoID, true}
		}

		// Sort tickers
		var tickers []string
		for t := range allMappings {
			tickers = append(tickers, t)
		}
		sort.Strings(tickers)

		// Print
		fmt.Println("Ticker Mappings:")
		fmt.Println()

		// Show custom mappings first
		hasCustom := false
		for _, ticker := range tickers {
			m := allMappings[ticker]
			if m.isCustom {
				if !hasCustom {
					fmt.Println("Custom mappings:")
					hasCustom = true
				}
				fmt.Printf("  %-8s -> %s\n", ticker, m.geckoID)
			}
		}

		if hasCustom {
			fmt.Println()
		}

		// Show all default mappings if --all flag is set
		if showAll {
			fmt.Println("Default mappings:")
			for _, ticker := range tickers {
				m := allMappings[ticker]
				if !m.isCustom {
					fmt.Printf("  %-8s -> %s\n", ticker, m.geckoID)
				}
			}
		} else {
			fmt.Printf("Default mappings: %d built-in\n", len(defaultMappings))
			fmt.Println("Use 'follyo ticker list --all' to see all default mappings")
		}
	},
}

var tickerSearchCmd = &cobra.Command{
	Use:   "search QUERY [TICKER]",
	Short: "Search CoinGecko for a coin and optionally map it",
	Long: `Search CoinGecko for coins matching the query.

If TICKER is provided, you can interactively select a result to map.

Examples:
  follyo ticker search bitcoin     # Just search
  follyo ticker search mute MUTE   # Search and map result to MUTE`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		var targetTicker string
		if len(args) > 1 {
			targetTicker = strings.ToUpper(args[1])
		}

		fmt.Printf("Searching CoinGecko for \"%s\"...\n\n", query)

		ps := prices.New()
		results, err := ps.SearchCoins(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("No results found.")
			return
		}

		// Display results
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "  #\tID\tName\tSymbol\tRank")
		for i, r := range results {
			rank := "-"
			if r.Rank > 0 {
				rank = fmt.Sprintf("#%d", r.Rank)
			}
			fmt.Fprintf(w, "  %d\t%s\t%s\t%s\t%s\n",
				i+1, r.ID, r.Name, strings.ToUpper(r.Symbol), rank)
		}
		w.Flush()

		// If no ticker specified, just show results
		if targetTicker == "" {
			fmt.Println("\nTo map a result, run: follyo ticker search <query> <TICKER>")
			return
		}

		// Interactive selection
		fmt.Printf("\nSelect a result (1-%d) to map to %s, or 0 to cancel: ", len(results), targetTicker)

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}

		input = strings.TrimSpace(input)
		selection, err := strconv.Atoi(input)
		if err != nil || selection < 0 || selection > len(results) {
			fmt.Println("Invalid selection.")
			return
		}

		if selection == 0 {
			fmt.Println("Cancelled.")
			return
		}

		// Map the selected result
		selected := results[selection-1]
		cfg := loadConfig()
		if err := cfg.SetTickerMapping(targetTicker, selected.ID); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving mapping: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nMapped %s -> %s (%s)\n", targetTicker, selected.ID, selected.Name)
	},
}

// loadConfig loads the configuration from the default path
func loadConfig() *config.ConfigStore {
	configPath := filepath.Join("data", "config.json")
	cfg, err := config.New(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}
