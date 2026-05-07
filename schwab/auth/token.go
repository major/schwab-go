package auth

import (
	"context"
	"time"
)

const (
	accessTokenExpiryBuffer = 5 * 60
	refreshTokenMaxAge      = 6*24*60*60 + 12*60*60
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

// TokenStatus describes token lifecycle state without exposing token secrets.
type TokenStatus struct {
	AccessTokenExpiresAt  time.Time
	AccessTokenExpired    bool
	RefreshTokenCreatedAt time.Time
	RefreshTokenExpiresAt time.Time
	RefreshTokenStale     bool
	CanRefresh            bool
	LoginRequired         bool
}

// IsAccessTokenExpired reports whether the access token in tf has expired,
// using a 5-minute buffer to account for clock skew and in-flight requests.
func IsAccessTokenExpired(tf TokenFile) bool {
	return time.Now().Unix() >= tf.Token.ExpiresAt-accessTokenExpiryBuffer
}

// IsRefreshTokenStale reports whether the refresh token in tf is older
// than 6.5 days and should be considered invalid.
func IsRefreshTokenStale(tf TokenFile) bool {
	return time.Now().Unix() >= tf.CreationTimestamp+refreshTokenMaxAge
}

// InspectToken reports token lifecycle state for tf at now without refreshing or
// saving tokens. It does not validate token authenticity or contact Schwab.
func InspectToken(tf TokenFile, now time.Time) TokenStatus {
	if now.IsZero() {
		now = time.Now()
	}

	if tf.Token.AccessToken == "" || tf.Token.RefreshToken == "" || tf.CreationTimestamp == 0 {
		return TokenStatus{LoginRequired: true}
	}

	accessTokenExpiresAt := time.Unix(tf.Token.ExpiresAt, 0).UTC()
	refreshTokenCreatedAt := time.Unix(tf.CreationTimestamp, 0).UTC()
	refreshTokenExpiresAt := time.Unix(tf.CreationTimestamp+refreshTokenMaxAge, 0).UTC()
	accessTokenExpired := now.Unix() >= tf.Token.ExpiresAt-accessTokenExpiryBuffer
	refreshTokenStale := now.Unix() >= tf.CreationTimestamp+refreshTokenMaxAge

	return TokenStatus{
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		AccessTokenExpired:    accessTokenExpired,
		RefreshTokenCreatedAt: refreshTokenCreatedAt,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
		RefreshTokenStale:     refreshTokenStale,
		CanRefresh:            !refreshTokenStale,
		LoginRequired:         accessTokenExpired && refreshTokenStale,
	}
}
