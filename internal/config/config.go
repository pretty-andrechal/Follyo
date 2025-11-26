package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Config holds application configuration
type Config struct {
	TickerMappings map[string]string `json:"ticker_mappings"`
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

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
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

	return os.WriteFile(cs.path, data, 0644)
}

// GetTickerMapping returns the CoinGecko ID for a ticker, or empty string if not found
func (cs *ConfigStore) GetTickerMapping(ticker string) string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.config.TickerMappings[strings.ToUpper(ticker)]
}

// SetTickerMapping sets a ticker to CoinGecko ID mapping
func (cs *ConfigStore) SetTickerMapping(ticker, geckoID string) error {
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
