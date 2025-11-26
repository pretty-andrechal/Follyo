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

### Holdings

```bash
# Add a holding
./follyo holding add BTC 0.5 45000 -p "Ledger" -n "DCA purchase"

# List all holdings
./follyo holding list

# Remove a holding
./follyo holding remove <holding-id>
```

### Sales

```bash
# Add a sale
./follyo sale add BTC 0.1 55000 -p "Binance" -n "Taking profits"

# List all sales
./follyo sale list

# Remove a sale
./follyo sale remove <sale-id>
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
