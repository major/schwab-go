package marketdata

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schwab "github.com/major/schwab-go/schwab"
)

const recordedInstrumentByCUSIPResponse = `{
  "instruments": [
    {
      "cusip": "037833100",
      "symbol": "AAPL",
      "description": "Apple Inc",
      "exchange": "NASDAQ",
      "assetType": "EQUITY"
    }
  ]
}`

func TestSearchInstruments(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/instruments", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify query params
		symbol := r.URL.Query().Get("symbol")
		assert.Equal(t, "AAPL", symbol)
		projection := r.URL.Query().Get("projection")
		assert.Equal(t, "symbol-search", projection)

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
	})

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
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/instruments", r.URL.Path)

		// Verify projection param
		projection := r.URL.Query().Get("projection")
		assert.Equal(t, "fundamental", projection)

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
	})

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
	assert.InDelta(t, 199.62, fund.High52, 0.000001)
	assert.InDelta(t, 124.17, fund.Low52, 0.000001)
	assert.InDelta(t, 0.24, fund.DividendAmount, 0.000001)
	assert.InDelta(t, 0.45, fund.DividendYield, 0.000001)
	assert.InDelta(t, 28.5, fund.PeRatio, 0.000001)
	assert.InDelta(t, 1.2, fund.Beta, 0.000001)
	assert.InDelta(t, 2800000000000.0, fund.MarketCap, 0.000001)
	assert.InDelta(t, 6.05, fund.EpsTTM, 0.000001)
	assert.InDelta(t, 0.85, fund.ReturnOnEquity, 0.000001)
	assert.InDelta(t, 1.08, fund.CurrentRatio, 0.000001)
}

func TestGetInstrumentByCUSIP(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/instruments/037833100", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

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
	})

	result, err := client.GetInstrumentByCUSIP(context.Background(), "037833100")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify structure
	assert.Equal(t, "037833100", result.Cusip)
	assert.Equal(t, "AAPL", result.Symbol)
	assert.Equal(t, "Apple Inc", result.Description)
	assert.Equal(t, "NASDAQ", result.Exchange)
	assert.Equal(t, schwab.AssetType("EQUITY"), result.AssetType)
}

func TestGetInstrumentByCUSIP_RecordedWrappedResponse(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/instruments/037833100", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprint(w, recordedInstrumentByCUSIPResponse)
		assert.NoError(t, err)
	})

	result, err := client.GetInstrumentByCUSIP(context.Background(), "037833100")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "037833100", result.Cusip)
	assert.Equal(t, "AAPL", result.Symbol)
	assert.Equal(t, schwab.AssetType("EQUITY"), result.AssetType)
}

func TestGetInstrumentByCUSIP_EmptyInstruments(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/instruments/000000000", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, InstrumentResponse{
			Instruments: []Instrument{},
		})
	})

	result, err := client.GetInstrumentByCUSIP(context.Background(), "000000000")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no instrument found for CUSIP")
}

func TestSearchInstruments_Error(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := client.SearchInstruments(context.Background(), "INVALID", ProjectionSymbolSearch)
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestGetInstrumentByCUSIP_Error(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := client.GetInstrumentByCUSIP(context.Background(), "INVALID")
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
