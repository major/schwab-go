package auth

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	t.Parallel()

	t.Run("full flow saves token and returns provider", func(t *testing.T) {
		t.Parallel()

		server := newLoginTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "/token", r.URL.Path)
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "test-code", r.PostForm.Get("code"))

			w.Header().Set("Content-Type", "application/json")
			_, err := fmt.Fprint(w, `{"access_token":"login-access-token","token_type":"Bearer","expires_in":1800,"refresh_token":"login-refresh-token","scope":"api"}`)
			require.NoError(t, err)
		})

		cfg := newLoginTestConfig(t, server.URL)
		store := newLoginMemoryStore()
		callbackErrs := make(chan error, 1)
		var capturedAuthorizeURL string
		urlHandler := func(authorizeURL string) error {
			capturedAuthorizeURL = authorizeURL
			go func() {
				callbackErrs <- sendLoginCallback(cfg.CallbackURL, authorizeURL, "test-code")
			}()
			return nil
		}

		provider, err := Login(context.Background(), cfg, store, urlHandler, WithLoginHTTPClient(server.Client()))

		require.NoError(t, err)
		require.NoError(t, <-callbackErrs)
		require.NotNil(t, provider)
		assert.NotEmpty(t, capturedAuthorizeURL)

		savedTokenFile, err := store.Load(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "login-access-token", savedTokenFile.Token.AccessToken)
		assert.Equal(t, "login-refresh-token", savedTokenFile.Token.RefreshToken)

		accessToken, err := provider.Token(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "login-access-token", accessToken)
	})

	t.Run("urlHandler error returns error", func(t *testing.T) {
		t.Parallel()

		server := newLoginTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "unexpected token exchange", http.StatusInternalServerError)
		})
		cfg := newLoginTestConfig(t, server.URL)
		store := newLoginMemoryStore()
		urlHandlerErr := errors.New("open browser failed")

		provider, err := Login(context.Background(), cfg, store, func(string) error {
			return urlHandlerErr
		}, WithLoginHTTPClient(server.Client()))

		require.Error(t, err)
		assert.Nil(t, provider)
		assert.ErrorIs(t, err, urlHandlerErr)
	})

	t.Run("state mismatch returns callback error", func(t *testing.T) {
		t.Parallel()

		var tokenExchangeCalls atomic.Int64
		server := newLoginTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
			tokenExchangeCalls.Add(1)
			http.Error(w, "unexpected token exchange", http.StatusInternalServerError)
		})
		cfg := newLoginTestConfig(t, server.URL)
		store := newLoginMemoryStore()
		callbackErrs := make(chan error, 1)
		urlHandler := func(authorizeURL string) error {
			_ = authorizeURL
			go func() {
				callbackErrs <- getLoginCallback(cfg.CallbackURL, url.Values{"code": {"test-code"}, "state": {"wrong-state"}})
			}()
			return nil
		}

		provider, err := Login(context.Background(), cfg, store, urlHandler, WithLoginHTTPClient(server.Client()))

		require.Error(t, err)
		require.NoError(t, <-callbackErrs)
		assert.Nil(t, provider)
		var callbackErr *AuthCallbackError
		assert.ErrorAs(t, err, &callbackErr)
		assert.Equal(t, "state mismatch", callbackErr.Reason)
		assert.Equal(t, int64(0), tokenExchangeCalls.Load())
	})

	t.Run("context cancellation returns context error", func(t *testing.T) {
		t.Parallel()

		var tokenExchangeCalls atomic.Int64
		server := newLoginTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
			tokenExchangeCalls.Add(1)
			http.Error(w, "unexpected token exchange", http.StatusInternalServerError)
		})
		cfg := newLoginTestConfig(t, server.URL)
		store := newLoginMemoryStore()
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		provider, err := Login(ctx, cfg, store, func(string) error {
			cancel()
			return nil
		}, WithLoginHTTPClient(server.Client()))

		require.Error(t, err)
		assert.Nil(t, provider)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, int64(0), tokenExchangeCalls.Load())
	})
}

type loginMemoryStore struct {
	mu    sync.Mutex
	token TokenFile
}

func newLoginMemoryStore() *loginMemoryStore {
	return &loginMemoryStore{}
}

func (s *loginMemoryStore) Save(ctx context.Context, token TokenFile) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	s.token = token
	return nil
}

func (s *loginMemoryStore) Load(ctx context.Context) (TokenFile, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.token, nil
}

func newLoginTokenServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func newLoginTestConfig(t *testing.T, oauthBaseURL string) Config {
	t.Helper()

	return Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		CallbackURL:  newLoginCallbackURL(t),
		OAuthBaseURL: oauthBaseURL,
	}
}

func newLoginCallbackURL(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()
	require.NoError(t, listener.Close())

	return "https://" + addr + "/callback"
}

func sendLoginCallback(callbackURL, authorizeURL, code string) error {
	parsedURL, err := url.Parse(authorizeURL)
	if err != nil {
		return fmt.Errorf("parse authorize URL: %w", err)
	}

	state := parsedURL.Query().Get("state")
	return getLoginCallback(callbackURL, url.Values{"code": {code}, "state": {state}})
}

func getLoginCallback(callbackURL string, query url.Values) error {
	parsedURL, err := url.Parse(callbackURL)
	if err != nil {
		return fmt.Errorf("parse callback URL: %w", err)
	}
	parsedURL.RawQuery = query.Encode()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Tests connect to an in-memory self-signed loopback certificate.
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(parsedURL.String())
	if err != nil {
		return fmt.Errorf("send OAuth callback: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send OAuth callback status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	return nil
}
