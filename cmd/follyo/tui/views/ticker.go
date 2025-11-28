package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui"
	"github.com/pretty-andrechal/follyo/cmd/follyo/tui/format"
	"github.com/pretty-andrechal/follyo/internal/config"
	"github.com/pretty-andrechal/follyo/internal/prices"
)

// TickerViewMode represents the current mode of the ticker view
type TickerViewMode int

const (
	TickerList TickerViewMode = iota
	TickerAdd
	TickerSearch
	TickerSearchResults
	TickerConfirmDelete
	TickerShowDefaults
)

// Ticker form field indices
const (
	tickerFieldTicker = iota
	tickerFieldGeckoID
	tickerFieldCount
)

// MappingItem represents a ticker mapping for display
type MappingItem struct {
	Ticker   string
	GeckoID  string
	IsCustom bool
}

// TickerModel represents the ticker mapping view
type TickerModel struct {
	config          *config.ConfigStore
	priceService    *prices.PriceService
	mappings        []MappingItem
	defaultMappings []MappingItem
	cursor          int
	defaultCursor   int
	mode            TickerViewMode
	inputs          []textinput.Model
	searchInput     textinput.Model
	focusIndex      int
	searchResults   []prices.SearchResult
	searchCursor    int
	searchTicker    string // ticker to map search result to
	keys            tui.KeyMap
	width           int
	height          int
	err             error
	statusMsg       string
	searching       bool
}

// NewTickerModel creates a new ticker view model
func NewTickerModel(cfg *config.ConfigStore) TickerModel {
	inputs := make([]textinput.Model, tickerFieldCount)

	// Ticker input
	inputs[tickerFieldTicker] = textinput.New()
	inputs[tickerFieldTicker].Placeholder = "BTC, ETH, MUTE..."
	inputs[tickerFieldTicker].CharLimit = tui.InputCoinCharLimit
	inputs[tickerFieldTicker].Width = tui.InputCoinWidth

	// GeckoID input
	inputs[tickerFieldGeckoID] = textinput.New()
	inputs[tickerFieldGeckoID].Placeholder = "bitcoin, ethereum..."
	inputs[tickerFieldGeckoID].CharLimit = tui.InputGeckoIDCharLimit
	inputs[tickerFieldGeckoID].Width = tui.InputGeckoIDWidth

	// Search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search coin name or symbol..."
	searchInput.CharLimit = tui.InputSearchCharLimit
	searchInput.Width = tui.InputSearchWidth

	m := TickerModel{
		config:       cfg,
		priceService: prices.New(),
		inputs:       inputs,
		searchInput:  searchInput,
		keys:         tui.DefaultKeyMap(),
		mode:         TickerList,
	}

	m.loadMappings()
	return m
}

func (m *TickerModel) loadMappings() {
	// Load custom mappings
	customMap := m.config.GetAllTickerMappings()
	m.mappings = make([]MappingItem, 0, len(customMap))
	for ticker, geckoID := range customMap {
		m.mappings = append(m.mappings, MappingItem{
			Ticker:   ticker,
			GeckoID:  geckoID,
			IsCustom: true,
		})
	}
	// Sort by ticker
	sort.Slice(m.mappings, func(i, j int) bool {
		return m.mappings[i].Ticker < m.mappings[j].Ticker
	})

	// Load default mappings
	defaultMap := prices.GetDefaultMappings()
	m.defaultMappings = make([]MappingItem, 0, len(defaultMap))
	for ticker, geckoID := range defaultMap {
		m.defaultMappings = append(m.defaultMappings, MappingItem{
			Ticker:   ticker,
			GeckoID:  geckoID,
			IsCustom: false,
		})
	}
	sort.Slice(m.defaultMappings, func(i, j int) bool {
		return m.defaultMappings[i].Ticker < m.defaultMappings[j].Ticker
	})
}

// Init initializes the ticker model
func (m TickerModel) Init() tea.Cmd {
	return nil
}

