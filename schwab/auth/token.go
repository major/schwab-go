package auth

import (
	"context"
	"time"
)

const (
	accessTokenExpiryBuffer = 30
	refreshTokenMaxAge      = 7 * 24 * 60 * 60
)

// TokenData holds the OAuth2 token response fields from Schwab.
type TokenData struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	ExpiresAt    int64  `json:"expires_at"` // Unix timestamp
}

// TokenFile wraps TokenData with a creation timestamp for persistence.
type TokenFile struct {
	CreationTimestamp int64     `json:"creation_timestamp"` // Unix timestamp
	Token             TokenData `json:"token"`
}

// TokenStore is the interface for persisting and loading OAuth2 tokens.
// Implementations must be safe for concurrent use.
type TokenStore interface {
	Save(ctx context.Context, tf TokenFile) error
	Load(ctx context.Context) (TokenFile, error)
}

// IsAccessTokenExpired reports whether the access token in tf has expired,
// using a 30-second buffer to account for clock skew.
func IsAccessTokenExpired(tf TokenFile) bool {
	return time.Now().Unix() >= tf.Token.ExpiresAt-accessTokenExpiryBuffer
}

// IsRefreshTokenStale reports whether the refresh token in tf is older
// than 7 days and should be considered invalid.
func IsRefreshTokenStale(tf TokenFile) bool {
	return time.Now().Unix() >= tf.CreationTimestamp+refreshTokenMaxAge
}
