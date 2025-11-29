package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
	"github.com/pretty-andrechal/follyo/internal/models"
	"github.com/pretty-andrechal/follyo/internal/storage"
)

func setupCoinHistoryTest(t *testing.T) (*storage.SnapshotStore, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "follyo-coinhistory-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	snapshotPath := filepath.Join(tmpDir, "snapshots.json")
	store, err := storage.NewSnapshotStore(snapshotPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create snapshot store: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return store, cleanup
}

var testSnapshotCounter int

func addTestSnapshot(t *testing.T, store *storage.SnapshotStore, timestamp time.Time, coinValues map[string]models.CoinSnapshot) {
	t.Helper()

	testSnapshotCounter++
	snapshot := models.Snapshot{
		ID:         fmt.Sprintf("test%d", testSnapshotCounter),
		Timestamp:  timestamp,
		CoinValues: coinValues,
		NetValue:   1000.0,
	}
	if err := store.Add(snapshot); err != nil {
		t.Fatalf("failed to add snapshot: %v", err)
	}
}

func TestNewCoinHistoryModel(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	m := NewCoinHistoryModel(store)

	if m.store != store {
		t.Error("store not set correctly")
	}

	if m.mode != CoinHistoryCoinSelect {
		t.Errorf("expected mode CoinHistoryCoinSelect, got %d", m.mode)
	}

	if m.availableCoins == nil {
		t.Error("availableCoins should be initialized")
	}
}

func TestCoinHistoryModel_ExtractAvailableCoins(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	// Add snapshots with different coins
	addTestSnapshot(t, store, time.Now().Add(-2*time.Hour), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 50000, Value: 50000},
		"ETH": {Amount: 10.0, Price: 3000, Value: 30000},
	})
	addTestSnapshot(t, store, time.Now().Add(-1*time.Hour), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.5, Price: 52000, Value: 78000},
		"SOL": {Amount: 100, Price: 100, Value: 10000},
	})

	m := NewCoinHistoryModel(store)

	if len(m.availableCoins) != 3 {
		t.Errorf("expected 3 coins, got %d", len(m.availableCoins))
	}

	// Should be sorted alphabetically
	expected := []string{"BTC", "ETH", "SOL"}
	for i, coin := range expected {
		if m.availableCoins[i] != coin {
			t.Errorf("expected coin %s at position %d, got %s", coin, i, m.availableCoins[i])
		}
	}
}

func TestCoinHistoryModel_LoadCoinHistory(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	now := time.Now()
	addTestSnapshot(t, store, now.Add(-2*time.Hour), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 50000, Value: 50000},
	})
	addTestSnapshot(t, store, now.Add(-1*time.Hour), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.5, Price: 52000, Value: 78000},
	})
	addTestSnapshot(t, store, now, map[string]models.CoinSnapshot{
		"BTC": {Amount: 2.0, Price: 55000, Value: 110000},
	})

	m := NewCoinHistoryModel(store)
	m.loadCoinHistory("BTC")

	if len(m.coinData) != 3 {
		t.Fatalf("expected 3 data points, got %d", len(m.coinData))
	}

	// Should be in chronological order (oldest first)
	if m.coinData[0].Price != 50000 {
		t.Errorf("expected first price 50000, got %f", m.coinData[0].Price)
	}
	if m.coinData[2].Price != 55000 {
		t.Errorf("expected last price 55000, got %f", m.coinData[2].Price)
	}
}

func TestCoinHistoryModel_Init(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	m := NewCoinHistoryModel(store)
	cmd := m.Init()

	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestCoinHistoryModel_NavigateCoins(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	addTestSnapshot(t, store, time.Now(), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 50000, Value: 50000},
		"ETH": {Amount: 10.0, Price: 3000, Value: 30000},
		"SOL": {Amount: 100, Price: 100, Value: 10000},
	})

	m := NewCoinHistoryModel(store)

	// Press down
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(downMsg)
	m = newModel.(CoinHistoryModel)

	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after down, got %d", m.cursor)
	}

	// Press up
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = m.Update(upMsg)
	m = newModel.(CoinHistoryModel)

	if m.cursor != 0 {
		t.Errorf("expected cursor at 0 after up, got %d", m.cursor)
	}
}

func TestCoinHistoryModel_SelectCoin(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	addTestSnapshot(t, store, time.Now(), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 50000, Value: 50000},
		"ETH": {Amount: 10.0, Price: 3000, Value: 30000},
	})

	m := NewCoinHistoryModel(store)

	// Select first coin (BTC)
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(enterMsg)
	m = newModel.(CoinHistoryModel)

	if m.mode != CoinHistoryDisplay {
		t.Errorf("expected mode CoinHistoryDisplay, got %d", m.mode)
	}

	if m.selectedCoin != "BTC" {
		t.Errorf("expected selectedCoin BTC, got %s", m.selectedCoin)
	}

	if len(m.coinData) != 1 {
		t.Errorf("expected 1 data point, got %d", len(m.coinData))
	}
}

