package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/pretty-andrechal/follyo/internal/models"
)

// SnapshotStore manages snapshot persistence
type SnapshotStore struct {
	path      string
	snapshots []models.Snapshot
	mu        sync.RWMutex
}

// NewSnapshotStore creates a new SnapshotStore with the given path
func NewSnapshotStore(path string) (*SnapshotStore, error) {
	ss := &SnapshotStore{
		path:      path,
		snapshots: []models.Snapshot{},
	}

	// Ensure directory exists with restricted permissions for privacy
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	// Load existing snapshots if file exists
	if err := ss.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return ss, nil
}

// load reads snapshots from disk
func (ss *SnapshotStore) load() error {
	data, err := os.ReadFile(ss.path)
	if err != nil {
		return err
	}

	ss.mu.Lock()
	defer ss.mu.Unlock()

	if err := json.Unmarshal(data, &ss.snapshots); err != nil {
		return err
	}

	return nil
}

// save writes snapshots to disk
func (ss *SnapshotStore) save() error {
	ss.mu.RLock()
	data, err := json.MarshalIndent(ss.snapshots, "", "  ")
	ss.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(ss.path, data, 0600)
}

// Add adds a new snapshot
func (ss *SnapshotStore) Add(snapshot models.Snapshot) error {
	ss.mu.Lock()
	ss.snapshots = append(ss.snapshots, snapshot)
	ss.mu.Unlock()

	return ss.save()
}

// List returns all snapshots sorted by timestamp (newest first)
func (ss *SnapshotStore) List() []models.Snapshot {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	// Return a copy sorted by timestamp descending
	result := make([]models.Snapshot, len(ss.snapshots))
	copy(result, ss.snapshots)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})

	return result
}

// Get returns a snapshot by ID
func (ss *SnapshotStore) Get(id string) (*models.Snapshot, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	for _, s := range ss.snapshots {
		if s.ID == id {
			return &s, true
		}
	}
	return nil, false
}

// Remove removes a snapshot by ID
func (ss *SnapshotStore) Remove(id string) (bool, error) {
	ss.mu.Lock()
	found := false
	newSnapshots := make([]models.Snapshot, 0, len(ss.snapshots))
	for _, s := range ss.snapshots {
		if s.ID == id {
			found = true
		} else {
			newSnapshots = append(newSnapshots, s)
		}
	}
	ss.snapshots = newSnapshots
	ss.mu.Unlock()

	if found {
		return true, ss.save()
	}
	return false, nil
}

// Count returns the number of snapshots
func (ss *SnapshotStore) Count() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return len(ss.snapshots)
}
