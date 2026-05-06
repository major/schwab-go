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

func TestGetPriceHistory(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/pricehistory", r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify symbol query param
		symbol := r.URL.Query().Get("symbol")
		require.Equal(t, "AAPL", symbol)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"candles": []map[string]interface{}{
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
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithToken("test-token"),
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	result, err := client.GetPriceHistory(context.Background(), "AAPL", nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify structure
	assert.Equal(t, "AAPL", result.Symbol)
	assert.False(t, result.Empty)
	require.Len(t, result.Candles, 1)

	// Verify candle OHLCV fields
	candle := result.Candles[0]
	assert.Equal(t, 150.0, candle.Open)
	assert.Equal(t, 155.0, candle.High)
	assert.Equal(t, 149.0, candle.Low)
	assert.Equal(t, 153.0, candle.Close)
	assert.Equal(t, int64(1000000), candle.Volume)
	assert.Equal(t, int64(1705276800000), candle.Datetime)
}

func TestGetPriceHistory_WithParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/pricehistory", r.URL.Path)

		// Verify all query params are present
		symbol := r.URL.Query().Get("symbol")
		require.Equal(t, "AAPL", symbol)

		periodType := r.URL.Query().Get("periodType")
		require.Equal(t, "day", periodType)

		period := r.URL.Query().Get("period")
		require.Equal(t, "10", period)

		frequencyType := r.URL.Query().Get("frequencyType")
		require.Equal(t, "minute", frequencyType)

		frequency := r.URL.Query().Get("frequency")
		require.Equal(t, "5", frequency)

		startDate := r.URL.Query().Get("startDate")
		require.Equal(t, "1705276800000", startDate)

		endDate := r.URL.Query().Get("endDate")
		require.Equal(t, "1705363200000", endDate)

		needExtendedHoursData := r.URL.Query().Get("needExtendedHoursData")
		require.Equal(t, "true", needExtendedHoursData)

		needPreviousClose := r.URL.Query().Get("needPreviousClose")
		require.Equal(t, "false", needPreviousClose)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"candles": []map[string]interface{}{},
			"symbol":  "AAPL",
			"empty":   true,
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/pricehistory", r.URL.Path)

		// Verify only symbol is present
		symbol := r.URL.Query().Get("symbol")
		require.Equal(t, "AAPL", symbol)

		// Verify optional params are absent
		require.Empty(t, r.URL.Query().Get("periodType"))
		require.Empty(t, r.URL.Query().Get("period"))
		require.Empty(t, r.URL.Query().Get("frequencyType"))
		require.Empty(t, r.URL.Query().Get("frequency"))
		require.Empty(t, r.URL.Query().Get("startDate"))
		require.Empty(t, r.URL.Query().Get("endDate"))
		require.Empty(t, r.URL.Query().Get("needExtendedHoursData"))
		require.Empty(t, r.URL.Query().Get("needPreviousClose"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"candles": []map[string]interface{}{},
			"symbol":  "AAPL",
			"empty":   true,
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	result, err := client.GetPriceHistory(context.Background(), "AAPL", nil)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestGetPriceHistory_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetPriceHistory(context.Background(), "AAPL", nil)
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
}
