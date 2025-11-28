package components

// ListNavigator handles cursor movement with bounds checking.
type ListNavigator struct {
	Cursor int
	Length int
}

// NewListNavigator creates a new list navigator.
func NewListNavigator(length int) *ListNavigator {
	return &ListNavigator{
		Cursor: 0,
		Length: length,
	}
}

// MoveUp moves cursor up, respecting bounds.
// Returns true if the cursor actually moved.
func (n *ListNavigator) MoveUp() bool {
	if n.Cursor > 0 {
		n.Cursor--
		return true
	}
	return false
}

// MoveDown moves cursor down, respecting bounds.
// Returns true if the cursor actually moved.
func (n *ListNavigator) MoveDown() bool {
	if n.Cursor < n.Length-1 {
		n.Cursor++
		return true
	}
	return false
}

// ClampCursor ensures cursor is within valid range after list changes.
// Call this after adding or removing items from the list.
func (n *ListNavigator) ClampCursor() {
	if n.Length == 0 {
		n.Cursor = 0
		return
	}
	if n.Cursor >= n.Length {
		n.Cursor = n.Length - 1
	}
	if n.Cursor < 0 {
		n.Cursor = 0
	}
}

// SetLength updates the list length and clamps the cursor.
func (n *ListNavigator) SetLength(length int) {
	n.Length = length
	n.ClampCursor()
}

// IsEmpty returns true if the list is empty.
func (n *ListNavigator) IsEmpty() bool {
	return n.Length == 0
}

// HasSelection returns true if there's a valid selection.
func (n *ListNavigator) HasSelection() bool {
	return n.Length > 0 && n.Cursor >= 0 && n.Cursor < n.Length
}

// Reset resets the cursor to the beginning.
func (n *ListNavigator) Reset() {
	n.Cursor = 0
}

// MoveCursorUp is a helper function for simple cursor up operations.
// Returns the new cursor position.
func MoveCursorUp(cursor, length int) int {
	if cursor > 0 {
		return cursor - 1
	}
	return cursor
}

// MoveCursorDown is a helper function for simple cursor down operations.
// Returns the new cursor position.
func MoveCursorDown(cursor, length int) int {
	if cursor < length-1 {
		return cursor + 1
	}
	return cursor
}

// ClampCursor is a helper function to clamp a cursor value.
// Returns the clamped cursor position.
func ClampCursor(cursor, length int) int {
	if length == 0 {
		return 0
	}
	if cursor >= length {
		return length - 1
	}
	if cursor < 0 {
		return 0
	}
	return cursor
}
