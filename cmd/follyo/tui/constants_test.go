package tui

import "testing"

func TestInputCharLimits(t *testing.T) {
	// Verify char limits are positive and sensible
	tests := []struct {
		name  string
		value int
		min   int
		max   int
	}{
		{"InputCoinCharLimit", InputCoinCharLimit, 3, 20},
		{"InputAmountCharLimit", InputAmountCharLimit, 10, 30},
		{"InputPriceCharLimit", InputPriceCharLimit, 10, 30},
		{"InputPlatformCharLimit", InputPlatformCharLimit, 20, 100},
		{"InputNotesCharLimit", InputNotesCharLimit, 50, 500},
		{"InputRateCharLimit", InputRateCharLimit, 5, 20},
		{"InputGeckoIDCharLimit", InputGeckoIDCharLimit, 20, 100},
		{"InputSearchCharLimit", InputSearchCharLimit, 20, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value < tt.min {
				t.Errorf("%s = %d, want >= %d", tt.name, tt.value, tt.min)
			}
			if tt.value > tt.max {
				t.Errorf("%s = %d, want <= %d", tt.name, tt.value, tt.max)
			}
		})
	}
}

func TestInputWidths(t *testing.T) {
	// Verify widths are positive and sensible for terminal display
	tests := []struct {
		name  string
		value int
		min   int
		max   int
	}{
		{"InputCoinWidth", InputCoinWidth, 10, 30},
		{"InputAmountWidth", InputAmountWidth, 10, 30},
		{"InputPriceWidth", InputPriceWidth, 10, 30},
		{"InputPlatformWidth", InputPlatformWidth, 20, 50},
		{"InputNotesWidth", InputNotesWidth, 30, 60},
		{"InputRateWidth", InputRateWidth, 10, 30},
		{"InputGeckoIDWidth", InputGeckoIDWidth, 20, 50},
		{"InputSearchWidth", InputSearchWidth, 30, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value < tt.min {
				t.Errorf("%s = %d, want >= %d", tt.name, tt.value, tt.min)
			}
			if tt.value > tt.max {
				t.Errorf("%s = %d, want <= %d", tt.name, tt.value, tt.max)
			}
		})
	}
}

func TestDisplayTruncation(t *testing.T) {
	// Verify truncation lengths are consistent (truncate length + suffix length <= max display)
	suffixLen := len(TruncationSuffix)

	tests := []struct {
		name        string
		maxDisplay  int
		truncateLen int
	}{
		{"Platform short", PlatformDisplayMaxShort, PlatformTruncateLengthShort},
		{"Platform long", PlatformDisplayMaxLong, PlatformTruncateLengthLong},
		{"Note", NoteDisplayMax, NoteTruncateLength},
		{"ID", IDDisplayMax, IDTruncateLength},
		{"Name", NameDisplayMax, NameTruncateLength},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Truncated string (truncateLen) + suffix should fit within or equal to maxDisplay
			totalLen := tt.truncateLen + suffixLen
			if totalLen > tt.maxDisplay {
				t.Errorf("%s: truncate length (%d) + suffix (%d) = %d, exceeds max display (%d)",
					tt.name, tt.truncateLen, suffixLen, totalLen, tt.maxDisplay)
			}

			// Truncation should only happen for strings longer than maxDisplay
			if tt.truncateLen >= tt.maxDisplay {
				t.Errorf("%s: truncate length (%d) should be less than max display (%d)",
					tt.name, tt.truncateLen, tt.maxDisplay)
			}
		})
	}
}

func TestSeparatorWidths(t *testing.T) {
	// Verify separator widths are reasonable for terminal display
	tests := []struct {
		name  string
		value int
		min   int
		max   int
	}{
		{"SeparatorWidthBuy", SeparatorWidthBuy, 50, 120},
		{"SeparatorWidthSell", SeparatorWidthSell, 50, 120},
		{"SeparatorWidthStake", SeparatorWidthStake, 50, 120},
		{"SeparatorWidthLoan", SeparatorWidthLoan, 50, 120},
		{"SeparatorWidthTickerSearch", SeparatorWidthTickerSearch, 50, 100},
		{"SeparatorWidthTickerDefaults", SeparatorWidthTickerDefaults, 30, 80},
		{"SeparatorWidthSnapshots", SeparatorWidthSnapshots, 50, 120},
		{"SeparatorWidthSnapshotDetail", SeparatorWidthSnapshotDetail, 30, 80},
		{"SeparatorWidthSummary", SeparatorWidthSummary, 20, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value < tt.min {
				t.Errorf("%s = %d, want >= %d", tt.name, tt.value, tt.min)
			}
			if tt.value > tt.max {
				t.Errorf("%s = %d, want <= %d", tt.name, tt.value, tt.max)
			}
		})
	}
}

func TestDateDisplayLength(t *testing.T) {
	// YYYY-MM-DD format is exactly 10 characters
	if DateDisplayLength != 10 {
		t.Errorf("DateDisplayLength = %d, want 10 for YYYY-MM-DD format", DateDisplayLength)
	}
}

func TestTruncationSuffix(t *testing.T) {
	// Verify the suffix is the expected "..."
	expected := "..."
	if TruncationSuffix != expected {
		t.Errorf("TruncationSuffix = %q, want %q", TruncationSuffix, expected)
	}
}

func TestDefaultMappingsVisibleCount(t *testing.T) {
	// Should show a reasonable number of items for scrolling
	if DefaultMappingsVisibleCount < 5 {
		t.Errorf("DefaultMappingsVisibleCount = %d, want >= 5", DefaultMappingsVisibleCount)
	}
	if DefaultMappingsVisibleCount > 30 {
		t.Errorf("DefaultMappingsVisibleCount = %d, want <= 30", DefaultMappingsVisibleCount)
	}
}

func TestViewportHorizontalMargin(t *testing.T) {
	// Margin should be reasonable for padding
	if ViewportHorizontalMargin < 0 {
		t.Errorf("ViewportHorizontalMargin = %d, want >= 0", ViewportHorizontalMargin)
	}
	if ViewportHorizontalMargin > 10 {
		t.Errorf("ViewportHorizontalMargin = %d, want <= 10", ViewportHorizontalMargin)
	}
}
