package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/pretty-andrechal/follyo/internal/prices"
	"golang.org/x/term"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// colorPreference caches the user's color preference
var (
	colorPreference     bool
	colorPreferenceOnce sync.Once
)

// colorEnabled checks if color output should be used.
// Thread-safe via sync.Once.
func colorEnabled() bool {
	// Check if stdout is a terminal first
	isTerminal := false
	if f, ok := osStdout.(*os.File); ok {
		isTerminal = term.IsTerminal(int(f.Fd()))
	}
	if !isTerminal {
		return false
	}

	// Check user preference (lazy load, thread-safe)
	colorPreferenceOnce.Do(func() {
		cfg := loadConfig()
		colorPreference = cfg.GetColorOutput()
	})
	return colorPreference
}

// colorize wraps text in ANSI color codes if colors are enabled
func colorize(text, color string) string {
	if !colorEnabled() {
		return text
	}
	return color + text + colorReset
}

// colorGreenText returns green colored text
func colorGreenText(text string) string {
	return colorize(text, colorGreen)
}

// colorRedText returns red colored text
func colorRedText(text string) string {
	return colorize(text, colorRed)
}

// colorByValue returns green for positive, red for negative
func colorByValue(text string, value float64) string {
	if value > 0 {
		return colorGreenText(text)
	} else if value < 0 {
		return colorRedText(text)
	}
	return text
}

// parseFloat parses a float64 from a string, exiting on error
func parseFloat(s, name string) float64 {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	if err != nil {
		fmt.Fprintf(osStderr, "Error: invalid %s: %s\n", name, s)
		osExit(1)
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
	sortStrings(keys)
	return keys
}

// printCoinLine prints a coin line with optional price info and returns the computed value.
// showPrefix adds +/- prefix for amounts (used in NET HOLDINGS section).
func printCoinLine(w *tabwriter.Writer, coin string, amount float64, livePrices map[string]float64, showPrefix bool) float64 {
	amountPrefix := ""
	if showPrefix && amount > 0 {
		amountPrefix = "+"
	}

	if livePrices != nil {
		if price, ok := livePrices[coin]; ok {
			value := amount * price
			valuePrefix := ""
			if showPrefix && value > 0 {
				valuePrefix = "+"
			}
			fmt.Fprintf(w, "  %-8s\t%s%s\t@ %s\t= %s%s\t\n",
				coin+":", amountPrefix, formatAmountAligned(amount), formatUSD(price), valuePrefix, formatUSD(value))
			return value
		}
		fmt.Fprintf(w, "  %-8s\t%s%s\t@ %s\t= %s\t\n",
			coin+":", amountPrefix, formatAmountAligned(amount), "N/A", "N/A")
		return 0
	}
	fmt.Fprintf(w, "  %-8s\t%s%s\t\n", coin+":", amountPrefix, formatAmountAligned(amount))
	return 0
}

// safeDivide performs division with a guard against division by zero
func safeDivide(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

// parsePriceFromArgs parses price from either a PRICE argument or --total flag.
// Returns the calculated price per unit. Exits with error if neither or both are specified.
// Example: parsePriceFromArgs([]string{"BTC", "0.5", "50000"}, 0, 0.5) returns 50000
// Example: parsePriceFromArgs([]string{"BTC", "0.5"}, 1000, 0.5) returns 2000
func parsePriceFromArgs(args []string, total, amount float64) float64 {
	if len(args) == 3 && total > 0 {
		fmt.Fprintln(osStderr, "Error: specify either PRICE argument or --total flag, not both")
		fmt.Fprintln(osStderr, "Example: follyo buy add BTC 0.5 50000")
		fmt.Fprintln(osStderr, "     or: follyo buy add BTC 0.5 --total 25000")
		osExit(1)
	}

	if len(args) == 3 {
		return parseFloat(args[2], "price")
	} else if total > 0 {
		return total / amount
	}

	fmt.Fprintln(osStderr, "Error: specify either PRICE argument or --total flag")
	fmt.Fprintln(osStderr, "Example: follyo buy add BTC 0.5 50000")
	fmt.Fprintln(osStderr, "     or: follyo buy add BTC 0.5 --total 25000")
	osExit(1)
	return 0 // unreachable, but needed for compiler
}

// handleRemoveByID handles the common pattern of removing an item by ID.
// It calls the remover function and prints appropriate success/failure messages.
// itemType is used in messages (e.g., "purchase", "sale", "loan", "stake").
func handleRemoveByID(id, itemType string, remover func(string) (bool, error)) {
	removed, err := remover(id)
	if err != nil {
		fmt.Fprintf(osStderr, "Error: %v\n", err)
		osExit(1)
	}
	if removed {
		fmt.Printf("Removed %s %s\n", itemType, id)
	} else {
		fmt.Printf("%s %s not found\n", cases.Title(language.English).String(itemType), id)
	}
}

// getPlatformWithDefault returns the platform flag value, or the default platform from config if not set.
func getPlatformWithDefault(platform string) string {
	if platform == "" {
		cfg := loadConfig()
		return cfg.GetDefaultPlatform()
	}
	return platform
}

// PriceFetchResult contains the results of a price fetch operation
type PriceFetchResult struct {
	Prices          map[string]float64
	UnmappedTickers []string
	IsOffline       bool
	Error           error
}

// fetchPricesForCoins fetches live prices for a list of coins, handling errors gracefully.
// Returns prices, unmapped tickers, and whether the fetch failed (offline mode).
func fetchPricesForCoins(coins []string) PriceFetchResult {
	result := PriceFetchResult{
		Prices: make(map[string]float64),
	}

	if len(coins) == 0 {
		return result
	}

	ps := prices.New()

	// Load custom mappings from config
	cfg := loadConfig()
	customMappings := cfg.GetAllTickerMappings()
	for ticker, geckoID := range customMappings {
		ps.AddCoinMapping(ticker, geckoID)
	}

	// Check for unmapped tickers before fetching
	result.UnmappedTickers = ps.GetUnmappedTickers(coins)

	// Fetch prices
	priceMap, err := ps.GetPrices(coins)
	if err != nil {
		result.Error = err
		result.IsOffline = true
		return result
	}

	result.Prices = priceMap
	return result
}

// collectAllCoins collects all unique coins from a portfolio summary.
func collectAllCoins(holdingsByCoin, stakesByCoin, loansByCoin, netByCoin map[string]float64) []string {
	allCoins := make(map[string]bool)
	for coin := range holdingsByCoin {
		allCoins[coin] = true
	}
	for coin := range stakesByCoin {
		allCoins[coin] = true
	}
	for coin := range loansByCoin {
		allCoins[coin] = true
	}
	for coin := range netByCoin {
		allCoins[coin] = true
	}

	coins := make([]string, 0, len(allCoins))
	for coin := range allCoins {
		coins = append(coins, coin)
	}
	sortStrings(coins)
	return coins
}
