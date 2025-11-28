package components

import (
	"strings"
	"testing"
)

func TestRenderHelp(t *testing.T) {
	items := []HelpItem{
		{Key: "↑↓", Action: "navigate"},
		{Key: "a", Action: "add"},
		{Key: "q", Action: "quit"},
	}

	result := RenderHelp(items)

	if !strings.Contains(result, "navigate") {
		t.Error("help should contain 'navigate'")
	}
	if !strings.Contains(result, "add") {
		t.Error("help should contain 'add'")
	}
	if !strings.Contains(result, "quit") {
		t.Error("help should contain 'quit'")
	}
}

func TestRenderHelp_Empty(t *testing.T) {
	result := RenderHelp([]HelpItem{})
	// Should not panic and return something
	if result == "" {
		// This is acceptable - empty help renders empty string
	}
}

func TestListHelp_WithItems(t *testing.T) {
	items := ListHelp(true)

	if len(items) == 0 {
		t.Error("expected non-empty help items")
	}

	// Should contain navigate, add, delete, back, quit
	hasNavigate := false
	hasAdd := false
	hasDelete := false
	for _, item := range items {
		if item.Action == "navigate" {
			hasNavigate = true
		}
		if item.Action == "add" {
			hasAdd = true
		}
		if item.Action == "delete" {
			hasDelete = true
		}
	}

	if !hasNavigate {
		t.Error("list help with items should have navigate")
	}
	if !hasAdd {
		t.Error("list help with items should have add")
	}
	if !hasDelete {
		t.Error("list help with items should have delete")
	}
}

func TestListHelp_NoItems(t *testing.T) {
	items := ListHelp(false)

	if len(items) == 0 {
		t.Error("expected non-empty help items")
	}

	// Should not contain navigate or delete
	for _, item := range items {
		if item.Action == "navigate" {
			t.Error("list help without items should not have navigate")
		}
		if item.Action == "delete" {
			t.Error("list help without items should not have delete")
		}
	}

	// Should still have add
	hasAdd := false
	for _, item := range items {
		if item.Action == "add" {
			hasAdd = true
		}
	}
	if !hasAdd {
		t.Error("list help without items should still have add")
	}
}

func TestFormHelp(t *testing.T) {
	items := FormHelp()

	if len(items) == 0 {
		t.Error("expected non-empty help items")
	}

	// Should have tab, save, cancel
	hasSave := false
	hasCancel := false
	for _, item := range items {
		if item.Action == "save" {
			hasSave = true
		}
		if item.Action == "cancel" {
			hasCancel = true
		}
	}

	if !hasSave {
		t.Error("form help should have save")
	}
	if !hasCancel {
		t.Error("form help should have cancel")
	}
}

func TestDeleteConfirmHelp(t *testing.T) {
	items := DeleteConfirmHelp()

	if len(items) == 0 {
		t.Error("expected non-empty help items")
	}

	// Should have confirm and cancel
	hasConfirm := false
	hasCancel := false
	for _, item := range items {
		if item.Action == "confirm" {
			hasConfirm = true
		}
		if item.Action == "cancel" {
			hasCancel = true
		}
	}

	if !hasConfirm {
		t.Error("delete confirm help should have confirm")
	}
	if !hasCancel {
		t.Error("delete confirm help should have cancel")
	}
}
