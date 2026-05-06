package marketdata

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schwab "github.com/major/schwab-go/schwab"
)

func marketHoursDate(t *testing.T, years, days int) string {
	t.Helper()
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return today.AddDate(years, 0, days).Format(marketHoursDateLayout)
}

func TestGetMarketHours(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
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
	})

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

func TestGetMarketHoursTyped(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/markets", r.URL.Path)
		assert.Equal(t, "equity,option", r.URL.Query().Get("markets"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, MarketHoursMap{})
	})

	_, err := client.GetMarketHoursTyped(context.Background(), []MarketID{MarketIDEquity, MarketIDOption}, "")
	require.NoError(t, err)
}

func TestGetMarketHours_WithDate(t *testing.T) {
	requestDate := marketHoursDate(t, 0, 1)
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/markets", r.URL.Path)

		// Verify date query param is present
		date := r.URL.Query().Get("date")
		assert.Equal(t, requestDate, date)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, MarketHoursMap{})
	})

	_, err := client.GetMarketHours(context.Background(), []string{"equity"}, requestDate)
	require.NoError(t, err)
}

func TestGetMarketHours_NoDate(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/markets", r.URL.Path)

		// Verify date query param is absent
		date := r.URL.Query().Get("date")
		assert.Empty(t, date)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, MarketHoursMap{})
	})

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

func TestGetMarketHoursRejectsInvalidMarket(t *testing.T) {
	client := NewClient()

	_, err := client.GetMarketHours(context.Background(), []string{"invalid"}, "")
	require.EqualError(t, err, `invalid market "invalid": expected one of equity, option, bond, future, forex`)
}

func TestGetMarketHoursRejectsDuplicateMarkets(t *testing.T) {
	client := NewClient()

	_, err := client.GetMarketHours(context.Background(), []string{"equity", "equity"}, "")
	require.EqualError(t, err, `duplicate market "equity"`)
}

func TestGetMarketHoursRejectsInvalidDate(t *testing.T) {
	client := NewClient()

	for _, date := range []string{"2024/01/15", "01-15-2024", "2024-1-15"} {
		t.Run(date, func(t *testing.T) {
			_, err := client.GetMarketHours(context.Background(), []string{"equity"}, date)
			require.EqualError(t, err, `invalid date "`+date+`": expected YYYY-MM-DD`)
		})
	}
}

func TestValidateMarketHoursDateRejectsOutOfRangeDate(t *testing.T) {
	now := time.Date(2026, time.May, 6, 12, 0, 0, 0, time.Local)
	pastDate := now.AddDate(0, 0, -1).Format(marketHoursDateLayout)
	futureDate := now.AddDate(1, 0, 1).Format(marketHoursDateLayout)

	err := validateMarketHoursDateWithNow(pastDate, now)
	require.EqualError(t, err, `date "`+pastDate+`" is before today`)

	err = validateMarketHoursDateWithNow(futureDate, now)
	require.EqualError(t, err, `date "`+futureDate+`" is more than one year from today`)
}

func TestValidateMarketHoursDateAllowsInclusiveRange(t *testing.T) {
	now := time.Date(2026, time.May, 6, 12, 0, 0, 0, time.UTC)
	today := now.Format(marketHoursDateLayout)
	oneYearOut := now.AddDate(1, 0, 0).Format(marketHoursDateLayout)

	require.NoError(t, validateMarketHoursDateWithNow(today, now))
	require.NoError(t, validateMarketHoursDateWithNow(oneYearOut, now))
}

func TestGetMarketHoursSingle(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
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
	})

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

func TestGetMarketHoursSingleTyped(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/markets/equity", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, MarketHoursMap{})
	})

	_, err := client.GetMarketHoursSingleTyped(context.Background(), MarketIDEquity, "")
	require.NoError(t, err)
}

func TestGetMarketHoursSingle_WithDate(t *testing.T) {
	requestDate := marketHoursDate(t, 0, 1)
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/markets/equity", r.URL.Path)

		// Verify date query param is present
		date := r.URL.Query().Get("date")
		assert.Equal(t, requestDate, date)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, MarketHoursMap{})
	})

	_, err := client.GetMarketHoursSingle(context.Background(), "equity", requestDate)
	require.NoError(t, err)
}

func TestGetMarketHoursSingleRejectsInvalidMarket(t *testing.T) {
	client := NewClient()

	_, err := client.GetMarketHoursSingle(context.Background(), "invalid", "")
	require.EqualError(t, err, `invalid market "invalid": expected one of equity, option, bond, future, forex`)
}

func TestGetMarketHoursSingleRejectsInvalidDate(t *testing.T) {
	client := NewClient()

	_, err := client.GetMarketHoursSingle(context.Background(), "equity", "2024/01/15")
	require.EqualError(t, err, `invalid date "2024/01/15": expected YYYY-MM-DD`)
}

func TestGetMarketHours_Error(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := client.GetMarketHours(context.Background(), []string{"equity"}, "")
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestGetMarketHoursSingle_Error(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := client.GetMarketHoursSingle(context.Background(), "equity", "")
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
}
