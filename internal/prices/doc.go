// Package prices provides cryptocurrency price fetching from CoinGecko API.
//
// The [PriceService] fetches current USD prices for multiple coins in a single
// API call. It supports custom ticker-to-CoinGecko ID mappings for coins that
// use non-standard identifiers.
//
// # Usage
//
//	ps := prices.New()
//	ps.AddCoinMapping("MATIC", "polygon")  // Custom mapping
//	prices, err := ps.GetPrices([]string{"BTC", "ETH", "MATIC"})
//
// # Stablecoins
//
// Common stablecoins (USDT, USDC, DAI, etc.) are automatically mapped to
// return a price of $1.00 without making API calls.
//
// # Rate Limiting
//
// The CoinGecko free API has rate limits. For production use, consider
// implementing caching or using a paid API key.
package prices
