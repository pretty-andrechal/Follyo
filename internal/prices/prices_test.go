package prices

import (
	"net/http"
	"net/http/httptest"
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
