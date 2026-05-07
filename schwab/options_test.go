package schwab

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

const relativeBaseURLError = "invalid base URL \"relative/path\": absolute URL with scheme and host required"

type staticTokenProvider struct {
	token string
}

func (p staticTokenProvider) Token(context.Context) (string, error) {
	return p.token, nil
}

func TestWithToken(t *testing.T) {
	cfg := &ClientConfig{}
	opt := WithToken("test-token-123")
	opt(cfg)
	require.Equal(t, "test-token-123", cfg.Token)
}

func TestWithTokenProvider(t *testing.T) {
	provider := staticTokenProvider{token: "dynamic-token-123"}
	cfg := &ClientConfig{}
	opt := WithTokenProvider(provider)
	opt(cfg)
	require.Equal(t, provider, cfg.TokenProvider)
}

func TestWithTokenProvider_Nil(t *testing.T) {
	existingProvider := staticTokenProvider{token: "keep-me"}
	cfg := &ClientConfig{TokenProvider: existingProvider}
	opt := WithTokenProvider(nil)
	opt(cfg)
	require.Equal(t, existingProvider, cfg.TokenProvider)
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

func TestWithTLSConfig(t *testing.T) {
	tlsCfg := &tls.Config{ServerName: "api.schwabapi.com", MinVersion: tls.VersionTLS12}
	cfg := &ClientConfig{}

	WithTLSConfig(tlsCfg)(cfg)

	require.NotNil(t, cfg.HTTPClient)
	transport, ok := cfg.HTTPClient.Transport.(*http.Transport)
	require.True(t, ok)
	require.Same(t, tlsCfg, transport.TLSClientConfig)
	require.NotSame(t, http.DefaultTransport, transport)
}

func TestWithTLSConfig_Nil(t *testing.T) {
	existingClient := &http.Client{Timeout: 0}
	cfg := &ClientConfig{HTTPClient: existingClient}

	WithTLSConfig(nil)(cfg)

	require.Equal(t, existingClient, cfg.HTTPClient)
}

func TestWithTLSConfig_ClonesExistingHTTPClientTransport(t *testing.T) {
	tlsCfg := &tls.Config{ServerName: "api.schwabapi.com", MinVersion: tls.VersionTLS12}
	transport := &http.Transport{MaxIdleConns: 42}
	existingClient := &http.Client{Transport: transport}
	cfg := &ClientConfig{HTTPClient: existingClient}

	WithTLSConfig(tlsCfg)(cfg)

	require.NotSame(t, existingClient, cfg.HTTPClient)
	configuredTransport, ok := cfg.HTTPClient.Transport.(*http.Transport)
	require.True(t, ok)
	require.NotSame(t, transport, configuredTransport)
	require.Equal(t, 42, configuredTransport.MaxIdleConns)
	require.Same(t, tlsCfg, configuredTransport.TLSClientConfig)
}

func TestWithTLSConfig_DoesNotOverrideCustomRoundTripper(t *testing.T) {
	tlsCfg := &tls.Config{ServerName: "api.schwabapi.com", MinVersion: tls.VersionTLS12}
	roundTripper := staticRoundTripper{}
	cfg := &ClientConfig{HTTPClient: &http.Client{Transport: roundTripper}}

	WithTLSConfig(tlsCfg)(cfg)

	require.Equal(t, roundTripper, cfg.HTTPClient.Transport)
}

func TestWithTLSConfig_OrderWithHTTPClient(t *testing.T) {
	tlsCfg := &tls.Config{ServerName: "api.schwabapi.com", MinVersion: tls.VersionTLS12}
	customClient := &http.Client{Timeout: 0}

	cfg := &ClientConfig{}
	ApplyOptions(cfg, []Option{WithTLSConfig(tlsCfg), WithHTTPClient(customClient)})
	require.Equal(t, customClient, cfg.HTTPClient)

	cfg = &ClientConfig{}
	ApplyOptions(cfg, []Option{WithHTTPClient(customClient), WithTLSConfig(tlsCfg)})
	require.NotEqual(t, customClient, cfg.HTTPClient)
	transport, ok := cfg.HTTPClient.Transport.(*http.Transport)
	require.True(t, ok)
	require.Same(t, tlsCfg, transport.TLSClientConfig)
}

func TestWithBaseURL_Valid(t *testing.T) {
	cfg := &ClientConfig{}
	opt := WithBaseURL("https://example.com/api")
	opt(cfg)
	require.NotNil(t, cfg.BaseURL)
	require.Equal(t, "https://example.com/api", cfg.BaseURL.String())
	require.NoError(t, cfg.OptionError)
}

type staticRoundTripper struct{}

func (staticRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, http.ErrNotSupported
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
