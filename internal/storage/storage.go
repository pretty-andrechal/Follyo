package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

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
// Data is cached in memory after first load to reduce disk I/O.
type Storage struct {
	dataPath string
	data     *PortfolioData
	mu       sync.RWMutex
}

// New creates a new Storage instance.
func New(dataPath string) (*Storage, error) {
	s := &Storage{dataPath: dataPath}
	if err := s.ensureDataFile(); err != nil {
		return nil, fmt.Errorf("ensuring data file: %w", err)
	}
	// Load data into memory
	if err := s.loadData(); err != nil {
		return nil, fmt.Errorf("loading data: %w", err)
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
	// Use 0700 for privacy - portfolio data should only be readable by owner
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	if _, err := os.Stat(s.dataPath); os.IsNotExist(err) {
		s.data = &PortfolioData{
			Holdings: []models.Holding{},
			Loans:    []models.Loan{},
			Sales:    []models.Sale{},
			Stakes:   []models.Stake{},
		}
		return s.saveData()
	}
	return nil
}

func (s *Storage) loadData() error {
	file, err := os.ReadFile(s.dataPath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = &PortfolioData{}
	if err := json.Unmarshal(file, s.data); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	// Ensure slices are initialized (not nil)
	if s.data.Holdings == nil {
		s.data.Holdings = []models.Holding{}
	}
	if s.data.Loans == nil {
		s.data.Loans = []models.Loan{}
	}
	if s.data.Sales == nil {
		s.data.Sales = []models.Sale{}
	}
	if s.data.Stakes == nil {
		s.data.Stakes = []models.Stake{}
	}

	return nil
}

func (s *Storage) saveData() error {
	s.mu.RLock()
	file, err := json.MarshalIndent(s.data, "", "  ")
	s.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	if err := os.WriteFile(s.dataPath, file, 0600); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}

// Holdings operations

// GetHoldings returns all holdings.
func (s *Storage) GetHoldings() ([]models.Holding, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]models.Holding, len(s.data.Holdings))
	copy(result, s.data.Holdings)
	return result, nil
}

// AddHolding adds a new holding.
func (s *Storage) AddHolding(holding models.Holding) error {
	s.mu.Lock()
	s.data.Holdings = append(s.data.Holdings, holding)
	s.mu.Unlock()

	return s.saveData()
}

// RemoveHolding removes a holding by ID.
func (s *Storage) RemoveHolding(id string) (bool, error) {
	s.mu.Lock()
	originalLen := len(s.data.Holdings)
	filtered := make([]models.Holding, 0, len(s.data.Holdings))
	for _, h := range s.data.Holdings {
		if h.ID != id {
			filtered = append(filtered, h)
		}
	}
	s.data.Holdings = filtered
	removed := len(s.data.Holdings) < originalLen
	s.mu.Unlock()

	if removed {
		return true, s.saveData()
	}
	return false, nil
}

// Loans operations

// GetLoans returns all loans.
func (s *Storage) GetLoans() ([]models.Loan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.Loan, len(s.data.Loans))
	copy(result, s.data.Loans)
	return result, nil
}

// AddLoan adds a new loan.
func (s *Storage) AddLoan(loan models.Loan) error {
	s.mu.Lock()
	s.data.Loans = append(s.data.Loans, loan)
	s.mu.Unlock()

	return s.saveData()
}

// RemoveLoan removes a loan by ID.
func (s *Storage) RemoveLoan(id string) (bool, error) {
	s.mu.Lock()
	originalLen := len(s.data.Loans)
	filtered := make([]models.Loan, 0, len(s.data.Loans))
	for _, l := range s.data.Loans {
		if l.ID != id {
			filtered = append(filtered, l)
		}
	}
	s.data.Loans = filtered
	removed := len(s.data.Loans) < originalLen
	s.mu.Unlock()

	if removed {
		return true, s.saveData()
	}
	return false, nil
}

// Sales operations

// GetSales returns all sales.
func (s *Storage) GetSales() ([]models.Sale, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.Sale, len(s.data.Sales))
	copy(result, s.data.Sales)
	return result, nil
}

// AddSale adds a new sale.
func (s *Storage) AddSale(sale models.Sale) error {
	s.mu.Lock()
	s.data.Sales = append(s.data.Sales, sale)
	s.mu.Unlock()

	return s.saveData()
}

// RemoveSale removes a sale by ID.
func (s *Storage) RemoveSale(id string) (bool, error) {
	s.mu.Lock()
	originalLen := len(s.data.Sales)
	filtered := make([]models.Sale, 0, len(s.data.Sales))
	for _, sl := range s.data.Sales {
		if sl.ID != id {
			filtered = append(filtered, sl)
		}
	}
	s.data.Sales = filtered
	removed := len(s.data.Sales) < originalLen
	s.mu.Unlock()

	if removed {
		return true, s.saveData()
	}
	return false, nil
}

// Stakes operations

// GetStakes returns all stakes.
func (s *Storage) GetStakes() ([]models.Stake, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.Stake, len(s.data.Stakes))
	copy(result, s.data.Stakes)
	return result, nil
}

// AddStake adds a new stake.
func (s *Storage) AddStake(stake models.Stake) error {
	s.mu.Lock()
	s.data.Stakes = append(s.data.Stakes, stake)
	s.mu.Unlock()

	return s.saveData()
}

// RemoveStake removes a stake by ID.
func (s *Storage) RemoveStake(id string) (bool, error) {
	s.mu.Lock()
	originalLen := len(s.data.Stakes)
	filtered := make([]models.Stake, 0, len(s.data.Stakes))
	for _, st := range s.data.Stakes {
		if st.ID != id {
			filtered = append(filtered, st)
		}
	}
	s.data.Stakes = filtered
	removed := len(s.data.Stakes) < originalLen
	s.mu.Unlock()

	if removed {
		return true, s.saveData()
	}
	return false, nil
}
