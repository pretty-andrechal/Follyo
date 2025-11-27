# TUI Implementation Plan

## Overview
Add an interactive Terminal User Interface (TUI) to Follyo using Bubble Tea, accessible via `follyo tui` command. The existing CLI commands remain fully functional.

---

## Phase 1: Foundation & Setup

### 1.1 Add Dependencies
```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/bubbles
go get github.com/charmbracelet/lipgloss
```

- **bubbletea**: Core TUI framework (Elm architecture)
- **bubbles**: Pre-built components (list, textinput, table, spinner, etc.)
- **lipgloss**: Styling and layout

### 1.2 Create TUI Package Structure
```
cmd/follyo/
├── tui.go                 # Main TUI command entry point
├── tui/
│   ├── app.go             # Main application model
│   ├── styles.go          # Shared styles (colors, borders)
│   ├── keys.go            # Keybindings
│   ├── views/
│   │   ├── menu.go        # Main menu view
│   │   ├── summary.go     # Portfolio summary view
│   │   ├── list.go        # Generic list view (holdings, sales, etc.)
│   │   └── form.go        # Generic form view (add operations)
│   └── components/
│       ├── table.go       # Reusable table component
│       ├── statusbar.go   # Bottom status bar
│       └── confirm.go     # Confirmation dialog
```

### 1.3 Architecture Pattern
Use Bubble Tea's Elm architecture:
- **Model**: Application state
- **Update**: Handle messages/events
- **View**: Render UI

```go
// Main app model holds current view and shared state
type App struct {
    portfolio    *portfolio.Portfolio
    config       *config.ConfigStore
    currentView  View
    width        int
    height       int
    err          error
}

type View interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (View, tea.Cmd)
    View() string
}
```

---

## Phase 2: Core Views Implementation

### 2.1 Main Menu View
**File:** `tui/views/menu.go`

```
╭─────────────────────────────────────╮
│         FOLLYO - Portfolio          │
│                                     │
│   → Portfolio Summary               │
│     Buy                             │
│     Sell                            │
│     Stake                           │
│     Loan                            │
│     Snapshots                       │
│     Settings                        │
│                                     │
│   Press q to quit, ? for help       │
╰─────────────────────────────────────╯
```

**Features:**
- Arrow keys to navigate
- Enter to select
- Highlight current selection
- Show keyboard shortcuts

**Implementation:**
- Use `bubbles/list` component
- Custom item delegate for styling

### 2.2 Portfolio Summary View
**File:** `tui/views/summary.go`

```
╭─ Portfolio Summary ─────────────────────────────────────────╮
│                                                             │
│  HOLDINGS BY COIN              STAKED BY COIN               │
│  ┌────────────────────────┐    ┌────────────────────────┐   │
│  │ BTC    1.5000 @ $97,000│    │ ETH   5.0000 @ $3,400  │   │
│  │ ETH   10.0000 @ $3,400 │    └────────────────────────┘   │
│  │ SOL  100.0000 @ $180   │                                 │
│  └────────────────────────┘    LOANS BY COIN                │
│                                ┌────────────────────────┐   │
│  Net Value: $215,450.00        │ USDC  5,000 on Nexo    │   │
│  Profit/Loss: +$15,200 (7.6%)  └────────────────────────┘   │
│                                                             │
│  Last updated: 2025-01-15 10:30:00  [R]efresh  [B]ack       │
╰─────────────────────────────────────────────────────────────╯
```

**Features:**
- Display all portfolio sections
- Live price integration
- Auto-refresh option
- Color-coded profit/loss

**Implementation:**
- Reuse existing `portfolio.GetSummary()`
- Use `lipgloss` for layout (columns, borders)
- Spinner while fetching prices

### 2.3 List View (Generic)
**File:** `tui/views/list.go`

Used for: Buy list, Sell list, Stake list, Loan list

