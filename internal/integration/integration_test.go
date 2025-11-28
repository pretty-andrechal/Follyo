// Package integration provides end-to-end workflow tests for the Follyo portfolio tracker.
// These tests verify complete user workflows to catch regressions in feature changes.
package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pretty-andrechal/follyo/internal/portfolio"
	"github.com/pretty-andrechal/follyo/internal/storage"
)

// testEnv holds test fixtures for integration tests.
type testEnv struct {
	tmpDir        string
	portfolioPath string
	snapshotPath  string
	portfolio     *portfolio.Portfolio
	snapshots     *storage.SnapshotStore
}

// setup creates a fresh test environment with storage and portfolio.
func setup(t *testing.T) *testEnv {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "follyo-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	portfolioPath := filepath.Join(tmpDir, "portfolio.json")
	snapshotPath := filepath.Join(tmpDir, "snapshots.json")

	store, err := storage.New(portfolioPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create storage: %v", err)
	}

	snapStore, err := storage.NewSnapshotStore(snapshotPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create snapshot store: %v", err)
	}

	p := portfolio.New(store)

	return &testEnv{
		tmpDir:        tmpDir,
		portfolioPath: portfolioPath,
		snapshotPath:  snapshotPath,
		portfolio:     p,
		snapshots:     snapStore,
	}
}

// cleanup removes test fixtures.
func (e *testEnv) cleanup() {
	os.RemoveAll(e.tmpDir)
}

