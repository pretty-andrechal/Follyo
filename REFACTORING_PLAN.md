# Follyo Codebase Refactoring Plan

This document outlines a phased approach to improving the maintainability, testability, and organization of the Follyo codebase.

## Overview

The refactoring is divided into 6 phases, each building upon the previous. Each phase includes:
- Implementation changes
- Test updates/additions
- Documentation updates

**Estimated total effort:** 4-6 development sessions

---

## Phase 1: Constants and Configuration (Quick Wins)

**Goal:** Extract magic numbers and hardcoded values into named constants.

### 1.1 Create TUI Constants File

**New file:** `cmd/follyo/tui/constants.go`

```go
package tui

// Input field constraints
const (
    InputCoinCharLimit     = 10
    InputCoinWidth         = 20
    InputAmountCharLimit   = 20
    InputAmountWidth       = 20
    InputPriceCharLimit    = 20
    InputPriceWidth        = 20
    InputPlatformCharLimit = 50
    InputPlatformWidth     = 30
    InputNotesCharLimit    = 200
    InputNotesWidth        = 40
    InputSearchCharLimit   = 50
    InputSearchWidth       = 40
)

// Display constraints
const (
    DateDisplayLength        = 10  // YYYY-MM-DD
    PlatformDisplayMaxLength = 12
    PlatformTruncateLength   = 9
    TruncationSuffix         = "..."
)

// Table/list widths
const (
    BuyListSeparatorWidth      = 85
    SellListSeparatorWidth     = 85
    StakeListSeparatorWidth    = 75
    LoanListSeparatorWidth     = 75
    TickerListSeparatorWidth   = 50
    SnapshotListSeparatorWidth = 90
    SearchResultsSeparatorWidth = 70
)

// Pagination
const (
    DefaultMappingsVisibleCount = 15
)
```

### 1.2 Update View Files to Use Constants

**Files to update:**
- `cmd/follyo/tui/views/buy.go`
- `cmd/follyo/tui/views/sell.go`
- `cmd/follyo/tui/views/stake.go`
- `cmd/follyo/tui/views/loan.go`
- `cmd/follyo/tui/views/ticker.go`
- `cmd/follyo/tui/views/snapshots.go`

**Changes:**
- Replace hardcoded `CharLimit` values with constants
- Replace hardcoded `Width` values with constants
- Replace hardcoded separator lengths with constants
- Replace hardcoded truncation lengths with constants

### 1.3 Tests

**New file:** `cmd/follyo/tui/constants_test.go`
- Verify constants are positive/valid values
- Ensure truncation length < max length

### 1.4 Documentation

- Add comments to constants explaining their purpose
- Update any relevant README sections

### Deliverables
- [x] `cmd/follyo/tui/constants.go` created
- [x] All view files updated to use constants
- [x] `cmd/follyo/tui/constants_test.go` created
- [x] All existing tests pass

**Status: ✅ COMPLETED**

---

## Phase 2: Shared Formatting Utilities

**Goal:** Consolidate duplicate formatting functions into a shared package.

### 2.1 Create Format Package

**New file:** `cmd/follyo/tui/format/format.go`

```go
package format

// USD formats a float64 as USD currency with thousand separators
func USD(amount float64) string

// USDSimple formats a float64 as USD without thousand separators (for alignment)
func USDSimple(amount float64) string

// ProfitLoss formats a profit/loss amount with +/- prefix
func ProfitLoss(amount float64) string

// Percentage formats a float64 as a percentage
func Percentage(value float64) string

// TruncateString truncates a string to maxLen, adding suffix if truncated
func TruncateString(s string, maxLen int, suffix string) string

// TruncatePlatform truncates platform names for display
func TruncatePlatform(platform string) string

// TruncateDate truncates date to YYYY-MM-DD format
func TruncateDate(date string) string

// FormatAmount formats a crypto amount (handles small decimals)
func FormatAmount(amount float64) string
```

### 2.2 Create Error Formatting

**New file:** `cmd/follyo/tui/format/errors.go`

```go
package format

// StatusError formats an error for status bar display
func StatusError(context string, err error) string

// StatusSuccess formats a success message for status bar display
func StatusSuccess(action string, details string) string
```

