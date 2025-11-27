// Package tui provides an interactive terminal user interface for Follyo.
package tui

import "github.com/charmbracelet/lipgloss"

// Dark theme color palette
var (
	// Primary colors
	PrimaryColor   = lipgloss.Color("#7C3AED") // Purple
	SecondaryColor = lipgloss.Color("#6366F1") // Indigo
	AccentColor    = lipgloss.Color("#8B5CF6") // Light purple

	// Status colors
	SuccessColor = lipgloss.Color("#10B981") // Green
	ErrorColor   = lipgloss.Color("#EF4444") // Red
	WarningColor = lipgloss.Color("#F59E0B") // Yellow/Orange
	InfoColor    = lipgloss.Color("#3B82F6") // Blue

	// Neutral colors
	TextColor       = lipgloss.Color("#F9FAFB") // Almost white
	SubtleTextColor = lipgloss.Color("#9CA3AF") // Gray
	MutedColor      = lipgloss.Color("#6B7280") // Darker gray
	BorderColor     = lipgloss.Color("#374151") // Dark border
	BackgroundColor = lipgloss.Color("#111827") // Very dark blue-gray
	SurfaceColor    = lipgloss.Color("#1F2937") // Slightly lighter surface
)

// Common styles
var (
	// Base text styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			Bold(true).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(SubtleTextColor).
			Italic(true)

	// Panel/box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1, 2)

	FocusedBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1, 2)

	// Menu styles
	MenuTitleStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)

	MenuItemStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			PaddingLeft(2)

	SelectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(TextColor).
				Background(PrimaryColor).
				Bold(true).
				PaddingLeft(2).
				PaddingRight(2)

	// Status bar style
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(SubtleTextColor).
			Background(SurfaceColor).
			Padding(0, 1)

	// Help style
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true).
				Padding(0, 1)

	TableCellStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			Padding(0, 1)

	TableSelectedStyle = lipgloss.NewStyle().
				Foreground(TextColor).
				Background(SurfaceColor).
				Bold(true).
				Padding(0, 1)

	// Form styles
	FormLabelStyle = lipgloss.NewStyle().
			Foreground(SubtleTextColor).
			Width(12)

	FormInputStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			Border(lipgloss.NormalBorder()).
			BorderForeground(BorderColor).
			Padding(0, 1)

	FormFocusedInputStyle = lipgloss.NewStyle().
				Foreground(TextColor).
				Border(lipgloss.NormalBorder()).
				BorderForeground(PrimaryColor).
				Padding(0, 1)

	// Button styles
	ButtonStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			Background(SurfaceColor).
			Padding(0, 2).
			MarginRight(1)

	FocusedButtonStyle = lipgloss.NewStyle().
				Foreground(TextColor).
				Background(PrimaryColor).
				Padding(0, 2).
				MarginRight(1)

	// Value display styles
	PositiveValueStyle = lipgloss.NewStyle().
				Foreground(SuccessColor)

	NegativeValueStyle = lipgloss.NewStyle().
				Foreground(ErrorColor)

	NeutralValueStyle = lipgloss.NewStyle().
				Foreground(TextColor)

	// Notification styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningColor)

	// Logo/header style
	LogoStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Padding(0, 1)
)

// Helper function to style values based on sign
func StyleValue(value float64, formatted string) string {
	if value > 0 {
		return PositiveValueStyle.Render("+" + formatted)
	} else if value < 0 {
		return NegativeValueStyle.Render(formatted)
	}
	return NeutralValueStyle.Render(formatted)
}

// Helper to create a centered box with title
func TitledBox(title, content string, width int) string {
	titleRendered := MenuTitleStyle.Render(title)
	box := BoxStyle.Width(width).Render(content)
	return lipgloss.JoinVertical(lipgloss.Left, titleRendered, box)
}