// TestCompletePortfolioWorkflow tests the full lifecycle of a portfolio:
// add holdings, sales, stakes, loans -> verify summary -> create snapshot.
func TestCompletePortfolioWorkflow(t *testing.T) {
	env := setup(t)
	defer env.cleanup()

	p := env.portfolio

	// Step 1: Add initial holdings
	t.Run("add holdings", func(t *testing.T) {
		_, err := p.AddHolding("BTC", 2.0, 45000, "Coinbase", "initial purchase", "2024-01-15")
		if err != nil {
			t.Fatalf("AddHolding BTC failed: %v", err)
		}

		_, err = p.AddHolding("ETH", 20.0, 2500, "Binance", "", "2024-01-20")
		if err != nil {
			t.Fatalf("AddHolding ETH failed: %v", err)
		}

		_, err = p.AddHolding("SOL", 100.0, 100, "Kraken", "", "")
		if err != nil {
			t.Fatalf("AddHolding SOL failed: %v", err)
		}

		holdings, _ := p.ListHoldings()
		if len(holdings) != 3 {
			t.Errorf("expected 3 holdings, got %d", len(holdings))
		}
	})

	// Step 2: Make some sales
	t.Run("add sales", func(t *testing.T) {
		_, err := p.AddSale("BTC", 0.5, 55000, "Coinbase", "taking profit", "2024-02-01")
		if err != nil {
			t.Fatalf("AddSale failed: %v", err)
		}

		sales, _ := p.ListSales()
		if len(sales) != 1 {
			t.Errorf("expected 1 sale, got %d", len(sales))
		}
	})

	// Step 3: Stake some holdings
	t.Run("add stakes", func(t *testing.T) {
		apy := 4.5
		_, err := p.AddStake("ETH", 10.0, "Lido", &apy, "liquid staking", "2024-02-15")
		if err != nil {
			t.Fatalf("AddStake ETH failed: %v", err)
		}

		apy2 := 6.0
		_, err = p.AddStake("SOL", 50.0, "Marinade", &apy2, "", "")
		if err != nil {
			t.Fatalf("AddStake SOL failed: %v", err)
		}

		stakes, _ := p.ListStakes()
		if len(stakes) != 2 {
			t.Errorf("expected 2 stakes, got %d", len(stakes))
		}
	})

	// Step 4: Add a loan
	t.Run("add loans", func(t *testing.T) {
		rate := 8.5
		_, err := p.AddLoan("USDC", 10000, "Aave", &rate, "collateral loan", "2024-03-01")
		if err != nil {
			t.Fatalf("AddLoan failed: %v", err)
		}

		loans, _ := p.ListLoans()
		if len(loans) != 1 {
			t.Errorf("expected 1 loan, got %d", len(loans))
		}
	})

	// Step 5: Verify summary calculations
	t.Run("verify summary", func(t *testing.T) {
		summary, err := p.GetSummary()
		if err != nil {
			t.Fatalf("GetSummary failed: %v", err)
		}

		// Holdings: BTC(2*45000) + ETH(20*2500) + SOL(100*100) = 90000 + 50000 + 10000 = 150000
		expectedInvested := 150000.0
		if summary.TotalInvestedUSD != expectedInvested {
			t.Errorf("expected invested %f, got %f", expectedInvested, summary.TotalInvestedUSD)
		}

		// Sales: BTC(0.5*55000) = 27500
		expectedSold := 27500.0
		if summary.TotalSoldUSD != expectedSold {
			t.Errorf("expected sold %f, got %f", expectedSold, summary.TotalSoldUSD)
		}

		// Current holdings (after sales): BTC=1.5, ETH=20, SOL=100
		if summary.HoldingsByCoin["BTC"] != 1.5 {
			t.Errorf("expected BTC holdings 1.5, got %f", summary.HoldingsByCoin["BTC"])
		}

		// Stakes: ETH=10, SOL=50
		if summary.StakesByCoin["ETH"] != 10 {
			t.Errorf("expected ETH staked 10, got %f", summary.StakesByCoin["ETH"])
		}
		if summary.StakesByCoin["SOL"] != 50 {
			t.Errorf("expected SOL staked 50, got %f", summary.StakesByCoin["SOL"])
		}

		// Available (holdings - stakes): ETH=20-10=10, SOL=100-50=50, BTC=1.5
		if summary.AvailableByCoin["ETH"] != 10 {
			t.Errorf("expected ETH available 10, got %f", summary.AvailableByCoin["ETH"])
		}
		if summary.AvailableByCoin["SOL"] != 50 {
			t.Errorf("expected SOL available 50, got %f", summary.AvailableByCoin["SOL"])
		}
		if summary.AvailableByCoin["BTC"] != 1.5 {
			t.Errorf("expected BTC available 1.5, got %f", summary.AvailableByCoin["BTC"])
		}

		// Loans: USDC=10000
		if summary.LoansByCoin["USDC"] != 10000 {
			t.Errorf("expected USDC loan 10000, got %f", summary.LoansByCoin["USDC"])
		}
	})

	// Step 6: Create and verify snapshot
	t.Run("create snapshot", func(t *testing.T) {
		prices := map[string]float64{
			"BTC":  60000, // price went up
			"ETH":  3000,  // price went up
			"SOL":  150,   // price went up
			"USDC": 1,
		}

		snapshot, err := p.CreateSnapshot(prices, "Q1 2024 review")
		if err != nil {
			t.Fatalf("CreateSnapshot failed: %v", err)
		}

		// Holdings value: BTC(1.5*60000) + ETH(20*3000) + SOL(100*150) = 90000 + 60000 + 15000 = 165000
		expectedHoldings := 165000.0
		if snapshot.HoldingsValue != expectedHoldings {
			t.Errorf("expected holdings value %f, got %f", expectedHoldings, snapshot.HoldingsValue)
		}

		// Loans value: USDC(10000*1) = 10000
		expectedLoans := 10000.0
		if snapshot.LoansValue != expectedLoans {
			t.Errorf("expected loans value %f, got %f", expectedLoans, snapshot.LoansValue)
		}

		// Net value: 165000 - 10000 = 155000
		expectedNet := 155000.0
		if snapshot.NetValue != expectedNet {
			t.Errorf("expected net value %f, got %f", expectedNet, snapshot.NetValue)
		}

		// Profit/loss: net_value - total_invested + total_sold = 155000 - 150000 + 27500 = 32500
		expectedPL := 32500.0
		if snapshot.ProfitLoss != expectedPL {
			t.Errorf("expected profit/loss %f, got %f", expectedPL, snapshot.ProfitLoss)
		}

		// Save to store
		if err := env.snapshots.Add(snapshot); err != nil {
			t.Fatalf("failed to save snapshot: %v", err)
		}

		snapshots := env.snapshots.List()
		if len(snapshots) != 1 {
			t.Errorf("expected 1 snapshot in store, got %d", len(snapshots))
		}
	})
}

