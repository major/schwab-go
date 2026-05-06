package trader

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

func TestGetAccountNumbers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/accounts/accountNumbers", r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]AccountNumberHash{
			{
				AccountNumber: "123456789",
				HashValue:     "HASH_ABC123",
			},
			{
				AccountNumber: "987654321",
				HashValue:     "HASH_XYZ789",
			},
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithToken("test-token"),
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	result, err := client.GetAccountNumbers(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result, 2)

	// Verify first account
	assert.Equal(t, "123456789", result[0].AccountNumber)
	assert.Equal(t, "HASH_ABC123", result[0].HashValue)

	// Verify second account
	assert.Equal(t, "987654321", result[1].AccountNumber)
	assert.Equal(t, "HASH_XYZ789", result[1].HashValue)
}

func TestGetAccountNumbers_Empty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/accounts/accountNumbers", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]AccountNumberHash{})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	result, err := client.GetAccountNumbers(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result, 0)
}

func TestGetAccountNumbers_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetAccountNumbers(context.Background())
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
}
