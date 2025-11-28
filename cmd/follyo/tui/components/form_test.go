package components

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
)

func TestBuildFormInputs(t *testing.T) {
	fields := []FormField{
		{Placeholder: "Name", CharLimit: 50, Width: 30},
		{Placeholder: "Email", CharLimit: 100, Width: 40, DefaultValue: "test@example.com"},
		{Placeholder: "Notes", CharLimit: 200, Width: 50},
	}

	inputs := BuildFormInputs(fields)

	if len(inputs) != 3 {
		t.Errorf("expected 3 inputs, got %d", len(inputs))
	}

	// Check first input
	if inputs[0].Placeholder != "Name" {
		t.Errorf("expected placeholder 'Name', got '%s'", inputs[0].Placeholder)
	}
	if inputs[0].CharLimit != 50 {
		t.Errorf("expected char limit 50, got %d", inputs[0].CharLimit)
	}

	// Check second input has default value
	if inputs[1].Value() != "test@example.com" {
		t.Errorf("expected default value 'test@example.com', got '%s'", inputs[1].Value())
	}
}

func TestBuildFormInputs_Empty(t *testing.T) {
	inputs := BuildFormInputs([]FormField{})
	if len(inputs) != 0 {
		t.Errorf("expected 0 inputs, got %d", len(inputs))
	}
}

func TestResetFormInputs(t *testing.T) {
	inputs := make([]textinput.Model, 3)
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].SetValue("some value")
	}

	defaults := []string{"default1", "", "default3"}
	ResetFormInputs(inputs, defaults)

	if inputs[0].Value() != "default1" {
		t.Errorf("expected 'default1', got '%s'", inputs[0].Value())
	}
	if inputs[1].Value() != "" {
		t.Errorf("expected empty string, got '%s'", inputs[1].Value())
	}
	if inputs[2].Value() != "default3" {
		t.Errorf("expected 'default3', got '%s'", inputs[2].Value())
	}
}

func TestResetFormInputs_NilDefaults(t *testing.T) {
	inputs := make([]textinput.Model, 2)
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].SetValue("value")
	}

	ResetFormInputs(inputs, nil)

	for i, input := range inputs {
		if input.Value() != "" {
			t.Errorf("input %d: expected empty string, got '%s'", i, input.Value())
		}
	}
}

func TestResetFormInputs_ShorterDefaults(t *testing.T) {
	inputs := make([]textinput.Model, 3)
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].SetValue("value")
	}

	defaults := []string{"only one"}
	ResetFormInputs(inputs, defaults)

	if inputs[0].Value() != "only one" {
		t.Errorf("expected 'only one', got '%s'", inputs[0].Value())
	}
	if inputs[1].Value() != "" {
		t.Errorf("expected empty string, got '%s'", inputs[1].Value())
	}
	if inputs[2].Value() != "" {
		t.Errorf("expected empty string, got '%s'", inputs[2].Value())
	}
}

func TestFocusField(t *testing.T) {
	inputs := make([]textinput.Model, 3)
	for i := range inputs {
		inputs[i] = textinput.New()
	}

	// Focus second field
	cmd := FocusField(inputs, 1)

	if cmd == nil {
		t.Error("expected command, got nil")
	}

	// Check that only the second field is focused
	if inputs[0].Focused() {
		t.Error("input 0 should not be focused")
	}
	if !inputs[1].Focused() {
		t.Error("input 1 should be focused")
	}
	if inputs[2].Focused() {
		t.Error("input 2 should not be focused")
	}
}

func TestFocusField_OutOfBounds(t *testing.T) {
	inputs := make([]textinput.Model, 2)
	for i := range inputs {
		inputs[i] = textinput.New()
	}

	// Should not panic with out of bounds index
	FocusField(inputs, 5)
	FocusField(inputs, -1)

	// Both should be blurred
	for i, input := range inputs {
		if input.Focused() {
			t.Errorf("input %d should not be focused", i)
		}
	}
}

func TestBlurAll(t *testing.T) {
	inputs := make([]textinput.Model, 3)
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Focus()
	}

	BlurAll(inputs)

	for i, input := range inputs {
		if input.Focused() {
			t.Errorf("input %d should not be focused after BlurAll", i)
		}
	}
}

func TestNextField(t *testing.T) {
	inputs := make([]textinput.Model, 3)
	for i := range inputs {
		inputs[i] = textinput.New()
	}

	// From 0 to 1
	newIndex, cmd := NextField(inputs, 0)
	if newIndex != 1 {
		t.Errorf("expected index 1, got %d", newIndex)
	}
	if cmd == nil {
		t.Error("expected command, got nil")
	}

	// From 2 to 0 (wrap)
	newIndex, _ = NextField(inputs, 2)
	if newIndex != 0 {
		t.Errorf("expected index 0 (wrap), got %d", newIndex)
	}
}

func TestPrevField(t *testing.T) {
	inputs := make([]textinput.Model, 3)
	for i := range inputs {
		inputs[i] = textinput.New()
	}

	// From 1 to 0
	newIndex, cmd := PrevField(inputs, 1)
	if newIndex != 0 {
		t.Errorf("expected index 0, got %d", newIndex)
	}
	if cmd == nil {
		t.Error("expected command, got nil")
	}

	// From 0 to 2 (wrap)
	newIndex, _ = PrevField(inputs, 0)
	if newIndex != 2 {
		t.Errorf("expected index 2 (wrap), got %d", newIndex)
	}
}
