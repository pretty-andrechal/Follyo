package prices

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// PriceService fetches cryptocurrency prices
type PriceService struct {
	client      *http.Client
	cache       map[string]cachedPrice
	cacheMu     sync.RWMutex
	cacheTTL    time.Duration
	coinIDMap   map[string]string // maps ticker (BTC) to CoinGecko ID (bitcoin)
	lastRequest time.Time
	rateMu      sync.Mutex
	minInterval time.Duration // minimum time between API requests
}

type cachedPrice struct {
	price     float64
	fetchedAt time.Time
}

// Common ticker to CoinGecko ID mappings
var defaultCoinIDMap = map[string]string{
	"BTC":   "bitcoin",
	"ETH":   "ethereum",
	"USDT":  "tether",
	"USDC":  "usd-coin",
	"BNB":   "binancecoin",
	"XRP":   "ripple",
	"ADA":   "cardano",
	"DOGE":  "dogecoin",
	"SOL":   "solana",
	"DOT":   "polkadot",
	"MATIC": "matic-network",
	"LTC":   "litecoin",
	"SHIB":  "shiba-inu",
	"TRX":   "tron",
	"AVAX":  "avalanche-2",
	"LINK":  "chainlink",
	"ATOM":  "cosmos",
	"UNI":   "uniswap",
	"XLM":   "stellar",
	"ETC":   "ethereum-classic",
	"ALGO":  "algorand",
	"NEAR":  "near",
	"FTM":   "fantom",
	"APE":   "apecoin",
	"MANA":  "decentraland",
	"SAND":  "the-sandbox",
	"AXS":   "axie-infinity",
	"CRO":   "crypto-com-chain",
	"AAVE":  "aave",
	"MKR":   "maker",
	"COMP":  "compound-governance-token",
	"SNX":   "havven",
	"YFI":   "yearn-finance",
	"SUSHI": "sushi",
	"ZEC":   "zcash",
	"DASH":  "dash",
	"XMR":   "monero",
	"NEO":   "neo",
	"EOS":   "eos",
	"XTZ":   "tezos",
	"THETA": "theta-token",
	"VET":   "vechain",
	"FIL":   "filecoin",
	"ICP":   "internet-computer",
	"HBAR":  "hedera-hashgraph",
	"EGLD":  "elrond-erd-2",
	"FLOW":  "flow",
	"KCS":   "kucoin-shares",
	"HT":    "huobi-token",
	"OKB":   "okb",
	"LEO":   "leo-token",
	"QNT":   "quant-network",
	"ARB":   "arbitrum",
	"OP":    "optimism",
	"APT":   "aptos",
	"SUI":   "sui",
	"INJ":   "injective-protocol",
	"IMX":   "immutable-x",
	"STX":   "blockstack",
	"RNDR":  "render-token",
	"GRT":   "the-graph",
	"LDO":   "lido-dao",
	"RPL":   "rocket-pool",
	"PEPE":  "pepe",
	"WLD":   "worldcoin-wld",
	"SEI":   "sei-network",
	"TIA":   "celestia",
	"MUTE":  "mute", // zkSync token
}

// Default rate limit: 2 seconds between requests (~30 req/min for CoinGecko free tier)
const defaultMinInterval = 2 * time.Second

// Default cache TTL
const defaultCacheTTL = 2 * time.Minute

// Default HTTP timeout
const defaultHTTPTimeout = 10 * time.Second

// Option is a functional option for configuring PriceService
type Option func(*PriceService)

// WithCacheTTL sets the cache time-to-live duration
func WithCacheTTL(ttl time.Duration) Option {
	return func(ps *PriceService) {
		ps.cacheTTL = ttl
	}
}

