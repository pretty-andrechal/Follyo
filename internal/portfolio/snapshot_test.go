package portfolio

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pretty-andrechal/follyo/internal/models"
	"github.com/pretty-andrechal/follyo/internal/storage"
)

func setupSnapshotTestPortfolio(t *testing.T) (*Portfolio, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "follyo-snapshot-test-*")
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

func TestPortfolio_CreateSnapshot(t *testing.T) {
	p, cleanup := setupSnapshotTestPortfolio(t)
	defer cleanup()

	// Add some holdings and loans
	p.AddHolding("BTC", 1.0, 40000, "Coinbase", "", "")
	p.AddHolding("ETH", 10, 2000, "Binance", "", "")
	p.AddLoan("USDT", 5000, "Nexo", nil, "", "")

	prices := map[string]float64{
		"BTC":  50000,
		"ETH":  3000,
		"USDT": 1,
	}

	snapshot, err := p.CreateSnapshot(prices, "Test snapshot")
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	// Check snapshot ID
	if len(snapshot.ID) != models.IDLength {
		t.Errorf("expected %d-char ID, got %d chars", models.IDLength, len(snapshot.ID))
	}

	// Check timestamp is recent
	if time.Since(snapshot.Timestamp) > time.Minute {
		t.Error("expected recent timestamp")
	}

	// Check holdings value: BTC(1*50000) + ETH(10*3000) = 50000 + 30000 = 80000
	expectedHoldingsValue := 80000.0
	if snapshot.HoldingsValue != expectedHoldingsValue {
		t.Errorf("expected holdings value %f, got %f", expectedHoldingsValue, snapshot.HoldingsValue)
	}

	// Check loans value: USDT(5000*1) = 5000
	expectedLoansValue := 5000.0
	if snapshot.LoansValue != expectedLoansValue {
		t.Errorf("expected loans value %f, got %f", expectedLoansValue, snapshot.LoansValue)
	}

	// Check net value: 80000 - 5000 = 75000
	expectedNetValue := 75000.0
	if snapshot.NetValue != expectedNetValue {
		t.Errorf("expected net value %f, got %f", expectedNetValue, snapshot.NetValue)
	}

	// Check total invested: BTC(1*40000) + ETH(10*2000) = 40000 + 20000 = 60000
	expectedTotalInvested := 60000.0
	if snapshot.TotalInvested != expectedTotalInvested {
		t.Errorf("expected total invested %f, got %f", expectedTotalInvested, snapshot.TotalInvested)
	}

	// Check profit/loss: 75000 - 60000 + 0 = 15000
	expectedProfitLoss := 15000.0
	if snapshot.ProfitLoss != expectedProfitLoss {
		t.Errorf("expected profit/loss %f, got %f", expectedProfitLoss, snapshot.ProfitLoss)
	}

	// Check profit percent: 15000 / 60000 * 100 = 25%
	expectedProfitPercent := 25.0
	if snapshot.ProfitPercent != expectedProfitPercent {
		t.Errorf("expected profit percent %f, got %f", expectedProfitPercent, snapshot.ProfitPercent)
	}

	// Check coin values
	if len(snapshot.CoinValues) != 2 {
		t.Errorf("expected 2 coin values, got %d", len(snapshot.CoinValues))
	}

	btcSnap := snapshot.CoinValues["BTC"]
	if btcSnap.Amount != 1.0 {
		t.Errorf("expected BTC amount 1.0, got %f", btcSnap.Amount)
	}
	if btcSnap.Price != 50000 {
		t.Errorf("expected BTC price 50000, got %f", btcSnap.Price)
	}
	if btcSnap.Value != 50000 {
		t.Errorf("expected BTC value 50000, got %f", btcSnap.Value)
	}

	// Check note
	if snapshot.Note != "Test snapshot" {
		t.Errorf("expected note 'Test snapshot', got %s", snapshot.Note)
	}
}

