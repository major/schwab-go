package auth

import (
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
			err:  &AuthRequiredError{Msg: "login required"},
			want: "login required",
		},
		{
			name: "expired",
			err:  &AuthExpiredError{Msg: "token expired"},
			want: "token expired",
		},
		{
			name: "callback",
			err:  &AuthCallbackError{Msg: "state mismatch", Code: 400},
			want: "state mismatch",
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
			err:  &AuthRequiredError{Msg: "login required"},
			check: func(t *testing.T, err error) {
				t.Helper()

				var target *AuthRequiredError
				require.ErrorAs(t, err, &target)
				require.NotNil(t, target)
				assert.Equal(t, "login required", target.Msg)
			},
		},
		{
			name: "expired",
			err:  &AuthExpiredError{Msg: "token expired"},
			check: func(t *testing.T, err error) {
				t.Helper()

				var target *AuthExpiredError
				require.ErrorAs(t, err, &target)
				require.NotNil(t, target)
				assert.Equal(t, "token expired", target.Msg)
			},
		},
		{
			name: "callback",
			err:  &AuthCallbackError{Msg: "state mismatch", Code: 400},
			check: func(t *testing.T, err error) {
				t.Helper()

				var target *AuthCallbackError
				require.ErrorAs(t, err, &target)
				require.NotNil(t, target)
				assert.Equal(t, "state mismatch", target.Msg)
				assert.Equal(t, 400, target.Code)
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
