package auth

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"
)

const defaultProviderHTTPTimeout = 30 * time.Second

// Provider returns OAuth2 access tokens with automatic refresh. It is safe for
// concurrent use.
type Provider struct {
	mu    sync.Mutex
	cfg   Config
	store TokenStore
	http  *http.Client // for token refresh HTTP calls
}

// NewProvider creates a Provider that manages tokens using the given
// config and store. The http client is optional; nil creates a new
// client with a 30-second timeout used only for refresh
// requests.
func NewProvider(cfg Config, store TokenStore, httpClient *http.Client) (*Provider, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}

	if store == nil {
		return nil, errors.New("token store is required")
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

type tokenProvider interface {
	Token(context.Context) (string, error)
}

var _ tokenProvider = (*Provider)(nil)

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
	if tokenFile.CreationTimestamp != 0 {
		refreshedTokenFile.CreationTimestamp = tokenFile.CreationTimestamp
	}

	err = p.store.Save(ctx, refreshedTokenFile)
	if err != nil {
		return "", err
	}

	return refreshedTokenFile.Token.AccessToken, nil
}