// TickerMappingAddedMsg is sent when a mapping is added
type TickerMappingAddedMsg struct {
	Ticker  string
	GeckoID string
	Error   error
}

// TickerMappingDeletedMsg is sent when a mapping is deleted
type TickerMappingDeletedMsg struct {
	Ticker string
	Error  error
}

// TickerSearchResultsMsg is sent when search results are received
type TickerSearchResultsMsg struct {
	Results []prices.SearchResult
	Error   error
}

// Update handles messages for the ticker model
func (m TickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case TickerAdd:
			return m.handleAddKeys(msg)
		case TickerSearch:
			return m.handleSearchKeys(msg)
		case TickerSearchResults:
			return m.handleSearchResultsKeys(msg)
		case TickerConfirmDelete:
			return m.handleDeleteConfirmKeys(msg)
		case TickerShowDefaults:
			return m.handleDefaultsKeys(msg)
		default:
			return m.handleListKeys(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case TickerMappingAddedMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.statusMsg = fmt.Sprintf("Error: %v", msg.Error)
		} else {
			m.statusMsg = fmt.Sprintf("Mapped %s -> %s", msg.Ticker, msg.GeckoID)
			m.loadMappings()
		}
		m.mode = TickerList

	case TickerMappingDeletedMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.statusMsg = fmt.Sprintf("Error: %v", msg.Error)
		} else {
			// Check if default exists
			defaults := prices.GetDefaultMappings()
			if defaultID, ok := defaults[msg.Ticker]; ok {
				m.statusMsg = fmt.Sprintf("Removed %s (will use default: %s)", msg.Ticker, defaultID)
			} else {
				m.statusMsg = fmt.Sprintf("Removed mapping for %s", msg.Ticker)
			}
			m.loadMappings()
			if m.cursor >= len(m.mappings) && m.cursor > 0 {
				m.cursor--
			}
		}

	case TickerSearchResultsMsg:
		m.searching = false
		if msg.Error != nil {
			m.err = msg.Error
			m.statusMsg = fmt.Sprintf("Search error: %v", msg.Error)
			m.mode = TickerList
		} else if len(msg.Results) == 0 {
			m.statusMsg = "No results found"
			m.mode = TickerList
		} else {
			m.searchResults = msg.Results
			m.searchCursor = 0
			m.mode = TickerSearchResults
		}
	}

	return m, nil
}

func (m TickerModel) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		if m.cursor < len(m.mappings)-1 {
			m.cursor++
		}

	case msg.String() == "a" || msg.String() == "n":
		// Add new mapping manually
		m.mode = TickerAdd
		m.focusIndex = 0
		m.resetForm()
		m.inputs[tickerFieldTicker].Focus()
		m.statusMsg = ""
		return m, textinput.Blink

	case msg.String() == "s" || msg.String() == "/":
		// Search CoinGecko
		m.mode = TickerSearch
		m.searchInput.SetValue("")
		m.searchInput.Focus()
		m.searchTicker = ""
		m.statusMsg = ""
		return m, textinput.Blink

	case msg.String() == "d" || msg.String() == "x":
		// Delete mapping
		if len(m.mappings) > 0 {
			m.mode = TickerConfirmDelete
			m.statusMsg = ""
		}

	case msg.String() == "v":
		// View default mappings
		m.mode = TickerShowDefaults
		m.defaultCursor = 0
		m.statusMsg = ""
	}

	return m, nil
}

