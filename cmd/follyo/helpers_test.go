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
		{1234.56789, "1234.56789"},
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
