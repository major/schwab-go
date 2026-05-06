package marketdata

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schwab "github.com/major/schwab-go/schwab"
)

func TestGetMovers(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/movers/$DJI", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, MoverResponse{
			Screeners: []Screener{
				{
					Symbol:           "AAPL",
					Description:      "Apple Inc.",
					Direction:        "up",
					Last:             150.25,
					Change:           2.50,
					NetPercentChange: 1.69,
					MarketShare:      0.05,
					TotalVolume:      50000000,
					Trades:           1000000,
				},
				{
					Symbol:           "MSFT",
					Description:      "Microsoft Corporation",
					Direction:        "up",
					Last:             380.50,
					Change:           3.75,
					NetPercentChange: 0.99,
					MarketShare:      0.04,
					TotalVolume:      40000000,
					Trades:           900000,
				},
			},
		})
	})

	result, err := client.GetMovers(context.Background(), "$DJI", "", nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify structure
	require.Len(t, result.Screeners, 2)

	// First screener
	assert.Equal(t, "AAPL", result.Screeners[0].Symbol)
	assert.Equal(t, "Apple Inc.", result.Screeners[0].Description)
	assert.InDelta(t, 150.25, result.Screeners[0].Last, 0.000001)
	assert.InDelta(t, 2.50, result.Screeners[0].Change, 0.000001)
	assert.Equal(t, "up", result.Screeners[0].Direction)
	assert.InDelta(t, 1.69, result.Screeners[0].NetPercentChange, 0.000001)
	assert.InDelta(t, 0.05, result.Screeners[0].MarketShare, 0.000001)
	assert.Equal(t, int64(50000000), result.Screeners[0].TotalVolume)
	assert.Equal(t, int64(1000000), result.Screeners[0].Trades)

	// Second screener
	assert.Equal(t, "MSFT", result.Screeners[1].Symbol)
	assert.Equal(t, "Microsoft Corporation", result.Screeners[1].Description)
	assert.InDelta(t, 380.50, result.Screeners[1].Last, 0.000001)
}

func TestGetMovers_WithSort(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/movers/$COMPX", r.URL.Path)

		// Verify sort query param is present
		sort := r.URL.Query().Get("sort")
		assert.Equal(t, "VOLUME", sort)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, MoverResponse{
			Screeners: []Screener{},
		})
	})

	_, err := client.GetMovers(context.Background(), "$COMPX", MoverSortVolume, nil)
	require.NoError(t, err)
}

func TestGetMovers_WithFrequency(t *testing.T) {
	freq := 60

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/movers/$SPX", r.URL.Path)

		// Verify frequency query param is present
		frequency := r.URL.Query().Get("frequency")
		assert.Equal(t, "60", frequency)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, MoverResponse{
			Screeners: []Screener{},
		})
	})

	_, err := client.GetMovers(context.Background(), "$SPX", "", &freq)
	require.NoError(t, err)
}

func TestGetMovers_FrequencyZero(t *testing.T) {
	zero := 0

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/movers/$DJI", r.URL.Path)

		frequency := r.URL.Query().Get("frequency")
		assert.Equal(t, "0", frequency)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, MoverResponse{
			Screeners: []Screener{},
		})
	})

	_, err := client.GetMovers(context.Background(), "$DJI", "", &zero)
	require.NoError(t, err)
}

func TestGetMovers_NoOptionalParams(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/movers/$DJI", r.URL.Path)

		// Verify optional params are absent
		sort := r.URL.Query().Get("sort")
		assert.Empty(t, sort)
		frequency := r.URL.Query().Get("frequency")
		assert.Empty(t, frequency)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, MoverResponse{
			Screeners: []Screener{},
		})
	})

	_, err := client.GetMovers(context.Background(), "$DJI", "", nil)
	require.NoError(t, err)
}

func TestGetMovers_Error(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := client.GetMovers(context.Background(), "$DJI", "", nil)
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
