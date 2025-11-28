package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pretty-andrechal/follyo/internal/portfolio"
	"github.com/pretty-andrechal/follyo/internal/storage"
)

func setupTestApp(t *testing.T) (*App, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "follyo-app-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dataPath := filepath.Join(tmpDir, "test.json")
	s, err := storage.New(dataPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create storage: %v", err)
	}

	p := portfolio.New(s)
	app := NewApp(s, p)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return app, cleanup
}

func TestNewApp(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	if app.portfolio == nil {
		t.Error("portfolio should be set")
	}

	if app.storage == nil {
		t.Error("storage should be set")
	}

	if app.currentView != ViewMenu {
		t.Errorf("expected initial view to be ViewMenu, got %v", app.currentView)
	}

	if app.tickerMappings == nil {
		t.Error("tickerMappings should be initialized")
	}
}

func TestApp_Init(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	cmd := app.Init()
	if cmd == nil {
		t.Error("Init should return commands")
	}
}

func TestApp_SetMenuModel(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create a mock menu model
	mockMenu := &mockModel{}
	app.SetMenuModel(mockMenu)

	if !app.views.Has(ViewMenu) {
		t.Error("menu model not set correctly")
	}
}

func TestApp_SetSummaryModel(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	mockSummary := &mockModel{}
	app.SetSummaryModel(mockSummary)

	if !app.views.Has(ViewSummary) {
		t.Error("summary model not set correctly")
	}
}

func TestApp_SetTickerMappings(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	mappings := map[string]string{"BTC": "bitcoin", "ETH": "ethereum"}
	app.SetTickerMappings(mappings)

	if len(app.tickerMappings) != 2 {
		t.Errorf("expected 2 mappings, got %d", len(app.tickerMappings))
	}

	if app.tickerMappings["BTC"] != "bitcoin" {
		t.Error("BTC mapping not set correctly")
	}
}

func TestApp_GetPortfolio(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	p := app.GetPortfolio()
	if p == nil {
		t.Error("GetPortfolio should return portfolio")
	}
}

func TestApp_GetTickerMappings(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	mappings := map[string]string{"TEST": "test-id"}
	app.SetTickerMappings(mappings)

	result := app.GetTickerMappings()
	if result["TEST"] != "test-id" {
		t.Error("GetTickerMappings should return set mappings")
	}
}

func TestApp_QuitOnCtrlC(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	newApp, cmd := app.Update(msg)
	app = newApp.(*App)

	if !app.quitting {
		t.Error("app should be quitting after Ctrl+C")
	}

	if cmd == nil {
		t.Error("should return quit command")
	}
}

func TestApp_MenuSelectSummary(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	mockSummary := &mockModel{}
	app.SetSummaryModel(mockSummary)

	msg := MenuSelectMsg{Action: "summary"}
	newApp, _ := app.Update(msg)
	app = newApp.(*App)

	if app.currentView != ViewSummary {
		t.Errorf("expected ViewSummary, got %v", app.currentView)
	}
}

func TestApp_MenuSelectBuy(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	msg := MenuSelectMsg{Action: "buy"}
	newApp, _ := app.Update(msg)
	app = newApp.(*App)

	if app.currentView != ViewBuy {
		t.Errorf("expected ViewBuy, got %v", app.currentView)
	}

	// Buy is now implemented, so statusMsg should be empty
	if app.statusMsg != "" {
		t.Errorf("status message should be empty for implemented view, got '%s'", app.statusMsg)
	}
}

func TestApp_MenuSelectAllOptions(t *testing.T) {
	tests := []struct {
		action       string
		expectedView ViewType
	}{
		{"summary", ViewSummary},
		{"buy", ViewBuy},
		{"sell", ViewSell},
		{"stake", ViewStake},
		{"loan", ViewLoan},
		{"snapshots", ViewSnapshots},
		{"settings", ViewSettings},
	}

	for _, tt := range tests {
		app, cleanup := setupTestApp(t)

		msg := MenuSelectMsg{Action: tt.action}
		newApp, _ := app.Update(msg)
		app = newApp.(*App)

		if app.currentView != tt.expectedView {
			t.Errorf("action '%s': expected view %v, got %v",
				tt.action, tt.expectedView, app.currentView)
		}

		cleanup()
	}
}

func TestApp_BackToMenu(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Switch to summary view
	app.currentView = ViewSummary
	app.statusMsg = "some message"

	msg := BackToMenuMsg{}
	newApp, _ := app.Update(msg)
	app = newApp.(*App)

	if app.currentView != ViewMenu {
		t.Errorf("expected ViewMenu, got %v", app.currentView)
	}

	if app.statusMsg != "" {
		t.Error("status message should be cleared")
	}
}

func TestApp_WindowResize(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	mockMenu := &mockModel{}
	app.SetMenuModel(mockMenu)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newApp, _ := app.Update(msg)
	app = newApp.(*App)

	if app.width != 100 {
		t.Errorf("expected width 100, got %d", app.width)
	}
	if app.height != 50 {
		t.Errorf("expected height 50, got %d", app.height)
	}
}

func TestApp_ViewMenu(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.width = 80
	app.height = 24

	mockMenu := &mockModel{viewContent: "Menu Content"}
	app.SetMenuModel(mockMenu)

	view := app.View()

	if !strings.Contains(view, "Menu Content") {
		t.Error("view should contain menu content")
	}
}