### 2.3 Update Consumers

**Files to update:**
- `cmd/follyo/tui/views/summary.go` - Remove local `formatUSD`, use `format.USD`
- `cmd/follyo/tui/views/buy.go` - Use format functions
- `cmd/follyo/tui/views/sell.go` - Use format functions
- `cmd/follyo/tui/views/stake.go` - Use format functions
- `cmd/follyo/tui/views/loan.go` - Use format functions
- `cmd/follyo/tui/views/snapshots.go` - Use format functions
- `cmd/follyo/tui/views/ticker.go` - Use format functions
- `cmd/follyo/helpers.go` - Keep CLI version but consider consolidation

### 2.4 Tests

**New file:** `cmd/follyo/tui/format/format_test.go`

```go
func TestUSD(t *testing.T) {
    tests := []struct {
        input    float64
        expected string
    }{
        {1000.50, "$1,000.50"},
        {-500.25, "-$500.25"},
        {0.00, "$0.00"},
        {1000000.99, "$1,000,000.99"},
    }
    // ...
}

func TestTruncateString(t *testing.T) { ... }
func TestTruncatePlatform(t *testing.T) { ... }
func TestTruncateDate(t *testing.T) { ... }
func TestStatusError(t *testing.T) { ... }
```

### 2.5 Documentation

- Package-level godoc for `format` package
- Document each public function

### Deliverables
- [x] `cmd/follyo/tui/format/format.go` created
- [x] `cmd/follyo/tui/format/errors.go` created
- [x] All view files updated to use format package
- [x] `cmd/follyo/tui/format/format_test.go` with comprehensive tests
- [x] `cmd/follyo/tui/format/errors_test.go` with tests for error utilities
- [x] All existing tests pass

**Status: ✅ COMPLETED**

---

## Phase 3: Shared View Components

**Goal:** Extract common view patterns into reusable components.

### 3.1 Create Form Builder

**New file:** `cmd/follyo/tui/components/form.go`

```go
package components

import "github.com/charmbracelet/bubbles/textinput"

// FormFieldSpec defines a form field configuration
type FormFieldSpec struct {
    Placeholder  string
    CharLimit    int
    Width        int
    DefaultValue string
}

// BuildFormInputs creates textinput models from field specifications
func BuildFormInputs(fields []FormFieldSpec) []textinput.Model

// ResetFormInputs clears all form inputs, optionally setting defaults
func ResetFormInputs(inputs []textinput.Model, defaults []string)

// FocusField sets focus to a specific field, blurring others
func FocusField(inputs []textinput.Model, index int) tea.Cmd
```

### 3.2 Create Table Row Renderer

**New file:** `cmd/follyo/tui/components/table.go`

```go
package components

// RowRenderer handles common table row rendering patterns
type RowRenderer struct {
    CursorStyle    lipgloss.Style
    SelectedStyle  lipgloss.Style
    NormalStyle    lipgloss.Style
}

// NewRowRenderer creates a row renderer with default TUI styles
func NewRowRenderer() *RowRenderer

// RenderRow renders a table row with cursor and selection state
func (r *RowRenderer) RenderRow(index, cursor int, content string) string

// RenderHeader renders a table header
func RenderHeader(columns []string, widths []int) string

// RenderSeparator renders a horizontal separator line
func RenderSeparator(width int) string
```

### 3.3 Create List Navigation Helper

**New file:** `cmd/follyo/tui/components/navigation.go`

```go
package components

// ListNavigator handles cursor movement with bounds checking
type ListNavigator struct {
    Cursor int
    Length int
}

// MoveUp moves cursor up, respecting bounds
func (n *ListNavigator) MoveUp() bool

// MoveDown moves cursor down, respecting bounds
func (n *ListNavigator) MoveDown() bool

// ClampCursor ensures cursor is within valid range after list changes
func (n *ListNavigator) ClampCursor()
```

### 3.4 Update View Files

**Files to update:**
- `cmd/follyo/tui/views/buy.go` - Use components
- `cmd/follyo/tui/views/sell.go` - Use components
- `cmd/follyo/tui/views/stake.go` - Use components
- `cmd/follyo/tui/views/loan.go` - Use components
- `cmd/follyo/tui/views/ticker.go` - Use components

