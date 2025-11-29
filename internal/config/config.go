package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pretty-andrechal/follyo/internal/models"
)

// Preferences holds user preferences
type Preferences struct {
	// FetchPrices controls whether to fetch live prices by default in summary
	FetchPrices *bool `json:"fetch_prices,omitempty"`
	// ColorOutput controls whether to use colored output
	ColorOutput *bool `json:"color_output,omitempty"`
	// DefaultPlatform is the default platform for new entries
	DefaultPlatform string `json:"default_platform,omitempty"`
}

// Config holds application configuration
type Config struct {
	TickerMappings map[string]string `json:"ticker_mappings"`
	Preferences    Preferences       `json:"preferences"`
}

// ConfigStore manages configuration persistence
type ConfigStore struct {
	path   string
	config *Config
	mu     sync.RWMutex
}

// New creates a new ConfigStore with the given path
func New(path string) (*ConfigStore, error) {
	cs := &ConfigStore{
		path: path,
		config: &Config{
			TickerMappings: make(map[string]string),
		},
	}

	// Ensure directory exists with restricted permissions for privacy
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	// Load existing config if it exists
	if err := cs.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return cs, nil
}

// load reads config from disk
func (cs *ConfigStore) load() error {
	data, err := os.ReadFile(cs.path)
	if err != nil {
		return err
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	if err := json.Unmarshal(data, cs.config); err != nil {
		return err
	}

	// Ensure map is initialized
	if cs.config.TickerMappings == nil {
		cs.config.TickerMappings = make(map[string]string)
	}

	return nil
}

// save writes config to disk
func (cs *ConfigStore) save() error {
	cs.mu.RLock()
	data, err := json.MarshalIndent(cs.config, "", "  ")
	cs.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(cs.path, data, 0600)
}

// GetTickerMapping returns the CoinGecko ID for a ticker, or empty string if not found
func (cs *ConfigStore) GetTickerMapping(ticker string) string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.config.TickerMappings[strings.ToUpper(ticker)]
}

// SetTickerMapping sets a ticker to CoinGecko ID mapping
func (cs *ConfigStore) SetTickerMapping(ticker, geckoID string) error {
	if err := models.ValidateCoinSymbol(ticker); err != nil {
		return fmt.Errorf("invalid ticker: %w", err)
	}
	if geckoID == "" {
		return fmt.Errorf("CoinGecko ID cannot be empty")
	}

	cs.mu.Lock()
	cs.config.TickerMappings[strings.ToUpper(ticker)] = geckoID
	cs.mu.Unlock()

	return cs.save()
}

// RemoveTickerMapping removes a ticker mapping
func (cs *ConfigStore) RemoveTickerMapping(ticker string) error {
	cs.mu.Lock()
	delete(cs.config.TickerMappings, strings.ToUpper(ticker))
	cs.mu.Unlock()

	return cs.save()
}

// GetAllTickerMappings returns all custom ticker mappings
func (cs *ConfigStore) GetAllTickerMappings() map[string]string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	// Return a copy
	result := make(map[string]string)
	for k, v := range cs.config.TickerMappings {
		result[k] = v
	}
	return result
}

// HasTickerMapping checks if a custom mapping exists for a ticker
func (cs *ConfigStore) HasTickerMapping(ticker string) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	_, ok := cs.config.TickerMappings[strings.ToUpper(ticker)]
	return ok
}

// GetPreferences returns a copy of the preferences
func (cs *ConfigStore) GetPreferences() Preferences {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.config.Preferences
}

// GetFetchPrices returns whether to fetch prices by default (true if not set)
func (cs *ConfigStore) GetFetchPrices() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	if cs.config.Preferences.FetchPrices == nil {
		return true // default to true
	}
	return *cs.config.Preferences.FetchPrices
}

// SetFetchPrices sets whether to fetch prices by default
func (cs *ConfigStore) SetFetchPrices(value bool) error {
	cs.mu.Lock()
	cs.config.Preferences.FetchPrices = &value
	cs.mu.Unlock()
	return cs.save()
}

// GetColorOutput returns whether to use colored output (true if not set)
func (cs *ConfigStore) GetColorOutput() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	if cs.config.Preferences.ColorOutput == nil {
		return true // default to true
	}
	return *cs.config.Preferences.ColorOutput
}

// SetColorOutput sets whether to use colored output
func (cs *ConfigStore) SetColorOutput(value bool) error {
	cs.mu.Lock()
	cs.config.Preferences.ColorOutput = &value
	cs.mu.Unlock()
	return cs.save()
}

// GetDefaultPlatform returns the default platform (empty if not set)
func (cs *ConfigStore) GetDefaultPlatform() string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.config.Preferences.DefaultPlatform
}

// SetDefaultPlatform sets the default platform
func (cs *ConfigStore) SetDefaultPlatform(platform string) error {
	if err := models.ValidatePlatform(platform); err != nil {
		return fmt.Errorf("invalid platform: %w", err)
	}

	cs.mu.Lock()
	cs.config.Preferences.DefaultPlatform = platform
	cs.mu.Unlock()
	return cs.save()
}

// ClearDefaultPlatform removes the default platform
func (cs *ConfigStore) ClearDefaultPlatform() error {
	cs.mu.Lock()
	cs.config.Preferences.DefaultPlatform = ""
	cs.mu.Unlock()
	return cs.save()
}
