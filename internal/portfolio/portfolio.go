// Package portfolio provides business logic for managing cryptocurrency portfolios.
// It handles holdings (purchases), sales, loans, and staking operations with
// proper validation and summary calculations.
package portfolio

import (
	"fmt"
	"strings"

	"github.com/pretty-andrechal/follyo/internal/models"
	"github.com/pretty-andrechal/follyo/internal/storage"
)

// Summary contains portfolio summary data.
type Summary struct {
	TotalHoldingsCount int
	TotalSalesCount    int
	TotalLoansCount    int
	TotalStakesCount   int
	TotalInvestedUSD   float64
	TotalSoldUSD       float64
	HoldingsByCoin     map[string]float64 // Current holdings: purchases - sales
	LoansByCoin        map[string]float64
	StakesByCoin       map[string]float64
	AvailableByCoin    map[string]float64 // Holdings - staked
	NetByCoin          map[string]float64 // Holdings - loans
}

// Portfolio manages crypto holdings, sales, and loans.
type Portfolio struct {
	storage *storage.Storage
}

// New creates a new Portfolio instance.
func New(s *storage.Storage) *Portfolio {
	return &Portfolio{storage: s}
}

// Holdings

// AddHolding adds a new coin holding.
func (p *Portfolio) AddHolding(coin string, amount, purchasePriceUSD float64, platform, notes, date string) (models.Holding, error) {
	// Validate inputs
	if err := models.ValidateCoinSymbol(coin); err != nil {
		return models.Holding{}, fmt.Errorf("invalid coin: %w", err)
	}
	if err := models.ValidateAmount(amount); err != nil {
		return models.Holding{}, fmt.Errorf("invalid amount: %w", err)
	}
	if err := models.ValidatePrice(purchasePriceUSD); err != nil {
		return models.Holding{}, fmt.Errorf("invalid price: %w", err)
	}
	if err := models.ValidateDate(date); err != nil {
		return models.Holding{}, fmt.Errorf("invalid date: %w", err)
	}
	if err := models.ValidateNotes(notes); err != nil {
		return models.Holding{}, fmt.Errorf("invalid notes: %w", err)
	}

	holding := models.NewHolding(strings.ToUpper(coin), amount, purchasePriceUSD, platform, notes, date)
	if err := p.storage.AddHolding(holding); err != nil {
		return models.Holding{}, fmt.Errorf("saving holding: %w", err)
	}
	return holding, nil
}

// RemoveHolding removes a holding by ID.
func (p *Portfolio) RemoveHolding(id string) (bool, error) {
	return p.storage.RemoveHolding(id)
}

// ListHoldings lists all holdings.
func (p *Portfolio) ListHoldings() ([]models.Holding, error) {
	return p.storage.GetHoldings()
}

// Loans

// AddLoan adds a new loan.
func (p *Portfolio) AddLoan(coin string, amount float64, platform string, interestRate *float64, notes, date string) (models.Loan, error) {
	// Validate inputs
	if err := models.ValidateCoinSymbol(coin); err != nil {
		return models.Loan{}, fmt.Errorf("invalid coin: %w", err)
	}
	if err := models.ValidateAmount(amount); err != nil {
		return models.Loan{}, fmt.Errorf("invalid amount: %w", err)
	}
	if platform == "" {
		return models.Loan{}, fmt.Errorf("platform is required for loans")
	}
	if err := models.ValidateDate(date); err != nil {
		return models.Loan{}, fmt.Errorf("invalid date: %w", err)
	}
	if err := models.ValidateNotes(notes); err != nil {
		return models.Loan{}, fmt.Errorf("invalid notes: %w", err)
	}

	loan := models.NewLoan(strings.ToUpper(coin), amount, platform, interestRate, notes, date)
	if err := p.storage.AddLoan(loan); err != nil {
		return models.Loan{}, fmt.Errorf("saving loan: %w", err)
	}
	return loan, nil
}

