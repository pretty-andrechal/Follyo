package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/pretty-andrechal/follyo/internal/portfolio"
	"github.com/pretty-andrechal/follyo/internal/storage"
)

// setupTestEnv creates a temp directory and initializes the portfolio for testing
func setupTestEnv(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "follyo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dataFile := filepath.Join(tmpDir, "portfolio.json")
	s, err := storage.New(dataFile)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create storage: %v", err)
	}
	p = portfolio.New(s)
	dataPath = dataFile

	// Setup mock for osStdout/osStderr to capture output
	oldStdout := osStdout
	oldStderr := osStderr

	cleanup := func() {
		osStdout = oldStdout
		osStderr = oldStderr
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// captureOutput returns a buffer for capturing stdout
func captureOutput() (*bytes.Buffer, func()) {
	var buf bytes.Buffer
	oldStdout := osStdout
	osStdout = &buf
	return &buf, func() {
		osStdout = oldStdout
	}
}

// TestBuyCommands tests the buy add, list, and remove commands
func TestBuyCommands(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Test buy add with price
	t.Run("buy add with price", func(t *testing.T) {
		err := buyAddCmd.Flags().Set("platform", "Coinbase")
		if err != nil {
			t.Fatalf("Failed to set platform flag: %v", err)
		}
		buyAddCmd.Run(buyAddCmd, []string{"BTC", "0.5", "50000"})

		// Verify the holding was added
		holdings, err := p.ListHoldings()
		if err != nil {
			t.Fatalf("Failed to list holdings: %v", err)
		}
		if len(holdings) != 1 {
			t.Errorf("Expected 1 holding, got %d", len(holdings))
		}
		if holdings[0].Coin != "BTC" {
			t.Errorf("Expected coin BTC, got %s", holdings[0].Coin)
		}
		if holdings[0].Amount != 0.5 {
			t.Errorf("Expected amount 0.5, got %f", holdings[0].Amount)
		}
		if holdings[0].PurchasePriceUSD != 50000 {
			t.Errorf("Expected price 50000, got %f", holdings[0].PurchasePriceUSD)
		}
	})

	// Test buy list
	t.Run("buy list", func(t *testing.T) {
		buf, restore := captureOutput()
		defer restore()

		buyListCmd.Run(buyListCmd, []string{})

		output := buf.String()
		if !strings.Contains(output, "BTC") {
			t.Errorf("Expected output to contain BTC, got: %s", output)
		}
	})

	// Test buy remove
	t.Run("buy remove", func(t *testing.T) {
		holdings, _ := p.ListHoldings()
		if len(holdings) == 0 {
			t.Fatal("No holdings to remove")
		}

		buyRemoveCmd.Run(buyRemoveCmd, []string{holdings[0].ID})

		// Verify removal
		holdings, _ = p.ListHoldings()
		if len(holdings) != 0 {
			t.Errorf("Expected 0 holdings after removal, got %d", len(holdings))
		}
	})
}

// TestBuyAddWithTotal tests the --total flag
func TestBuyAddWithTotal(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flags
	buyAddCmd.Flags().Set("total", "10000")
	buyAddCmd.Flags().Set("platform", "")

	buyAddCmd.Run(buyAddCmd, []string{"ETH", "5"})

	holdings, err := p.ListHoldings()
	if err != nil {
		t.Fatalf("Failed to list holdings: %v", err)
	}
	if len(holdings) != 1 {
		t.Fatalf("Expected 1 holding, got %d", len(holdings))
	}
	// Price should be 10000 / 5 = 2000
	if holdings[0].PurchasePriceUSD != 2000 {
		t.Errorf("Expected price 2000 (10000/5), got %f", holdings[0].PurchasePriceUSD)
	}

	// Reset flag
	buyAddCmd.Flags().Set("total", "0")
}

// TestSellCommands tests sell add, list, and remove commands
func TestSellCommands(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// First add a holding to sell from
	p.AddHolding("BTC", 1.0, 50000, "Coinbase", "", "")

	// Test sell add
	t.Run("sell add", func(t *testing.T) {
		sellAddCmd.Run(sellAddCmd, []string{"BTC", "0.5", "55000"})

		sales, err := p.ListSales()
		if err != nil {
			t.Fatalf("Failed to list sales: %v", err)
		}
		if len(sales) != 1 {
			t.Errorf("Expected 1 sale, got %d", len(sales))
		}
		if sales[0].Coin != "BTC" {
			t.Errorf("Expected coin BTC, got %s", sales[0].Coin)
		}
	})

	// Test sell list
	t.Run("sell list", func(t *testing.T) {
		buf, restore := captureOutput()
		defer restore()

		sellListCmd.Run(sellListCmd, []string{})

		output := buf.String()
		if !strings.Contains(output, "BTC") {
			t.Errorf("Expected output to contain BTC, got: %s", output)
		}
	})

	// Test sell remove
	t.Run("sell remove", func(t *testing.T) {
		sales, _ := p.ListSales()
		if len(sales) == 0 {
			t.Fatal("No sales to remove")
		}

		sellRemoveCmd.Run(sellRemoveCmd, []string{sales[0].ID})

		sales, _ = p.ListSales()
		if len(sales) != 0 {
			t.Errorf("Expected 0 sales after removal, got %d", len(sales))
		}
	})
}

// TestLoanCommands tests loan add, list, and remove commands
func TestLoanCommands(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Test loan add
	t.Run("loan add", func(t *testing.T) {
		loanAddCmd.Flags().Set("rate", "5.5")
		loanAddCmd.Run(loanAddCmd, []string{"USDC", "10000", "Nexo"})

		loans, err := p.ListLoans()
		if err != nil {
			t.Fatalf("Failed to list loans: %v", err)
		}
		if len(loans) != 1 {
			t.Errorf("Expected 1 loan, got %d", len(loans))
		}
		if loans[0].Coin != "USDC" {
			t.Errorf("Expected coin USDC, got %s", loans[0].Coin)
		}
		if loans[0].Amount != 10000 {
			t.Errorf("Expected amount 10000, got %f", loans[0].Amount)
		}
		loanAddCmd.Flags().Set("rate", "0")
	})

	// Test loan list
	t.Run("loan list", func(t *testing.T) {
		buf, restore := captureOutput()
		defer restore()

		loanListCmd.Run(loanListCmd, []string{})

		output := buf.String()
		if !strings.Contains(output, "USDC") {
			t.Errorf("Expected output to contain USDC, got: %s", output)
		}
	})

	// Test loan remove
	t.Run("loan remove", func(t *testing.T) {
		loans, _ := p.ListLoans()
		if len(loans) == 0 {
			t.Fatal("No loans to remove")
		}

		loanRemoveCmd.Run(loanRemoveCmd, []string{loans[0].ID})

		loans, _ = p.ListLoans()
		if len(loans) != 0 {
			t.Errorf("Expected 0 loans after removal, got %d", len(loans))
		}
	})
}

// TestStakeCommands tests stake add, list, and remove commands
func TestStakeCommands(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// First add a holding to stake from
	p.AddHolding("ETH", 10.0, 3000, "Coinbase", "", "")

	// Test stake add
	t.Run("stake add", func(t *testing.T) {
		stakeAddCmd.Flags().Set("apy", "4.5")
		stakeAddCmd.Run(stakeAddCmd, []string{"ETH", "5", "Lido"})

		stakes, err := p.ListStakes()
		if err != nil {
			t.Fatalf("Failed to list stakes: %v", err)
		}
		if len(stakes) != 1 {
			t.Errorf("Expected 1 stake, got %d", len(stakes))
		}
		if stakes[0].Coin != "ETH" {
			t.Errorf("Expected coin ETH, got %s", stakes[0].Coin)
		}
		stakeAddCmd.Flags().Set("apy", "0")
	})

	// Test stake list
	t.Run("stake list", func(t *testing.T) {
		buf, restore := captureOutput()
		defer restore()

		stakeListCmd.Run(stakeListCmd, []string{})

		output := buf.String()
		if !strings.Contains(output, "ETH") {
			t.Errorf("Expected output to contain ETH, got: %s", output)
		}
	})

	// Test stake remove
	t.Run("stake remove", func(t *testing.T) {
		stakes, _ := p.ListStakes()
		if len(stakes) == 0 {
			t.Fatal("No stakes to remove")
		}

		stakeRemoveCmd.Run(stakeRemoveCmd, []string{stakes[0].ID})

		stakes, _ = p.ListStakes()
		if len(stakes) != 0 {
			t.Errorf("Expected 0 stakes after removal, got %d", len(stakes))
		}
	})
}

// TestSummaryCommand tests the summary command
func TestSummaryCommand(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Add some data
	p.AddHolding("BTC", 1.0, 50000, "Coinbase", "", "")
	p.AddHolding("ETH", 10.0, 3000, "Binance", "", "")
	p.AddSale("BTC", 0.5, 55000, "Coinbase", "", "")
	p.AddLoan("USDC", 5000, "Nexo", nil, "", "")
	p.AddStake("ETH", 5.0, "Lido", nil, "", "")

	t.Run("summary without prices", func(t *testing.T) {
		buf, restore := captureOutput()
		defer restore()

		summaryCmd.Flags().Set("prices", "false")
		summaryCmd.Run(summaryCmd, []string{})

		output := buf.String()

		// Check for expected sections
		if !strings.Contains(output, "PORTFOLIO SUMMARY") {
			t.Error("Expected PORTFOLIO SUMMARY header")
		}
		if !strings.Contains(output, "HOLDINGS BY COIN") {
			t.Error("Expected HOLDINGS BY COIN section")
		}
		if !strings.Contains(output, "STAKED BY COIN") {
			t.Error("Expected STAKED BY COIN section")
		}
		if !strings.Contains(output, "LOANS BY COIN") {
			t.Error("Expected LOANS BY COIN section")
		}
		if !strings.Contains(output, "NET HOLDINGS") {
			t.Error("Expected NET HOLDINGS section")
		}
		if !strings.Contains(output, "Total Holdings: 2") {
			t.Error("Expected Total Holdings: 2")
		}
		if !strings.Contains(output, "Total Sales: 1") {
			t.Error("Expected Total Sales: 1")
		}
	})
}

// TestSummaryEmptyPortfolio tests summary with no data
func TestSummaryEmptyPortfolio(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	buf, restore := captureOutput()
	defer restore()

	summaryCmd.Run(summaryCmd, []string{})

	output := buf.String()
	if !strings.Contains(output, "(none)") {
		t.Error("Expected (none) for empty sections")
	}
}

// TestBuyListEmpty tests buy list with no holdings
func TestBuyListEmpty(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	buf, restore := captureOutput()
	defer restore()

	buyListCmd.Run(buyListCmd, []string{})

	output := buf.String()
	if !strings.Contains(output, "No purchases found") {
		t.Errorf("Expected 'No purchases found', got: %s", output)
	}
}

// TestSellListEmpty tests sell list with no sales
func TestSellListEmpty(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	buf, restore := captureOutput()
	defer restore()

	sellListCmd.Run(sellListCmd, []string{})

	output := buf.String()
	if !strings.Contains(output, "No sales found") {
		t.Errorf("Expected 'No sales found', got: %s", output)
	}
}

// TestLoanListEmpty tests loan list with no loans
func TestLoanListEmpty(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	buf, restore := captureOutput()
	defer restore()

	loanListCmd.Run(loanListCmd, []string{})

	output := buf.String()
	if !strings.Contains(output, "No loans found") {
		t.Errorf("Expected 'No loans found', got: %s", output)
	}
}

// TestStakeListEmpty tests stake list with no stakes
func TestStakeListEmpty(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	buf, restore := captureOutput()
	defer restore()

	stakeListCmd.Run(stakeListCmd, []string{})

	output := buf.String()
	if !strings.Contains(output, "No stakes found") {
		t.Errorf("Expected 'No stakes found', got: %s", output)
	}
}

// TestPrintCoinLine tests the printCoinLine helper function
func TestPrintCoinLine(t *testing.T) {
	tests := []struct {
		name       string
		coin       string
		amount     float64
		prices     map[string]float64
		showPrefix bool
		wantValue  float64
		wantOutput string
	}{
		{
			name:       "with price no prefix",
			coin:       "BTC",
			amount:     1.5,
			prices:     map[string]float64{"BTC": 50000},
			showPrefix: false,
			wantValue:  75000,
			wantOutput: "BTC:",
		},
		{
			name:       "with price and prefix positive",
			coin:       "ETH",
			amount:     10,
			prices:     map[string]float64{"ETH": 3000},
			showPrefix: true,
			wantValue:  30000,
			wantOutput: "+",
		},
		{
			name:       "no price available",
			coin:       "UNKNOWN",
			amount:     100,
			prices:     map[string]float64{"BTC": 50000},
			showPrefix: false,
			wantValue:  0,
			wantOutput: "N/A",
		},
		{
			name:       "nil prices",
			coin:       "BTC",
			amount:     1.0,
			prices:     nil,
			showPrefix: false,
			wantValue:  0,
			wantOutput: "BTC:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', tabwriter.AlignRight)

			got := printCoinLine(w, tt.coin, tt.amount, tt.prices, tt.showPrefix)
			w.Flush()

			if got != tt.wantValue {
				t.Errorf("printCoinLine() returned value = %f, want %f", got, tt.wantValue)
			}

			output := buf.String()
			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("printCoinLine() output = %s, want to contain %s", output, tt.wantOutput)
			}
		})
	}
}

