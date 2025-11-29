package models

import "time"

// Snapshot represents a point-in-time snapshot of the portfolio value
type Snapshot struct {
	ID            string                  `json:"id"`
	Timestamp     time.Time               `json:"timestamp"`
	HoldingsValue float64                 `json:"holdings_value"`
	LoansValue    float64                 `json:"loans_value"`
	NetValue      float64                 `json:"net_value"`
	TotalInvested float64                 `json:"total_invested"`
	TotalSold     float64                 `json:"total_sold"`
	ProfitLoss    float64                 `json:"profit_loss"`
	ProfitPercent float64                 `json:"profit_percent"`
	CoinValues    map[string]CoinSnapshot `json:"coin_values"`
	Note          string                  `json:"note,omitempty"`
}

// CoinSnapshot represents a coin's value at snapshot time
type CoinSnapshot struct {
	Amount float64 `json:"amount"`
	Price  float64 `json:"price"`
	Value  float64 `json:"value"`
}
