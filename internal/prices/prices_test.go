package prices

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetPrices(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		ids := r.URL.Query().Get("ids")
		if ids == "" {
			t.Error("Expected ids parameter")
		}
		vsCurrencies := r.URL.Query().Get("vs_currencies")
		if vsCurrencies != "usd" {
			t.Errorf("Expected vs_currencies=usd, got %s", vsCurrencies)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"bitcoin":{"usd":97000.50},"ethereum":{"usd":3400.25}}`))
	}))
	defer server.Close()

	// Create price service with mock client that redirects to test server
	ps := NewWithClient(&http.Client{
		Transport: &mockTransport{server.URL},
	})

	prices, err := ps.GetPrices([]string{"BTC", "ETH"})
	if err != nil {
		t.Fatalf("GetPrices failed: %v", err)
	}

	if prices["BTC"] != 97000.50 {
		t.Errorf("Expected BTC price 97000.50, got %f", prices["BTC"])
	}
	if prices["ETH"] != 3400.25 {
		t.Errorf("Expected ETH price 3400.25, got %f", prices["ETH"])
	}
}

func TestGetPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"bitcoin":{"usd":95000}}`))
	}))
	defer server.Close()

	ps := NewWithClient(&http.Client{
		Transport: &mockTransport{server.URL},
	})

	price, err := ps.GetPrice("BTC")
	if err != nil {
		t.Fatalf("GetPrice failed: %v", err)
	}

	if price != 95000 {
		t.Errorf("Expected price 95000, got %f", price)
	}
}

func TestCaching(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"bitcoin":{"usd":97000}}`))
	}))
	defer server.Close()

	ps := NewWithClient(&http.Client{
		Transport: &mockTransport{server.URL},
	})
	ps.SetCacheTTL(1 * time.Hour) // Long TTL for test

	// First call should hit the server
	_, err := ps.GetPrice("BTC")
	if err != nil {
		t.Fatalf("First GetPrice failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 API call, got %d", callCount)
	}

	// Second call should use cache
	_, err = ps.GetPrice("BTC")
	if err != nil {
		t.Fatalf("Second GetPrice failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected still 1 API call (cached), got %d", callCount)
	}
}

func TestCacheExpiry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"bitcoin":{"usd":97000}}`))
	}))
	defer server.Close()

	ps := NewWithClient(&http.Client{
		Transport: &mockTransport{server.URL},
	})
	ps.SetCacheTTL(1 * time.Millisecond) // Very short TTL

	// First call
	_, err := ps.GetPrice("BTC")
	if err != nil {
		t.Fatalf("First GetPrice failed: %v", err)
	}

	// Wait for cache to expire
	time.Sleep(5 * time.Millisecond)

	// Second call should hit server again
	_, err = ps.GetPrice("BTC")
	if err != nil {
		t.Fatalf("Second GetPrice failed: %v", err)
	}
	if callCount != 2 {
		t.Errorf("Expected 2 API calls after cache expiry, got %d", callCount)
	}
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	ps := NewWithClient(&http.Client{
		Transport: &mockTransport{server.URL},
	})

	_, err := ps.GetPrice("BTC")
	if err == nil {
		t.Error("Expected error for 429 response")
	}
}

func TestAddCoinMapping(t *testing.T) {
	ps := New()

	// Add custom mapping
	ps.AddCoinMapping("MYCOIN", "my-custom-coin")

	geckoID := ps.GetCoinGeckoID("MYCOIN")
	if geckoID != "my-custom-coin" {
		t.Errorf("Expected my-custom-coin, got %s", geckoID)
	}

	// Test case insensitivity
	geckoID = ps.GetCoinGeckoID("mycoin")
	if geckoID != "my-custom-coin" {
		t.Errorf("Expected my-custom-coin for lowercase, got %s", geckoID)
	}
}

