// Package format provides shared formatting utilities for the Follyo TUI.
package format

import (
	"fmt"
	"strings"

	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
)

// AddCommas adds thousand separators to a numeric string.
func AddCommas(s string) string {
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

// USD formats a float64 as USD currency with thousand separators.
// Example: 1234.56 -> "$1,234.56"
func USD(amount float64) string {
	s := fmt.Sprintf("%.2f", amount)
	return "$" + AddCommas(s)
}

// USDSimple formats a float64 as USD without thousand separators.
// Handles negative values with the minus sign before the dollar sign.
// Example: -500.25 -> "-$500.25"
func USDSimple(amount float64) string {
	if amount < 0 {
		return fmt.Sprintf("-$%.2f", -amount)
	}
	return fmt.Sprintf("$%.2f", amount)
}

// USDWithSign formats USD with explicit + sign for positive values.
// Example: 100.50 -> "+$100.50", -50.25 -> "-$50.25"
func USDWithSign(amount float64) string {
	if amount >= 0 {
		return "+" + USD(amount)
	}
	return USD(amount)
}

// Amount formats a crypto amount, trimming unnecessary trailing zeros.
// Example: 1.50000000 -> "1.5", 100.00000000 -> "100"
func Amount(amount float64) string {
	s := fmt.Sprintf("%.8f", amount)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return AddCommas(s)
}

// AmountAligned formats amount with exactly 4 decimal places for alignment.
// Keeps trailing zeros to ensure decimal points line up.
// Example: 1.5 -> "1.5000"
func AmountAligned(amount float64) string {
	s := fmt.Sprintf("%.4f", amount)
	return AddCommas(s)
}

// ProfitLoss formats a profit/loss amount with sign and percentage.
// Example: (2500.00, 15.5) -> "+$2,500.00 (15.5%)"
func ProfitLoss(change, percent float64) string {
	sign := ""
	if change >= 0 {
		sign = "+"
	}
	return fmt.Sprintf("%s%s (%.1f%%)", sign, USD(change), percent)
}

// Percentage formats a float64 as a percentage with one decimal place.
// Example: 15.5 -> "15.5%"
func Percentage(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

// PercentageWithSign formats a percentage with explicit sign.
// Example: 15.5 -> "+15.5%", -5.2 -> "-5.2%"
func PercentageWithSign(value float64) string {
	if value >= 0 {
		return fmt.Sprintf("+%.1f%%", value)
	}
	return fmt.Sprintf("%.1f%%", value)
}

// TruncateString truncates a string to maxLen, adding suffix if truncated.
// If the string is shorter than or equal to maxLen, it is returned unchanged.
func TruncateString(s string, maxLen int, suffix string) string {
	if len(s) <= maxLen {
		return s
	}
	truncateAt := maxLen - len(suffix)
	if truncateAt < 0 {
		truncateAt = 0
	}
	return s[:truncateAt] + suffix
}

// TruncatePlatformShort truncates platform names for buy/sell views.
// Uses the short display constants from tui package.
func TruncatePlatformShort(platform string) string {
	if len(platform) > tui.PlatformDisplayMaxShort {
		return platform[:tui.PlatformTruncateLengthShort] + tui.TruncationSuffix
	}
	return platform
}

// TruncatePlatformLong truncates platform names for stake/loan views.
// Uses the long display constants from tui package.
func TruncatePlatformLong(platform string) string {
	if len(platform) > tui.PlatformDisplayMaxLong {
		return platform[:tui.PlatformTruncateLengthLong] + tui.TruncationSuffix
	}
	return platform
}

// TruncateDate truncates a date string to YYYY-MM-DD format.
func TruncateDate(date string) string {
	if len(date) > tui.DateDisplayLength {
		return date[:tui.DateDisplayLength]
	}
	return date
}

// TruncateNote truncates notes for list display.
func TruncateNote(note string) string {
	if len(note) > tui.NoteDisplayMax {
		return note[:tui.NoteTruncateLength] + tui.TruncationSuffix
	}
	return note
}

// TruncateID truncates IDs (like CoinGecko IDs) for display.
func TruncateID(id string) string {
	if len(id) > tui.IDDisplayMax {
		return id[:tui.IDTruncateLength] + tui.TruncationSuffix
	}
	return id
}

// TruncateName truncates names for search result display.
func TruncateName(name string) string {
	if len(name) > tui.NameDisplayMax {
		return name[:tui.NameTruncateLength] + tui.TruncationSuffix
	}
	return name
}

// SafeDivide performs division, returning 0 if denominator is 0.
func SafeDivide(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}
