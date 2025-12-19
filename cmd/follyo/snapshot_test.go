package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitSnapshotStore_Success(t *testing.T) {
	// Set up temp directory
	tmpDir, err := os.MkdirTemp("", "follyo-snapshot-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create data directory
	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}

	// Change to temp dir so initSnapshotStore finds "data" directory
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Reset snapshotStore to force re-initialization
	oldStore := snapshotStore
	snapshotStore = nil
	defer func() { snapshotStore = oldStore }()

	// Call initSnapshotStore
	store := initSnapshotStore()

	if store == nil {
		t.Error("initSnapshotStore returned nil")
	}

	// Call again to test caching
	store2 := initSnapshotStore()
	if store2 != store {
		t.Error("initSnapshotStore should return cached instance")
	}
}

func TestSnapshotCmd_Exists(t *testing.T) {
	// Verify the snapshot command structure
	if snapshotCmd.Use != "snapshot" {
		t.Errorf("expected Use 'snapshot', got %s", snapshotCmd.Use)
	}

	// Check aliases
	found := false
	for _, alias := range snapshotCmd.Aliases {
		if alias == "snap" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'snap' alias for snapshot command")
	}
}

func TestSnapshotSaveCmd_Exists(t *testing.T) {
	if snapshotSaveCmd.Use != "save" {
		t.Errorf("expected Use 'save', got %s", snapshotSaveCmd.Use)
	}

	// Check note flag exists
	flag := snapshotSaveCmd.Flags().Lookup("note")
	if flag == nil {
		t.Error("expected 'note' flag to exist")
	}
}

func TestSnapshotListCmd_Exists(t *testing.T) {
	if snapshotListCmd.Use != "list" {
		t.Errorf("expected Use 'list', got %s", snapshotListCmd.Use)
	}

	if snapshotListCmd.Short == "" {
		t.Error("expected Short description for snapshot list command")
	}
}
