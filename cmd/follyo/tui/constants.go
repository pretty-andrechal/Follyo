// Package tui provides an interactive terminal user interface for Follyo.
package tui

// Input field character limits define the maximum number of characters
// a user can enter in form fields.
const (
	// InputCoinCharLimit is the max length for coin/ticker symbols (e.g., "BTC", "ETH")
	InputCoinCharLimit = 10

	// InputAmountCharLimit is the max length for numeric amount fields
	InputAmountCharLimit = 20

	// InputPriceCharLimit is the max length for price/rate fields
	InputPriceCharLimit = 20

	// InputPlatformCharLimit is the max length for platform names
	InputPlatformCharLimit = 50

	// InputNotesCharLimit is the max length for notes/comments
	InputNotesCharLimit = 200

	// InputRateCharLimit is the max length for APY/interest rate fields
	InputRateCharLimit = 10

	// InputGeckoIDCharLimit is the max length for CoinGecko ID fields
	InputGeckoIDCharLimit = 50

	// InputSearchCharLimit is the max length for search query fields
	InputSearchCharLimit = 50
)

// Input field display widths define how wide input fields appear in forms.
const (
	// InputCoinWidth is the display width for coin symbol inputs
	InputCoinWidth = 20

	// InputAmountWidth is the display width for amount inputs
	InputAmountWidth = 20

	// InputPriceWidth is the display width for price inputs
	InputPriceWidth = 20

	// InputPlatformWidth is the display width for platform inputs
	InputPlatformWidth = 30

	// InputNotesWidth is the display width for notes inputs
	InputNotesWidth = 40

	// InputRateWidth is the display width for rate inputs (APY, interest)
	InputRateWidth = 20

	// InputGeckoIDWidth is the display width for CoinGecko ID inputs
	InputGeckoIDWidth = 30

	// InputSearchWidth is the display width for search inputs
	InputSearchWidth = 40
)

// Display truncation constants control how text is shortened for display.
const (
	// DateDisplayLength is the length of dates in YYYY-MM-DD format
	DateDisplayLength = 10

	// PlatformDisplayMaxShort is the max display length for platforms in buy/sell views
	PlatformDisplayMaxShort = 12

	// PlatformTruncateLengthShort is where to truncate platforms in buy/sell views
	PlatformTruncateLengthShort = 9

	// PlatformDisplayMaxLong is the max display length for platforms in stake/loan views
	PlatformDisplayMaxLong = 20

	// PlatformTruncateLengthLong is where to truncate platforms in stake/loan views
	PlatformTruncateLengthLong = 17

	// NoteDisplayMax is the max display length for notes in list views
	NoteDisplayMax = 20

	// NoteTruncateLength is where to truncate notes
	NoteTruncateLength = 17

	// IDDisplayMax is the max display length for IDs (e.g., CoinGecko IDs)
	IDDisplayMax = 20

	// IDTruncateLength is where to truncate IDs
	IDTruncateLength = 17

	// NameDisplayMax is the max display length for names in search results
	NameDisplayMax = 25

	// NameTruncateLength is where to truncate names
	NameTruncateLength = 22

	// TruncationSuffix is appended to truncated strings
	TruncationSuffix = "..."
)

// Table separator widths define the width of horizontal separator lines
// in different list views.
const (
	// SeparatorWidthBuy is the separator width for the buy/holdings list
	SeparatorWidthBuy = 85

	// SeparatorWidthSell is the separator width for the sell/sales list
	SeparatorWidthSell = 85

	// SeparatorWidthStake is the separator width for the stake list
	SeparatorWidthStake = 75

	// SeparatorWidthLoan is the separator width for the loan list
	SeparatorWidthLoan = 75

	// SeparatorWidthTickerSearch is the separator width for ticker search results
	SeparatorWidthTickerSearch = 70

	// SeparatorWidthTickerDefaults is the separator width for default mappings list
	SeparatorWidthTickerDefaults = 50

	// SeparatorWidthSnapshots is the separator width for the snapshots list
	SeparatorWidthSnapshots = 90

	// SeparatorWidthSnapshotDetail is the separator width for snapshot details
	SeparatorWidthSnapshotDetail = 55

	// SeparatorWidthSummary is the separator width for summary sections
	SeparatorWidthSummary = 40
)

// Pagination constants control scrollable list behavior.
const (
	// DefaultMappingsVisibleCount is how many default ticker mappings to show at once
	DefaultMappingsVisibleCount = 15
)

// Viewport margins for scrollable content.
const (
	// ViewportHorizontalMargin is the horizontal padding for viewport content
	ViewportHorizontalMargin = 4
)
