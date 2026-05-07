package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "required",
			err:  &AuthRequiredError{},
			want: "authentication required: run the login flow to obtain a token",
		},
		{
			name: "expired",
			err:  &AuthExpiredError{},
			want: "authentication expired: refresh token is no longer valid, re-login required",
		},
		{
			name: "callback",
			err:  &AuthCallbackError{Reason: "state mismatch"},
			want: "OAuth callback error: state mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.err.Error()
			require.NotEmpty(t, got)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuthErrorsAs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		err   error
		check func(t *testing.T, err error)
	}{
		{
			name: "required",
			err:  &AuthRequiredError{},
			check: func(t *testing.T, err error) {
				t.Helper()

				var target *AuthRequiredError
				require.True(t, errors.As(err, &target))
				require.NotNil(t, target)
			},
		},
		{
			name: "expired",
			err:  &AuthExpiredError{},
			check: func(t *testing.T, err error) {
				t.Helper()

				var target *AuthExpiredError
				require.True(t, errors.As(err, &target))
				require.NotNil(t, target)
			},
		},
		{
			name: "callback",
			err:  &AuthCallbackError{Reason: "state mismatch"},
			check: func(t *testing.T, err error) {
				t.Helper()

				var target *AuthCallbackError
				require.True(t, errors.As(err, &target))
				require.NotNil(t, target)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.check(t, tt.err)
		})
	}
}