// RemoveLoan removes a loan by ID.
func (p *Portfolio) RemoveLoan(id string) (bool, error) {
	return p.storage.RemoveLoan(id)
}

// ListLoans lists all loans.
func (p *Portfolio) ListLoans() ([]models.Loan, error) {
	return p.storage.GetLoans()
}

// Sales

// AddSale adds a new sale.
// Validates that you can only sell coins that are available (not staked).
func (p *Portfolio) AddSale(coin string, amount, sellPriceUSD float64, platform, notes, date string) (models.Sale, error) {
	// Validate inputs
	if err := models.ValidateCoinSymbol(coin); err != nil {
		return models.Sale{}, fmt.Errorf("invalid coin: %w", err)
	}
	if err := models.ValidateAmount(amount); err != nil {
		return models.Sale{}, fmt.Errorf("invalid amount: %w", err)
	}
	if err := models.ValidatePrice(sellPriceUSD); err != nil {
		return models.Sale{}, fmt.Errorf("invalid price: %w", err)
	}
	if err := models.ValidateDate(date); err != nil {
		return models.Sale{}, fmt.Errorf("invalid date: %w", err)
	}
	if err := models.ValidateNotes(notes); err != nil {
		return models.Sale{}, fmt.Errorf("invalid notes: %w", err)
	}

	coin = strings.ToUpper(coin)

	// Validate that you have enough available (non-staked) coins to sell
	available, err := p.GetAvailableByCoin()
	if err != nil {
		return models.Sale{}, fmt.Errorf("checking available balance: %w", err)
	}

	availableAmount := available[coin]
	if amount > availableAmount {
		if availableAmount <= 0 {
			// Check if they have staked coins
			stakes, _ := p.GetStakesByCoin()
			if stakes[coin] > 0 {
				return models.Sale{}, fmt.Errorf("cannot sell %.8g %s: all your %s is staked (unstake first)", amount, coin, coin)
			}
			return models.Sale{}, fmt.Errorf("cannot sell %.8g %s: you have no %s to sell", amount, coin, coin)
		}
		// Check if the difference is due to staking
		stakes, _ := p.GetStakesByCoin()
		if stakes[coin] > 0 {
			return models.Sale{}, fmt.Errorf("cannot sell %.8g %s: only %.8g %s available (%.8g staked - unstake first to sell more)", amount, coin, availableAmount, coin, stakes[coin])
		}
		return models.Sale{}, fmt.Errorf("cannot sell %.8g %s: only %.8g %s available", amount, coin, availableAmount, coin)
	}

	sale := models.NewSale(coin, amount, sellPriceUSD, platform, notes, date)
	if err := p.storage.AddSale(sale); err != nil {
		return models.Sale{}, fmt.Errorf("saving sale: %w", err)
	}
	return sale, nil
}

// RemoveSale removes a sale by ID.
func (p *Portfolio) RemoveSale(id string) (bool, error) {
	return p.storage.RemoveSale(id)
}

// ListSales lists all sales.
func (p *Portfolio) ListSales() ([]models.Sale, error) {
	return p.storage.GetSales()
}

// Stakes

