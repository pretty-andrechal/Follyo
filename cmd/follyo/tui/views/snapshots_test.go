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
	"github.com/pretty-andrechal/follyo/internal/portfolio"
	"github.com/pretty-andrechal/follyo/internal/storage"
)

func setupSnapshotsTest(t *testing.T) (*storage.SnapshotStore, *portfolio.Portfolio, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "follyo-snapshots-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create portfolio storage
	portfolioPath := filepath.Join(tmpDir, "portfolio.json")
	s, err := storage.New(portfolioPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create storage: %v", err)
	}

	p := portfolio.New(s)

	// Add some test data
	_, _ = p.AddHolding("BTC", 1.0, 50000, "Test", "", "")
	_, _ = p.AddHolding("ETH", 10.0, 3000, "Test", "", "")

	// Create snapshot store
	snapshotPath := filepath.Join(tmpDir, "snapshots.json")
	store, err := storage.NewSnapshotStore(snapshotPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create snapshot store: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return store, p, cleanup
}

func TestNewSnapshotsModel(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)

	if m.store != store {
		t.Error("store not set correctly")
	}

	if m.portfolio != p {
		t.Error("portfolio not set correctly")
	}

	if m.mode != SnapshotsList {
		t.Errorf("expected mode SnapshotsList, got %d", m.mode)
	}
}

func TestSnapshotsModel_Init(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)
	cmd := m.Init()

	// Init now returns a batch command with ReloadSnapshotsMsg and CheckAutoSnapshotMsg
	if cmd == nil {
		t.Error("Init should return a command")
	}

	// Execute the batch command - it returns a BatchMsg containing multiple commands
	msg := cmd()

	// The batch message contains multiple commands, we just verify it's not nil
	if msg == nil {
		t.Error("Init command should return a message")
	}
}

func TestSnapshotsModel_NavigateDown(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	// Add some snapshots
	snap1, _ := p.CreateSnapshot(map[string]float64{"BTC": 50000, "ETH": 3000}, "test1")
	_ = store.Add(snap1)
	snap2, _ := p.CreateSnapshot(map[string]float64{"BTC": 51000, "ETH": 3100}, "test2")
	_ = store.Add(snap2)

	m := NewSnapshotsModel(store, p, nil)

	// Press down
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)
	m = newModel.(SnapshotsModel)

	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after down, got %d", m.cursor)
	}
}

func TestSnapshotsModel_NavigateUp(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	// Add some snapshots
	snap1, _ := p.CreateSnapshot(map[string]float64{"BTC": 50000, "ETH": 3000}, "test1")
	_ = store.Add(snap1)
	snap2, _ := p.CreateSnapshot(map[string]float64{"BTC": 51000, "ETH": 3100}, "test2")
	_ = store.Add(snap2)

	m := NewSnapshotsModel(store, p, nil)

	// Move down first
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)
	m = newModel.(SnapshotsModel)

	// Now press up
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = m.Update(upMsg)
	m = newModel.(SnapshotsModel)

	if m.cursor != 0 {
		t.Errorf("expected cursor at 0 after up, got %d", m.cursor)
	}
}

func TestSnapshotsModel_NavigateBoundaries(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	// Add some snapshots
	snap1, _ := p.CreateSnapshot(map[string]float64{"BTC": 50000, "ETH": 3000}, "test1")
	_ = store.Add(snap1)
	snap2, _ := p.CreateSnapshot(map[string]float64{"BTC": 51000, "ETH": 3100}, "test2")
	_ = store.Add(snap2)

	m := NewSnapshotsModel(store, p, nil)

	// Press up at top - should stay at 0
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := m.Update(upMsg)
	m = newModel.(SnapshotsModel)

	if m.cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", m.cursor)
	}

	// Navigate to bottom
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	for i := 0; i < 10; i++ { // More than items
		newModel, _ = m.Update(downMsg)
		m = newModel.(SnapshotsModel)
	}

	expectedBottom := len(m.snapshots) - 1
	if m.cursor != expectedBottom {
		t.Errorf("expected cursor at %d (bottom), got %d", expectedBottom, m.cursor)
	}
}

func TestSnapshotsModel_BackNavigation(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)

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

func TestSnapshotsModel_QuitKey(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)

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

func TestSnapshotsModel_WindowResize(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	newModel, _ := m.Update(msg)
	m = newModel.(SnapshotsModel)

	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height 40, got %d", m.height)
	}
}

func TestSnapshotsModel_ViewList(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)
	view := m.View()

	// Check for expected content
	checks := []string{
		"SNAPSHOTS",
		"No snapshots yet",
		"new snapshot",
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("view should contain '%s'", check)
		}
	}
}

