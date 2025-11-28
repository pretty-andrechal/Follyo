package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
)

// HelpItem represents a single help key-action pair.
type HelpItem struct {
	Key    string
	Action string
}

// RenderHelp renders a help line with key bindings.
func RenderHelp(items []HelpItem) string {
	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)

	var parts []string
	for _, item := range items {
		parts = append(parts, tui.HelpKeyStyle.Render(item.Key)+" "+item.Action)
	}

	return helpStyle.Render(strings.Join(parts, "  "))
}

// Common help items that can be reused across views.
var (
	HelpNavigate = HelpItem{Key: "↑↓", Action: "navigate"}
	HelpAdd      = HelpItem{Key: "a", Action: "add"}
	HelpDelete   = HelpItem{Key: "d", Action: "delete"}
	HelpBack     = HelpItem{Key: "esc", Action: "back"}
	HelpQuit     = HelpItem{Key: "q", Action: "quit"}
	HelpSave     = HelpItem{Key: "enter", Action: "save"}
	HelpCancel   = HelpItem{Key: "esc", Action: "cancel"}
	HelpConfirm  = HelpItem{Key: "y", Action: "confirm"}
	HelpDeny     = HelpItem{Key: "n/esc", Action: "cancel"}
	HelpTab      = HelpItem{Key: "tab/↑↓", Action: "navigate"}
	HelpRefresh  = HelpItem{Key: "r", Action: "refresh"}
	HelpScroll   = HelpItem{Key: "↑↓", Action: "scroll"}
	HelpSearch   = HelpItem{Key: "enter", Action: "search"}
)

// ListHelp returns common help items for list views.
func ListHelp(hasItems bool) []HelpItem {
	if hasItems {
		return []HelpItem{HelpNavigate, HelpAdd, HelpDelete, HelpBack, HelpQuit}
	}
	return []HelpItem{HelpAdd, HelpBack, HelpQuit}
}

// FormHelp returns common help items for form views.
func FormHelp() []HelpItem {
	return []HelpItem{HelpTab, HelpSave, HelpCancel}
}

// DeleteConfirmHelp returns help items for delete confirmation.
func DeleteConfirmHelp() []HelpItem {
	return []HelpItem{HelpConfirm, HelpDeny}
}