func TestPortfolio_CreateSnapshotWithSales(t *testing.T) {
	p, cleanup := setupSnapshotTestPortfolio(t)
	defer cleanup()

	// Add holdings and sales
	p.AddHolding("BTC", 2.0, 40000, "", "", "")
	p.AddSale("BTC", 0.5, 50000, "", "", "")

	prices := map[string]float64{
		"BTC": 60000,
	}

	snapshot, err := p.CreateSnapshot(prices, "")
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	// Current holdings = 2.0 - 0.5 = 1.5 BTC
	// Holdings value = 1.5 * 60000 = 90000
	expectedHoldingsValue := 90000.0
	if snapshot.HoldingsValue != expectedHoldingsValue {
		t.Errorf("expected holdings value %f, got %f", expectedHoldingsValue, snapshot.HoldingsValue)
	}

	// Total invested = 2.0 * 40000 = 80000
	// Total sold = 0.5 * 50000 = 25000
	// Profit/loss = 90000 - 80000 + 25000 = 35000
	expectedProfitLoss := 35000.0
	if snapshot.ProfitLoss != expectedProfitLoss {
		t.Errorf("expected profit/loss %f, got %f", expectedProfitLoss, snapshot.ProfitLoss)
	}
}

func TestCompareSnapshots(t *testing.T) {
	now := time.Now()

	older := &models.Snapshot{
		ID:            "older123",
		Timestamp:     now.Add(-24 * time.Hour),
		NetValue:      50000,
		ProfitLoss:    5000,
		HoldingsValue: 50000,
		CoinValues: map[string]models.CoinSnapshot{
			"BTC": {Amount: 1.0, Price: 40000, Value: 40000},
			"ETH": {Amount: 5.0, Price: 2000, Value: 10000},
		},
	}

	newer := &models.Snapshot{
		ID:            "newer456",
		Timestamp:     now,
		NetValue:      75000,
		ProfitLoss:    20000,
		HoldingsValue: 75000,
		CoinValues: map[string]models.CoinSnapshot{
			"BTC": {Amount: 1.0, Price: 50000, Value: 50000},
			"ETH": {Amount: 10.0, Price: 2500, Value: 25000},
		},
	}

	comparison := CompareSnapshots(older, newer)

	// Check snapshots are set
	if comparison.OlderSnapshot.ID != "older123" {
		t.Errorf("expected older snapshot ID older123, got %s", comparison.OlderSnapshot.ID)
	}
	if comparison.NewerSnapshot.ID != "newer456" {
		t.Errorf("expected newer snapshot ID newer456, got %s", comparison.NewerSnapshot.ID)
	}

	// Check net value change: 75000 - 50000 = 25000
	expectedNetValueChange := 25000.0
	if comparison.NetValueChange != expectedNetValueChange {
		t.Errorf("expected net value change %f, got %f", expectedNetValueChange, comparison.NetValueChange)
	}

	// Check net value percent: 25000 / 50000 * 100 = 50%
	expectedNetValuePercent := 50.0
	if comparison.NetValuePercent != expectedNetValuePercent {
		t.Errorf("expected net value percent %f, got %f", expectedNetValuePercent, comparison.NetValuePercent)
	}

	// Check profit/loss change: 20000 - 5000 = 15000
	expectedProfitLossChange := 15000.0
	if comparison.ProfitLossChange != expectedProfitLossChange {
		t.Errorf("expected profit/loss change %f, got %f", expectedProfitLossChange, comparison.ProfitLossChange)
	}

	// Check coin changes
	if len(comparison.CoinChanges) != 2 {
		t.Errorf("expected 2 coin changes, got %d", len(comparison.CoinChanges))
	}

	btcChange := comparison.CoinChanges["BTC"]
	if btcChange.OldAmount != 1.0 {
		t.Errorf("expected BTC old amount 1.0, got %f", btcChange.OldAmount)
	}
	if btcChange.NewAmount != 1.0 {
		t.Errorf("expected BTC new amount 1.0, got %f", btcChange.NewAmount)
	}
	if btcChange.OldPrice != 40000 {
		t.Errorf("expected BTC old price 40000, got %f", btcChange.OldPrice)
	}
	if btcChange.NewPrice != 50000 {
		t.Errorf("expected BTC new price 50000, got %f", btcChange.NewPrice)
	}
	// BTC value change: 50000 - 40000 = 10000
	if btcChange.ValueChange != 10000 {
		t.Errorf("expected BTC value change 10000, got %f", btcChange.ValueChange)
	}

	ethChange := comparison.CoinChanges["ETH"]
	if ethChange.OldAmount != 5.0 {
		t.Errorf("expected ETH old amount 5.0, got %f", ethChange.OldAmount)
	}
	if ethChange.NewAmount != 10.0 {
		t.Errorf("expected ETH new amount 10.0, got %f", ethChange.NewAmount)
	}
	// ETH value change: 25000 - 10000 = 15000
	if ethChange.ValueChange != 15000 {
		t.Errorf("expected ETH value change 15000, got %f", ethChange.ValueChange)
	}
}