func TestApp_ViewSummary(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.width = 80
	app.height = 24
	app.currentView = ViewSummary

	mockSummary := &mockModel{viewContent: "Summary Content"}
	app.SetSummaryModel(mockSummary)

	view := app.View()

	if !strings.Contains(view, "Summary Content") {
		t.Error("view should contain summary content")
	}
}

func TestApp_ViewQuitting(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.quitting = true
	view := app.View()

	if view != "" {
		t.Error("quitting view should be empty")
	}
}

func TestApp_StatusBar(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.width = 80
	app.height = 24

	mockMenu := &mockModel{viewContent: ""}
	app.SetMenuModel(mockMenu)

	view := app.View()

	if !strings.Contains(view, "FOLLYO") {
		t.Error("status bar should contain FOLLYO")
	}
}

func TestApp_StatusBarWithError(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.width = 80
	app.height = 24
	app.err = os.ErrNotExist

	mockMenu := &mockModel{viewContent: ""}
	app.SetMenuModel(mockMenu)

	view := app.View()

	if !strings.Contains(view, "Error") {
		t.Error("status bar should show error")
	}
}

func TestApp_SetAllModels(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	mock := &mockModel{}

	// Test all Set*Model functions
	app.SetSettingsModel(mock)
	if !app.views.Has(ViewSettings) {
		t.Error("settings model not set")
	}

	app.SetSnapshotsModel(mock)
	if !app.views.Has(ViewSnapshots) {
		t.Error("snapshots model not set")
	}

	app.SetBuyModel(mock)
	if !app.views.Has(ViewBuy) {
		t.Error("buy model not set")
	}

	app.SetSellModel(mock)
	if !app.views.Has(ViewSell) {
		t.Error("sell model not set")
	}

	app.SetStakeModel(mock)
	if !app.views.Has(ViewStake) {
		t.Error("stake model not set")
	}

	app.SetLoanModel(mock)
	if !app.views.Has(ViewLoan) {
		t.Error("loan model not set")
	}

	app.SetTickerModel(mock)
	if !app.views.Has(ViewTicker) {
		t.Error("ticker model not set")
	}
}

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	// Test that key bindings are configured
	if len(km.Up.Keys()) == 0 {
		t.Error("Up key not configured")
	}
	if len(km.Down.Keys()) == 0 {
		t.Error("Down key not configured")
	}
	if len(km.Select.Keys()) == 0 {
		t.Error("Select key not configured")
	}
	if len(km.Back.Keys()) == 0 {
		t.Error("Back key not configured")
	}
	if len(km.Quit.Keys()) == 0 {
		t.Error("Quit key not configured")
	}

	// Test help functions
	short := km.ShortHelp()
	if len(short) == 0 {
		t.Error("ShortHelp should return key bindings")
	}

	full := km.FullHelp()
	if len(full) == 0 {
		t.Error("FullHelp should return key binding groups")
	}
}

func TestViewRegistry_Get(t *testing.T) {
	registry := NewViewRegistry()

	mock := &mockModel{viewContent: "test"}
	registry.Set(ViewMenu, mock)

	// Test Get existing
	model := registry.Get(ViewMenu)
	if model == nil {
		t.Error("expected menu to exist")
	}
	if model.View() != "test" {
		t.Error("expected model content to match")
	}

	// Test Get non-existing
	model = registry.Get(ViewSummary)
	if model != nil {
		t.Error("expected summary to be nil")
	}
}

func TestViewRegistry_View(t *testing.T) {
	registry := NewViewRegistry()

	mock := &mockModel{viewContent: "view content"}
	registry.Set(ViewMenu, mock)

	// Test View for existing model
	view := registry.View(ViewMenu)
	if view != "view content" {
		t.Errorf("expected 'view content', got '%s'", view)
	}

	// Test View for non-existing model returns loading text
	view = registry.View(ViewSummary)
	if !strings.Contains(view, "Loading") {
		t.Errorf("expected loading text for non-existing view, got '%s'", view)
	}
}

func TestApp_RenderPlaceholder(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.width = 80
	app.height = 24
	app.currentView = ViewTicker // A view without a model set

	// Don't set any model for ViewTicker - should render placeholder
	view := app.View()

	// The view should render something (even if placeholder)
	if view == "" {
		t.Error("should render something for view without model")
	}
}

func TestApp_ViewTitles(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.width = 80
	app.height = 24

	// Set mock models for all views
	mock := &mockModel{viewContent: "content"}
	app.SetMenuModel(mock)
	app.SetSummaryModel(mock)
	app.SetBuyModel(mock)
	app.SetSellModel(mock)
	app.SetStakeModel(mock)
	app.SetLoanModel(mock)
	app.SetSnapshotsModel(mock)
	app.SetSettingsModel(mock)
	app.SetTickerModel(mock)

	// Test that each view renders without error
	views := []ViewType{
		ViewMenu, ViewSummary, ViewBuy, ViewSell,
		ViewStake, ViewLoan, ViewSnapshots, ViewSettings, ViewTicker,
	}

	for _, v := range views {
		app.currentView = v
		view := app.View()
		if view == "" {
			t.Errorf("view %v should render content", v)
		}
		// Status bar should always contain FOLLYO
		if !strings.Contains(view, "FOLLYO") {
			t.Errorf("view %v should contain FOLLYO in status bar", v)
		}
	}
}

// mockModel is a simple mock for testing
type mockModel struct {
	viewContent string
}

func (m *mockModel) Init() tea.Cmd {
	return nil
}

func (m *mockModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *mockModel) View() string {
	return m.viewContent
}
