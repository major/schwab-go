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

func TestGetExpirationChain(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/expirationchain", r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]any{
			"expirationList": []map[string]any{
				{
					"expiration":       "2024-01-19",
					"daysToExpiration": 4,
					"expirationType":   "R",
					"settlementType":   "P",
					"optionRoots":      "AAPL",
					"standard":         true,
				},
				{
					"expiration":       "2024-01-26",
					"daysToExpiration": 11,
					"expirationType":   "R",
					"settlementType":   "P",
					"optionRoots":      "AAPL",
					"standard":         true,
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

	result, err := client.GetExpirationChain(context.Background(), "AAPL")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify structure
	require.Len(t, result.ExpirationList, 2)

	// First expiration
	exp1 := result.ExpirationList[0]
	assert.Equal(t, "2024-01-19", exp1.ExpirationDate)
	assert.Equal(t, 4, exp1.DaysToExpiration)
	assert.Equal(t, "R", exp1.ExpirationType)
	assert.Equal(t, "P", exp1.SettlementType)
	assert.Equal(t, "AAPL", exp1.OptionRoots)
	assert.True(t, exp1.Standard)

	// Second expiration
	exp2 := result.ExpirationList[1]
	assert.Equal(t, "2024-01-26", exp2.ExpirationDate)
	assert.Equal(t, 11, exp2.DaysToExpiration)
	assert.Equal(t, "R", exp2.ExpirationType)
	assert.Equal(t, "P", exp2.SettlementType)
	assert.Equal(t, "AAPL", exp2.OptionRoots)
	assert.True(t, exp2.Standard)
}

func TestGetExpirationChain_SymbolParam(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/expirationchain", r.URL.Path)

		// Verify symbol query param is set correctly
		symbol := r.URL.Query().Get("symbol")
		require.Equal(t, "TSLA", symbol)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]any{
			"expirationList": []map[string]any{},
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetExpirationChain(context.Background(), "TSLA")
	require.NoError(t, err)
}

func TestGetExpirationChain_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetExpirationChain(context.Background(), "INVALID")
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
