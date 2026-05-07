package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxOAuthErrorBodyBytes = 1 << 20

// ExchangeCode exchanges a Schwab OAuth2 authorization code for access and
// refresh tokens using the configured OAuth token endpoint.
func ExchangeCode(
	ctx context.Context,
	cfg Config,
	code string,
	httpClient *http.Client,
) (TokenFile, error) {
	client := httpClient
	if client == nil {
		client = http.DefaultClient
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", cfg.CallbackURL)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		tokenExchangeURL(cfg),
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return TokenFile{}, fmt.Errorf("failed to create token exchange request: %w", err)
	}

	credentials := base64.StdEncoding.EncodeToString(
		[]byte(url.QueryEscape(cfg.ClientID) + ":" + url.QueryEscape(cfg.ClientSecret)),
	)
	request.Header.Set("Authorization", "Basic "+credentials)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)
	if err != nil {
		return TokenFile{}, fmt.Errorf("token exchange request failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, readErr := io.ReadAll(io.LimitReader(response.Body, maxOAuthErrorBodyBytes))
		if readErr != nil {
			return TokenFile{}, fmt.Errorf(
				"token exchange failed with status %d and unreadable response body: %w",
				response.StatusCode,
				readErr,
			)
		}

		trimmedBody := strings.TrimSpace(string(body))
		if trimmedBody == "" {
			return TokenFile{}, fmt.Errorf("token exchange failed with status %d", response.StatusCode)
		}

		return TokenFile{}, fmt.Errorf("token exchange failed with status %d: %s", response.StatusCode, trimmedBody)
	}

	var token TokenData
	err = json.NewDecoder(response.Body).Decode(&token)
	if err != nil {
		return TokenFile{}, fmt.Errorf("failed to parse token exchange response: %w", err)
	}

	now := time.Now().Unix()
	if token.ExpiresAt == 0 && token.ExpiresIn > 0 {
		token.ExpiresAt = now + int64(token.ExpiresIn)
	}

	return TokenFile{
		CreationTimestamp: now,
		Token:             token,
	}, nil
}

func tokenExchangeURL(cfg Config) string {
	return strings.TrimRight(cfg.oauthBaseURL(), "/") + "/token"
}
