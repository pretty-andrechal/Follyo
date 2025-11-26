# Follyo - Personal Crypto Portfolio Tracker

A simple CLI tool to track your cryptocurrency holdings and loans across platforms.

## Features

- Track coin purchases with price and platform info
- Track loans on platforms like Nexo, Celsius, etc.
- View net holdings (holdings minus loans)
- Simple JSON-based storage

## Installation

Requires [uv](https://docs.astral.sh/uv/getting-started/installation/).

```bash
# Sync dependencies
uv sync
```

## Usage

### Holdings

```bash
# Add a holding
uv run follyo holding add BTC 0.5 45000 --platform "Ledger" --notes "DCA purchase"

# List all holdings
uv run follyo holding list

# Remove a holding
uv run follyo holding remove <holding-id>
```

### Loans

```bash
# Add a loan
uv run follyo loan add USDT 5000 Nexo --rate 6.9 --notes "Credit line"

# List all loans
uv run follyo loan list

# Remove a loan
uv run follyo loan remove <loan-id>
```

### Portfolio Summary

```bash
# View summary with net holdings
uv run follyo summary
```

## Data Storage

Portfolio data is stored in `data/portfolio.json`.

## Future Enhancements

- Live price fetching
- Profit/Loss calculations
- Multiple portfolios
- Export to CSV
- Interest calculations for loans
