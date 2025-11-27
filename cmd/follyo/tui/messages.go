// Package tui provides an interactive terminal user interface for Follyo.
package tui

// BackToMenuMsg signals returning to the main menu.
// This is defined here so both tui and views packages can use it.
type BackToMenuMsg struct{}

// MenuSelectMsg is sent when a menu item is selected.
type MenuSelectMsg struct {
	Action string
}
