package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypedAuthErrorContracts(t *testing.T) {
	t.Parallel()

	t.Run("missing token file returns auth required", func(t *testing.T) {
		t.Parallel()

		store := NewFileTokenStore(filepath.Join(t.TempDir(), "missing-token.json"))

		_, err := store.Load(context.Background())

		require.Error(t, err)
		var requiredErr *AuthRequiredError
		require.ErrorAs(t, err, &requiredErr)
	})

	t.Run("stale refresh token returns auth expired", func(t *testing.T) {
		t.Parallel()

		store := newProviderMemoryStore(providerTokenFile(
			"expired-access-token",
			"stale-refresh-token",
			time.Now().Add(-time.Hour),
			time.Now().Add(-8*24*time.Hour),
		))
		provider := newTestProvider(t, "https://127.0.0.1:8182/oauth", store, nil)

		_, err := provider.Token(context.Background())

		require.Error(t, err)
		var expiredErr *AuthExpiredError
		require.ErrorAs(t, err, &expiredErr)
	})

	t.Run("refresh invalid grant returns auth expired", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, err := fmt.Fprint(w, `{"error":"invalid_grant"}`)
			assert.NoError(t, err)
		}))
		t.Cleanup(server.Close)

		_, err := RefreshAccessToken(
			context.Background(),
			refreshTestConfig(server.URL),
			"revoked-refresh-token",
			server.Client(),
		)

		require.Error(t, err)
		var expiredErr *AuthExpiredError
		require.ErrorAs(t, err, &expiredErr)
	})

	t.Run("callback listener setup failure returns auth callback", func(t *testing.T) {
		t.Parallel()

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		t.Cleanup(func() {
			assert.NoError(t, listener.Close())
		})

		callbackURL := fmt.Sprintf("https://%s/callback", listener.Addr().String())
		_, _, _, err = StartCallbackServer(context.Background(), callbackURL)

		require.Error(t, err)
		var callbackErr *AuthCallbackError
		require.ErrorAs(t, err, &callbackErr)
	})

	t.Run("callback missing code returns auth callback", func(t *testing.T) {
		t.Parallel()

		callbackURL := newCallbackTestURL(t)
		_, errs, shutdown, err := StartCallbackServer(context.Background(), callbackURL)
		require.NoError(t, err)
		t.Cleanup(shutdown)

		resp := getCallback(t, callbackURL, url.Values{"state": {"csrf-state"}})
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		_ = readResponseBody(t, resp)

		var callbackErr *AuthCallbackError
		select {
		case gotErr := <-errs:
			require.ErrorAs(t, gotErr, &callbackErr)
		case <-time.After(callbackTestTimeout):
			t.Fatal("StartCallbackServer() did not send missing code error")
		}
	})

	t.Run("callback missing state returns auth callback", func(t *testing.T) {
		t.Parallel()

		callbackURL := newCallbackTestURL(t)
		_, errs, shutdown, err := StartCallbackServer(context.Background(), callbackURL)
		require.NoError(t, err)
		t.Cleanup(shutdown)

		resp := getCallback(t, callbackURL, url.Values{"code": {"auth-code"}})
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		_ = readResponseBody(t, resp)

		var callbackErr *AuthCallbackError
		select {
		case gotErr := <-errs:
			require.ErrorAs(t, gotErr, &callbackErr)
		case <-time.After(callbackTestTimeout):
			t.Fatal("StartCallbackServer() did not send missing state error")
		}
	})
}
