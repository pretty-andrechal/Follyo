package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
)

// RowRenderer handles common table row rendering patterns.
type RowRenderer struct {
	CursorText     string
	CursorStyle    lipgloss.Style
	SelectedStyle  lipgloss.Style
	NormalStyle    lipgloss.Style
	NoCursorIndent string
}

// NewRowRenderer creates a row renderer with default TUI styles.
func NewRowRenderer() *RowRenderer {
	return &RowRenderer{
		CursorText:     "> ",
		CursorStyle:    lipgloss.NewStyle().Foreground(tui.PrimaryColor),
		SelectedStyle:  lipgloss.NewStyle().Bold(true).Foreground(tui.PrimaryColor),
		NormalStyle:    lipgloss.NewStyle().Foreground(tui.TextColor),
		NoCursorIndent: "  ",
	}
}

// RenderRow renders a table row with cursor and selection state.
// If index == cursor, the row is rendered as selected with a cursor indicator.
func (r *RowRenderer) RenderRow(index, cursor int, content string) string {
	isSelected := index == cursor

	var prefix string
	var style lipgloss.Style

	if isSelected {
		prefix = r.CursorStyle.Render(r.CursorText)
		style = r.SelectedStyle
	} else {
		prefix = r.NoCursorIndent
		style = r.NormalStyle
	}

	return prefix + style.Render(content) + "\n"
}

// RenderHeader renders a table header with column names.
func RenderHeader(columns []string, widths []int, separator string) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(tui.MutedColor).
		Bold(true)

	var parts []string
	for i, col := range columns {
		width := 0
		if i < len(widths) {
			width = widths[i]
		}
		if width > 0 {
			parts = append(parts, fmt.Sprintf("%-*s", width, col))
		} else {
			parts = append(parts, col)
		}
	}

	return headerStyle.Render("  " + strings.Join(parts, separator))
}

// RenderSeparator renders a horizontal separator line.
func RenderSeparator(width int) string {
	sepStyle := lipgloss.NewStyle().Foreground(tui.BorderColor)
	return sepStyle.Render(strings.Repeat("─", width))
}

// RenderEmptyState renders a message when there are no items in a list.
func RenderEmptyState(message string) string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(tui.MutedColor).
		Italic(true)
	return emptyStyle.Render("  " + message)
}

// RenderTitle renders a view title.
func RenderTitle(title string) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2)
	return titleStyle.Render(title)
}

// RenderErrorTitle renders an error/warning view title.
func RenderErrorTitle(title string) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(tui.ErrorColor).
		Bold(true).
		Padding(0, 2)
	return titleStyle.Render(title)
}

// RenderStatusMessage renders a status message.
func RenderStatusMessage(msg string, isError bool) string {
	if msg == "" {
		return ""
	}
	var style lipgloss.Style
	if isError {
		style = lipgloss.NewStyle().Foreground(tui.ErrorColor).Italic(true)
	} else {
		style = lipgloss.NewStyle().Foreground(tui.AccentColor).Italic(true)
	}
	return style.Render(msg)
}

// RenderBox wraps content in a rounded border box.
func RenderBox(content string, borderColor lipgloss.Color) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2)
	return boxStyle.Render(content)
}

// RenderBoxDefault wraps content in a rounded border box with default border color.
func RenderBoxDefault(content string) string {
	return RenderBox(content, tui.BorderColor)
}

// RenderBoxError wraps content in a rounded border box with error border color.
func RenderBoxError(content string) string {
	return RenderBox(content, tui.ErrorColor)
}
