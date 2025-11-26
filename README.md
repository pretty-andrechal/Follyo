# Follyo - Personal Crypto Portfolio Tracker

A simple CLI tool to track your cryptocurrency holdings and loans across platforms.

## Features

- Track coin purchases with price and platform info
- Track loans on platforms like Nexo, Celsius, etc.
- View net holdings (holdings minus loans)
- Simple JSON-based storage

## Installation

```bash
pip install -e .
```

## Usage

### Holdings

```bash
# Add a holding
follyo holding add BTC 0.5 45000 --platform "Ledger" --notes "DCA purchase"

# List all holdings
follyo holding list

# Remove a holding
follyo holding remove <holding-id>
```

### Loans

```bash
# Add a loan
follyo loan add USDT 5000 Nexo --rate 6.9 --notes "Credit line"

# List all loans
follyo loan list

# Remove a loan
follyo loan remove <loan-id>
```

### Portfolio Summary

```bash
# View summary with net holdings
follyo summary
```

## Data Storage

Portfolio data is stored in `data/portfolio.json`.

## Future Enhancements

- Live price fetching
- Profit/Loss calculations
- Multiple portfolios
- Export to CSV
- Interest calculations for loans
