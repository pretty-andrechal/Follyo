package views

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pretty-andrechal/follyo/internal/config"
	"github.com/pretty-andrechal/follyo/internal/prices"
)

func setupTestTickerModel(t *testing.T) (TickerModel, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "follyo-ticker-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create config
	configPath := filepath.Join(tmpDir, "config.json")
	cfg, err := config.New(configPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create config: %v", err)
	}

	model := NewTickerModel(cfg)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return model, cleanup
}

func TestTickerModel_Init(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	cmd := model.Init()

	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestTickerModel_InitialState(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	if model.mode != TickerList {
		t.Error("initial mode should be TickerList")
	}

	if model.cursor != 0 {
		t.Error("initial cursor should be 0")
	}

	if model.focusIndex != 0 {
		t.Error("initial focusIndex should be 0")
	}
}

func TestTickerModel_LoadDefaultMappings(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	// Should have loaded default mappings
	if len(model.defaultMappings) == 0 {
		t.Error("should have loaded default mappings")
	}

	// Default mappings should include common tickers
	found := false
	for _, m := range model.defaultMappings {
		if m.Ticker == "BTC" {
			found = true
			break
		}
	}
	if !found {
		t.Error("default mappings should include BTC")
	}
}

func TestTickerModel_NavigationKeys(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	// Add test mappings
	cfg := model.GetConfig()
	cfg.SetTickerMapping("TEST1", "test-1")
	cfg.SetTickerMapping("TEST2", "test-2")
	model.loadMappings()

	// Test down navigation
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.cursor != 1 {
		t.Error("cursor should move down")
	}

	// Test up navigation
	msg = tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = model.Update(msg)
	model = newModel.(TickerModel)

	if model.cursor != 0 {
		t.Error("cursor should move up")
	}

	// Test vim j navigation
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ = model.Update(msg)
	model = newModel.(TickerModel)

	if model.cursor != 1 {
		t.Error("j key should move cursor down")
	}

	// Test vim k navigation
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ = model.Update(msg)
	model = newModel.(TickerModel)

	if model.cursor != 0 {
		t.Error("k key should move cursor up")
	}
}

func TestTickerModel_CursorBounds(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	// Add test mappings
	cfg := model.GetConfig()
	cfg.SetTickerMapping("TEST1", "test-1")
	cfg.SetTickerMapping("TEST2", "test-2")
	model.loadMappings()

	// Try to go below 0
	msg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.cursor != 0 {
		t.Error("cursor should not go below 0")
	}

	// Move to end
	model.cursor = 1

	// Try to go past end
	msg = tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ = model.Update(msg)
	model = newModel.(TickerModel)

	if model.cursor != 1 {
		t.Error("cursor should not go past end")
	}
}

func TestTickerModel_AddMode(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	// Press 'a' to enter add mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.mode != TickerAdd {
		t.Error("should be in add mode")
	}

	if model.focusIndex != 0 {
		t.Error("focus should be on first field")
	}
}

func TestTickerModel_AddModeWithN(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	// Press 'n' to enter add mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.mode != TickerAdd {
		t.Error("should be in add mode")
	}
}

func TestTickerModel_SearchMode(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	// Press 's' to enter search mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.mode != TickerSearch {
		t.Error("should be in search mode")
	}
}

func TestTickerModel_SearchModeWithSlash(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	// Press '/' to enter search mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.mode != TickerSearch {
		t.Error("should be in search mode")
	}
}

func TestTickerModel_ViewDefaultsMode(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	// Press 'v' to view defaults
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.mode != TickerShowDefaults {
		t.Error("should be in show defaults mode")
	}
}

func TestTickerModel_AddFormNavigation(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	// Enter add mode
	model.mode = TickerAdd
	model.focusIndex = 0

	// Tab to next field
	msg := tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.focusIndex != 1 {
		t.Error("tab should move to next field")
	}

	// Shift+Tab to previous field
	msg = tea.KeyMsg{Type: tea.KeyShiftTab}
	newModel, _ = model.Update(msg)
	model = newModel.(TickerModel)

	if model.focusIndex != 0 {
		t.Error("shift+tab should move to previous field")
	}
}

func TestTickerModel_AddFormWrapAround(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	// Enter add mode at last field
	model.mode = TickerAdd
	model.focusIndex = tickerFieldCount - 1

	// Tab should wrap to first field
	msg := tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.focusIndex != 0 {
		t.Error("tab at last field should wrap to first")
	}

	// Shift+Tab should wrap to last field
	msg = tea.KeyMsg{Type: tea.KeyShiftTab}
	newModel, _ = model.Update(msg)
	model = newModel.(TickerModel)

	if model.focusIndex != tickerFieldCount-1 {
		t.Error("shift+tab at first field should wrap to last")
	}
}

