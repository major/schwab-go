package httpclient

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithPathPrefix(t *testing.T) {
	for _, tt := range []struct {
		name       string
		baseURL    string
		pathPrefix string
		want       string
	}{
		{
			name:       "root URL",
			baseURL:    "https://custom.example.com",
			pathPrefix: "/marketdata/v1",
			want:       "https://custom.example.com/marketdata/v1",
		},
		{
			name:       "proxy path",
			baseURL:    "https://custom.example.com/proxy",
			pathPrefix: "/marketdata/v1",
			want:       "https://custom.example.com/proxy/marketdata/v1",
		},
		{
			name:       "already prefixed",
			baseURL:    "https://custom.example.com/marketdata/v1",
			pathPrefix: "/marketdata/v1",
			want:       "https://custom.example.com/marketdata/v1",
		},
		{
			name:       "already prefixed with trailing slash",
			baseURL:    "https://custom.example.com/proxy/marketdata/v1/",
			pathPrefix: "/marketdata/v1",
			want:       "https://custom.example.com/proxy/marketdata/v1/",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			baseURL, err := url.Parse(tt.baseURL)
			require.NoError(t, err)

			got := WithPathPrefix(baseURL, tt.pathPrefix)

			require.Equal(t, tt.want, got.String())
		})
	}
}

func TestWithPathPrefixClonesAlreadyPrefixedURL(t *testing.T) {
	baseURL, err := url.Parse("https://custom.example.com/marketdata/v1")
	require.NoError(t, err)

	got := WithPathPrefix(baseURL, "/marketdata/v1")

	require.NotSame(t, baseURL, got)
	require.Equal(t, baseURL.String(), got.String())
}
