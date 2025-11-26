package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigStore(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")

	// Create new config store
	cs, err := New(configPath)
	if err != nil {
		t.Fatalf("Failed to create config store: %v", err)
	}

	// Test setting a mapping
	err = cs.SetTickerMapping("MUTE", "mute-io")
	if err != nil {
		t.Fatalf("Failed to set mapping: %v", err)
	}

	// Test getting a mapping
	geckoID := cs.GetTickerMapping("MUTE")
	if geckoID != "mute-io" {
		t.Errorf("Expected mute-io, got %s", geckoID)
	}

	// Test case insensitivity
	geckoID = cs.GetTickerMapping("mute")
	if geckoID != "mute-io" {
		t.Errorf("Expected mute-io for lowercase, got %s", geckoID)
	}

	// Test HasTickerMapping
	if !cs.HasTickerMapping("MUTE") {
		t.Error("Expected HasTickerMapping to return true")
	}
	if cs.HasTickerMapping("NOTEXIST") {
		t.Error("Expected HasTickerMapping to return false for non-existent")
	}

	// Test GetAllTickerMappings
	all := cs.GetAllTickerMappings()
	if len(all) != 1 {
		t.Errorf("Expected 1 mapping, got %d", len(all))
	}
	if all["MUTE"] != "mute-io" {
		t.Errorf("Expected MUTE -> mute-io in all mappings")
	}

	// Test removing a mapping
	err = cs.RemoveTickerMapping("MUTE")
	if err != nil {
		t.Fatalf("Failed to remove mapping: %v", err)
	}

	geckoID = cs.GetTickerMapping("MUTE")
	if geckoID != "" {
		t.Errorf("Expected empty string after removal, got %s", geckoID)
	}
}

func TestConfigPersistence(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")

	// Create and save config
	cs1, err := New(configPath)
	if err != nil {
		t.Fatalf("Failed to create config store: %v", err)
	}

	err = cs1.SetTickerMapping("TEST", "test-coin")
	if err != nil {
		t.Fatalf("Failed to set mapping: %v", err)
	}

	// Create new config store from same path - should load saved data
	cs2, err := New(configPath)
	if err != nil {
		t.Fatalf("Failed to create second config store: %v", err)
	}

	geckoID := cs2.GetTickerMapping("TEST")
	if geckoID != "test-coin" {
		t.Errorf("Expected test-coin from persisted config, got %s", geckoID)
	}
}

func TestConfigNonExistentPath(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Path with non-existent subdirectory
	configPath := filepath.Join(tmpDir, "subdir", "config.json")

	cs, err := New(configPath)
	if err != nil {
		t.Fatalf("Failed to create config store with new subdirectory: %v", err)
	}

	// Should work fine
	err = cs.SetTickerMapping("NEW", "new-coin")
	if err != nil {
		t.Fatalf("Failed to set mapping: %v", err)
	}
}
