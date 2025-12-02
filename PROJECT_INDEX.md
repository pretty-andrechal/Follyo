# Project Index: Follyo

**Generated**: 2025-12-02
**Type**: Go CLI Application
**Purpose**: Personal Cryptocurrency Portfolio Tracker

## Project Structure

```
follyo/
├── cmd/follyo/           # CLI entry point and commands
│   ├── main.go           # Application entry point
│   ├── tui.go            # TUI launcher command
│   ├── buy.go            # Buy command (add/list/remove)
│   ├── sell.go           # Sell command (add/list/remove)
│   ├── stake.go          # Stake command (add/list/remove)
│   ├── loan.go           # Loan command (add/list/remove)
│   ├── snapshot.go       # Snapshot command (save/list/show/compare/remove/daily)
│   ├── ticker.go         # Ticker mapping command (map/unmap/list/search)
│   ├── config.go         # Config command (get/set)
│   ├── summary.go        # Summary command
│   ├── watch.go          # Watch command (live updates)
│   ├── helpers.go        # CLI helper functions
│   └── tui/              # TUI implementation
│       ├── app.go        # Main TUI application (Bubbletea)
│       ├── styles.go     # UI colors and styles (Lipgloss)
│       ├── keys.go       # Keyboard bindings
│       ├── messages.go   # TUI message types
│       ├── registry.go   # View registry
│       ├── components/   # Reusable TUI components
│       │   ├── form.go       # Form input handling
│       │   ├── table.go      # Table rendering
│       │   ├── help.go       # Help bar
│       │   └── navigation.go # Navigation breadcrumbs
│       ├── format/       # Display formatting
│       │   ├── format.go     # USD/percentage formatting
│       │   └── errors.go     # Error message formatting
│       └── views/        # TUI view implementations
│           ├── menu.go           # Main menu
│           ├── summary.go        # Portfolio summary
│           ├── buy.go            # Buy management
│           ├── sell.go           # Sell management
│           ├── stake.go          # Stake management
│           ├── loan.go           # Loan management
│           ├── snapshots.go      # Snapshot management
│           ├── coinhistory.go    # Coin history tracking
│           ├── coinhistory_charts.go  # ASCII charts
│           ├── coinhistory_tables.go  # Data tables
│           ├── ticker.go         # Ticker mappings
│           ├── settings.go       # Settings management
│           └── entity_view.go    # Generic CRUD view
├── internal/             # Internal packages
│   ├── models/           # Data models
│   │   ├── models.go     # Holding, Sale, Stake, Loan structs
│   │   └── snapshot.go   # Snapshot, CoinSnapshot structs
│   ├── portfolio/        # Business logic
│   │   ├── portfolio.go  # Portfolio operations
│   │   ├── interfaces.go # Interface definitions
│   │   ├── snapshot.go   # Snapshot creation
│   │   ├── errors.go     # Custom errors
│   │   └── mock.go       # Test mocks
│   ├── storage/          # Data persistence
│   │   ├── storage.go    # JSON file storage
│   │   ├── crud.go       # Generic CRUD operations
│   │   └── snapshots.go  # Snapshot store
│   ├── prices/           # Price fetching
│   │   └── prices.go     # CoinGecko API integration
│   └── config/           # User configuration
│       └── config.go     # Config storage and defaults
└── data/                 # Runtime data (gitignored)
    ├── portfolio.json    # Holdings, sales, stakes, loans
    ├── config.json       # User settings, ticker mappings
    └── snapshots.json    # Historical snapshots
```

## Entry Points

| Entry | Path | Description |
|-------|------|-------------|
| CLI | `cmd/follyo/main.go` | Cobra-based CLI with subcommands |
| TUI | `cmd/follyo/tui/app.go` | Bubbletea interactive interface |
| Build | `go build -o follyo ./cmd/follyo` | Produces `follyo` binary |

## Core Modules

### models
- **Path**: `internal/models/`
- **Purpose**: Data structures with validation
- **Exports**: `Holding`, `Sale`, `Stake`, `Loan`, `Snapshot`, `CoinSnapshot`
- **Features**: UUID generation, date parsing, field validation

### portfolio
- **Path**: `internal/portfolio/`
- **Purpose**: Business logic layer
- **Exports**: `Portfolio`, `Summary`, various interfaces
- **Features**: CRUD for all entities, aggregations, staking validation