// TestTickerListCommand tests the ticker list command
func TestTickerListCommand(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	buf, restore := captureOutput()
	defer restore()

	tickerListCmd.Flags().Set("all", "false")
	tickerListCmd.Run(tickerListCmd, []string{})

	output := buf.String()
	if !strings.Contains(output, "Ticker Mappings") {
		t.Error("Expected 'Ticker Mappings' header")
	}
	if !strings.Contains(output, "built-in") {
		t.Error("Expected mention of built-in mappings")
	}
}

// TestRootCmd tests that root command exists and has correct info
func TestRootCmd(t *testing.T) {
	if rootCmd.Use != "follyo" {
		t.Errorf("Expected root command Use to be 'follyo', got %s", rootCmd.Use)
	}
	if rootCmd.Short == "" {
		t.Error("Expected root command Short description to be non-empty")
	}
}

// setupSnapshotTestEnv creates a temp directory for snapshots and initializes the snapshot store
func setupSnapshotTestEnv(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, cleanup := setupTestEnv(t)

	// Initialize snapshot store in temp dir
	snapshotPath := filepath.Join(tmpDir, "snapshots.json")
	ss, err := storage.NewSnapshotStore(snapshotPath)
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create snapshot store: %v", err)
	}
	snapshotStore = ss

	return tmpDir, func() {
		snapshotStore = nil
		cleanup()
	}
}

