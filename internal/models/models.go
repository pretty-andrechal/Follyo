package models

import (
	"time"

	"github.com/google/uuid"
)

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
		ID:               uuid.New().String()[:8],
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
		ID:           uuid.New().String()[:8],
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
		ID:           uuid.New().String()[:8],
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