func (m TickerModel) handleAddKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.mode = TickerList
		m.inputs[m.focusIndex].Blur()
		m.statusMsg = ""
		return m, nil

	case tea.KeyTab, tea.KeyShiftTab, tea.KeyDown, tea.KeyUp:
		// Navigate between fields
		if msg.Type == tea.KeyUp || msg.Type == tea.KeyShiftTab {
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = tickerFieldCount - 1
			}
		} else {
			m.focusIndex++
			if m.focusIndex >= tickerFieldCount {
				m.focusIndex = 0
			}
		}

		// Update focus
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focusIndex].Focus()
		return m, textinput.Blink

	case tea.KeyEnter:
		// If on last field, try to save
		if m.focusIndex == tickerFieldCount-1 {
			return m.submitForm()
		}
		// Otherwise move to next field
		m.focusIndex++
		if m.focusIndex >= tickerFieldCount {
			m.focusIndex = 0
		}
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focusIndex].Focus()
		return m, textinput.Blink

	default:
		// Update the focused input
		var cmd tea.Cmd
		m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
		return m, cmd
	}
}

func (m TickerModel) handleSearchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.mode = TickerList
		m.searchInput.Blur()
		m.statusMsg = ""
		return m, nil

	case tea.KeyEnter:
		// Perform search
		query := strings.TrimSpace(m.searchInput.Value())
		if query == "" {
			m.statusMsg = "Enter a search query"
			return m, nil
		}
		m.searching = true
		m.statusMsg = "Searching..."
		m.searchInput.Blur()
		return m, m.performSearch(query)

	default:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}
}

func (m TickerModel) handleSearchResultsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.mode = TickerList
		m.statusMsg = ""
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.searchCursor > 0 {
			m.searchCursor--
		}

	case key.Matches(msg, m.keys.Down):
		if m.searchCursor < len(m.searchResults)-1 {
			m.searchCursor++
		}

	case msg.Type == tea.KeyEnter:
		// Select this result - prompt for ticker
		if len(m.searchResults) > 0 {
			selected := m.searchResults[m.searchCursor]
			// Use the symbol as default ticker
			m.searchTicker = strings.ToUpper(selected.Symbol)
			m.mode = TickerAdd
			m.focusIndex = 0
			m.inputs[tickerFieldTicker].SetValue(m.searchTicker)
			m.inputs[tickerFieldGeckoID].SetValue(selected.ID)
			m.inputs[tickerFieldTicker].Focus()
			return m, textinput.Blink
		}
	}

	return m, nil
}

func (m TickerModel) handleDeleteConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm delete
		if len(m.mappings) > 0 {
			ticker := m.mappings[m.cursor].Ticker
			m.mode = TickerList
			return m, m.deleteMapping(ticker)
		}
	case "n", "N", "escape":
		m.mode = TickerList
		m.statusMsg = ""
	}
	return m, nil
}

func (m TickerModel) handleDefaultsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.mode = TickerList
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.defaultCursor > 0 {
			m.defaultCursor--
		}

	case key.Matches(msg, m.keys.Down):
		if m.defaultCursor < len(m.defaultMappings)-1 {
			m.defaultCursor++
		}

	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	}

	return m, nil
}

func (m *TickerModel) resetForm() {
	m.inputs[tickerFieldTicker].SetValue("")
	m.inputs[tickerFieldGeckoID].SetValue("")
}

func (m TickerModel) submitForm() (tea.Model, tea.Cmd) {
	ticker := strings.ToUpper(strings.TrimSpace(m.inputs[tickerFieldTicker].Value()))
	if ticker == "" {
		m.statusMsg = "Ticker is required"
		return m, nil
	}

	geckoID := strings.ToLower(strings.TrimSpace(m.inputs[tickerFieldGeckoID].Value()))
	if geckoID == "" {
		m.statusMsg = "CoinGecko ID is required"
		return m, nil
	}

	// Blur all inputs
	for i := range m.inputs {
		m.inputs[i].Blur()
	}

	return m, m.addMapping(ticker, geckoID)
}

func (m TickerModel) addMapping(ticker, geckoID string) tea.Cmd {
	return func() tea.Msg {
		err := m.config.SetTickerMapping(ticker, geckoID)
		if err != nil {
			return TickerMappingAddedMsg{Error: err}
		}
		return TickerMappingAddedMsg{Ticker: ticker, GeckoID: geckoID}
	}
}