func TestTickerModel_AddFormEscape(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	// Enter add mode
	model.mode = TickerAdd

	// Press escape
	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.mode != TickerList {
		t.Error("escape should return to list mode")
	}
}

func TestTickerModel_SearchFormEscape(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	// Enter search mode
	model.mode = TickerSearch

	// Press escape
	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.mode != TickerList {
		t.Error("escape should return to list mode")
	}
}

func TestTickerModel_DefaultsEscape(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	// Enter defaults mode
	model.mode = TickerShowDefaults

	// Press escape
	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.mode != TickerList {
		t.Error("escape should return to list mode")
	}
}

func TestTickerModel_DeleteMode(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	// Add test mapping
	cfg := model.GetConfig()
	cfg.SetTickerMapping("TEST", "test-coin")
	model.loadMappings()

	// Press 'd' to enter delete mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.mode != TickerConfirmDelete {
		t.Error("should be in confirm delete mode")
	}
}

func TestTickerModel_DeleteModeNoMappings(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	// Try to delete with no custom mappings
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.mode != TickerList {
		t.Error("should stay in list mode when no mappings")
	}
}

func TestTickerModel_DeleteConfirmCancel(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	// Add test mapping
	cfg := model.GetConfig()
	cfg.SetTickerMapping("TEST", "test-coin")
	model.loadMappings()

	// Enter delete mode
	model.mode = TickerConfirmDelete

	// Press 'n' to cancel
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.mode != TickerList {
		t.Error("n should cancel delete and return to list")
	}
}

func TestTickerModel_FormValidation_EmptyTicker(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	model.mode = TickerAdd
	model.focusIndex = tickerFieldCount - 1

	// Leave ticker empty
	model.inputs[tickerFieldGeckoID].SetValue("test-coin")

	// Try to submit
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.statusMsg != "Ticker is required" {
		t.Error("should show ticker required error")
	}
}

func TestTickerModel_FormValidation_EmptyGeckoID(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	model.mode = TickerAdd
	model.focusIndex = tickerFieldCount - 1

	model.inputs[tickerFieldTicker].SetValue("TEST")
	// Leave geckoID empty

	// Try to submit
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.statusMsg != "CoinGecko ID is required" {
		t.Error("should show geckoID required error")
	}
}

func TestTickerModel_SearchEmptyQuery(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	model.mode = TickerSearch
	// Leave search empty

	// Try to submit
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.statusMsg != "Enter a search query" {
		t.Error("should show search query required error")
	}
}

func TestTickerModel_View_EmptyList(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	view := model.View()

	if !strings.Contains(view, "No custom mappings") {
		t.Error("empty list should show no mappings message")
	}

	if !strings.Contains(view, "Default Mappings") {
		t.Error("should show default mappings count")
	}
}

func TestTickerModel_View_WithMappings(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	cfg := model.GetConfig()
	cfg.SetTickerMapping("TEST", "test-coin")
	model.loadMappings()

	view := model.View()

	if !strings.Contains(view, "TEST") {
		t.Error("view should show ticker")
	}

	if !strings.Contains(view, "test-coin") {
		t.Error("view should show gecko ID")
	}
}

func TestTickerModel_View_AddForm(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	model.mode = TickerAdd

	view := model.View()

	if !strings.Contains(view, "ADD MAPPING") {
		t.Error("add form should show title")
	}

	if !strings.Contains(view, "Ticker") {
		t.Error("add form should show ticker label")
	}

	if !strings.Contains(view, "CoinGecko ID") {
		t.Error("add form should show geckoID label")
	}
}

func TestTickerModel_View_SearchForm(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	model.mode = TickerSearch

	view := model.View()

	if !strings.Contains(view, "SEARCH") {
		t.Error("search form should show title")
	}
}

func TestTickerModel_View_SearchResults(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	model.mode = TickerSearchResults
	model.searchResults = []prices.SearchResult{
		{ID: "bitcoin", Name: "Bitcoin", Symbol: "btc", Rank: 1},
		{ID: "ethereum", Name: "Ethereum", Symbol: "eth", Rank: 2},
	}

	view := model.View()

	if !strings.Contains(view, "SEARCH RESULTS") {
		t.Error("should show search results title")
	}

	if !strings.Contains(view, "bitcoin") {
		t.Error("should show bitcoin result")
	}
}

func TestTickerModel_View_DeleteConfirm(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	cfg := model.GetConfig()
	cfg.SetTickerMapping("TEST", "test-coin")
	model.loadMappings()

	model.mode = TickerConfirmDelete

	view := model.View()

	if !strings.Contains(view, "CONFIRM") {
		t.Error("delete confirm should show title")
	}

	if !strings.Contains(view, "TEST") {
		t.Error("delete confirm should show ticker")
	}
}

