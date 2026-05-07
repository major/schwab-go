package auth

import (
	"errors"

	schwab "github.com/major/schwab-go/schwab"
)

// AuthRequiredError indicates no valid tokens exist and login is needed.
//
//nolint:revive // Auth-prefixed names are intentional for clarity when used as auth.AuthRequiredError.
type AuthRequiredError struct{ Msg string }

// Error returns a human-readable representation of the auth required error.
func (e *AuthRequiredError) Error() string {
	return e.Msg
}

// IsRequired reports whether err indicates that a user must log in before the
// application can obtain Schwab API tokens.
func IsRequired(err error) bool {
	target, ok := errors.AsType[*AuthRequiredError](err)
	return ok && target != nil
}

// AuthExpiredError indicates the refresh token has expired or been revoked.
//
//nolint:revive // Auth-prefixed names are intentional for clarity when used as auth.AuthExpiredError.
type AuthExpiredError struct{ Msg string }

// Error returns a human-readable representation of the auth expired error.
func (e *AuthExpiredError) Error() string {
	return e.Msg
}

// IsExpired reports whether err indicates that stored auth credentials are
// expired, revoked, or otherwise unable to refresh the Schwab access token.
func IsExpired(err error) bool {
	target, ok := errors.AsType[*AuthExpiredError](err)
	return ok && target != nil || schwab.IsUnauthorized(err)
}

// AuthCallbackError indicates the OAuth callback failed.
//
//nolint:revive // Auth-prefixed names are intentional for clarity when used as auth.AuthCallbackError.
type AuthCallbackError struct {
	Msg  string
	Code int // HTTP status from Schwab, if any.
}

// Error returns a human-readable representation of the auth callback error.
func (e *AuthCallbackError) Error() string {
	return e.Msg
}

// IsCallback reports whether err indicates that the OAuth callback listener or
// callback request failed.
func IsCallback(err error) bool {
	target, ok := errors.AsType[*AuthCallbackError](err)
	return ok && target != nil
}
