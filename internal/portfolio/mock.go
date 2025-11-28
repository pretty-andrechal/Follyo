package portfolio

import "github.com/pretty-andrechal/follyo/internal/models"

// MockPortfolio is a test double for Portfolio that implements PortfolioService.
// It stores data in memory and allows for easy test setup and verification.
type MockPortfolio struct {
	Holdings []models.Holding
	Sales    []models.Sale
	Loans    []models.Loan
	Stakes   []models.Stake

	// Error fields to simulate failures
	HoldingsErr    error
	SalesErr       error
	LoansErr       error
	StakesErr      error
	SummaryErr     error
	AggregationErr error

	// Counters for verification
	AddHoldingCalls    int
	RemoveHoldingCalls int
	AddSaleCalls       int
	RemoveSaleCalls    int
	AddLoanCalls       int
	RemoveLoanCalls    int
	AddStakeCalls      int
	RemoveStakeCalls   int
}

// NewMockPortfolio creates a new mock portfolio for testing.
func NewMockPortfolio() *MockPortfolio {
	return &MockPortfolio{
		Holdings: make([]models.Holding, 0),
		Sales:    make([]models.Sale, 0),
		Loans:    make([]models.Loan, 0),
		Stakes:   make([]models.Stake, 0),
	}
}

// Holdings methods

func (m *MockPortfolio) AddHolding(coin string, amount, purchasePriceUSD float64, platform, notes, date string) (models.Holding, error) {
	m.AddHoldingCalls++
	if m.HoldingsErr != nil {
		return models.Holding{}, m.HoldingsErr
	}
	h := models.NewHolding(coin, amount, purchasePriceUSD, platform, notes, date)
	m.Holdings = append(m.Holdings, h)
	return h, nil
}

