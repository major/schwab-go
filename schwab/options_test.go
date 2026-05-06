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

func TestWithUserAgent(t *testing.T) {
	cfg := &ClientConfig{}
	opt := WithUserAgent("schwab-test/1.0")
	opt(cfg)
	require.Equal(t, "schwab-test/1.0", cfg.Headers.Get("User-Agent"))
}

func TestWithHeader(t *testing.T) {
	cfg := &ClientConfig{}
	opt := WithHeader("X-Request-ID", "abc123")
	opt(cfg)
	require.Equal(t, "abc123", cfg.Headers.Get("X-Request-ID"))
}

func TestWithHeader_EmptyName(t *testing.T) {
	cfg := &ClientConfig{Headers: http.Header{"X-Keep": []string{"yes"}}}
	opt := WithHeader("", "ignored")
	opt(cfg)
	require.Equal(t, "yes", cfg.Headers.Get("X-Keep"))
}

func TestWithHeaders(t *testing.T) {
	headers := http.Header{
		"X-Trace-Id": []string{"one", "two"},
		"X-Empty":    []string{""},
	}
	cfg := &ClientConfig{}
	opt := WithHeaders(headers)
	opt(cfg)
	headers.Add("X-Trace-Id", "mutated")

	require.Equal(t, []string{"one", "two"}, cfg.Headers.Values("X-Trace-Id"))
	require.Equal(t, []string{""}, cfg.Headers.Values("X-Empty"))
}

func TestApplyOptions(t *testing.T) {
	customClient := &http.Client{Timeout: 0}
	opts := []Option{
		WithToken("my-token"),
		WithHTTPClient(customClient),
		WithBaseURL("https://api.schwab.com"),
		WithResponseBodyLimit(2048),
		WithUserAgent("schwab-test/1.0"),
		WithHeader("X-Trace-Id", "trace-123"),
	}
	cfg := &ClientConfig{}
	ApplyOptions(cfg, opts)

	require.Equal(t, "my-token", cfg.Token)
	require.Equal(t, customClient, cfg.HTTPClient)
	require.NotNil(t, cfg.BaseURL)
	require.Equal(t, "https://api.schwab.com", cfg.BaseURL.String())
	require.NoError(t, cfg.OptionError)
	require.Equal(t, int64(2048), cfg.ResponseBodyLimit)
	require.Equal(t, "schwab-test/1.0", cfg.Headers.Get("User-Agent"))
	require.Equal(t, "trace-123", cfg.Headers.Get("X-Trace-Id"))
}
