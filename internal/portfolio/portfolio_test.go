package portfolio

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pretty-andrechal/follyo/internal/storage"
)

func setupTestPortfolio(t *testing.T) (*Portfolio, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "follyo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dataPath := filepath.Join(tmpDir, "portfolio.json")
	s, err := storage.New(dataPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create storage: %v", err)
	}

	p := New(s)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return p, cleanup
}

func TestPortfolio_Holdings(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	// Add holdings
	h1, err := p.AddHolding("btc", 1.0, 50000, "Binance", "test", "2024-01-01")
	if err != nil {
		t.Fatalf("AddHolding failed: %v", err)
	}
	if h1.Coin != "BTC" {
		t.Errorf("expected coin to be uppercased to BTC, got %s", h1.Coin)
	}

	h2, err := p.AddHolding("ETH", 10, 3000, "", "", "")
	if err != nil {
		t.Fatalf("AddHolding failed: %v", err)
	}

	// List holdings
	holdings, err := p.ListHoldings()
	if err != nil {
		t.Fatalf("ListHoldings failed: %v", err)
	}
	if len(holdings) != 2 {
		t.Errorf("expected 2 holdings, got %d", len(holdings))
	}

	// Remove holding
	removed, err := p.RemoveHolding(h1.ID)
	if err != nil {
		t.Fatalf("RemoveHolding failed: %v", err)
	}
	if !removed {
		t.Error("expected holding to be removed")
	}

	holdings, err = p.ListHoldings()
	if err != nil {
		t.Fatalf("ListHoldings failed: %v", err)
	}
	if len(holdings) != 1 {
		t.Errorf("expected 1 holding, got %d", len(holdings))
	}

	// Remove non-existent
	removed, err = p.RemoveHolding("nonexistent")
	if err != nil {
		t.Fatalf("RemoveHolding failed: %v", err)
	}
	if removed {
		t.Error("expected holding not to be removed")
	}

	_ = h2 // silence unused warning
}

func TestPortfolio_Loans(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	// Add loans
	rate := 6.9
	l1, err := p.AddLoan("usdt", 5000, "Nexo", &rate, "credit", "2024-01-01")
	if err != nil {
		t.Fatalf("AddLoan failed: %v", err)
	}
	if l1.Coin != "USDT" {
		t.Errorf("expected coin to be uppercased to USDT, got %s", l1.Coin)
	}

	l2, err := p.AddLoan("BTC", 0.5, "Celsius", nil, "", "")
	if err != nil {
		t.Fatalf("AddLoan failed: %v", err)
	}

	// List loans
	loans, err := p.ListLoans()
	if err != nil {
		t.Fatalf("ListLoans failed: %v", err)
	}
	if len(loans) != 2 {
		t.Errorf("expected 2 loans, got %d", len(loans))
	}

	// Remove loan
	removed, err := p.RemoveLoan(l1.ID)
	if err != nil {
		t.Fatalf("RemoveLoan failed: %v", err)
	}
	if !removed {
		t.Error("expected loan to be removed")
	}

	loans, err = p.ListLoans()
	if err != nil {
		t.Fatalf("ListLoans failed: %v", err)
	}
	if len(loans) != 1 {
		t.Errorf("expected 1 loan, got %d", len(loans))
	}

	_ = l2
}

func TestPortfolio_Sales(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	// Add sales
	s1, err := p.AddSale("btc", 0.5, 55000, "Binance", "profit", "2024-01-01")
	if err != nil {
		t.Fatalf("AddSale failed: %v", err)
	}
	if s1.Coin != "BTC" {
		t.Errorf("expected coin to be uppercased to BTC, got %s", s1.Coin)
	}

	s2, err := p.AddSale("ETH", 5, 3500, "", "", "")
	if err != nil {
		t.Fatalf("AddSale failed: %v", err)
	}

	// List sales
	sales, err := p.ListSales()
	if err != nil {
		t.Fatalf("ListSales failed: %v", err)
	}
	if len(sales) != 2 {
		t.Errorf("expected 2 sales, got %d", len(sales))
	}

	// Remove sale
	removed, err := p.RemoveSale(s1.ID)
	if err != nil {
		t.Fatalf("RemoveSale failed: %v", err)
	}
	if !removed {
		t.Error("expected sale to be removed")
	}

	sales, err = p.ListSales()
	if err != nil {
		t.Fatalf("ListSales failed: %v", err)
	}
	if len(sales) != 1 {
		t.Errorf("expected 1 sale, got %d", len(sales))
	}

	_ = s2
}

func TestPortfolio_GetHoldingsByCoin(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	p.AddHolding("BTC", 1.0, 50000, "", "", "")
	p.AddHolding("BTC", 0.5, 55000, "", "", "")
	p.AddHolding("ETH", 10, 3000, "", "", "")

	byCoin, err := p.GetHoldingsByCoin()
	if err != nil {
		t.Fatalf("GetHoldingsByCoin failed: %v", err)
	}

	if byCoin["BTC"] != 1.5 {
		t.Errorf("expected BTC holdings 1.5, got %f", byCoin["BTC"])
	}
	if byCoin["ETH"] != 10 {
		t.Errorf("expected ETH holdings 10, got %f", byCoin["ETH"])
	}
}

