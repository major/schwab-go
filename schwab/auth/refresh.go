package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	invalidGrantError        = "invalid_grant"
	refreshTokenExpiredError = "refresh token expired or revoked"
)

type tokenErrorResponse struct {
	Error string `json:"error"`
}

// RefreshAccessToken exchanges a Schwab OAuth2 refresh token for new access
// and refresh tokens using the configured OAuth token endpoint.
func RefreshAccessToken(
	ctx context.Context,
	cfg Config,
	refreshToken string,
	httpClient *http.Client,
) (TokenFile, error) {
	err := cfg.Validate()
	if err != nil {
		return TokenFile{}, err
	}

	if refreshToken == "" {
		return TokenFile{}, errors.New("refresh token must not be empty")
	}

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
		tokenExchangeURL(cfg),
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return TokenFile{}, fmt.Errorf("failed to create token refresh request: %w", err)
	}

	credentials := base64.StdEncoding.EncodeToString(
		[]byte(url.QueryEscape(cfg.ClientID) + ":" + url.QueryEscape(cfg.ClientSecret)),
	)
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
	err = json.NewDecoder(response.Body).Decode(&token)
	if err != nil {
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

// RefreshTokenFile refreshes the token stored in tf and preserves the original
// creation timestamp when it is set so refresh-token age remains accurate.
func RefreshTokenFile(
	ctx context.Context,
	cfg Config,
	tf TokenFile,
	httpClient *http.Client,
) (TokenFile, error) {
	if err := cfg.Validate(); err != nil {
		return TokenFile{}, err
	}
	if tf.Token.RefreshToken == "" {
		return TokenFile{}, errors.New("refresh token must not be empty")
	}
	if IsRefreshTokenStale(tf) {
		return TokenFile{}, &AuthExpiredError{Msg: refreshTokenExpiredError}
	}

	refreshedTokenFile, err := RefreshAccessToken(ctx, cfg, tf.Token.RefreshToken, httpClient)
	if err != nil {
		return TokenFile{}, err
	}
	refreshedTokenFile.CreationTimestamp = tf.CreationTimestamp

	return refreshedTokenFile, nil
}

func tokenRefreshHTTPError(response *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(response.Body, maxOAuthErrorBodyBytes))
	if err != nil {
		return fmt.Errorf(
			"token refresh failed with status %d and unreadable response body: %w",
			response.StatusCode,
			err,
		)
	}

	if response.StatusCode == http.StatusBadRequest {
		var tokenErr tokenErrorResponse
		err = json.Unmarshal(body, &tokenErr)
		if err == nil && tokenErr.Error == invalidGrantError {
			return &AuthExpiredError{Msg: refreshTokenExpiredError}
		}
	}

	trimmedBody := strings.TrimSpace(string(body))
	if trimmedBody == "" {
		return fmt.Errorf("token refresh failed with status %d", response.StatusCode)
	}

	return fmt.Errorf("token refresh failed with status %d: %s", response.StatusCode, trimmedBody)
}