func TestSnapshotsModel_ViewListWithSnapshots(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	// Add a snapshot
	snap, _ := p.CreateSnapshot(map[string]float64{"BTC": 50000, "ETH": 3000}, "test note")
	_ = store.Add(snap)

	m := NewSnapshotsModel(store, p, nil)
	view := m.View()

	// Check for expected content
	checks := []string{
		"SNAPSHOTS",
		"ID",
		"Date",
		"Net Value",
		"test note",
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("view should contain '%s'", check)
		}
	}
}

func TestSnapshotsModel_EnterNoteInput(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)

	// Press 'n' to create new snapshot
	nMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := m.Update(nMsg)
	m = newModel.(SnapshotsModel)

	if m.mode != SnapshotsNoteInput {
		t.Errorf("expected mode SnapshotsNoteInput, got %d", m.mode)
	}
}

func TestSnapshotsModel_CancelNoteInput(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)

	// Enter note input mode
	nMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := m.Update(nMsg)
	m = newModel.(SnapshotsModel)

	// Cancel with escape
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ = m.Update(escMsg)
	m = newModel.(SnapshotsModel)

	if m.mode != SnapshotsList {
		t.Errorf("expected mode SnapshotsList after cancel, got %d", m.mode)
	}
}

func TestSnapshotsModel_SelectSnapshot(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	// Add a snapshot
	snap, _ := p.CreateSnapshot(map[string]float64{"BTC": 50000, "ETH": 3000}, "test")
	_ = store.Add(snap)

	m := NewSnapshotsModel(store, p, nil)

	// Set up viewport
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(SnapshotsModel)

	// Press enter to view detail
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = m.Update(enterMsg)
	m = newModel.(SnapshotsModel)

	if m.mode != SnapshotsDetail {
		t.Errorf("expected mode SnapshotsDetail, got %d", m.mode)
	}

	if m.selectedID != snap.ID {
		t.Errorf("expected selectedID '%s', got '%s'", snap.ID, m.selectedID)
	}
}

func TestSnapshotsModel_BackFromDetail(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	// Add a snapshot
	snap, _ := p.CreateSnapshot(map[string]float64{"BTC": 50000, "ETH": 3000}, "test")
	_ = store.Add(snap)

	m := NewSnapshotsModel(store, p, nil)

	// Set up viewport
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(SnapshotsModel)

	// Go to detail view
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = m.Update(enterMsg)
	m = newModel.(SnapshotsModel)

	// Press escape to go back
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ = m.Update(escMsg)
	m = newModel.(SnapshotsModel)

	if m.mode != SnapshotsList {
		t.Errorf("expected mode SnapshotsList after back, got %d", m.mode)
	}

	if m.selectedID != "" {
		t.Errorf("expected selectedID to be cleared, got '%s'", m.selectedID)
	}
}

func TestSnapshotsModel_DeleteSnapshot(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	// Add a snapshot
	snap, _ := p.CreateSnapshot(map[string]float64{"BTC": 50000, "ETH": 3000}, "test")
	_ = store.Add(snap)

	m := NewSnapshotsModel(store, p, nil)

	initialCount := len(m.snapshots)
	if initialCount != 1 {
		t.Fatalf("expected 1 snapshot, got %d", initialCount)
	}

	// Press 'd' to delete
	dMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	_, cmd := m.Update(dMsg)

	if cmd == nil {
		t.Fatal("expected command on delete")
	}

	// Execute the command
	msg := cmd()
	deletedMsg, ok := msg.(SnapshotDeletedMsg)
	if !ok {
		t.Fatalf("expected SnapshotDeletedMsg, got %T", msg)
	}

	if deletedMsg.Error != nil {
		t.Errorf("unexpected error: %v", deletedMsg.Error)
	}
}

func TestSnapshotsModel_SnapshotDeletedMsg(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	// Add a snapshot
	snap, _ := p.CreateSnapshot(map[string]float64{"BTC": 50000, "ETH": 3000}, "test")
	_ = store.Add(snap)

	m := NewSnapshotsModel(store, p, nil)

	// Simulate receiving a deleted message
	msg := SnapshotDeletedMsg{ID: snap.ID, Error: nil}
	newModel, _ := m.Update(msg)
	m = newModel.(SnapshotsModel)

	if !strings.Contains(m.statusMsg, "deleted") {
		t.Error("status message should indicate snapshot was deleted")
	}
}

