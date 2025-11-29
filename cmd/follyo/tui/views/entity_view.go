package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui/components"
)

// EntityViewMode represents the current mode of an entity CRUD view
type EntityViewMode int

const (
	EntityModeList EntityViewMode = iota
	EntityModeAdd
	EntityModeConfirmDelete
)

// EntityAddedMsg is a generic message for when an entity is added
type EntityAddedMsg struct {
	EntityName string
	EntityID   string
	Error      error
}

// EntityDeletedMsg is a generic message for when an entity is deleted
type EntityDeletedMsg struct {
	EntityID string
	Error    error
}

// FormFieldConfig defines configuration for a form field
type FormFieldConfig struct {
	Label        string
	Placeholder  string
	CharLimit    int
	Width        int
	DefaultValue string
}

// EntityViewConfig defines the configuration for an entity CRUD view
type EntityViewConfig struct {
	// Display settings
	Title          string // e.g., "PURCHASES", "SALES"
	EntityName     string // e.g., "purchase", "sale" (for messages)
	EmptyMessage   string // e.g., "No purchases yet. Press 'a' to add one."
	ColumnHeader   string // Pre-formatted column header string
	SeparatorWidth int

	// Form configuration
	FormTitle  string            // e.g., "ADD PURCHASE"
	FormFields []FormFieldConfig // Form field definitions
	FormLabels []string          // Labels for form fields

	// Callbacks for rendering
	RenderRow func(index, cursor int, item interface{}) string

	// Callbacks for delete confirmation
	RenderDeleteInfo func(item interface{}) string
}

// EntityViewState holds the common state for entity views
type EntityViewState struct {
	Cursor     int
	Mode       EntityViewMode
	Inputs     []textinput.Model
	FocusIndex int
	Keys       tui.KeyMap
	Width      int
	Height     int
	Err        error
	StatusMsg  string
}

// NewEntityViewState creates a new entity view state with form inputs
func NewEntityViewState(fields []FormFieldConfig) EntityViewState {
	formFields := make([]components.FormField, len(fields))
	for i, f := range fields {
		formFields[i] = components.FormField{
			Placeholder:  f.Placeholder,
			CharLimit:    f.CharLimit,
			Width:        f.Width,
			DefaultValue: f.DefaultValue,
		}
	}

	return EntityViewState{
		Inputs: components.BuildFormInputs(formFields),
		Keys:   tui.DefaultKeyMap(),
		Mode:   EntityModeList,
	}
}

// HandleWindowSize updates the view dimensions
func (s *EntityViewState) HandleWindowSize(width, height int) {
	s.Width = width
	s.Height = height
}

// HandleListNavigation handles up/down navigation in list mode
func (s *EntityViewState) HandleListNavigation(msg tea.KeyMsg, itemCount int) bool {
	switch {
	case key.Matches(msg, s.Keys.Up):
		s.Cursor = components.MoveCursorUp(s.Cursor, itemCount)
		return true
	case key.Matches(msg, s.Keys.Down):
		s.Cursor = components.MoveCursorDown(s.Cursor, itemCount)
		return true
	}
	return false
}

// EnterAddMode switches to add mode
func (s *EntityViewState) EnterAddMode(defaults []string) tea.Cmd {
	s.Mode = EntityModeAdd
	s.FocusIndex = 0
	components.ResetFormInputs(s.Inputs, defaults)
	s.StatusMsg = ""
	return components.FocusField(s.Inputs, 0)
}

// EnterDeleteMode switches to delete confirmation mode
func (s *EntityViewState) EnterDeleteMode() {
	s.Mode = EntityModeConfirmDelete
	s.StatusMsg = ""
}

// ExitToList exits to list mode
func (s *EntityViewState) ExitToList() {
	s.Mode = EntityModeList
	components.BlurAll(s.Inputs)
	s.StatusMsg = ""
}

// HandleFormNavigation handles navigation within the add form
func (s *EntityViewState) HandleFormNavigation(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		s.ExitToList()
		return true, nil

	case tea.KeyTab, tea.KeyShiftTab, tea.KeyDown, tea.KeyUp:
		var cmd tea.Cmd
		if msg.Type == tea.KeyUp || msg.Type == tea.KeyShiftTab {
			s.FocusIndex, cmd = components.PrevField(s.Inputs, s.FocusIndex)
		} else {
			s.FocusIndex, cmd = components.NextField(s.Inputs, s.FocusIndex)
		}
		return true, cmd
	}
	return false, nil
}

