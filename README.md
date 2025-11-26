# Follyo - Personal Crypto Portfolio Tracker

A simple CLI tool to track your cryptocurrency holdings, sales, loans, and stakes across platforms.

## Features

- Track coin purchases with price and platform info
- Track coin sales with sell price
- Track loans on platforms like Nexo, Celsius, etc.
- Track staked crypto with APY and platform info
- View available coins (holdings - sales - staked)
- View net holdings (holdings - sales - loans)
- Validation: can only stake what you own
- Simple JSON-based storage

## Installation

Requires [Go](https://go.dev/doc/install) 1.21+.

```bash
# Build the binary
go build -o follyo ./cmd/follyo

# Or install to $GOPATH/bin
go install ./cmd/follyo
```

## Usage

### Buy (Purchases)

```bash
# Record a purchase
./follyo buy add BTC 0.5 45000 -p "Ledger" -n "DCA purchase"

# List all purchases
./follyo buy list

# Remove a purchase
./follyo buy remove <id>
```

### Sell (Sales)

```bash
# Record a sale
./follyo sell add BTC 0.1 55000 -p "Binance" -n "Taking profits"

# List all sales
./follyo sell list

# Remove a sale
./follyo sell remove <id>
```

### Loans

```bash
# Add a loan
./follyo loan add USDT 5000 Nexo -r 6.9 -n "Credit line"

# List all loans
./follyo loan list

# Remove a loan
./follyo loan remove <loan-id>
```

### Staking

```bash
# Stake crypto (validates you own enough)
./follyo stake add ETH 5 Lido -a 4.5 -n "ETH staking"

# List all stakes
./follyo stake list

# Remove a stake (unstake)
./follyo stake remove <stake-id>
```

Note: You can only stake coins you actually own. The system validates that `holdings - sales - already_staked >= stake_amount`.

### Portfolio Summary

```bash
# View summary with holdings, sales, staked, available, loans, and net holdings
./follyo summary
```

The summary shows:
- Holdings by coin (total purchased)
- Sales by coin
- Staked by coin
- Available by coin (holdings - sales - staked)
- Loans by coin
- Net holdings (holdings - sales - loans)

## Data Storage

Portfolio data is stored in `data/portfolio.json` (relative to current directory).

You can specify a custom path with the `--data` flag:

```bash
./follyo --data /path/to/portfolio.json summary
```

## Future Enhancements

- Live price fetching
- Profit/Loss calculations
- Multiple portfolios
- Export to CSV
- Interest calculations for loans
- Staking rewards tracking
