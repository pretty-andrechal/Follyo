// Package models defines the core data structures for the Follyo portfolio tracker.
//
// This package contains the domain models used throughout the application:
//   - Holding: Represents a cryptocurrency purchase/holding
//   - Sale: Represents a cryptocurrency sale transaction
//   - Loan: Represents a borrowed cryptocurrency position
//   - Stake: Represents a staked cryptocurrency position
//   - Snapshot: Represents a point-in-time portfolio valuation
//   - CoinSnapshot: Represents a single coin's value within a snapshot
//
// All models include an ID field generated using [GenerateID] for unique identification.
// Coin symbols are normalized to uppercase for consistency.
package models
