# Follyo - Personal Crypto Portfolio Tracker

A CLI tool to track your cryptocurrency holdings, sales, loans, and stakes across platforms with live price tracking.

## Features

- Track coin purchases with price and platform info
- Track coin sales with sell price
- Track loans on platforms like Nexo, Celsius, etc.
- Track staked crypto with APY and platform info
- **Live price tracking** via CoinGecko API (enabled by default)
- **Profit/Loss calculation** with colored output (green/red)
- **Ticker mapping** to customize CoinGecko ID mappings
- **Historical snapshots** to track portfolio value over time
- **Interactive TUI** with keyboard navigation and live updates
- View current holdings (purchased - sold)
- View available coins (holdings - staked)
- View net holdings (holdings - loans)
- Net portfolio value calculation (holdings value - loans value)
- Validation: can only stake what you own
- Simple JSON-based storage
- Command aliases for faster usage
- **User preferences** for customizing default behavior

## Installation

Requires [Go](https://go.dev/doc/install) 1.21+.

```bash
# Build the binary
go build -o follyo ./cmd/follyo

# Or install to $GOPATH/bin
go install ./cmd/follyo
```

## Quick Start

```bash
# Record a purchase
follyo buy add BTC 0.5 45000

# View portfolio summary with live prices
follyo summary

# View summary without prices (faster)
follyo summary --no-prices

# Launch interactive TUI
follyo tui
```

## Interactive TUI

Launch an interactive terminal interface with keyboard navigation:

```bash
follyo tui
```

**Features:**
- Dark theme with purple accent colors
- Keyboard navigation (arrow keys or vim-style j/k)
- Mouse support
- Live price fetching with loading spinner
- Portfolio summary with real-time values
- Profit/loss display with color coding

**Keybindings:**
| Key | Action |
|-----|--------|
| `↑`/`k` | Move up |
| `↓`/`j` | Move down |
| `Enter` | Select |
| `Esc` | Go back |
| `r` | Refresh data |
| `q` | Quit |

**Menu Options:**
- Portfolio Summary - View holdings with live prices
- Buy - Manage purchases (coming soon)
- Sell - Manage sales (coming soon)
- Stake - Manage staking positions (coming soon)
- Loan - Manage loans (coming soon)
- Snapshots - Save and compare snapshots (coming soon)
- Settings - Configure preferences (coming soon)

## Usage

### Command Aliases

All main commands have short aliases for faster typing:

| Command    | Alias  |
|------------|--------|
| `buy`      | `b`    |
| `sell`     | `sl`   |
| `loan`     | `l`    |
| `stake`    | `st`   |
| `summary`  | `s`    |
| `ticker`   | `t`    |
| `config`   | `cfg`  |
| `snapshot` | `snap` |

### Buy (Purchases)

```bash
# Record a purchase (price per unit)
follyo buy add BTC 0.5 45000 -p "Ledger" -n "DCA purchase"

# Record a purchase (total cost) - calculates price per unit automatically
follyo buy add BTC 2.3 --total 170000

# Using alias
follyo b add ETH 10 3000

# List all purchases
follyo buy list

# Remove a purchase
follyo buy remove <id>
```

### Sell (Sales)

```bash
# Record a sale (price per unit)
follyo sell add BTC 0.1 55000 -p "Binance" -n "Taking profits"

# Record a sale (total amount) - calculates price per unit automatically
follyo sell add BTC 1.5 --total 120000

# Using alias
follyo sl add ETH 2 4000

# List all sales
follyo sell list

# Remove a sale
follyo sell remove <id>
```

### Loans

```bash
# Add a loan
follyo loan add USDT 5000 Nexo -r 6.9 -n "Credit line"

# Using alias
follyo l add USDC 10000 Celsius

# List all loans
follyo loan list

# Remove a loan
follyo loan remove <loan-id>
```

### Staking

```bash
# Stake crypto (validates you own enough)
follyo stake add ETH 5 Lido -a 4.5 -n "ETH staking"

# Using alias
follyo st add SOL 100 Marinade

# List all stakes
follyo stake list

# Remove a stake (unstake)
follyo stake remove <stake-id>
```

Note: You can only stake coins you actually own. The system validates that `holdings - sales - already_staked >= stake_amount`.

### Portfolio Summary

```bash
# View summary with live prices (default)
follyo summary

# Using alias
follyo s

# View summary without live prices
follyo summary --no-prices
```

The summary shows:
- Holdings by coin (what you actually own: purchased - sold)
- Staked by coin
- Available by coin (holdings - staked)
- Loans by coin
- Net holdings (holdings - loans)
- **Current value** based on live prices
- **Profit/Loss** with percentage (colored green/red in terminal)

### Portfolio Snapshots

Save point-in-time snapshots of your portfolio value to track performance over time:

```bash
# Save a snapshot of current portfolio value
follyo snapshot save

# Save with a note
follyo snapshot save -n "Before major purchase"

# Using alias
follyo snap save

# List all snapshots
follyo snapshot list

# Show details of a specific snapshot
follyo snapshot show <id>

# Compare a snapshot to current portfolio
follyo snapshot compare <id>

# Compare two snapshots
follyo snapshot compare <id1> <id2>

# Remove a snapshot
follyo snapshot remove <id>
```

Snapshots capture:
- Net portfolio value at point-in-time
- Individual coin values and prices
- Profit/loss calculation
- Total invested and sold amounts

This enables you to track portfolio growth, compare historical performance, and see how specific coins have changed over time.

### Ticker Mapping

Map your portfolio tickers to CoinGecko IDs for accurate price lookups:

```bash
# Search for a coin on CoinGecko
follyo ticker search bitcoin

# Search and interactively map to a ticker
follyo ticker search mute MUTE

# Manually map a ticker
follyo ticker map MUTE mute-io

# List all mappings (custom only)
follyo ticker list

# List all mappings including defaults
follyo ticker list --all

# Remove a custom mapping
follyo ticker unmap MUTE
```

68 common tickers are pre-mapped by default (BTC, ETH, SOL, etc.).

### Configuration

Customize Follyo's default behavior with the config command:

```bash
# View all settings
follyo config get

# View a specific setting
follyo config get prices

# Enable/disable live price fetching by default
follyo config set prices on
follyo config set prices off

# Enable/disable colored output
follyo config set colors on
follyo config set colors off

# Set a default platform for new entries
follyo config set platform Coinbase

# Clear the default platform
follyo config set platform clear
```

Available settings:
- `prices` - Enable/disable live price fetching by default (on/off)
- `colors` - Enable/disable colored output in terminal (on/off)
- `platform` - Set default platform for new buy/sell entries

## Data Storage

Portfolio data is stored in `data/portfolio.json` (relative to current directory).
Configuration (custom ticker mappings and user preferences) is stored in `data/config.json`.
Snapshots are stored in `data/snapshots.json`.

You can specify a custom data path with the `--data` flag:

```bash
follyo --data /path/to/portfolio.json summary
```

## Example Output

```
Fetching live prices...

=== PORTFOLIO SUMMARY ===

HOLDINGS BY COIN:
  BTC:      0.5000    @ $97,000.00  = $48,500.00
  ETH:     10.0000    @ $3,500.00   = $35,000.00

STAKED BY COIN:
  ETH:      5.0000    @ $3,500.00   = $17,500.00

AVAILABLE BY COIN (Holdings - Staked):
  BTC:      0.5000    @ $97,000.00  = $48,500.00
  ETH:      5.0000    @ $3,500.00   = $17,500.00

LOANS BY COIN:
  USDC:  5000.0000    @ $1.00       = $5,000.00

NET HOLDINGS (Holdings - Loans):
  BTC:     +0.5000    @ $97,000.00  = +$48,500.00
  ETH:    +10.0000    @ $3,500.00   = +$35,000.00
  USDC: -5000.0000    @ $1.00       = -$5,000.00

---------------------------
Total Holdings: 2
Total Sales: 0
Total Stakes: 1
Total Loans: 1
Total Invested: $52,500.00
Total Sold: $0.00

---------------------------
Holdings Value: $83,500.00
Loans Value:   -$5,000.00
Net Value:      $78,500.00
Profit/Loss:    +$26,000.00 (49.5%)
```

## Future Enhancements

- TUI: Add/edit/delete operations for all entry types
- TUI: Snapshots management view
- TUI: Settings and ticker mapping views
- Edit commands for existing entries
- Transaction fee tracking
- Export to CSV/JSON
- Interest calculations for loans
- Staking rewards tracking