**Expected reduction:** ~30-50 lines per view file

### 3.5 Tests

**New files:**
- `cmd/follyo/tui/components/form_test.go`
- `cmd/follyo/tui/components/table_test.go`
- `cmd/follyo/tui/components/navigation_test.go`

### 3.6 Documentation

- Package-level godoc for `components` package
- Usage examples in comments

### Deliverables
- [x] `cmd/follyo/tui/components/form.go` created
- [x] `cmd/follyo/tui/components/table.go` created
- [x] `cmd/follyo/tui/components/navigation.go` created
- [x] `cmd/follyo/tui/components/help.go` created (bonus)
- [x] `buy.go` refactored to use components (example implementation)
- [x] Component test files created (form_test.go, table_test.go, navigation_test.go, help_test.go)
- [x] All existing tests pass

**Status: ✅ COMPLETED**

Note: Components are now available for gradual adoption in other view files. The buy.go file demonstrates the usage pattern.

---

## Phase 4: Interfaces and Dependency Injection

**Goal:** Reduce tight coupling by introducing interfaces and improving DI.

### 4.1 Define Portfolio Interfaces

**New file:** `internal/portfolio/interfaces.go`

```go
package portfolio

import "github.com/pretty-andrechal/follyo/internal/models"

// HoldingService defines operations for managing purchases
type HoldingService interface {
    AddHolding(coin string, amount, price float64, platform, notes, date string) (models.Holding, error)
    RemoveHolding(id string) (bool, error)
    ListHoldings() ([]models.Holding, error)
}

// SaleService defines operations for managing sales
type SaleService interface {
    AddSale(coin string, amount, price float64, platform, notes, date string) (models.Sale, error)
    RemoveSale(id string) (bool, error)
    ListSales() ([]models.Sale, error)
}

// StakeService defines operations for managing stakes
type StakeService interface {
    AddStake(coin string, amount float64, platform string, apy *float64, notes, date string) (models.Stake, error)
    RemoveStake(id string) (bool, error)
    ListStakes() ([]models.Stake, error)
}

// LoanService defines operations for managing loans
type LoanService interface {
    AddLoan(coin string, amount float64, platform string, interestRate *float64, notes, date string) (models.Loan, error)
    RemoveLoan(id string) (bool, error)
    ListLoans() ([]models.Loan, error)
}

// SummaryService defines operations for portfolio summaries
type SummaryService interface {
    GetHoldingsByCoin() (map[string]float64, error)
    GetSalesByCoin() (map[string]float64, error)
    GetStakesByCoin() (map[string]float64, error)
    GetLoansByCoin() (map[string]float64, error)
    GetTotalInvested() (float64, error)
    GetTotalSold() (float64, error)
}
```

### 4.2 Ensure Portfolio Implements Interfaces

**File to update:** `internal/portfolio/portfolio.go`

Add compile-time interface checks:
```go
var _ HoldingService = (*Portfolio)(nil)
var _ SaleService = (*Portfolio)(nil)
var _ StakeService = (*Portfolio)(nil)
var _ LoanService = (*Portfolio)(nil)
var _ SummaryService = (*Portfolio)(nil)
```

### 4.3 Update View Models to Use Interfaces

**Files to update:**
- `cmd/follyo/tui/views/buy.go` - Use `HoldingService` instead of `*Portfolio`
- `cmd/follyo/tui/views/sell.go` - Use `SaleService` instead of `*Portfolio`
- `cmd/follyo/tui/views/stake.go` - Use `StakeService` instead of `*Portfolio`
- `cmd/follyo/tui/views/loan.go` - Use `LoanService` instead of `*Portfolio`
- `cmd/follyo/tui/views/summary.go` - Use `SummaryService` instead of `*Portfolio`

### 4.4 Create Mock Implementations for Testing

**New file:** `internal/portfolio/mocks/mocks.go`

```go
package mocks

// MockHoldingService is a test double for HoldingService
type MockHoldingService struct {
    Holdings     []models.Holding
    AddError     error
    RemoveError  error
    ListError    error
}

func (m *MockHoldingService) AddHolding(...) (models.Holding, error) { ... }
func (m *MockHoldingService) RemoveHolding(id string) (bool, error) { ... }
func (m *MockHoldingService) ListHoldings() ([]models.Holding, error) { ... }

// Similar mocks for SaleService, StakeService, LoanService
```