// TestSnapshotComparisonWorkflow tests taking multiple snapshots and comparing them.
func TestSnapshotComparisonWorkflow(t *testing.T) {
	env := setup(t)
	defer env.cleanup()

	p := env.portfolio

	// Initial state
	p.AddHolding("BTC", 1.0, 50000, "", "", "")
	p.AddHolding("ETH", 10.0, 2000, "", "", "")

	// First snapshot
	prices1 := map[string]float64{"BTC": 50000, "ETH": 2000}
	snap1, err := p.CreateSnapshot(prices1, "initial")
	if err != nil {
		t.Fatalf("CreateSnapshot 1 failed: %v", err)
	}
	env.snapshots.Add(snap1)

	// Make changes
	p.AddHolding("BTC", 0.5, 55000, "", "", "")  // Buy more BTC
	p.AddSale("ETH", 5.0, 2500, "", "", "")      // Sell some ETH

	// Second snapshot with price changes
	prices2 := map[string]float64{"BTC": 60000, "ETH": 3000}
	snap2, err := p.CreateSnapshot(prices2, "after changes")
	if err != nil {
		t.Fatalf("CreateSnapshot 2 failed: %v", err)
	}
	env.snapshots.Add(snap2)

	// Compare snapshots
	comparison := portfolio.CompareSnapshots(&snap1, &snap2)

	// Verify comparison results
	t.Run("verify net value change", func(t *testing.T) {
		// Snap1: BTC(1*50000) + ETH(10*2000) = 50000 + 20000 = 70000
		// Snap2: BTC(1.5*60000) + ETH(5*3000) = 90000 + 15000 = 105000
		// Change: 105000 - 70000 = 35000
		expectedChange := 35000.0
		if comparison.NetValueChange != expectedChange {
			t.Errorf("expected net value change %f, got %f", expectedChange, comparison.NetValueChange)
		}
	})

	t.Run("verify coin changes", func(t *testing.T) {
		btcChange := comparison.CoinChanges["BTC"]
		if btcChange.OldAmount != 1.0 {
			t.Errorf("expected BTC old amount 1.0, got %f", btcChange.OldAmount)
		}
		if btcChange.NewAmount != 1.5 {
			t.Errorf("expected BTC new amount 1.5, got %f", btcChange.NewAmount)
		}

		ethChange := comparison.CoinChanges["ETH"]
		if ethChange.OldAmount != 10.0 {
			t.Errorf("expected ETH old amount 10.0, got %f", ethChange.OldAmount)
		}
		if ethChange.NewAmount != 5.0 {
			t.Errorf("expected ETH new amount 5.0, got %f", ethChange.NewAmount)
		}
	})
}

