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

const invalidGrantError = "invalid_grant"

type tokenErrorResponse struct {
	Error string `json:"error"`
}

// RefreshAccessToken exchanges a Schwab OAuth2 refresh token for new access
// and refresh tokens using the configured OAuth token endpoint.
func RefreshAccessToken(ctx context.Context, cfg Config, refreshToken string, httpClient *http.Client) (TokenFile, error) {
	client := httpClient
	if client == nil {
		client = http.DefaultClient
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		tokenExchangeURL(cfg.OAuthBaseURL),
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return TokenFile{}, fmt.Errorf("failed to create token refresh request: %w", err)
	}

	credentials := base64.StdEncoding.EncodeToString([]byte(cfg.ClientID + ":" + cfg.ClientSecret))
	request.Header.Set("Authorization", "Basic "+credentials)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)
	if err != nil {
		return TokenFile{}, fmt.Errorf("token refresh request failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return TokenFile{}, tokenRefreshHTTPError(response)
	}

	var token TokenData
	if err := json.NewDecoder(response.Body).Decode(&token); err != nil {
		return TokenFile{}, fmt.Errorf("failed to parse token refresh response: %w", err)
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

func tokenRefreshHTTPError(response *http.Response) error {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("token refresh failed with status %d and unreadable response body: %w", response.StatusCode, err)
	}

	if response.StatusCode == http.StatusBadRequest {
		var tokenErr tokenErrorResponse
		if err := json.Unmarshal(body, &tokenErr); err == nil && tokenErr.Error == invalidGrantError {
			return &AuthExpiredError{}
		}
	}

	trimmedBody := strings.TrimSpace(string(body))
	if trimmedBody == "" {
		return fmt.Errorf("token refresh failed with status %d", response.StatusCode)
	}

	return fmt.Errorf("token refresh failed with status %d: %s", response.StatusCode, trimmedBody)
}
