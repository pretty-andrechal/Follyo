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
