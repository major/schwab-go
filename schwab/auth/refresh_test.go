package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshAccessToken(t *testing.T) {
	t.Parallel()

	t.Run("success returns token file", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, err := fmt.Fprint(
				w,
				`{"access_token":"new-access-token","token_type":"Bearer","expires_in":1800,"refresh_token":"new-refresh-token","scope":"api"}`,
			)
			assert.NoError(t, err)
		}))
		t.Cleanup(server.Close)

		before := time.Now().Unix()
		tokenFile, err := RefreshAccessToken(
			context.Background(),
			refreshTestConfig(server.URL),
			"old-refresh-token",
			server.Client(),
		)
		after := time.Now().Unix()

		require.NoError(t, err)
		assert.GreaterOrEqual(t, tokenFile.CreationTimestamp, before)
		assert.LessOrEqual(t, tokenFile.CreationTimestamp, after)
		assert.Equal(t, "new-access-token", tokenFile.Token.AccessToken)
		assert.Equal(t, "Bearer", tokenFile.Token.TokenType)
		assert.Equal(t, 1800, tokenFile.Token.ExpiresIn)
		assert.Equal(t, "new-refresh-token", tokenFile.Token.RefreshToken)
		assert.Equal(t, "api", tokenFile.Token.Scope)
		assert.Equal(t, tokenFile.CreationTimestamp+1800, tokenFile.Token.ExpiresAt)
	})

	t.Run("expired refresh token returns auth expired error", func(t *testing.T) {
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
			"expired-refresh-token",
			server.Client(),
		)

		require.Error(t, err)
		var expiredErr *AuthExpiredError
		assert.ErrorAs(t, err, &expiredErr)
	})

	t.Run("http error returns status", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "server failed", http.StatusInternalServerError)
		}))
		t.Cleanup(server.Close)

		_, err := RefreshAccessToken(
			context.Background(),
			refreshTestConfig(server.URL),
			"refresh-token",
			server.Client(),
		)

		require.Error(t, err)
		assert.ErrorContains(t, err, "status 500")
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		t.Parallel()

		cfg := Config{ClientID: "", ClientSecret: "secret", CallbackURL: "https://127.0.0.1:8182/callback"}
		_, err := RefreshAccessToken(context.Background(), cfg, "refresh-token", nil)

		require.Error(t, err)
		assert.ErrorContains(t, err, "client_id")
	})

	t.Run("request format uses basic auth and refresh form body", func(t *testing.T) {
		t.Parallel()

		requests := make(chan refreshRequestSnapshot, 1)
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			assert.NoError(t, err)

			requests <- refreshRequestSnapshot{
				method:        r.Method,
				path:          r.URL.Path,
				authorization: r.Header.Get("Authorization"),
				contentType:   r.Header.Get("Content-Type"),
				form:          r.PostForm,
			}

			w.Header().Set("Content-Type", "application/json")
			_, err = fmt.Fprint(
				w,
				`{"access_token":"new-access-token","token_type":"Bearer","expires_in":1800,"refresh_token":"new-refresh-token"}`,
			)
			assert.NoError(t, err)
		}))
		t.Cleanup(server.Close)

		cfg := refreshTestConfig(server.URL + "/")
		_, err := RefreshAccessToken(context.Background(), cfg, "old-refresh-token", server.Client())
		require.NoError(t, err)

		snapshot := <-requests
		expectedBasicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("client-id:client-secret"))
		assert.Equal(t, http.MethodPost, snapshot.method)
		assert.Equal(t, "/token", snapshot.path)
		assert.Equal(t, expectedBasicAuth, snapshot.authorization)
		assert.Equal(t, "application/x-www-form-urlencoded", snapshot.contentType)
		assert.Equal(t, "refresh_token", snapshot.form.Get("grant_type"))
		assert.Equal(t, "old-refresh-token", snapshot.form.Get("refresh_token"))
	})
}

func TestRefreshTokenFile(t *testing.T) {
	t.Parallel()

	t.Run("success preserves original creation timestamp", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, err := fmt.Fprint(
				w,
				`{"access_token":"new-access-token","token_type":"Bearer","expires_in":1800,"refresh_token":"new-refresh-token"}`,
			)
			assert.NoError(t, err)
		}))
		t.Cleanup(server.Close)

		createdAt := time.Now().Add(-time.Hour).Unix()
		tokenFile := TokenFile{
			CreationTimestamp: createdAt,
			Token: TokenData{
				AccessToken:  "old-access-token",
				RefreshToken: "old-refresh-token",
			},
		}

		refreshed, err := RefreshTokenFile(
			context.Background(),
			refreshTestConfig(server.URL),
			tokenFile,
			server.Client(),
		)

		require.NoError(t, err)
		assert.Equal(t, createdAt, refreshed.CreationTimestamp)
		assert.Equal(t, "new-access-token", refreshed.Token.AccessToken)
		assert.Equal(t, "new-refresh-token", refreshed.Token.RefreshToken)
	})

	t.Run("stale refresh token returns expired before request", func(t *testing.T) {
		t.Parallel()

		var requestCount atomic.Int32
		server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			requestCount.Add(1)
		}))
		t.Cleanup(server.Close)

		_, err := RefreshTokenFile(
			context.Background(),
			refreshTestConfig(server.URL),
			TokenFile{
				CreationTimestamp: time.Now().Add(-8 * 24 * time.Hour).Unix(),
				Token:             TokenData{RefreshToken: "stale-refresh-token"},
			},
			server.Client(),
		)

		require.Error(t, err)
		var expiredErr *AuthExpiredError
		require.ErrorAs(t, err, &expiredErr)
		assert.Equal(t, int32(0), requestCount.Load())
	})

	t.Run("missing refresh token returns error", func(t *testing.T) {
		t.Parallel()

		_, err := RefreshTokenFile(
			context.Background(),
			refreshTestConfig("https://api.schwabapi.com/v1/oauth"),
			TokenFile{CreationTimestamp: time.Now().Unix()},
			nil,
		)

		require.Error(t, err)
		assert.ErrorContains(t, err, "refresh token")
	})

	t.Run("refresh failure returns error", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "temporarily unavailable", http.StatusServiceUnavailable)
		}))
		t.Cleanup(server.Close)

		_, err := RefreshTokenFile(
			context.Background(),
			refreshTestConfig(server.URL),
			TokenFile{
				CreationTimestamp: time.Now().Add(-time.Hour).Unix(),
				Token:             TokenData{RefreshToken: "refresh-token"},
			},
			server.Client(),
		)

		require.Error(t, err)
		assert.ErrorContains(t, err, "token refresh failed with status 503")
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		t.Parallel()

		_, err := RefreshTokenFile(
			context.Background(),
			Config{ClientSecret: "client-secret", CallbackURL: "https://127.0.0.1:8182/callback"},
			TokenFile{CreationTimestamp: time.Now().Unix(), Token: TokenData{RefreshToken: "refresh-token"}},
			nil,
		)

		require.Error(t, err)
		assert.ErrorContains(t, err, "client_id")
	})
}

type refreshRequestSnapshot struct {
	method        string
	path          string
	authorization string
	contentType   string
	form          url.Values
}

func refreshTestConfig(oauthBaseURL string) Config {
	return Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		CallbackURL:  "https://127.0.0.1:8182/callback",
		OAuthBaseURL: oauthBaseURL,
	}
}