func TestClearCache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"bitcoin":{"usd":97000}}`))
	}))
	defer server.Close()

	ps := NewWithClient(&http.Client{
		Transport: &mockTransport{server.URL},
	})
	ps.SetCacheTTL(1 * time.Hour)

	// First call
	_, _ = ps.GetPrice("BTC")
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}

	// Clear cache
	ps.ClearCache()

	// Should hit server again
	_, _ = ps.GetPrice("BTC")
	if callCount != 2 {
		t.Errorf("Expected 2 calls after cache clear, got %d", callCount)
	}
}

func TestDefaultCoinMappings(t *testing.T) {
	ps := New()

	// Test some common mappings
	mappings := map[string]string{
		"BTC":   "bitcoin",
		"ETH":   "ethereum",
		"USDT":  "tether",
		"SOL":   "solana",
		"MATIC": "matic-network",
	}

	for ticker, expectedID := range mappings {
		geckoID := ps.GetCoinGeckoID(ticker)
		if geckoID != expectedID {
			t.Errorf("Expected %s for %s, got %s", expectedID, ticker, geckoID)
		}
	}
}

// TestRateLimitSetting tests the rate limit configuration
func TestRateLimitSetting(t *testing.T) {
	ps := New()

	// Set a short rate limit for testing
	ps.SetRateLimit(100 * time.Millisecond)

	// Verify internal state (using a simple timing check)
	start := time.Now()
	ps.waitForRateLimit()
	first := time.Since(start)

	start = time.Now()
	ps.waitForRateLimit()
	second := time.Since(start)

	// First call should be immediate, second should wait
	if first > 50*time.Millisecond {
		t.Errorf("First rate limit call took too long: %v", first)
	}
	if second < 80*time.Millisecond {
		t.Errorf("Second rate limit call should have waited ~100ms, took: %v", second)
	}
}

// TestNetworkConnectionRefused tests behavior when the server is unreachable
func TestNetworkConnectionRefused(t *testing.T) {
	// Use a port that's not listening
	ps := NewWithClient(&http.Client{
		Timeout:   1 * time.Second,
		Transport: &mockTransport{"http://127.0.0.1:1"}, // Port 1 is privileged and unlikely to be listening
	})

	_, err := ps.GetPrice("BTC")
	if err == nil {
		t.Error("Expected connection error, got nil")
	}
}

// TestAPIServerError tests handling of 500 Internal Server Error
func TestAPIServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	ps := NewWithClient(&http.Client{
		Transport: &mockTransport{server.URL},
	})

	_, err := ps.GetPrice("BTC")
	if err == nil {
		t.Error("Expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected error to mention status 500, got: %v", err)
	}
}

// TestMalformedJSONResponse tests handling of invalid JSON responses
func TestMalformedJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	ps := NewWithClient(&http.Client{
		Transport: &mockTransport{server.URL},
	})

	_, err := ps.GetPrice("BTC")
	if err == nil {
		t.Error("Expected error for malformed JSON response")
	}
}

// TestEmptyResponse tests handling of empty API response
func TestEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	ps := NewWithClient(&http.Client{
		Transport: &mockTransport{server.URL},
	})

	prices, err := ps.GetPrices([]string{"BTC"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// BTC should not be in the result since API returned no data for it
	if _, ok := prices["BTC"]; ok {
		t.Error("Expected BTC to not be in prices when API returns empty response")
	}
}

// TestSearchCoinsNetworkError tests SearchCoins with network error
func TestSearchCoinsNetworkError(t *testing.T) {
	ps := NewWithClient(&http.Client{
		Timeout:   1 * time.Second,
		Transport: &mockTransport{"http://127.0.0.1:1"},
	})

	_, err := ps.SearchCoins("bitcoin")
	if err == nil {
		t.Error("Expected connection error, got nil")
	}
}

// TestSearchCoinsAPIError tests SearchCoins with API error
func TestSearchCoinsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	ps := NewWithClient(&http.Client{
		Transport: &mockTransport{server.URL},
	})

	_, err := ps.SearchCoins("bitcoin")
	if err == nil {
		t.Error("Expected error for 503 response")
	}
}

// TestSearchCoinsSuccess tests successful search
func TestSearchCoinsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"coins":[{"id":"bitcoin","name":"Bitcoin","symbol":"btc","market_cap_rank":1},{"id":"bitcoin-cash","name":"Bitcoin Cash","symbol":"bch","market_cap_rank":20}]}`))
	}))
	defer server.Close()

	ps := NewWithClient(&http.Client{
		Transport: &mockTransport{server.URL},
	})

	results, err := ps.SearchCoins("bitcoin")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if results[0].ID != "bitcoin" {
		t.Errorf("Expected first result to be bitcoin, got %s", results[0].ID)
	}
}

