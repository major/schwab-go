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

	refreshedTokenFile, err := p.refreshLocked(ctx, tokenFile)
	if err != nil {
		return "", err
	}

	return refreshedTokenFile.Token.AccessToken, nil
}

// Refresh refreshes the stored access token regardless of current access-token
// expiry and saves the refreshed token through the configured TokenStore.
// It preserves the original TokenFile.CreationTimestamp so refresh-token age
// remains accurate across access-token rotations.
func (p *Provider) Refresh(ctx context.Context) (TokenFile, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	tokenFile, err := p.store.Load(ctx)
	if err != nil {
		return TokenFile{}, err
	}

	return p.refreshLocked(ctx, tokenFile)
}

// Status reports the currently stored token lifecycle state without refreshing
// or saving tokens. Missing token files return a status with LoginRequired set
// and no error; other TokenStore load errors are returned.
func (p *Provider) Status(ctx context.Context, now time.Time) (TokenStatus, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	tokenFile, err := p.store.Load(ctx)
	if err != nil {
		if _, ok := errors.AsType[*AuthRequiredError](err); ok {
			return TokenStatus{LoginRequired: true}, nil
		}
		return TokenStatus{}, err
	}

	return InspectToken(tokenFile, now), nil
}

func (p *Provider) refreshLocked(ctx context.Context, tokenFile TokenFile) (TokenFile, error) {
	if IsRefreshTokenStale(tokenFile) {
		return TokenFile{}, &AuthExpiredError{Msg: "refresh token expired or revoked"}
	}

	refreshedTokenFile, err := RefreshAccessToken(ctx, p.cfg, tokenFile.Token.RefreshToken, p.http)
	if err != nil {
		return TokenFile{}, err
	}
	if tokenFile.CreationTimestamp != 0 {
		refreshedTokenFile.CreationTimestamp = tokenFile.CreationTimestamp
	}

	err = p.store.Save(ctx, refreshedTokenFile)
	if err != nil {
		return TokenFile{}, err
	}

	return refreshedTokenFile, nil
}
