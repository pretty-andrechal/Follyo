package portfolio

import (
	"errors"
	"testing"
)

func TestMockPortfolio_Holdings(t *testing.T) {
	mock := NewMockPortfolio()

	// Test AddHolding
	h, err := mock.AddHolding("BTC", 1.5, 50000, "Coinbase", "Test purchase", "2024-01-15")
	if err != nil {
		t.Fatalf("AddHolding failed: %v", err)
	}
	if h.Coin != "BTC" || h.Amount != 1.5 {
		t.Errorf("AddHolding returned wrong data: got %+v", h)
	}
	if mock.AddHoldingCalls != 1 {
		t.Errorf("AddHoldingCalls = %d, want 1", mock.AddHoldingCalls)
	}

	// Test ListHoldings
	holdings, err := mock.ListHoldings()
	if err != nil {
		t.Fatalf("ListHoldings failed: %v", err)
	}
	if len(holdings) != 1 {
		t.Errorf("ListHoldings returned %d items, want 1", len(holdings))
	}

	// Test RemoveHolding
	removed, err := mock.RemoveHolding(h.ID)
	if err != nil || !removed {
		t.Errorf("RemoveHolding failed: removed=%v, err=%v", removed, err)
	}
	if mock.RemoveHoldingCalls != 1 {
		t.Errorf("RemoveHoldingCalls = %d, want 1", mock.RemoveHoldingCalls)
	}

	holdings, _ = mock.ListHoldings()
	if len(holdings) != 0 {
		t.Errorf("After removal, ListHoldings returned %d items, want 0", len(holdings))
	}
}

func TestMockPortfolio_Sales(t *testing.T) {
	mock := NewMockPortfolio()

	// Test AddSale
	s, err := mock.AddSale("ETH", 2.0, 3000, "Binance", "Test sale", "2024-02-01")
	if err != nil {
		t.Fatalf("AddSale failed: %v", err)
	}
	if s.Coin != "ETH" || s.Amount != 2.0 {
		t.Errorf("AddSale returned wrong data: got %+v", s)
	}

	// Test ListSales
	sales, err := mock.ListSales()
	if err != nil {
		t.Fatalf("ListSales failed: %v", err)
	}
	if len(sales) != 1 {
		t.Errorf("ListSales returned %d items, want 1", len(sales))
	}

	// Test RemoveSale
	removed, err := mock.RemoveSale(s.ID)
	if err != nil || !removed {
		t.Errorf("RemoveSale failed: removed=%v, err=%v", removed, err)
	}
}

func TestMockPortfolio_Loans(t *testing.T) {
	mock := NewMockPortfolio()
	interestRate := 5.0

	// Test AddLoan
	l, err := mock.AddLoan("USDT", 1000, "Nexo", &interestRate, "Test loan", "2024-03-01")
	if err != nil {
		t.Fatalf("AddLoan failed: %v", err)
	}
	if l.Coin != "USDT" || l.Amount != 1000 {
		t.Errorf("AddLoan returned wrong data: got %+v", l)
	}

	// Test ListLoans
	loans, err := mock.ListLoans()
	if err != nil {
		t.Fatalf("ListLoans failed: %v", err)
	}
	if len(loans) != 1 {
		t.Errorf("ListLoans returned %d items, want 1", len(loans))
	}

	// Test RemoveLoan
	removed, err := mock.RemoveLoan(l.ID)
	if err != nil || !removed {
		t.Errorf("RemoveLoan failed: removed=%v, err=%v", removed, err)
	}
}

func TestMockPortfolio_Stakes(t *testing.T) {
	mock := NewMockPortfolio()
	apy := 8.5

	// Test AddStake
	s, err := mock.AddStake("SOL", 10, "Phantom", &apy, "Test stake", "2024-04-01")
	if err != nil {
		t.Fatalf("AddStake failed: %v", err)
	}
	if s.Coin != "SOL" || s.Amount != 10 {
		t.Errorf("AddStake returned wrong data: got %+v", s)
	}

	// Test ListStakes
	stakes, err := mock.ListStakes()
	if err != nil {
		t.Fatalf("ListStakes failed: %v", err)
	}
	if len(stakes) != 1 {
		t.Errorf("ListStakes returned %d items, want 1", len(stakes))
	}

	// Test RemoveStake
	removed, err := mock.RemoveStake(s.ID)
	if err != nil || !removed {
		t.Errorf("RemoveStake failed: removed=%v, err=%v", removed, err)
	}
}