func (m TickerModel) deleteMapping(ticker string) tea.Cmd {
	return func() tea.Msg {
		err := m.config.RemoveTickerMapping(ticker)
		if err != nil {
			return TickerMappingDeletedMsg{Ticker: ticker, Error: err}
		}
		return TickerMappingDeletedMsg{Ticker: ticker}
	}
}

func (m TickerModel) performSearch(query string) tea.Cmd {
	return func() tea.Msg {
		results, err := m.priceService.SearchCoins(query)
		if err != nil {
			return TickerSearchResultsMsg{Error: err}
		}
		return TickerSearchResultsMsg{Results: results}
	}
}

// View renders the ticker view
func (m TickerModel) View() string {
	switch m.mode {
	case TickerAdd:
		return m.renderAddForm()
	case TickerSearch:
		return m.renderSearchForm()
	case TickerSearchResults:
		return m.renderSearchResults()
	case TickerConfirmDelete:
		return m.renderDeleteConfirm()
	case TickerShowDefaults:
		return m.renderDefaults()
	default:
		return m.renderList()
	}
}

func (m TickerModel) renderList() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Render("TICKER MAPPINGS")

	b.WriteString(title)
	b.WriteString("\n\n")

	// Custom mappings section
	customHeader := lipgloss.NewStyle().
		Foreground(tui.AccentColor).
		Bold(true).
		Render("Custom Mappings:")
	b.WriteString(customHeader)
	b.WriteString("\n")

	if len(m.mappings) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Italic(true)
		b.WriteString(emptyStyle.Render("  No custom mappings. Press 'a' to add or 's' to search."))
		b.WriteString("\n")
	} else {
		for i, mapping := range m.mappings {
			b.WriteString(m.renderMappingRow(i, mapping, true))
		}
	}

	// Default mappings info
	b.WriteString("\n")
	defaultInfo := lipgloss.NewStyle().
		Foreground(tui.MutedColor).
		Render(fmt.Sprintf("Default Mappings: %d built-in (press 'v' to view)", len(m.defaultMappings)))
	b.WriteString(defaultInfo)

	// Status message
	if m.statusMsg != "" {
		b.WriteString("\n\n")
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

func (m TickerModel) renderMappingRow(index int, mapping MappingItem, isCustomList bool) string {
	isSelected := false
	if isCustomList {
		isSelected = index == m.cursor
	} else {
		isSelected = index == m.defaultCursor
	}

	// Cursor
	cursor := "  "
	if isSelected {
		cursor = lipgloss.NewStyle().
			Foreground(tui.PrimaryColor).
			Render("> ")
	}

	// Build row
	rowStyle := lipgloss.NewStyle().Foreground(tui.TextColor)
	if isSelected {
		rowStyle = rowStyle.Bold(true).Foreground(tui.PrimaryColor)
	}

	row := fmt.Sprintf("%-8s -> %s", mapping.Ticker, mapping.GeckoID)

	return cursor + rowStyle.Render(row) + "\n"
}

func (m TickerModel) renderListHelp() string {
	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)

	var help string
	if len(m.mappings) > 0 {
		help = fmt.Sprintf("%s navigate  %s add  %s search  %s delete  %s defaults  %s back",
			tui.HelpKeyStyle.Render("↑↓"),
			tui.HelpKeyStyle.Render("a"),
			tui.HelpKeyStyle.Render("s"),
			tui.HelpKeyStyle.Render("d"),
			tui.HelpKeyStyle.Render("v"),
			tui.HelpKeyStyle.Render("esc"))
	} else {
		help = fmt.Sprintf("%s add  %s search  %s defaults  %s back",
			tui.HelpKeyStyle.Render("a"),
			tui.HelpKeyStyle.Render("s"),
			tui.HelpKeyStyle.Render("v"),
			tui.HelpKeyStyle.Render("esc"))
	}

	return helpStyle.Render(help)
}

