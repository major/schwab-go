package marketdata

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schwab "github.com/major/schwab-go/schwab"
)

func TestGetMarketHours(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/markets", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify markets query param
		markets := r.URL.Query().Get("markets")
		assert.Equal(t, "equity,option", markets)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]map[string]MarketHours{
			"equity": {
				"EQ": {
					Date:        "2024-01-15",
					MarketType:  "EQUITY",
					Exchange:    "NYSE",
					Category:    "EQUITY",
					Product:     "EQ",
					ProductName: "equity",
					IsOpen:      true,
					SessionHours: map[string][]SessionHours{
						"regularMarket": {
							{
								Start: "2024-01-15T09:30:00-05:00",
								End:   "2024-01-15T16:00:00-05:00",
							},
						},
					},
				},
			},
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithToken("test-token"),
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	result, err := client.GetMarketHours(context.Background(), []string{"equity", "option"}, "")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify structure
	equityMarkets, ok := result["equity"]
	require.True(t, ok)
	eq, ok := equityMarkets["EQ"]
	require.True(t, ok)
	assert.Equal(t, "2024-01-15", eq.Date)
	assert.Equal(t, "EQUITY", eq.MarketType)
	assert.Equal(t, "NYSE", eq.Exchange)
	assert.Equal(t, "EQ", eq.Product)
	assert.Equal(t, "equity", eq.ProductName)
	assert.True(t, eq.IsOpen)

	// Verify session hours
	sessions, ok := eq.SessionHours["regularMarket"]
	require.True(t, ok)
	require.Len(t, sessions, 1)
	assert.Equal(t, "2024-01-15T09:30:00-05:00", sessions[0].Start)
	assert.Equal(t, "2024-01-15T16:00:00-05:00", sessions[0].End)
}

func TestGetMarketHours_WithDate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/markets", r.URL.Path)

		// Verify date query param is present
		date := r.URL.Query().Get("date")
		assert.Equal(t, "2024-01-15", date)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, MarketHoursMap{})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetMarketHours(context.Background(), []string{"equity"}, "2024-01-15")
	require.NoError(t, err)
}

func TestGetMarketHours_NoDate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/markets", r.URL.Path)

		// Verify date query param is absent
		date := r.URL.Query().Get("date")
		assert.Empty(t, date)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, MarketHoursMap{})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetMarketHours(context.Background(), []string{"equity"}, "")
	require.NoError(t, err)
}

func TestGetMarketHoursRequiresMarkets(t *testing.T) {
	client := NewClient()

	_, err := client.GetMarketHours(context.Background(), nil, "")
	require.EqualError(t, err, "markets is required")

	_, err = client.GetMarketHours(context.Background(), []string{}, "")
	require.EqualError(t, err, "markets is required")
}

func TestGetMarketHoursSingle(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/markets/equity", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]map[string]MarketHours{
			"equity": {
				"EQ": {
					Date:        "2024-01-15",
					MarketType:  "EQUITY",
					Exchange:    "NYSE",
					Category:    "EQUITY",
					Product:     "EQ",
					ProductName: "equity",
					IsOpen:      true,
					SessionHours: map[string][]SessionHours{
						"regularMarket": {
							{
								Start: "2024-01-15T09:30:00-05:00",
								End:   "2024-01-15T16:00:00-05:00",
							},
						},
					},
				},
			},
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithToken("test-token"),
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	result, err := client.GetMarketHoursSingle(context.Background(), "equity", "")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify structure
	equityMarkets, ok := result["equity"]
	require.True(t, ok)
	eq, ok := equityMarkets["EQ"]
	require.True(t, ok)
	assert.Equal(t, "2024-01-15", eq.Date)
	assert.Equal(t, "EQUITY", eq.MarketType)
	assert.Equal(t, "NYSE", eq.Exchange)
}

func TestGetMarketHoursSingle_WithDate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/markets/equity", r.URL.Path)

		// Verify date query param is present
		date := r.URL.Query().Get("date")
		assert.Equal(t, "2024-01-15", date)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, MarketHoursMap{})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetMarketHoursSingle(context.Background(), "equity", "2024-01-15")
	require.NoError(t, err)
}

func TestGetMarketHours_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetMarketHours(context.Background(), []string{"equity"}, "")
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestGetMarketHoursSingle_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetMarketHoursSingle(context.Background(), "equity", "")
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
}
