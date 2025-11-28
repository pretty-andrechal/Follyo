// Package format provides formatting utilities for displaying values in the TUI.
//
// This package handles consistent formatting of:
//   - Currency values (USD amounts with proper decimal places)
//   - Percentages (profit/loss with signs)
//   - Dates (truncated for table display)
//   - Strings (truncation with ellipsis)
//
// # Currency Formatting
//
//   - [USD]: Full precision USD with $ prefix
//   - [USDSimple]: Simplified USD without cents for large values
//   - [ProfitLoss]: Profit/loss with percentage and +/- sign
//
// # String Formatting
//
//   - [TruncatePlatformShort], [TruncatePlatformLong]: Platform name truncation
//   - [TruncateNote]: Note text truncation
//   - [TruncateDate]: Date display truncation
//
// # Error Formatting
//
//   - [FormatError]: User-friendly error messages
package format