// WithRateLimit sets the minimum interval between API requests
func WithRateLimit(interval time.Duration) Option {
	return func(ps *PriceService) {
		ps.minInterval = interval
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(ps *PriceService) {
		ps.client = client
	}
}

// WithCoinMappings adds custom ticker to CoinGecko ID mappings
func WithCoinMappings(mappings map[string]string) Option {
	return func(ps *PriceService) {
		for ticker, geckoID := range mappings {
			ps.coinIDMap[strings.ToUpper(ticker)] = geckoID
		}
	}
}

// New creates a new PriceService with the given options.
// Default settings: 2 minute cache TTL, 2 second rate limit, 10 second HTTP timeout.
func New(opts ...Option) *PriceService {
	// Create with defaults
	ps := &PriceService{
		client: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		cache:       make(map[string]cachedPrice),
		cacheTTL:    defaultCacheTTL,
		coinIDMap:   copyDefaultMappings(),
		minInterval: defaultMinInterval,
	}

	// Apply options
	for _, opt := range opts {
		opt(ps)
	}

	return ps
}

// copyDefaultMappings returns a copy of the default coin ID map
func copyDefaultMappings() map[string]string {
	result := make(map[string]string, len(defaultCoinIDMap))
	for k, v := range defaultCoinIDMap {
		result[k] = v
	}
	return result
}

// NewWithClient creates a PriceService with a custom HTTP client (for testing).
// Deprecated: Use New(WithHTTPClient(client)) instead.
func NewWithClient(client *http.Client) *PriceService {
	return New(WithHTTPClient(client))
}

// SetCacheTTL sets the cache time-to-live duration
func (ps *PriceService) SetCacheTTL(ttl time.Duration) {
	ps.cacheTTL = ttl
}

// SetRateLimit sets the minimum interval between API requests
func (ps *PriceService) SetRateLimit(interval time.Duration) {
	ps.rateMu.Lock()
	ps.minInterval = interval
	ps.rateMu.Unlock()
}

// waitForRateLimit blocks until it's safe to make another API request
func (ps *PriceService) waitForRateLimit() {
	ps.rateMu.Lock()
	defer ps.rateMu.Unlock()

	elapsed := time.Since(ps.lastRequest)
	if elapsed < ps.minInterval {
		time.Sleep(ps.minInterval - elapsed)
	}
	ps.lastRequest = time.Now()
}

// AddCoinMapping adds a custom ticker to CoinGecko ID mapping
func (ps *PriceService) AddCoinMapping(ticker, geckoID string) {
	ps.coinIDMap[strings.ToUpper(ticker)] = geckoID
}

// GetPrice fetches the current USD price for a single coin.
// Uses a background context. For cancellation support, use GetPriceWithContext.
func (ps *PriceService) GetPrice(ticker string) (float64, error) {
	return ps.GetPriceWithContext(context.Background(), ticker)
}

// GetPriceWithContext fetches the current USD price for a single coin with context support.
func (ps *PriceService) GetPriceWithContext(ctx context.Context, ticker string) (float64, error) {
	prices, err := ps.GetPricesWithContext(ctx, []string{ticker})
	if err != nil {
		return 0, err
	}
	price, ok := prices[strings.ToUpper(ticker)]
	if !ok {
		return 0, fmt.Errorf("price not found for %s", ticker)
	}
	return price, nil
}

// GetPrices fetches current USD prices for multiple coins.
// Uses a background context. For cancellation support, use GetPricesWithContext.
func (ps *PriceService) GetPrices(tickers []string) (map[string]float64, error) {
	return ps.GetPricesWithContext(context.Background(), tickers)
}

// GetPricesWithContext fetches current USD prices for multiple coins with context support.
// Returns a map of ticker -> price. The context can be used for cancellation and timeouts.
func (ps *PriceService) GetPricesWithContext(ctx context.Context, tickers []string) (map[string]float64, error) {
	result := make(map[string]float64)
	var toFetch []string
	tickerToGeckoID := make(map[string]string)

	// Check cache first
	ps.cacheMu.RLock()
	for _, ticker := range tickers {
		upperTicker := strings.ToUpper(ticker)
		if cached, ok := ps.cache[upperTicker]; ok {
			if time.Since(cached.fetchedAt) < ps.cacheTTL {
				result[upperTicker] = cached.price
				continue
			}
		}
		// Need to fetch this one
		geckoID, ok := ps.coinIDMap[upperTicker]
		if !ok {
			// Try lowercase ticker as gecko ID
			geckoID = strings.ToLower(upperTicker)
		}
		toFetch = append(toFetch, geckoID)
		tickerToGeckoID[upperTicker] = geckoID
	}
	ps.cacheMu.RUnlock()

	// If everything was cached, return early
	if len(toFetch) == 0 {
		return result, nil
	}

	// Check context before making API call
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Fetch from CoinGecko
	prices, err := ps.fetchFromCoinGeckoWithContext(ctx, toFetch)
	if err != nil {
		return nil, err
	}

	// Map gecko IDs back to tickers and update cache
	ps.cacheMu.Lock()
	for ticker, geckoID := range tickerToGeckoID {
		if price, ok := prices[geckoID]; ok {
			result[ticker] = price
			ps.cache[ticker] = cachedPrice{
				price:     price,
				fetchedAt: time.Now(),
			}
		}
	}
	ps.cacheMu.Unlock()

	return result, nil
}

// fetchFromCoinGecko fetches prices from the CoinGecko API.
// Uses a background context. For cancellation support, use fetchFromCoinGeckoWithContext.
func (ps *PriceService) fetchFromCoinGecko(geckoIDs []string) (map[string]float64, error) {
	return ps.fetchFromCoinGeckoWithContext(context.Background(), geckoIDs)
}

// fetchFromCoinGeckoWithContext fetches prices from the CoinGecko API with context support.
func (ps *PriceService) fetchFromCoinGeckoWithContext(ctx context.Context, geckoIDs []string) (map[string]float64, error) {
	if len(geckoIDs) == 0 {
		return make(map[string]float64), nil
	}

	// Build URL
	baseURL := "https://api.coingecko.com/api/v3/simple/price"
	params := url.Values{}
	params.Set("ids", strings.Join(geckoIDs, ","))
	params.Set("vs_currencies", "usd")

	reqURL := baseURL + "?" + params.Encode()

	// Wait for rate limit before making request
	ps.waitForRateLimit()

	// Check context after rate limit wait
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Make request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := ps.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CoinGecko API returned status %d", resp.StatusCode)
	}

	// Parse response
	// Response format: {"bitcoin":{"usd":97000},"ethereum":{"usd":3400}}
	var data map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse price response: %w", err)
	}

	// Extract USD prices
	result := make(map[string]float64)
	for geckoID, currencies := range data {
		if usdPrice, ok := currencies["usd"]; ok {
			result[geckoID] = usdPrice
		}
	}

	return result, nil
}

