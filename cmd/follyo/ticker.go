package main

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/pretty-andrechal/follyo/internal/config"
	"github.com/pretty-andrechal/follyo/internal/prices"
	"github.com/spf13/cobra"
)

var tickerCmd = &cobra.Command{
	Use:     "ticker",
	Aliases: []string{"t"},
	Short:   "Manage ticker to CoinGecko ID mappings",
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
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
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
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
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
		sortStrings(tickers)

		// Print
		fmt.Fprintln(osStdout, "Ticker Mappings:")
		fmt.Fprintln(osStdout)

		// Show custom mappings first
		hasCustom := false
		for _, ticker := range tickers {
			m := allMappings[ticker]
			if m.isCustom {
				if !hasCustom {
					fmt.Fprintln(osStdout, "Custom mappings:")
					hasCustom = true
				}
				fmt.Fprintf(osStdout, "  %-8s -> %s\n", ticker, m.geckoID)
			}
		}

		if hasCustom {
			fmt.Fprintln(osStdout)
		}

		// Show all default mappings if --all flag is set
		if showAll {
			fmt.Fprintln(osStdout, "Default mappings:")
			for _, ticker := range tickers {
				m := allMappings[ticker]
				if !m.isCustom {
					fmt.Fprintf(osStdout, "  %-8s -> %s\n", ticker, m.geckoID)
				}
			}
		} else {
			fmt.Fprintf(osStdout, "Default mappings: %d built-in\n", len(defaultMappings))
			fmt.Fprintln(osStdout, "Use 'follyo ticker list --all' to see all default mappings")
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
			fmt.Fprintf(osStderr, "Error: %v\n", err)
			osExit(1)
		}

		if len(results) == 0 {
			fmt.Println("No results found.")
			return
		}

		// Display results
		w := tabwriter.NewWriter(osStdout, 0, 0, 2, ' ', 0)
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

		reader := bufio.NewReader(osStdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(osStderr, "Error reading input: %v\n", err)
			osExit(1)
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
			fmt.Fprintf(osStderr, "Error saving mapping: %v\n", err)
			osExit(1)
		}

		fmt.Printf("\nMapped %s -> %s (%s)\n", targetTicker, selected.ID, selected.Name)
	},
}

// cachedConfig holds the cached configuration to avoid repeated disk reads
var (
	cachedConfig     *config.ConfigStore
	configOnce       sync.Once
	configInitErr    error
)

// loadConfig loads the configuration from the default path, using a cached instance if available.
// Thread-safe via sync.Once.
func loadConfig() *config.ConfigStore {
	configOnce.Do(func() {
		configPath := filepath.Join("data", "config.json")
		cachedConfig, configInitErr = config.New(configPath)
	})

	if configInitErr != nil {
		fmt.Fprintf(osStderr, "Error loading config: %v\n", configInitErr)
		osExit(1)
	}
	return cachedConfig
}
