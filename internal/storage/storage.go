package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pretty-andrechal/follyo/internal/models"
)

// PortfolioData represents the structure of the JSON file.
type PortfolioData struct {
	Holdings []models.Holding `json:"holdings"`
	Loans    []models.Loan    `json:"loans"`
	Sales    []models.Sale    `json:"sales"`
	Stakes   []models.Stake   `json:"stakes"`
}

// Storage handles persistence of portfolio data to JSON.
type Storage struct {
	dataPath string
}

// New creates a new Storage instance.
func New(dataPath string) (*Storage, error) {
	s := &Storage{dataPath: dataPath}
	if err := s.ensureDataFile(); err != nil {
		return nil, err
	}
	return s, nil
}

// DefaultDataPath returns the default path for portfolio data.
func DefaultDataPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "data/portfolio.json"
	}
	return filepath.Join(filepath.Dir(exe), "data", "portfolio.json")
}

func (s *Storage) ensureDataFile() error {
	dir := filepath.Dir(s.dataPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if _, err := os.Stat(s.dataPath); os.IsNotExist(err) {
		data := PortfolioData{
			Holdings: []models.Holding{},
			Loans:    []models.Loan{},
			Sales:    []models.Sale{},
			Stakes:   []models.Stake{},
		}
		return s.saveData(data)
	}
	return nil
}

func (s *Storage) loadData() (PortfolioData, error) {
	var data PortfolioData

	file, err := os.ReadFile(s.dataPath)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(file, &data)
	return data, err
}

func (s *Storage) saveData(data PortfolioData) error {
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.dataPath, file, 0644)
}

// Holdings operations

// GetHoldings returns all holdings.
func (s *Storage) GetHoldings() ([]models.Holding, error) {
	data, err := s.loadData()
	if err != nil {
		return nil, err
	}
	return data.Holdings, nil
}

// AddHolding adds a new holding.
func (s *Storage) AddHolding(holding models.Holding) error {
	data, err := s.loadData()
	if err != nil {
		return err
	}
	data.Holdings = append(data.Holdings, holding)
	return s.saveData(data)
}

// RemoveHolding removes a holding by ID.
func (s *Storage) RemoveHolding(id string) (bool, error) {
	data, err := s.loadData()
	if err != nil {
		return false, err
	}

	originalLen := len(data.Holdings)
	filtered := make([]models.Holding, 0, len(data.Holdings))
	for _, h := range data.Holdings {
		if h.ID != id {
			filtered = append(filtered, h)
		}
	}
	data.Holdings = filtered

	if len(data.Holdings) < originalLen {
		return true, s.saveData(data)
	}
	return false, nil
}

// Loans operations

// GetLoans returns all loans.
func (s *Storage) GetLoans() ([]models.Loan, error) {
	data, err := s.loadData()
	if err != nil {
		return nil, err
	}
	return data.Loans, nil
}

// AddLoan adds a new loan.
func (s *Storage) AddLoan(loan models.Loan) error {
	data, err := s.loadData()
	if err != nil {
		return err
	}
	data.Loans = append(data.Loans, loan)
	return s.saveData(data)
}

// RemoveLoan removes a loan by ID.
func (s *Storage) RemoveLoan(id string) (bool, error) {
	data, err := s.loadData()
	if err != nil {
		return false, err
	}

	originalLen := len(data.Loans)
	filtered := make([]models.Loan, 0, len(data.Loans))
	for _, l := range data.Loans {
		if l.ID != id {
			filtered = append(filtered, l)
		}
	}
	data.Loans = filtered

	if len(data.Loans) < originalLen {
		return true, s.saveData(data)
	}
	return false, nil
}

// Sales operations

// GetSales returns all sales.
func (s *Storage) GetSales() ([]models.Sale, error) {
	data, err := s.loadData()
	if err != nil {
		return nil, err
	}
	return data.Sales, nil
}

// AddSale adds a new sale.
func (s *Storage) AddSale(sale models.Sale) error {
	data, err := s.loadData()
	if err != nil {
		return err
	}
	data.Sales = append(data.Sales, sale)
	return s.saveData(data)
}

// RemoveSale removes a sale by ID.
func (s *Storage) RemoveSale(id string) (bool, error) {
	data, err := s.loadData()
	if err != nil {
		return false, err
	}

	originalLen := len(data.Sales)
	filtered := make([]models.Sale, 0, len(data.Sales))
	for _, sl := range data.Sales {
		if sl.ID != id {
			filtered = append(filtered, sl)
		}
	}
	data.Sales = filtered

	if len(data.Sales) < originalLen {
		return true, s.saveData(data)
	}
	return false, nil
}

// Stakes operations

// GetStakes returns all stakes.
func (s *Storage) GetStakes() ([]models.Stake, error) {
	data, err := s.loadData()
	if err != nil {
		return nil, err
	}
	return data.Stakes, nil
}

// AddStake adds a new stake.
func (s *Storage) AddStake(stake models.Stake) error {
	data, err := s.loadData()
	if err != nil {
		return err
	}
	if data.Stakes == nil {
		data.Stakes = []models.Stake{}
	}
	data.Stakes = append(data.Stakes, stake)
	return s.saveData(data)
}

// RemoveStake removes a stake by ID.
func (s *Storage) RemoveStake(id string) (bool, error) {
	data, err := s.loadData()
	if err != nil {
		return false, err
	}

	originalLen := len(data.Stakes)
	filtered := make([]models.Stake, 0, len(data.Stakes))
	for _, st := range data.Stakes {
		if st.ID != id {
			filtered = append(filtered, st)
		}
	}
	data.Stakes = filtered

	if len(data.Stakes) < originalLen {
		return true, s.saveData(data)
	}
	return false, nil
}
