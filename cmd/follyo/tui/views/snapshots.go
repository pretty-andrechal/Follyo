package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui/format"
	"github.com/pretty-andrechal/follyo/internal/models"
	"github.com/pretty-andrechal/follyo/internal/portfolio"
	"github.com/pretty-andrechal/follyo/internal/prices"
	"github.com/pretty-andrechal/follyo/internal/storage"
)

// SnapshotsViewMode represents the current mode of the snapshots view
type SnapshotsViewMode int

const (
	SnapshotsList SnapshotsViewMode = iota
	SnapshotsDetail
	SnapshotsSaving
	SnapshotsNoteInput
)

// SnapshotsModel represents the snapshots view
type SnapshotsModel struct {
	store          *storage.SnapshotStore
	portfolio      portfolio.SnapshotsManager
	tickerMappings map[string]string
	snapshots      []models.Snapshot
	cursor         int
	mode           SnapshotsViewMode
	selectedID     string
	spinner        spinner.Model
	textInput      textinput.Model
	viewport       viewport.Model
	viewportReady  bool
	keys           tui.KeyMap
	width          int
	height         int
	err            error
	statusMsg      string
}

// NewSnapshotsModel creates a new snapshots view model
func NewSnapshotsModel(store *storage.SnapshotStore, p portfolio.SnapshotsManager, tickerMappings map[string]string) SnapshotsModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(tui.PrimaryColor)

	ti := textinput.New()
	ti.Placeholder = "Optional note for this snapshot..."
	ti.CharLimit = tui.InputNotesCharLimit

	m := SnapshotsModel{
		store:          store,
		portfolio:      p,
		tickerMappings: tickerMappings,
		spinner:        s,
		textInput:      ti,
		keys:           tui.DefaultKeyMap(),
		mode:           SnapshotsList,
	}

	m.loadSnapshots()
	return m
}

func (m *SnapshotsModel) loadSnapshots() {
	m.snapshots = m.store.List()
}

// Init initializes the snapshots model
func (m SnapshotsModel) Init() tea.Cmd {
	return nil
}

// SnapshotSavedMsg is sent when a snapshot is saved
type SnapshotSavedMsg struct {
	Snapshot *models.Snapshot
	Error    error
}

// SnapshotDeletedMsg is sent when a snapshot is deleted
type SnapshotDeletedMsg struct {
	ID    string
	Error error
}

