package schwab

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// ClientConfig holds configuration for a Schwab API client.
// It is exported so sub-packages (marketdata, trader) can construct it with
// defaults and call ApplyOptions to apply user-provided options.
type ClientConfig struct {
	Token             string
	HTTPClient        *http.Client
	BaseURL           *url.URL
	OptionError       error
	ResponseBodyLimit int64
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

// ApplyOptions applies a slice of Option functions to cfg.
func ApplyOptions(cfg *ClientConfig, opts []Option) {
	for _, opt := range opts {
		opt(cfg)
	}
}
