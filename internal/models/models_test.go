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
			if len(h.ID) != IDLength {
				t.Errorf("expected ID length %d, got %d", IDLength, len(h.ID))
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
			if len(l.ID) != IDLength {
				t.Errorf("expected ID length %d, got %d", IDLength, len(l.ID))
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
			if len(s.ID) != IDLength {
				t.Errorf("expected ID length %d, got %d", IDLength, len(s.ID))
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
			if len(st.ID) != IDLength {
				t.Errorf("expected ID length %d, got %d", IDLength, len(st.ID))
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

// TestValidateCoinSymbol tests coin symbol validation
func TestValidateCoinSymbol(t *testing.T) {
	tests := []struct {
		coin    string
		wantErr bool
	}{
		{"BTC", false},
		{"ETH", false},
		{"USDT", false},
		{"btc", false},       // lowercase valid
		{"BTC123", false},    // alphanumeric valid
		{"", true},           // empty invalid
		{"BTC!", true},       // special char invalid
		{"BTC ETH", true},    // space invalid
		{"VERYLONGCOIN", true}, // too long (>10 chars)
	}

	for _, tt := range tests {
		t.Run(tt.coin, func(t *testing.T) {
			err := ValidateCoinSymbol(tt.coin)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCoinSymbol(%q) error = %v, wantErr %v", tt.coin, err, tt.wantErr)
			}
		})
	}
}

// TestValidateAmount tests amount validation
func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		wantErr bool
	}{
		{"positive", 1.0, false},
		{"small positive", 0.00001, false},
		{"large positive", 1000000.0, false},
		{"zero", 0, true},
		{"negative", -1.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmount(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAmount(%f) error = %v, wantErr %v", tt.amount, err, tt.wantErr)
			}
		})
	}
}

// TestValidatePrice tests price validation
func TestValidatePrice(t *testing.T) {
	tests := []struct {
		name    string
		price   float64
		wantErr bool
	}{
		{"positive", 50000.0, false},
		{"zero", 0, false},  // Zero is valid (free/airdrop)
		{"small", 0.0001, false},
		{"negative", -1.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePrice(tt.price)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePrice(%f) error = %v, wantErr %v", tt.price, err, tt.wantErr)
			}
		})
	}
}

// TestValidateDate tests date validation
func TestValidateDate(t *testing.T) {
	tests := []struct {
		date    string
		wantErr bool
	}{
		{"2024-01-15", false},
		{"2023-12-31", false},
		{"", false},             // empty is valid (defaults to today)
		{"2024-1-15", true},     // wrong format
		{"01-15-2024", true},    // US format invalid
		{"2024/01/15", true},    // slashes invalid
		{"not-a-date", true},
		{"2024-02-30", true},    // invalid day
	}

	for _, tt := range tests {
		t.Run(tt.date, func(t *testing.T) {
			err := ValidateDate(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDate(%q) error = %v, wantErr %v", tt.date, err, tt.wantErr)
			}
		})
	}
}

// TestValidatePlatform tests platform name validation
func TestValidatePlatform(t *testing.T) {
	tests := []struct {
		platform string
		wantErr  bool
	}{
		{"Coinbase", false},
		{"Binance", false},
		{"Binance US", false},           // space valid
		{"FTX-backup", false},           // dash valid
		{"cold_storage", false},         // underscore valid
		{"My Platform 123", false},      // alphanumeric with space
		{"", false},                     // empty valid (optional)
		{"Platform!@#", true},           // special chars invalid
		{"Platform<script>", true},      // XSS attempt invalid
		{string(make([]byte, 51)), true}, // too long
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			err := ValidatePlatform(tt.platform)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePlatform(%q) error = %v, wantErr %v", tt.platform, err, tt.wantErr)
			}
		})
	}
}

// TestValidateNotes tests notes validation
func TestValidateNotes(t *testing.T) {
	tests := []struct {
		name    string
		notes   string
		wantErr bool
	}{
		{"empty notes", "", false},
		{"short notes", "This is a short note", false},
		{"notes at max length", string(make([]byte, MaxNotesLength)), false},
		{"notes over max length", string(make([]byte, MaxNotesLength+1)), true},
		{"unicode notes", "These are notes with émojis 🚀 and ñ characters", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotes(tt.notes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNotes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
