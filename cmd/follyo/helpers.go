package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"golang.org/x/term"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// colorPreference caches the user's color preference (nil = not loaded)
var colorPreference *bool

// colorEnabled checks if color output should be used
func colorEnabled() bool {
	// Check if stdout is a terminal first
	isTerminal := false
	if f, ok := osStdout.(*os.File); ok {
		isTerminal = term.IsTerminal(int(f.Fd()))
	}
	if !isTerminal {
		return false
	}

	// Check user preference (lazy load)
	if colorPreference == nil {
		cfg := loadConfig()
		pref := cfg.GetColorOutput()
		colorPreference = &pref
	}
	return *colorPreference
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