### 4.5 Update View Tests to Use Mocks

**Files to update:**
- `cmd/follyo/tui/views/buy_test.go` - Use mock instead of real portfolio
- `cmd/follyo/tui/views/sell_test.go`
- `cmd/follyo/tui/views/stake_test.go`
- `cmd/follyo/tui/views/loan_test.go`

### 4.6 Tests

**New file:** `internal/portfolio/interfaces_test.go`
- Verify Portfolio implements all interfaces

### 4.7 Documentation

- Document interfaces with usage examples
- Update view model documentation to reference interfaces

### Deliverables
- [ ] `internal/portfolio/interfaces.go` created
- [ ] Portfolio implements all interfaces (with compile-time checks)
- [ ] View models updated to use interfaces
- [ ] `internal/portfolio/mocks/mocks.go` created
- [ ] View tests updated to use mocks
- [ ] All existing tests pass

---

## Phase 5: App.go Refactoring

**Goal:** Simplify app.go by reducing repetitive view handler methods.

### 5.1 Create View Registry

**New file:** `cmd/follyo/tui/registry.go`

```go
package tui

import tea "github.com/charmbracelet/bubbletea"

// ViewRegistry manages view models and their lifecycle
type ViewRegistry struct {
    views map[ViewType]tea.Model
}

// NewViewRegistry creates a new view registry
func NewViewRegistry() *ViewRegistry

// Register adds a view model to the registry
func (r *ViewRegistry) Register(vt ViewType, model tea.Model)

// Get retrieves a view model by type
func (r *ViewRegistry) Get(vt ViewType) tea.Model

// Update forwards a message to the specified view
func (r *ViewRegistry) Update(vt ViewType, msg tea.Msg) (tea.Model, tea.Cmd)

// View renders the specified view
func (r *ViewRegistry) View(vt ViewType) string
```

### 5.2 Refactor App.go

**File to update:** `cmd/follyo/tui/app.go`

**Before:** 8 separate update methods + 8 view render cases
**After:** Single registry-based dispatch

```go
type App struct {
    registry    *ViewRegistry
    currentView ViewType
    // ... other fields
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // ... global handling ...

    // Delegate to current view via registry
    model, cmd := a.registry.Update(a.currentView, msg)
    if model != nil {
        a.registry.Register(a.currentView, model)
    }
    return a, cmd
}
```

### 5.3 Consolidate Menu Action Handling

Create a menu action dispatcher to reduce switch statement size:

```go
type menuAction struct {
    view     ViewType
    initFunc func() tea.Cmd
}

var menuActions = map[string]menuAction{
    "summary":   {ViewSummary, nil},
    "buy":       {ViewBuy, nil},
    "sell":      {ViewSell, nil},
    // ...
}

func (a *App) handleMenuSelect(msg MenuSelectMsg) (tea.Model, tea.Cmd) {
    action, ok := menuActions[msg.Action]
    if !ok {
        return a, nil
    }
    a.currentView = action.view
    // ...
}
```

### 5.4 Tests

**File to update:** `cmd/follyo/tui/app_test.go`
- Update tests to work with registry
- Add registry-specific tests

**New file:** `cmd/follyo/tui/registry_test.go`

### 5.5 Documentation

- Document ViewRegistry usage
- Update app.go comments

### Deliverables
- [ ] `cmd/follyo/tui/registry.go` created
- [ ] `app.go` refactored to use registry
- [ ] Menu action handling consolidated
- [ ] `cmd/follyo/tui/registry_test.go` created
- [ ] `app_test.go` updated
- [ ] All existing tests pass

---

## Phase 6: Test Coverage and Documentation

**Goal:** Improve test coverage and add comprehensive documentation.

### 6.1 Add Missing View Tests

**Files to enhance:**
- `cmd/follyo/tui/views/buy_test.go`
- `cmd/follyo/tui/views/sell_test.go`
- `cmd/follyo/tui/views/stake_test.go`
- `cmd/follyo/tui/views/loan_test.go`
- `cmd/follyo/tui/views/ticker_test.go`