// TestSnapshotListEmpty tests snapshot list with no snapshots
func TestSnapshotListEmpty(t *testing.T) {
	_, cleanup := setupSnapshotTestEnv(t)
	defer cleanup()

	buf, restore := captureOutput()
	defer restore()

	snapshotListCmd.Run(snapshotListCmd, []string{})

	output := buf.String()
	if !strings.Contains(output, "No snapshots found") {
		t.Errorf("Expected 'No snapshots found', got: %s", output)
	}
}

// TestSnapshotShowNotFound tests snapshot show with non-existent ID
func TestSnapshotShowNotFound(t *testing.T) {
	_, cleanup := setupSnapshotTestEnv(t)
	defer cleanup()

	// Capture stderr
	var errBuf bytes.Buffer
	oldStderr := osStderr
	osStderr = &errBuf
	defer func() { osStderr = oldStderr }()

	// Mock osExit with panic to stop execution (like real exit would)
	exitCalled := false
	oldExit := osExit
	osExit = func(code int) {
		exitCalled = true
		panic("exit called")
	}
	defer func() { osExit = oldExit }()

	// Recover from the panic
	func() {
		defer func() { recover() }()
		snapshotShowCmd.Run(snapshotShowCmd, []string{"nonexistent"})
	}()

	if !exitCalled {
		t.Error("Expected exit to be called for non-existent snapshot")
	}
	if !strings.Contains(errBuf.String(), "not found") {
		t.Errorf("Expected 'not found' error, got: %s", errBuf.String())
	}
}