func (m TickerModel) renderAddForm() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Render("ADD MAPPING")

	b.WriteString(title)
	b.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().
		Foreground(tui.SubtleTextColor).
		Width(14)

	focusedLabelStyle := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Width(14)

	fields := []struct {
		label string
		index int
	}{
		{"Ticker:", tickerFieldTicker},
		{"CoinGecko ID:", tickerFieldGeckoID},
	}

	for _, f := range fields {
		ls := labelStyle
		if m.focusIndex == f.index {
			ls = focusedLabelStyle
		}
		b.WriteString(ls.Render(f.label))
		b.WriteString(m.inputs[f.index].View())
		b.WriteString("\n")
	}

	// Hint
	b.WriteString("\n")
	hintStyle := lipgloss.NewStyle().Foreground(tui.MutedColor).Italic(true)
	b.WriteString(hintStyle.Render("Tip: Use 's' from main screen to search CoinGecko"))

	// Status message (for errors)
	if m.statusMsg != "" {
		b.WriteString("\n")
		errorStyle := lipgloss.NewStyle().
			Foreground(tui.ErrorColor).
			Italic(true)
		b.WriteString(errorStyle.Render(m.statusMsg))
	}

	// Help
	b.WriteString("\n\n")
	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)
	help := fmt.Sprintf("%s navigate  %s save  %s cancel",
		tui.HelpKeyStyle.Render("tab"),
		tui.HelpKeyStyle.Render("enter"),
		tui.HelpKeyStyle.Render("esc"))
	b.WriteString(helpStyle.Render(help))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.BorderColor).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

func (m TickerModel) renderSearchForm() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Render("SEARCH COINGECKO")

	b.WriteString(title)
	b.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Width(10)

	b.WriteString(labelStyle.Render("Search:"))
	b.WriteString(m.searchInput.View())
	b.WriteString("\n")

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
	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)
	help := fmt.Sprintf("%s search  %s cancel",
		tui.HelpKeyStyle.Render("enter"),
		tui.HelpKeyStyle.Render("esc"))
	b.WriteString(helpStyle.Render(help))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.BorderColor).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

func (m TickerModel) renderSearchResults() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Render("SEARCH RESULTS")

	b.WriteString(title)
	b.WriteString("\n\n")

	// Column headers
	headerStyle := lipgloss.NewStyle().
		Foreground(tui.MutedColor).
		Bold(true)
	header := fmt.Sprintf("  %-20s  %-25s  %-8s  %s",
		"ID", "Name", "Symbol", "Rank")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// Separator
	sepStyle := lipgloss.NewStyle().Foreground(tui.BorderColor)
	b.WriteString(sepStyle.Render(strings.Repeat("─", tui.SeparatorWidthTickerSearch)))
	b.WriteString("\n")

	// Results (show up to 10)
	maxResults := 10
	if len(m.searchResults) < maxResults {
		maxResults = len(m.searchResults)
	}

	for i := 0; i < maxResults; i++ {
		r := m.searchResults[i]
		isSelected := i == m.searchCursor

		cursor := "  "
		if isSelected {
			cursor = lipgloss.NewStyle().
				Foreground(tui.PrimaryColor).
				Render("> ")
		}

		rowStyle := lipgloss.NewStyle().Foreground(tui.TextColor)
		if isSelected {
			rowStyle = rowStyle.Bold(true).Foreground(tui.PrimaryColor)
		}

		// Truncate long values
		id := format.TruncateID(r.ID)
		name := format.TruncateName(r.Name)

		rank := "-"
		if r.Rank > 0 {
			rank = fmt.Sprintf("#%d", r.Rank)
		}

		row := fmt.Sprintf("%-20s  %-25s  %-8s  %s",
			id, name, strings.ToUpper(r.Symbol), rank)
		b.WriteString(cursor + rowStyle.Render(row) + "\n")
	}

	if len(m.searchResults) > maxResults {
		moreStyle := lipgloss.NewStyle().Foreground(tui.MutedColor).Italic(true)
		b.WriteString(moreStyle.Render(fmt.Sprintf("  ... and %d more results", len(m.searchResults)-maxResults)))
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)
	help := fmt.Sprintf("%s navigate  %s select & map  %s back",
		tui.HelpKeyStyle.Render("↑↓"),
		tui.HelpKeyStyle.Render("enter"),
		tui.HelpKeyStyle.Render("esc"))
	b.WriteString(helpStyle.Render(help))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.BorderColor).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

