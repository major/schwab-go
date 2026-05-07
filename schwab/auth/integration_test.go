package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	schwab "github.com/major/schwab-go/schwab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_FullOAuthFlow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	requests := make(chan integrationTokenRequest, 2)
	server := newIntegrationTokenServer(t, requests)
	cfg := newLoginTestConfig(t, server.URL)
	tokenPath := filepath.Join(t.TempDir(), "tokens.json")
	store := NewFileTokenStore(tokenPath)
	callbackErrs := make(chan error, 1)
	var capturedAuthorizeURL string

	urlHandler := func(authorizeURL string) error {
		capturedAuthorizeURL = authorizeURL
		state := parseIntegrationAuthorizeState(t, authorizeURL)
		go func() {
			callbackErrs <- getLoginCallback(cfg.CallbackURL, url.Values{"code": {"integration-code"}, "state": {state}})
		}()
		return nil
	}

	provider, err := Login(ctx, cfg, store, urlHandler, WithLoginHTTPClient(server.Client()))

	require.NoError(t, err)
	require.NoError(t, <-callbackErrs)
	require.NotNil(t, provider)
	assert.NotEmpty(t, capturedAuthorizeURL)

	exchangeRequest := <-requests
	require.NoError(t, exchangeRequest.err)
	assert.Equal(t, http.MethodPost, exchangeRequest.method)
	assert.Equal(t, "/token", exchangeRequest.path)
	assert.Equal(t, "authorization_code", exchangeRequest.grantType)
	assert.Equal(t, "integration-code", exchangeRequest.code)
	assert.Equal(t, cfg.CallbackURL, exchangeRequest.redirectURI)

	accessToken, err := provider.Token(ctx)
	require.NoError(t, err)
	assert.Equal(t, "integration-access-token", accessToken)

	tf, err := store.Load(ctx)
	require.NoError(t, err)
	tf.Token.ExpiresAt = time.Now().Unix() - 100
	tf.CreationTimestamp = time.Now().Unix()
	require.NoError(t, store.Save(ctx, tf))

	refreshedToken, err := provider.Token(ctx)
	require.NoError(t, err)
	assert.Equal(t, "integration-refreshed-token", refreshedToken)

	refreshRequest := <-requests
	require.NoError(t, refreshRequest.err)
	assert.Equal(t, http.MethodPost, refreshRequest.method)
	assert.Equal(t, "/token", refreshRequest.path)
	assert.Equal(t, "refresh_token", refreshRequest.grantType)
	assert.Equal(t, "integration-refresh-token", refreshRequest.refreshToken)

	fileInfo, err := os.Stat(tokenPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())
}

func TestIntegration_ProviderFromSaved(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewFileTokenStore(filepath.Join(t.TempDir(), "tokens.json"))
	cfg := providerTestConfig("https://127.0.0.1:8182/oauth")
	require.NoError(t, store.Save(ctx, providerTokenFile("saved-access-token", "saved-refresh-token", time.Now().Add(time.Hour), time.Now())))
	provider, err := NewProvider(cfg, store, nil)
	require.NoError(t, err)

	accessToken, err := provider.Token(ctx)

	require.NoError(t, err)
	assert.Equal(t, "saved-access-token", accessToken)
}

func TestIntegration_WithTokenProvider(t *testing.T) {
	t.Parallel()

	store := newProviderMemoryStore(providerTokenFile("root-access-token", "root-refresh-token", time.Now().Add(time.Hour), time.Now()))
	provider := newTestProvider(t, "https://127.0.0.1:8182/oauth", store, nil)

	option := schwab.WithTokenProvider(provider)
	config := schwab.ClientConfig{}
	option(&config)

	require.NotNil(t, option)
	assert.Equal(t, provider, config.TokenProvider)
}

type integrationTokenRequest struct {
	method       string
	path         string
	grantType    string
	code         string
	refreshToken string
	redirectURI  string
	err          error
}

func newIntegrationTokenServer(t *testing.T, requests chan<- integrationTokenRequest) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parseErr := r.ParseForm()
		grantType := r.PostForm.Get("grant_type")
		requests <- integrationTokenRequest{
			method:       r.Method,
			path:         r.URL.Path,
			grantType:    grantType,
			code:         r.PostForm.Get("code"),
			refreshToken: r.PostForm.Get("refresh_token"),
			redirectURI:  r.PostForm.Get("redirect_uri"),
			err:          parseErr,
		}
		if parseErr != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		switch grantType {
		case "authorization_code":
			_, err := fmt.Fprint(w, `{"access_token":"integration-access-token","token_type":"Bearer","expires_in":1800,"refresh_token":"integration-refresh-token","scope":"api"}`)
			assert.NoError(t, err)
		case "refresh_token":
			_, err := fmt.Fprint(w, `{"access_token":"integration-refreshed-token","token_type":"Bearer","expires_in":1800,"refresh_token":"integration-new-refresh-token","scope":"api"}`)
			assert.NoError(t, err)
		default:
			http.Error(w, "unsupported grant", http.StatusBadRequest)
		}
	}))
	t.Cleanup(server.Close)

	return server
}

func parseIntegrationAuthorizeState(t *testing.T, authorizeURL string) string {
	t.Helper()

	parsedURL, err := url.Parse(authorizeURL)
	require.NoError(t, err)

	return parsedURL.Query().Get("state")
}
