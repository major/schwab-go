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
			name: "thirty second buffer expires token",
			tf: TokenFile{
				Token: TokenData{ExpiresAt: now + 30},
			},
			want: true,
		},
		{
			name: "thirty one seconds remains active",
			tf: TokenFile{
				Token: TokenData{ExpiresAt: now + 31},
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
			name: "seven day token is stale",
			tf: TokenFile{
				CreationTimestamp: now - int64((7 * 24 * time.Hour).Seconds()),
			},
			want: true,
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
