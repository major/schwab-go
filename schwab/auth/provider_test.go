package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider(t *testing.T) {
	t.Parallel()

	t.Run("valid non-expired token returns without refresh", func(t *testing.T) {
		t.Parallel()

		var refreshCalls atomic.Int64
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			refreshCalls.Add(1)
			http.Error(w, "unexpected refresh", http.StatusInternalServerError)
		}))
		t.Cleanup(server.Close)

		store := newProviderMemoryStore(providerTokenFile(
			"access-token",
			"refresh-token",
			time.Now().Add(time.Hour),
			time.Now(),
		))
		provider := newTestProvider(t, server.URL, store, server.Client())

		accessToken, err := provider.Token(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "access-token", accessToken)
		assert.Equal(t, int64(0), refreshCalls.Load())
	})

	t.Run("expired access token refreshes and returns new token", func(t *testing.T) {
		t.Parallel()

		var refreshCalls atomic.Int64
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			refreshCalls.Add(1)
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/token", r.URL.Path)

			w.Header().Set("Content-Type", "application/json")
			_, err := fmt.Fprint(
				w,
				`{"access_token":"new-access-token","token_type":"Bearer","expires_in":1800,"refresh_token":"new-refresh-token","scope":"api"}`,
			)
			assert.NoError(t, err)
		}))
		t.Cleanup(server.Close)

		createdAt := time.Now().Add(-time.Hour)
		store := newProviderMemoryStore(providerTokenFile(
			"old-access-token",
			"old-refresh-token",
			time.Now().Add(-time.Hour),
			createdAt,
		))
		provider := newTestProvider(t, server.URL, store, server.Client())

		accessToken, err := provider.Token(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "new-access-token", accessToken)
		assert.Equal(t, int64(1), refreshCalls.Load())

		savedTokenFile, err := store.Load(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "new-access-token", savedTokenFile.Token.AccessToken)
		assert.Equal(t, "new-refresh-token", savedTokenFile.Token.RefreshToken)
		assert.Equal(t, createdAt.Unix(), savedTokenFile.CreationTimestamp)
	})

	t.Run("explicit refresh ignores active access token and saves refreshed token", func(t *testing.T) {
		t.Parallel()

		var refreshCalls atomic.Int64
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			refreshCalls.Add(1)
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/token", r.URL.Path)

			w.Header().Set("Content-Type", "application/json")
			_, err := fmt.Fprint(
				w,
				`{"access_token":"forced-new-access-token","token_type":"Bearer","expires_in":1800,"refresh_token":"forced-new-refresh-token","scope":"api"}`,
			)
			assert.NoError(t, err)
		}))
		t.Cleanup(server.Close)

		createdAt := time.Now().Add(-time.Hour)
		store := newProviderMemoryStore(providerTokenFile(
			"active-access-token",
			"refresh-token",
			time.Now().Add(time.Hour),
			createdAt,
		))
		provider := newTestProvider(t, server.URL, store, server.Client())

		refreshedTokenFile, err := provider.Refresh(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "forced-new-access-token", refreshedTokenFile.Token.AccessToken)
		assert.Equal(t, createdAt.Unix(), refreshedTokenFile.CreationTimestamp)
		assert.Equal(t, int64(1), refreshCalls.Load())
		assert.Equal(t, int64(1), store.saveCalls.Load())

		savedTokenFile, err := store.Load(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "forced-new-access-token", savedTokenFile.Token.AccessToken)
		assert.Equal(t, createdAt.Unix(), savedTokenFile.CreationTimestamp)
	})

	t.Run("explicit refresh returns auth expired for stale refresh token", func(t *testing.T) {
		t.Parallel()

		var refreshCalls atomic.Int64
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			refreshCalls.Add(1)
			http.Error(w, "unexpected refresh", http.StatusInternalServerError)
		}))
		t.Cleanup(server.Close)

		store := newProviderMemoryStore(providerTokenFile(
			"active-access-token",
			"stale-refresh-token",
			time.Now().Add(time.Hour),
			time.Now().Add(-8*24*time.Hour),
		))
		provider := newTestProvider(t, server.URL, store, server.Client())

		refreshedTokenFile, err := provider.Refresh(context.Background())

		require.Error(t, err)
		assert.Empty(t, refreshedTokenFile)
		var expiredErr *AuthExpiredError
		require.ErrorAs(t, err, &expiredErr)
		assert.Equal(t, int64(0), refreshCalls.Load())
		assert.Equal(t, int64(0), store.saveCalls.Load())
	})

	t.Run("explicit refresh propagates load error", func(t *testing.T) {
		t.Parallel()

		loadErr := errors.New("load failed")
		store := newProviderMemoryStore(TokenFile{})
		store.loadErr = loadErr
		provider := newTestProvider(t, "https://127.0.0.1:8182/oauth", store, nil)

		refreshedTokenFile, err := provider.Refresh(context.Background())

		require.Error(t, err)
		assert.Empty(t, refreshedTokenFile)
		assert.ErrorIs(t, err, loadErr)
	})

	t.Run("explicit refresh propagates refresh error", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "refresh failed", http.StatusInternalServerError)
		}))
		t.Cleanup(server.Close)

		store := newProviderMemoryStore(providerTokenFile(
			"active-access-token",
			"refresh-token",
			time.Now().Add(time.Hour),
			time.Now().Add(-time.Hour),
		))
		provider := newTestProvider(t, server.URL, store, server.Client())

		refreshedTokenFile, err := provider.Refresh(context.Background())

		require.Error(t, err)
		assert.Empty(t, refreshedTokenFile)
		assert.Equal(t, int64(0), store.saveCalls.Load())
	})

	t.Run("explicit refresh propagates save error", func(t *testing.T) {
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

		saveErr := errors.New("save failed")
		store := newProviderMemoryStore(providerTokenFile(
			"active-access-token",
			"refresh-token",
			time.Now().Add(time.Hour),
			time.Now().Add(-time.Hour),
		))
		store.saveErr = saveErr
		provider := newTestProvider(t, server.URL, store, server.Client())

		refreshedTokenFile, err := provider.Refresh(context.Background())

		require.Error(t, err)
		assert.Empty(t, refreshedTokenFile)
		assert.ErrorIs(t, err, saveErr)
	})

	t.Run("status inspects stored token without refresh or save", func(t *testing.T) {
		t.Parallel()

		now := time.Date(2026, time.May, 7, 12, 0, 0, 0, time.UTC)
		createdAt := now.Add(-time.Hour)
		store := newProviderMemoryStore(providerTokenFile(
			"access-token",
			"refresh-token",
			now.Add(time.Hour),
			createdAt,
		))
		provider := newTestProvider(t, "https://127.0.0.1:8182/oauth", store, nil)

		status, err := provider.Status(context.Background(), now)

		require.NoError(t, err)
		assert.Equal(t, now.Add(time.Hour), status.AccessTokenExpiresAt)
		assert.False(t, status.AccessTokenExpired)
		assert.Equal(t, createdAt, status.RefreshTokenCreatedAt)
		assert.Equal(t, createdAt.Add(refreshTokenMaxAge*time.Second), status.RefreshTokenExpiresAt)
		assert.False(t, status.RefreshTokenStale)
		assert.True(t, status.CanRefresh)
		assert.False(t, status.LoginRequired)
		assert.Equal(t, int64(0), store.saveCalls.Load())
	})

	t.Run("status reports login required for missing token", func(t *testing.T) {
		t.Parallel()

		store := newProviderMemoryStore(TokenFile{})
		store.loadErr = &AuthRequiredError{Msg: authRequiredLoginMessage}
		provider := newTestProvider(t, "https://127.0.0.1:8182/oauth", store, nil)

		status, err := provider.Status(context.Background(), time.Now())

		require.NoError(t, err)
		assert.True(t, status.LoginRequired)
		assert.False(t, status.CanRefresh)
		assert.Equal(t, int64(0), store.saveCalls.Load())
	})

	t.Run("status propagates non-auth load error", func(t *testing.T) {
		t.Parallel()

		loadErr := errors.New("read failed")
		store := newProviderMemoryStore(TokenFile{})
		store.loadErr = loadErr
		provider := newTestProvider(t, "https://127.0.0.1:8182/oauth", store, nil)

		status, err := provider.Status(context.Background(), time.Now())

		require.Error(t, err)
		assert.Empty(t, status)
		assert.ErrorIs(t, err, loadErr)
	})

	t.Run("stale refresh token returns auth expired without HTTP call", func(t *testing.T) {
		t.Parallel()

		var refreshCalls atomic.Int64
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			refreshCalls.Add(1)
			http.Error(w, "unexpected refresh", http.StatusInternalServerError)
		}))
		t.Cleanup(server.Close)

		store := newProviderMemoryStore(providerTokenFile(
			"old-access-token",
			"old-refresh-token",
			time.Now().Add(-time.Hour),
			time.Now().Add(-8*24*time.Hour),
		))
		provider := newTestProvider(t, server.URL, store, server.Client())

		accessToken, err := provider.Token(context.Background())

		require.Error(t, err)
		assert.Empty(t, accessToken)
		var expiredErr *AuthExpiredError
		require.ErrorAs(t, err, &expiredErr)
		assert.Equal(t, int64(0), refreshCalls.Load())
	})

	t.Run("no token returns auth required", func(t *testing.T) {
		t.Parallel()

		store := newProviderMemoryStore(TokenFile{})
		store.loadErr = &AuthRequiredError{Msg: authRequiredLoginMessage}
		provider := newTestProvider(t, "https://127.0.0.1:8182/oauth", store, nil)

		accessToken, err := provider.Token(context.Background())

		require.Error(t, err)
		assert.Empty(t, accessToken)
		var requiredErr *AuthRequiredError
		require.ErrorAs(t, err, &requiredErr)
	})

	t.Run("concurrent expired token refreshes only once", func(t *testing.T) {
		t.Parallel()

		var refreshCalls atomic.Int64
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			refreshCalls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, err := fmt.Fprint(
				w,
				`{"access_token":"shared-new-access-token","token_type":"Bearer","expires_in":1800,"refresh_token":"shared-new-refresh-token","scope":"api"}`,
			)
			assert.NoError(t, err)
		}))
		t.Cleanup(server.Close)

		store := newProviderMemoryStore(
			providerTokenFile("old-access-token", "old-refresh-token", time.Now().Add(-time.Hour), time.Now()),
		)
		provider := newTestProvider(t, server.URL, store, server.Client())

		const goroutines = 10
		var wg sync.WaitGroup
		start := make(chan struct{})
		results := make(chan providerTokenResult, goroutines)

		wg.Add(goroutines)
		for range goroutines {
			go func() {
				defer wg.Done()
				<-start
				accessToken, err := provider.Token(context.Background())
				results <- providerTokenResult{accessToken: accessToken, err: err}
			}()
		}

		close(start)
		wg.Wait()
		close(results)

		for result := range results {
			require.NoError(t, result.err)
			assert.Equal(t, "shared-new-access-token", result.accessToken)
		}
		assert.Equal(t, int64(1), refreshCalls.Load())
	})

	t.Run("load error propagates", func(t *testing.T) {
		t.Parallel()

		loadErr := errors.New("load failed")
		store := newProviderMemoryStore(TokenFile{})
		store.loadErr = loadErr
		provider := newTestProvider(t, "https://127.0.0.1:8182/oauth", store, nil)

		accessToken, err := provider.Token(context.Background())

		require.Error(t, err)
		assert.Empty(t, accessToken)
		assert.ErrorIs(t, err, loadErr)
	})

	t.Run("nil store returns error", func(t *testing.T) {
		t.Parallel()

		provider, err := NewProvider(providerTestConfig("https://127.0.0.1:8182/oauth"), nil, nil)

		require.Error(t, err)
		assert.Nil(t, provider)
		assert.ErrorContains(t, err, "token store is required")
	})

	t.Run("invalid config returns validation error", func(t *testing.T) {
		t.Parallel()

		provider, err := NewProvider(Config{}, newProviderMemoryStore(TokenFile{}), nil)

		require.Error(t, err)
		assert.Nil(t, provider)
		assert.ErrorContains(t, err, "client_id is required")
	})
}

