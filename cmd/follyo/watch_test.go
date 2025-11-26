package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pretty-andrechal/follyo/internal/portfolio"
	"github.com/pretty-andrechal/follyo/internal/storage"
)

func TestWatchCmd_Flags(t *testing.T) {
	// Test that the watch command has correct configuration
	if watchCmd.Use != "watch" {
		t.Errorf("expected Use 'watch', got %s", watchCmd.Use)
	}

	// Check aliases
	expectedAliases := []string{"w", "live"}
	if len(watchCmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(watchCmd.Aliases))
	}

	// Check interval flag exists
	flag := watchCmd.Flags().Lookup("interval")
	if flag == nil {
		t.Error("expected 'interval' flag to exist")
	}
	if flag.Shorthand != "i" {
		t.Errorf("expected shorthand 'i', got '%s'", flag.Shorthand)
	}
}

func TestDisplayDashboard(t *testing.T) {
	// Set up temp directory and portfolio
	tmpDir, err := os.MkdirTemp("", "follyo-watch-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create data directory
	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}

	dataPath := filepath.Join(dataDir, "portfolio.json")
	s, err := storage.New(dataPath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// Set up global portfolio for the test
	oldP := p
	p = portfolio.New(s)
	defer func() { p = oldP }()

	// Set up config path
	configPath := filepath.Join(dataDir, "config.json")
	oldCachedConfig := cachedConfig
	cachedConfig = nil
	defer func() { cachedConfig = oldCachedConfig }()

	// Change to temp dir so config loads from correct location
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Capture output
	var buf bytes.Buffer
	oldStdout := osStdout
	osStdout = &buf
	defer func() { osStdout = oldStdout }()

	// Ensure config file exists
	os.WriteFile(configPath, []byte("{}"), 0600)

	// Test display with empty portfolio
	displayDashboard()

	output := buf.String()

	// Check for expected content
	if !bytes.Contains([]byte(output), []byte("FOLLYO - LIVE PORTFOLIO DASHBOARD")) {
		t.Error("expected dashboard header in output")
	}
	if !bytes.Contains([]byte(output), []byte("Last Update:")) {
		t.Error("expected 'Last Update:' in output")
	}
	if !bytes.Contains([]byte(output), []byte("NET HOLDINGS")) {
		t.Error("expected 'NET HOLDINGS' section in output")
	}
	if !bytes.Contains([]byte(output), []byte("PORTFOLIO TOTALS")) {
		t.Error("expected 'PORTFOLIO TOTALS' section in output")
	}
}

func TestDefaultRefreshInterval(t *testing.T) {
	expected := 2 * time.Minute
	if defaultRefreshInterval != expected {
		t.Errorf("expected default interval %v, got %v", expected, defaultRefreshInterval)
	}
}
