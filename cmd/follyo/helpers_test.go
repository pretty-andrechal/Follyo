package main

import (
	"testing"
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
			got := addCommas(tt.input)
			if got != tt.want {
				t.Errorf("addCommas(%s) = %s, want %s", tt.input, got, tt.want)
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
