package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		token string
		want  string
	}{
		{
			name: "empty",
			want: "",
		},
		{
			name:  "short token",
			token: "short",
			want:  redactedValue,
		},
		{
			name:  "boundary length token",
			token: "1234567890",
			want:  redactedValue,
		},
		{
			name:  "long token",
			token: "abcdef1234567890",
			want:  "abcdef...7890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := RedactToken(tt.token)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRedactClientID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		clientID string
		want     string
	}{
		{
			name: "empty",
			want: "",
		},
		{
			name:     "short client ID",
			clientID: "short",
			want:     redactedValue,
		},
		{
			name:     "boundary length client ID",
			clientID: "12345678",
			want:     redactedValue,
		},
		{
			name:     "long client ID",
			clientID: "ABCD12345678WXYZ",
			want:     "ABCD...WXYZ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := RedactClientID(tt.clientID)

			assert.Equal(t, tt.want, got)
		})
	}
}