// UpdateFocusedInput updates the currently focused input field
func (s *EntityViewState) UpdateFocusedInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	s.Inputs[s.FocusIndex], cmd = s.Inputs[s.FocusIndex].Update(msg)
	return cmd
}

// MoveToNextField moves focus to the next field
func (s *EntityViewState) MoveToNextField() tea.Cmd {
	var cmd tea.Cmd
	s.FocusIndex, cmd = components.NextField(s.Inputs, s.FocusIndex)
	return cmd
}

// IsLastField checks if currently on the last field
func (s *EntityViewState) IsLastField() bool {
	return s.FocusIndex == len(s.Inputs)-1
}

// GetFieldValue returns the trimmed value of a field
func (s *EntityViewState) GetFieldValue(index int) string {
	if index < 0 || index >= len(s.Inputs) {
		return ""
	}
	return strings.TrimSpace(s.Inputs[index].Value())
}

// SetStatusMsg sets the status message
func (s *EntityViewState) SetStatusMsg(msg string) {
	s.StatusMsg = msg
}

// AdjustCursorAfterDelete adjusts cursor position after item deletion
func (s *EntityViewState) AdjustCursorAfterDelete(newItemCount int) {
	if s.Cursor >= newItemCount && s.Cursor > 0 {
		s.Cursor--
	}
}

// RenderListView renders the list view
func RenderListView(cfg EntityViewConfig, state *EntityViewState, items []interface{}) string {
	var b strings.Builder

	b.WriteString(components.RenderTitle(cfg.Title))
	b.WriteString("\n\n")

	if len(items) == 0 {
		b.WriteString(components.RenderEmptyState(cfg.EmptyMessage))
		b.WriteString("\n")
	} else {
		// Column headers
		headerStyle := lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Bold(true)
		b.WriteString(headerStyle.Render(cfg.ColumnHeader))
		b.WriteString("\n")

		// Separator
		b.WriteString(components.RenderSeparator(cfg.SeparatorWidth))
		b.WriteString("\n")

		// List items
		for i, item := range items {
			b.WriteString(cfg.RenderRow(i, state.Cursor, item))
		}
	}

	// Status message
	if state.StatusMsg != "" {
		b.WriteString("\n")
		b.WriteString(components.RenderStatusMessage(state.StatusMsg, false))
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(components.RenderHelp(components.ListHelp(len(items) > 0)))

	return components.RenderBoxDefault(b.String())
}

// RenderAddForm renders the add form view
func RenderAddForm(cfg EntityViewConfig, state *EntityViewState) string {
	var b strings.Builder

	b.WriteString(components.RenderTitle(cfg.FormTitle))
	b.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().
		Foreground(tui.SubtleTextColor).
		Width(12)

	focusedLabelStyle := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Width(12)

	for i, label := range cfg.FormLabels {
		ls := labelStyle
		if state.FocusIndex == i {
			ls = focusedLabelStyle
		}
		b.WriteString(ls.Render(label))
		b.WriteString(state.Inputs[i].View())
		b.WriteString("\n")
	}

	// Status message (for errors)
	if state.StatusMsg != "" {
		b.WriteString("\n")
		b.WriteString(components.RenderStatusMessage(state.StatusMsg, true))
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(components.RenderHelp(components.FormHelp()))

	return components.RenderBoxDefault(b.String())
}

// RenderDeleteConfirm renders the delete confirmation dialog
func RenderDeleteConfirm(cfg EntityViewConfig, item interface{}) string {
	var b strings.Builder

	b.WriteString(components.RenderErrorTitle("CONFIRM DELETE"))
	b.WriteString("\n\n")

	if cfg.RenderDeleteInfo != nil {
		infoStyle := lipgloss.NewStyle().Foreground(tui.TextColor)
		b.WriteString(infoStyle.Render(cfg.RenderDeleteInfo(item)))
		b.WriteString("\n\n")
	}

	b.WriteString(components.RenderHelp(components.DeleteConfirmHelp()))

	return components.RenderBoxError(b.String())
}
