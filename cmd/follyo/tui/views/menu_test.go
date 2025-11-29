package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
)

func TestNewMenuModel(t *testing.T) {
	m := NewMenuModel()

	if len(m.items) != 9 {
		t.Errorf("expected 9 menu items, got %d", len(m.items))
	}

	if m.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.cursor)
	}

	// Check that all items have required fields
	for i, item := range m.items {
		if item.Title == "" {
			t.Errorf("item %d has empty title", i)
		}
		if item.Description == "" {
			t.Errorf("item %d has empty description", i)
		}
		if item.Action == "" {
			t.Errorf("item %d has empty action", i)
		}
	}
}

func TestMenuModel_NavigateDown(t *testing.T) {
	m := NewMenuModel()

	// Press down key
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)
	m = newModel.(MenuModel)

	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after down, got %d", m.cursor)
	}

	// Press down again
	newModel, _ = m.Update(msg)
	m = newModel.(MenuModel)

	if m.cursor != 2 {
		t.Errorf("expected cursor at 2 after second down, got %d", m.cursor)
	}
}

func TestMenuModel_NavigateUp(t *testing.T) {
	m := NewMenuModel()

	// Move down first
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)
	m = newModel.(MenuModel)
	newModel, _ = m.Update(msg)
	m = newModel.(MenuModel)

	// Now at position 2, press up
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = m.Update(upMsg)
	m = newModel.(MenuModel)

	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after up, got %d", m.cursor)
	}
}

func TestMenuModel_NavigateBoundaries(t *testing.T) {
	m := NewMenuModel()

	// Press up at top - should stay at 0
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := m.Update(upMsg)
	m = newModel.(MenuModel)

	if m.cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", m.cursor)
	}

	// Navigate to bottom
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	for i := 0; i < 10; i++ { // More than menu items
		newModel, _ = m.Update(downMsg)
		m = newModel.(MenuModel)
	}

	expectedBottom := len(m.items) - 1
	if m.cursor != expectedBottom {
		t.Errorf("expected cursor at %d (bottom), got %d", expectedBottom, m.cursor)
	}
}

func TestMenuModel_SelectItem(t *testing.T) {
	m := NewMenuModel()

	// Press enter to select
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := m.Update(enterMsg)

	if cmd == nil {
		t.Fatal("expected a command to be returned on select")
	}

	// Execute the command to get the message
	msg := cmd()
	selectMsg, ok := msg.(tui.MenuSelectMsg)
	if !ok {
		t.Fatalf("expected MenuSelectMsg, got %T", msg)
	}

	if selectMsg.Action != "summary" {
		t.Errorf("expected action 'summary', got '%s'", selectMsg.Action)
	}
}

func TestMenuModel_SelectDifferentItems(t *testing.T) {
	tests := []struct {
		cursorPos      int
		expectedAction string
	}{
		{0, "summary"},
		{1, "buy"},
		{2, "sell"},
		{3, "stake"},
		{4, "loan"},
		{5, "snapshots"},
		{6, "coinhistory"},
		{7, "ticker"},
		{8, "settings"},
	}

	for _, tt := range tests {
		m := NewMenuModel()

		// Navigate to position
		downMsg := tea.KeyMsg{Type: tea.KeyDown}
		for i := 0; i < tt.cursorPos; i++ {
			newModel, _ := m.Update(downMsg)
			m = newModel.(MenuModel)
		}

		// Select
		enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
		_, cmd := m.Update(enterMsg)

		msg := cmd()
		selectMsg := msg.(tui.MenuSelectMsg)

		if selectMsg.Action != tt.expectedAction {
			t.Errorf("at position %d: expected action '%s', got '%s'",
				tt.cursorPos, tt.expectedAction, selectMsg.Action)
		}
	}
}

func TestMenuModel_View(t *testing.T) {
	m := NewMenuModel()
	view := m.View()

	// Check that the view contains expected elements
	// Logo uses ASCII art with box drawing characters
	if !strings.Contains(view, "███") {
		t.Error("view should contain ASCII logo")
	}

	if !strings.Contains(view, "Portfolio Summary") {
		t.Error("view should contain 'Portfolio Summary' menu item")
	}

	if !strings.Contains(view, "navigate") {
		t.Error("view should contain help text")
	}
}

func TestMenuModel_WindowResize(t *testing.T) {
	m := NewMenuModel()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := m.Update(msg)
	m = newModel.(MenuModel)

	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("expected height 50, got %d", m.height)
	}
}

func TestMenuModel_SelectedItem(t *testing.T) {
	m := NewMenuModel()

	item := m.SelectedItem()
	if item.Title != "Portfolio Summary" {
		t.Errorf("expected first item to be Portfolio Summary, got %s", item.Title)
	}

	// Navigate down and check again
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(downMsg)
	m = newModel.(MenuModel)

	item = m.SelectedItem()
	if item.Title != "Buy" {
		t.Errorf("expected second item to be Buy, got %s", item.Title)
	}
}

func TestMenuModel_VimKeys(t *testing.T) {
	m := NewMenuModel()

	// Test 'j' for down
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := m.Update(jMsg)
	m = newModel.(MenuModel)

	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after 'j', got %d", m.cursor)
	}

	// Test 'k' for up
	kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ = m.Update(kMsg)
	m = newModel.(MenuModel)

	if m.cursor != 0 {
		t.Errorf("expected cursor at 0 after 'k', got %d", m.cursor)
	}
}
