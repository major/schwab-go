package schwab

import (
	"net/http"
	"net/url"
)

// ClientConfig holds configuration for a Schwab API client.
// It is exported so sub-packages (marketdata, trader) can construct it with
// defaults and call ApplyOptions to apply user-provided options.
type ClientConfig struct {
	Token      string
	HTTPClient *http.Client
	BaseURL    *url.URL
}

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

// WithBaseURL overrides the default base URL. An invalid URL is silently ignored.
func WithBaseURL(rawURL string) Option {
	return func(cfg *ClientConfig) {
		u, err := url.Parse(rawURL)
		if err == nil {
			cfg.BaseURL = u
		}
	}
}

// ApplyOptions applies a slice of Option functions to cfg.
func ApplyOptions(cfg *ClientConfig, opts []Option) {
	for _, opt := range opts {
		opt(cfg)
	}
}