**Test scenarios to add:**
- Form validation (empty fields, invalid input)
- State transitions (List → Add → List, etc.)
- Error message formatting
- Boundary conditions (empty lists, cursor bounds)
- Message handling for all message types

### 6.2 Add CLI Integration Tests

**New file:** `cmd/follyo/integration_test.go`

```go
func TestBuyAddCommand_Success(t *testing.T) { ... }
func TestBuyAddCommand_InvalidAmount(t *testing.T) { ... }
func TestSellAddCommand_Success(t *testing.T) { ... }
func TestSummaryCommand_WithPrices(t *testing.T) { ... }
func TestSummaryCommand_NoPrices(t *testing.T) { ... }
```

### 6.3 Add Helper Function Tests

**New file:** `cmd/follyo/helpers_test.go`

```go
func TestAddCommas(t *testing.T) { ... }
func TestParsePriceFromArgs(t *testing.T) { ... }
func TestCollectAllCoins(t *testing.T) { ... }
func TestFormatUSD(t *testing.T) { ... }
```

### 6.4 Package Documentation

**Files to add/update:**

`cmd/follyo/tui/doc.go`:
```go
// Package tui provides an interactive terminal user interface for Follyo.
//
// Architecture:
// The TUI follows the Bubble Tea Model-View-Update pattern...
//
// Views:
// - MenuModel: Main menu navigation
// - SummaryModel: Portfolio overview with live prices
// - BuyModel, SellModel, StakeModel, LoanModel: CRUD views for entries
// - SnapshotsModel: Historical portfolio snapshots
// - TickerModel: CoinGecko ticker mappings
// - SettingsModel: User preferences
//
// Usage:
//   app := tui.NewApp(storage, portfolio)
//   program := tea.NewProgram(app)
//   program.Run()
package tui
```

`cmd/follyo/tui/views/doc.go`:
```go
// Package views provides individual view models for the Follyo TUI.
//
// View Lifecycle:
// Each view follows a three-mode pattern:
// 1. List mode: Browse and select items
// 2. Add mode: Create new entries via form
// 3. ConfirmDelete mode: Confirm item deletion
//
// Key Bindings:
// - ↑/↓ or j/k: Navigate
// - a/n: Add new item
// - d/x: Delete selected item
// - Enter: Select/confirm
// - Esc: Cancel/go back
// - q: Quit application
package views
```

### 6.5 Update README

Add sections for:
- Architecture overview
- Development guide
- Testing instructions
- Contributing guidelines

### 6.6 Generate Test Coverage Report

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

Target: >80% coverage for core packages

### Deliverables
- [ ] View test files enhanced with comprehensive scenarios
- [ ] `cmd/follyo/integration_test.go` created
- [ ] `cmd/follyo/helpers_test.go` created
- [ ] Package documentation files created
- [ ] README updated with development guide
- [ ] Test coverage >80% for core packages
- [ ] All tests pass

---

## Summary

| Phase | Focus | New Files | Modified Files | Effort |
|-------|-------|-----------|----------------|--------|
| 1 | Constants | 2 | 6 | Small |
| 2 | Formatting | 3 | 8 | Small |
| 3 | Components | 4 | 5 | Medium |
| 4 | Interfaces | 3 | 10 | Medium |
| 5 | App Refactor | 2 | 2 | Medium |
| 6 | Tests & Docs | 5 | 8 | Medium |

**Total new files:** ~19
**Total modified files:** ~25 (with overlap)

---

## Success Criteria

After completing all phases:

1. **No duplicate code** - Common patterns extracted to shared packages
2. **Testable** - All views can be tested with mocks
3. **Documented** - All packages have godoc, README is comprehensive
4. **Test coverage** - >80% coverage on core packages
5. **Maintainable** - Adding a new view requires <100 lines of unique code
6. **Consistent** - All magic numbers are named constants

---

## Getting Started

Begin with Phase 1 by running:

```bash
# Create the constants file
touch cmd/follyo/tui/constants.go

# Run tests to establish baseline
go test ./... -v
```

After each phase, verify:
```bash
go build ./...
go test ./...
go vet ./...
```
