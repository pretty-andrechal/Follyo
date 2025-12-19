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

func TestStorage_Stakes(t *testing.T) {
	s, cleanup := setupTestStorage(t)
	defer cleanup()

	// Initially empty
	stakes, err := s.GetStakes()
	if err != nil {
		t.Fatalf("GetStakes failed: %v", err)
	}
	if len(stakes) != 0 {
		t.Errorf("expected 0 stakes, got %d", len(stakes))
	}

	// Add a stake
	apy := 4.5
	st1 := models.NewStake("ETH", 10, "Lido", &apy, "staking rewards", "2024-03-01")
	err = s.AddStake(st1)
	if err != nil {
		t.Fatalf("AddStake failed: %v", err)
	}

	// Verify it was added
	stakes, err = s.GetStakes()
	if err != nil {
		t.Fatalf("GetStakes failed: %v", err)
	}
	if len(stakes) != 1 {
		t.Errorf("expected 1 stake, got %d", len(stakes))
	}
	if stakes[0].Coin != "ETH" {
		t.Errorf("expected ETH, got %s", stakes[0].Coin)
	}
	if stakes[0].Platform != "Lido" {
		t.Errorf("expected Lido, got %s", stakes[0].Platform)
	}

	// Add another stake
	st2 := models.NewStake("SOL", 100, "Coinbase", nil, "", "2024-03-02")
	err = s.AddStake(st2)
	if err != nil {
		t.Fatalf("AddStake failed: %v", err)
	}

	stakes, err = s.GetStakes()
	if err != nil {
		t.Fatalf("GetStakes failed: %v", err)
	}
	if len(stakes) != 2 {
		t.Errorf("expected 2 stakes, got %d", len(stakes))
	}

	// Remove first stake
	removed, err := s.RemoveStake(st1.ID)
	if err != nil {
		t.Fatalf("RemoveStake failed: %v", err)
	}
	if !removed {
		t.Error("expected stake to be removed")
	}

	stakes, err = s.GetStakes()
	if err != nil {
		t.Fatalf("GetStakes failed: %v", err)
	}
	if len(stakes) != 1 {
		t.Errorf("expected 1 stake, got %d", len(stakes))
	}

	// Try to remove non-existent stake
	removed, err = s.RemoveStake("nonexistent")
	if err != nil {
		t.Fatalf("RemoveStake failed: %v", err)
	}
	if removed {
		t.Error("expected stake not to be removed")
	}
}

func TestDefaultDataPath(t *testing.T) {
	path := DefaultDataPath()
	if path == "" {
		t.Error("expected non-empty default data path")
	}
}

// TestCorruptedJSONFile tests that corrupted JSON files are handled gracefully
func TestCorruptedJSONFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "follyo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataPath := filepath.Join(tmpDir, "portfolio.json")

	// Write corrupted JSON to file
	if err := os.WriteFile(dataPath, []byte(`{invalid json content`), 0600); err != nil {
		t.Fatalf("failed to write corrupted file: %v", err)
	}

	// Attempting to create storage should fail
	_, err = New(dataPath)
	if err == nil {
		t.Error("expected error when loading corrupted JSON file")
	}
}

// TestEmptyJSONFile tests that empty JSON files are handled
func TestEmptyJSONFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "follyo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataPath := filepath.Join(tmpDir, "portfolio.json")

	// Write empty file
	if err := os.WriteFile(dataPath, []byte(``), 0600); err != nil {
		t.Fatalf("failed to write empty file: %v", err)
	}

	// Storage should handle empty file (either error or initialize fresh)
	_, err = New(dataPath)
	// Empty file handling - either it errors or creates fresh data
	// The current implementation will error on invalid JSON
	if err == nil {
		// If no error, check that it initialized properly
		t.Log("Empty file was handled by initializing fresh data")
	}
}

// TestPartiallyCorruptedData tests behavior with valid JSON but invalid data types
func TestPartiallyCorruptedData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "follyo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataPath := filepath.Join(tmpDir, "portfolio.json")

	// Write valid JSON with wrong structure
	if err := os.WriteFile(dataPath, []byte(`{"holdings": "not_an_array"}`), 0600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err = New(dataPath)
	if err == nil {
		t.Error("expected error when loading JSON with wrong structure")
	}
}

