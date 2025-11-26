package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pretty-andrechal/follyo/internal/models"
)

func setupTestStorage(t *testing.T) (*Storage, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "follyo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dataPath := filepath.Join(tmpDir, "portfolio.json")
	s, err := New(dataPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create storage: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return s, cleanup
}

func TestNew(t *testing.T) {
	s, cleanup := setupTestStorage(t)
	defer cleanup()

	if s == nil {
		t.Fatal("expected storage to be created")
	}
}

func TestStorage_Holdings(t *testing.T) {
	s, cleanup := setupTestStorage(t)
	defer cleanup()

	// Initially empty
	holdings, err := s.GetHoldings()
	if err != nil {
		t.Fatalf("GetHoldings failed: %v", err)
	}
	if len(holdings) != 0 {
		t.Errorf("expected 0 holdings, got %d", len(holdings))
	}

	// Add a holding
	h1 := models.NewHolding("BTC", 1.0, 50000, "Binance", "test", "2024-01-01")
	err = s.AddHolding(h1)
	if err != nil {
		t.Fatalf("AddHolding failed: %v", err)
	}

	// Verify it was added
	holdings, err = s.GetHoldings()
	if err != nil {
		t.Fatalf("GetHoldings failed: %v", err)
	}
	if len(holdings) != 1 {
		t.Errorf("expected 1 holding, got %d", len(holdings))
	}
	if holdings[0].Coin != "BTC" {
		t.Errorf("expected BTC, got %s", holdings[0].Coin)
	}

	// Add another holding
	h2 := models.NewHolding("ETH", 10, 3000, "Ledger", "", "2024-01-02")
	err = s.AddHolding(h2)
	if err != nil {
		t.Fatalf("AddHolding failed: %v", err)
	}

	holdings, err = s.GetHoldings()
	if err != nil {
		t.Fatalf("GetHoldings failed: %v", err)
	}
	if len(holdings) != 2 {
		t.Errorf("expected 2 holdings, got %d", len(holdings))
	}

	// Remove first holding
	removed, err := s.RemoveHolding(h1.ID)
	if err != nil {
		t.Fatalf("RemoveHolding failed: %v", err)
	}
	if !removed {
		t.Error("expected holding to be removed")
	}

	holdings, err = s.GetHoldings()
	if err != nil {
		t.Fatalf("GetHoldings failed: %v", err)
	}
	if len(holdings) != 1 {
		t.Errorf("expected 1 holding, got %d", len(holdings))
	}
	if holdings[0].Coin != "ETH" {
		t.Errorf("expected ETH, got %s", holdings[0].Coin)
	}

	// Try to remove non-existent holding
	removed, err = s.RemoveHolding("nonexistent")
	if err != nil {
		t.Fatalf("RemoveHolding failed: %v", err)
	}
	if removed {
		t.Error("expected holding not to be removed")
	}
}

func TestStorage_Loans(t *testing.T) {
	s, cleanup := setupTestStorage(t)
	defer cleanup()

	// Initially empty
	loans, err := s.GetLoans()
	if err != nil {
		t.Fatalf("GetLoans failed: %v", err)
	}
	if len(loans) != 0 {
		t.Errorf("expected 0 loans, got %d", len(loans))
	}

	// Add a loan
	rate := 6.9
	l1 := models.NewLoan("USDT", 5000, "Nexo", &rate, "credit line", "2024-01-01")
	err = s.AddLoan(l1)
	if err != nil {
		t.Fatalf("AddLoan failed: %v", err)
	}

	// Verify it was added
	loans, err = s.GetLoans()
	if err != nil {
		t.Fatalf("GetLoans failed: %v", err)
	}
	if len(loans) != 1 {
		t.Errorf("expected 1 loan, got %d", len(loans))
	}
	if loans[0].Coin != "USDT" {
		t.Errorf("expected USDT, got %s", loans[0].Coin)
	}
	if loans[0].Platform != "Nexo" {
		t.Errorf("expected Nexo, got %s", loans[0].Platform)
	}

	// Add another loan
	l2 := models.NewLoan("BTC", 0.5, "Celsius", nil, "", "2024-01-02")
	err = s.AddLoan(l2)
	if err != nil {
		t.Fatalf("AddLoan failed: %v", err)
	}

	loans, err = s.GetLoans()
	if err != nil {
		t.Fatalf("GetLoans failed: %v", err)
	}
	if len(loans) != 2 {
		t.Errorf("expected 2 loans, got %d", len(loans))
	}

	// Remove first loan
	removed, err := s.RemoveLoan(l1.ID)
	if err != nil {
		t.Fatalf("RemoveLoan failed: %v", err)
	}
	if !removed {
		t.Error("expected loan to be removed")
	}

	loans, err = s.GetLoans()
	if err != nil {
		t.Fatalf("GetLoans failed: %v", err)
	}
	if len(loans) != 1 {
		t.Errorf("expected 1 loan, got %d", len(loans))
	}

	// Try to remove non-existent loan
	removed, err = s.RemoveLoan("nonexistent")
	if err != nil {
		t.Fatalf("RemoveLoan failed: %v", err)
	}
	if removed {
		t.Error("expected loan not to be removed")
	}
}

func TestStorage_Sales(t *testing.T) {
	s, cleanup := setupTestStorage(t)
	defer cleanup()

	// Initially empty
	sales, err := s.GetSales()
	if err != nil {
		t.Fatalf("GetSales failed: %v", err)
	}
	if len(sales) != 0 {
		t.Errorf("expected 0 sales, got %d", len(sales))
	}

	// Add a sale
	s1 := models.NewSale("BTC", 0.5, 55000, "Binance", "profit taking", "2024-01-01")
	err = s.AddSale(s1)
	if err != nil {
		t.Fatalf("AddSale failed: %v", err)
	}

	// Verify it was added
	sales, err = s.GetSales()
	if err != nil {
		t.Fatalf("GetSales failed: %v", err)
	}
	if len(sales) != 1 {
		t.Errorf("expected 1 sale, got %d", len(sales))
	}
	if sales[0].Coin != "BTC" {
		t.Errorf("expected BTC, got %s", sales[0].Coin)
	}

	// Add another sale
	s2 := models.NewSale("ETH", 5, 3500, "Kraken", "", "2024-01-02")
	err = s.AddSale(s2)
	if err != nil {
		t.Fatalf("AddSale failed: %v", err)
	}

	sales, err = s.GetSales()
	if err != nil {
		t.Fatalf("GetSales failed: %v", err)
	}
	if len(sales) != 2 {
		t.Errorf("expected 2 sales, got %d", len(sales))
	}

	// Remove first sale
	removed, err := s.RemoveSale(s1.ID)
	if err != nil {
		t.Fatalf("RemoveSale failed: %v", err)
	}
	if !removed {
		t.Error("expected sale to be removed")
	}

	sales, err = s.GetSales()
	if err != nil {
		t.Fatalf("GetSales failed: %v", err)
	}
	if len(sales) != 1 {
		t.Errorf("expected 1 sale, got %d", len(sales))
	}

	// Try to remove non-existent sale
	removed, err = s.RemoveSale("nonexistent")
	if err != nil {
		t.Fatalf("RemoveSale failed: %v", err)
	}
	if removed {
		t.Error("expected sale not to be removed")
	}
}

func TestDefaultDataPath(t *testing.T) {
	path := DefaultDataPath()
	if path == "" {
		t.Error("expected non-empty default data path")
	}
}
