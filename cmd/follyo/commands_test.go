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
