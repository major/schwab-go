package schwab

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithToken(t *testing.T) {
	cfg := &ClientConfig{}
	opt := WithToken("test-token-123")
	opt(cfg)
	require.Equal(t, "test-token-123", cfg.Token)
}

func TestWithHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 0}
	cfg := &ClientConfig{}
	opt := WithHTTPClient(customClient)
	opt(cfg)
	require.Equal(t, customClient, cfg.HTTPClient)
}

func TestWithHTTPClient_Nil(t *testing.T) {
	existingClient := &http.Client{Timeout: 0}
	cfg := &ClientConfig{HTTPClient: existingClient}
	opt := WithHTTPClient(nil)
	opt(cfg)
	require.Equal(t, existingClient, cfg.HTTPClient)
}

func TestWithBaseURL_Valid(t *testing.T) {
	cfg := &ClientConfig{}
	opt := WithBaseURL("https://example.com/api")
	opt(cfg)
	require.NotNil(t, cfg.BaseURL)
	require.Equal(t, "https://example.com/api", cfg.BaseURL.String())
}

func TestWithBaseURL_Invalid(t *testing.T) {
	existingURL, err := url.Parse("https://existing.com")
	require.NoError(t, err)

	for _, rawURL := range []string{"://invalid", "", "relative/path"} {
		t.Run(rawURL, func(t *testing.T) {
			cfg := &ClientConfig{BaseURL: existingURL}
			opt := WithBaseURL(rawURL)
			opt(cfg)
			require.Equal(t, existingURL, cfg.BaseURL)
		})
	}
}

func TestApplyOptions(t *testing.T) {
	customClient := &http.Client{Timeout: 0}
	opts := []Option{
		WithToken("my-token"),
		WithHTTPClient(customClient),
		WithBaseURL("https://api.schwab.com"),
	}
	cfg := &ClientConfig{}
	ApplyOptions(cfg, opts)

	require.Equal(t, "my-token", cfg.Token)
	require.Equal(t, customClient, cfg.HTTPClient)
	require.NotNil(t, cfg.BaseURL)
	require.Equal(t, "https://api.schwab.com", cfg.BaseURL.String())
}
