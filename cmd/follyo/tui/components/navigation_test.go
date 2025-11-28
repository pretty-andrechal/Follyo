package components

import "testing"

func TestNewListNavigator(t *testing.T) {
	nav := NewListNavigator(10)

	if nav.Cursor != 0 {
		t.Errorf("expected initial cursor 0, got %d", nav.Cursor)
	}
	if nav.Length != 10 {
		t.Errorf("expected length 10, got %d", nav.Length)
	}
}

func TestListNavigator_MoveUp(t *testing.T) {
	nav := NewListNavigator(5)
	nav.Cursor = 3

	// Should move up
	moved := nav.MoveUp()
	if !moved {
		t.Error("expected move to succeed")
	}
	if nav.Cursor != 2 {
		t.Errorf("expected cursor 2, got %d", nav.Cursor)
	}

	// Move to top
	nav.Cursor = 0
	moved = nav.MoveUp()
	if moved {
		t.Error("expected move to fail at top")
	}
	if nav.Cursor != 0 {
		t.Errorf("expected cursor 0, got %d", nav.Cursor)
	}
}

func TestListNavigator_MoveDown(t *testing.T) {
	nav := NewListNavigator(5)

	// Should move down
	moved := nav.MoveDown()
	if !moved {
		t.Error("expected move to succeed")
	}
	if nav.Cursor != 1 {
		t.Errorf("expected cursor 1, got %d", nav.Cursor)
	}

	// Move to bottom
	nav.Cursor = 4
	moved = nav.MoveDown()
	if moved {
		t.Error("expected move to fail at bottom")
	}
	if nav.Cursor != 4 {
		t.Errorf("expected cursor 4, got %d", nav.Cursor)
	}
}

func TestListNavigator_ClampCursor(t *testing.T) {
	tests := []struct {
		name           string
		cursor         int
		length         int
		expectedCursor int
	}{
		{"cursor too high", 10, 5, 4},
		{"cursor negative", -1, 5, 0},
		{"cursor valid", 2, 5, 2},
		{"empty list", 5, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nav := &ListNavigator{Cursor: tt.cursor, Length: tt.length}
			nav.ClampCursor()
			if nav.Cursor != tt.expectedCursor {
				t.Errorf("expected cursor %d, got %d", tt.expectedCursor, nav.Cursor)
			}
		})
	}
}

func TestListNavigator_SetLength(t *testing.T) {
	nav := NewListNavigator(10)
	nav.Cursor = 8

	nav.SetLength(5)

	if nav.Length != 5 {
		t.Errorf("expected length 5, got %d", nav.Length)
	}
	if nav.Cursor != 4 {
		t.Errorf("expected clamped cursor 4, got %d", nav.Cursor)
	}
}

func TestListNavigator_IsEmpty(t *testing.T) {
	nav := NewListNavigator(0)
	if !nav.IsEmpty() {
		t.Error("expected IsEmpty to return true for empty list")
	}

	nav.Length = 5
	if nav.IsEmpty() {
		t.Error("expected IsEmpty to return false for non-empty list")
	}
}

func TestListNavigator_HasSelection(t *testing.T) {
	nav := NewListNavigator(0)
	if nav.HasSelection() {
		t.Error("expected HasSelection to return false for empty list")
	}

	nav.Length = 5
	nav.Cursor = 2
	if !nav.HasSelection() {
		t.Error("expected HasSelection to return true")
	}
}

func TestListNavigator_Reset(t *testing.T) {
	nav := NewListNavigator(10)
	nav.Cursor = 5

	nav.Reset()

	if nav.Cursor != 0 {
		t.Errorf("expected cursor 0 after reset, got %d", nav.Cursor)
	}
}

func TestMoveCursorUp(t *testing.T) {
	tests := []struct {
		cursor   int
		length   int
		expected int
	}{
		{5, 10, 4},
		{0, 10, 0},
		{1, 10, 0},
	}

	for _, tt := range tests {
		result := MoveCursorUp(tt.cursor, tt.length)
		if result != tt.expected {
			t.Errorf("MoveCursorUp(%d, %d) = %d, want %d", tt.cursor, tt.length, result, tt.expected)
		}
	}
}

func TestMoveCursorDown(t *testing.T) {
	tests := []struct {
		cursor   int
		length   int
		expected int
	}{
		{5, 10, 6},
		{9, 10, 9},
		{0, 10, 1},
	}

	for _, tt := range tests {
		result := MoveCursorDown(tt.cursor, tt.length)
		if result != tt.expected {
			t.Errorf("MoveCursorDown(%d, %d) = %d, want %d", tt.cursor, tt.length, result, tt.expected)
		}
	}
}

func TestClampCursor(t *testing.T) {
	tests := []struct {
		cursor   int
		length   int
		expected int
	}{
		{5, 10, 5},    // valid
		{15, 10, 9},   // too high
		{-1, 10, 0},   // negative
		{5, 0, 0},     // empty list
		{0, 0, 0},     // empty list at zero
	}

	for _, tt := range tests {
		result := ClampCursor(tt.cursor, tt.length)
		if result != tt.expected {
			t.Errorf("ClampCursor(%d, %d) = %d, want %d", tt.cursor, tt.length, result, tt.expected)
		}
	}
}