// TestStakingWorkflow tests the complete staking lifecycle with validation.
func TestStakingWorkflow(t *testing.T) {
	env := setup(t)
	defer env.cleanup()

	p := env.portfolio

	// Add holdings
	p.AddHolding("ETH", 10.0, 2000, "", "", "")

	// Try to stake more than holdings - should fail
	t.Run("reject over-staking", func(t *testing.T) {
		_, err := p.AddStake("ETH", 15.0, "Lido", nil, "", "")
		if err == nil {
			t.Error("expected error when staking more than holdings")
		}
	})

	// Stake within limit
	t.Run("stake within limit", func(t *testing.T) {
		stake, err := p.AddStake("ETH", 6.0, "Lido", nil, "", "")
		if err != nil {
			t.Fatalf("AddStake failed: %v", err)
		}

		available, _ := p.GetAvailableByCoin()
		if available["ETH"] != 4.0 {
			t.Errorf("expected 4.0 ETH available after staking, got %f", available["ETH"])
		}

		// Try to stake more than remaining - should fail
		_, err = p.AddStake("ETH", 5.0, "Rocket Pool", nil, "", "")
		if err == nil {
			t.Error("expected error when staking more than available")
		}

		// Remove stake and verify balance restored
		removed, err := p.RemoveStake(stake.ID)
		if err != nil || !removed {
			t.Fatalf("RemoveStake failed: %v", err)
		}

		available, _ = p.GetAvailableByCoin()
		if available["ETH"] != 10.0 {
			t.Errorf("expected 10.0 ETH available after unstaking, got %f", available["ETH"])
		}
	})

	// Staking with sales
	t.Run("stake respects sales", func(t *testing.T) {
		p.AddSale("ETH", 3.0, 2500, "", "", "")

		// Current: 10 - 3 = 7 ETH
		_, err := p.AddStake("ETH", 8.0, "Lido", nil, "", "")
		if err == nil {
			t.Error("expected error when staking more than current holdings (after sales)")
		}

		_, err = p.AddStake("ETH", 7.0, "Lido", nil, "", "")
		if err != nil {
			t.Fatalf("AddStake for remaining should succeed: %v", err)
		}
	})
}

// TestLoanImpactWorkflow tests how loans affect net value calculations.
func TestLoanImpactWorkflow(t *testing.T) {
	env := setup(t)
	defer env.cleanup()

	p := env.portfolio

	// Add holdings
	p.AddHolding("BTC", 1.0, 50000, "", "", "")

	// Snapshot without loan
	prices := map[string]float64{"BTC": 60000}
	snapNoLoan, _ := p.CreateSnapshot(prices, "no loan")

	// Add loan
	p.AddLoan("USDC", 20000, "Aave", nil, "", "")

	// Snapshot with loan
	prices["USDC"] = 1.0
	snapWithLoan, _ := p.CreateSnapshot(prices, "with loan")

	t.Run("loan reduces net value", func(t *testing.T) {
		// Without loan: net = 60000
		// With loan: net = 60000 - 20000 = 40000
		if snapNoLoan.NetValue != 60000 {
			t.Errorf("expected net value 60000 without loan, got %f", snapNoLoan.NetValue)
		}
		if snapWithLoan.NetValue != 40000 {
			t.Errorf("expected net value 40000 with loan, got %f", snapWithLoan.NetValue)
		}
		if snapWithLoan.LoansValue != 20000 {
			t.Errorf("expected loans value 20000, got %f", snapWithLoan.LoansValue)
		}
	})

	t.Run("net by coin includes loans", func(t *testing.T) {
		net, _ := p.GetNetHoldingsByCoin()

		// BTC: 1.0 (no loans against it)
		if net["BTC"] != 1.0 {
			t.Errorf("expected BTC net 1.0, got %f", net["BTC"])
		}

		// USDC: 0 - 20000 = -20000 (borrowed, not owned)
		if net["USDC"] != -20000 {
			t.Errorf("expected USDC net -20000, got %f", net["USDC"])
		}
	})
}

