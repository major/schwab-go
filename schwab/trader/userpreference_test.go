package trader

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schwab "github.com/major/schwab-go/schwab"
)

func TestGetUserPreference(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/userPreference", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, []map[string]any{
			{
				"accounts": []map[string]any{
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
				"streamerInfo": []map[string]any{
					{
						"streamerSocketUrl":      "wss://streamer.schwab.com/ws",
						"streamerUrl":            "wss://streamer.schwab.com/legacy",
						"token":                  "test-stream-token-placeholder",
						"tokenExpTime":           "2026-05-07T12:00:00Z",
						"appId":                  "APP-123",
						"acl":                    "test-acl-placeholder",
						"schwabClientCustomerId": "customer123",
						"schwabClientCorrelId":   "correl456",
						"schwabClientChannel":    "IO",
						"schwabClientFunctionId": "APIAPP",
					},
				},
				"offers": []map[string]any{
					{
						"level2Permissions": true,
						"mktDataPermission": "NP",
					},
				},
			},
		})
	})

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
	assert.Equal(t, "...6789", account.DisplayAcctID)
	assert.False(t, account.AutoPositionEffect)

	// Verify streamer info
	require.Len(t, result[0].StreamerInfo, 1)
	streamer := result[0].StreamerInfo[0]
	assert.Equal(t, "wss://streamer.schwab.com/ws", streamer.StreamerSocketURL)
	assert.Equal(t, "wss://streamer.schwab.com/legacy", streamer.StreamerURL)
	assert.Equal(t, "test-stream-token-placeholder", streamer.Token)
	assert.Equal(t, "2026-05-07T12:00:00Z", streamer.TokenExpirationTime)
	assert.Equal(t, "APP-123", streamer.AppID)
	assert.Equal(t, "test-acl-placeholder", streamer.ACL)
	assert.Equal(t, "customer123", streamer.SchwabClientCustomerID)
	assert.Equal(t, "correl456", streamer.SchwabClientCorrelID)
	assert.Equal(t, "IO", streamer.SchwabClientChannel)
	assert.Equal(t, "APIAPP", streamer.SchwabClientFunctionID)

	// Verify offers
	require.Len(t, result[0].Offers, 1)
	offer := result[0].Offers[0]
	assert.True(t, offer.Level2Permissions)
	assert.Equal(t, "NP", offer.MktDataPermission)
}

func TestStreamerInfoStringRedactsSecrets(t *testing.T) {
	streamer := StreamerInfo{
		StreamerSocketURL:   "wss://streamer.schwab.com/ws",
		StreamerURL:         "wss://streamer.schwab.com/legacy",
		Token:               "test-stream-token-placeholder",
		TokenExpirationTime: "2026-05-07T12:00:00Z",
		AppID:               "APP-123",
		ACL:                 "test-acl-placeholder",
	}

	formatted := fmt.Sprint(streamer)
	assert.NotContains(t, formatted, streamer.Token)
	assert.NotContains(t, formatted, streamer.ACL)
	assert.Equal(t, 2, strings.Count(formatted, "<redacted>"))
}

func TestGetUserPreference_Error(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	_, err := client.GetUserPreference(context.Background())
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
}