// Update handles messages for the snapshots model
func (m SnapshotsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle based on mode
		switch m.mode {
		case SnapshotsNoteInput:
			return m.handleNoteInputKeys(msg)
		case SnapshotsDetail:
			return m.handleDetailKeys(msg)
		case SnapshotsSaving:
			// Ignore keys while saving
			return m, nil
		default:
			return m.handleListKeys(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 3
		footerHeight := 3
		verticalMargins := headerHeight + footerHeight

		if !m.viewportReady {
			m.viewport = viewport.New(msg.Width-4, msg.Height-verticalMargins)
			m.viewport.YPosition = headerHeight
			m.viewportReady = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - verticalMargins
		}

	case spinner.TickMsg:
		if m.mode == SnapshotsSaving {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case SnapshotSavedMsg:
		m.mode = SnapshotsList
		if msg.Error != nil {
			m.err = msg.Error
			m.statusMsg = fmt.Sprintf("Error saving snapshot: %v", msg.Error)
		} else {
			m.statusMsg = "Snapshot saved!"
			m.loadSnapshots()
			m.cursor = 0 // Move to newest
		}

	case SnapshotDeletedMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.statusMsg = fmt.Sprintf("Error deleting snapshot: %v", msg.Error)
		} else {
			m.statusMsg = "Snapshot deleted"
			m.loadSnapshots()
			if m.cursor >= len(m.snapshots) && m.cursor > 0 {
				m.cursor--
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m SnapshotsModel) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		return m, func() tea.Msg { return tui.BackToMenuMsg{} }

	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.snapshots)-1 {
			m.cursor++
		}

	case key.Matches(msg, m.keys.Select):
		if len(m.snapshots) > 0 {
			m.selectedID = m.snapshots[m.cursor].ID
			m.mode = SnapshotsDetail
			m.updateDetailViewport()
		}

	case msg.String() == "n" || msg.String() == "s":
		// New snapshot - go to note input
		m.mode = SnapshotsNoteInput
		m.textInput.SetValue("")
		m.textInput.Focus()
		m.statusMsg = "Enter a note (optional), then press Enter to save"
		return m, textinput.Blink

	case msg.String() == "d" || msg.String() == "x":
		// Delete snapshot
		if len(m.snapshots) > 0 {
			return m, m.deleteSnapshot(m.snapshots[m.cursor].ID)
		}
	}

	return m, nil
}

func (m SnapshotsModel) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.mode = SnapshotsList
		m.selectedID = ""
		return m, nil

	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case msg.String() == "d" || msg.String() == "x":
		// Delete this snapshot
		id := m.selectedID
		m.mode = SnapshotsList
		m.selectedID = ""
		return m, m.deleteSnapshot(id)

	default:
		// Forward to viewport for scrolling
		if m.viewportReady {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m SnapshotsModel) handleNoteInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.mode = SnapshotsList
		m.textInput.Blur()
		m.statusMsg = ""
		return m, nil

	case tea.KeyEnter:
		// Start saving
		note := strings.TrimSpace(m.textInput.Value())
		m.textInput.Blur()
		m.mode = SnapshotsSaving
		m.statusMsg = "Fetching prices and saving snapshot..."
		return m, tea.Batch(m.spinner.Tick, m.saveSnapshot(note))

	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

func (m SnapshotsModel) saveSnapshot(note string) tea.Cmd {
	return func() tea.Msg {
		// Get portfolio summary to find all coins
		summary, err := m.portfolio.GetSummary()
		if err != nil {
			return SnapshotSavedMsg{Error: fmt.Errorf("getting summary: %w", err)}
		}

		// Collect all coins
		allCoins := make(map[string]bool)
		for coin := range summary.HoldingsByCoin {
			allCoins[coin] = true
		}
		for coin := range summary.LoansByCoin {
			allCoins[coin] = true
		}

		coins := make([]string, 0, len(allCoins))
		for coin := range allCoins {
			coins = append(coins, coin)
		}

		if len(coins) == 0 {
			return SnapshotSavedMsg{Error: fmt.Errorf("no holdings to snapshot")}
		}

		// Fetch prices
		ps := prices.New()
		for ticker, geckoID := range m.tickerMappings {
			ps.AddCoinMapping(ticker, geckoID)
		}

		priceMap, err := ps.GetPrices(coins)
		if err != nil {
			return SnapshotSavedMsg{Error: fmt.Errorf("fetching prices: %w", err)}
		}

		// Create snapshot
		snapshot, err := m.portfolio.CreateSnapshot(priceMap, note)
		if err != nil {
			return SnapshotSavedMsg{Error: err}
		}

		// Save to store
		if err := m.store.Add(snapshot); err != nil {
			return SnapshotSavedMsg{Error: fmt.Errorf("saving snapshot: %w", err)}
		}

		return SnapshotSavedMsg{Snapshot: &snapshot}
	}
}

func (m SnapshotsModel) deleteSnapshot(id string) tea.Cmd {
	return func() tea.Msg {
		removed, err := m.store.Remove(id)
		if err != nil {
			return SnapshotDeletedMsg{ID: id, Error: err}
		}
		if !removed {
			return SnapshotDeletedMsg{ID: id, Error: fmt.Errorf("snapshot not found")}
		}
		return SnapshotDeletedMsg{ID: id}
	}
}

func (m *SnapshotsModel) updateDetailViewport() {
	if !m.viewportReady {
		return
	}

	snapshot, ok := m.store.Get(m.selectedID)
	if !ok {
		return
	}

	m.viewport.SetContent(m.renderSnapshotDetail(snapshot))
	m.viewport.GotoTop()
}

// View renders the snapshots view
func (m SnapshotsModel) View() string {
	switch m.mode {
	case SnapshotsSaving:
		return m.renderSaving()
	case SnapshotsNoteInput:
		return m.renderNoteInput()
	case SnapshotsDetail:
		return m.renderDetail()
	default:
		return m.renderList()
	}
}

func (m SnapshotsModel) renderSaving() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.PrimaryColor).
		Padding(2, 4)

	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Render("SAVING SNAPSHOT")

	loadingText := lipgloss.NewStyle().
		Foreground(tui.SubtleTextColor).
		Render(fmt.Sprintf("\n\n%s %s", m.spinner.View(), m.statusMsg))

	return boxStyle.Render(title + loadingText)
}

func (m SnapshotsModel) renderNoteInput() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Render("NEW SNAPSHOT")

	b.WriteString(title)
	b.WriteString("\n\n")

	// Note input
	labelStyle := lipgloss.NewStyle().Foreground(tui.SubtleTextColor)
	b.WriteString(labelStyle.Render("Note (optional):"))
	b.WriteString("\n")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	// Help
	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)
	help := fmt.Sprintf("%s save  %s cancel",
		tui.HelpKeyStyle.Render("enter"),
		tui.HelpKeyStyle.Render("esc"))
	b.WriteString(helpStyle.Render(help))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.BorderColor).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