// AddStake adds a new stake with validation that you can only stake what you own.
func (p *Portfolio) AddStake(coin string, amount float64, platform string, apy *float64, notes, date string) (models.Stake, error) {
	// Validate inputs
	if err := models.ValidateCoinSymbol(coin); err != nil {
		return models.Stake{}, fmt.Errorf("invalid coin: %w", err)
	}
	if err := models.ValidateAmount(amount); err != nil {
		return models.Stake{}, fmt.Errorf("invalid amount: %w", err)
	}
	if platform == "" {
		return models.Stake{}, fmt.Errorf("platform is required for stakes")
	}
	if err := models.ValidateDate(date); err != nil {
		return models.Stake{}, fmt.Errorf("invalid date: %w", err)
	}
	if err := models.ValidateNotes(notes); err != nil {
		return models.Stake{}, fmt.Errorf("invalid notes: %w", err)
	}

	coin = strings.ToUpper(coin)

	// Calculate available balance for this coin
	available, err := p.GetAvailableByCoin()
	if err != nil {
		return models.Stake{}, fmt.Errorf("checking available balance: %w", err)
	}

	availableAmount := available[coin]
	if amount > availableAmount {
		if availableAmount <= 0 {
			return models.Stake{}, fmt.Errorf("cannot stake %.8g %s: you have no available %s to stake", amount, coin, coin)
		}
		return models.Stake{}, fmt.Errorf("cannot stake %.8g %s: only %.8g %s available (holdings - sales - already staked)", amount, coin, availableAmount, coin)
	}

	stake := models.NewStake(coin, amount, platform, apy, notes, date)
	if err := p.storage.AddStake(stake); err != nil {
		return models.Stake{}, fmt.Errorf("saving stake: %w", err)
	}
	return stake, nil
}

// RemoveStake removes a stake by ID.
func (p *Portfolio) RemoveStake(id string) (bool, error) {
	return p.storage.RemoveStake(id)
}

// ListStakes lists all stakes.
func (p *Portfolio) ListStakes() ([]models.Stake, error) {
	return p.storage.GetStakes()
}

// Summary methods

// GetHoldingsByCoin returns total holdings aggregated by coin.
func (p *Portfolio) GetHoldingsByCoin() (map[string]float64, error) {
	holdings, err := p.ListHoldings()
	if err != nil {
		return nil, err
	}

	byCoin := make(map[string]float64)
	for _, h := range holdings {
		byCoin[h.Coin] += h.Amount
	}
	return byCoin, nil
}

// GetLoansByCoin returns total loans aggregated by coin.
func (p *Portfolio) GetLoansByCoin() (map[string]float64, error) {
	loans, err := p.ListLoans()
	if err != nil {
		return nil, err
	}

	byCoin := make(map[string]float64)
	for _, l := range loans {
		byCoin[l.Coin] += l.Amount
	}
	return byCoin, nil
}

// GetSalesByCoin returns total sales aggregated by coin.
func (p *Portfolio) GetSalesByCoin() (map[string]float64, error) {
	sales, err := p.ListSales()
	if err != nil {
		return nil, err
	}

	byCoin := make(map[string]float64)
	for _, s := range sales {
		byCoin[s.Coin] += s.Amount
	}
	return byCoin, nil
}

// GetCurrentHoldingsByCoin returns current holdings (purchases - sales) by coin.
// This represents what you actually own right now.
func (p *Portfolio) GetCurrentHoldingsByCoin() (map[string]float64, error) {
	purchases, err := p.GetHoldingsByCoin()
	if err != nil {
		return nil, err
	}

	sales, err := p.GetSalesByCoin()
	if err != nil {
		return nil, err
	}

	// Collect all coins
	allCoins := make(map[string]bool)
	for coin := range purchases {
		allCoins[coin] = true
	}
	for coin := range sales {
		allCoins[coin] = true
	}

	current := make(map[string]float64)
	for coin := range allCoins {
		current[coin] = purchases[coin] - sales[coin]
	}
	return current, nil
}

// GetStakesByCoin returns total stakes aggregated by coin.
func (p *Portfolio) GetStakesByCoin() (map[string]float64, error) {
	stakes, err := p.ListStakes()
	if err != nil {
		return nil, err
	}

	byCoin := make(map[string]float64)
	for _, st := range stakes {
		byCoin[st.Coin] += st.Amount
	}
	return byCoin, nil
}