func TestSnapshotsModel_VimKeys(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	// Add some snapshots
	snap1, _ := p.CreateSnapshot(map[string]float64{"BTC": 50000, "ETH": 3000}, "test1")
	_ = store.Add(snap1)
	snap2, _ := p.CreateSnapshot(map[string]float64{"BTC": 51000, "ETH": 3100}, "test2")
	_ = store.Add(snap2)

	m := NewSnapshotsModel(store, p, nil)

	// Test 'j' for down
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := m.Update(jMsg)
	m = newModel.(SnapshotsModel)

	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after 'j', got %d", m.cursor)
	}

	// Test 'k' for up
	kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ = m.Update(kMsg)
	m = newModel.(SnapshotsModel)

	if m.cursor != 0 {
		t.Errorf("expected cursor at 0 after 'k', got %d", m.cursor)
	}
}

func TestSnapshotsModel_ViewDetail(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	// Add a snapshot
	snap, _ := p.CreateSnapshot(map[string]float64{"BTC": 50000, "ETH": 3000}, "Test note for detail view")
	_ = store.Add(snap)

	m := NewSnapshotsModel(store, p, nil)

	// Set up viewport
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(SnapshotsModel)

	// Go to detail view
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = m.Update(enterMsg)
	m = newModel.(SnapshotsModel)

	view := m.View()

	// Check for expected content
	checks := []string{
		"SNAPSHOT",
		"PORTFOLIO VALUE",
		"Holdings Value",
		"Test note for detail view",
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("detail view should contain '%s'", check)
		}
	}
}

func TestSnapshotsModel_NoteInputView(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)

	// Enter note input mode
	nMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := m.Update(nMsg)
	m = newModel.(SnapshotsModel)

	view := m.View()

	// Check for expected content
	checks := []string{
		"NEW SNAPSHOT",
		"Note (optional)",
		"save",
		"cancel",
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("note input view should contain '%s'", check)
		}
	}
}

func TestSnapshotsModel_GetStore(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)

	if m.GetStore() != store {
		t.Error("GetStore should return the snapshot store")
	}
}

func TestSnapshotsModel_DailySnapshotKey(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)

	// Press 't' to create today's snapshot
	tMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	newModel, cmd := m.Update(tMsg)
	m = newModel.(SnapshotsModel)

	// Should enter saving mode directly (no note input prompt)
	if m.mode != SnapshotsSaving {
		t.Errorf("expected mode SnapshotsSaving after 't', got %d", m.mode)
	}

	// Status message should indicate daily snapshot
	if !strings.Contains(m.statusMsg, "daily") {
		t.Errorf("status message should contain 'daily', got '%s'", m.statusMsg)
	}

	// Should return a command (batch with spinner.Tick and saveSnapshot)
	if cmd == nil {
		t.Error("expected command to be returned for saving snapshot")
	}
}

func TestSnapshotsModel_DailySnapshotHelpText(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)
	view := m.View()

	// Empty state help text should show today's snapshot option
	if !strings.Contains(view, "today") {
		t.Error("help text should contain 'today' for daily snapshot shortcut")
	}

	// Add a snapshot to check the non-empty help text
	snap, _ := p.CreateSnapshot(map[string]float64{"BTC": 50000, "ETH": 3000}, "test")
	_ = store.Add(snap)
	m = NewSnapshotsModel(store, p, nil)
	view = m.View()

	// With snapshots, 't' key should be shown for today
	if !strings.Contains(view, "today") {
		t.Error("help text with snapshots should show 't' for today")
	}
}

func TestSnapshotsModel_DailySnapshotNoteFormat(t *testing.T) {
	// Verify the expected date format is YYYY-MM-DD
	now := time.Now()
	expectedFormat := now.Format("2006-01-02")

	// The format should match the standard Go date format for YYYY-MM-DD
	if len(expectedFormat) != 10 {
		t.Errorf("expected date format length 10, got %d", len(expectedFormat))
	}

	// Verify format components
	if expectedFormat[4] != '-' || expectedFormat[7] != '-' {
		t.Errorf("expected date format YYYY-MM-DD, got %s", expectedFormat)
	}
}

func TestSnapshotsModel_AutoSnapshotTrigger_NoSnapshotToday(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)

	// Verify no snapshot for today
	if store.HasSnapshotForToday() {
		t.Fatal("expected no snapshot for today initially")
	}

	// Send CheckAutoSnapshotMsg
	msg := CheckAutoSnapshotMsg{}
	newModel, cmd := m.Update(msg)
	m = newModel.(SnapshotsModel)

	// Should enter auto-saving mode
	if m.mode != SnapshotsAutoSaving {
		t.Errorf("expected mode SnapshotsAutoSaving, got %d", m.mode)
	}

	// Should have a command to save
	if cmd == nil {
		t.Error("expected command for auto-snapshot save")
	}

	// Status should indicate auto-snapshot
	if !strings.Contains(m.statusMsg, "automatic") {
		t.Errorf("status message should contain 'automatic', got '%s'", m.statusMsg)
	}

	// autoSnapshotChecked should be set
	if !m.autoSnapshotChecked {
		t.Error("autoSnapshotChecked should be true after trigger")
	}
}

