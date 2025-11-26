package portfolio

import (
	"strings"

	"github.com/pretty-andrechal/follyo/internal/models"
	"github.com/pretty-andrechal/follyo/internal/storage"
)

// Summary contains portfolio summary data.
type Summary struct {
	TotalHoldingsCount int
	TotalSalesCount    int
	TotalLoansCount    int
	TotalInvestedUSD   float64
	TotalSoldUSD       float64
	HoldingsByCoin     map[string]float64
	SalesByCoin        map[string]float64
	LoansByCoin        map[string]float64
	NetByCoin          map[string]float64
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
	holding := models.NewHolding(strings.ToUpper(coin), amount, purchasePriceUSD, platform, notes, date)
	err := p.storage.AddHolding(holding)
	return holding, err
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
	loan := models.NewLoan(strings.ToUpper(coin), amount, platform, interestRate, notes, date)
	err := p.storage.AddLoan(loan)
	return loan, err
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
func (p *Portfolio) AddSale(coin string, amount, sellPriceUSD float64, platform, notes, date string) (models.Sale, error) {
	sale := models.NewSale(strings.ToUpper(coin), amount, sellPriceUSD, platform, notes, date)
	err := p.storage.AddSale(sale)
	return sale, err
}

// RemoveSale removes a sale by ID.
func (p *Portfolio) RemoveSale(id string) (bool, error) {
	return p.storage.RemoveSale(id)
}

// ListSales lists all sales.
func (p *Portfolio) ListSales() ([]models.Sale, error) {
	return p.storage.GetSales()
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

// GetNetHoldingsByCoin returns net holdings (holdings - sales - loans) by coin.
func (p *Portfolio) GetNetHoldingsByCoin() (map[string]float64, error) {
	holdings, err := p.GetHoldingsByCoin()
	if err != nil {
		return nil, err
	}

	sales, err := p.GetSalesByCoin()
	if err != nil {
		return nil, err
	}

	loans, err := p.GetLoansByCoin()
	if err != nil {
		return nil, err
	}

	// Collect all coins
	allCoins := make(map[string]bool)
	for coin := range holdings {
		allCoins[coin] = true
	}
	for coin := range sales {
		allCoins[coin] = true
	}
	for coin := range loans {
		allCoins[coin] = true
	}

	net := make(map[string]float64)
	for coin := range allCoins {
		net[coin] = holdings[coin] - sales[coin] - loans[coin]
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

	totalInvested, err := p.GetTotalInvestedUSD()
	if err != nil {
		return Summary{}, err
	}

	totalSold, err := p.GetTotalSoldUSD()
	if err != nil {
		return Summary{}, err
	}

	holdingsByCoin, err := p.GetHoldingsByCoin()
	if err != nil {
		return Summary{}, err
	}

	salesByCoin, err := p.GetSalesByCoin()
	if err != nil {
		return Summary{}, err
	}

	loansByCoin, err := p.GetLoansByCoin()
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
		TotalInvestedUSD:   totalInvested,
		TotalSoldUSD:       totalSold,
		HoldingsByCoin:     holdingsByCoin,
		SalesByCoin:        salesByCoin,
		LoansByCoin:        loansByCoin,
		NetByCoin:          netByCoin,
	}, nil
}