// TestSnapshotRemoveNotFound tests snapshot remove with non-existent ID
func TestSnapshotRemoveNotFound(t *testing.T) {
	_, cleanup := setupSnapshotTestEnv(t)
	defer cleanup()

	buf, restore := captureOutput()
	defer restore()

	snapshotRemoveCmd.Run(snapshotRemoveCmd, []string{"nonexistent"})

	output := buf.String()
	if !strings.Contains(output, "not found") {
		t.Errorf("Expected 'not found' message, got: %s", output)
	}
}

// TestSnapshotCommands tests the snapshot command structure
func TestSnapshotCommands(t *testing.T) {
	// Test snapshotCmd
	if snapshotCmd.Use != "snapshot" {
		t.Errorf("Expected snapshotCmd Use to be 'snapshot', got %s", snapshotCmd.Use)
	}

	// Check alias
	found := false
	for _, alias := range snapshotCmd.Aliases {
		if alias == "snap" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected snapshotCmd to have 'snap' alias")
	}

	// Test subcommands exist
	subCmds := map[string]bool{
		"save":    false,
		"list":    false,
		"show":    false,
		"compare": false,
		"remove":  false,
	}

	for _, cmd := range snapshotCmd.Commands() {
		if _, ok := subCmds[cmd.Use]; ok {
			subCmds[cmd.Use] = true
		} else if strings.HasPrefix(cmd.Use, "save") ||
			strings.HasPrefix(cmd.Use, "show") ||
			strings.HasPrefix(cmd.Use, "compare") ||
			strings.HasPrefix(cmd.Use, "remove") {
			// Handle commands with arguments like "show ID"
			for name := range subCmds {
				if strings.HasPrefix(cmd.Use, name) {
					subCmds[name] = true
				}
			}
		}
	}

	for name, found := range subCmds {
		if !found {
			t.Errorf("Expected snapshotCmd to have '%s' subcommand", name)
		}
	}
}

