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

func TestExchangeCode(t *testing.T) {
	t.Parallel()

	t.Run("success returns token file", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, err := fmt.Fprint(w, `{"access_token":"access-token","token_type":"Bearer","expires_in":1800,"refresh_token":"refresh-token","scope":"api"}`)
			require.NoError(t, err)
		}))
		t.Cleanup(server.Close)

		cfg := exchangeTestConfig(server.URL)
		before := time.Now().Unix()
		tokenFile, err := ExchangeCode(context.Background(), cfg, "authorization-code", server.Client())
		after := time.Now().Unix()

		require.NoError(t, err)
		assert.GreaterOrEqual(t, tokenFile.CreationTimestamp, before)
		assert.LessOrEqual(t, tokenFile.CreationTimestamp, after)
		assert.Equal(t, "access-token", tokenFile.Token.AccessToken)
		assert.Equal(t, "Bearer", tokenFile.Token.TokenType)
		assert.Equal(t, 1800, tokenFile.Token.ExpiresIn)
		assert.Equal(t, "refresh-token", tokenFile.Token.RefreshToken)
		assert.Equal(t, "api", tokenFile.Token.Scope)
		assert.Equal(t, tokenFile.CreationTimestamp+1800, tokenFile.Token.ExpiresAt)
	})

	t.Run("http error returns status", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "invalid grant", http.StatusBadRequest)
		}))
		t.Cleanup(server.Close)

		_, err := ExchangeCode(context.Background(), exchangeTestConfig(server.URL), "bad-code", server.Client())

		require.Error(t, err)
		assert.ErrorContains(t, err, "status 400")
	})

	t.Run("invalid json response returns error", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := fmt.Fprint(w, `not-json`)
			require.NoError(t, err)
		}))
		t.Cleanup(server.Close)

		_, err := ExchangeCode(context.Background(), exchangeTestConfig(server.URL), "authorization-code", server.Client())

		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to parse token exchange response")
	})

	t.Run("request format uses basic auth and form body", func(t *testing.T) {
		t.Parallel()

		requests := make(chan exchangeRequestSnapshot, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			require.NoError(t, err)

			requests <- exchangeRequestSnapshot{
				method:        r.Method,
				path:          r.URL.Path,
				authorization: r.Header.Get("Authorization"),
				contentType:   r.Header.Get("Content-Type"),
				form:          r.PostForm,
			}

			w.Header().Set("Content-Type", "application/json")
			_, err = fmt.Fprint(w, `{"access_token":"access-token","token_type":"Bearer","expires_in":1800,"refresh_token":"refresh-token"}`)
			require.NoError(t, err)
		}))
		t.Cleanup(server.Close)

		cfg := exchangeTestConfig(server.URL + "/")
		_, err := ExchangeCode(context.Background(), cfg, "authorization-code", server.Client())
		require.NoError(t, err)

		snapshot := <-requests
		assert.Equal(t, http.MethodPost, snapshot.method)
		assert.Equal(t, "/token", snapshot.path)
		assert.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte("client-id:client-secret")), snapshot.authorization)
		assert.Equal(t, "application/x-www-form-urlencoded", snapshot.contentType)
		assert.Equal(t, "authorization_code", snapshot.form.Get("grant_type"))
		assert.Equal(t, "authorization-code", snapshot.form.Get("code"))
		assert.Equal(t, cfg.CallbackURL, snapshot.form.Get("redirect_uri"))
	})

	t.Run("empty oauth base url uses default token endpoint", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, err := fmt.Fprint(w, `{"access_token":"access-token","token_type":"Bearer","expires_in":1800,"refresh_token":"refresh-token"}`)
			require.NoError(t, err)
		}))
		t.Cleanup(server.Close)

		rewriteTransport := newExchangeRewriteTransport(t, server.URL)
		client := &http.Client{Transport: rewriteTransport}
		cfg := exchangeTestConfig("")

		_, err := ExchangeCode(context.Background(), cfg, "authorization-code", client)

		require.NoError(t, err)
		assert.Equal(t, defaultOAuthBaseURL+"/token", <-rewriteTransport.originalURLs)
	})
}

type exchangeRequestSnapshot struct {
	method        string
	path          string
	authorization string
	contentType   string
	form          url.Values
}

type exchangeRewriteTransport struct {
	targetURL    *url.URL
	base         http.RoundTripper
	originalURLs chan string
}

func newExchangeRewriteTransport(t *testing.T, targetServerURL string) *exchangeRewriteTransport {
	t.Helper()

	targetURL, err := url.Parse(targetServerURL)
	require.NoError(t, err)

	return &exchangeRewriteTransport{
		targetURL:    targetURL,
		base:         http.DefaultTransport,
		originalURLs: make(chan string, 1),
	}
}

func (t *exchangeRewriteTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	t.originalURLs <- request.URL.String()

	rewrittenRequest := request.Clone(request.Context())
	rewrittenURL := *request.URL
	rewrittenURL.Scheme = t.targetURL.Scheme
	rewrittenURL.Host = t.targetURL.Host
	rewrittenRequest.URL = &rewrittenURL

	return t.base.RoundTrip(rewrittenRequest)
}

func exchangeTestConfig(oauthBaseURL string) Config {
	return Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		CallbackURL:  "https://127.0.0.1:8182/callback",
		OAuthBaseURL: oauthBaseURL,
	}
}
