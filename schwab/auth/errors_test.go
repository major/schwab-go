package auth

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schwab "github.com/major/schwab-go/schwab"
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

func TestAuthErrorsCanBeJoinedWithCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("underlying failure")
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "required",
			err:  errors.Join(&AuthRequiredError{Msg: "login required"}, cause),
		},
		{
			name: "expired",
			err:  errors.Join(&AuthExpiredError{Msg: "token expired"}, cause),
		},
		{
			name: "callback",
			err:  errors.Join(&AuthCallbackError{Msg: "state mismatch", Code: 400}, cause),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.ErrorIs(t, tt.err, cause)
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

func TestAuthErrorClassifiers(t *testing.T) {
	t.Parallel()

	cause := errors.New("underlying failure")
	tests := []struct {
		name         string
		err          error
		wantRequired bool
		wantExpired  bool
		wantCallback bool
	}{
		{
			name:         "nil",
			wantRequired: false,
			wantExpired:  false,
			wantCallback: false,
		},
		{
			name:         "unrelated",
			err:          cause,
			wantRequired: false,
			wantExpired:  false,
			wantCallback: false,
		},
		{
			name:         "required",
			err:          &AuthRequiredError{Msg: "login required"},
			wantRequired: true,
			wantExpired:  false,
			wantCallback: false,
		},
		{
			name:         "expired",
			err:          &AuthExpiredError{Msg: "token expired"},
			wantRequired: false,
			wantExpired:  true,
			wantCallback: false,
		},
		{
			name:         "callback",
			err:          &AuthCallbackError{Msg: "state mismatch", Code: 400},
			wantRequired: false,
			wantExpired:  false,
			wantCallback: true,
		},
		{
			name:         "wrapped required",
			err:          errors.Join(&AuthRequiredError{Msg: "login required"}, cause),
			wantRequired: true,
			wantExpired:  false,
			wantCallback: false,
		},
		{
			name:         "fmt wrapped required",
			err:          fmt.Errorf("load token: %w", &AuthRequiredError{Msg: "login required"}),
			wantRequired: true,
			wantExpired:  false,
			wantCallback: false,
		},
		{
			name:         "wrapped expired",
			err:          errors.Join(&AuthExpiredError{Msg: "token expired"}, cause),
			wantRequired: false,
			wantExpired:  true,
			wantCallback: false,
		},
		{
			name:         "fmt wrapped expired",
			err:          fmt.Errorf("refresh token: %w", &AuthExpiredError{Msg: "token expired"}),
			wantRequired: false,
			wantExpired:  true,
			wantCallback: false,
		},
		{
			name:         "API unauthorized",
			err:          &schwab.APIError{StatusCode: http.StatusUnauthorized, Message: "unauthorized"},
			wantRequired: false,
			wantExpired:  true,
			wantCallback: false,
		},
		{
			name: "wrapped API unauthorized",
			err: fmt.Errorf(
				"get accounts: %w",
				&schwab.APIError{StatusCode: http.StatusUnauthorized, Message: "unauthorized"},
			),
			wantRequired: false,
			wantExpired:  true,
			wantCallback: false,
		},
		{
			name:         "API forbidden",
			err:          &schwab.APIError{StatusCode: http.StatusForbidden, Message: "forbidden"},
			wantRequired: false,
			wantExpired:  false,
			wantCallback: false,
		},
		{
			name:         "wrapped callback",
			err:          errors.Join(&AuthCallbackError{Msg: "state mismatch", Code: 400}, cause),
			wantRequired: false,
			wantExpired:  false,
			wantCallback: true,
		},
		{
			name:         "fmt wrapped callback",
			err:          fmt.Errorf("login callback: %w", &AuthCallbackError{Msg: "state mismatch", Code: 400}),
			wantRequired: false,
			wantExpired:  false,
			wantCallback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equalf(t, tt.wantRequired, IsRequired(tt.err), "IsRequired(%v)", tt.err)
			assert.Equalf(t, tt.wantExpired, IsExpired(tt.err), "IsExpired(%v)", tt.err)
			assert.Equalf(t, tt.wantCallback, IsCallback(tt.err), "IsCallback(%v)", tt.err)
		})
	}
}
