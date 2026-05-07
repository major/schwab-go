package schwab

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// TokenProvider supplies bearer tokens dynamically. Implementations
// must be safe for concurrent use.
type TokenProvider interface {
	Token(ctx context.Context) (string, error)
}

// ClientConfig holds configuration for a Schwab API client.
// It is exported so sub-packages (marketdata, trader) can construct it with
// defaults and call ApplyOptions to apply user-provided options.
type ClientConfig struct {
	Token             string
	TokenProvider     TokenProvider
	HTTPClient        *http.Client
	BaseURL           *url.URL
	OptionError       error
	ResponseBodyLimit int64
	Headers           http.Header
}

// DefaultResponseBodyLimit is the maximum response body size read by default.
const DefaultResponseBodyLimit int64 = 10 << 20

// Option is a functional option for configuring a Schwab API client.
type Option func(*ClientConfig)

// WithToken sets the bearer token for API authentication.
func WithToken(token string) Option {
	return func(cfg *ClientConfig) {
		cfg.Token = token
	}
}

// WithTokenProvider sets the bearer token provider for API authentication.
// A nil value is ignored.
func WithTokenProvider(tp TokenProvider) Option {
	return func(cfg *ClientConfig) {
		if tp == nil {
			return
		}
		cfg.TokenProvider = tp
	}
}

// WithHTTPClient sets a custom HTTP client. A nil value is ignored.
func WithHTTPClient(c *http.Client) Option {
	return func(cfg *ClientConfig) {
		if c != nil {
			cfg.HTTPClient = c
		}
	}
}

// WithBaseURL overrides the default base URL.
// Sub-clients may append their own API path prefix to this URL.
// Invalid or relative URLs are recorded as option errors so request creation fails before any HTTP call.
func WithBaseURL(rawURL string) Option {
	return func(cfg *ClientConfig) {
		u, err := url.Parse(rawURL)
		if err != nil {
			cfg.OptionError = errors.Join(cfg.OptionError, fmt.Errorf("invalid base URL %q: %w", rawURL, err))
			return
		}
		if u.Scheme == "" || u.Host == "" {
			cfg.OptionError = errors.Join(
				cfg.OptionError,
				fmt.Errorf("invalid base URL %q: absolute URL with scheme and host required", rawURL),
			)
			return
		}
		cfg.BaseURL = u
	}
}

// WithResponseBodyLimit sets the maximum response body size read by the client.
// Non-positive values are ignored.
func WithResponseBodyLimit(limit int64) Option {
	return func(cfg *ClientConfig) {
		if limit > 0 {
			cfg.ResponseBodyLimit = limit
		}
	}
}

// WithUserAgent sets the User-Agent header sent with each request.
func WithUserAgent(userAgent string) Option {
	return WithHeader("User-Agent", userAgent)
}

// WithHeader sets a request header sent with each request. An empty name is ignored.
// Accept, Authorization, and Content-Type are controlled by the library and cannot be set this way.
func WithHeader(name, value string) Option {
	return func(cfg *ClientConfig) {
		if name == "" {
			return
		}
		if cfg.Headers == nil {
			cfg.Headers = http.Header{}
		}
		cfg.Headers.Set(name, value)
	}
}

// WithHeaders sets request headers sent with each request. Header values are copied.
// Accept, Authorization, and Content-Type are controlled by the library and cannot be set this way.
func WithHeaders(headers http.Header) Option {
	return func(cfg *ClientConfig) {
		for name, values := range headers {
			if name == "" {
				continue
			}
			if cfg.Headers == nil {
				cfg.Headers = http.Header{}
			}
			cfg.Headers.Del(name)
			for _, value := range values {
				cfg.Headers.Add(name, value)
			}
		}
	}
}

// ApplyOptions applies a slice of Option functions to cfg.
func ApplyOptions(cfg *ClientConfig, opts []Option) {
	for _, opt := range opts {
		opt(cfg)
	}
}
