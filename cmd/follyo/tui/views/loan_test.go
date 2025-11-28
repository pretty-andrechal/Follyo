package views

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pretty-andrechal/follyo/internal/portfolio"
	"github.com/pretty-andrechal/follyo/internal/storage"
)

func setupTestLoanModel(t *testing.T) (LoanModel, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "follyo-loan-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create storage and portfolio
	dataPath := filepath.Join(tmpDir, "portfolio.json")
	s, err := storage.New(dataPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create storage: %v", err)
	}

	p := portfolio.New(s)

	model := NewLoanModel(p, "")

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return model, cleanup
}

func TestLoanModel_Init(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	cmd := model.Init()

	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestLoanModel_InitialState(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	if model.mode != LoanList {
		t.Error("initial mode should be LoanList")
	}

	if model.cursor != 0 {
		t.Error("initial cursor should be 0")
	}

	if model.focusIndex != 0 {
		t.Error("initial focusIndex should be 0")
	}
}

func TestLoanModel_DefaultPlatform(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "follyo-loan-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataPath := filepath.Join(tmpDir, "portfolio.json")
	s, err := storage.New(dataPath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	p := portfolio.New(s)

	model := NewLoanModel(p, "Nexo")

	if model.defaultPlatform != "Nexo" {
		t.Error("default platform should be set")
	}

	// Check if the platform input has the default value
	if model.inputs[loanFieldPlatform].Value() != "Nexo" {
		t.Error("platform input should have default value")
	}
}

func TestLoanModel_NavigationKeys(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	// Add test loans
	p := model.GetPortfolio()
	p.AddLoan("USDT", 5000, "Nexo", nil, "", "")
	p.AddLoan("USDC", 10000, "Celsius", nil, "", "")
	model.loadLoans()

	// Test down navigation
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.cursor != 1 {
		t.Error("cursor should move down")
	}

	// Test up navigation
	msg = tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = model.Update(msg)
	model = newModel.(LoanModel)

	if model.cursor != 0 {
		t.Error("cursor should move up")
	}

	// Test vim j navigation
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ = model.Update(msg)
	model = newModel.(LoanModel)

	if model.cursor != 1 {
		t.Error("j key should move cursor down")
	}

	// Test vim k navigation
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ = model.Update(msg)
	model = newModel.(LoanModel)

	if model.cursor != 0 {
		t.Error("k key should move cursor up")
	}
}

func TestLoanModel_CursorBounds(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	// Add test loans
	p := model.GetPortfolio()
	p.AddLoan("USDT", 5000, "Nexo", nil, "", "")
	p.AddLoan("USDC", 10000, "Celsius", nil, "", "")
	model.loadLoans()

	// Try to go below 0
	msg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.cursor != 0 {
		t.Error("cursor should not go below 0")
	}

	// Move to end
	model.cursor = 1

	// Try to go past end
	msg = tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ = model.Update(msg)
	model = newModel.(LoanModel)

	if model.cursor != 1 {
		t.Error("cursor should not go past end")
	}
}

func TestLoanModel_AddMode(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	// Press 'a' to enter add mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.mode != LoanAdd {
		t.Error("should be in add mode")
	}

	if model.focusIndex != 0 {
		t.Error("focus should be on first field")
	}
}

func TestLoanModel_AddModeWithN(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	// Press 'n' to enter add mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.mode != LoanAdd {
		t.Error("should be in add mode")
	}
}

func TestLoanModel_AddFormNavigation(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	// Enter add mode
	model.mode = LoanAdd
	model.focusIndex = 0

	// Tab to next field
	msg := tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.focusIndex != 1 {
		t.Error("tab should move to next field")
	}

	// Shift+Tab to previous field
	msg = tea.KeyMsg{Type: tea.KeyShiftTab}
	newModel, _ = model.Update(msg)
	model = newModel.(LoanModel)

	if model.focusIndex != 0 {
		t.Error("shift+tab should move to previous field")
	}
}

