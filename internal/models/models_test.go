package models

import (
	"testing"
	"time"
)

func TestNewHolding(t *testing.T) {
	tests := []struct {
		name             string
		coin             string
		amount           float64
		purchasePriceUSD float64
		platform         string
		notes            string
		date             string
		wantDate         string
	}{
		{
			name:             "basic holding",
			coin:             "BTC",
			amount:           1.5,
			purchasePriceUSD: 50000,
			platform:         "Binance",
			notes:            "DCA",
			date:             "2024-01-15",
			wantDate:         "2024-01-15",
		},
		{
			name:             "empty date uses today",
			coin:             "ETH",
			amount:           10,
			purchasePriceUSD: 3000,
			platform:         "",
			notes:            "",
			date:             "",
			wantDate:         time.Now().Format("2006-01-02"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHolding(tt.coin, tt.amount, tt.purchasePriceUSD, tt.platform, tt.notes, tt.date)

			if h.ID == "" {
				t.Error("expected ID to be generated")
			}
			if len(h.ID) != 8 {
				t.Errorf("expected ID length 8, got %d", len(h.ID))
			}
			if h.Coin != tt.coin {
				t.Errorf("expected coin %s, got %s", tt.coin, h.Coin)
			}
			if h.Amount != tt.amount {
				t.Errorf("expected amount %f, got %f", tt.amount, h.Amount)
			}
			if h.PurchasePriceUSD != tt.purchasePriceUSD {
				t.Errorf("expected price %f, got %f", tt.purchasePriceUSD, h.PurchasePriceUSD)
			}
			if h.Platform != tt.platform {
				t.Errorf("expected platform %s, got %s", tt.platform, h.Platform)
			}
			if h.Notes != tt.notes {
				t.Errorf("expected notes %s, got %s", tt.notes, h.Notes)
			}
			if h.Date != tt.wantDate {
				t.Errorf("expected date %s, got %s", tt.wantDate, h.Date)
			}
		})
	}
}

func TestHolding_TotalValueUSD(t *testing.T) {
	h := Holding{
		Amount:           2.5,
		PurchasePriceUSD: 40000,
	}

	want := 100000.0
	got := h.TotalValueUSD()

	if got != want {
		t.Errorf("TotalValueUSD() = %f, want %f", got, want)
	}
}

func TestNewLoan(t *testing.T) {
	tests := []struct {
		name         string
		coin         string
		amount       float64
		platform     string
		interestRate *float64
		notes        string
		date         string
		wantDate     string
	}{
		{
			name:         "loan with rate",
			coin:         "USDT",
			amount:       5000,
			platform:     "Nexo",
			interestRate: floatPtr(6.9),
			notes:        "Credit line",
			date:         "2024-02-01",
			wantDate:     "2024-02-01",
		},
		{
			name:         "loan without rate",
			coin:         "BTC",
			amount:       0.5,
			platform:     "Celsius",
			interestRate: nil,
			notes:        "",
			date:         "",
			wantDate:     time.Now().Format("2006-01-02"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLoan(tt.coin, tt.amount, tt.platform, tt.interestRate, tt.notes, tt.date)

			if l.ID == "" {
				t.Error("expected ID to be generated")
			}
			if len(l.ID) != 8 {
				t.Errorf("expected ID length 8, got %d", len(l.ID))
			}
			if l.Coin != tt.coin {
				t.Errorf("expected coin %s, got %s", tt.coin, l.Coin)
			}
			if l.Amount != tt.amount {
				t.Errorf("expected amount %f, got %f", tt.amount, l.Amount)
			}
			if l.Platform != tt.platform {
				t.Errorf("expected platform %s, got %s", tt.platform, l.Platform)
			}
			if tt.interestRate != nil && (l.InterestRate == nil || *l.InterestRate != *tt.interestRate) {
				t.Errorf("expected interest rate %v, got %v", tt.interestRate, l.InterestRate)
			}
			if tt.interestRate == nil && l.InterestRate != nil {
				t.Errorf("expected nil interest rate, got %v", l.InterestRate)
			}
			if l.Date != tt.wantDate {
				t.Errorf("expected date %s, got %s", tt.wantDate, l.Date)
			}
		})
	}
}

