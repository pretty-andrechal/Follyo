package views

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
	"github.com/pretty-andrechal/follyo/internal/portfolio"
	"github.com/pretty-andrechal/follyo/internal/storage"
)

func setupTestPortfolio(t *testing.T) (*portfolio.Portfolio, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "follyo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dataPath := filepath.Join(tmpDir, "test.json")
	s, err := storage.New(dataPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create storage: %v", err)
	}

	p := portfolio.New(s)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return p, cleanup
}

func TestNewSummaryModel(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	m := NewSummaryModel(p, nil)

	if m.portfolio != p {
		t.Error("portfolio not set correctly")
	}

	if !m.loading {
		t.Error("should start in loading state")
	}
}

func TestSummaryModel_Init(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	m := NewSummaryModel(p, nil)
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestSummaryModel_LoadDataMessage(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	// Add some test data
	_, err := p.AddHolding("BTC", 1.5, 50000, "Coinbase", "", "")
	if err != nil {
		t.Fatalf("failed to add holding: %v", err)
	}

	m := NewSummaryModel(p, nil)

	// Simulate receiving data
	summary, _ := p.GetSummary()
	msg := SummaryDataMsg{
		Summary: &summary,
		Prices:  map[string]float64{"BTC": 60000},
	}

	newModel, _ := m.Update(msg)
	m = newModel.(SummaryModel)

	if m.loading {
		t.Error("should not be loading after receiving data")
	}

	if m.summary == nil {
		t.Error("summary should be set after receiving data")
	}

	if m.livePrices["BTC"] != 60000 {
		t.Errorf("expected BTC price 60000, got %v", m.livePrices["BTC"])
	}
}

func TestSummaryModel_LoadDataError(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	m := NewSummaryModel(p, nil)

	// Simulate error
	msg := SummaryDataMsg{
		Error: os.ErrNotExist,
	}

	newModel, _ := m.Update(msg)
	m = newModel.(SummaryModel)

	if m.err == nil {
		t.Error("error should be set")
	}

	if m.loading {
		t.Error("should not be loading after error")
	}
}

func TestSummaryModel_BackNavigation(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	m := NewSummaryModel(p, nil)

	// Press escape
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	_, cmd := m.Update(escMsg)

	if cmd == nil {
		t.Fatal("expected command on escape")
	}

	// Execute command
	msg := cmd()
	_, ok := msg.(tui.BackToMenuMsg)
	if !ok {
		t.Errorf("expected BackToMenuMsg, got %T", msg)
	}
}

func TestSummaryModel_QuitKey(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	m := NewSummaryModel(p, nil)

	// Press q
	qMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(qMsg)

	if cmd == nil {
		t.Fatal("expected command on 'q'")
	}

	// The command should be tea.Quit
	msg := cmd()
	if msg != tea.Quit() {
		t.Error("expected tea.Quit message")
	}
}

func TestSummaryModel_RefreshKey(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	m := NewSummaryModel(p, nil)
	m.loading = false // Pretend we're done loading

	// Press r
	rMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	newModel, cmd := m.Update(rMsg)
	m = newModel.(SummaryModel)

	if !m.loading {
		t.Error("should be loading after refresh")
	}

	if cmd == nil {
		t.Error("should return command to reload data")
	}
}

func TestSummaryModel_WindowResize(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	m := NewSummaryModel(p, nil)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	newModel, _ := m.Update(msg)
	m = newModel.(SummaryModel)

	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height 40, got %d", m.height)
	}
}

func TestSummaryModel_ViewLoading(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	m := NewSummaryModel(p, nil)
	view := m.View()

	if !strings.Contains(view, "PORTFOLIO SUMMARY") {
		t.Error("loading view should contain 'PORTFOLIO SUMMARY'")
	}
	if !strings.Contains(view, "Fetching") {
		t.Error("loading view should contain 'Fetching'")
	}
}

func TestSummaryModel_ViewError(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	m := NewSummaryModel(p, nil)
	m.loading = false
	m.err = os.ErrNotExist

	view := m.View()

	if !strings.Contains(view, "Error") {
		t.Error("error view should contain 'Error'")
	}
}

func TestSummaryModel_ViewWithData(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	// Add test data
	p.AddHolding("BTC", 1.0, 50000, "", "", "")
	p.AddHolding("ETH", 10.0, 3000, "", "", "")

	m := NewSummaryModel(p, nil)
	summary, _ := p.GetSummary()
	m.loading = false
	m.summary = &summary
	m.livePrices = map[string]float64{"BTC": 60000, "ETH": 3500}

	// Initialize viewport with window size
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(SummaryModel)

	view := m.View()

	// Check for expected content
	checks := []string{
		"PORTFOLIO SUMMARY",
		"HOLDINGS BY COIN",
		"BTC",
		"ETH",
		"Holdings:",
		"refresh",
		"back",
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("view should contain '%s'", check)
		}
	}
}

func TestSummaryModel_ViewOffline(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	p.AddHolding("BTC", 1.0, 50000, "", "", "")

	m := NewSummaryModel(p, nil)
	summary, _ := p.GetSummary()
	m.loading = false
	m.summary = &summary
	m.isOffline = true

	// Initialize viewport with window size
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(SummaryModel)

	view := m.View()

	if !strings.Contains(view, "Offline") || !strings.Contains(view, "unavailable") {
		t.Error("offline view should show offline message")
	}
}

func TestSummaryModel_ViewUnmappedTickers(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	p.AddHolding("UNKNOWN", 100, 1, "", "", "")

	m := NewSummaryModel(p, nil)
	summary, _ := p.GetSummary()
	m.loading = false
	m.summary = &summary
	m.unmappedTickers = []string{"UNKNOWN"}

	// Initialize viewport with window size
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(SummaryModel)

	view := m.View()

	if !strings.Contains(view, "UNKNOWN") {
		t.Error("view should show unmapped ticker")
	}

	if !strings.Contains(view, "mapping") || !strings.Contains(view, "CoinGecko") {
		t.Error("view should mention missing CoinGecko mapping")
	}
}

func TestCollectCoins(t *testing.T) {
	summary := portfolio.Summary{
		HoldingsByCoin: map[string]float64{"BTC": 1.0, "ETH": 2.0},
		StakesByCoin:   map[string]float64{"ETH": 1.0, "SOL": 10.0},
		LoansByCoin:    map[string]float64{"USDC": 1000},
		NetByCoin:      map[string]float64{"BTC": 1.0},
	}

	coins := collectCoins(summary)

	expected := map[string]bool{"BTC": true, "ETH": true, "SOL": true, "USDC": true}

	if len(coins) != len(expected) {
		t.Errorf("expected %d coins, got %d", len(expected), len(coins))
	}

	for _, coin := range coins {
		if !expected[coin] {
			t.Errorf("unexpected coin: %s", coin)
		}
	}
}

func TestSummaryModel_WithTickerMappings(t *testing.T) {
	p, cleanup := setupTestPortfolio(t)
	defer cleanup()

	mappings := map[string]string{
		"MYTOKEN": "my-token-id",
	}

	m := NewSummaryModel(p, mappings)

	if len(m.tickerMappings) != 1 {
		t.Errorf("expected 1 ticker mapping, got %d", len(m.tickerMappings))
	}

	if m.tickerMappings["MYTOKEN"] != "my-token-id" {
		t.Error("ticker mapping not set correctly")
	}
}
