package format

import (
	"errors"
	"testing"
)

func TestStatusError(t *testing.T) {
	tests := []struct {
		name     string
		context  string
		err      error
		expected string
	}{
		{"with error", "saving", errors.New("file not found"), "Error saving: file not found"},
		{"nil error", "loading", nil, ""},
		{"complex error", "connecting", errors.New("connection refused"), "Error connecting: connection refused"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StatusError(tt.context, tt.err)
			if result != tt.expected {
				t.Errorf("StatusError(%q, %v) = %q, want %q", tt.context, tt.err, result, tt.expected)
			}
		})
	}
}

func TestStatusSuccess(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		details  string
		expected string
	}{
		{"with details", "Added", "BTC purchase", "Added BTC purchase!"},
		{"no details", "Saved", "", "Saved!"},
		{"different action", "Deleted", "stake #123", "Deleted stake #123!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StatusSuccess(tt.action, tt.details)
			if result != tt.expected {
				t.Errorf("StatusSuccess(%q, %q) = %q, want %q", tt.action, tt.details, result, tt.expected)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		reason   string
		expected string
	}{
		{"amount error", "amount", "must be positive", "Invalid amount: must be positive"},
		{"date error", "date", "must be in YYYY-MM-DD format", "Invalid date: must be in YYYY-MM-DD format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidationError(tt.field, tt.reason)
			if result != tt.expected {
				t.Errorf("ValidationError(%q, %q) = %q, want %q", tt.field, tt.reason, result, tt.expected)
			}
		})
	}
}

func TestRequiredFieldError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected string
	}{
		{"coin", "Coin", "Coin is required"},
		{"amount", "Amount", "Amount is required"},
		{"platform", "Platform", "Platform is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RequiredFieldError(tt.field)
			if result != tt.expected {
				t.Errorf("RequiredFieldError(%q) = %q, want %q", tt.field, result, tt.expected)
			}
		})
	}
}

func TestNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		item     string
		expected string
	}{
		{"stake", "stake", "stake not found"},
		{"loan", "loan", "loan not found"},
		{"holding", "holding", "holding not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NotFoundError(tt.item)
			if result != tt.expected {
				t.Errorf("NotFoundError(%q) = %q, want %q", tt.item, result, tt.expected)
			}
		})
	}
}
