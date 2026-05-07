package auth

import (
	"context"
	"net/http"
	"sync"
	"time"

	schwab "github.com/major/schwab-go/schwab"
)

const defaultProviderHTTPTimeout = 30 * time.Second

// Provider implements schwab.TokenProvider with automatic token
// refresh. It is safe for concurrent use.
type Provider struct {
	mu    sync.Mutex
	cfg   Config
	store TokenStore
	http  *http.Client // for token refresh HTTP calls
}

// NewProvider creates a Provider that manages tokens using the given
// config and store. The http client is optional (nil uses
// [http.DefaultClient]) and is used only for refresh requests.
func NewProvider(cfg Config, store TokenStore, httpClient *http.Client) (*Provider, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}

	client := httpClient
	if client == nil {
		client = &http.Client{Timeout: defaultProviderHTTPTimeout}
	}

	return &Provider{
		cfg:   cfg,
		store: store,
		http:  client,
	}, nil
}

// Verify interface compliance.
var _ schwab.TokenProvider = (*Provider)(nil)

// Token returns a valid access token, refreshing if expired.
// It is safe for concurrent use from multiple goroutines.
func (p *Provider) Token(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	tokenFile, err := p.store.Load(ctx)
	if err != nil {
		return "", err
	}

	if !IsAccessTokenExpired(tokenFile) {
		return tokenFile.Token.AccessToken, nil
	}

	if IsRefreshTokenStale(tokenFile) {
		return "", &AuthExpiredError{Msg: "refresh token expired or revoked"}
	}

	refreshedTokenFile, err := RefreshAccessToken(ctx, p.cfg, tokenFile.Token.RefreshToken, p.http)
	if err != nil {
		return "", err
	}

	err = p.store.Save(ctx, refreshedTokenFile)
	if err != nil {
		return "", err
	}

	return refreshedTokenFile.Token.AccessToken, nil
}