func TestNewSale(t *testing.T) {
	tests := []struct {
		name         string
		coin         string
		amount       float64
		sellPriceUSD float64
		platform     string
		notes        string
		date         string
		wantDate     string
	}{
		{
			name:         "basic sale",
			coin:         "BTC",
			amount:       0.5,
			sellPriceUSD: 55000,
			platform:     "Kraken",
			notes:        "Taking profits",
			date:         "2024-03-01",
			wantDate:     "2024-03-01",
		},
		{
			name:         "sale without date",
			coin:         "ETH",
			amount:       5,
			sellPriceUSD: 3500,
			platform:     "",
			notes:        "",
			date:         "",
			wantDate:     time.Now().Format("2006-01-02"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSale(tt.coin, tt.amount, tt.sellPriceUSD, tt.platform, tt.notes, tt.date)

			if s.ID == "" {
				t.Error("expected ID to be generated")
			}
			if len(s.ID) != 8 {
				t.Errorf("expected ID length 8, got %d", len(s.ID))
			}
			if s.Coin != tt.coin {
				t.Errorf("expected coin %s, got %s", tt.coin, s.Coin)
			}
			if s.Amount != tt.amount {
				t.Errorf("expected amount %f, got %f", tt.amount, s.Amount)
			}
			if s.SellPriceUSD != tt.sellPriceUSD {
				t.Errorf("expected price %f, got %f", tt.sellPriceUSD, s.SellPriceUSD)
			}
			if s.Platform != tt.platform {
				t.Errorf("expected platform %s, got %s", tt.platform, s.Platform)
			}
			if s.Date != tt.wantDate {
				t.Errorf("expected date %s, got %s", tt.wantDate, s.Date)
			}
		})
	}
}

func TestSale_TotalValueUSD(t *testing.T) {
	s := Sale{
		Amount:       0.5,
		SellPriceUSD: 60000,
	}

	want := 30000.0
	got := s.TotalValueUSD()

	if got != want {
		t.Errorf("TotalValueUSD() = %f, want %f", got, want)
	}
}

func TestNewStake(t *testing.T) {
	tests := []struct {
		name     string
		coin     string
		amount   float64
		platform string
		apy      *float64
		notes    string
		date     string
		wantDate string
	}{
		{
			name:     "stake with APY",
			coin:     "ETH",
			amount:   10,
			platform: "Lido",
			apy:      floatPtr(4.5),
			notes:    "Staking rewards",
			date:     "2024-03-01",
			wantDate: "2024-03-01",
		},
		{
			name:     "stake without APY",
			coin:     "SOL",
			amount:   100,
			platform: "Coinbase",
			apy:      nil,
			notes:    "",
			date:     "",
			wantDate: time.Now().Format("2006-01-02"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := NewStake(tt.coin, tt.amount, tt.platform, tt.apy, tt.notes, tt.date)

			if st.ID == "" {
				t.Error("expected ID to be generated")
			}
			if len(st.ID) != 8 {
				t.Errorf("expected ID length 8, got %d", len(st.ID))
			}
			if st.Coin != tt.coin {
				t.Errorf("expected coin %s, got %s", tt.coin, st.Coin)
			}
			if st.Amount != tt.amount {
				t.Errorf("expected amount %f, got %f", tt.amount, st.Amount)
			}
			if st.Platform != tt.platform {
				t.Errorf("expected platform %s, got %s", tt.platform, st.Platform)
			}
			if tt.apy != nil && (st.APY == nil || *st.APY != *tt.apy) {
				t.Errorf("expected APY %v, got %v", tt.apy, st.APY)
			}
			if tt.apy == nil && st.APY != nil {
				t.Errorf("expected nil APY, got %v", st.APY)
			}
			if st.Date != tt.wantDate {
				t.Errorf("expected date %s, got %s", tt.wantDate, st.Date)
			}
		})
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