// TestFilePermissionError tests behavior when file is not readable
func TestFilePermissionError(t *testing.T) {
	// Skip on Windows as permissions work differently
	if os.Getenv("OS") == "Windows_NT" {
		t.Skip("Skipping permission test on Windows")
	}

	tmpDir, err := os.MkdirTemp("", "follyo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataPath := filepath.Join(tmpDir, "portfolio.json")

	// Write valid data first
	validData := `{"holdings":[],"loans":[],"sales":[],"stakes":[]}`
	if err := os.WriteFile(dataPath, []byte(validData), 0600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Make file unreadable
	if err := os.Chmod(dataPath, 0000); err != nil {
		t.Fatalf("failed to change permissions: %v", err)
	}

	_, err = New(dataPath)
	if err == nil {
		t.Error("expected error when file is not readable")
	}

	// Restore permissions for cleanup
	os.Chmod(dataPath, 0600)
}

// TestNonExistentDirectory tests creating storage in a non-existent directory
func TestNonExistentDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "follyo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use a path in a non-existent subdirectory
	dataPath := filepath.Join(tmpDir, "nonexistent", "subdir", "portfolio.json")

	// Creating storage should create the directory structure
	s, err := New(dataPath)
	if err != nil {
		t.Fatalf("failed to create storage in new directory: %v", err)
	}

	// Verify it works
	h := models.NewHolding("BTC", 1.0, 50000, "", "", "")
	if err := s.AddHolding(h); err != nil {
		t.Fatalf("failed to add holding: %v", err)
	}
}

// TestLargeDataSet tests storage with many entries
func TestLargeDataSet(t *testing.T) {
	s, cleanup := setupTestStorage(t)
	defer cleanup()

	// Add 100 holdings
	for i := 0; i < 100; i++ {
		h := models.NewHolding("BTC", float64(i)+1, 50000, "", "", "")
		if err := s.AddHolding(h); err != nil {
			t.Fatalf("failed to add holding %d: %v", i, err)
		}
	}

	holdings, err := s.GetHoldings()
	if err != nil {
		t.Fatalf("failed to get holdings: %v", err)
	}
	if len(holdings) != 100 {
		t.Errorf("expected 100 holdings, got %d", len(holdings))
	}
}

// TestConcurrentAccess tests thread safety of storage operations
func TestConcurrentAccess(t *testing.T) {
	s, cleanup := setupTestStorage(t)
	defer cleanup()

	// Run concurrent reads and writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				h := models.NewHolding("BTC", float64(id*10+j), 50000, "", "", "")
				s.AddHolding(h)
				s.GetHoldings()
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	holdings, err := s.GetHoldings()
	if err != nil {
		t.Fatalf("failed to get holdings after concurrent access: %v", err)
	}
	if len(holdings) != 100 {
		t.Errorf("expected 100 holdings after concurrent access, got %d", len(holdings))
	}
}

// TestReload tests that Reload picks up external changes
func TestReload(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "follyo-reload-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataPath := filepath.Join(tmpDir, "portfolio.json")

	// Create first storage instance
	s1, err := New(dataPath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// Add a holding via s1
	h1 := models.NewHolding("BTC", 1.0, 50000, "Coinbase", "", "")
	if err := s1.AddHolding(h1); err != nil {
		t.Fatalf("failed to add holding: %v", err)
	}

	// Create second storage instance (simulating another process)
	s2, err := New(dataPath)
	if err != nil {
		t.Fatalf("failed to create second storage: %v", err)
	}

	// Add a holding via s2
	h2 := models.NewHolding("ETH", 10.0, 3000, "Binance", "", "")
	if err := s2.AddHolding(h2); err != nil {
		t.Fatalf("failed to add holding via s2: %v", err)
	}

	// s1 should still only see 1 holding (cached)
	holdings, _ := s1.GetHoldings()
	if len(holdings) != 1 {
		t.Errorf("expected 1 holding before reload, got %d", len(holdings))
	}

	// Reload s1
	if err := s1.Reload(); err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	// Now s1 should see 2 holdings
	holdings, _ = s1.GetHoldings()
	if len(holdings) != 2 {
		t.Errorf("expected 2 holdings after reload, got %d", len(holdings))
	}

	// Verify both holdings are present
	coins := make(map[string]bool)
	for _, h := range holdings {
		coins[h.Coin] = true
	}
	if !coins["BTC"] || !coins["ETH"] {
		t.Errorf("expected BTC and ETH holdings, got %v", coins)
	}
}
