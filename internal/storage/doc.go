// Package storage provides JSON-based persistence for portfolio data.
//
// The package includes two main storage types:
//
// # Storage
//
// [Storage] handles the main portfolio data including holdings, sales, loans,
// and stakes. Data is persisted to a JSON file and loaded on initialization.
// All operations are immediately persisted to disk.
//
// # SnapshotStore
//
// [SnapshotStore] handles portfolio snapshots separately from the main data.
// Snapshots are stored in their own JSON file, sorted by timestamp (newest first).
//
// # Thread Safety
//
// Storage operations use mutex locks for thread-safe access. Multiple goroutines
// can safely read and write to the same storage instance.
//
// # File Format
//
// Data is stored as pretty-printed JSON for human readability and version control
// compatibility.
package storage