func TestCoinHistoryModel_BackFromDisplay(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	addTestSnapshot(t, store, time.Now(), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 50000, Value: 50000},
	})

	m := NewCoinHistoryModel(store)

	// Select coin
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(enterMsg)
	m = newModel.(CoinHistoryModel)

	// Go back
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ = m.Update(escMsg)
	m = newModel.(CoinHistoryModel)

	if m.mode != CoinHistoryCoinSelect {
		t.Errorf("expected mode CoinHistoryCoinSelect, got %d", m.mode)
	}

	if m.selectedCoin != "" {
		t.Errorf("expected empty selectedCoin, got %s", m.selectedCoin)
	}
}

func TestCoinHistoryModel_BackToMenu(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	m := NewCoinHistoryModel(store)

	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	_, cmd := m.Update(escMsg)

	if cmd == nil {
		t.Fatal("expected command on escape")
	}

	msg := cmd()
	_, ok := msg.(tui.BackToMenuMsg)
	if !ok {
		t.Errorf("expected BackToMenuMsg, got %T", msg)
	}
}

func TestCoinHistoryModel_QuitKey(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	m := NewCoinHistoryModel(store)

	qMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(qMsg)

	if cmd == nil {
		t.Fatal("expected command on 'q'")
	}

	msg := cmd()
	if msg != tea.Quit() {
		t.Error("expected tea.Quit message")
	}
}

func TestCoinHistoryModel_WindowResize(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	m := NewCoinHistoryModel(store)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	newModel, _ := m.Update(msg)
	m = newModel.(CoinHistoryModel)

	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height 40, got %d", m.height)
	}
	if !m.viewportReady {
		t.Error("viewport should be ready after window resize")
	}
}

func TestCoinHistoryModel_ViewCoinSelectEmpty(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	m := NewCoinHistoryModel(store)
	view := m.View()

	checks := []string{
		"COIN HISTORY",
		"No coins found",
		"snapshots",
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("view should contain '%s'", check)
		}
	}
}

func TestCoinHistoryModel_ViewCoinSelectWithCoins(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	addTestSnapshot(t, store, time.Now(), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 50000, Value: 50000},
		"ETH": {Amount: 10.0, Price: 3000, Value: 30000},
	})

	m := NewCoinHistoryModel(store)
	view := m.View()

	checks := []string{
		"COIN HISTORY",
		"BTC",
		"ETH",
		"navigate",
		"toggle", // space to toggle selection
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("view should contain '%s'", check)
		}
	}
}

func TestCoinHistoryModel_ViewDisplay(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	now := time.Now()
	addTestSnapshot(t, store, now.Add(-1*time.Hour), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 50000, Value: 50000},
	})
	addTestSnapshot(t, store, now, map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.5, Price: 55000, Value: 82500},
	})

	m := NewCoinHistoryModel(store)

	// Set up viewport
	msg := tea.WindowSizeMsg{Width: 100, Height: 40}
	newModel, _ := m.Update(msg)
	m = newModel.(CoinHistoryModel)

	// Select BTC
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = m.Update(enterMsg)
	m = newModel.(CoinHistoryModel)

	view := m.View()

	checks := []string{
		"BTC HISTORY",
		"scroll",
		"back",
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("display view should contain '%s'", check)
		}
	}
}

func TestCoinHistoryModel_CalculatePriceStats(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	now := time.Now()
	addTestSnapshot(t, store, now.Add(-2*time.Hour), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 50000, Value: 50000},
	})
	addTestSnapshot(t, store, now.Add(-1*time.Hour), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 60000, Value: 60000},
	})
	addTestSnapshot(t, store, now, map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 55000, Value: 55000},
	})

	m := NewCoinHistoryModel(store)
	m.loadCoinHistory("BTC")

	min, max, avg := m.calculatePriceStats()

	if min != 50000 {
		t.Errorf("expected min 50000, got %f", min)
	}
	if max != 60000 {
		t.Errorf("expected max 60000, got %f", max)
	}
	expectedAvg := (50000.0 + 60000.0 + 55000.0) / 3
	if avg != expectedAvg {
		t.Errorf("expected avg %f, got %f", expectedAvg, avg)
	}
}

func TestCoinHistoryModel_CalculatePriceChange(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	now := time.Now()
	addTestSnapshot(t, store, now.Add(-1*time.Hour), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 50000, Value: 50000},
	})
	addTestSnapshot(t, store, now, map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 55000, Value: 55000},
	})

	m := NewCoinHistoryModel(store)
	m.loadCoinHistory("BTC")

	change, percent := m.calculatePriceChange()

	if change != 5000 {
		t.Errorf("expected change 5000, got %f", change)
	}
	expectedPercent := (5000.0 / 50000.0) * 100
	if percent != expectedPercent {
		t.Errorf("expected percent %f, got %f", expectedPercent, percent)
	}
}