```
╭─ Purchases ─────────────────────────────────────────────────╮
│                                                             │
│  ID          Coin    Amount      Price      Total    Date   │
│  ─────────────────────────────────────────────────────────  │
│  c7a482a3    BTC     1.0000   $50,000   $50,000  2024-01-15 │
│→ 28b7bd8a    BTC     0.5000   $55,000   $27,500  2024-02-01 │
│  f9e12c34    ETH    10.0000    $3,000   $30,000  2024-02-15 │
│                                                             │
│  [A]dd  [D]elete  [B]ack                    Page 1/1        │
╰─────────────────────────────────────────────────────────────╯
```

**Features:**
- Scrollable list with selection
- Delete selected item (with confirmation)
- Navigate to Add form
- Pagination for large lists

**Implementation:**
- Use `bubbles/table` component
- Custom row styling for selection

### 2.4 Form View (Generic)
**File:** `tui/views/form.go`

Used for: Add purchase, Add sale, Add stake, Add loan

```
╭─ Add Purchase ──────────────────────────────────────────────╮
│                                                             │
│  Coin:      [BTC________]                                   │
│  Amount:    [1.5________]                                   │
│  Price:     [97000______]    ○ Per unit  ● Total            │
│  Platform:  [Coinbase___]    (optional)                     │
│  Date:      [2025-01-15_]    (optional, default: today)     │
│  Notes:     [DCA buy____]    (optional)                     │
│                                                             │
│           [ Submit ]    [ Cancel ]                          │
│                                                             │
│  Tab: next field  Shift+Tab: prev  Enter: submit            │
╰─────────────────────────────────────────────────────────────╯
```

**Features:**
- Tab between fields
- Input validation with error messages
- Radio buttons for price type (per unit vs total)
- Auto-complete for coin symbols (optional, phase 4)
- Success/error feedback

**Implementation:**
- Use `bubbles/textinput` for each field
- Custom focus management
- Validate on submit using existing `models.Validate*` functions

---

## Phase 3: Sub-menu Views

### 3.1 Buy Menu
```
╭─ Buy ───────────────────────────────╮
│                                     │
│   → Add Purchase                    │
│     List Purchases                  │
│                                     │
│   Press Esc to go back              │
╰─────────────────────────────────────╯
```

### 3.2 Sell Menu
Same structure as Buy

### 3.3 Stake Menu
Same structure, form has APY field instead of price

### 3.4 Loan Menu
Same structure, form has Interest Rate field, platform required

### 3.5 Snapshots Menu
```
╭─ Snapshots ─────────────────────────╮
│                                     │
│   → Save Snapshot                   │
│     List Snapshots                  │
│     Compare Snapshots               │
│                                     │
╰─────────────────────────────────────╯
```

### 3.6 Settings Menu
```
╭─ Settings ──────────────────────────╮
│                                     │
│   → Ticker Mappings                 │
│     Preferences                     │
│                                     │
╰─────────────────────────────────────╯
```

---

## Phase 4: Components

### 4.1 Status Bar
**File:** `tui/components/statusbar.go`

Always visible at bottom:
```
 FOLLYO v1.0.0 │ Portfolio: $215,450 │ ↑↓ Navigate │ Enter Select │ q Quit
```

### 4.2 Confirmation Dialog
**File:** `tui/components/confirm.go`

```
╭─ Confirm ───────────────────────────╮
│                                     │
│  Delete purchase c7a482a3?          │
│  BTC 1.0000 @ $50,000               │
│                                     │
│       [ Yes ]    [ No ]             │
╰─────────────────────────────────────╯
```

### 4.3 Notification/Toast
Brief messages that auto-dismiss:
```
✓ Purchase added successfully
```

---

## Phase 5: Keybindings

### Global Keys
| Key | Action |
|-----|--------|
| `q`, `Ctrl+C` | Quit |
| `Esc` | Go back / Cancel |
| `?` | Show help |
| `r` | Refresh data |

### Navigation Keys
| Key | Action |
|-----|--------|
| `↑`, `k` | Move up |
| `↓`, `j` | Move down |
| `Enter` | Select / Submit |
| `Tab` | Next field (in forms) |
| `Shift+Tab` | Previous field |