func TestMockPortfolio_Summary(t *testing.T) {
	mock := NewMockPortfolio()

	// Add some test data
	mock.AddHolding("BTC", 1.0, 50000, "Coinbase", "", "2024-01-01")
	mock.AddHolding("ETH", 5.0, 3000, "Binance", "", "2024-01-02")
	mock.AddSale("BTC", 0.5, 55000, "Coinbase", "", "2024-02-01")
	interestRate := 5.0
	mock.AddLoan("USDT", 1000, "Nexo", &interestRate, "", "2024-03-01")
	apy := 8.0
	mock.AddStake("ETH", 2.0, "Lido", &apy, "", "2024-04-01")

	// Test GetSummary
	summary, err := mock.GetSummary()
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}

	if summary.TotalHoldingsCount != 2 {
		t.Errorf("TotalHoldingsCount = %d, want 2", summary.TotalHoldingsCount)
	}
	if summary.TotalSalesCount != 1 {
		t.Errorf("TotalSalesCount = %d, want 1", summary.TotalSalesCount)
	}
	if summary.TotalLoansCount != 1 {
		t.Errorf("TotalLoansCount = %d, want 1", summary.TotalLoansCount)
	}
	if summary.TotalStakesCount != 1 {
		t.Errorf("TotalStakesCount = %d, want 1", summary.TotalStakesCount)
	}

	// BTC holdings: 1.0 - 0.5 = 0.5
	if summary.HoldingsByCoin["BTC"] != 0.5 {
		t.Errorf("HoldingsByCoin[BTC] = %f, want 0.5", summary.HoldingsByCoin["BTC"])
	}

	// ETH holdings: 5.0 (no sales)
	if summary.HoldingsByCoin["ETH"] != 5.0 {
		t.Errorf("HoldingsByCoin[ETH] = %f, want 5.0", summary.HoldingsByCoin["ETH"])
	}

	// ETH available: 5.0 - 2.0 staked = 3.0
	if summary.AvailableByCoin["ETH"] != 3.0 {
		t.Errorf("AvailableByCoin[ETH] = %f, want 3.0", summary.AvailableByCoin["ETH"])
	}
}

func TestMockPortfolio_ErrorSimulation(t *testing.T) {
	mock := NewMockPortfolio()
	testErr := errors.New("simulated error")

	// Test holdings error
	mock.HoldingsErr = testErr
	_, err := mock.AddHolding("BTC", 1.0, 50000, "Test", "", "")
	if err != testErr {
		t.Errorf("Expected HoldingsErr, got %v", err)
	}
	_, err = mock.ListHoldings()
	if err != testErr {
		t.Errorf("Expected HoldingsErr on List, got %v", err)
	}
	mock.HoldingsErr = nil

	// Test sales error
	mock.SalesErr = testErr
	_, err = mock.AddSale("BTC", 1.0, 50000, "Test", "", "")
	if err != testErr {
		t.Errorf("Expected SalesErr, got %v", err)
	}
	mock.SalesErr = nil

	// Test loans error
	mock.LoansErr = testErr
	_, err = mock.AddLoan("BTC", 1.0, "Test", nil, "", "")
	if err != testErr {
		t.Errorf("Expected LoansErr, got %v", err)
	}
	mock.LoansErr = nil

	// Test stakes error
	mock.StakesErr = testErr
	_, err = mock.AddStake("BTC", 1.0, "Test", nil, "", "")
	if err != testErr {
		t.Errorf("Expected StakesErr, got %v", err)
	}
	mock.StakesErr = nil

	// Test summary error
	mock.SummaryErr = testErr
	_, err = mock.GetSummary()
	if err != testErr {
		t.Errorf("Expected SummaryErr, got %v", err)
	}
}

func TestMockPortfolio_ImplementsInterface(t *testing.T) {
	// This is a compile-time check that MockPortfolio implements PortfolioService
	var _ PortfolioService = (*MockPortfolio)(nil)

	// Also verify Portfolio implements the interface
	var _ PortfolioService = (*Portfolio)(nil)
}