// TestFormatUSDWithSign tests the formatUSDWithSign helper
func TestFormatUSDWithSign(t *testing.T) {
	tests := []struct {
		amount   float64
		expected string
	}{
		{1000, "+$1,000.00"},
		{-500, "$-500.00"},
		{0, "+$0.00"},
		{12345.67, "+$12,345.67"},
	}

	for _, tt := range tests {
		result := formatUSDWithSign(tt.amount)
		if result != tt.expected {
			t.Errorf("formatUSDWithSign(%f) = %s, want %s", tt.amount, result, tt.expected)
		}
	}
}

// TestFormatChangeWithPercent tests the formatChangeWithPercent helper
func TestFormatChangeWithPercent(t *testing.T) {
	tests := []struct {
		change  float64
		percent float64
		want    string
	}{
		{1000, 10, "+$1,000.00 (10.0%)"},
		{-500, -5, "$-500.00 (-5.0%)"},
		{0, 0, "+$0.00 (0.0%)"},
	}

	for _, tt := range tests {
		result := formatChangeWithPercent(tt.change, tt.percent)
		if result != tt.want {
			t.Errorf("formatChangeWithPercent(%f, %f) = %s, want %s", tt.change, tt.percent, result, tt.want)
		}
	}
}

// Integration Tests - Test complete workflows