func (m TickerModel) renderDeleteConfirm() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(tui.ErrorColor).
		Bold(true).
		Padding(0, 2).
		Render("CONFIRM DELETE")

	b.WriteString(title)
	b.WriteString("\n\n")

	if m.cursor < len(m.mappings) {
		mapping := m.mappings[m.cursor]

		infoStyle := lipgloss.NewStyle().Foreground(tui.TextColor)
		info := fmt.Sprintf("Remove mapping %s -> %s?", mapping.Ticker, mapping.GeckoID)
		b.WriteString(infoStyle.Render(info))
		b.WriteString("\n")

		// Check if default exists
		defaults := prices.GetDefaultMappings()
		if defaultID, ok := defaults[mapping.Ticker]; ok {
			noteStyle := lipgloss.NewStyle().Foreground(tui.MutedColor).Italic(true)
			b.WriteString(noteStyle.Render(fmt.Sprintf("(will revert to default: %s)", defaultID)))
		}
		b.WriteString("\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)
	help := fmt.Sprintf("%s confirm  %s cancel",
		tui.HelpKeyStyle.Render("y"),
		tui.HelpKeyStyle.Render("n/esc"))
	b.WriteString(helpStyle.Render(help))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.ErrorColor).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

func (m TickerModel) renderDefaults() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Render("DEFAULT MAPPINGS")

	b.WriteString(title)
	b.WriteString("\n\n")

	// Show count
	countStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)
	b.WriteString(countStyle.Render(fmt.Sprintf("%d built-in mappings (read-only)", len(m.defaultMappings))))
	b.WriteString("\n\n")

	// Column headers
	headerStyle := lipgloss.NewStyle().
		Foreground(tui.MutedColor).
		Bold(true)
	header := fmt.Sprintf("  %-8s  %s", "Ticker", "CoinGecko ID")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// Separator
	sepStyle := lipgloss.NewStyle().Foreground(tui.BorderColor)
	b.WriteString(sepStyle.Render(strings.Repeat("─", tui.SeparatorWidthTickerDefaults)))
	b.WriteString("\n")

	// Show mappings with scrolling
	visibleCount := tui.DefaultMappingsVisibleCount
	start := m.defaultCursor - visibleCount/2
	if start < 0 {
		start = 0
	}
	end := start + visibleCount
	if end > len(m.defaultMappings) {
		end = len(m.defaultMappings)
		start = end - visibleCount
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		b.WriteString(m.renderMappingRow(i, m.defaultMappings[i], false))
	}

	// Scroll indicator
	if len(m.defaultMappings) > visibleCount {
		scrollInfo := lipgloss.NewStyle().Foreground(tui.MutedColor).Italic(true)
		b.WriteString(scrollInfo.Render(fmt.Sprintf("\n  Showing %d-%d of %d", start+1, end, len(m.defaultMappings))))
	}

	// Help
	b.WriteString("\n\n")
	helpStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)
	help := fmt.Sprintf("%s scroll  %s back",
		tui.HelpKeyStyle.Render("↑↓"),
		tui.HelpKeyStyle.Render("esc"))
	b.WriteString(helpStyle.Render(help))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.BorderColor).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

// GetConfig returns the config instance
func (m TickerModel) GetConfig() *config.ConfigStore {
	return m.config
}
