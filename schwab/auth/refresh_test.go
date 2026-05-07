package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshAccessToken(t *testing.T) {
	t.Parallel()

	t.Run("success returns token file", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

	t.Run("request format uses basic auth and refresh form body", func(t *testing.T) {
		t.Parallel()

		requests := make(chan refreshRequestSnapshot, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