func (m SnapshotsModel) renderList() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Render("SNAPSHOTS")

	b.WriteString(title)
	b.WriteString("\n\n")

	if len(m.snapshots) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Italic(true)
		b.WriteString(emptyStyle.Render("  No snapshots yet. Press 'n' to create one."))
		b.WriteString("\n")
	} else {
		// Column headers
		headerStyle := lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Bold(true)
		header := fmt.Sprintf("  %-12s  %-19s  %14s  %-22s  %s",
			"ID", "Date", "Net Value", "P/L", "Note")
		b.WriteString(headerStyle.Render(header))
		b.WriteString("\n")

		// Separator
		sepStyle := lipgloss.NewStyle().Foreground(tui.BorderColor)
		b.WriteString(sepStyle.Render(strings.Repeat("─", tui.SeparatorWidthSnapshots)))
		b.WriteString("\n")

		// List items
		for i, snap := range m.snapshots {
			b.WriteString(m.renderSnapshotRow(i, snap))
		}
	}

	// Status message
	if m.statusMsg != "" {
		b.WriteString("\n")
		statusStyle := lipgloss.NewStyle().
			Foreground(tui.AccentColor).
			Italic(true)
		b.WriteString(statusStyle.Render(m.statusMsg))
	}

	// Help
	b.WriteString("\n\n")
	help := m.renderListHelp()
	b.WriteString(help)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.BorderColor).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

func (m SnapshotsModel) renderSnapshotRow(index int, snap models.Snapshot) string {
	isSelected := index == m.cursor

	// Cursor
	cursor := "  "
	if isSelected {
		cursor = lipgloss.NewStyle().
			Foreground(tui.PrimaryColor).
			Render("> ")
	}

	// Format values
	id := snap.ID
	date := snap.Timestamp.Format("2006-01-02 15:04")
	netValue := format.USDSimple(snap.NetValue)

	// Profit/Loss with color
	plText := format.ProfitLoss(snap.ProfitLoss, snap.ProfitPercent)

	plStyle := lipgloss.NewStyle()
	if snap.ProfitLoss > 0 {
		plStyle = plStyle.Foreground(tui.SuccessColor)
	} else if snap.ProfitLoss < 0 {
		plStyle = plStyle.Foreground(tui.ErrorColor)
	} else {
		plStyle = plStyle.Foreground(tui.TextColor)
	}

	// Truncate note
	note := format.TruncateNote(snap.Note)

	// Build row
	rowStyle := lipgloss.NewStyle().Foreground(tui.TextColor)
	if isSelected {
		rowStyle = rowStyle.Bold(true).Foreground(tui.PrimaryColor)
	}

	row := fmt.Sprintf("%-12s  %-19s  %14s  ", id, date, netValue)
	plPart := fmt.Sprintf("%-22s", plText)
	notePart := lipgloss.NewStyle().Foreground(tui.SubtleTextColor).Render(note)

	return cursor + rowStyle.Render(row) + plStyle.Render(plPart) + "  " + notePart + "\n"
}

func (m SnapshotsModel) renderListHelp() string {
	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)

	var help string
	if len(m.snapshots) > 0 {
		help = fmt.Sprintf("%s navigate  %s view  %s new  %s delete  %s back  %s quit",
			tui.HelpKeyStyle.Render("↑↓"),
			tui.HelpKeyStyle.Render("enter"),
			tui.HelpKeyStyle.Render("n"),
			tui.HelpKeyStyle.Render("d"),
			tui.HelpKeyStyle.Render("esc"),
			tui.HelpKeyStyle.Render("q"))
	} else {
		help = fmt.Sprintf("%s new snapshot  %s back  %s quit",
			tui.HelpKeyStyle.Render("n"),
			tui.HelpKeyStyle.Render("esc"),
			tui.HelpKeyStyle.Render("q"))
	}

	return helpStyle.Render(help)
}

