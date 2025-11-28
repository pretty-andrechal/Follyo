// Package components provides reusable TUI components for Follyo views.
package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// FormField defines a form field configuration.
type FormField struct {
	Placeholder  string
	CharLimit    int
	Width        int
	DefaultValue string
}

// BuildFormInputs creates textinput models from field specifications.
func BuildFormInputs(fields []FormField) []textinput.Model {
	inputs := make([]textinput.Model, len(fields))
	for i, f := range fields {
		inputs[i] = textinput.New()
		inputs[i].Placeholder = f.Placeholder
		inputs[i].CharLimit = f.CharLimit
		inputs[i].Width = f.Width
		if f.DefaultValue != "" {
			inputs[i].SetValue(f.DefaultValue)
		}
	}
	return inputs
}

// ResetFormInputs clears all form inputs, optionally setting default values.
// If defaults is nil or shorter than inputs, remaining fields are cleared to empty.
func ResetFormInputs(inputs []textinput.Model, defaults []string) {
	for i := range inputs {
		if defaults != nil && i < len(defaults) {
			inputs[i].SetValue(defaults[i])
		} else {
			inputs[i].SetValue("")
		}
	}
}

// FocusField sets focus to a specific field, blurring all others.
// Returns a command to start the cursor blink.
func FocusField(inputs []textinput.Model, index int) tea.Cmd {
	for i := range inputs {
		inputs[i].Blur()
	}
	if index >= 0 && index < len(inputs) {
		inputs[index].Focus()
	}
	return textinput.Blink
}

// BlurAll removes focus from all inputs.
func BlurAll(inputs []textinput.Model) {
	for i := range inputs {
		inputs[i].Blur()
	}
}

// NextField moves focus to the next field, wrapping around.
// Returns the new focus index and a command to start the cursor blink.
func NextField(inputs []textinput.Model, currentIndex int) (int, tea.Cmd) {
	newIndex := (currentIndex + 1) % len(inputs)
	return newIndex, FocusField(inputs, newIndex)
}

// PrevField moves focus to the previous field, wrapping around.
// Returns the new focus index and a command to start the cursor blink.
func PrevField(inputs []textinput.Model, currentIndex int) (int, tea.Cmd) {
	newIndex := currentIndex - 1
	if newIndex < 0 {
		newIndex = len(inputs) - 1
	}
	return newIndex, FocusField(inputs, newIndex)
}

// UpdateFocusedInput updates only the currently focused input.
// Returns the updated input and any command.
func UpdateFocusedInput(inputs []textinput.Model, focusIndex int, msg tea.Msg) (textinput.Model, tea.Cmd) {
	if focusIndex >= 0 && focusIndex < len(inputs) {
		return inputs[focusIndex].Update(msg)
	}
	return textinput.Model{}, nil
}
