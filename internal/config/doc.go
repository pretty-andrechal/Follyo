// Package config handles application configuration loading and management.
//
// Configuration is stored in a YAML file (default: ~/.follyo/config.yaml) and
// includes settings for:
//   - Data directory path
//   - Default exchange/platform for transactions
//   - Custom ticker-to-CoinGecko ID mappings
//   - Display preferences
//
// # Loading Configuration
//
// Use [Load] to read configuration from the default path, or [LoadFromPath]
// for a custom location. If the config file doesn't exist, default values
// are used.
//
// # Environment Variables
//
// The data directory can be overridden with the FOLLYO_DATA_DIR environment
// variable.
package config
