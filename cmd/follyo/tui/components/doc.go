// Package components provides reusable UI components for the Follyo TUI.
//
// This package contains shared rendering functions and helpers used across
// all views, promoting consistency and reducing code duplication.
//
// # Form Components
//
// Form handling utilities for text inputs:
//   - [BuildFormInputs]: Create form inputs from field specifications
//   - [ResetFormInputs]: Reset inputs to default values
//   - [FocusField], [NextField], [PrevField]: Field navigation
//   - [BlurAll]: Remove focus from all inputs
//
// # Table Components
//
// Rendering helpers for list views and tables:
//   - [RenderTitle]: Styled section title
//   - [RenderEmptyState]: Empty list message
//   - [RenderSeparator]: Horizontal line separator
//   - [RenderStatusMessage]: Success/error messages
//   - [RenderBoxDefault], [RenderBoxError]: Bordered containers
//
// # Navigation Components
//
// Cursor and navigation helpers:
//   - [MoveCursorUp], [MoveCursorDown]: List cursor movement with bounds checking
//
// # Help Components
//
// Help text rendering:
//   - [RenderHelp]: Render key binding help from [HelpItem] list
//   - [ListHelp], [FormHelp], [DeleteConfirmHelp]: Common help presets
package components
