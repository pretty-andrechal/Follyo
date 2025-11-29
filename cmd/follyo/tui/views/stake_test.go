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

func setupStakeTest(t *testing.T) (*portfolio.Portfolio, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "follyo-stake-test-*")
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

	// Add some holdings so staking validation passes
	_, _ = p.AddHolding("ETH", 100, 3000, "Test", "", "")
	_, _ = p.AddHolding("SOL", 500, 100, "Test", "", "")

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return p, cleanup
}

func TestNewStakeModel(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	m := NewStakeModel(p, "TestPlatform")

	if m.portfolio != p {
		t.Error("portfolio not set correctly")
	}

	if m.defaultPlatform != "TestPlatform" {
		t.Errorf("expected default platform 'TestPlatform', got '%s'", m.defaultPlatform)
	}

	if m.state.Mode != EntityModeList {
		t.Errorf("expected mode EntityModeList, got %d", m.state.Mode)
	}

	expectedFieldCount := 5 // coin, amount, platform, apy, notes
	if len(m.state.Inputs) != expectedFieldCount {
		t.Errorf("expected %d inputs, got %d", expectedFieldCount, len(m.state.Inputs))
	}
}

func TestStakeModel_Init(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	m := NewStakeModel(p, "")
	cmd := m.Init()

	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestStakeModel_NavigateDown(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	// Add some stakes
	apy1 := 4.5
	apy2 := 5.0
	_, _ = p.AddStake("ETH", 10, "Lido", &apy1, "", "")
	_, _ = p.AddStake("SOL", 100, "Marinade", &apy2, "", "")

	m := NewStakeModel(p, "")

	// Press down
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)
	m = newModel.(StakeModel)

	if m.state.Cursor != 1 {
		t.Errorf("expected cursor at 1 after down, got %d", m.state.Cursor)
	}
}

func TestStakeModel_NavigateUp(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	// Add some stakes
	apy1 := 4.5
	apy2 := 5.0
	_, _ = p.AddStake("ETH", 10, "Lido", &apy1, "", "")
	_, _ = p.AddStake("SOL", 100, "Marinade", &apy2, "", "")

	m := NewStakeModel(p, "")

	// Move down first
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)
	m = newModel.(StakeModel)

	// Now press up
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = m.Update(upMsg)
	m = newModel.(StakeModel)

	if m.state.Cursor != 0 {
		t.Errorf("expected cursor at 0 after up, got %d", m.state.Cursor)
	}
}

func TestStakeModel_NavigateBoundaries(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	// Add some stakes
	apy := 4.5
	_, _ = p.AddStake("ETH", 10, "Lido", &apy, "", "")
	_, _ = p.AddStake("SOL", 100, "Marinade", &apy, "", "")

	m := NewStakeModel(p, "")

	// Press up at top - should stay at 0
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := m.Update(upMsg)
	m = newModel.(StakeModel)

	if m.state.Cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", m.state.Cursor)
	}

	// Navigate to bottom
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	for i := 0; i < 10; i++ {
		newModel, _ = m.Update(downMsg)
		m = newModel.(StakeModel)
	}

	expectedBottom := len(m.stakes) - 1
	if m.state.Cursor != expectedBottom {
		t.Errorf("expected cursor at %d (bottom), got %d", expectedBottom, m.state.Cursor)
	}
}

func TestStakeModel_BackNavigation(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	m := NewStakeModel(p, "")

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

func TestStakeModel_QuitKey(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	m := NewStakeModel(p, "")

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

func TestStakeModel_WindowResize(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	m := NewStakeModel(p, "")

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	newModel, _ := m.Update(msg)
	m = newModel.(StakeModel)

	if m.state.Width != 120 {
		t.Errorf("expected width 120, got %d", m.state.Width)
	}
	if m.state.Height != 40 {
		t.Errorf("expected height 40, got %d", m.state.Height)
	}
}

func TestStakeModel_ViewListEmpty(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	m := NewStakeModel(p, "")
	view := m.View()

	checks := []string{
		"STAKING",
		"No staked positions",
		"add",
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("view should contain '%s'", check)
		}
	}
}

func TestStakeModel_ViewListWithStakes(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	apy := 4.5
	_, _ = p.AddStake("ETH", 10, "Lido", &apy, "", "")

	m := NewStakeModel(p, "")
	view := m.View()

	checks := []string{
		"STAKING",
		"ETH",
		"Lido",
		"4.50%",
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("view should contain '%s'", check)
		}
	}
}