func TestPortfolio_GetLoansByCoin(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	p.AddLoan("USDT", 5000, "Nexo", nil, "", "")
	p.AddLoan("USDT", 3000, "Celsius", nil, "", "")
	p.AddLoan("BTC", 0.1, "Nexo", nil, "", "")

	byCoin, err := p.GetLoansByCoin()
	if err != nil {
		t.Fatalf("GetLoansByCoin failed: %v", err)
	}

	if byCoin["USDT"] != 8000 {
		t.Errorf("expected USDT loans 8000, got %f", byCoin["USDT"])
	}
	if byCoin["BTC"] != 0.1 {
		t.Errorf("expected BTC loans 0.1, got %f", byCoin["BTC"])
	}
}

func TestPortfolio_GetSalesByCoin(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	p.AddSale("BTC", 0.3, 55000, "", "", "")
	p.AddSale("BTC", 0.2, 60000, "", "", "")
	p.AddSale("ETH", 5, 3500, "", "", "")

	byCoin, err := p.GetSalesByCoin()
	if err != nil {
		t.Fatalf("GetSalesByCoin failed: %v", err)
	}

	if byCoin["BTC"] != 0.5 {
		t.Errorf("expected BTC sales 0.5, got %f", byCoin["BTC"])
	}
	if byCoin["ETH"] != 5 {
		t.Errorf("expected ETH sales 5, got %f", byCoin["ETH"])
	}
}

func TestPortfolio_GetNetHoldingsByCoin(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	// Add holdings
	p.AddHolding("BTC", 1.0, 50000, "", "", "")
	p.AddHolding("ETH", 10, 3000, "", "", "")

	// Add sales
	p.AddSale("BTC", 0.3, 55000, "", "", "")

	// Add loans
	p.AddLoan("BTC", 0.1, "Nexo", nil, "", "")
	p.AddLoan("USDT", 5000, "Nexo", nil, "", "")

	net, err := p.GetNetHoldingsByCoin()
	if err != nil {
		t.Fatalf("GetNetHoldingsByCoin failed: %v", err)
	}

	// BTC: 1.0 - 0.3 - 0.1 = 0.6
	if net["BTC"] != 0.6 {
		t.Errorf("expected BTC net 0.6, got %f", net["BTC"])
	}

	// ETH: 10 - 0 - 0 = 10
	if net["ETH"] != 10 {
		t.Errorf("expected ETH net 10, got %f", net["ETH"])
	}

	// USDT: 0 - 0 - 5000 = -5000
	if net["USDT"] != -5000 {
		t.Errorf("expected USDT net -5000, got %f", net["USDT"])
	}
}

func TestPortfolio_GetTotalInvestedUSD(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	p.AddHolding("BTC", 1.0, 50000, "", "", "")   // 50000
	p.AddHolding("ETH", 10, 3000, "", "", "")     // 30000

	total, err := p.GetTotalInvestedUSD()
	if err != nil {
		t.Fatalf("GetTotalInvestedUSD failed: %v", err)
	}

	if total != 80000 {
		t.Errorf("expected total invested 80000, got %f", total)
	}
}

func TestPortfolio_GetTotalSoldUSD(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	p.AddSale("BTC", 0.5, 55000, "", "", "")   // 27500
	p.AddSale("ETH", 5, 3500, "", "", "")      // 17500

	total, err := p.GetTotalSoldUSD()
	if err != nil {
		t.Fatalf("GetTotalSoldUSD failed: %v", err)
	}

	if total != 45000 {
		t.Errorf("expected total sold 45000, got %f", total)
	}
}

func TestPortfolio_GetSummary(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	// Add some data
	p.AddHolding("BTC", 1.0, 50000, "", "", "")
	p.AddHolding("ETH", 10, 3000, "", "", "")
	p.AddSale("BTC", 0.3, 55000, "", "", "")
	p.AddLoan("USDT", 5000, "Nexo", nil, "", "")

	summary, err := p.GetSummary()
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}

	if summary.TotalHoldingsCount != 2 {
		t.Errorf("expected 2 holdings, got %d", summary.TotalHoldingsCount)
	}
	if summary.TotalSalesCount != 1 {
		t.Errorf("expected 1 sale, got %d", summary.TotalSalesCount)
	}
	if summary.TotalLoansCount != 1 {
		t.Errorf("expected 1 loan, got %d", summary.TotalLoansCount)
	}
	if summary.TotalInvestedUSD != 80000 {
		t.Errorf("expected invested 80000, got %f", summary.TotalInvestedUSD)
	}
	if summary.TotalSoldUSD != 16500 {
		t.Errorf("expected sold 16500, got %f", summary.TotalSoldUSD)
	}

	// Check holdings by coin
	if summary.HoldingsByCoin["BTC"] != 1.0 {
		t.Errorf("expected BTC holdings 1.0, got %f", summary.HoldingsByCoin["BTC"])
	}

	// Check sales by coin
	if summary.SalesByCoin["BTC"] != 0.3 {
		t.Errorf("expected BTC sales 0.3, got %f", summary.SalesByCoin["BTC"])
	}

	// Check loans by coin
	if summary.LoansByCoin["USDT"] != 5000 {
		t.Errorf("expected USDT loans 5000, got %f", summary.LoansByCoin["USDT"])
	}

	// Check net by coin
	if summary.NetByCoin["BTC"] != 0.7 {
		t.Errorf("expected BTC net 0.7, got %f", summary.NetByCoin["BTC"])
	}
}

func TestNew(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	if p == nil {
		t.Fatal("expected portfolio to be created")
	}
}
