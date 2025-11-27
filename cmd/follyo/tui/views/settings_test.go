package views

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
	"github.com/pretty-andrechal/follyo/internal/config"
)

func setupTestConfig(t *testing.T) (*config.ConfigStore, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "follyo-settings-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	configPath := filepath.Join(tmpDir, "config.json")
	cfg, err := config.New(configPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create config: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return cfg, cleanup
}

func TestNewSettingsModel(t *testing.T) {
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	m := NewSettingsModel(cfg)

	if m.config != cfg {
		t.Error("config not set correctly")
	}

	if len(m.items) != 3 {
		t.Errorf("expected 3 setting items, got %d", len(m.items))
	}

	// Check setting keys
	expectedKeys := []string{"prices", "colors", "platform"}
	for i, key := range expectedKeys {
		if m.items[i].Key != key {
			t.Errorf("expected item %d key '%s', got '%s'", i, key, m.items[i].Key)
		}
	}
}

func TestSettingsModel_Init(t *testing.T) {
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	m := NewSettingsModel(cfg)
	cmd := m.Init()

	// Init should return nil (no async initialization needed)
	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestSettingsModel_NavigateDown(t *testing.T) {
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	m := NewSettingsModel(cfg)

	// Press down
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)
	m = newModel.(SettingsModel)

	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after down, got %d", m.cursor)
	}
}

func TestSettingsModel_NavigateUp(t *testing.T) {
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	m := NewSettingsModel(cfg)

	// Move down first
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)
	m = newModel.(SettingsModel)

	// Now press up
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = m.Update(upMsg)
	m = newModel.(SettingsModel)

	if m.cursor != 0 {
		t.Errorf("expected cursor at 0 after up, got %d", m.cursor)
	}
}

func TestSettingsModel_NavigateBoundaries(t *testing.T) {
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	m := NewSettingsModel(cfg)

	// Press up at top - should stay at 0
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := m.Update(upMsg)
	m = newModel.(SettingsModel)

	if m.cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", m.cursor)
	}

	// Navigate to bottom
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	for i := 0; i < 10; i++ { // More than items
		newModel, _ = m.Update(downMsg)
		m = newModel.(SettingsModel)
	}

	expectedBottom := len(m.items) - 1
	if m.cursor != expectedBottom {
		t.Errorf("expected cursor at %d (bottom), got %d", expectedBottom, m.cursor)
	}
}

func TestSettingsModel_ToggleSetting(t *testing.T) {
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	m := NewSettingsModel(cfg)

	// First item should be "prices" toggle, default true
	if !m.items[0].BoolValue {
		t.Error("prices should default to true")
	}

	// Press enter to toggle
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := m.Update(enterMsg)

	if cmd == nil {
		t.Fatal("expected command on toggle")
	}

	// Execute the command
	msg := cmd()
	savedMsg, ok := msg.(SettingsSavedMsg)
	if !ok {
		t.Fatalf("expected SettingsSavedMsg, got %T", msg)
	}

	if savedMsg.Key != "prices" {
		t.Errorf("expected key 'prices', got '%s'", savedMsg.Key)
	}

	if savedMsg.Error != nil {
		t.Errorf("unexpected error: %v", savedMsg.Error)
	}

	// Verify the config was updated
	if cfg.GetFetchPrices() != false {
		t.Error("expected prices to be toggled to false")
	}
}

func TestSettingsModel_BackNavigation(t *testing.T) {
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	m := NewSettingsModel(cfg)

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

func TestSettingsModel_QuitKey(t *testing.T) {
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	m := NewSettingsModel(cfg)

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

func TestSettingsModel_WindowResize(t *testing.T) {
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	m := NewSettingsModel(cfg)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	newModel, _ := m.Update(msg)
	m = newModel.(SettingsModel)

	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height 40, got %d", m.height)
	}
}

func TestSettingsModel_View(t *testing.T) {
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	m := NewSettingsModel(cfg)
	view := m.View()

	// Check for expected content
	checks := []string{
		"SETTINGS",
		"Live Prices",
		"Color Output",
		"Default Platform",
		"[ON]",
		"navigate",
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("view should contain '%s'", check)
		}
	}
}

func TestSettingsModel_TextInputMode(t *testing.T) {
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	m := NewSettingsModel(cfg)

	// Navigate to platform setting (index 2)
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(downMsg)
	m = newModel.(SettingsModel)
	newModel, _ = m.Update(downMsg)
	m = newModel.(SettingsModel)

	if m.cursor != 2 {
		t.Fatalf("expected cursor at 2, got %d", m.cursor)
	}

	// Press enter to edit
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = m.Update(enterMsg)
	m = newModel.(SettingsModel)

	if !m.editing {
		t.Error("should be in editing mode")
	}

	// Type something
	typeMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'C', 'o', 'i', 'n', 'b', 'a', 's', 'e'}}
	newModel, _ = m.Update(typeMsg)
	m = newModel.(SettingsModel)

	// Cancel with escape
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ = m.Update(escMsg)
	m = newModel.(SettingsModel)

	if m.editing {
		t.Error("should not be in editing mode after escape")
	}
}

func TestSettingsModel_SaveTextSetting(t *testing.T) {
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	m := NewSettingsModel(cfg)

	// Navigate to platform setting (index 2)
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(downMsg)
	m = newModel.(SettingsModel)
	newModel, _ = m.Update(downMsg)
	m = newModel.(SettingsModel)

	// Enter edit mode
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = m.Update(enterMsg)
	m = newModel.(SettingsModel)

	// Set value directly in text input for testing
	m.textInput.SetValue("TestPlatform")

	// Press enter to save
	newModel, cmd := m.Update(enterMsg)
	m = newModel.(SettingsModel)

	if m.editing {
		t.Error("should not be in editing mode after save")
	}

	if cmd == nil {
		t.Fatal("expected command on save")
	}

	// Execute the command
	msg := cmd()
	savedMsg, ok := msg.(SettingsSavedMsg)
	if !ok {
		t.Fatalf("expected SettingsSavedMsg, got %T", msg)
	}

	if savedMsg.Key != "platform" {
		t.Errorf("expected key 'platform', got '%s'", savedMsg.Key)
	}

	if savedMsg.Error != nil {
		t.Errorf("unexpected error: %v", savedMsg.Error)
	}

	// Verify config was updated
	if cfg.GetDefaultPlatform() != "TestPlatform" {
		t.Errorf("expected platform 'TestPlatform', got '%s'", cfg.GetDefaultPlatform())
	}
}

func TestSettingsModel_SettingsSavedMsg(t *testing.T) {
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	m := NewSettingsModel(cfg)

	// Simulate receiving a saved message
	msg := SettingsSavedMsg{Key: "prices", Error: nil}
	newModel, _ := m.Update(msg)
	m = newModel.(SettingsModel)

	if !strings.Contains(m.statusMsg, "Saved") {
		t.Error("status message should indicate setting was saved")
	}
}

func TestSettingsModel_VimKeys(t *testing.T) {
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	m := NewSettingsModel(cfg)

	// Test 'j' for down
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := m.Update(jMsg)
	m = newModel.(SettingsModel)

	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after 'j', got %d", m.cursor)
	}

	// Test 'k' for up
	kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ = m.Update(kMsg)
	m = newModel.(SettingsModel)

	if m.cursor != 0 {
		t.Errorf("expected cursor at 0 after 'k', got %d", m.cursor)
	}
}