func TestStakeModel_EnterAddMode(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	m := NewStakeModel(p, "")

	// Press 'a' to add
	aMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := m.Update(aMsg)
	m = newModel.(StakeModel)

	if m.state.Mode != EntityModeAdd {
		t.Errorf("expected mode EntityModeAdd, got %d", m.state.Mode)
	}
}

func TestStakeModel_CancelAddMode(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	m := NewStakeModel(p, "")

	// Enter add mode
	aMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := m.Update(aMsg)
	m = newModel.(StakeModel)

	// Cancel with escape
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ = m.Update(escMsg)
	m = newModel.(StakeModel)

	if m.state.Mode != EntityModeList {
		t.Errorf("expected mode EntityModeList after cancel, got %d", m.state.Mode)
	}
}

func TestStakeModel_AddFormView(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	m := NewStakeModel(p, "")

	// Enter add mode
	aMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := m.Update(aMsg)
	m = newModel.(StakeModel)

	view := m.View()

	checks := []string{
		"ADD STAKE",
		"Coin:",
		"Amount:",
		"Platform:",
		"APY",
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

func TestStakeModel_DeleteConfirmMode(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	apy := 4.5
	_, _ = p.AddStake("ETH", 10, "Lido", &apy, "", "")

	m := NewStakeModel(p, "")

	// Press 'd' to delete
	dMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := m.Update(dMsg)
	m = newModel.(StakeModel)

	if m.state.Mode != EntityModeConfirmDelete {
		t.Errorf("expected mode EntityModeConfirmDelete, got %d", m.state.Mode)
	}
}

func TestStakeModel_CancelDelete(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	apy := 4.5
	_, _ = p.AddStake("ETH", 10, "Lido", &apy, "", "")

	m := NewStakeModel(p, "")

	// Enter delete confirm mode
	dMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := m.Update(dMsg)
	m = newModel.(StakeModel)

	// Cancel with 'n'
	nMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ = m.Update(nMsg)
	m = newModel.(StakeModel)

	if m.state.Mode != EntityModeList {
		t.Errorf("expected mode EntityModeList after cancel, got %d", m.state.Mode)
	}
}

func TestStakeModel_ConfirmDelete(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	apy := 4.5
	_, _ = p.AddStake("ETH", 10, "Lido", &apy, "", "")

	m := NewStakeModel(p, "")

	initialCount := len(m.stakes)
	if initialCount != 1 {
		t.Fatalf("expected 1 stake, got %d", initialCount)
	}

	// Enter delete confirm mode
	dMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := m.Update(dMsg)
	m = newModel.(StakeModel)

	// Confirm with 'y'
	yMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	_, cmd := m.Update(yMsg)

	if cmd == nil {
		t.Fatal("expected command on delete confirm")
	}

	// Execute the command
	msg := cmd()
	deletedMsg, ok := msg.(StakeDeletedMsg)
	if !ok {
		t.Fatalf("expected StakeDeletedMsg, got %T", msg)
	}

	if deletedMsg.Error != nil {
		t.Errorf("unexpected error: %v", deletedMsg.Error)
	}
}

func TestStakeModel_DeleteConfirmView(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	apy := 4.5
	_, _ = p.AddStake("ETH", 10, "Lido", &apy, "", "")

	m := NewStakeModel(p, "")

	// Enter delete confirm mode
	dMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := m.Update(dMsg)
	m = newModel.(StakeModel)

	view := m.View()

	checks := []string{
		"CONFIRM DELETE",
		"ETH",
		"Lido",
		"confirm",
		"cancel",
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("delete confirm view should contain '%s'", check)
		}
	}
}

func TestStakeModel_StakeAddedMsg(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	m := NewStakeModel(p, "")

	// Simulate receiving an added message
	apy := 4.5
	stake, _ := p.AddStake("ETH", 10, "Lido", &apy, "", "")
	msg := StakeAddedMsg{Stake: &stake, Error: nil}
	newModel, _ := m.Update(msg)
	m = newModel.(StakeModel)

	if !strings.Contains(m.state.StatusMsg, "Staked") {
		t.Error("status message should indicate stake was added")
	}

	if m.state.Mode != EntityModeList {
		t.Errorf("expected mode EntityModeList after add, got %d", m.state.Mode)
	}
}

func TestStakeModel_StakeDeletedMsg(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	apy := 4.5
	stake, _ := p.AddStake("ETH", 10, "Lido", &apy, "", "")

	m := NewStakeModel(p, "")

	// Simulate receiving a deleted message
	msg := StakeDeletedMsg{ID: stake.ID, Error: nil}
	newModel, _ := m.Update(msg)
	m = newModel.(StakeModel)

	if !strings.Contains(m.state.StatusMsg, "removed") {
		t.Error("status message should indicate stake was removed")
	}
}

func TestStakeModel_VimKeys(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	apy := 4.5
	_, _ = p.AddStake("ETH", 10, "Lido", &apy, "", "")
	_, _ = p.AddStake("SOL", 100, "Marinade", &apy, "", "")

	m := NewStakeModel(p, "")

	// Test 'j' for down
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := m.Update(jMsg)
	m = newModel.(StakeModel)

	if m.state.Cursor != 1 {
		t.Errorf("expected cursor at 1 after 'j', got %d", m.state.Cursor)
	}

	// Test 'k' for up
	kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ = m.Update(kMsg)
	m = newModel.(StakeModel)

	if m.state.Cursor != 0 {
		t.Errorf("expected cursor at 0 after 'k', got %d", m.state.Cursor)
	}
}

func TestStakeModel_DefaultPlatformInForm(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	m := NewStakeModel(p, "Lido")

	// The platform input should have the default value
	if m.state.Inputs[stakeFieldPlatform].Value() != "Lido" {
		t.Errorf("expected platform default 'Lido', got '%s'", m.state.Inputs[stakeFieldPlatform].Value())
	}
}

func TestStakeModel_NavigateFormFields(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	m := NewStakeModel(p, "")

	// Enter add mode
	aMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := m.Update(aMsg)
	m = newModel.(StakeModel)

	if m.state.FocusIndex != 0 {
		t.Errorf("expected focus at 0, got %d", m.state.FocusIndex)
	}

	// Press tab to move to next field
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ = m.Update(tabMsg)
	m = newModel.(StakeModel)

	if m.state.FocusIndex != 1 {
		t.Errorf("expected focus at 1 after tab, got %d", m.state.FocusIndex)
	}

	// Press shift+tab to go back
	shiftTabMsg := tea.KeyMsg{Type: tea.KeyShiftTab}
	newModel, _ = m.Update(shiftTabMsg)
	m = newModel.(StakeModel)

	if m.state.FocusIndex != 0 {
		t.Errorf("expected focus at 0 after shift+tab, got %d", m.state.FocusIndex)
	}
}

func TestStakeModel_GetPortfolio(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	m := NewStakeModel(p, "")

	if m.GetPortfolio() != p {
		t.Error("GetPortfolio should return the portfolio")
	}
}

func TestStakeModel_APYOptional(t *testing.T) {
	p, cleanup := setupStakeTest(t)
	defer cleanup()

	// Add a stake without APY
	_, _ = p.AddStake("ETH", 10, "Lido", nil, "", "")

	m := NewStakeModel(p, "")
	view := m.View()

	// Should show "-" for no APY
	if !strings.Contains(view, "-") {
		t.Error("view should show '-' for stake without APY")
	}
}
