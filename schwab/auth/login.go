package auth

import (
	"context"
	"errors"
	"net/http"
	"sync"
)

const errStateMismatch = "state mismatch"

// LoginOption configures the login flow.
type LoginOption func(*loginConfig)

// LoginWaitFunc waits for the OAuth callback, exchanges the returned code,
// saves the resulting tokens, and returns a Provider.
type LoginWaitFunc func(context.Context) (*Provider, error)

type loginConfig struct {
	httpClient *http.Client
}

// WithLoginHTTPClient sets the HTTP client used for token exchange.
// If not set, [http.DefaultClient] is used.
func WithLoginHTTPClient(c *http.Client) LoginOption {
	return func(cfg *loginConfig) {
		if c != nil {
			cfg.httpClient = c
		}
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
func Login(
	ctx context.Context,
	cfg Config,
	store TokenStore,
	urlHandler func(string) error,
	opts ...LoginOption,
) (*Provider, error) {
	if urlHandler == nil {
		return nil, errors.New("urlHandler is required")
	}

	serverCtx, cancelServer := context.WithCancel(ctx)
	defer cancelServer()

	authorizeURL, wait, err := StartLogin(serverCtx, cfg, store, opts...)
	if err != nil {
		return nil, err
	}

	err = urlHandler(authorizeURL)
	if err != nil {
		return nil, err
	}

	return wait(ctx)
}

// StartLogin starts the OAuth2 callback listener and returns the Schwab
// authorization URL plus a wait function that completes the login.
//
// Callers should present the returned URL to the user, then call the wait
// function with the context that should bound callback waiting and token
// exchange. Cancel ctx to stop the callback listener if the returned wait
// function will not be called.
func StartLogin(
	ctx context.Context,
	cfg Config,
	store TokenStore,
	opts ...LoginOption,
) (string, LoginWaitFunc, error) {
	loginCfg := loginConfig{httpClient: http.DefaultClient}
	for _, opt := range opts {
		opt(&loginCfg)
	}

	err := cfg.Validate()
	if err != nil {
		return "", nil, err
	}

	if store == nil {
		return "", nil, errors.New("token store is required")
	}

	authorizeURL, expectedState, err := AuthorizeURL(cfg)
	if err != nil {
		return "", nil, err
	}

	serverCtx, cancelServer := context.WithCancel(ctx)
	results, errs, shutdown, err := StartCallbackServer(serverCtx, cfg.CallbackURL)
	if err != nil {
		cancelServer()
		return "", nil, err
	}

	var waitOnce sync.Once
	waitDone := make(chan struct{})
	var provider *Provider
	var waitErr error
	wait := func(waitCtx context.Context) (*Provider, error) {
		waitOnce.Do(func() {
			defer close(waitDone)
			defer cancelServer()
			defer shutdown()
			provider, waitErr = finishLogin(
				waitCtx,
				serverCtx,
				cfg,
				store,
				loginCfg.httpClient,
				expectedState,
				results,
				errs,
			)
		})
		<-waitDone
		return provider, waitErr
	}

	return authorizeURL, wait, nil
}

func finishLogin(
	waitCtx context.Context,
	serverCtx context.Context,
	cfg Config,
	store TokenStore,
	httpClient *http.Client,
	expectedState string,
	results <-chan CallbackResult,
	errs <-chan error,
) (*Provider, error) {
	var result CallbackResult
	select {
	case result = <-results:
		if result.State != expectedState {
			return nil, &AuthCallbackError{Msg: errStateMismatch, Code: http.StatusBadRequest}
		}
	case err := <-errs:
		return nil, err
	case <-waitCtx.Done():
		return nil, waitCtx.Err()
	case <-serverCtx.Done():
		return nil, serverCtx.Err()
	}

	tokenFile, err := ExchangeCode(waitCtx, cfg, result.Code, httpClient)
	if err != nil {
		return nil, err
	}

	err = store.Save(waitCtx, tokenFile)
	if err != nil {
		return nil, err
	}

	return NewProvider(cfg, store, httpClient)
}