// TestCompleteWorkflow_BuySellStake tests a complete workflow of buying, selling, and staking
func TestCompleteWorkflow_BuySellStake(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// 1. Buy some BTC
	buyAddCmd.Flags().Set("platform", "Coinbase")
	buyAddCmd.Flags().Set("total", "0")
	buyAddCmd.Run(buyAddCmd, []string{"BTC", "2.0", "50000"})

	// Verify purchase
	holdings, _ := p.ListHoldings()
	if len(holdings) != 1 {
		t.Fatalf("Expected 1 holding, got %d", len(holdings))
	}
	if holdings[0].Amount != 2.0 {
		t.Errorf("Expected 2.0 BTC, got %f", holdings[0].Amount)
	}

	// 2. Buy more BTC
	buyAddCmd.Run(buyAddCmd, []string{"BTC", "1.0", "55000"})

	holdings, _ = p.ListHoldings()
	if len(holdings) != 2 {
		t.Fatalf("Expected 2 holdings, got %d", len(holdings))
	}

	// 3. Sell some BTC
	sellAddCmd.Run(sellAddCmd, []string{"BTC", "0.5", "60000"})

	sales, _ := p.ListSales()
	if len(sales) != 1 {
		t.Fatalf("Expected 1 sale, got %d", len(sales))
	}

	// 4. Stake some BTC
	stakeAddCmd.Flags().Set("apy", "5.0")
	stakeAddCmd.Run(stakeAddCmd, []string{"BTC", "1.0", "Lido"})

	stakes, _ := p.ListStakes()
	if len(stakes) != 1 {
		t.Fatalf("Expected 1 stake, got %d", len(stakes))
	}

	// 5. Verify summary data
	summary, err := p.GetSummary()
	if err != nil {
		t.Fatalf("Failed to get summary: %v", err)
	}

	// Holdings should be 3.0 - 0.5 = 2.5 BTC
	if summary.HoldingsByCoin["BTC"] != 2.5 {
		t.Errorf("Expected 2.5 BTC holdings, got %f", summary.HoldingsByCoin["BTC"])
	}

	// Available should be 2.5 - 1.0 (staked) = 1.5 BTC
	if summary.AvailableByCoin["BTC"] != 1.5 {
		t.Errorf("Expected 1.5 BTC available, got %f", summary.AvailableByCoin["BTC"])
	}

	// Total invested = (2 * 50000) + (1 * 55000) = 155000
	expectedInvested := 155000.0
	if summary.TotalInvestedUSD != expectedInvested {
		t.Errorf("Expected invested $%f, got $%f", expectedInvested, summary.TotalInvestedUSD)
	}

	// Total sold = 0.5 * 60000 = 30000
	expectedSold := 30000.0
	if summary.TotalSoldUSD != expectedSold {
		t.Errorf("Expected sold $%f, got $%f", expectedSold, summary.TotalSoldUSD)
	}

	// Reset flags
	stakeAddCmd.Flags().Set("apy", "0")
	buyAddCmd.Flags().Set("platform", "")
}

// TestCompleteWorkflow_LoanManagement tests a complete loan workflow
func TestCompleteWorkflow_LoanManagement(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// 1. Take out a loan
	loanAddCmd.Flags().Set("rate", "6.5")
	loanAddCmd.Run(loanAddCmd, []string{"USDC", "10000", "Nexo"})

	loans, _ := p.ListLoans()
	if len(loans) != 1 {
		t.Fatalf("Expected 1 loan, got %d", len(loans))
	}

	// 2. Take another loan
	loanAddCmd.Run(loanAddCmd, []string{"USDT", "5000", "Celsius"})

	loans, _ = p.ListLoans()
	if len(loans) != 2 {
		t.Fatalf("Expected 2 loans, got %d", len(loans))
	}

	// 3. Pay off first loan (remove it)
	loanRemoveCmd.Run(loanRemoveCmd, []string{loans[0].ID})

	loans, _ = p.ListLoans()
	if len(loans) != 1 {
		t.Fatalf("Expected 1 loan after removal, got %d", len(loans))
	}

	// 4. Verify summary
	summary, _ := p.GetSummary()
	if summary.TotalLoansCount != 1 {
		t.Errorf("Expected 1 loan in summary, got %d", summary.TotalLoansCount)
	}

	// Reset flags
	loanAddCmd.Flags().Set("rate", "0")
}

