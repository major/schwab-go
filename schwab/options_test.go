package schwab

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

const relativeBaseURLError = "invalid base URL \"relative/path\": absolute URL with scheme and host required"

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
	require.NoError(t, cfg.OptionError)
}

func TestWithBaseURL_Invalid(t *testing.T) {
	existingURL, err := url.Parse("https://existing.com")
	require.NoError(t, err)

	for _, tt := range []struct {
		name    string
		rawURL  string
		wantErr string
	}{
		{name: "parse error", rawURL: "://invalid", wantErr: "invalid base URL \"://invalid\":"},
		{name: "empty", rawURL: "", wantErr: "invalid base URL \"\": absolute URL with scheme and host required"},
		{name: "relative", rawURL: "relative/path", wantErr: relativeBaseURLError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ClientConfig{BaseURL: existingURL}
			opt := WithBaseURL(tt.rawURL)
			opt(cfg)
			require.Equal(t, existingURL, cfg.BaseURL)
			require.Error(t, cfg.OptionError)
			require.ErrorContains(t, cfg.OptionError, tt.wantErr)
		})
	}
}

func TestWithBaseURL_JoinsInvalidOptions(t *testing.T) {
	cfg := &ClientConfig{}
	WithBaseURL("")(cfg)
	WithBaseURL("relative/path")(cfg)

	require.Error(t, cfg.OptionError)
	require.ErrorContains(t, cfg.OptionError, "invalid base URL \"\": absolute URL with scheme and host required")
	require.ErrorContains(t, cfg.OptionError, relativeBaseURLError)
}

func TestWithResponseBodyLimit(t *testing.T) {
	cfg := &ClientConfig{}
	opt := WithResponseBodyLimit(1234)
	opt(cfg)
	require.Equal(t, int64(1234), cfg.ResponseBodyLimit)
}

func TestWithResponseBodyLimit_NonPositive(t *testing.T) {
	for _, limit := range []int64{0, -1} {
		t.Run(strconv.FormatInt(limit, 10), func(t *testing.T) {
			cfg := &ClientConfig{ResponseBodyLimit: 42}
			opt := WithResponseBodyLimit(limit)
			opt(cfg)
			require.Equal(t, int64(42), cfg.ResponseBodyLimit)
		})
	}
}

func TestApplyOptions(t *testing.T) {
	customClient := &http.Client{Timeout: 0}
	opts := []Option{
		WithToken("my-token"),
		WithHTTPClient(customClient),
		WithBaseURL("https://api.schwab.com"),
		WithResponseBodyLimit(2048),
	}
	cfg := &ClientConfig{}
	ApplyOptions(cfg, opts)

	require.Equal(t, "my-token", cfg.Token)
	require.Equal(t, customClient, cfg.HTTPClient)
	require.NotNil(t, cfg.BaseURL)
	require.Equal(t, "https://api.schwab.com", cfg.BaseURL.String())
	require.NoError(t, cfg.OptionError)
	require.Equal(t, int64(2048), cfg.ResponseBodyLimit)
}
