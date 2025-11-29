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

func TestConfigPreferences(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	cs, err := New(configPath)
	if err != nil {
		t.Fatalf("Failed to create config store: %v", err)
	}

	// Test GetPreferences returns valid preferences
	prefs := cs.GetPreferences()
	// GetPreferences should return a struct, test it's callable
	_ = prefs

	// Test FetchPrices (default should be true)
	if !cs.GetFetchPrices() {
		t.Error("Expected default FetchPrices to be true")
	}

	// Test SetFetchPrices
	err = cs.SetFetchPrices(false)
	if err != nil {
		t.Fatalf("SetFetchPrices failed: %v", err)
	}
	if cs.GetFetchPrices() {
		t.Error("Expected FetchPrices to be false after setting")
	}

	// Test ColorOutput (default should be true)
	if !cs.GetColorOutput() {
		t.Error("Expected default ColorOutput to be true")
	}

	// Test SetColorOutput
	err = cs.SetColorOutput(false)
	if err != nil {
		t.Fatalf("SetColorOutput failed: %v", err)
	}
	if cs.GetColorOutput() {
		t.Error("Expected ColorOutput to be false after setting")
	}

	// Test DefaultPlatform (default should be empty)
	if cs.GetDefaultPlatform() != "" {
		t.Errorf("Expected empty default platform, got %s", cs.GetDefaultPlatform())
	}

	// Test SetDefaultPlatform
	err = cs.SetDefaultPlatform("Coinbase")
	if err != nil {
		t.Fatalf("SetDefaultPlatform failed: %v", err)
	}
	if cs.GetDefaultPlatform() != "Coinbase" {
		t.Errorf("Expected Coinbase, got %s", cs.GetDefaultPlatform())
	}

	// Test ClearDefaultPlatform
	err = cs.ClearDefaultPlatform()
	if err != nil {
		t.Fatalf("ClearDefaultPlatform failed: %v", err)
	}
	if cs.GetDefaultPlatform() != "" {
		t.Errorf("Expected empty platform after clear, got %s", cs.GetDefaultPlatform())
	}
}

func TestConfigPreferencesPersistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")

	// Create and configure
	cs1, err := New(configPath)
	if err != nil {
		t.Fatalf("Failed to create config store: %v", err)
	}

	cs1.SetFetchPrices(false)
	cs1.SetColorOutput(false)
	cs1.SetDefaultPlatform("Binance")

	// Reload from disk
	cs2, err := New(configPath)
	if err != nil {
		t.Fatalf("Failed to reload config store: %v", err)
	}

	if cs2.GetFetchPrices() {
		t.Error("FetchPrices not persisted correctly")
	}
	if cs2.GetColorOutput() {
		t.Error("ColorOutput not persisted correctly")
	}
	if cs2.GetDefaultPlatform() != "Binance" {
		t.Errorf("DefaultPlatform not persisted, got %s", cs2.GetDefaultPlatform())
	}
}

func TestConfigValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	cs, err := New(configPath)
	if err != nil {
		t.Fatalf("Failed to create config store: %v", err)
	}

	// Test SetTickerMapping validation
	t.Run("empty ticker rejected", func(t *testing.T) {
		err := cs.SetTickerMapping("", "some-id")
		if err == nil {
			t.Error("Expected error for empty ticker")
		}
	})

	t.Run("invalid ticker rejected", func(t *testing.T) {
		err := cs.SetTickerMapping("BTC!", "bitcoin")
		if err == nil {
			t.Error("Expected error for invalid ticker with special char")
		}
	})

	t.Run("empty geckoID rejected", func(t *testing.T) {
		err := cs.SetTickerMapping("BTC", "")
		if err == nil {
			t.Error("Expected error for empty geckoID")
		}
	})

	t.Run("valid ticker mapping accepted", func(t *testing.T) {
		err := cs.SetTickerMapping("BTC", "bitcoin")
		if err != nil {
			t.Errorf("Expected valid mapping to succeed: %v", err)
		}
	})

	// Test SetDefaultPlatform validation
	t.Run("invalid platform rejected", func(t *testing.T) {
		err := cs.SetDefaultPlatform("Invalid@Platform!")
		if err == nil {
			t.Error("Expected error for invalid platform with special chars")
		}
	})

	t.Run("platform too long rejected", func(t *testing.T) {
		longPlatform := "ThisPlatformNameIsWayTooLongAndShouldBeRejectedByValidation123"
		err := cs.SetDefaultPlatform(longPlatform)
		if err == nil {
			t.Error("Expected error for platform name too long")
		}
	})

	t.Run("valid platform accepted", func(t *testing.T) {
		err := cs.SetDefaultPlatform("Coinbase Pro")
		if err != nil {
			t.Errorf("Expected valid platform to succeed: %v", err)
		}
	})

	t.Run("empty platform accepted", func(t *testing.T) {
		err := cs.SetDefaultPlatform("")
		if err != nil {
			t.Errorf("Expected empty platform to succeed: %v", err)
		}
	})
}
