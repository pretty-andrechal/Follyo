package portfolio

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pretty-andrechal/follyo/internal/models"
)

// CreateSnapshot creates a new portfolio snapshot with the given prices.
// Returns an error if any coin in the portfolio is missing from the prices map.
func (p *Portfolio) CreateSnapshot(prices map[string]float64, note string) (models.Snapshot, error) {
	summary, err := p.GetSummary()
	if err != nil {
		return models.Snapshot{}, err
	}

	// Validate all coins have prices before calculating
	var missingPrices []string
	for coin := range summary.HoldingsByCoin {
		if _, exists := prices[coin]; !exists {
			missingPrices = append(missingPrices, coin)
		}
	}
	for coin := range summary.LoansByCoin {
		if _, exists := prices[coin]; !exists {
			// Avoid duplicates
			found := false
			for _, m := range missingPrices {
				if m == coin {
					found = true
					break
				}
			}
			if !found {
				missingPrices = append(missingPrices, coin)
			}
		}
	}
	if len(missingPrices) > 0 {
		return models.Snapshot{}, fmt.Errorf("missing prices for coins: %v", missingPrices)
	}

	// Calculate values for each coin
	coinValues := make(map[string]models.CoinSnapshot)
	var holdingsValue float64

	for coin, amount := range summary.HoldingsByCoin {
		price := prices[coin] // Safe: validated above
		value := amount * price
		holdingsValue += value

		coinValues[coin] = models.CoinSnapshot{
			Amount: amount,
			Price:  price,
			Value:  value,
		}
	}

	// Calculate loans value
	var loansValue float64
	for coin, amount := range summary.LoansByCoin {
		price := prices[coin] // Safe: validated above
		value := amount * price
		loansValue += value
	}

	// Calculate net value and profit/loss
	netValue := holdingsValue - loansValue
	profitLoss := netValue - summary.TotalInvestedUSD + summary.TotalSoldUSD

	var profitPercent float64
	if summary.TotalInvestedUSD > 0 {
		profitPercent = (profitLoss / summary.TotalInvestedUSD) * 100
	}

	snapshot := models.Snapshot{
		ID:            uuid.New().String()[:models.IDLength],
		Timestamp:     time.Now(),
		HoldingsValue: holdingsValue,
		LoansValue:    loansValue,
		NetValue:      netValue,
		TotalInvested: summary.TotalInvestedUSD,
		TotalSold:     summary.TotalSoldUSD,
		ProfitLoss:    profitLoss,
		ProfitPercent: profitPercent,
		CoinValues:    coinValues,
		Note:          note,
	}

	return snapshot, nil
}

// SnapshotComparison represents the difference between two snapshots
type SnapshotComparison struct {
	OlderSnapshot    *models.Snapshot
	NewerSnapshot    *models.Snapshot
	NetValueChange   float64
	NetValuePercent  float64
	ProfitLossChange float64
	CoinChanges      map[string]CoinChange
}

// CoinChange represents the change in a coin between snapshots
type CoinChange struct {
	OldAmount   float64
	NewAmount   float64
	OldPrice    float64
	NewPrice    float64
	OldValue    float64
	NewValue    float64
	ValueChange float64
}

// CompareSnapshots compares two snapshots and returns the differences
func CompareSnapshots(older, newer *models.Snapshot) SnapshotComparison {
	comparison := SnapshotComparison{
		OlderSnapshot:    older,
		NewerSnapshot:    newer,
		NetValueChange:   newer.NetValue - older.NetValue,
		ProfitLossChange: newer.ProfitLoss - older.ProfitLoss,
		CoinChanges:      make(map[string]CoinChange),
	}

	if older.NetValue > 0 {
		comparison.NetValuePercent = (comparison.NetValueChange / older.NetValue) * 100
	}

	// Collect all coins from both snapshots
	allCoins := make(map[string]bool)
	for coin := range older.CoinValues {
		allCoins[coin] = true
	}
	for coin := range newer.CoinValues {
		allCoins[coin] = true
	}

	// Calculate changes for each coin
	for coin := range allCoins {
		oldSnap := older.CoinValues[coin]
		newSnap := newer.CoinValues[coin]

		comparison.CoinChanges[coin] = CoinChange{
			OldAmount:   oldSnap.Amount,
			NewAmount:   newSnap.Amount,
			OldPrice:    oldSnap.Price,
			NewPrice:    newSnap.Price,
			OldValue:    oldSnap.Value,
			NewValue:    newSnap.Value,
			ValueChange: newSnap.Value - oldSnap.Value,
		}
	}

	return comparison
}