func TestLoanModel_AddFormWrapAround(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	// Enter add mode at last field
	model.mode = LoanAdd
	model.focusIndex = loanFieldCount - 1

	// Tab should wrap to first field
	msg := tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.focusIndex != 0 {
		t.Error("tab at last field should wrap to first")
	}

	// Shift+Tab should wrap to last field
	msg = tea.KeyMsg{Type: tea.KeyShiftTab}
	newModel, _ = model.Update(msg)
	model = newModel.(LoanModel)

	if model.focusIndex != loanFieldCount-1 {
		t.Error("shift+tab at first field should wrap to last")
	}
}

func TestLoanModel_AddFormEscape(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	// Enter add mode
	model.mode = LoanAdd

	// Press escape
	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.mode != LoanList {
		t.Error("escape should return to list mode")
	}
}

func TestLoanModel_DeleteMode(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	// Add test loan
	p := model.GetPortfolio()
	p.AddLoan("USDT", 5000, "Nexo", nil, "", "")
	model.loadLoans()

	// Press 'd' to enter delete mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.mode != LoanConfirmDelete {
		t.Error("should be in confirm delete mode")
	}
}

func TestLoanModel_DeleteModeWithX(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	// Add test loan
	p := model.GetPortfolio()
	p.AddLoan("USDT", 5000, "Nexo", nil, "", "")
	model.loadLoans()

	// Press 'x' to enter delete mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.mode != LoanConfirmDelete {
		t.Error("should be in confirm delete mode")
	}
}

func TestLoanModel_DeleteModeNoLoans(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	// Try to delete with no loans
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.mode != LoanList {
		t.Error("should stay in list mode when no loans")
	}
}

func TestLoanModel_DeleteConfirmCancel(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	// Add test loan
	p := model.GetPortfolio()
	p.AddLoan("USDT", 5000, "Nexo", nil, "", "")
	model.loadLoans()

	// Enter delete mode
	model.mode = LoanConfirmDelete

	// Press 'n' to cancel
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.mode != LoanList {
		t.Error("n should cancel delete and return to list")
	}
}

func TestLoanModel_FormValidation_EmptyCoin(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	model.mode = LoanAdd
	model.focusIndex = loanFieldCount - 1

	// Leave coin empty, enter amount and platform
	model.inputs[loanFieldAmount].SetValue("5000")
	model.inputs[loanFieldPlatform].SetValue("Nexo")

	// Try to submit
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.statusMsg != "Coin is required" {
		t.Error("should show coin required error")
	}
}

func TestLoanModel_FormValidation_InvalidAmount(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	model.mode = LoanAdd
	model.focusIndex = loanFieldCount - 1

	model.inputs[loanFieldCoin].SetValue("USDT")
	model.inputs[loanFieldAmount].SetValue("invalid")
	model.inputs[loanFieldPlatform].SetValue("Nexo")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.statusMsg != "Invalid amount" {
		t.Error("should show invalid amount error")
	}
}

func TestLoanModel_FormValidation_ZeroAmount(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	model.mode = LoanAdd
	model.focusIndex = loanFieldCount - 1

	model.inputs[loanFieldCoin].SetValue("USDT")
	model.inputs[loanFieldAmount].SetValue("0")
	model.inputs[loanFieldPlatform].SetValue("Nexo")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.statusMsg != "Invalid amount" {
		t.Error("should show invalid amount error for zero")
	}
}

func TestLoanModel_FormValidation_EmptyPlatform(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	model.mode = LoanAdd
	model.focusIndex = loanFieldCount - 1

	model.inputs[loanFieldCoin].SetValue("USDT")
	model.inputs[loanFieldAmount].SetValue("5000")
	model.inputs[loanFieldPlatform].SetValue("")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.statusMsg != "Platform is required for loans" {
		t.Error("should show platform required error")
	}
}

