// Package portfolio provides interfaces for portfolio management.
// These interfaces enable dependency injection and easier testing of
// components that depend on portfolio functionality.
package portfolio

import "github.com/pretty-andrechal/follyo/internal/models"

// HoldingsManager provides methods for managing coin holdings (purchases).
// Used by views that only need to work with holdings.
type HoldingsManager interface {
	// AddHolding adds a new coin holding.
	AddHolding(coin string, amount, purchasePriceUSD float64, platform, notes, date string) (models.Holding, error)
	// RemoveHolding removes a holding by ID.
	RemoveHolding(id string) (bool, error)
	// ListHoldings returns all holdings.
	ListHoldings() ([]models.Holding, error)
}

// SalesManager provides methods for managing coin sales.
// Used by views that only need to work with sales.
type SalesManager interface {
	// AddSale adds a new sale.
	AddSale(coin string, amount, sellPriceUSD float64, platform, notes, date string) (models.Sale, error)
	// RemoveSale removes a sale by ID.
	RemoveSale(id string) (bool, error)
	// ListSales returns all sales.
	ListSales() ([]models.Sale, error)
}

// LoansManager provides methods for managing loans.
// Used by views that only need to work with loans.
type LoansManager interface {
	// AddLoan adds a new loan.
	AddLoan(coin string, amount float64, platform string, interestRate *float64, notes, date string) (models.Loan, error)
	// RemoveLoan removes a loan by ID.
	RemoveLoan(id string) (bool, error)
	// ListLoans returns all loans.
	ListLoans() ([]models.Loan, error)
}

// StakesManager provides methods for managing stakes.
// Used by views that only need to work with stakes.
type StakesManager interface {
	// AddStake adds a new stake with validation that you can only stake what you own.
	AddStake(coin string, amount float64, platform string, apy *float64, notes, date string) (models.Stake, error)
	// RemoveStake removes a stake by ID.
	RemoveStake(id string) (bool, error)
	// ListStakes returns all stakes.
	ListStakes() ([]models.Stake, error)
}

// SummaryProvider provides portfolio summary data.
// Used by views that need to display portfolio overview.
type SummaryProvider interface {
	// GetSummary returns a portfolio summary.
	GetSummary() (Summary, error)
}

// SnapshotCreator provides snapshot creation capability.
// Used by views that need to create portfolio snapshots.
type SnapshotCreator interface {
	// CreateSnapshot creates a new portfolio snapshot with the given prices.
	CreateSnapshot(prices map[string]float64, note string) (models.Snapshot, error)
}

// AggregationProvider provides aggregated portfolio data by coin.
// Used by components that need detailed portfolio breakdowns.
type AggregationProvider interface {
	// GetHoldingsByCoin returns total holdings aggregated by coin.
	GetHoldingsByCoin() (map[string]float64, error)
	// GetSalesByCoin returns total sales aggregated by coin.
	GetSalesByCoin() (map[string]float64, error)
	// GetCurrentHoldingsByCoin returns current holdings (purchases - sales) by coin.
	GetCurrentHoldingsByCoin() (map[string]float64, error)
	// GetLoansByCoin returns total loans aggregated by coin.
	GetLoansByCoin() (map[string]float64, error)
	// GetStakesByCoin returns total stakes aggregated by coin.
	GetStakesByCoin() (map[string]float64, error)
	// GetAvailableByCoin returns available coins (current holdings - staked) by coin.
	GetAvailableByCoin() (map[string]float64, error)
	// GetNetHoldingsByCoin returns net holdings (current holdings - loans) by coin.
	GetNetHoldingsByCoin() (map[string]float64, error)
	// GetTotalInvestedUSD returns total USD invested in holdings.
	GetTotalInvestedUSD() (float64, error)
	// GetTotalSoldUSD returns total USD received from sales.
	GetTotalSoldUSD() (float64, error)
}

// SnapshotsManager provides snapshot-related portfolio operations.
// Used by views that need to display summaries and create snapshots.
type SnapshotsManager interface {
	SummaryProvider
	SnapshotCreator
}

// PortfolioService combines all portfolio management interfaces.
// Use this when a component needs full portfolio access.
type PortfolioService interface {
	HoldingsManager
	SalesManager
	LoansManager
	StakesManager
	SummaryProvider
	SnapshotCreator
	AggregationProvider
}

// Ensure *Portfolio implements PortfolioService at compile time.
var _ PortfolioService = (*Portfolio)(nil)