func TestNewFileProvider(t *testing.T) {
	t.Parallel()

	t.Run("valid file provider", func(t *testing.T) {
		t.Parallel()

		provider, err := NewFileProvider(
			providerTestConfig("https://127.0.0.1:8182/oauth"),
			filepath.Join(t.TempDir(), "tokens.json"),
			nil,
		)

		require.NoError(t, err)
		require.NotNil(t, provider)
	})

	t.Run("empty token path returns error", func(t *testing.T) {
		t.Parallel()

		provider, err := NewFileProvider(
			providerTestConfig("https://127.0.0.1:8182/oauth"),
			"",
			nil,
		)

		require.Error(t, err)
		assert.Nil(t, provider)
		assert.ErrorContains(t, err, "token path")
	})
}

type providerMemoryStore struct {
	mu        sync.Mutex
	token     TokenFile
	loadErr   error
	saveErr   error
	saveCalls atomic.Int64
}

func newProviderMemoryStore(token TokenFile) *providerMemoryStore {
	return &providerMemoryStore{token: token}
}

func (s *providerMemoryStore) Save(ctx context.Context, token TokenFile) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.saveErr != nil {
		return s.saveErr
	}

	s.saveCalls.Add(1)
	s.token = token
	return nil
}

func (s *providerMemoryStore) Load(ctx context.Context) (TokenFile, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.loadErr != nil {
		return TokenFile{}, s.loadErr
	}

	return s.token, nil
}

type providerTokenResult struct {
	accessToken string
	err         error
}

func newTestProvider(t *testing.T, oauthBaseURL string, store TokenStore, httpClient *http.Client) *Provider {
	t.Helper()

	provider, err := NewProvider(providerTestConfig(oauthBaseURL), store, httpClient)
	require.NoError(t, err)
	return provider
}

func providerTestConfig(oauthBaseURL string) Config {
	return Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		CallbackURL:  "https://127.0.0.1:8182/callback",
		OAuthBaseURL: oauthBaseURL,
	}
}

func providerTokenFile(accessToken, refreshToken string, expiresAt, createdAt time.Time) TokenFile {
	return TokenFile{
		CreationTimestamp: createdAt.Unix(),
		Token: TokenData{
			AccessToken:  accessToken,
			TokenType:    "Bearer",
			ExpiresIn:    1800,
			RefreshToken: refreshToken,
			Scope:        "api",
			ExpiresAt:    expiresAt.Unix(),
		},
	}
}
