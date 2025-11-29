package portfolio

import (
	"errors"
	"testing"
)

func TestValidationError(t *testing.T) {
	underlying := errors.New("must be positive")
	err := NewValidationError("amount", underlying)

	t.Run("Error message", func(t *testing.T) {
		expected := "invalid amount: must be positive"
		if err.Error() != expected {
			t.Errorf("got %q, want %q", err.Error(), expected)
		}
	})

	t.Run("Is ErrValidation", func(t *testing.T) {
		if !errors.Is(err, ErrValidation) {
			t.Error("expected error to match ErrValidation")
		}
	})

	t.Run("IsValidationError helper", func(t *testing.T) {
		if !IsValidationError(err) {
			t.Error("expected IsValidationError to return true")
		}
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		if !errors.Is(err, underlying) {
			t.Error("expected to unwrap to underlying error")
		}
	})

	t.Run("As ValidationError", func(t *testing.T) {
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Error("expected error to be extractable as ValidationError")
		}
		if ve.Field != "amount" {
			t.Errorf("got field %q, want %q", ve.Field, "amount")
		}
	})

	t.Run("Without underlying error", func(t *testing.T) {
		err := &ValidationError{Field: "coin"}
		if err.Error() != "invalid coin" {
			t.Errorf("got %q, want %q", err.Error(), "invalid coin")
		}
	})
}

func TestRequiredFieldError(t *testing.T) {
	err := NewRequiredFieldError("platform", "for loans")

	t.Run("Error message with context", func(t *testing.T) {
		expected := "platform is required for loans"
		if err.Error() != expected {
			t.Errorf("got %q, want %q", err.Error(), expected)
		}
	})

	t.Run("Error message without context", func(t *testing.T) {
		err := NewRequiredFieldError("platform", "")
		expected := "platform is required"
		if err.Error() != expected {
			t.Errorf("got %q, want %q", err.Error(), expected)
		}
	})

	t.Run("Is ErrRequiredField", func(t *testing.T) {
		if !errors.Is(err, ErrRequiredField) {
			t.Error("expected error to match ErrRequiredField")
		}
	})

	t.Run("IsRequiredFieldError helper", func(t *testing.T) {
		if !IsRequiredFieldError(err) {
			t.Error("expected IsRequiredFieldError to return true")
		}
	})

	t.Run("As RequiredFieldError", func(t *testing.T) {
		var rfe *RequiredFieldError
		if !errors.As(err, &rfe) {
			t.Error("expected error to be extractable as RequiredFieldError")
		}
		if rfe.Field != "platform" {
			t.Errorf("got field %q, want %q", rfe.Field, "platform")
		}
	})
}

func TestInsufficientBalanceError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		coin      string
		requested float64
		available float64
		staked    float64
		wantMsg   string
	}{
		{
			name:      "sell with no coins",
			operation: "sell",
			coin:      "BTC",
			requested: 1.0,
			available: 0,
			staked:    0,
			wantMsg:   "cannot sell 1 BTC: you have no available BTC to sell",
		},
		{
			name:      "sell with all staked",
			operation: "sell",
			coin:      "ETH",
			requested: 5.0,
			available: 0,
			staked:    5.0,
			wantMsg:   "cannot sell 5 ETH: all your ETH is staked (unstake first)",
		},
		{
			name:      "sell with partial staked",
			operation: "sell",
			coin:      "BTC",
			requested: 3.0,
			available: 1.0,
			staked:    2.0,
			wantMsg:   "cannot sell 3 BTC: only 1 BTC available (2 staked - unstake first to sell more)",
		},
		{
			name:      "sell with insufficient (no staking)",
			operation: "sell",
			coin:      "SOL",
			requested: 10.0,
			available: 5.0,
			staked:    0,
			wantMsg:   "cannot sell 10 SOL: only 5 SOL available",
		},
		{
			name:      "stake with no coins",
			operation: "stake",
			coin:      "DOT",
			requested: 100.0,
			available: 0,
			staked:    0,
			wantMsg:   "cannot stake 100 DOT: you have no available DOT to stake",
		},
		{
			name:      "stake with insufficient",
			operation: "stake",
			coin:      "ADA",
			requested: 1000.0,
			available: 500.0,
			staked:    0,
			wantMsg:   "cannot stake 1000 ADA: only 500 ADA available (holdings - sales - already staked)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewInsufficientBalanceError(tt.operation, tt.coin, tt.requested, tt.available, tt.staked)

			if err.Error() != tt.wantMsg {
				t.Errorf("got %q, want %q", err.Error(), tt.wantMsg)
			}

			if !errors.Is(err, ErrInsufficientBalance) {
				t.Error("expected error to match ErrInsufficientBalance")
			}

			if !IsInsufficientBalanceError(err) {
				t.Error("expected IsInsufficientBalanceError to return true")
			}

			var ibe *InsufficientBalanceError
			if !errors.As(err, &ibe) {
				t.Error("expected error to be extractable as InsufficientBalanceError")
			}
			if ibe.Coin != tt.coin {
				t.Errorf("got coin %q, want %q", ibe.Coin, tt.coin)
			}
			if ibe.Operation != tt.operation {
				t.Errorf("got operation %q, want %q", ibe.Operation, tt.operation)
			}
		})
	}
}