// GetAvailableByCoin returns available coins (current holdings - staked) by coin.
// This represents coins that you own and are not currently staked.
func (p *Portfolio) GetAvailableByCoin() (map[string]float64, error) {
	currentHoldings, err := p.GetCurrentHoldingsByCoin()
	if err != nil {
		return nil, err
	}

	stakes, err := p.GetStakesByCoin()
	if err != nil {
		return nil, err
	}

	// Collect all coins
	allCoins := make(map[string]bool)
	for coin := range currentHoldings {
		allCoins[coin] = true
	}
	for coin := range stakes {
		allCoins[coin] = true
	}

	available := make(map[string]float64)
	for coin := range allCoins {
		available[coin] = currentHoldings[coin] - stakes[coin]
	}
	return available, nil
}

// GetNetHoldingsByCoin returns net holdings (current holdings - loans) by coin.
// This represents what you'd have if all loans were paid back.
func (p *Portfolio) GetNetHoldingsByCoin() (map[string]float64, error) {
	currentHoldings, err := p.GetCurrentHoldingsByCoin()
	if err != nil {
		return nil, err
	}

	loans, err := p.GetLoansByCoin()
	if err != nil {
		return nil, err
	}

	// Collect all coins
	allCoins := make(map[string]bool)
	for coin := range currentHoldings {
		allCoins[coin] = true
	}
	for coin := range loans {
		allCoins[coin] = true
	}

	net := make(map[string]float64)
	for coin := range allCoins {
		net[coin] = currentHoldings[coin] - loans[coin]
	}
	return net, nil
}

// GetTotalInvestedUSD returns total USD invested in holdings.
func (p *Portfolio) GetTotalInvestedUSD() (float64, error) {
	holdings, err := p.ListHoldings()
	if err != nil {
		return 0, err
	}

	var total float64
	for _, h := range holdings {
		total += h.TotalValueUSD()
	}
	return total, nil
}

// GetTotalSoldUSD returns total USD received from sales.
func (p *Portfolio) GetTotalSoldUSD() (float64, error) {
	sales, err := p.ListSales()
	if err != nil {
		return 0, err
	}

	var total float64
	for _, s := range sales {
		total += s.TotalValueUSD()
	}
	return total, nil
}

// GetSummary returns a portfolio summary.
func (p *Portfolio) GetSummary() (Summary, error) {
	holdings, err := p.ListHoldings()
	if err != nil {
		return Summary{}, err
	}

	loans, err := p.ListLoans()
	if err != nil {
		return Summary{}, err
	}

	sales, err := p.ListSales()
	if err != nil {
		return Summary{}, err
	}

	stakes, err := p.ListStakes()
	if err != nil {
		return Summary{}, err
	}

	totalInvested, err := p.GetTotalInvestedUSD()
	if err != nil {
		return Summary{}, err
	}

	totalSold, err := p.GetTotalSoldUSD()
	if err != nil {
		return Summary{}, err
	}

	// Current holdings = purchases - sales (what you actually own)
	currentHoldingsByCoin, err := p.GetCurrentHoldingsByCoin()
	if err != nil {
		return Summary{}, err
	}

	loansByCoin, err := p.GetLoansByCoin()
	if err != nil {
		return Summary{}, err
	}

	stakesByCoin, err := p.GetStakesByCoin()
	if err != nil {
		return Summary{}, err
	}

	availableByCoin, err := p.GetAvailableByCoin()
	if err != nil {
		return Summary{}, err
	}

	netByCoin, err := p.GetNetHoldingsByCoin()
	if err != nil {
		return Summary{}, err
	}

	return Summary{
		TotalHoldingsCount: len(holdings),
		TotalSalesCount:    len(sales),
		TotalLoansCount:    len(loans),
		TotalStakesCount:   len(stakes),
		TotalInvestedUSD:   totalInvested,
		TotalSoldUSD:       totalSold,
		HoldingsByCoin:     currentHoldingsByCoin,
		LoansByCoin:        loansByCoin,
		StakesByCoin:       stakesByCoin,
		AvailableByCoin:    availableByCoin,
		NetByCoin:          netByCoin,
	}, nil
}