// TestDataPersistence verifies that data survives storage reload.
func TestDataPersistence(t *testing.T) {
	env := setup(t)
	defer env.cleanup()

	// Add data with first portfolio instance
	p1 := env.portfolio
	p1.AddHolding("BTC", 1.5, 45000, "Coinbase", "test", "")
	p1.AddSale("BTC", 0.5, 50000, "", "", "")
	p1.AddLoan("USDC", 5000, "Aave", nil, "", "")

	apy := 5.0
	p1.AddHolding("ETH", 10, 2000, "", "", "")
	p1.AddStake("ETH", 5, "Lido", &apy, "", "")

	// Create new portfolio instance with same storage path
	store2, err := storage.New(env.portfolioPath)
	if err != nil {
		t.Fatalf("failed to create second storage: %v", err)
	}
	p2 := portfolio.New(store2)

	// Verify all data persisted
	t.Run("holdings persisted", func(t *testing.T) {
		holdings, _ := p2.ListHoldings()
		if len(holdings) != 2 {
			t.Errorf("expected 2 holdings, got %d", len(holdings))
		}
	})

	t.Run("sales persisted", func(t *testing.T) {
		sales, _ := p2.ListSales()
		if len(sales) != 1 {
			t.Errorf("expected 1 sale, got %d", len(sales))
		}
	})

	t.Run("loans persisted", func(t *testing.T) {
		loans, _ := p2.ListLoans()
		if len(loans) != 1 {
			t.Errorf("expected 1 loan, got %d", len(loans))
		}
	})

	t.Run("stakes persisted", func(t *testing.T) {
		stakes, _ := p2.ListStakes()
		if len(stakes) != 1 {
			t.Errorf("expected 1 stake, got %d", len(stakes))
		}
	})

	t.Run("calculations correct after reload", func(t *testing.T) {
		summary, _ := p2.GetSummary()

		// BTC: 1.5 purchased, 0.5 sold = 1.0 current
		if summary.HoldingsByCoin["BTC"] != 1.0 {
			t.Errorf("expected BTC holdings 1.0, got %f", summary.HoldingsByCoin["BTC"])
		}

		// ETH: 10 purchased, 5 staked = 5 available
		if summary.AvailableByCoin["ETH"] != 5 {
			t.Errorf("expected ETH available 5, got %f", summary.AvailableByCoin["ETH"])
		}
	})
}

// TestEdgeCases tests boundary conditions and edge cases.
func TestEdgeCases(t *testing.T) {
	env := setup(t)
	defer env.cleanup()

	p := env.portfolio

	t.Run("empty portfolio summary", func(t *testing.T) {
		summary, err := p.GetSummary()
		if err != nil {
			t.Fatalf("GetSummary on empty portfolio failed: %v", err)
		}
		if summary.TotalInvestedUSD != 0 {
			t.Errorf("expected 0 invested in empty portfolio, got %f", summary.TotalInvestedUSD)
		}
	})

	t.Run("zero amount handling", func(t *testing.T) {
		_, err := p.AddHolding("BTC", 0, 50000, "", "", "")
		if err == nil {
			t.Error("expected error for zero amount holding")
		}
	})

	t.Run("negative amount handling", func(t *testing.T) {
		_, err := p.AddHolding("BTC", -1, 50000, "", "", "")
		if err == nil {
			t.Error("expected error for negative amount holding")
		}
	})

	t.Run("sell more than owned", func(t *testing.T) {
		p.AddHolding("ETH", 5, 2000, "", "", "")

		// This should succeed (system allows it for flexibility)
		// but the current holdings will show negative
		_, err := p.AddSale("ETH", 10, 2500, "", "", "")
		if err != nil {
			// If the system rejects over-selling, that's also valid
			t.Logf("System rejects over-selling: %v", err)
		} else {
			current, _ := p.GetCurrentHoldingsByCoin()
			if current["ETH"] != -5 {
				t.Errorf("expected ETH current -5 after over-selling, got %f", current["ETH"])
			}
		}
	})

	t.Run("coin name normalization", func(t *testing.T) {
		p.AddHolding("btc", 0.1, 50000, "", "", "")
		p.AddHolding("BTC", 0.1, 50000, "", "", "")
		p.AddHolding("Btc", 0.1, 50000, "", "", "")

		byCoin, _ := p.GetHoldingsByCoin()
		// All should be normalized to BTC
		if byCoin["BTC"] < 0.3 {
			t.Errorf("expected at least 0.3 BTC after case-insensitive adds, got %f", byCoin["BTC"])
		}
		if _, exists := byCoin["btc"]; exists {
			t.Error("lowercase 'btc' should not exist as separate key")
		}
	})
}