### storage
- **Path**: `internal/storage/`
- **Purpose**: JSON file persistence
- **Exports**: `Storage`, `SnapshotStore`
- **Features**: Generic CRUD, atomic writes, snapshot management

### prices
- **Path**: `internal/prices/`
- **Purpose**: Live price fetching
- **Exports**: `Fetcher`, `GetPrices()`
- **Features**: CoinGecko API, batch requests, caching

### config
- **Path**: `internal/config/`
- **Purpose**: User preferences and ticker mappings
- **Exports**: `Config`, `UserPreferences`
- **Features**: 68 default ticker mappings, preference storage

### tui
- **Path**: `cmd/follyo/tui/`
- **Purpose**: Interactive terminal interface
- **Framework**: Bubbletea + Lipgloss
- **Features**: Vim keybindings, forms, tables, ASCII charts

## Key Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| spf13/cobra | v1.10.1 | CLI framework |
| charmbracelet/bubbletea | v1.3.10 | TUI framework |
| charmbracelet/lipgloss | v1.1.0 | TUI styling |
| charmbracelet/bubbles | v0.21.0 | TUI components |
| guptarohit/asciigraph | v0.7.3 | ASCII chart rendering |
| google/uuid | v1.6.0 | ID generation |

## Test Coverage

| Package | Coverage |
|---------|----------|
| cmd/follyo | 47.9% |
| tui | 68.8% |
| tui/components | 97.3% |
| tui/format | 98.9% |
| tui/views | 78.5% |
| internal/config | 93.9% |
| internal/models | 92.0% |
| internal/portfolio | 78.2% |
| internal/prices | 93.3% |
| internal/storage | 91.9% |

**Files**: 58 source, 32 test

## CLI Commands

```
follyo [command]
├── tui              # Interactive TUI
├── summary (s)      # Portfolio summary
├── buy (b)          # Buy management
│   ├── add          # Add purchase
│   ├── list         # List purchases
│   └── remove       # Remove purchase
├── sell (sl)        # Sell management
│   ├── add          # Add sale
│   ├── list         # List sales
│   └── remove       # Remove sale
├── stake (st)       # Stake management
│   ├── add          # Add stake (validates ownership)
│   ├── list         # List stakes
│   └── remove       # Remove stake
├── loan (l)         # Loan management
│   ├── add          # Add loan
│   ├── list         # List loans
│   └── remove       # Remove loan
├── snapshot (snap)  # Snapshot management
│   ├── save         # Save snapshot
│   ├── daily        # Daily snapshot (YYYY-MM-DD note)
│   ├── list         # List snapshots
│   ├── show         # Show snapshot details
│   ├── compare      # Compare snapshots
│   └── remove       # Remove snapshot
├── ticker (t)       # Ticker mapping
│   ├── map          # Add mapping
│   ├── unmap        # Remove mapping
│   ├── list         # List mappings
│   └── search       # Search CoinGecko
├── config (cfg)     # Configuration
│   ├── get          # Get setting
│   └── set          # Set setting
└── watch            # Live price updates
```

## TUI Views

| View | Key | Features |
|------|-----|----------|
| Menu | - | Main navigation |
| Summary | - | Holdings, P/L, live prices |
| Buy | a/d | Add/delete purchases (7 fields: coin, amount, price, total, date, platform, notes) |
| Sell | a/d | Add/delete sales (7 fields: coin, amount, price, total, date, platform, notes) |
| Stake | a/d | Add/delete stakes |
| Loan | a/d | Add/delete loans |
| Snapshots | n/d/t | New/delete/today snapshot |
| Coin History | Enter | Price/holdings charts, comparison |
| Ticker | a/s/d | Add/search/delete mappings |
| Settings | Enter | Toggle settings |

## Quick Start

```bash
# Build
go build -o follyo ./cmd/follyo

# Add a purchase
./follyo buy add BTC 0.5 45000

# View summary
./follyo summary

# Launch TUI
./follyo tui

# Run tests
go test ./...
```

## Project Conventions

- **Issue Tracking**: bd (beads) - use `bd` commands, not markdown TODOs
- **Git**: Commit after each logical change, never leave untracked files
- **Testing**: Table-driven tests, mocks in `*_test.go` files
- **Errors**: Custom error types in `errors.go`
- **Docs**: Package docs in `doc.go` files
