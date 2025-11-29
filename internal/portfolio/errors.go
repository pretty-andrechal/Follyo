// Package portfolio provides business logic for managing cryptocurrency portfolios.
package portfolio

import (
	"errors"
	"fmt"
	"strings"
)

// Error sentinel values for error type checking
var (
	ErrValidation          = errors.New("validation error")
	ErrRequiredField       = errors.New("required field missing")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrStorage             = errors.New("storage error")
	ErrMissingPrices       = errors.New("missing prices")
)

// ValidationError represents an input validation error.
type ValidationError struct {
	Field string // The field that failed validation
	Err   error  // The underlying validation error
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("invalid %s: %v", e.Field, e.Err)
	}
	return fmt.Sprintf("invalid %s", e.Field)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func (e *ValidationError) Is(target error) bool {
	return target == ErrValidation
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field string, err error) *ValidationError {
	return &ValidationError{Field: field, Err: err}
}

// RequiredFieldError represents a missing required field error.
type RequiredFieldError struct {
	Field   string // The required field name
	Context string // Optional context (e.g., "for loans")
}

func (e *RequiredFieldError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("%s is required %s", e.Field, e.Context)
	}
	return fmt.Sprintf("%s is required", e.Field)
}

func (e *RequiredFieldError) Is(target error) bool {
	return target == ErrRequiredField
}

// NewRequiredFieldError creates a new RequiredFieldError.
func NewRequiredFieldError(field, context string) *RequiredFieldError {
	return &RequiredFieldError{Field: field, Context: context}
}

// InsufficientBalanceError represents an error when trying to sell or stake
// more coins than available.
type InsufficientBalanceError struct {
	Coin            string  // The coin symbol
	RequestedAmount float64 // Amount user tried to sell/stake
	AvailableAmount float64 // Amount actually available
	StakedAmount    float64 // Amount currently staked (if relevant)
	Operation       string  // "sell" or "stake"
}

func (e *InsufficientBalanceError) Error() string {
	if e.AvailableAmount <= 0 {
		if e.StakedAmount > 0 && e.Operation == "sell" {
			return fmt.Sprintf("cannot %s %.8g %s: all your %s is staked (unstake first)",
				e.Operation, e.RequestedAmount, e.Coin, e.Coin)
		}
		return fmt.Sprintf("cannot %s %.8g %s: you have no available %s to %s",
			e.Operation, e.RequestedAmount, e.Coin, e.Coin, e.Operation)
	}
	if e.StakedAmount > 0 && e.Operation == "sell" {
		return fmt.Sprintf("cannot %s %.8g %s: only %.8g %s available (%.8g staked - unstake first to %s more)",
			e.Operation, e.RequestedAmount, e.Coin, e.AvailableAmount, e.Coin, e.StakedAmount, e.Operation)
	}
	if e.Operation == "stake" {
		return fmt.Sprintf("cannot %s %.8g %s: only %.8g %s available (holdings - sales - already staked)",
			e.Operation, e.RequestedAmount, e.Coin, e.AvailableAmount, e.Coin)
	}
	return fmt.Sprintf("cannot %s %.8g %s: only %.8g %s available",
		e.Operation, e.RequestedAmount, e.Coin, e.AvailableAmount, e.Coin)
}

func (e *InsufficientBalanceError) Is(target error) bool {
	return target == ErrInsufficientBalance
}

// NewInsufficientBalanceError creates a new InsufficientBalanceError.
func NewInsufficientBalanceError(operation, coin string, requested, available, staked float64) *InsufficientBalanceError {
	return &InsufficientBalanceError{
		Coin:            coin,
		RequestedAmount: requested,
		AvailableAmount: available,
		StakedAmount:    staked,
		Operation:       operation,
	}
}

// StorageError represents an error during storage operations.
type StorageError struct {
	Operation string // e.g., "saving holding", "removing loan"
	Err       error  // The underlying storage error
}

func (e *StorageError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Operation, e.Err)
	}
	return e.Operation
}

func (e *StorageError) Unwrap() error {
	return e.Err
}

func (e *StorageError) Is(target error) bool {
	return target == ErrStorage
}

// NewStorageError creates a new StorageError.
func NewStorageError(operation string, err error) *StorageError {
	return &StorageError{Operation: operation, Err: err}
}

// MissingPricesError represents an error when prices are missing for coins.
type MissingPricesError struct {
	Coins []string // List of coins missing prices
}

func (e *MissingPricesError) Error() string {
	return fmt.Sprintf("missing prices for coins: %s", strings.Join(e.Coins, ", "))
}

func (e *MissingPricesError) Is(target error) bool {
	return target == ErrMissingPrices
}

// NewMissingPricesError creates a new MissingPricesError.
func NewMissingPricesError(coins []string) *MissingPricesError {
	return &MissingPricesError{Coins: coins}
}

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidation)
}

// IsRequiredFieldError checks if an error is a required field error.
func IsRequiredFieldError(err error) bool {
	return errors.Is(err, ErrRequiredField)
}

// IsInsufficientBalanceError checks if an error is an insufficient balance error.
func IsInsufficientBalanceError(err error) bool {
	return errors.Is(err, ErrInsufficientBalance)
}

// IsStorageError checks if an error is a storage error.
func IsStorageError(err error) bool {
	return errors.Is(err, ErrStorage)
}

// IsMissingPricesError checks if an error is a missing prices error.
func IsMissingPricesError(err error) bool {
	return errors.Is(err, ErrMissingPrices)
}
