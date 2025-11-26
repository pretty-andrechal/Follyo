# Follyo - Personal Crypto Portfolio Tracker

A simple CLI tool to track your cryptocurrency holdings, sales, and loans across platforms.

## Features

- Track coin purchases with price and platform info
- Track coin sales with sell price
- Track loans on platforms like Nexo, Celsius, etc.
- View net holdings (holdings - sales - loans)
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

### Portfolio Summary

```bash
# View summary with net holdings
./follyo summary
```

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
