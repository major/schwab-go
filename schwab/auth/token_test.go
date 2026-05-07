package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsAccessTokenExpired(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()
	tests := []struct {
		name string
		tf   TokenFile
		want bool
	}{
		{
			name: "future token is active",
			tf: TokenFile{
				Token: TokenData{ExpiresAt: now + int64(time.Hour.Seconds())},
			},
		},
		{
			name: "past token is expired",
			tf: TokenFile{
				Token: TokenData{ExpiresAt: now - int64(time.Minute.Seconds())},
			},
			want: true,
		},
		{
			name: "five minute buffer expires token",
			tf: TokenFile{
				Token: TokenData{ExpiresAt: now + int64((5 * time.Minute).Seconds())},
			},
			want: true,
		},
		{
			name: "six minutes remains active",
			tf: TokenFile{
				Token: TokenData{ExpiresAt: now + int64((6 * time.Minute).Seconds())},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsAccessTokenExpired(tt.tf)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsRefreshTokenStale(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()
	tests := []struct {
		name string
		tf   TokenFile
		want bool
	}{
		{
			name: "six day token is fresh",
			tf: TokenFile{
				CreationTimestamp: now - int64((6 * 24 * time.Hour).Seconds()),
			},
		},
		{
			name: "eight day token is stale",
			tf: TokenFile{
				CreationTimestamp: now - int64((8 * 24 * time.Hour).Seconds()),
			},
			want: true,
		},
		{
			name: "six and a half day token is stale",
			tf: TokenFile{
				CreationTimestamp: now - int64((156 * time.Hour).Seconds()),
			},
			want: true,
		},
		{
			name: "six days and twelve hours minus one minute is fresh",
			tf: TokenFile{
				CreationTimestamp: now - int64((156*time.Hour - time.Minute).Seconds()),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsRefreshTokenStale(tt.tf)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInspectToken(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.May, 7, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		tf   TokenFile
		want TokenStatus
	}{
		{
			name: "valid token can refresh",
			tf: TokenFile{
				CreationTimestamp: now.Add(-time.Hour).Unix(),
				Token: TokenData{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
					ExpiresAt:    now.Add(time.Hour).Unix(),
				},
			},
			want: TokenStatus{
				AccessTokenExpiresAt:  now.Add(time.Hour),
				RefreshTokenCreatedAt: now.Add(-time.Hour),
				RefreshTokenExpiresAt: now.Add(-time.Hour).Add(refreshTokenMaxAge * time.Second),
				CanRefresh:            true,
			},
		},
		{
			name: "access token inside buffer is expired but refreshable",
			tf: TokenFile{
				CreationTimestamp: now.Add(-time.Hour).Unix(),
				Token: TokenData{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
					ExpiresAt:    now.Add(4 * time.Minute).Unix(),
				},
			},
			want: TokenStatus{
				AccessTokenExpiresAt:  now.Add(4 * time.Minute),
				AccessTokenExpired:    true,
				RefreshTokenCreatedAt: now.Add(-time.Hour),
				RefreshTokenExpiresAt: now.Add(-time.Hour).Add(refreshTokenMaxAge * time.Second),
				CanRefresh:            true,
			},
		},
		{
			name: "stale refresh token requires login",
			tf: TokenFile{
				CreationTimestamp: now.Add(-8 * 24 * time.Hour).Unix(),
				Token: TokenData{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
					ExpiresAt:    now.Add(-time.Hour).Unix(),
				},
			},
			want: TokenStatus{
				AccessTokenExpiresAt:  now.Add(-time.Hour),
				AccessTokenExpired:    true,
				RefreshTokenCreatedAt: now.Add(-8 * 24 * time.Hour),
				RefreshTokenExpiresAt: now.Add(-8 * 24 * time.Hour).Add(refreshTokenMaxAge * time.Second),
				RefreshTokenStale:     true,
				LoginRequired:         true,
			},
		},
		{
			name: "stale refresh token with active access token warns without requiring login",
			tf: TokenFile{
				CreationTimestamp: now.Add(-8 * 24 * time.Hour).Unix(),
				Token: TokenData{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
					ExpiresAt:    now.Add(time.Hour).Unix(),
				},
			},
			want: TokenStatus{
				AccessTokenExpiresAt:  now.Add(time.Hour),
				RefreshTokenCreatedAt: now.Add(-8 * 24 * time.Hour),
				RefreshTokenExpiresAt: now.Add(-8 * 24 * time.Hour).Add(refreshTokenMaxAge * time.Second),
				RefreshTokenStale:     true,
			},
		},
		{
			name: "empty token requires login",
			want: TokenStatus{LoginRequired: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := InspectToken(tt.tf, now)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInspectTokenZeroNowUsesCurrentTime(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)
	createdAt := now.Add(-time.Hour)
	expiresAt := now.Add(time.Hour)
	tf := TokenFile{
		CreationTimestamp: createdAt.Unix(),
		Token: TokenData{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresAt:    expiresAt.Unix(),
		},
	}

	got := InspectToken(tf, time.Time{})

	assert.Equal(t, expiresAt, got.AccessTokenExpiresAt)
	assert.False(t, got.AccessTokenExpired)
	assert.Equal(t, createdAt, got.RefreshTokenCreatedAt)
	assert.Equal(t, createdAt.Add(refreshTokenMaxAge*time.Second), got.RefreshTokenExpiresAt)
	assert.False(t, got.RefreshTokenStale)
	assert.True(t, got.CanRefresh)
	assert.False(t, got.LoginRequired)
}