// TestStakingValidation tests that staking more than available fails
func TestStakingValidation(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Add holdings
	p.AddHolding("ETH", 10.0, 3000, "", "", "")

	// Try to stake more than owned - should fail
	_, err := p.AddStake("ETH", 15.0, "Lido", nil, "", "")
	if err == nil {
		t.Error("Expected error when staking more than owned")
	}

	// Stake exactly what's available - should succeed
	_, err = p.AddStake("ETH", 10.0, "Lido", nil, "", "")
	if err != nil {
		t.Errorf("Expected successful stake, got error: %v", err)
	}

	// Try to stake more - should fail (nothing left)
	_, err = p.AddStake("ETH", 0.1, "Coinbase", nil, "", "")
	if err == nil {
		t.Error("Expected error when staking with no available balance")
	}
}

// TestMultipleCoinPortfolio tests a portfolio with multiple different coins
func TestMultipleCoinPortfolio(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Add multiple coins
	coins := []struct {
		coin   string
		amount float64
		price  float64
	}{
		{"BTC", 1.0, 50000},
		{"ETH", 10.0, 3000},
		{"SOL", 100.0, 100},
		{"ADA", 5000.0, 0.5},
	}

	for _, c := range coins {
		p.AddHolding(c.coin, c.amount, c.price, "", "", "")
	}

	// Verify all holdings
	holdings, _ := p.ListHoldings()
	if len(holdings) != 4 {
		t.Errorf("Expected 4 holdings, got %d", len(holdings))
	}

	// Get summary
	summary, _ := p.GetSummary()

	// Verify each coin
	expectedHoldings := map[string]float64{
		"BTC": 1.0,
		"ETH": 10.0,
		"SOL": 100.0,
		"ADA": 5000.0,
	}

	for coin, expected := range expectedHoldings {
		if summary.HoldingsByCoin[coin] != expected {
			t.Errorf("Expected %f %s, got %f", expected, coin, summary.HoldingsByCoin[coin])
		}
	}

	// Verify total invested
	// (1*50000) + (10*3000) + (100*100) + (5000*0.5) = 50000 + 30000 + 10000 + 2500 = 92500
	expectedTotal := 92500.0
	if summary.TotalInvestedUSD != expectedTotal {
		t.Errorf("Expected total invested $%f, got $%f", expectedTotal, summary.TotalInvestedUSD)
	}
}

// TestRemoveOperations tests various remove operations
func TestRemoveOperations(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Add test data
	h, _ := p.AddHolding("BTC", 1.0, 50000, "", "", "")
	s, _ := p.AddSale("BTC", 0.5, 55000, "", "", "")
	l, _ := p.AddLoan("USDC", 1000, "Nexo", nil, "", "")

	// Add holding first, then stake
	p.AddHolding("ETH", 10.0, 3000, "", "", "")
	st, _ := p.AddStake("ETH", 5.0, "Lido", nil, "", "")

	// Verify initial counts
	holdings, _ := p.ListHoldings()
	if len(holdings) != 2 {
		t.Fatalf("Expected 2 holdings before removal, got %d", len(holdings))
	}

	// Remove each type using the command handlers
	buyRemoveCmd.Run(buyRemoveCmd, []string{h.ID})
	sellRemoveCmd.Run(sellRemoveCmd, []string{s.ID})
	loanRemoveCmd.Run(loanRemoveCmd, []string{l.ID})
	stakeRemoveCmd.Run(stakeRemoveCmd, []string{st.ID})

	// Verify all removed
	holdings, _ = p.ListHoldings()
	if len(holdings) != 1 { // Still have ETH holding
		t.Errorf("Expected 1 holding (ETH), got %d", len(holdings))
	}

	sales, _ := p.ListSales()
	if len(sales) != 0 {
		t.Errorf("Expected 0 sales, got %d", len(sales))
	}

	loans, _ := p.ListLoans()
	if len(loans) != 0 {
		t.Errorf("Expected 0 loans, got %d", len(loans))
	}

	stakes, _ := p.ListStakes()
	if len(stakes) != 0 {
		t.Errorf("Expected 0 stakes, got %d", len(stakes))
	}
}
