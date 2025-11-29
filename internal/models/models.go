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
	// Platform can contain alphanumeric, spaces, dashes, and underscores
	platformRegex = regexp.MustCompile(`^[A-Za-z0-9\s\-_]+$`)
)

// MaxPlatformLength is the maximum length for platform names
const MaxPlatformLength = 50

// MaxNotesLength is the maximum length for notes fields
const MaxNotesLength = 500

// generateID creates a new unique ID with sufficient entropy
func generateID() string {
	return uuid.New().String()[:IDLength]
}

// ValidateCoinSymbol validates a coin symbol.
// Valid examples: BTC, ETH, USDT, sol
// Invalid examples: empty string, BTC!, BTC ETH
func ValidateCoinSymbol(coin string) error {
	if coin == "" {
		return fmt.Errorf("coin symbol cannot be empty (example: BTC, ETH)")
	}
	if len(coin) > MaxCoinSymbolLength {
		return fmt.Errorf("coin symbol too long (max %d characters, got %d)", MaxCoinSymbolLength, len(coin))
	}
	if !coinSymbolRegex.MatchString(coin) {
		return fmt.Errorf("coin symbol must contain only letters and numbers (example: BTC, USDT)")
	}
	return nil
}

// ValidateAmount validates an amount value.
// Amount must be positive (greater than zero).
func ValidateAmount(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive (got %.8g)", amount)
	}
	return nil
}

// ValidatePrice validates a price value.
// Price can be zero (for airdrops/free) but cannot be negative.
func ValidatePrice(price float64) error {
	if price < 0 {
		return fmt.Errorf("price cannot be negative (got %.2f)", price)
	}
	return nil
}

// ValidateDate validates a date string (YYYY-MM-DD format).
// Empty string is valid and defaults to today's date.
// Example valid dates: 2024-01-15, 2023-12-31
func ValidateDate(date string) error {
	if date == "" {
		return nil // Empty date is allowed (defaults to today)
	}
	if !dateFormatRegex.MatchString(date) {
		return fmt.Errorf("date must be in YYYY-MM-DD format (example: 2024-01-15, got: %s)", date)
	}
	// Verify it's a valid date
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date %s: not a real calendar date", date)
	}
	return nil
}

// ValidatePlatform validates a platform name.
// Platform names can contain alphanumeric characters, spaces, dashes, and underscores.
// Example valid platforms: "Coinbase", "Binance US", "FTX-backup", "cold_storage"
func ValidatePlatform(platform string) error {
	if platform == "" {
		return nil // Empty platform is allowed (optional field)
	}
	if len(platform) > MaxPlatformLength {
		return fmt.Errorf("platform name too long (max %d characters)", MaxPlatformLength)
	}
	if !platformRegex.MatchString(platform) {
		return fmt.Errorf("platform name can only contain letters, numbers, spaces, dashes, and underscores")
	}
	return nil
}

// ValidateNotes validates a notes field.
// Notes are optional but have a maximum length.
func ValidateNotes(notes string) error {
	if len(notes) > MaxNotesLength {
		return fmt.Errorf("notes too long (max %d characters, got %d)", MaxNotesLength, len(notes))
	}
	return nil
}

// Entity is an interface for types that have an ID field.
// This enables generic CRUD operations in the storage layer.
type Entity interface {
	GetID() string
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

// GetID returns the holding's ID.
func (h Holding) GetID() string { return h.ID }

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

// GetID returns the loan's ID.
func (l Loan) GetID() string { return l.ID }

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

// GetID returns the sale's ID.
func (s Sale) GetID() string { return s.ID }

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

// GetID returns the stake's ID.
func (st Stake) GetID() string { return st.ID }

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
