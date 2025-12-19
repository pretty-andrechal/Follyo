package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pretty-andrechal/follyo/internal/config"
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

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{0, "0s"},
		{5 * time.Second, "5s"},
		{30 * time.Second, "30s"},
		{59 * time.Second, "59s"},
		{60 * time.Second, "1m 00s"},
		{61 * time.Second, "1m 01s"},
		{90 * time.Second, "1m 30s"},
		{2 * time.Minute, "2m 00s"},
		{2*time.Minute + 30*time.Second, "2m 30s"},
		{-5 * time.Second, "0s"}, // Negative should be treated as 0
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %s, want %s", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestPrintStatusLine(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	oldStdout := osStdout
	osStdout = &buf
	defer func() { osStdout = oldStdout }()

	nextRefresh := time.Now().Add(30 * time.Second)
	printStatusLine(2*time.Minute, nextRefresh)

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("Next refresh in")) {
		t.Error("expected 'Next refresh in' in output")
	}
	if !bytes.Contains([]byte(output), []byte("Press Ctrl+C to exit")) {
		t.Error("expected 'Press Ctrl+C to exit' in output")
	}
}

func TestUpdateCountdown(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	oldStdout := osStdout
	osStdout = &buf
	defer func() { osStdout = oldStdout }()

	updateCountdown(2*time.Minute, 45*time.Second)

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("45s")) {
		t.Error("expected '45s' in countdown output")
	}
	// Interval is printed as Duration.String() format (2m0s)
	if !bytes.Contains([]byte(output), []byte("every 2m0s")) {
		t.Error("expected 'every 2m0s' interval in output")
	}
}

func TestPrintDashboardCoinLine(t *testing.T) {
	// Create a tabwriter
	var buf bytes.Buffer
	w := NewTable(&buf, true)

	// Create live prices map
	livePrices := map[string]float64{
		"BTC": 45000.0,
	}

	// Test the function with sample data
	value := printDashboardCoinLine(w.Writer(), "BTC", 1.5, livePrices)

	w.Flush()

	if value != 67500.0 {
		t.Errorf("expected value 67500.0, got %f", value)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("BTC")) {
		t.Error("expected 'BTC' in output")
	}
}

func TestPrintDashboardCoinLineNoPrices(t *testing.T) {
	// Create a tabwriter
	var buf bytes.Buffer
	w := NewTable(&buf, true)

	// Test with nil prices
	value := printDashboardCoinLine(w.Writer(), "BTC", 1.5, nil)

	w.Flush()

	if value != 0 {
		t.Errorf("expected value 0 for nil prices, got %f", value)
	}
}

func TestPrintDashboardCoinLine_PriceNotFound(t *testing.T) {
	var buf bytes.Buffer
	w := NewTable(&buf, true)

	// Price map exists but doesn't have the coin
	livePrices := map[string]float64{"BTC": 50000.0}
	value := printDashboardCoinLine(w.Writer(), "ETH", 10.0, livePrices)

	w.Flush()

	if value != 0 {
		t.Errorf("expected value 0 for missing price, got %f", value)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("N/A")) {
		t.Error("expected N/A in output for missing price")
	}
}