// ClearCache clears the price cache
func (ps *PriceService) ClearCache() {
	ps.cacheMu.Lock()
	ps.cache = make(map[string]cachedPrice)
	ps.cacheMu.Unlock()
}

// GetCoinGeckoID returns the CoinGecko ID for a ticker, or empty string if unknown
func (ps *PriceService) GetCoinGeckoID(ticker string) string {
	return ps.coinIDMap[strings.ToUpper(ticker)]
}

// HasMapping checks if a ticker has a mapping (either default or custom)
func (ps *PriceService) HasMapping(ticker string) bool {
	_, ok := ps.coinIDMap[strings.ToUpper(ticker)]
	return ok
}

// GetUnmappedTickers returns tickers that don't have a CoinGecko mapping
func (ps *PriceService) GetUnmappedTickers(tickers []string) []string {
	var unmapped []string
	for _, ticker := range tickers {
		if !ps.HasMapping(ticker) {
			unmapped = append(unmapped, strings.ToUpper(ticker))
		}
	}
	return unmapped
}

// GetDefaultMappings returns a copy of the default ticker mappings
func GetDefaultMappings() map[string]string {
	result := make(map[string]string)
	for k, v := range defaultCoinIDMap {
		result[k] = v
	}
	return result
}

// SearchResult represents a coin from CoinGecko search
type SearchResult struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Rank   int    `json:"market_cap_rank"`
}

// SearchCoins searches CoinGecko for coins matching the query.
// Uses a background context. For cancellation support, use SearchCoinsWithContext.
func (ps *PriceService) SearchCoins(query string) ([]SearchResult, error) {
	return ps.SearchCoinsWithContext(context.Background(), query)
}

// SearchCoinsWithContext searches CoinGecko for coins matching the query with context support.
func (ps *PriceService) SearchCoinsWithContext(ctx context.Context, query string) ([]SearchResult, error) {
	baseURL := "https://api.coingecko.com/api/v3/search"
	params := url.Values{}
	params.Set("query", query)

	reqURL := baseURL + "?" + params.Encode()

	// Wait for rate limit before making request
	ps.waitForRateLimit()

	// Check context after rate limit wait
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Make request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := ps.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search coins: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CoinGecko API returned status %d", resp.StatusCode)
	}

	// Response format: {"coins":[{"id":"bitcoin","name":"Bitcoin","symbol":"btc","market_cap_rank":1},...]}
	var data struct {
		Coins []SearchResult `json:"coins"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Limit to top 10 results
	if len(data.Coins) > 10 {
		data.Coins = data.Coins[:10]
	}

	return data.Coins, nil
}