func TestTickerModel_View_Defaults(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	model.mode = TickerShowDefaults

	view := model.View()

	if !strings.Contains(view, "DEFAULT MAPPINGS") {
		t.Error("defaults view should show title")
	}

	if !strings.Contains(view, "built-in") {
		t.Error("defaults view should show built-in count")
	}
}

func TestTickerModel_MappingAddedMsg(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	model.mode = TickerAdd

	msg := TickerMappingAddedMsg{Ticker: "TEST", GeckoID: "test-coin"}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.mode != TickerList {
		t.Error("should return to list mode after adding")
	}

	if !strings.Contains(model.statusMsg, "TEST") {
		t.Error("status should mention ticker")
	}
}

func TestTickerModel_MappingAddedMsgError(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	model.mode = TickerAdd

	msg := TickerMappingAddedMsg{Error: os.ErrNotExist}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if !strings.Contains(model.statusMsg, "Error") {
		t.Error("should show error message")
	}
}

func TestTickerModel_MappingDeletedMsg(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	msg := TickerMappingDeletedMsg{Ticker: "TEST"}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if !strings.Contains(model.statusMsg, "Removed") {
		t.Error("should show removed message")
	}
}

func TestTickerModel_SearchResultsMsg(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	results := []prices.SearchResult{
		{ID: "bitcoin", Name: "Bitcoin", Symbol: "btc", Rank: 1},
	}

	msg := TickerSearchResultsMsg{Results: results}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.mode != TickerSearchResults {
		t.Error("should switch to search results mode")
	}

	if len(model.searchResults) != 1 {
		t.Error("should have search results")
	}
}

func TestTickerModel_SearchResultsMsgEmpty(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	msg := TickerSearchResultsMsg{Results: []prices.SearchResult{}}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.mode != TickerList {
		t.Error("should return to list mode when no results")
	}

	if !strings.Contains(model.statusMsg, "No results") {
		t.Error("should show no results message")
	}
}

func TestTickerModel_SearchResultsMsgError(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	msg := TickerSearchResultsMsg{Error: os.ErrNotExist}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.mode != TickerList {
		t.Error("should return to list mode on error")
	}

	if !strings.Contains(model.statusMsg, "error") {
		t.Error("should show error message")
	}
}

func TestTickerModel_SearchResultsNavigation(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	model.mode = TickerSearchResults
	model.searchResults = []prices.SearchResult{
		{ID: "bitcoin", Name: "Bitcoin", Symbol: "btc", Rank: 1},
		{ID: "ethereum", Name: "Ethereum", Symbol: "eth", Rank: 2},
	}
	model.searchCursor = 0

	// Navigate down
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.searchCursor != 1 {
		t.Error("should move cursor down in search results")
	}

	// Navigate up
	msg = tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = model.Update(msg)
	model = newModel.(TickerModel)

	if model.searchCursor != 0 {
		t.Error("should move cursor up in search results")
	}
}

func TestTickerModel_SearchResultsSelect(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	model.mode = TickerSearchResults
	model.searchResults = []prices.SearchResult{
		{ID: "bitcoin", Name: "Bitcoin", Symbol: "btc", Rank: 1},
	}
	model.searchCursor = 0

	// Select result
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.mode != TickerAdd {
		t.Error("should switch to add mode after selecting")
	}

	if model.inputs[tickerFieldTicker].Value() != "BTC" {
		t.Error("ticker should be pre-filled with symbol")
	}

	if model.inputs[tickerFieldGeckoID].Value() != "bitcoin" {
		t.Error("gecko ID should be pre-filled")
	}
}

func TestTickerModel_BackToMenu(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("escape should return command")
	}
}

func TestTickerModel_Quit(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("q should return quit command")
	}
}

func TestTickerModel_WindowSizeMsg(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.width != 100 {
		t.Error("width should be updated")
	}

	if model.height != 50 {
		t.Error("height should be updated")
	}
}

func TestTickerModel_DefaultsNavigation(t *testing.T) {
	model, cleanup := setupTestTickerModel(t)
	defer cleanup()

	model.mode = TickerShowDefaults
	model.defaultCursor = 0

	// Navigate down
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := model.Update(msg)
	model = newModel.(TickerModel)

	if model.defaultCursor != 1 {
		t.Error("should move cursor down in defaults")
	}

	// Navigate up
	msg = tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = model.Update(msg)
	model = newModel.(TickerModel)

	if model.defaultCursor != 0 {
		t.Error("should move cursor up in defaults")
	}
}
