package marketdata

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schwab "github.com/major/schwab-go/schwab"
)

func TestGetMovers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/movers/$DJI", r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MoverResponse{
			Screeners: []Screener{
				{
					Symbol:            "AAPL",
					Description:       "Apple Inc.",
					LastPrice:         150.25,
					NetChange:         2.50,
					NetPercentChange:  1.69,
					MarketShare:       0.05,
					TotalVolume:       50000000,
					Trades:            1000000,
				},
				{
					Symbol:            "MSFT",
					Description:       "Microsoft Corporation",
					LastPrice:         380.50,
					NetChange:         3.75,
					NetPercentChange:  0.99,
					MarketShare:       0.04,
					TotalVolume:       40000000,
					Trades:            900000,
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

	result, err := client.GetMovers(context.Background(), "$DJI", "", 0)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify structure
	require.Len(t, result.Screeners, 2)

	// First screener
	assert.Equal(t, "AAPL", result.Screeners[0].Symbol)
	assert.Equal(t, "Apple Inc.", result.Screeners[0].Description)
	assert.Equal(t, 150.25, result.Screeners[0].LastPrice)
	assert.Equal(t, 2.50, result.Screeners[0].NetChange)
	assert.Equal(t, 1.69, result.Screeners[0].NetPercentChange)
	assert.Equal(t, 0.05, result.Screeners[0].MarketShare)
	assert.Equal(t, int64(50000000), result.Screeners[0].TotalVolume)
	assert.Equal(t, int64(1000000), result.Screeners[0].Trades)

	// Second screener
	assert.Equal(t, "MSFT", result.Screeners[1].Symbol)
	assert.Equal(t, "Microsoft Corporation", result.Screeners[1].Description)
	assert.Equal(t, 380.50, result.Screeners[1].LastPrice)
}

func TestGetMovers_WithSort(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/movers/$COMPX", r.URL.Path)

		// Verify sort query param is present
		sort := r.URL.Query().Get("sort")
		require.Equal(t, "VOLUME", sort)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MoverResponse{
			Screeners: []Screener{},
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetMovers(context.Background(), "$COMPX", MoverSortVolume, 0)
	require.NoError(t, err)
}

func TestGetMovers_WithFrequency(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/movers/$SPX", r.URL.Path)

		// Verify frequency query param is present
		frequency := r.URL.Query().Get("frequency")
		require.Equal(t, "60", frequency)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MoverResponse{
			Screeners: []Screener{},
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetMovers(context.Background(), "$SPX", "", 60)
	require.NoError(t, err)
}

func TestGetMovers_NoOptionalParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/movers/$DJI", r.URL.Path)

		// Verify optional params are absent
		sort := r.URL.Query().Get("sort")
		require.Empty(t, sort)
		frequency := r.URL.Query().Get("frequency")
		require.Empty(t, frequency)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MoverResponse{
			Screeners: []Screener{},
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetMovers(context.Background(), "$DJI", "", 0)
	require.NoError(t, err)
}

func TestGetMovers_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetMovers(context.Background(), "$DJI", "", 0)
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
