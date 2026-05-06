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

func TestGetExpirationChain(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/expirationchain", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

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
	})

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
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/expirationchain", r.URL.Path)

		// Verify symbol query param is set correctly
		symbol := r.URL.Query().Get("symbol")
		assert.Equal(t, "TSLA", symbol)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]any{
			"expirationList": []map[string]any{},
		})
	})

	_, err := client.GetExpirationChain(context.Background(), "TSLA")
	require.NoError(t, err)
}

func TestGetExpirationChain_Error(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := client.GetExpirationChain(context.Background(), "INVALID")
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
