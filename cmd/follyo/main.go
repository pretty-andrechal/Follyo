package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/pretty-andrechal/follyo/internal/portfolio"
	"github.com/pretty-andrechal/follyo/internal/storage"
	"github.com/spf13/cobra"
)

var (
	p        *portfolio.Portfolio
	dataPath string
)

// Testable wrappers for os functions
var (
	osExit    = os.Exit
	osStderr  io.Writer = os.Stderr
	osStdout  io.Writer = os.Stdout
	osStdin   io.Reader = os.Stdin
	sortStrings = sort.Strings
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
