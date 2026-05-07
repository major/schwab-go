package auth

import (
	"encoding/hex"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizeURL(t *testing.T) {
	t.Parallel()

	baseConfig := Config{
		ClientID:     "client-id",
		CallbackURL:  "https://127.0.0.1:8182/callback",
		OAuthBaseURL: "https://auth.example.test/oauth",
	}

	t.Run("valid config builds authorization query", func(t *testing.T) {
		t.Parallel()

		authorizeURL, state, err := AuthorizeURL(baseConfig)
		require.NoError(t, err)

		parsedURL := requireParsedAuthorizeURL(t, authorizeURL)
		query := parsedURL.Query()

		assert.Equal(t, "code", query.Get("response_type"))
		assert.Equal(t, baseConfig.ClientID, query.Get("client_id"))
		assert.Equal(t, baseConfig.CallbackURL, query.Get("redirect_uri"))
		assert.Equal(t, state, query.Get("state"))
	})

	t.Run("state values are unique", func(t *testing.T) {
		t.Parallel()

		_, firstState, err := AuthorizeURL(baseConfig)
		require.NoError(t, err)

		_, secondState, err := AuthorizeURL(baseConfig)
		require.NoError(t, err)

		assert.NotEqual(t, firstState, secondState)
	})

	t.Run("empty oauth base url uses default", func(t *testing.T) {
		t.Parallel()

		cfg := baseConfig
		cfg.OAuthBaseURL = ""

		authorizeURL, _, err := AuthorizeURL(cfg)
		require.NoError(t, err)

		parsedURL := requireParsedAuthorizeURL(t, authorizeURL)
		assert.Equal(t, "https", parsedURL.Scheme)
		assert.Equal(t, "api.schwabapi.com", parsedURL.Host)
		assert.Equal(t, "/v1/oauth/authorize", parsedURL.Path)
	})

	t.Run("custom oauth base url is used", func(t *testing.T) {
		t.Parallel()

		authorizeURL, _, err := AuthorizeURL(baseConfig)
		require.NoError(t, err)

		parsedURL := requireParsedAuthorizeURL(t, authorizeURL)
		assert.Equal(t, "https", parsedURL.Scheme)
		assert.Equal(t, "auth.example.test", parsedURL.Host)
		assert.Equal(t, "/oauth/authorize", parsedURL.Path)
	})

	t.Run("state is 32 byte hex value", func(t *testing.T) {
		t.Parallel()

		_, state, err := AuthorizeURL(baseConfig)
		require.NoError(t, err)

		decodedState, err := hex.DecodeString(state)
		require.NoError(t, err)
		assert.Len(t, state, 64)
		assert.Len(t, decodedState, 32)
	})
}

func requireParsedAuthorizeURL(t *testing.T, authorizeURL string) *url.URL {
	t.Helper()

	parsedURL, err := url.Parse(authorizeURL)
	require.NoError(t, err)
	return parsedURL
}