func (m SnapshotsModel) renderDetail() string {
	snapshot, ok := m.store.Get(m.selectedID)
	if !ok {
		return "Snapshot not found"
	}

	// Header
	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Render(fmt.Sprintf("SNAPSHOT %s", snapshot.ID))

	// Scroll indicator
	scrollInfo := ""
	if m.viewportReady {
		scrollPercent := m.viewport.ScrollPercent() * 100
		if m.viewport.TotalLineCount() > m.viewport.Height {
			scrollInfo = lipgloss.NewStyle().
				Foreground(tui.MutedColor).
				Render(fmt.Sprintf(" (%.0f%%)", scrollPercent))
		}
	}

	header := lipgloss.JoinHorizontal(lipgloss.Center, title, scrollInfo)

	// Footer with help
	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)
	help := fmt.Sprintf("%s scroll  %s delete  %s back  %s quit",
		tui.HelpKeyStyle.Render("↑↓"),
		tui.HelpKeyStyle.Render("d"),
		tui.HelpKeyStyle.Render("esc"),
		tui.HelpKeyStyle.Render("q"))
	footer := helpStyle.Render(help)

	// Viewport with border
	viewportStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.BorderColor).
		Padding(0, 1)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		viewportStyle.Render(m.viewport.View()),
		footer,
	)
}

func (m SnapshotsModel) renderSnapshotDetail(snap *models.Snapshot) string {
	var b strings.Builder

	// Timestamp
	dateStyle := lipgloss.NewStyle().Foreground(tui.SubtleTextColor)
	b.WriteString(dateStyle.Render(fmt.Sprintf("Saved: %s", snap.Timestamp.Format("2006-01-02 15:04:05"))))
	b.WriteString("\n")

	// Note
	if snap.Note != "" {
		noteStyle := lipgloss.NewStyle().Foreground(tui.AccentColor).Italic(true)
		b.WriteString(noteStyle.Render(fmt.Sprintf("Note: %s", snap.Note)))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Summary section
	sectionStyle := lipgloss.NewStyle().
		Foreground(tui.AccentColor).
		Bold(true)

	b.WriteString(sectionStyle.Render("PORTFOLIO VALUE"))
	b.WriteString("\n")

	labelStyle := lipgloss.NewStyle().
		Foreground(tui.SubtleTextColor).
		Width(18)
	valueStyle := lipgloss.NewStyle().
		Foreground(tui.TextColor).
		Width(14).
		Align(lipgloss.Right)

	summaryItems := []struct {
		label string
		value string
		style lipgloss.Style
	}{
		{"Holdings Value:", format.USDSimple(snap.HoldingsValue), valueStyle.Foreground(tui.SuccessColor)},
		{"Loans Value:", "-" + format.USDSimple(snap.LoansValue), valueStyle.Foreground(tui.ErrorColor)},
		{"Net Value:", format.USDSimple(snap.NetValue), valueStyle.Bold(true)},
		{"Total Invested:", format.USDSimple(snap.TotalInvested), valueStyle},
		{"Total Sold:", format.USDSimple(snap.TotalSold), valueStyle},
	}

	for _, item := range summaryItems {
		b.WriteString(labelStyle.Render(item.label))
		b.WriteString(item.style.Render(item.value))
		b.WriteString("\n")
	}

	// Profit/Loss
	plText := format.ProfitLoss(snap.ProfitLoss, snap.ProfitPercent)
	plStyle := valueStyle
	if snap.ProfitLoss > 0 {
		plStyle = plStyle.Foreground(tui.SuccessColor)
	} else if snap.ProfitLoss < 0 {
		plStyle = plStyle.Foreground(tui.ErrorColor)
	}
	b.WriteString(labelStyle.Render("Profit/Loss:"))
	b.WriteString(plStyle.Render(plText))
	b.WriteString("\n\n")

	// Coin values section
	if len(snap.CoinValues) > 0 {
		b.WriteString(sectionStyle.Render("COIN VALUES"))
		b.WriteString("\n")

		// Header
		headerStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)
		coinHeader := fmt.Sprintf("  %-8s  %14s  %12s  %14s",
			"Coin", "Amount", "Price", "Value")
		b.WriteString(headerStyle.Render(coinHeader))
		b.WriteString("\n")

		// Separator
		sepStyle := lipgloss.NewStyle().Foreground(tui.BorderColor)
		b.WriteString(sepStyle.Render(strings.Repeat("─", tui.SeparatorWidthSnapshotDetail)))
		b.WriteString("\n")

		// Sort coins alphabetically
		coins := make([]string, 0, len(snap.CoinValues))
		for coin := range snap.CoinValues {
			coins = append(coins, coin)
		}
		sort.Strings(coins)

		for _, coin := range coins {
			cv := snap.CoinValues[coin]
			row := fmt.Sprintf("  %-8s  %14.6f  %12s  %14s",
				coin,
				cv.Amount,
				format.USDSimple(cv.Price),
				format.USDSimple(cv.Value))
			b.WriteString(lipgloss.NewStyle().Foreground(tui.TextColor).Render(row))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// GetStore returns the snapshot store
func (m SnapshotsModel) GetStore() *storage.SnapshotStore {
	return m.store
}
