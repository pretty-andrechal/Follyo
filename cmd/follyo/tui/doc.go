// Package tui provides the terminal user interface for Follyo using Bubble Tea.
//
// The TUI is built on the Bubble Tea framework (github.com/charmbracelet/bubbletea)
// and uses Lip Gloss for styling (github.com/charmbracelet/lipgloss).
//
// # Architecture
//
// The TUI follows the Elm architecture pattern:
//   - Model: Application state (see [App])
//   - Update: Message handlers that update state
//   - View: Render functions that produce terminal output
//
// # Views
//
// The application uses a view-based navigation system:
//   - Menu: Main navigation menu
//   - Summary: Portfolio overview with current values
//   - Buy/Sell: Manage holdings and sales
//   - Stake/Loan: Manage staking and loans
//   - Snapshots: Historical portfolio snapshots
//   - Ticker: Manage coin ticker mappings
//   - Settings: Application configuration
//
// # Key Bindings
//
// Common key bindings are defined in [KeyMap] and include:
//   - Arrow keys: Navigation
//   - Enter: Select/confirm
//   - Escape: Back/cancel
//   - q: Quit application
//   - a/n: Add new item
//   - d/x: Delete item
package tui