### Action Keys
| Key | Action |
|-----|--------|
| `a` | Add (in list views) |
| `d`, `Delete` | Delete selected |
| `e` | Edit (future) |

---

## Phase 6: Styling

### Color Palette (lipgloss)
```go
var (
    primaryColor   = lipgloss.Color("#7C3AED")  // Purple
    successColor   = lipgloss.Color("#10B981")  // Green
    errorColor     = lipgloss.Color("#EF4444")  // Red
    warningColor   = lipgloss.Color("#F59E0B")  // Yellow
    mutedColor     = lipgloss.Color("#6B7280")  // Gray
    backgroundColor = lipgloss.Color("#1F2937") // Dark gray
)
```

### Consistent Styling
- Rounded borders for panels
- Highlighted selection with background color
- Dimmed inactive elements
- Bold headers

---

## Implementation Order

### Sprint 1: Foundation (2-3 files)
1. [ ] Add dependencies to go.mod
2. [ ] Create `tui.go` command entry point
3. [ ] Create `tui/app.go` with basic model
4. [ ] Create `tui/styles.go` with color palette
5. [ ] Create `tui/keys.go` with keybindings
6. [ ] Implement main menu view

**Deliverable:** `follyo tui` shows main menu, can quit with `q`

### Sprint 2: Summary View (1-2 files)
1. [ ] Create `tui/views/summary.go`
2. [ ] Integrate with `portfolio.GetSummary()`
3. [ ] Add price fetching with spinner
4. [ ] Style with lipgloss

**Deliverable:** Can view portfolio summary from TUI

### Sprint 3: List View (1-2 files)
1. [ ] Create `tui/views/list.go` (generic)
2. [ ] Implement for holdings/purchases
3. [ ] Add delete with confirmation
4. [ ] Add pagination

**Deliverable:** Can list and delete purchases

### Sprint 4: Form View (1-2 files)
1. [ ] Create `tui/views/form.go` (generic)
2. [ ] Implement for add purchase
3. [ ] Add validation feedback
4. [ ] Tab navigation between fields

**Deliverable:** Can add purchases via TUI form

### Sprint 5: Extend to All Operations
1. [ ] Sell submenu + list + form
2. [ ] Stake submenu + list + form
3. [ ] Loan submenu + list + form

**Deliverable:** All CRUD operations work in TUI

### Sprint 6: Snapshots & Settings
1. [ ] Snapshots submenu
2. [ ] Settings submenu
3. [ ] Ticker mapping management

**Deliverable:** Full feature parity with CLI

### Sprint 7: Polish
1. [ ] Status bar component
2. [ ] Help overlay (`?` key)
3. [ ] Error handling improvements
4. [ ] Keyboard shortcut hints

**Deliverable:** Production-ready TUI

---

## Testing Strategy

1. **Unit tests** for view state transitions
2. **Integration tests** using bubbletea's test utilities
3. **Manual testing** checklist for each view

---

## Files to Create (Summary)

```
cmd/follyo/
├── tui.go                      # NEW - Command entry
└── tui/
    ├── app.go                  # NEW - Main model
    ├── styles.go               # NEW - Lipgloss styles
    ├── keys.go                 # NEW - Keybindings
    ├── views/
    │   ├── menu.go             # NEW - Main menu
    │   ├── submenu.go          # NEW - Sub-menus (buy, sell, etc.)
    │   ├── summary.go          # NEW - Portfolio summary
    │   ├── list.go             # NEW - Generic list view
    │   └── form.go             # NEW - Generic form view
    └── components/
        ├── statusbar.go        # NEW - Bottom bar
        └── confirm.go          # NEW - Confirmation dialog
```

**Total: ~12 new files**

---

## Design Decisions

1. **Live refresh**: Yes - prices auto-refresh periodically in TUI
2. **Color scheme**: Dark theme
3. **Mouse support**: Yes - full mouse support enabled