// TestHasMapping tests the HasMapping function
func TestHasMapping(t *testing.T) {
	ps := New()

	if !ps.HasMapping("BTC") {
		t.Error("Expected BTC to have mapping")
	}
	if !ps.HasMapping("btc") { // Case insensitivity
		t.Error("Expected btc (lowercase) to have mapping")
	}
	if ps.HasMapping("UNKNOWN_COIN_XYZ") {
		t.Error("Expected UNKNOWN_COIN_XYZ to not have mapping")
	}
}

// TestGetUnmappedTickers tests the GetUnmappedTickers function
func TestGetUnmappedTickers(t *testing.T) {
	ps := New()

	unmapped := ps.GetUnmappedTickers([]string{"BTC", "ETH", "UNKNOWN1", "UNKNOWN2"})
	if len(unmapped) != 2 {
		t.Errorf("Expected 2 unmapped tickers, got %d", len(unmapped))
	}

	// Check the unmapped ones
	unmappedMap := make(map[string]bool)
	for _, u := range unmapped {
		unmappedMap[u] = true
	}
	if !unmappedMap["UNKNOWN1"] || !unmappedMap["UNKNOWN2"] {
		t.Errorf("Expected UNKNOWN1 and UNKNOWN2 to be unmapped, got %v", unmapped)
	}
}

// TestGetDefaultMappings tests the GetDefaultMappings function
func TestGetDefaultMappings(t *testing.T) {
	mappings := GetDefaultMappings()

	// Check it's a copy (modifying shouldn't affect original)
	originalLen := len(mappings)
	mappings["NEW_COIN"] = "new-coin"

	mappings2 := GetDefaultMappings()
	if len(mappings2) != originalLen {
		t.Error("GetDefaultMappings should return a copy, not the original map")
	}
}

// mockTransport redirects all requests to the test server
type mockTransport struct {
	testServerURL string
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Redirect to test server while preserving query params
	testReq, err := http.NewRequest(req.Method, t.testServerURL+req.URL.Path+"?"+req.URL.RawQuery, req.Body)
	if err != nil {
		return nil, err
	}
	return http.DefaultTransport.RoundTrip(testReq)
}

