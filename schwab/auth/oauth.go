package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
)

const (
	oauthStateBytes = 32
)

// AuthorizeURL builds the Schwab OAuth2 authorization URL and returns the
// generated state value that callers must verify during callback handling.
//
//nolint:nonamedreturns // Named returns document the three distinct string, string, and error return values.
func AuthorizeURL(cfg Config) (authorizeURL string, state string, err error) {
	err = cfg.Validate()
	if err != nil {
		return "", "", err
	}

	state, err = randomOAuthState()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate OAuth state: %w", err)
	}

	parsedURL, err := url.Parse(strings.TrimRight(cfg.oauthBaseURL(), "/") + "/authorize")
	if err != nil {
		return "", "", fmt.Errorf("oauth authorization URL must be valid: %w", err)
	}

	query := url.Values{}
	query.Set("response_type", "code")
	query.Set("client_id", cfg.ClientID)
	query.Set("redirect_uri", cfg.CallbackURL)
	query.Set("state", state)
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), state, nil
}

func randomOAuthState() (string, error) {
	buf := make([]byte, oauthStateBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
