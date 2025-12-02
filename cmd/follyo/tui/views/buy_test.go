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

func setupBuyTest(t *testing.T) (*portfolio.Portfolio, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "follyo-buy-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	portfolioPath := filepath.Join(tmpDir, "portfolio.json")
	s, err := storage.New(portfolioPath)
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

func TestNewBuyModel(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	m := NewBuyModel(p, "TestPlatform")

	if m.portfolio != p {
		t.Error("portfolio not set correctly")
	}

	if m.defaultPlatform != "TestPlatform" {
		t.Errorf("expected default platform 'TestPlatform', got '%s'", m.defaultPlatform)
	}

	if m.state.Mode != EntityModeList {
		t.Errorf("expected mode EntityModeList, got %d", m.state.Mode)
	}

	// 7 fields: coin, amount, price, total, date, platform, notes
	if len(m.state.Inputs) != 7 {
		t.Errorf("expected 7 inputs, got %d", len(m.state.Inputs))
	}
}

func TestBuyModel_Init(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	m := NewBuyModel(p, "")
	cmd := m.Init()

	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestBuyModel_NavigateDown(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	// Add some holdings
	_, _ = p.AddHolding("BTC", 1.0, 50000, "Test", "", "")
	_, _ = p.AddHolding("ETH", 10.0, 3000, "Test", "", "")

	m := NewBuyModel(p, "")

	// Press down
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)
	m = newModel.(BuyModel)

	if m.state.Cursor != 1 {
		t.Errorf("expected cursor at 1 after down, got %d", m.state.Cursor)
	}
}

func TestBuyModel_NavigateUp(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	// Add some holdings
	_, _ = p.AddHolding("BTC", 1.0, 50000, "Test", "", "")
	_, _ = p.AddHolding("ETH", 10.0, 3000, "Test", "", "")

	m := NewBuyModel(p, "")

	// Move down first
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)
	m = newModel.(BuyModel)

	// Now press up
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = m.Update(upMsg)
	m = newModel.(BuyModel)

	if m.state.Cursor != 0 {
		t.Errorf("expected cursor at 0 after up, got %d", m.state.Cursor)
	}
}

func TestBuyModel_NavigateBoundaries(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	// Add some holdings
	_, _ = p.AddHolding("BTC", 1.0, 50000, "Test", "", "")
	_, _ = p.AddHolding("ETH", 10.0, 3000, "Test", "", "")

	m := NewBuyModel(p, "")

	// Press up at top - should stay at 0
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := m.Update(upMsg)
	m = newModel.(BuyModel)

	if m.state.Cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", m.state.Cursor)
	}

	// Navigate to bottom
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	for i := 0; i < 10; i++ {
		newModel, _ = m.Update(downMsg)
		m = newModel.(BuyModel)
	}

	expectedBottom := len(m.holdings) - 1
	if m.state.Cursor != expectedBottom {
		t.Errorf("expected cursor at %d (bottom), got %d", expectedBottom, m.state.Cursor)
	}
}

func TestBuyModel_BackNavigation(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	m := NewBuyModel(p, "")

	// Press escape
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

func TestBuyModel_QuitKey(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	m := NewBuyModel(p, "")

	// Press q
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

func TestBuyModel_WindowResize(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	m := NewBuyModel(p, "")

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	newModel, _ := m.Update(msg)
	m = newModel.(BuyModel)

	if m.state.Width != 120 {
		t.Errorf("expected width 120, got %d", m.state.Width)
	}
	if m.state.Height != 40 {
		t.Errorf("expected height 40, got %d", m.state.Height)
	}
}

func TestBuyModel_ViewListEmpty(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	m := NewBuyModel(p, "")
	view := m.View()

	checks := []string{
		"PURCHASES",
		"No purchases yet",
		"add",
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("view should contain '%s'", check)
		}
	}
}

func TestBuyModel_ViewListWithHoldings(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	_, _ = p.AddHolding("BTC", 0.5, 50000, "Coinbase", "", "")

	m := NewBuyModel(p, "")
	view := m.View()

	checks := []string{
		"PURCHASES",
		"BTC",
		"0.5",
		"Coinbase",
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("view should contain '%s'", check)
		}
	}
}

func TestBuyModel_EnterAddMode(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	m := NewBuyModel(p, "")

	// Press 'a' to add
	aMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := m.Update(aMsg)
	m = newModel.(BuyModel)

	if m.state.Mode != EntityModeAdd {
		t.Errorf("expected mode EntityModeAdd, got %d", m.state.Mode)
	}
}

func TestBuyModel_CancelAddMode(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	m := NewBuyModel(p, "")

	// Enter add mode
	aMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := m.Update(aMsg)
	m = newModel.(BuyModel)

	// Cancel with escape
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ = m.Update(escMsg)
	m = newModel.(BuyModel)

	if m.state.Mode != EntityModeList {
		t.Errorf("expected mode EntityModeList after cancel, got %d", m.state.Mode)
	}
}

func TestBuyModel_AddFormView(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	m := NewBuyModel(p, "")

	// Enter add mode
	aMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := m.Update(aMsg)
	m = newModel.(BuyModel)

	view := m.View()

	checks := []string{
		"ADD PURCHASE",
		"Coin:",
		"Amount:",
		"Price",
		"Platform:",
		"Notes:",
		"save",
		"cancel",
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("add form view should contain '%s'", check)
		}
	}
}

func TestBuyModel_DeleteConfirmMode(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	_, _ = p.AddHolding("BTC", 1.0, 50000, "Test", "", "")

	m := NewBuyModel(p, "")

	// Press 'd' to delete
	dMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := m.Update(dMsg)
	m = newModel.(BuyModel)

	if m.state.Mode != EntityModeConfirmDelete {
		t.Errorf("expected mode EntityModeConfirmDelete, got %d", m.state.Mode)
	}
}