func TestCoinHistoryModel_HasVaryingAmounts(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	now := time.Now()

	// Test with varying amounts
	addTestSnapshot(t, store, now.Add(-1*time.Hour), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 50000, Value: 50000},
	})
	addTestSnapshot(t, store, now, map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.5, Price: 55000, Value: 82500},
	})

	m := NewCoinHistoryModel(store)
	m.loadCoinHistory("BTC")

	if !m.hasVaryingAmounts() {
		t.Error("should have varying amounts")
	}

	// Test with same amounts
	store2, cleanup2 := setupCoinHistoryTest(t)
	defer cleanup2()

	addTestSnapshot(t, store2, now.Add(-1*time.Hour), map[string]models.CoinSnapshot{
		"ETH": {Amount: 10.0, Price: 3000, Value: 30000},
	})
	addTestSnapshot(t, store2, now, map[string]models.CoinSnapshot{
		"ETH": {Amount: 10.0, Price: 3500, Value: 35000},
	})

	m2 := NewCoinHistoryModel(store2)
	m2.loadCoinHistory("ETH")

	if m2.hasVaryingAmounts() {
		t.Error("should not have varying amounts")
	}
}

func TestCoinHistoryModel_VimKeys(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	addTestSnapshot(t, store, time.Now(), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 50000, Value: 50000},
		"ETH": {Amount: 10.0, Price: 3000, Value: 30000},
	})

	m := NewCoinHistoryModel(store)

	// Test 'j' for down
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := m.Update(jMsg)
	m = newModel.(CoinHistoryModel)

	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after 'j', got %d", m.cursor)
	}

	// Test 'k' for up
	kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ = m.Update(kMsg)
	m = newModel.(CoinHistoryModel)

	if m.cursor != 0 {
		t.Errorf("expected cursor at 0 after 'k', got %d", m.cursor)
	}
}

func TestCoinHistoryModel_GetStore(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	m := NewCoinHistoryModel(store)

	if m.GetStore() != store {
		t.Error("GetStore should return the store")
	}
}

func TestCoinHistoryModel_CountDataPoints(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	now := time.Now()
	addTestSnapshot(t, store, now.Add(-2*time.Hour), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 50000, Value: 50000},
	})
	addTestSnapshot(t, store, now.Add(-1*time.Hour), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 52000, Value: 52000},
		"ETH": {Amount: 10.0, Price: 3000, Value: 30000},
	})
	addTestSnapshot(t, store, now, map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 55000, Value: 55000},
	})

	m := NewCoinHistoryModel(store)

	if m.countDataPoints("BTC") != 3 {
		t.Errorf("expected 3 data points for BTC, got %d", m.countDataPoints("BTC"))
	}

	if m.countDataPoints("ETH") != 1 {
		t.Errorf("expected 1 data point for ETH, got %d", m.countDataPoints("ETH"))
	}

	if m.countDataPoints("SOL") != 0 {
		t.Errorf("expected 0 data points for SOL, got %d", m.countDataPoints("SOL"))
	}
}

func TestCoinHistoryModel_CompareMultipleCoins(t *testing.T) {
	store, cleanup := setupCoinHistoryTest(t)
	defer cleanup()

	// Add snapshots with multiple coins
	now := time.Now()
	addTestSnapshot(t, store, now.Add(-48*time.Hour), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 50000, Value: 50000},
		"ETH": {Amount: 10.0, Price: 3000, Value: 30000},
	})
	addTestSnapshot(t, store, now.Add(-24*time.Hour), map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 52000, Value: 52000},
		"ETH": {Amount: 10.0, Price: 3200, Value: 32000},
	})
	addTestSnapshot(t, store, now, map[string]models.CoinSnapshot{
		"BTC": {Amount: 1.0, Price: 55000, Value: 55000},
		"ETH": {Amount: 10.0, Price: 3500, Value: 35000},
	})

	m := NewCoinHistoryModel(store)

	// Simulate window size to initialize viewport
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m = updated.(CoinHistoryModel)

	// Verify coins are available
	if len(m.availableCoins) != 2 {
		t.Fatalf("expected 2 available coins, got %d", len(m.availableCoins))
	}

	// Select first coin (space key to toggle)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updated.(CoinHistoryModel)

	// Move down and select second coin
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(CoinHistoryModel)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updated.(CoinHistoryModel)

	// Check that both are selected
	if len(m.GetSelectedCoins()) != 2 {
		t.Fatalf("expected 2 selected coins, got %d", len(m.GetSelectedCoins()))
	}

	// Press enter to compare - this should not panic
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(CoinHistoryModel)

	// Check that mode changed to compare
	if m.mode != CoinHistoryCompare {
		t.Errorf("expected mode to be CoinHistoryCompare, got %v", m.mode)
	}

	// Try to render the view - this is where crash may occur
	view := m.View()

	if view == "" {
		t.Error("view should not be empty")
	}

	// Check for expected content in compare view
	if !strings.Contains(view, "COMPARE") {
		t.Error("view should contain 'COMPARE'")
	}
}
