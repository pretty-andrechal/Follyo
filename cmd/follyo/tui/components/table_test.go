package components

import (
	"strings"
	"testing"
)

func TestNewRowRenderer(t *testing.T) {
	r := NewRowRenderer()

	if r == nil {
		t.Fatal("expected non-nil RowRenderer")
	}
	if r.CursorText != "> " {
		t.Errorf("expected cursor '> ', got '%s'", r.CursorText)
	}
	if r.NoCursorIndent != "  " {
		t.Errorf("expected indent '  ', got '%s'", r.NoCursorIndent)
	}
}

func TestRowRenderer_RenderRow_Selected(t *testing.T) {
	r := NewRowRenderer()
	result := r.RenderRow(2, 2, "test content")

	if !strings.Contains(result, "test content") {
		t.Error("rendered row should contain content")
	}
	if !strings.Contains(result, ">") {
		t.Error("selected row should contain cursor")
	}
	if !strings.HasSuffix(result, "\n") {
		t.Error("row should end with newline")
	}
}

func TestRowRenderer_RenderRow_NotSelected(t *testing.T) {
	r := NewRowRenderer()
	result := r.RenderRow(1, 2, "test content")

	if !strings.Contains(result, "test content") {
		t.Error("rendered row should contain content")
	}
	// Should have indent instead of cursor
	if strings.Contains(result, ">") {
		t.Error("non-selected row should not contain cursor")
	}
}

func TestRenderHeader(t *testing.T) {
	columns := []string{"Name", "Value", "Date"}
	widths := []int{10, 8, 12}

	result := RenderHeader(columns, widths, "  ")

	if !strings.Contains(result, "Name") {
		t.Error("header should contain 'Name'")
	}
	if !strings.Contains(result, "Value") {
		t.Error("header should contain 'Value'")
	}
	if !strings.Contains(result, "Date") {
		t.Error("header should contain 'Date'")
	}
}

func TestRenderHeader_NoWidths(t *testing.T) {
	columns := []string{"A", "B", "C"}

	// Should not panic with empty widths
	result := RenderHeader(columns, []int{}, "  ")

	if !strings.Contains(result, "A") {
		t.Error("header should contain 'A'")
	}
}

func TestRenderSeparator(t *testing.T) {
	result := RenderSeparator(20)

	// Should contain separator characters
	if !strings.Contains(result, "─") {
		t.Error("separator should contain '─' character")
	}
}

func TestRenderEmptyState(t *testing.T) {
	result := RenderEmptyState("No items found")

	if !strings.Contains(result, "No items found") {
		t.Error("empty state should contain message")
	}
}

func TestRenderTitle(t *testing.T) {
	result := RenderTitle("PURCHASES")

	if !strings.Contains(result, "PURCHASES") {
		t.Error("title should contain text")
	}
}

func TestRenderErrorTitle(t *testing.T) {
	result := RenderErrorTitle("ERROR")

	if !strings.Contains(result, "ERROR") {
		t.Error("error title should contain text")
	}
}

func TestRenderStatusMessage(t *testing.T) {
	// Non-error message
	result := RenderStatusMessage("Success!", false)
	if !strings.Contains(result, "Success!") {
		t.Error("status message should contain text")
	}

	// Error message
	errorResult := RenderStatusMessage("Failed!", true)
	if !strings.Contains(errorResult, "Failed!") {
		t.Error("error status message should contain text")
	}

	// Empty message
	emptyResult := RenderStatusMessage("", false)
	if emptyResult != "" {
		t.Error("empty message should return empty string")
	}
}

func TestRenderBox(t *testing.T) {
	content := "Box content"
	result := RenderBoxDefault(content)

	if !strings.Contains(result, content) {
		t.Error("box should contain content")
	}
	// Should contain border characters
	if !strings.Contains(result, "╭") && !strings.Contains(result, "┌") && !strings.Contains(result, "─") {
		t.Error("box should contain border characters")
	}
}

func TestRenderBoxError(t *testing.T) {
	content := "Error content"
	result := RenderBoxError(content)

	if !strings.Contains(result, content) {
		t.Error("error box should contain content")
	}
}