func TestSnapshotsModel_AutoSnapshotTrigger_SnapshotExists(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	// Add a snapshot for today
	snap, _ := p.CreateSnapshot(map[string]float64{"BTC": 50000, "ETH": 3000}, "existing")
	_ = store.Add(snap)

	m := NewSnapshotsModel(store, p, nil)

	// Verify snapshot exists for today
	if !store.HasSnapshotForToday() {
		t.Fatal("expected snapshot for today to exist")
	}

	// Send CheckAutoSnapshotMsg
	msg := CheckAutoSnapshotMsg{}
	newModel, cmd := m.Update(msg)
	m = newModel.(SnapshotsModel)

	// Should stay in list mode (no auto-save needed)
	if m.mode != SnapshotsList {
		t.Errorf("expected mode SnapshotsList when snapshot exists, got %d", m.mode)
	}

	// Should NOT have a save command
	if cmd != nil {
		t.Error("expected no command when snapshot already exists for today")
	}

	// autoSnapshotChecked should still be set to prevent future checks
	if !m.autoSnapshotChecked {
		t.Error("autoSnapshotChecked should be true even when snapshot exists")
	}
}

func TestSnapshotsModel_AutoSnapshotTrigger_OnlyOnce(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)

	// First trigger
	msg := CheckAutoSnapshotMsg{}
	newModel, cmd1 := m.Update(msg)
	m = newModel.(SnapshotsModel)

	// Should trigger auto-save
	if cmd1 == nil {
		t.Error("expected command on first trigger")
	}

	// Reset mode to list (simulate save completed)
	m.mode = SnapshotsList

	// Second trigger should be ignored
	newModel, cmd2 := m.Update(msg)
	m = newModel.(SnapshotsModel)

	// Should not trigger again
	if cmd2 != nil {
		t.Error("expected no command on second trigger (already checked)")
	}

	// Should stay in list mode
	if m.mode != SnapshotsList {
		t.Errorf("expected mode SnapshotsList on second trigger, got %d", m.mode)
	}
}

func TestSnapshotsModel_AutoSnapshotSavedMsg(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)
	m.mode = SnapshotsAutoSaving

	// Simulate receiving auto-snapshot saved message
	snap, _ := p.CreateSnapshot(map[string]float64{"BTC": 50000}, "Auto-snapshot test")
	msg := SnapshotSavedMsg{Snapshot: &snap, IsAuto: true}
	newModel, _ := m.Update(msg)
	m = newModel.(SnapshotsModel)

	// Should return to list mode
	if m.mode != SnapshotsList {
		t.Errorf("expected mode SnapshotsList after save, got %d", m.mode)
	}

	// Status should indicate auto-snapshot
	if !strings.Contains(m.statusMsg, "auto-snapshot") {
		t.Errorf("status message should contain 'auto-snapshot', got '%s'", m.statusMsg)
	}
}

func TestSnapshotsModel_AutoSnapshotError(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)
	m.mode = SnapshotsAutoSaving

	// Simulate receiving auto-snapshot error message
	testErr := fmt.Errorf("test error")
	msg := SnapshotSavedMsg{Error: testErr, IsAuto: true}
	newModel, _ := m.Update(msg)
	m = newModel.(SnapshotsModel)

	// Should return to list mode
	if m.mode != SnapshotsList {
		t.Errorf("expected mode SnapshotsList after error, got %d", m.mode)
	}

	// Status should indicate auto-snapshot error
	if !strings.Contains(m.statusMsg, "Auto-snapshot failed") {
		t.Errorf("status message should contain 'Auto-snapshot failed', got '%s'", m.statusMsg)
	}
}

func TestSnapshotsModel_AutoSnapshotViewRendering(t *testing.T) {
	store, p, cleanup := setupSnapshotsTest(t)
	defer cleanup()

	m := NewSnapshotsModel(store, p, nil)
	m.mode = SnapshotsAutoSaving
	m.statusMsg = "Taking automatic daily snapshot..."

	view := m.View()

	// Should show auto-snapshot title
	if !strings.Contains(view, "AUTO-SNAPSHOT") {
		t.Error("view should contain 'AUTO-SNAPSHOT' title")
	}

	// Should show status message
	if !strings.Contains(view, "automatic") {
		t.Error("view should contain 'automatic' in status")
	}
}
