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

func TestGetPriceHistory(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/pricehistory", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify symbol query param
		symbol := r.URL.Query().Get("symbol")
		assert.Equal(t, "AAPL", symbol)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]any{
			"candles": []map[string]any{
				{
					"open":     150.0,
					"high":     155.0,
					"low":      149.0,
					"close":    153.0,
					"volume":   1000000,
					"datetime": 1705276800000,
				},
			},
			"symbol": "AAPL",
			"empty":  false,
		})
	})

	result, err := client.GetPriceHistory(context.Background(), "AAPL", nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify structure
	assert.Equal(t, "AAPL", result.Symbol)
	assert.False(t, result.Empty)
	require.Len(t, result.Candles, 1)

	// Verify candle OHLCV fields
	candle := result.Candles[0]
	assert.InDelta(t, 150.0, candle.Open, 0.000001)
	assert.InDelta(t, 155.0, candle.High, 0.000001)
	assert.InDelta(t, 149.0, candle.Low, 0.000001)
	assert.InDelta(t, 153.0, candle.Close, 0.000001)
	assert.Equal(t, int64(1000000), candle.Volume)
	assert.Equal(t, int64(1705276800000), candle.Datetime)
}

func TestGetPriceHistory_WithParams(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/pricehistory", r.URL.Path)

		// Verify all query params are present
		symbol := r.URL.Query().Get("symbol")
		assert.Equal(t, "AAPL", symbol)

		periodType := r.URL.Query().Get("periodType")
		assert.Equal(t, "day", periodType)

		period := r.URL.Query().Get("period")
		assert.Equal(t, "10", period)

		frequencyType := r.URL.Query().Get("frequencyType")
		assert.Equal(t, "minute", frequencyType)

		frequency := r.URL.Query().Get("frequency")
		assert.Equal(t, "5", frequency)

		startDate := r.URL.Query().Get("startDate")
		assert.Equal(t, "1705276800000", startDate)

		endDate := r.URL.Query().Get("endDate")
		assert.Equal(t, "1705363200000", endDate)

		needExtendedHoursData := r.URL.Query().Get("needExtendedHoursData")
		assert.Equal(t, "true", needExtendedHoursData)

		needPreviousClose := r.URL.Query().Get("needPreviousClose")
		assert.Equal(t, "false", needPreviousClose)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]any{
			"candles": []map[string]any{},
			"symbol":  "AAPL",
			"empty":   true,
		})
	})

	trueVal := true
	falseVal := false
	params := &PriceHistoryParams{
		PeriodType:            PeriodTypeDay,
		Period:                10,
		FrequencyType:         FrequencyTypeMinute,
		Frequency:             5,
		StartDate:             1705276800000,
		EndDate:               1705363200000,
		NeedExtendedHoursData: &trueVal,
		NeedPreviousClose:     &falseVal,
	}

	result, err := client.GetPriceHistory(context.Background(), "AAPL", params)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "AAPL", result.Symbol)
	assert.True(t, result.Empty)
}

func TestGetPriceHistory_NilParams(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/pricehistory", r.URL.Path)

		// Verify only symbol is present
		symbol := r.URL.Query().Get("symbol")
		assert.Equal(t, "AAPL", symbol)

		// Verify optional params are absent
		assert.Empty(t, r.URL.Query().Get("periodType"))
		assert.Empty(t, r.URL.Query().Get("period"))
		assert.Empty(t, r.URL.Query().Get("frequencyType"))
		assert.Empty(t, r.URL.Query().Get("frequency"))
		assert.Empty(t, r.URL.Query().Get("startDate"))
		assert.Empty(t, r.URL.Query().Get("endDate"))
		assert.Empty(t, r.URL.Query().Get("needExtendedHoursData"))
		assert.Empty(t, r.URL.Query().Get("needPreviousClose"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]any{
			"candles": []map[string]any{},
			"symbol":  "AAPL",
			"empty":   true,
		})
	})

	result, err := client.GetPriceHistory(context.Background(), "AAPL", nil)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestGetPriceHistory_Error(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	_, err := client.GetPriceHistory(context.Background(), "AAPL", nil)
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
}