func TestCompareSnapshots_NewCoin(t *testing.T) {
	older := &models.Snapshot{
		ID:        "old",
		Timestamp: time.Now().Add(-time.Hour),
		NetValue:  10000,
		CoinValues: map[string]models.CoinSnapshot{
			"BTC": {Amount: 0.5, Price: 20000, Value: 10000},
		},
	}

	newer := &models.Snapshot{
		ID:        "new",
		Timestamp: time.Now(),
		NetValue:  25000,
		CoinValues: map[string]models.CoinSnapshot{
			"BTC": {Amount: 0.5, Price: 20000, Value: 10000},
			"ETH": {Amount: 5.0, Price: 3000, Value: 15000},
		},
	}

	comparison := CompareSnapshots(older, newer)

	// Should have both coins
	if len(comparison.CoinChanges) != 2 {
		t.Errorf("expected 2 coin changes, got %d", len(comparison.CoinChanges))
	}

	// ETH is new
	ethChange := comparison.CoinChanges["ETH"]
	if ethChange.OldAmount != 0 {
		t.Errorf("expected ETH old amount 0, got %f", ethChange.OldAmount)
	}
	if ethChange.NewAmount != 5.0 {
		t.Errorf("expected ETH new amount 5.0, got %f", ethChange.NewAmount)
	}
	if ethChange.ValueChange != 15000 {
		t.Errorf("expected ETH value change 15000, got %f", ethChange.ValueChange)
	}
}

func TestCompareSnapshots_RemovedCoin(t *testing.T) {
	older := &models.Snapshot{
		ID:        "old",
		Timestamp: time.Now().Add(-time.Hour),
		NetValue:  25000,
		CoinValues: map[string]models.CoinSnapshot{
			"BTC": {Amount: 0.5, Price: 20000, Value: 10000},
			"ETH": {Amount: 5.0, Price: 3000, Value: 15000},
		},
	}

	newer := &models.Snapshot{
		ID:        "new",
		Timestamp: time.Now(),
		NetValue:  12000,
		CoinValues: map[string]models.CoinSnapshot{
			"BTC": {Amount: 0.5, Price: 24000, Value: 12000},
		},
	}

	comparison := CompareSnapshots(older, newer)

	// Should have both coins
	if len(comparison.CoinChanges) != 2 {
		t.Errorf("expected 2 coin changes, got %d", len(comparison.CoinChanges))
	}

	// ETH was removed
	ethChange := comparison.CoinChanges["ETH"]
	if ethChange.OldAmount != 5.0 {
		t.Errorf("expected ETH old amount 5.0, got %f", ethChange.OldAmount)
	}
	if ethChange.NewAmount != 0 {
		t.Errorf("expected ETH new amount 0, got %f", ethChange.NewAmount)
	}
	if ethChange.ValueChange != -15000 {
		t.Errorf("expected ETH value change -15000, got %f", ethChange.ValueChange)
	}
}

func TestCompareSnapshots_ZeroNetValue(t *testing.T) {
	older := &models.Snapshot{
		ID:        "old",
		Timestamp: time.Now().Add(-time.Hour),
		NetValue:  0,
		CoinValues: map[string]models.CoinSnapshot{},
	}

	newer := &models.Snapshot{
		ID:        "new",
		Timestamp: time.Now(),
		NetValue:  10000,
		CoinValues: map[string]models.CoinSnapshot{
			"BTC": {Amount: 0.2, Price: 50000, Value: 10000},
		},
	}

	comparison := CompareSnapshots(older, newer)

	// Percent should be 0 when older net value is 0
	if comparison.NetValuePercent != 0 {
		t.Errorf("expected net value percent 0 for zero base, got %f", comparison.NetValuePercent)
	}
}