func TestLoanModel_FormValidation_InvalidInterestRate(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	model.mode = LoanAdd
	model.focusIndex = loanFieldCount - 1

	model.inputs[loanFieldCoin].SetValue("USDT")
	model.inputs[loanFieldAmount].SetValue("5000")
	model.inputs[loanFieldPlatform].SetValue("Nexo")
	model.inputs[loanFieldInterestRate].SetValue("invalid")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.statusMsg != "Invalid interest rate" {
		t.Error("should show invalid interest rate error")
	}
}

func TestLoanModel_InterestRateOptional(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	model.mode = LoanAdd
	model.focusIndex = loanFieldCount - 1

	model.inputs[loanFieldCoin].SetValue("USDT")
	model.inputs[loanFieldAmount].SetValue("5000")
	model.inputs[loanFieldPlatform].SetValue("Nexo")
	// Leave interest rate empty (optional)

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := model.Update(msg)
	model = newModel.(LoanModel)

	// Should not show error and should return command to add
	if model.statusMsg == "Invalid interest rate" {
		t.Error("empty interest rate should be valid")
	}

	if cmd == nil {
		t.Error("valid form should return add command")
	}
}

func TestLoanModel_View_EmptyList(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	view := model.View()

	if !strings.Contains(view, "No loans yet") {
		t.Error("empty list should show no loans message")
	}
}

func TestLoanModel_View_WithLoans(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	p := model.GetPortfolio()
	rate := 6.9
	p.AddLoan("USDT", 5000, "Nexo", &rate, "", "")
	model.loadLoans()

	view := model.View()

	if !strings.Contains(view, "USDT") {
		t.Error("view should show coin")
	}

	if !strings.Contains(view, "Nexo") {
		t.Error("view should show platform")
	}

	if !strings.Contains(view, "6.90%") {
		t.Error("view should show interest rate")
	}
}

func TestLoanModel_View_AddForm(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	model.mode = LoanAdd

	view := model.View()

	if !strings.Contains(view, "ADD LOAN") {
		t.Error("add form should show title")
	}

	if !strings.Contains(view, "Coin") {
		t.Error("add form should show coin label")
	}

	if !strings.Contains(view, "Platform") {
		t.Error("add form should show platform label")
	}
}

func TestLoanModel_View_DeleteConfirm(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	p := model.GetPortfolio()
	p.AddLoan("USDT", 5000, "Nexo", nil, "", "")
	model.loadLoans()

	model.mode = LoanConfirmDelete

	view := model.View()

	if !strings.Contains(view, "CONFIRM") {
		t.Error("delete confirm should show title")
	}

	if !strings.Contains(view, "USDT") {
		t.Error("delete confirm should show loan details")
	}
}

func TestLoanModel_LoanAddedMsg(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	p := model.GetPortfolio()
	loan, _ := p.AddLoan("USDT", 5000, "Nexo", nil, "", "")

	model.mode = LoanAdd

	msg := LoanAddedMsg{Loan: &loan}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.mode != LoanList {
		t.Error("should return to list mode after adding")
	}

	if !strings.Contains(model.statusMsg, "USDT") {
		t.Error("status should mention coin")
	}
}

func TestLoanModel_LoanAddedMsgError(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	model.mode = LoanAdd

	msg := LoanAddedMsg{Error: os.ErrNotExist}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if !strings.Contains(model.statusMsg, "Error") {
		t.Error("should show error message")
	}
}

func TestLoanModel_LoanDeletedMsg(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	msg := LoanDeletedMsg{ID: "test-id"}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if !strings.Contains(model.statusMsg, "removed") {
		t.Error("should show removed message")
	}
}

func TestLoanModel_BackToMenu(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("escape should return command")
	}
}

func TestLoanModel_Quit(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("q should return quit command")
	}
}

func TestLoanModel_WindowSizeMsg(t *testing.T) {
	model, cleanup := setupTestLoanModel(t)
	defer cleanup()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := model.Update(msg)
	model = newModel.(LoanModel)

	if model.width != 100 {
		t.Error("width should be updated")
	}

	if model.height != 50 {
		t.Error("height should be updated")
	}
}
