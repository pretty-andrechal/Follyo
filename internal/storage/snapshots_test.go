package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pretty-andrechal/follyo/internal/models"
)

func setupTestSnapshotStore(t *testing.T) (*SnapshotStore, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "follyo-snapshot-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	snapshotPath := filepath.Join(tmpDir, "snapshots.json")
	ss, err := NewSnapshotStore(snapshotPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create snapshot store: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return ss, cleanup
}

func TestNewSnapshotStore(t *testing.T) {
	ss, cleanup := setupTestSnapshotStore(t)
	defer cleanup()

	if ss == nil {
		t.Fatal("expected snapshot store to be created")
	}
}

func TestSnapshotStore_Add(t *testing.T) {
	ss, cleanup := setupTestSnapshotStore(t)
	defer cleanup()

	snapshot := models.Snapshot{
		ID:            "test1234",
		Timestamp:     time.Now(),
		HoldingsValue: 50000,
		LoansValue:    5000,
		NetValue:      45000,
		TotalInvested: 40000,
		TotalSold:     10000,
		ProfitLoss:    15000,
		ProfitPercent: 37.5,
		CoinValues: map[string]models.CoinSnapshot{
			"BTC": {Amount: 1.0, Price: 50000, Value: 50000},
		},
		Note: "Test snapshot",
	}

	err := ss.Add(snapshot)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Verify it was added
	list := ss.List()
	if len(list) != 1 {
		t.Errorf("expected 1 snapshot, got %d", len(list))
	}
	if list[0].ID != "test1234" {
		t.Errorf("expected ID test1234, got %s", list[0].ID)
	}
}

func TestSnapshotStore_List(t *testing.T) {
	ss, cleanup := setupTestSnapshotStore(t)
	defer cleanup()

	// Initially empty
	list := ss.List()
	if len(list) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(list))
	}

	// Add multiple snapshots
	now := time.Now()
	ss.Add(models.Snapshot{ID: "snap1", Timestamp: now.Add(-2 * time.Hour)})
	ss.Add(models.Snapshot{ID: "snap2", Timestamp: now.Add(-1 * time.Hour)})
	ss.Add(models.Snapshot{ID: "snap3", Timestamp: now})

	list = ss.List()
	if len(list) != 3 {
		t.Errorf("expected 3 snapshots, got %d", len(list))
	}

	// Verify sorted by timestamp (newest first)
	if list[0].ID != "snap3" {
		t.Errorf("expected first snapshot to be snap3, got %s", list[0].ID)
	}
	if list[2].ID != "snap1" {
		t.Errorf("expected last snapshot to be snap1, got %s", list[2].ID)
	}
}

func TestSnapshotStore_Get(t *testing.T) {
	ss, cleanup := setupTestSnapshotStore(t)
	defer cleanup()

	snapshot := models.Snapshot{
		ID:        "findme",
		Timestamp: time.Now(),
		NetValue:  12345,
		Note:      "Found it",
	}
	ss.Add(snapshot)

	// Find existing snapshot
	found, ok := ss.Get("findme")
	if !ok {
		t.Fatal("expected to find snapshot")
	}
	if found.NetValue != 12345 {
		t.Errorf("expected NetValue 12345, got %f", found.NetValue)
	}
	if found.Note != "Found it" {
		t.Errorf("expected Note 'Found it', got %s", found.Note)
	}

	// Try to find non-existent snapshot
	_, ok = ss.Get("nonexistent")
	if ok {
		t.Error("expected not to find non-existent snapshot")
	}
}

func TestSnapshotStore_Remove(t *testing.T) {
	ss, cleanup := setupTestSnapshotStore(t)
	defer cleanup()

	ss.Add(models.Snapshot{ID: "remove1", Timestamp: time.Now()})
	ss.Add(models.Snapshot{ID: "remove2", Timestamp: time.Now()})

	if ss.Count() != 2 {
		t.Errorf("expected 2 snapshots, got %d", ss.Count())
	}

	// Remove existing snapshot
	removed, err := ss.Remove("remove1")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	if !removed {
		t.Error("expected snapshot to be removed")
	}

	if ss.Count() != 1 {
		t.Errorf("expected 1 snapshot after removal, got %d", ss.Count())
	}

	// Try to remove non-existent snapshot
	removed, err = ss.Remove("nonexistent")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	if removed {
		t.Error("expected snapshot not to be removed")
	}
}

func TestSnapshotStore_Count(t *testing.T) {
	ss, cleanup := setupTestSnapshotStore(t)
	defer cleanup()

	if ss.Count() != 0 {
		t.Errorf("expected count 0, got %d", ss.Count())
	}

	ss.Add(models.Snapshot{ID: "c1", Timestamp: time.Now()})
	if ss.Count() != 1 {
		t.Errorf("expected count 1, got %d", ss.Count())
	}

	ss.Add(models.Snapshot{ID: "c2", Timestamp: time.Now()})
	if ss.Count() != 2 {
		t.Errorf("expected count 2, got %d", ss.Count())
	}

	ss.Remove("c1")
	if ss.Count() != 1 {
		t.Errorf("expected count 1 after removal, got %d", ss.Count())
	}
}

func TestSnapshotStore_Persistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "follyo-snapshot-persist-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	snapshotPath := filepath.Join(tmpDir, "snapshots.json")

	// Create store and add snapshot
	ss1, err := NewSnapshotStore(snapshotPath)
	if err != nil {
		t.Fatalf("failed to create snapshot store: %v", err)
	}

	ss1.Add(models.Snapshot{
		ID:        "persist",
		Timestamp: time.Now(),
		NetValue:  99999,
		Note:      "Persisted",
	})

	// Create new store instance and verify data persisted
	ss2, err := NewSnapshotStore(snapshotPath)
	if err != nil {
		t.Fatalf("failed to create second snapshot store: %v", err)
	}

	if ss2.Count() != 1 {
		t.Errorf("expected 1 snapshot after reload, got %d", ss2.Count())
	}

	found, ok := ss2.Get("persist")
	if !ok {
		t.Fatal("expected to find persisted snapshot")
	}
	if found.NetValue != 99999 {
		t.Errorf("expected NetValue 99999, got %f", found.NetValue)
	}
}
