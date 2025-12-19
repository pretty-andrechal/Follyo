package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/pretty-andrechal/follyo/cmd/follyo/tui/format"
)

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{1.0, "1"},
		{1.5, "1.5"},
		{0.123, "0.123"},
		{0.12300000, "0.123"},
		{100.00000000, "100"},
		{0.00000001, "0.00000001"},
		{1234.56789, "1,234.56789"},
		{1000000, "1,000,000"},
		{12345678.9, "12,345,678.9"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatAmount(tt.input)
			if got != tt.want {
				t.Errorf("formatAmount(%f) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestAddCommas(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1", "1"},
		{"12", "12"},
		{"123", "123"},
		{"1234", "1,234"},
		{"12345", "12,345"},
		{"123456", "123,456"},
		{"1234567", "1,234,567"},
		{"1234.56", "1,234.56"},
		{"-1234", "-1,234"},
		{"-1234567.89", "-1,234,567.89"},
		{"0", "0"},
		{"100", "100"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := format.AddCommas(tt.input)
			if got != tt.want {
				t.Errorf("format.AddCommas(%s) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatUSD(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "$0.00"},
		{1.5, "$1.50"},
		{1234.56, "$1,234.56"},
		{1000000, "$1,000,000.00"},
		{-500.25, "$-500.25"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatUSD(tt.input)
			if got != tt.want {
				t.Errorf("formatUSD(%f) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestSortedKeys(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]float64
		want []string
	}{
		{
			name: "basic sort",
			m:    map[string]float64{"BTC": 1.0, "ETH": 2.0, "ADA": 3.0},
			want: []string{"ADA", "BTC", "ETH"},
		},
		{
			name: "empty map",
			m:    map[string]float64{},
			want: []string{},
		},
		{
			name: "single element",
			m:    map[string]float64{"BTC": 1.0},
			want: []string{"BTC"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortedKeys(tt.m)
			if len(got) != len(tt.want) {
				t.Errorf("sortedKeys() returned %d elements, want %d", len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("sortedKeys()[%d] = %s, want %s", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestSortByUSDValue(t *testing.T) {
	tests := []struct {
		name    string
		amounts map[string]float64
		prices  map[string]float64
		want    []string
	}{
		{
			name:    "sort by value descending",
			amounts: map[string]float64{"BTC": 1.0, "ETH": 10.0, "ADA": 1000.0},
			prices:  map[string]float64{"BTC": 50000.0, "ETH": 3000.0, "ADA": 0.5},
			// BTC: 50000, ETH: 30000, ADA: 500
			want: []string{"BTC", "ETH", "ADA"},
		},
		{
			name:    "items without prices sorted alphabetically at end",
			amounts: map[string]float64{"BTC": 1.0, "UNKNOWN": 100.0, "ETH": 10.0},
			prices:  map[string]float64{"BTC": 50000.0, "ETH": 3000.0},
			// BTC: 50000, ETH: 30000, UNKNOWN: no price
			want: []string{"BTC", "ETH", "UNKNOWN"},
		},
		{
			name:    "nil prices falls back to alphabetical",
			amounts: map[string]float64{"BTC": 1.0, "ETH": 2.0, "ADA": 3.0},
			prices:  nil,
			want:    []string{"ADA", "BTC", "ETH"},
		},
		{
			name:    "equal values sorted alphabetically",
			amounts: map[string]float64{"BTC": 1.0, "ETH": 1.0, "ADA": 1.0},
			prices:  map[string]float64{"BTC": 100.0, "ETH": 100.0, "ADA": 100.0},
			want:    []string{"ADA", "BTC", "ETH"},
		},
		{
			name:    "empty map",
			amounts: map[string]float64{},
			prices:  map[string]float64{},
			want:    []string{},
		},
		{
			name:    "single element",
			amounts: map[string]float64{"BTC": 1.0},
			prices:  map[string]float64{"BTC": 50000.0},
			want:    []string{"BTC"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortByUSDValue(tt.amounts, tt.prices)
			if len(got) != len(tt.want) {
				t.Errorf("sortByUSDValue() returned %d elements, want %d", len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("sortByUSDValue()[%d] = %s, want %s", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestSafeDivide(t *testing.T) {
	tests := []struct {
		name        string
		numerator   float64
		denominator float64
		want        float64
	}{
		{
			name:        "normal division",
			numerator:   100,
			denominator: 4,
			want:        25,
		},
		{
			name:        "division by zero",
			numerator:   100,
			denominator: 0,
			want:        0,
		},
		{
			name:        "zero numerator",
			numerator:   0,
			denominator: 100,
			want:        0,
		},
		{
			name:        "negative numbers",
			numerator:   -50,
			denominator: 10,
			want:        -5,
		},
		{
			name:        "decimal result",
			numerator:   10,
			denominator: 3,
			want:        3.3333333333333335,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := safeDivide(tt.numerator, tt.denominator)
			if got != tt.want {
				t.Errorf("safeDivide(%f, %f) = %f, want %f", tt.numerator, tt.denominator, got, tt.want)
			}
		})
	}
}

func TestFormatAmountAligned(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{1.0, "1.0000"},
		{1.5, "1.5000"},
		{0.123, "0.1230"},
		{100.0, "100.0000"},
		{1234.56789, "1,234.5679"},
		{1000000, "1,000,000.0000"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatAmountAligned(tt.input)
			if got != tt.want {
				t.Errorf("formatAmountAligned(%f) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestBoolToOnOff(t *testing.T) {
	tests := []struct {
		input bool
		want  string
	}{
		{true, "on"},
		{false, "off"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := boolToOnOff(tt.input)
			if got != tt.want {
				t.Errorf("boolToOnOff(%v) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseOnOff(t *testing.T) {
	tests := []struct {
		input   string
		want    bool
		wantOK  bool
	}{
		{"on", true, true},
		{"ON", true, true},
		{"On", true, true},
		{"true", true, true},
		{"TRUE", true, true},
		{"1", true, true},
		{"yes", true, true},
		{"YES", true, true},
		{"off", false, true},
		{"OFF", false, true},
		{"false", false, true},
		{"FALSE", false, true},
		{"0", false, true},
		{"no", false, true},
		{"NO", false, true},
		{"invalid", false, false},
		{"maybe", false, false},
		{"", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := parseOnOff(tt.input)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("parseOnOff(%q) = (%v, %v), want (%v, %v)", tt.input, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestColorByValue(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		value float64
	}{
		{"positive value", "+$100", 100},
		{"negative value", "-$100", -100},
		{"zero value", "$0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorByValue(tt.text, tt.value)
			// Just verify it returns something (actual coloring depends on terminal)
			if got == "" {
				t.Errorf("colorByValue(%q, %f) returned empty string", tt.text, tt.value)
			}
			// For zero, should return text unchanged
			if tt.value == 0 && got != tt.text {
				t.Errorf("colorByValue(%q, 0) = %q, want %q", tt.text, got, tt.text)
			}
		})
	}
}

func TestTableWriter(t *testing.T) {
	// Test the Writer method
	var buf bytes.Buffer
	table := NewTable(&buf, false)

	w := table.Writer()
	if w == nil {
		t.Error("Writer() returned nil")
	}
}

func TestParsePriceFromArgs_WithPrice(t *testing.T) {
	// Test case: price provided as argument
	args := []string{"BTC", "0.5", "50000"}
	got := parsePriceFromArgs(args, 0, 0.5)
	want := 50000.0
	if got != want {
		t.Errorf("parsePriceFromArgs(%v, 0, 0.5) = %f, want %f", args, got, want)
	}
}

func TestParsePriceFromArgs_WithTotal(t *testing.T) {
	// Test case: total flag provided
	args := []string{"BTC", "0.5"}
	got := parsePriceFromArgs(args, 25000, 0.5)
	want := 50000.0 // 25000 / 0.5 = 50000
	if got != want {
		t.Errorf("parsePriceFromArgs(%v, 25000, 0.5) = %f, want %f", args, got, want)
	}
}

func TestParsePriceFromArgs_BothSpecified(t *testing.T) {
	// Save originals
	oldExit := osExit
	oldStderr := osStderr
	defer func() {
		osExit = oldExit
		osStderr = oldStderr
	}()

	// Capture stderr
	var errBuf bytes.Buffer
	osStderr = &errBuf

	// Mock osExit to panic for testing
	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
		panic("exit called")
	}

	defer func() {
		recover() // Catch the panic
	}()

	args := []string{"BTC", "0.5", "50000"}
	parsePriceFromArgs(args, 25000, 0.5) // Both price and total specified

	if !exitCalled {
		t.Error("expected osExit to be called when both price and total specified")
	}
}

func TestParsePriceFromArgs_NeitherSpecified(t *testing.T) {
	// Save originals
	oldExit := osExit
	oldStderr := osStderr
	defer func() {
		osExit = oldExit
		osStderr = oldStderr
	}()

	// Capture stderr
	var errBuf bytes.Buffer
	osStderr = &errBuf

	// Mock osExit to panic for testing
	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
		panic("exit called")
	}

	defer func() {
		recover() // Catch the panic
	}()

	args := []string{"BTC", "0.5"}
	parsePriceFromArgs(args, 0, 0.5) // Neither price nor total specified

	if !exitCalled {
		t.Error("expected osExit to be called when neither price nor total specified")
	}
}

func TestHandleRemoveByID_Success(t *testing.T) {
	// Test successful removal
	called := false
	remover := func(id string) (bool, error) {
		called = true
		return true, nil
	}

	// Capture stdout (handleRemoveByID prints to stdout)
	handleRemoveByID("test-id", "item", remover)

	if !called {
		t.Error("remover function was not called")
	}
}

func TestHandleRemoveByID_NotFound(t *testing.T) {
	// Test item not found
	remover := func(id string) (bool, error) {
		return false, nil
	}

	handleRemoveByID("test-id", "item", remover)
	// No error expected, just prints "not found"
}

func TestHandleRemoveByID_Error(t *testing.T) {
	// Save originals
	oldExit := osExit
	oldStderr := osStderr
	defer func() {
		osExit = oldExit
		osStderr = oldStderr
	}()

	// Capture stderr
	var errBuf bytes.Buffer
	osStderr = &errBuf

	// Mock osExit to panic for testing
	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
		panic("exit called")
	}

	defer func() {
		recover() // Catch the panic
	}()

	remover := func(id string) (bool, error) {
		return false, fmt.Errorf("database error")
	}

	handleRemoveByID("test-id", "item", remover)

	if !exitCalled {
		t.Error("expected osExit to be called on error")
	}
}

func TestParseFloat_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"123.45", 123.45},
		{"0", 0},
		{"1000000", 1000000},
		{"-50.5", -50.5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseFloat(tt.input, "test")
			if got != tt.want {
				t.Errorf("parseFloat(%q) = %f, want %f", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFloat_Invalid(t *testing.T) {
	// Save originals
	oldExit := osExit
	oldStderr := osStderr
	defer func() {
		osExit = oldExit
		osStderr = oldStderr
	}()

	// Capture stderr
	var errBuf bytes.Buffer
	osStderr = &errBuf

	// Mock osExit to panic for testing
	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
		panic("exit called")
	}

	defer func() {
		recover() // Catch the panic
	}()

	parseFloat("not-a-number", "amount")

	if !exitCalled {
		t.Error("expected osExit to be called for invalid input")
	}
}

func TestColorize_DisabledColors(t *testing.T) {
	// Test colorize when colors are disabled (non-terminal output)
	oldStdout := osStdout
	defer func() { osStdout = oldStdout }()

	// Use a non-file writer to simulate non-terminal
	var buf bytes.Buffer
	osStdout = &buf

	// Reset the once so we can test the non-terminal path
	// Note: We can't easily reset sync.Once, but the function
	// checks terminal first, so this should work
	result := colorize("test", colorGreen)

	// When not a terminal, should return text without color codes
	if result != "test" {
		t.Errorf("colorize with non-terminal should return plain text, got %q", result)
	}
}

func TestColorGreenText(t *testing.T) {
	// Test that colorGreenText returns something
	result := colorGreenText("test")
	if result == "" {
		t.Error("colorGreenText returned empty string")
	}
}

func TestColorRedText(t *testing.T) {
	// Test that colorRedText returns something
	result := colorRedText("test")
	if result == "" {
		t.Error("colorRedText returned empty string")
	}
}

func TestCollectAllCoins(t *testing.T) {
	tests := []struct {
		name      string
		holdings  map[string]float64
		stakes    map[string]float64
		loans     map[string]float64
		net       map[string]float64
		wantCount int
	}{
		{
			name:      "all different coins",
			holdings:  map[string]float64{"BTC": 1.0},
			stakes:    map[string]float64{"ETH": 2.0},
			loans:     map[string]float64{"SOL": 3.0},
			net:       map[string]float64{"ADA": 4.0},
			wantCount: 4,
		},
		{
			name:      "overlapping coins",
			holdings:  map[string]float64{"BTC": 1.0, "ETH": 2.0},
			stakes:    map[string]float64{"BTC": 0.5, "SOL": 1.0},
			loans:     map[string]float64{"ETH": 1.0},
			net:       map[string]float64{"BTC": 0.5, "ETH": 1.0, "SOL": 1.0},
			wantCount: 3, // BTC, ETH, SOL
		},
		{
			name:      "empty maps",
			holdings:  map[string]float64{},
			stakes:    map[string]float64{},
			loans:     map[string]float64{},
			net:       map[string]float64{},
			wantCount: 0,
		},
		{
			name:      "single coin in all maps",
			holdings:  map[string]float64{"BTC": 1.0},
			stakes:    map[string]float64{"BTC": 0.5},
			loans:     map[string]float64{"BTC": 0.2},
			net:       map[string]float64{"BTC": 0.3},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectAllCoins(tt.holdings, tt.stakes, tt.loans, tt.net)
			if len(got) != tt.wantCount {
				t.Errorf("collectAllCoins() returned %d coins, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestFetchPricesForCoins_EmptyList(t *testing.T) {
	result := fetchPricesForCoins([]string{})

	if result.IsOffline {
		t.Error("expected IsOffline to be false for empty coin list")
	}
	if result.Error != nil {
		t.Errorf("expected no error for empty coin list, got %v", result.Error)
	}
	if len(result.Prices) != 0 {
		t.Errorf("expected empty prices map, got %d entries", len(result.Prices))
	}
}

func TestGetPlatformWithDefault_Empty(t *testing.T) {
	// Test with empty platform - should try to load config
	// This tests the code path but config loading may fail in test env
	result := getPlatformWithDefault("")
	// Result could be empty or have a default depending on config
	// Just verify it doesn't panic
	_ = result
}

func TestGetPlatformWithDefault_Provided(t *testing.T) {
	// Test with provided platform - should return as-is
	result := getPlatformWithDefault("Coinbase")
	if result != "Coinbase" {
		t.Errorf("getPlatformWithDefault('Coinbase') = %q, want 'Coinbase'", result)
	}
}

func TestPrintCoinLine_WithPrices(t *testing.T) {
	var buf bytes.Buffer
	w := NewTable(&buf, true)

	livePrices := map[string]float64{"BTC": 50000.0}
	value := printCoinLine(w.Writer(), "BTC", 1.5, livePrices, false)
	w.Flush()

	expectedValue := 1.5 * 50000.0
	if value != expectedValue {
		t.Errorf("printCoinLine value = %f, want %f", value, expectedValue)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("BTC")) {
		t.Error("expected BTC in output")
	}
}

func TestPrintCoinLine_WithPricesAndPrefix(t *testing.T) {
	var buf bytes.Buffer
	w := NewTable(&buf, true)

	livePrices := map[string]float64{"BTC": 50000.0}
	value := printCoinLine(w.Writer(), "BTC", 1.5, livePrices, true)
	w.Flush()

	expectedValue := 1.5 * 50000.0
	if value != expectedValue {
		t.Errorf("printCoinLine value = %f, want %f", value, expectedValue)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("+")) {
		t.Error("expected + prefix in output for positive amount with showPrefix=true")
	}
}

func TestPrintCoinLine_PriceNotAvailable(t *testing.T) {
	var buf bytes.Buffer
	w := NewTable(&buf, true)

	// Price map exists but doesn't have ETH
	livePrices := map[string]float64{"BTC": 50000.0}
	value := printCoinLine(w.Writer(), "ETH", 10.0, livePrices, false)
	w.Flush()

	if value != 0 {
		t.Errorf("printCoinLine value = %f, want 0 for unavailable price", value)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("N/A")) {
		t.Error("expected N/A in output for unavailable price")
	}
}

func TestPrintCoinLine_NilPrices(t *testing.T) {
	var buf bytes.Buffer
	w := NewTable(&buf, true)

	value := printCoinLine(w.Writer(), "BTC", 1.5, nil, false)
	w.Flush()

	if value != 0 {
		t.Errorf("printCoinLine value = %f, want 0 for nil prices", value)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("BTC")) {
		t.Error("expected BTC in output")
	}
}

func TestMakeListRun_Empty(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := osStdout
	osStdout = &buf
	defer func() { osStdout = oldStdout }()

	cfg := ListConfig{
		EmptyMessage: "No items found",
		Headers:      []string{"ID", "Name"},
		FetchAndRender: func(t *Table) (int, error) {
			return 0, nil
		},
	}

	runFn := makeListRun(cfg)
	runFn(nil, nil)

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("No items found")) {
		t.Error("expected empty message in output")
	}
}

func TestMakeListRun_WithItems(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := osStdout
	osStdout = &buf
	defer func() { osStdout = oldStdout }()

	cfg := ListConfig{
		EmptyMessage: "No items found",
		Headers:      []string{"ID", "Name"},
		FetchAndRender: func(t *Table) (int, error) {
			t.Row("1", "Test Item")
			return 1, nil
		},
	}

	runFn := makeListRun(cfg)
	runFn(nil, nil)

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("Test Item")) {
		t.Error("expected item in output")
	}
	if bytes.Contains([]byte(output), []byte("No items found")) {
		t.Error("should not show empty message when items exist")
	}
}

func TestMakeListRun_Error(t *testing.T) {
	oldStdout := osStdout
	oldStderr := osStderr
	oldExit := osExit
	defer func() {
		osStdout = oldStdout
		osStderr = oldStderr
		osExit = oldExit
	}()

	var outBuf, errBuf bytes.Buffer
	osStdout = &outBuf
	osStderr = &errBuf

	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
		panic("exit called")
	}

	defer func() {
		recover()
	}()

	cfg := ListConfig{
		EmptyMessage: "No items found",
		Headers:      []string{"ID", "Name"},
		FetchAndRender: func(t *Table) (int, error) {
			return 0, fmt.Errorf("database error")
		},
	}

	runFn := makeListRun(cfg)
	runFn(nil, nil)

	if !exitCalled {
		t.Error("expected osExit to be called on error")
	}
}