func TestDisplayDashboard_WithHoldings(t *testing.T) {
	// Set up temp directory and portfolio
	tmpDir, err := os.MkdirTemp("", "follyo-watch-holdings-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}

	dataPath := filepath.Join(dataDir, "portfolio.json")
	s, err := storage.New(dataPath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	oldP := p
	p = portfolio.New(s)
	defer func() { p = oldP }()

	// Add a holding
	p.AddHolding("BTC", 1.0, 50000.0, "Test", "", "")

	// Set up config - create the config store directly
	configPath := filepath.Join(dataDir, "config.json")
	cfg, err := config.New(configPath)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	oldCachedConfig := cachedConfig
	cachedConfig = cfg
	defer func() { cachedConfig = oldCachedConfig }()

	// Capture output
	var buf bytes.Buffer
	oldStdout := osStdout
	osStdout = &buf
	defer func() { osStdout = oldStdout }()

	displayDashboard()

	output := buf.String()

	// Should have NET HOLDINGS section with content
	if !bytes.Contains([]byte(output), []byte("BTC")) {
		t.Error("expected BTC in output")
	}
	if !bytes.Contains([]byte(output), []byte("Total Invested")) {
		t.Error("expected 'Total Invested' in output")
	}
}

func TestDisplayDashboard_WithStakes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "follyo-watch-stakes-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataDir := filepath.Join(tmpDir, "data")
	os.MkdirAll(dataDir, 0755)

	dataPath := filepath.Join(dataDir, "portfolio.json")
	s, _ := storage.New(dataPath)

	oldP := p
	p = portfolio.New(s)
	defer func() { p = oldP }()

	// Add holding first, then stake
	p.AddHolding("ETH", 10.0, 3000.0, "Test", "", "")
	p.AddStake("ETH", 5.0, "Staking Pool", nil, "", "")

	// Set up config - create the config store directly
	configPath := filepath.Join(dataDir, "config.json")
	cfg, _ := config.New(configPath)
	oldCachedConfig := cachedConfig
	cachedConfig = cfg
	defer func() { cachedConfig = oldCachedConfig }()

	var buf bytes.Buffer
	oldStdout := osStdout
	osStdout = &buf
	defer func() { osStdout = oldStdout }()

	displayDashboard()

	output := buf.String()

	if !bytes.Contains([]byte(output), []byte("STAKED ASSETS")) {
		t.Error("expected 'STAKED ASSETS' section in output")
	}
}

func TestDisplayDashboard_WithLoans(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "follyo-watch-loans-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataDir := filepath.Join(tmpDir, "data")
	os.MkdirAll(dataDir, 0755)

	dataPath := filepath.Join(dataDir, "portfolio.json")
	s, _ := storage.New(dataPath)

	oldP := p
	p = portfolio.New(s)
	defer func() { p = oldP }()

	// Add holding first, then loan
	p.AddHolding("BTC", 2.0, 50000.0, "Test", "", "")
	p.AddLoan("BTC", 0.5, "Lending Platform", nil, "", "")

	// Set up config - create the config store directly
	configPath := filepath.Join(dataDir, "config.json")
	cfg, _ := config.New(configPath)
	oldCachedConfig := cachedConfig
	cachedConfig = cfg
	defer func() { cachedConfig = oldCachedConfig }()

	var buf bytes.Buffer
	oldStdout := osStdout
	osStdout = &buf
	defer func() { osStdout = oldStdout }()

	displayDashboard()

	output := buf.String()

	if !bytes.Contains([]byte(output), []byte("OUTSTANDING LOANS")) {
		t.Error("expected 'OUTSTANDING LOANS' section in output")
	}
}

func TestPrintStatusLine_NegativeRemaining(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := osStdout
	osStdout = &buf
	defer func() { osStdout = oldStdout }()

	// Test with past time (negative remaining)
	pastTime := time.Now().Add(-30 * time.Second)
	printStatusLine(2*time.Minute, pastTime)

	output := buf.String()
	// Should handle negative gracefully (show 0s or similar)
	if !bytes.Contains([]byte(output), []byte("Next refresh in")) {
		t.Error("expected 'Next refresh in' in output")
	}
}

func TestDisplayDashboard_ReloadsData(t *testing.T) {
	// Set up temp directory and portfolio
	tmpDir, err := os.MkdirTemp("", "follyo-watch-reload-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}

	dataPath := filepath.Join(dataDir, "portfolio.json")
	s, err := storage.New(dataPath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	oldP := p
	p = portfolio.New(s)
	defer func() { p = oldP }()

	// Add initial holding
	p.AddHolding("BTC", 1.0, 50000.0, "Test", "", "")

	// Set up config
	configPath := filepath.Join(dataDir, "config.json")
	cfg, err := config.New(configPath)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	oldCachedConfig := cachedConfig
	cachedConfig = cfg
	defer func() { cachedConfig = oldCachedConfig }()

	// Capture output
	var buf bytes.Buffer
	oldStdout := osStdout
	osStdout = &buf
	defer func() { osStdout = oldStdout }()

	// First display - should show BTC
	displayDashboard()
	output1 := buf.String()

	if !bytes.Contains([]byte(output1), []byte("BTC")) {
		t.Error("expected BTC in first display")
	}

	// Simulate another process adding ETH directly to storage
	s2, _ := storage.New(dataPath)
	p2 := portfolio.New(s2)
	p2.AddHolding("ETH", 10.0, 3000.0, "Test2", "", "")

	// Reset buffer and display again
	buf.Reset()
	displayDashboard()
	output2 := buf.String()

	// After reload, should show both BTC and ETH
	if !bytes.Contains([]byte(output2), []byte("BTC")) {
		t.Error("expected BTC in second display after reload")
	}
	if !bytes.Contains([]byte(output2), []byte("ETH")) {
		t.Error("expected ETH in second display after reload - data was not reloaded")
	}
}