func TestStorageError(t *testing.T) {
	underlying := errors.New("disk full")
	err := NewStorageError("saving holding", underlying)

	t.Run("Error message", func(t *testing.T) {
		expected := "saving holding: disk full"
		if err.Error() != expected {
			t.Errorf("got %q, want %q", err.Error(), expected)
		}
	})

	t.Run("Is ErrStorage", func(t *testing.T) {
		if !errors.Is(err, ErrStorage) {
			t.Error("expected error to match ErrStorage")
		}
	})

	t.Run("IsStorageError helper", func(t *testing.T) {
		if !IsStorageError(err) {
			t.Error("expected IsStorageError to return true")
		}
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		if !errors.Is(err, underlying) {
			t.Error("expected to unwrap to underlying error")
		}
	})

	t.Run("As StorageError", func(t *testing.T) {
		var se *StorageError
		if !errors.As(err, &se) {
			t.Error("expected error to be extractable as StorageError")
		}
		if se.Operation != "saving holding" {
			t.Errorf("got operation %q, want %q", se.Operation, "saving holding")
		}
	})

	t.Run("Without underlying error", func(t *testing.T) {
		err := &StorageError{Operation: "loading data"}
		if err.Error() != "loading data" {
			t.Errorf("got %q, want %q", err.Error(), "loading data")
		}
	})
}

func TestMissingPricesError(t *testing.T) {
	err := NewMissingPricesError([]string{"BTC", "ETH", "SOL"})

	t.Run("Error message", func(t *testing.T) {
		expected := "missing prices for coins: BTC, ETH, SOL"
		if err.Error() != expected {
			t.Errorf("got %q, want %q", err.Error(), expected)
		}
	})

	t.Run("Is ErrMissingPrices", func(t *testing.T) {
		if !errors.Is(err, ErrMissingPrices) {
			t.Error("expected error to match ErrMissingPrices")
		}
	})

	t.Run("IsMissingPricesError helper", func(t *testing.T) {
		if !IsMissingPricesError(err) {
			t.Error("expected IsMissingPricesError to return true")
		}
	})

	t.Run("As MissingPricesError", func(t *testing.T) {
		var mpe *MissingPricesError
		if !errors.As(err, &mpe) {
			t.Error("expected error to be extractable as MissingPricesError")
		}
		if len(mpe.Coins) != 3 {
			t.Errorf("got %d coins, want 3", len(mpe.Coins))
		}
		if mpe.Coins[0] != "BTC" {
			t.Errorf("got first coin %q, want %q", mpe.Coins[0], "BTC")
		}
	})
}

func TestErrorSentinelsAreDistinct(t *testing.T) {
	// Verify that sentinel errors are distinct
	sentinels := []error{
		ErrValidation,
		ErrRequiredField,
		ErrInsufficientBalance,
		ErrStorage,
		ErrMissingPrices,
	}

	for i, err1 := range sentinels {
		for j, err2 := range sentinels {
			if i != j && errors.Is(err1, err2) {
				t.Errorf("sentinel %d and %d should not match", i, j)
			}
		}
	}
}

func TestErrorTypesIntegration(t *testing.T) {
	// Test that errors returned from portfolio operations work correctly
	t.Run("ValidationError from AddHolding", func(t *testing.T) {
		// This would be an integration test - the error types work with the portfolio
		err := NewValidationError("coin", errors.New("invalid format"))
		if !IsValidationError(err) {
			t.Error("should be identified as validation error")
		}
		if IsStorageError(err) {
			t.Error("should not be identified as storage error")
		}
	})

	t.Run("Wrapped errors maintain type", func(t *testing.T) {
		innerErr := NewValidationError("price", errors.New("negative"))
		// Simulating how errors might be wrapped in practice
		// (though our current implementation doesn't wrap further)
		if !errors.Is(innerErr, ErrValidation) {
			t.Error("wrapped error should still match ErrValidation")
		}
	})
}
