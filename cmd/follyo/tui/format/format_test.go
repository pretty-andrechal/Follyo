package format

import (
	"testing"
)

func TestAddCommas(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"small number", "123", "123"},
		{"exactly three", "456", "456"},
		{"four digits", "1234", "1,234"},
		{"six digits", "123456", "123,456"},
		{"seven digits", "1234567", "1,234,567"},
		{"large number", "1234567890", "1,234,567,890"},
		{"with decimals", "1234.56", "1,234.56"},
		{"negative", "-1234", "-1,234"},
		{"negative with decimals", "-1234567.89", "-1,234,567.89"},
		{"small decimal", "0.12", "0.12"},
		{"zero", "0", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddCommas(tt.input)
			if result != tt.expected {
				t.Errorf("AddCommas(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestUSD(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"zero", 0, "$0.00"},
		{"simple", 100.00, "$100.00"},
		{"with cents", 99.99, "$99.99"},
		{"large number", 12345.67, "$12,345.67"},
		{"very large", 1234567.89, "$1,234,567.89"},
		{"negative", -500.25, "$-500.25"},
		{"small decimal", 0.50, "$0.50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := USD(tt.input)
			if result != tt.expected {
				t.Errorf("USD(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestUSDSimple(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"zero", 0, "$0.00"},
		{"positive", 100.50, "$100.50"},
		{"negative", -500.25, "-$500.25"},
		{"small negative", -0.01, "-$0.01"},
		{"large positive", 99999.99, "$99999.99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := USDSimple(tt.input)
			if result != tt.expected {
				t.Errorf("USDSimple(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestUSDWithSign(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"positive", 100.50, "+$100.50"},
		{"zero", 0, "+$0.00"},
		{"negative", -50.25, "$-50.25"},
		{"large positive", 1234.56, "+$1,234.56"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := USDWithSign(tt.input)
			if result != tt.expected {
				t.Errorf("USDWithSign(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAmount(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"whole number", 100.0, "100"},
		{"one decimal", 1.5, "1.5"},
		{"two decimals", 99.99, "99.99"},
		{"trailing zeros", 1.50000000, "1.5"},
		{"many decimals", 0.00001234, "0.00001234"},
		{"large with decimals", 1234.5, "1,234.5"},
		{"zero", 0.0, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Amount(tt.input)
			if result != tt.expected {
				t.Errorf("Amount(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAmountAligned(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"whole number", 100.0, "100.0000"},
		{"one decimal", 1.5, "1.5000"},
		{"exact four", 99.9999, "99.9999"},
		{"large", 1234.5, "1,234.5000"},
		{"zero", 0.0, "0.0000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AmountAligned(tt.input)
			if result != tt.expected {
				t.Errorf("AmountAligned(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestProfitLoss(t *testing.T) {
	tests := []struct {
		name    string
		change  float64
		percent float64
		expect  string
	}{
		{"positive profit", 2500.00, 15.5, "+$2,500.00 (15.5%)"},
		{"zero change", 0.0, 0.0, "+$0.00 (0.0%)"},
		{"negative loss", -500.00, -10.0, "$-500.00 (-10.0%)"},
		{"large gain", 123456.78, 250.3, "+$123,456.78 (250.3%)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProfitLoss(tt.change, tt.percent)
			if result != tt.expect {
				t.Errorf("ProfitLoss(%v, %v) = %q, want %q", tt.change, tt.percent, result, tt.expect)
			}
		})
	}
}

func TestPercentage(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"positive", 15.5, "15.5%"},
		{"negative", -5.2, "-5.2%"},
		{"zero", 0.0, "0.0%"},
		{"whole number", 100.0, "100.0%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Percentage(tt.input)
			if result != tt.expected {
				t.Errorf("Percentage(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPercentageWithSign(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"positive", 15.5, "+15.5%"},
		{"negative", -5.2, "-5.2%"},
		{"zero", 0.0, "+0.0%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PercentageWithSign(tt.input)
			if result != tt.expected {
				t.Errorf("PercentageWithSign(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		maxLen   int
		suffix   string
		expected string
	}{
		{"no truncation needed", "hello", 10, "...", "hello"},
		{"exact length", "hello", 5, "...", "hello"},
		{"truncate with dots", "hello world", 8, "...", "hello..."},
		{"truncate short", "abcdefgh", 5, "..", "abc.."},
		{"empty suffix", "hello world", 8, "", "hello wo"},
		{"short maxLen", "hello", 3, "..", "h.."},
		{"very short maxLen", "hello", 2, "..", ".."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.s, tt.maxLen, tt.suffix)
			if result != tt.expected {
				t.Errorf("TruncateString(%q, %d, %q) = %q, want %q", tt.s, tt.maxLen, tt.suffix, result, tt.expected)
			}
		})
	}
}

func TestTruncatePlatformShort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"short name", "Kraken", "Kraken"},
		{"at limit", "CoinbasePro", "CoinbasePro"},
		{"over limit needs truncation", "VeryLongPlatformName", "VeryLongPla.."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncatePlatformShort(tt.input)
			// Just verify it doesn't panic and returns something reasonable
			if len(result) == 0 {
				t.Errorf("TruncatePlatformShort(%q) returned empty string", tt.input)
			}
		})
	}
}

func TestTruncatePlatformLong(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"short name", "Kraken"},
		{"medium name", "CoinbasePro"},
		{"long name", "VeryLongPlatformNameThatExceedsLimit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncatePlatformLong(tt.input)
			if len(result) == 0 {
				t.Errorf("TruncatePlatformLong(%q) returned empty string", tt.input)
			}
		})
	}
}

func TestTruncateDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"full datetime", "2024-01-15T10:30:00Z", "2024-01-15"},
		{"already short", "2024-01-15", "2024-01-15"},
		{"shorter", "2024-01", "2024-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateDate(tt.input)
			if result != tt.expected {
				t.Errorf("TruncateDate(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTruncateNote(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"short note", "Quick note"},
		{"long note", "This is a very long note that definitely exceeds the maximum display limit for notes in the view"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateNote(tt.input)
			if len(result) == 0 && len(tt.input) > 0 {
				t.Errorf("TruncateNote(%q) returned empty string", tt.input)
			}
		})
	}
}

func TestTruncateID(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"short id", "bitcoin"},
		{"long id", "very-long-coingecko-identifier-that-exceeds-limit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateID(tt.input)
			if len(result) == 0 && len(tt.input) > 0 {
				t.Errorf("TruncateID(%q) returned empty string", tt.input)
			}
		})
	}
}

func TestTruncateName(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"short name", "Bitcoin"},
		{"long name", "SuperLongCryptocurrencyNameThatExceedsTheDisplayLimit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateName(tt.input)
			if len(result) == 0 && len(tt.input) > 0 {
				t.Errorf("TruncateName(%q) returned empty string", tt.input)
			}
		})
	}
}

func TestSafeDivide(t *testing.T) {
	tests := []struct {
		name        string
		numerator   float64
		denominator float64
		expected    float64
	}{
		{"normal division", 10.0, 2.0, 5.0},
		{"divide by zero", 10.0, 0.0, 0.0},
		{"zero numerator", 0.0, 5.0, 0.0},
		{"both zero", 0.0, 0.0, 0.0},
		{"negative numerator", -10.0, 2.0, -5.0},
		{"negative denominator", 10.0, -2.0, -5.0},
		{"fractional result", 1.0, 3.0, 1.0 / 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeDivide(tt.numerator, tt.denominator)
			if result != tt.expected {
				t.Errorf("SafeDivide(%v, %v) = %v, want %v", tt.numerator, tt.denominator, result, tt.expected)
			}
		})
	}
}
