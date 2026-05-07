package auth

import (
	"context"
	"errors"
	"net/http"
)

// LoginOption configures the login flow.
type LoginOption func(*loginConfig)

type loginConfig struct {
	httpClient *http.Client
}

// WithLoginHTTPClient sets the HTTP client used for token exchange.
// If not set, http.DefaultClient is used.
func WithLoginHTTPClient(c *http.Client) LoginOption {
	return func(cfg *loginConfig) {
		cfg.httpClient = c
	}
}

// Login runs the full OAuth2 authorization code flow:
//  1. Builds the authorization URL
//  2. Starts the HTTPS callback server
//  3. Calls urlHandler with the authorization URL
//  4. Waits for the callback with the authorization code
//  5. Exchanges the code for tokens
//  6. Saves tokens to the store
//  7. Returns a Provider for ongoing token management
//
// The urlHandler function receives the Schwab authorization URL and
// is responsible for presenting it to the user. For browser-based
// flows, pass a function that opens the URL (e.g., via os/exec with
// xdg-open). For headless/SSH environments, pass a function that
// prints or logs the URL for the user to visit manually. Login
// blocks until the OAuth callback arrives regardless of how the URL
// is handled.
func Login(ctx context.Context, cfg Config, store TokenStore, urlHandler func(string) error, opts ...LoginOption) (*Provider, error) {
	loginCfg := loginConfig{httpClient: http.DefaultClient}
	for _, opt := range opts {
		opt(&loginCfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if store == nil {
		return nil, errors.New("token store is required")
	}

	if urlHandler == nil {
		return nil, errors.New("urlHandler is required")
	}

	authorizeURL, expectedState, err := AuthorizeURL(cfg)
	if err != nil {
		return nil, err
	}

	serverCtx, cancelServer := context.WithCancel(ctx)
	defer cancelServer()

	results, errs, shutdown, err := StartCallbackServer(serverCtx, cfg.CallbackURL)
	if err != nil {
		return nil, err
	}
	defer shutdown()

	if err := urlHandler(authorizeURL); err != nil {
		return nil, err
	}

	var result CallbackResult
	select {
	case result = <-results:
		if result.State != expectedState {
			return nil, &AuthCallbackError{Reason: "state mismatch"}
		}
	case err := <-errs:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	tokenFile, err := ExchangeCode(ctx, cfg, result.Code, loginCfg.httpClient)
	if err != nil {
		return nil, err
	}

	if err := store.Save(ctx, tokenFile); err != nil {
		return nil, err
	}

	return NewProvider(cfg, store, loginCfg.httpClient)
}
