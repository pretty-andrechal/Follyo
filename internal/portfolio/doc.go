// Package portfolio provides the core business logic for managing cryptocurrency portfolios.
//
// The Portfolio type is the main entry point, offering methods to:
//   - Add, list, and remove holdings (purchases)
//   - Add, list, and remove sales
//   - Add, list, and remove loans
//   - Add, list, and remove stakes
//   - Calculate portfolio summaries and aggregations
//   - Create point-in-time snapshots for historical tracking
//
// # Summary Calculations
//
// The package tracks several key metrics:
//   - HoldingsByCoin: Total purchased amount per coin
//   - SalesByCoin: Total sold amount per coin
//   - LoansByCoin: Total borrowed amount per coin
//   - StakesByCoin: Total staked amount per coin
//   - AvailableByCoin: Holdings minus sales minus stakes (liquid balance)
//   - NetByCoin: Holdings minus sales minus loans (net exposure)
//
// # Staking Validation
//
// Stakes are validated against available balance. You cannot stake more than
// your current holdings minus sales minus existing stakes.
//
// # Snapshots
//
// Snapshots capture the portfolio value at a point in time using provided prices.
// They calculate holdings value, loans value, net value, and profit/loss.
// Use [CompareSnapshots] to analyze changes between two snapshots.
package portfolio
