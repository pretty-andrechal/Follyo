package models

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// Validation constants
const (
	// IDLength is the length of generated IDs (12 hex chars = 48 bits of entropy)
	IDLength = 12
	// MaxCoinSymbolLength is the maximum length for coin symbols
	MaxCoinSymbolLength = 10
)

// Validation regex patterns
var (
	coinSymbolRegex = regexp.MustCompile(`^[A-Za-z0-9]+$`)
	dateFormatRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

// generateID creates a new unique ID with sufficient entropy
func generateID() string {
	return uuid.New().String()[:IDLength]
}

// ValidateCoinSymbol validates a coin symbol
func ValidateCoinSymbol(coin string) error {
	if coin == "" {
		return fmt.Errorf("coin symbol cannot be empty")
	}
	if len(coin) > MaxCoinSymbolLength {
		return fmt.Errorf("coin symbol too long (max %d characters)", MaxCoinSymbolLength)
	}
	if !coinSymbolRegex.MatchString(coin) {
		return fmt.Errorf("coin symbol must contain only alphanumeric characters")
	}
	return nil
}

// ValidateAmount validates an amount value
func ValidateAmount(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return nil
}

// ValidatePrice validates a price value
func ValidatePrice(price float64) error {
	if price < 0 {
		return fmt.Errorf("price cannot be negative")
	}
	return nil
}

// ValidateDate validates a date string (YYYY-MM-DD format)
func ValidateDate(date string) error {
	if date == "" {
		return nil // Empty date is allowed (defaults to today)
	}
	if !dateFormatRegex.MatchString(date) {
		return fmt.Errorf("date must be in YYYY-MM-DD format")
	}
	// Verify it's a valid date
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date: %s", date)
	}
	return nil
}

// Holding represents a crypto holding/purchase.
type Holding struct {
	ID               string  `json:"id"`
	Coin             string  `json:"coin"`
	Amount           float64 `json:"amount"`
	PurchasePriceUSD float64 `json:"purchase_price_usd"`
	Date             string  `json:"date"`
	Platform         string  `json:"platform,omitempty"`
	Notes            string  `json:"notes,omitempty"`
}

// NewHolding creates a new holding with auto-generated ID and date.
func NewHolding(coin string, amount, purchasePriceUSD float64, platform, notes, date string) Holding {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	return Holding{
		ID:               generateID(),
		Coin:             coin,
		Amount:           amount,
		PurchasePriceUSD: purchasePriceUSD,
		Date:             date,
		Platform:         platform,
		Notes:            notes,
	}
}

// TotalValueUSD returns the total value at purchase price.
func (h Holding) TotalValueUSD() float64 {
	return h.Amount * h.PurchasePriceUSD
}

// Loan represents a crypto loan on a platform.
type Loan struct {
	ID           string   `json:"id"`
	Coin         string   `json:"coin"`
	Amount       float64  `json:"amount"`
	Platform     string   `json:"platform"`
	Date         string   `json:"date"`
	InterestRate *float64 `json:"interest_rate,omitempty"`
	Notes        string   `json:"notes,omitempty"`
}

// NewLoan creates a new loan with auto-generated ID and date.
func NewLoan(coin string, amount float64, platform string, interestRate *float64, notes, date string) Loan {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	return Loan{
		ID:           generateID(),
		Coin:         coin,
		Amount:       amount,
		Platform:     platform,
		Date:         date,
		InterestRate: interestRate,
		Notes:        notes,
	}
}

// Sale represents a crypto sale.
type Sale struct {
	ID           string  `json:"id"`
	Coin         string  `json:"coin"`
	Amount       float64 `json:"amount"`
	SellPriceUSD float64 `json:"sell_price_usd"`
	Date         string  `json:"date"`
	Platform     string  `json:"platform,omitempty"`
	Notes        string  `json:"notes,omitempty"`
}

// NewSale creates a new sale with auto-generated ID and date.
func NewSale(coin string, amount, sellPriceUSD float64, platform, notes, date string) Sale {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	return Sale{
		ID:           generateID(),
		Coin:         coin,
		Amount:       amount,
		SellPriceUSD: sellPriceUSD,
		Date:         date,
		Platform:     platform,
		Notes:        notes,
	}
}

// TotalValueUSD returns the total value at sell price.
func (s Sale) TotalValueUSD() float64 {
	return s.Amount * s.SellPriceUSD
}

// Stake represents crypto that is staked on a platform.
type Stake struct {
	ID       string   `json:"id"`
	Coin     string   `json:"coin"`
	Amount   float64  `json:"amount"`
	Platform string   `json:"platform"`
	Date     string   `json:"date"`
	APY      *float64 `json:"apy,omitempty"`
	Notes    string   `json:"notes,omitempty"`
}

// NewStake creates a new stake with auto-generated ID and date.
func NewStake(coin string, amount float64, platform string, apy *float64, notes, date string) Stake {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	return Stake{
		ID:       generateID(),
		Coin:     coin,
		Amount:   amount,
		Platform: platform,
		Date:     date,
		APY:      apy,
		Notes:    notes,
	}
}