func TestBuyModel_CancelDelete(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	_, _ = p.AddHolding("BTC", 1.0, 50000, "Test", "", "")

	m := NewBuyModel(p, "")

	// Enter delete confirm mode
	dMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := m.Update(dMsg)
	m = newModel.(BuyModel)

	// Cancel with 'n'
	nMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ = m.Update(nMsg)
	m = newModel.(BuyModel)

	if m.state.Mode != EntityModeList {
		t.Errorf("expected mode EntityModeList after cancel, got %d", m.state.Mode)
	}
}

func TestBuyModel_ConfirmDelete(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	_, _ = p.AddHolding("BTC", 1.0, 50000, "Test", "", "")

	m := NewBuyModel(p, "")

	initialCount := len(m.holdings)
	if initialCount != 1 {
		t.Fatalf("expected 1 holding, got %d", initialCount)
	}

	// Enter delete confirm mode
	dMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := m.Update(dMsg)
	m = newModel.(BuyModel)

	// Confirm with 'y'
	yMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	_, cmd := m.Update(yMsg)

	if cmd == nil {
		t.Fatal("expected command on delete confirm")
	}

	// Execute the command
	msg := cmd()
	deletedMsg, ok := msg.(HoldingDeletedMsg)
	if !ok {
		t.Fatalf("expected HoldingDeletedMsg, got %T", msg)
	}

	if deletedMsg.Error != nil {
		t.Errorf("unexpected error: %v", deletedMsg.Error)
	}
}

func TestBuyModel_DeleteConfirmView(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	_, _ = p.AddHolding("BTC", 1.0, 50000, "Test", "", "")

	m := NewBuyModel(p, "")

	// Enter delete confirm mode
	dMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := m.Update(dMsg)
	m = newModel.(BuyModel)

	view := m.View()

	checks := []string{
		"CONFIRM DELETE",
		"BTC",
		"confirm",
		"cancel",
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("delete confirm view should contain '%s'", check)
		}
	}
}

func TestBuyModel_HoldingAddedMsg(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	m := NewBuyModel(p, "")

	// Simulate receiving an added message
	holding, _ := p.AddHolding("BTC", 1.0, 50000, "Test", "", "")
	msg := HoldingAddedMsg{Holding: &holding, Error: nil}
	newModel, _ := m.Update(msg)
	m = newModel.(BuyModel)

	if !strings.Contains(m.state.StatusMsg, "Added") {
		t.Error("status message should indicate holding was added")
	}

	if m.state.Mode != EntityModeList {
		t.Errorf("expected mode EntityModeList after add, got %d", m.state.Mode)
	}
}

func TestBuyModel_HoldingDeletedMsg(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	holding, _ := p.AddHolding("BTC", 1.0, 50000, "Test", "", "")

	m := NewBuyModel(p, "")

	// Simulate receiving a deleted message
	msg := HoldingDeletedMsg{ID: holding.ID, Error: nil}
	newModel, _ := m.Update(msg)
	m = newModel.(BuyModel)

	if !strings.Contains(m.state.StatusMsg, "deleted") {
		t.Error("status message should indicate holding was deleted")
	}
}

func TestBuyModel_VimKeys(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	_, _ = p.AddHolding("BTC", 1.0, 50000, "Test", "", "")
	_, _ = p.AddHolding("ETH", 10.0, 3000, "Test", "", "")

	m := NewBuyModel(p, "")

	// Test 'j' for down
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := m.Update(jMsg)
	m = newModel.(BuyModel)

	if m.state.Cursor != 1 {
		t.Errorf("expected cursor at 1 after 'j', got %d", m.state.Cursor)
	}

	// Test 'k' for up
	kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ = m.Update(kMsg)
	m = newModel.(BuyModel)

	if m.state.Cursor != 0 {
		t.Errorf("expected cursor at 0 after 'k', got %d", m.state.Cursor)
	}
}

func TestBuyModel_DefaultPlatformInForm(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	m := NewBuyModel(p, "Coinbase")

	// The platform input should have the default value (index 3 for platform)
	if m.state.Inputs[buyFieldPlatform].Value() != "Coinbase" {
		t.Errorf("expected platform default 'Coinbase', got '%s'", m.state.Inputs[buyFieldPlatform].Value())
	}
}

func TestBuyModel_NavigateFormFields(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	m := NewBuyModel(p, "")

	// Enter add mode
	aMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := m.Update(aMsg)
	m = newModel.(BuyModel)

	if m.state.FocusIndex != 0 {
		t.Errorf("expected focus at 0, got %d", m.state.FocusIndex)
	}

	// Press tab to move to next field
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ = m.Update(tabMsg)
	m = newModel.(BuyModel)

	if m.state.FocusIndex != 1 {
		t.Errorf("expected focus at 1 after tab, got %d", m.state.FocusIndex)
	}

	// Press shift+tab to go back
	shiftTabMsg := tea.KeyMsg{Type: tea.KeyShiftTab}
	newModel, _ = m.Update(shiftTabMsg)
	m = newModel.(BuyModel)

	if m.state.FocusIndex != 0 {
		t.Errorf("expected focus at 0 after shift+tab, got %d", m.state.FocusIndex)
	}
}

func TestBuyModel_GetPortfolio(t *testing.T) {
	p, cleanup := setupBuyTest(t)
	defer cleanup()

	m := NewBuyModel(p, "")

	if m.GetPortfolio() != p {
		t.Error("GetPortfolio should return the portfolio")
	}
}
