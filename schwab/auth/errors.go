package auth

// AuthRequiredError indicates no valid tokens exist and login is needed.
//
//nolint:revive // Auth-prefixed names are intentional for clarity when used as auth.AuthRequiredError.
type AuthRequiredError struct{ Msg string }

// Error returns a human-readable representation of the auth required error.
func (e *AuthRequiredError) Error() string {
	return e.Msg
}

// AuthExpiredError indicates the refresh token has expired or been revoked.
//
//nolint:revive // Auth-prefixed names are intentional for clarity when used as auth.AuthExpiredError.
type AuthExpiredError struct{ Msg string }

// Error returns a human-readable representation of the auth expired error.
func (e *AuthExpiredError) Error() string {
	return e.Msg
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
