# Follyo Development Notes

> This file is a reference for Claude Code sessions to quickly understand the project state.

## Project Overview

**Follyo** is a CLI cryptocurrency portfolio tracker written in Go. It has two interfaces:
1. **CLI commands** - Traditional command-line interface (`follyo buy add BTC 1.5 50000`)
2. **TUI (Terminal UI)** - Interactive Bubble Tea-based interface (`follyo tui`)

## Directory Structure

```
cmd/follyo/
├── main.go              # Entry point, Cobra CLI setup
├── tui.go               # TUI command entry point
├── helpers.go           # CLI helper functions (formatUSD, addCommas, etc.)
├── tui/
│   ├── app.go           # Main TUI application (Bubble Tea model)
│   ├── app_test.go
│   ├── styles.go        # Lipgloss styles and colors
│   ├── keys.go          # Key bindings
│   ├── constants.go     # Named constants (Phase 1 refactoring)
│   ├── constants_test.go
│   ├── format/          # Formatting utilities (Phase 2 refactoring)
│   │   ├── format.go    # USD, Amount, Truncate functions
│   │   ├── format_test.go
│   │   ├── errors.go    # Error message formatting
│   │   └── errors_test.go
│   ├── components/      # Reusable TUI components (Phase 3 refactoring)
│   │   ├── form.go      # Form building (BuildFormInputs, FocusField, etc.)
│   │   ├── form_test.go
│   │   ├── table.go     # Table rendering (RenderTitle, RenderBox, etc.)
│   │   ├── table_test.go
│   │   ├── navigation.go # Cursor movement (MoveCursorUp/Down, ClampCursor)
│   │   ├── navigation_test.go
│   │   ├── help.go      # Help text rendering
│   │   └── help_test.go
│   └── views/           # Individual view models
│       ├── menu.go      # Main menu
│       ├── summary.go   # Portfolio summary with live prices
│       ├── buy.go       # Purchases CRUD (refactored to use components)
│       ├── sell.go      # Sales CRUD
│       ├── stake.go     # Staking CRUD
│       ├── loan.go      # Loans CRUD
│       ├── ticker.go    # CoinGecko ticker mappings
│       ├── snapshots.go # Historical portfolio snapshots
│       └── *_test.go    # Tests for each view

internal/
├── models/              # Data models (Holding, Sale, Stake, Loan, Snapshot)
├── portfolio/           # Portfolio business logic
├── storage/             # JSON file storage
├── prices/              # CoinGecko API integration
└── config/              # User preferences (default platform, etc.)
```

## Refactoring Progress

See `REFACTORING_PLAN.md` for full details. Current status:

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Constants and Configuration | ✅ COMPLETED |
| 2 | Shared Formatting Utilities | ✅ COMPLETED |
| 3 | Shared View Components | ✅ COMPLETED |
| 4 | Interfaces and Dependency Injection | ⏳ Pending |
| 5 | App.go Refactoring | ⏳ Pending |
| 6 | Test Coverage and Documentation | ⏳ Pending |

## Key Patterns

### TUI View Pattern
Each view follows a three-mode pattern:
```go
type ViewMode int
const (
    List ViewMode = iota  // Browse items
    Add                   // Form to add new item
    ConfirmDelete         // Confirm deletion
)
```

### Form Building (using components)
```go
fields := []components.FormField{
    {Placeholder: "BTC...", CharLimit: tui.InputCoinCharLimit, Width: tui.InputCoinWidth},
    // ...
}
inputs := components.BuildFormInputs(fields)
```

### Navigation (using components)
```go
m.cursor = components.MoveCursorUp(m.cursor, len(m.items))
m.cursor = components.MoveCursorDown(m.cursor, len(m.items))
```

### Formatting (using format package)
```go
format.USDSimple(1234.56)           // "$1234.56"
format.USD(1234.56)                 // "$1,234.56" (with commas)
format.TruncatePlatformShort(name)  // Truncates to constant length
format.SafeDivide(a, b)             // Returns 0 if b is 0
```

### Rendering (using components)
```go
components.RenderTitle("PURCHASES")
components.RenderEmptyState("No items yet")
components.RenderSeparator(tui.SeparatorWidthBuy)
components.RenderHelp(components.ListHelp(hasItems))
components.RenderBoxDefault(content)
```

## Important Constants

Located in `cmd/follyo/tui/constants.go`:
- Input field limits: `InputCoinCharLimit`, `InputAmountWidth`, etc.
- Display truncation: `PlatformDisplayMaxShort`, `DateDisplayLength`, etc.
- Separator widths: `SeparatorWidthBuy`, `SeparatorWidthSell`, etc.

## Testing

Run all tests:
```bash
go test ./...
```

Run specific package tests:
```bash
go test ./cmd/follyo/tui/components/... -v
go test ./cmd/follyo/tui/format/... -v
go test ./cmd/follyo/tui/views/... -v
```

## What's Been Done

1. **Phase 1**: Extracted magic numbers into `constants.go`
2. **Phase 2**: Created `format` package with USD formatting, truncation, safe division
3. **Phase 3**: Created `components` package with form, table, navigation, help utilities
4. **Refactored `buy.go`**: As example of using the new components (reduced ~80 lines)

## What's Left To Do

### Phase 4: Interfaces and DI
- Create `internal/portfolio/interfaces.go` with service interfaces
- Create mock implementations for testing
- Update views to use interfaces instead of `*portfolio.Portfolio`

### Phase 5: App.go Refactoring
- Create view registry to reduce boilerplate
- Consolidate menu action handling

### Phase 6: Tests and Docs
- Improve test coverage to >80%
- Add package documentation
- Create integration tests

### Optional: Apply Components to Other Views
The components from Phase 3 can be applied to:
- `sell.go` - Same pattern as buy.go
- `stake.go` - Similar form/list pattern
- `loan.go` - Similar form/list pattern
- `ticker.go` - Different but can use navigation/help
- `snapshots.go` - Can use table/navigation components

## Git Branch

Working branch: `claude/crypto-portfolio-tracker-01GaRjUPA3c83zTMFD3r8ZsW`
Main branch: `main`

## Recent Commits

```
bf6cd5b Complete Phase 3: Add shared view components package
b3a7bba Add TUI view files and complete Phases 1-2 of refactoring
5b99a1b Add Settings view to TUI (Sprint 3)
```

## Quick Commands

```bash
# Build
go build ./...

# Run TUI
go run . tui

# Run CLI
go run . buy list
go run . summary

# Test
go test ./...
```

## Notes for Future Sessions

1. **Always run tests** after making changes: `go test ./...`
2. **Check REFACTORING_PLAN.md** for detailed phase descriptions
3. **buy.go is the reference** for how to use the components package
4. **Constants are centralized** in `tui/constants.go` - add new ones there
5. **Format functions are centralized** in `tui/format/` - add new ones there
6. **The TUI uses Bubble Tea** - familiarize with Model-View-Update pattern