func (m *MockPortfolio) RemoveHolding(id string) (bool, error) {
	m.RemoveHoldingCalls++
	if m.HoldingsErr != nil {
		return false, m.HoldingsErr
	}
	for i, h := range m.Holdings {
		if h.ID == id {
			m.Holdings = append(m.Holdings[:i], m.Holdings[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

func (m *MockPortfolio) ListHoldings() ([]models.Holding, error) {
	if m.HoldingsErr != nil {
		return nil, m.HoldingsErr
	}
	return m.Holdings, nil
}

// Sales methods

func (m *MockPortfolio) AddSale(coin string, amount, sellPriceUSD float64, platform, notes, date string) (models.Sale, error) {
	m.AddSaleCalls++
	if m.SalesErr != nil {
		return models.Sale{}, m.SalesErr
	}
	s := models.NewSale(coin, amount, sellPriceUSD, platform, notes, date)
	m.Sales = append(m.Sales, s)
	return s, nil
}

func (m *MockPortfolio) RemoveSale(id string) (bool, error) {
	m.RemoveSaleCalls++
	if m.SalesErr != nil {
		return false, m.SalesErr
	}
	for i, s := range m.Sales {
		if s.ID == id {
			m.Sales = append(m.Sales[:i], m.Sales[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

func (m *MockPortfolio) ListSales() ([]models.Sale, error) {
	if m.SalesErr != nil {
		return nil, m.SalesErr
	}
	return m.Sales, nil
}

// Loans methods

func (m *MockPortfolio) AddLoan(coin string, amount float64, platform string, interestRate *float64, notes, date string) (models.Loan, error) {
	m.AddLoanCalls++
	if m.LoansErr != nil {
		return models.Loan{}, m.LoansErr
	}
	l := models.NewLoan(coin, amount, platform, interestRate, notes, date)
	m.Loans = append(m.Loans, l)
	return l, nil
}

func (m *MockPortfolio) RemoveLoan(id string) (bool, error) {
	m.RemoveLoanCalls++
	if m.LoansErr != nil {
		return false, m.LoansErr
	}
	for i, l := range m.Loans {
		if l.ID == id {
			m.Loans = append(m.Loans[:i], m.Loans[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

func (m *MockPortfolio) ListLoans() ([]models.Loan, error) {
	if m.LoansErr != nil {
		return nil, m.LoansErr
	}
	return m.Loans, nil
}

// Stakes methods

func (m *MockPortfolio) AddStake(coin string, amount float64, platform string, apy *float64, notes, date string) (models.Stake, error) {
	m.AddStakeCalls++
	if m.StakesErr != nil {
		return models.Stake{}, m.StakesErr
	}
	s := models.NewStake(coin, amount, platform, apy, notes, date)
	m.Stakes = append(m.Stakes, s)
	return s, nil
}

func (m *MockPortfolio) RemoveStake(id string) (bool, error) {
	m.RemoveStakeCalls++
	if m.StakesErr != nil {
		return false, m.StakesErr
	}
	for i, s := range m.Stakes {
		if s.ID == id {
			m.Stakes = append(m.Stakes[:i], m.Stakes[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

func (m *MockPortfolio) ListStakes() ([]models.Stake, error) {
	if m.StakesErr != nil {
		return nil, m.StakesErr
	}
	return m.Stakes, nil
}

// Summary methods

func (m *MockPortfolio) GetSummary() (Summary, error) {
	if m.SummaryErr != nil {
		return Summary{}, m.SummaryErr
	}

	holdingsByCoin, _ := m.GetCurrentHoldingsByCoin()
	loansByCoin, _ := m.GetLoansByCoin()
	stakesByCoin, _ := m.GetStakesByCoin()
	availableByCoin, _ := m.GetAvailableByCoin()
	netByCoin, _ := m.GetNetHoldingsByCoin()
	totalInvested, _ := m.GetTotalInvestedUSD()
	totalSold, _ := m.GetTotalSoldUSD()

	return Summary{
		TotalHoldingsCount: len(m.Holdings),
		TotalSalesCount:    len(m.Sales),
		TotalLoansCount:    len(m.Loans),
		TotalStakesCount:   len(m.Stakes),
		TotalInvestedUSD:   totalInvested,
		TotalSoldUSD:       totalSold,
		HoldingsByCoin:     holdingsByCoin,
		LoansByCoin:        loansByCoin,
		StakesByCoin:       stakesByCoin,
		AvailableByCoin:    availableByCoin,
		NetByCoin:          netByCoin,
	}, nil
}

// Aggregation methods

func (m *MockPortfolio) GetHoldingsByCoin() (map[string]float64, error) {
	if m.AggregationErr != nil {
		return nil, m.AggregationErr
	}
	result := make(map[string]float64)
	for _, h := range m.Holdings {
		result[h.Coin] += h.Amount
	}
	return result, nil
}

func (m *MockPortfolio) GetSalesByCoin() (map[string]float64, error) {
	if m.AggregationErr != nil {
		return nil, m.AggregationErr
	}
	result := make(map[string]float64)
	for _, s := range m.Sales {
		result[s.Coin] += s.Amount
	}
	return result, nil
}

func (m *MockPortfolio) GetCurrentHoldingsByCoin() (map[string]float64, error) {
	if m.AggregationErr != nil {
		return nil, m.AggregationErr
	}
	purchases, _ := m.GetHoldingsByCoin()
	sales, _ := m.GetSalesByCoin()

	allCoins := make(map[string]bool)
	for coin := range purchases {
		allCoins[coin] = true
	}
	for coin := range sales {
		allCoins[coin] = true
	}

	result := make(map[string]float64)
	for coin := range allCoins {
		result[coin] = purchases[coin] - sales[coin]
	}
	return result, nil
}

func (m *MockPortfolio) GetLoansByCoin() (map[string]float64, error) {
	if m.AggregationErr != nil {
		return nil, m.AggregationErr
	}
	result := make(map[string]float64)
	for _, l := range m.Loans {
		result[l.Coin] += l.Amount
	}
	return result, nil
}

func (m *MockPortfolio) GetStakesByCoin() (map[string]float64, error) {
	if m.AggregationErr != nil {
		return nil, m.AggregationErr
	}
	result := make(map[string]float64)
	for _, s := range m.Stakes {
		result[s.Coin] += s.Amount
	}
	return result, nil
}

func (m *MockPortfolio) GetAvailableByCoin() (map[string]float64, error) {
	if m.AggregationErr != nil {
		return nil, m.AggregationErr
	}
	currentHoldings, _ := m.GetCurrentHoldingsByCoin()
	stakes, _ := m.GetStakesByCoin()

	allCoins := make(map[string]bool)
	for coin := range currentHoldings {
		allCoins[coin] = true
	}
	for coin := range stakes {
		allCoins[coin] = true
	}

	result := make(map[string]float64)
	for coin := range allCoins {
		result[coin] = currentHoldings[coin] - stakes[coin]
	}
	return result, nil
}

func (m *MockPortfolio) GetNetHoldingsByCoin() (map[string]float64, error) {
	if m.AggregationErr != nil {
		return nil, m.AggregationErr
	}
	currentHoldings, _ := m.GetCurrentHoldingsByCoin()
	loans, _ := m.GetLoansByCoin()

	allCoins := make(map[string]bool)
	for coin := range currentHoldings {
		allCoins[coin] = true
	}
	for coin := range loans {
		allCoins[coin] = true
	}

	result := make(map[string]float64)
	for coin := range allCoins {
		result[coin] = currentHoldings[coin] - loans[coin]
	}
	return result, nil
}

func (m *MockPortfolio) GetTotalInvestedUSD() (float64, error) {
	if m.AggregationErr != nil {
		return 0, m.AggregationErr
	}
	var total float64
	for _, h := range m.Holdings {
		total += h.TotalValueUSD()
	}
	return total, nil
}

func (m *MockPortfolio) GetTotalSoldUSD() (float64, error) {
	if m.AggregationErr != nil {
		return 0, m.AggregationErr
	}
	var total float64
	for _, s := range m.Sales {
		total += s.TotalValueUSD()
	}
	return total, nil
}

// Snapshot methods

func (m *MockPortfolio) CreateSnapshot(prices map[string]float64, note string) (models.Snapshot, error) {
	if m.SummaryErr != nil {
		return models.Snapshot{}, m.SummaryErr
	}

	summary, err := m.GetSummary()
	if err != nil {
		return models.Snapshot{}, err
	}

	// Calculate values for each coin
	coinValues := make(map[string]models.CoinSnapshot)
	var holdingsValue float64

	for coin, amount := range summary.HoldingsByCoin {
		price := prices[coin]
		value := amount * price
		holdingsValue += value

		coinValues[coin] = models.CoinSnapshot{
			Amount: amount,
			Price:  price,
			Value:  value,
		}
	}

	// Calculate loans value
	var loansValue float64
	for coin, amount := range summary.LoansByCoin {
		price := prices[coin]
		loansValue += amount * price
	}

	netValue := holdingsValue - loansValue
	profitLoss := netValue - summary.TotalInvestedUSD + summary.TotalSoldUSD

	var profitPercent float64
	if summary.TotalInvestedUSD > 0 {
		profitPercent = (profitLoss / summary.TotalInvestedUSD) * 100
	}

	return models.Snapshot{
		ID:            "mock-snapshot",
		HoldingsValue: holdingsValue,
		LoansValue:    loansValue,
		NetValue:      netValue,
		TotalInvested: summary.TotalInvestedUSD,
		TotalSold:     summary.TotalSoldUSD,
		ProfitLoss:    profitLoss,
		ProfitPercent: profitPercent,
		CoinValues:    coinValues,
		Note:          note,
	}, nil
}

// Ensure MockPortfolio implements PortfolioService at compile time.
var _ PortfolioService = (*MockPortfolio)(nil)
