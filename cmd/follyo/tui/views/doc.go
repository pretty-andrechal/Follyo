// Package views contains all TUI view models for the Follyo application.
//
// Each view follows the Bubble Tea Model interface pattern with Init, Update,
// and View methods. Views are designed to be independent and communicate with
// the main application through messages.
//
// # Available Views
//
//   - [MenuModel]: Main navigation menu
//   - [SummaryModel]: Portfolio summary with values and allocation
//   - [BuyModel]: Manage holdings/purchases
//   - [SellModel]: Manage sales
//   - [StakeModel]: Manage staking positions
//   - [LoanModel]: Manage loans
//   - [SnapshotsModel]: View and create portfolio snapshots
//   - [TickerModel]: Manage coin ticker mappings
//   - [SettingsModel]: Application settings
//
// # View Modes
//
// Most views support multiple modes:
//   - List: Display items with cursor navigation
//   - Add: Form for adding new items
//   - ConfirmDelete: Deletion confirmation dialog
//
// # Messages
//
// Views emit messages to communicate with the parent application:
//   - *AddedMsg: Item was added successfully
//   - *DeletedMsg: Item was deleted
//   - BackToMenuMsg: User requested return to menu
package views