// TestContextCancellation tests that context cancellation works properly
func TestContextCancellation(t *testing.T) {
	t.Run("GetPricesWithContext canceled", func(t *testing.T) {
		ps := New(WithRateLimit(0)) // Disable rate limiting

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := ps.GetPricesWithContext(ctx, []string{"BTC"})
		if err == nil {
			t.Error("Expected context canceled error")
		}
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got: %v", err)
		}
	})

	t.Run("GetPriceWithContext canceled", func(t *testing.T) {
		ps := New(WithRateLimit(0))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := ps.GetPriceWithContext(ctx, "BTC")
		if err == nil {
			t.Error("Expected context canceled error")
		}
	})

	t.Run("SearchCoinsWithContext canceled", func(t *testing.T) {
		ps := New(WithRateLimit(0))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := ps.SearchCoinsWithContext(ctx, "bitcoin")
		if err == nil {
			t.Error("Expected context canceled error")
		}
	})

	t.Run("context with timeout applied to request", func(t *testing.T) {
		// Test that context is properly passed to HTTP request
		// We use a successful request with a valid context to ensure the code path works
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"bitcoin":{"usd":50000}}`))
		}))
		defer server.Close()

		ps := New(
			WithHTTPClient(&http.Client{Transport: &mockTransport{server.URL}}),
			WithRateLimit(0),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		prices, err := ps.GetPricesWithContext(ctx, []string{"BTC"})
		if err != nil {
			t.Fatalf("Expected success with valid context, got: %v", err)
		}
		if prices["BTC"] != 50000 {
			t.Errorf("Expected BTC price 50000, got %f", prices["BTC"])
		}
	})
}

// TestFunctionalOptions tests the functional options pattern
func TestFunctionalOptions(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		ps := New()
		// Verify defaults are set (we can't directly access private fields,
		// but we can test behavior)
		if ps.client == nil {
			t.Error("Expected client to be set")
		}
		if !ps.HasMapping("BTC") {
			t.Error("Expected default mappings to be loaded")
		}
	})

	t.Run("WithCacheTTL", func(t *testing.T) {
		ps := New(WithCacheTTL(5 * time.Minute))
		// Test by checking cache behavior - set up mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"bitcoin":{"usd":50000}}`))
		}))
		defer server.Close()

		ps.client = &http.Client{Transport: &mockTransport{server.URL}}
		ps.minInterval = 0 // Disable rate limiting for test

		// First call should hit API
		_, err := ps.GetPrice("BTC")
		if err != nil {
			t.Fatalf("First call failed: %v", err)
		}

		// Second call should use cache (within 5 min TTL)
		ps.client = nil // Would fail if tried to use
		_, err = ps.GetPrice("BTC")
		if err != nil {
			t.Fatalf("Second call should use cache: %v", err)
		}
	})

	t.Run("WithRateLimit", func(t *testing.T) {
		ps := New(WithRateLimit(50 * time.Millisecond))

		start := time.Now()
		ps.waitForRateLimit()
		first := time.Since(start)

		start = time.Now()
		ps.waitForRateLimit()
		second := time.Since(start)

		if first > 20*time.Millisecond {
			t.Errorf("First call took too long: %v", first)
		}
		if second < 30*time.Millisecond {
			t.Errorf("Second call should wait ~50ms, took: %v", second)
		}
	})

	t.Run("WithHTTPClient", func(t *testing.T) {
		customClient := &http.Client{Timeout: 30 * time.Second}
		ps := New(WithHTTPClient(customClient))
		if ps.client != customClient {
			t.Error("Expected custom client to be set")
		}
	})

	t.Run("WithCoinMappings", func(t *testing.T) {
		customMappings := map[string]string{
			"CUSTOM1": "custom-coin-1",
			"CUSTOM2": "custom-coin-2",
		}
		ps := New(WithCoinMappings(customMappings))

		if ps.GetCoinGeckoID("CUSTOM1") != "custom-coin-1" {
			t.Error("Expected CUSTOM1 mapping to be set")
		}
		if ps.GetCoinGeckoID("CUSTOM2") != "custom-coin-2" {
			t.Error("Expected CUSTOM2 mapping to be set")
		}
		// Verify default mappings are still present
		if ps.GetCoinGeckoID("BTC") != "bitcoin" {
			t.Error("Expected default BTC mapping to still be present")
		}
	})

	t.Run("multiple options", func(t *testing.T) {
		customClient := &http.Client{Timeout: 15 * time.Second}
		ps := New(
			WithCacheTTL(10*time.Minute),
			WithRateLimit(100*time.Millisecond),
			WithHTTPClient(customClient),
			WithCoinMappings(map[string]string{"TEST": "test-coin"}),
		)

		if ps.client != customClient {
			t.Error("Expected custom client")
		}
		if ps.GetCoinGeckoID("TEST") != "test-coin" {
			t.Error("Expected TEST mapping")
		}
		if !ps.HasMapping("BTC") {
			t.Error("Expected default mappings")
		}
	})
}
