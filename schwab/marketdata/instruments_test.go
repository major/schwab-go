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

func TestSearchInstruments(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/instruments", r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify query params
		symbol := r.URL.Query().Get("symbol")
		require.Equal(t, "AAPL", symbol)
		projection := r.URL.Query().Get("projection")
		require.Equal(t, "symbol-search", projection)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, InstrumentResponse{
			Instruments: []Instrument{
				{
					Cusip:       "037833100",
					Symbol:      "AAPL",
					Description: "Apple Inc",
					Exchange:    "NASDAQ",
					AssetType:   "EQUITY",
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

	result, err := client.SearchInstruments(context.Background(), "AAPL", ProjectionSymbolSearch)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify structure
	require.Len(t, result.Instruments, 1)
	inst := result.Instruments[0]
	assert.Equal(t, "037833100", inst.Cusip)
	assert.Equal(t, "AAPL", inst.Symbol)
	assert.Equal(t, "Apple Inc", inst.Description)
	assert.Equal(t, "NASDAQ", inst.Exchange)
	assert.Equal(t, schwab.AssetType("EQUITY"), inst.AssetType)
}

func TestSearchInstruments_Fundamental(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/instruments", r.URL.Path)

		// Verify projection param
		projection := r.URL.Query().Get("projection")
		require.Equal(t, "fundamental", projection)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, InstrumentResponse{
			Instruments: []Instrument{
				{
					Cusip:       "037833100",
					Symbol:      "AAPL",
					Description: "Apple Inc",
					Exchange:    "NASDAQ",
					AssetType:   "EQUITY",
					Fundamental: &FundamentalData{
						Symbol:         "AAPL",
						High52:         199.62,
						Low52:          124.17,
						DividendAmount: 0.24,
						DividendYield:  0.45,
						PeRatio:        28.5,
						Beta:           1.2,
						MarketCap:      2800000000000,
						EpsTTM:         6.05,
						ReturnOnEquity: 0.85,
						CurrentRatio:   1.08,
					},
				},
			},
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	result, err := client.SearchInstruments(context.Background(), "AAPL", ProjectionFundamental)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify structure
	require.Len(t, result.Instruments, 1)
	inst := result.Instruments[0]
	assert.Equal(t, "AAPL", inst.Symbol)
	require.NotNil(t, inst.Fundamental)

	// Verify fundamental data
	fund := inst.Fundamental
	assert.Equal(t, "AAPL", fund.Symbol)
	assert.Equal(t, 199.62, fund.High52)
	assert.Equal(t, 124.17, fund.Low52)
	assert.Equal(t, 0.24, fund.DividendAmount)
	assert.Equal(t, 0.45, fund.DividendYield)
	assert.Equal(t, 28.5, fund.PeRatio)
	assert.Equal(t, 1.2, fund.Beta)
	assert.Equal(t, 2800000000000.0, fund.MarketCap)
	assert.Equal(t, 6.05, fund.EpsTTM)
	assert.Equal(t, 0.85, fund.ReturnOnEquity)
	assert.Equal(t, 1.08, fund.CurrentRatio)
}

func TestGetInstrumentByCUSIP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/instruments/037833100", r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, Instrument{
			Cusip:       "037833100",
			Symbol:      "AAPL",
			Description: "Apple Inc",
			Exchange:    "NASDAQ",
			AssetType:   "EQUITY",
			Type:        "SWEEP_VEHICLE",
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithToken("test-token"),
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	result, err := client.GetInstrumentByCUSIP(context.Background(), "037833100")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify structure
	assert.Equal(t, "037833100", result.Cusip)
	assert.Equal(t, "AAPL", result.Symbol)
	assert.Equal(t, "Apple Inc", result.Description)
	assert.Equal(t, "NASDAQ", result.Exchange)
	assert.Equal(t, schwab.AssetType("EQUITY"), result.AssetType)
	assert.Equal(t, "SWEEP_VEHICLE", result.Type)
}

func TestSearchInstruments_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.SearchInstruments(context.Background(), "INVALID", ProjectionSymbolSearch)
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestGetInstrumentByCUSIP_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetInstrumentByCUSIP(context.Background(), "INVALID")
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
