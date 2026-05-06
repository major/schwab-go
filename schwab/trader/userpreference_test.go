package trader

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

func TestGetUserPreference(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/userPreference", r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, []map[string]interface{}{
			{
				"accounts": []map[string]interface{}{
					{
						"accountNumber":      "123456789",
						"primaryAccount":     true,
						"type":               "BROKERAGE",
						"nickName":           "My Account",
						"accountColor":       "Blue",
						"displayAcctId":      "...6789",
						"autoPositionEffect": false,
					},
				},
				"streamerInfo": []map[string]interface{}{
					{
						"streamerSocketUrl":      "wss://streamer.schwab.com/ws",
						"schwabClientCustomerId": "customer123",
						"schwabClientCorrelId":   "correl456",
						"schwabClientChannel":    "IO",
						"schwabClientFunctionId": "APIAPP",
					},
				},
				"offers": []map[string]interface{}{
					{
						"level2Permissions": true,
						"mktDataPermission": "NP",
					},
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

	result, err := client.GetUserPreference(context.Background())
	require.NoError(t, err)
	require.Len(t, result, 1)

	// Verify accounts
	require.Len(t, result[0].Accounts, 1)
	account := result[0].Accounts[0]
	assert.Equal(t, "123456789", account.AccountNumber)
	assert.True(t, account.PrimaryAccount)
	assert.Equal(t, "BROKERAGE", account.Type)
	assert.Equal(t, "My Account", account.NickName)
	assert.Equal(t, "Blue", account.AccountColor)
	assert.Equal(t, "...6789", account.DisplayAcctId)
	assert.False(t, account.AutoPositionEffect)

	// Verify streamer info
	require.Len(t, result[0].StreamerInfo, 1)
	streamer := result[0].StreamerInfo[0]
	assert.Equal(t, "wss://streamer.schwab.com/ws", streamer.StreamerSocketUrl)
	assert.Equal(t, "customer123", streamer.SchwabClientCustomerId)
	assert.Equal(t, "correl456", streamer.SchwabClientCorrelId)
	assert.Equal(t, "IO", streamer.SchwabClientChannel)
	assert.Equal(t, "APIAPP", streamer.SchwabClientFunctionId)

	// Verify offers
	require.Len(t, result[0].Offers, 1)
	offer := result[0].Offers[0]
	assert.True(t, offer.Level2Permissions)
	assert.Equal(t, "NP", offer.MktDataPermission)
}

func TestGetUserPreference_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetUserPreference(context.Background())
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
}
