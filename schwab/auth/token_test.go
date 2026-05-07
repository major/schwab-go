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
