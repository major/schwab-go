package auth

// AuthRequiredError is returned when no token exists and the user must
// complete the OAuth login flow before making API calls.
type AuthRequiredError struct{}

// Error returns a human-readable representation of the auth required error.
func (e *AuthRequiredError) Error() string {
	return "authentication required: run the login flow to obtain a token"
}

// AuthExpiredError is returned when the refresh token has expired and
// the user must re-authenticate via the full OAuth login flow.
type AuthExpiredError struct{}

// Error returns a human-readable representation of the auth expired error.
func (e *AuthExpiredError) Error() string {
	return "authentication expired: refresh token is no longer valid, re-login required"
}

// AuthCallbackError is returned when the OAuth callback contains an
// invalid or mismatched state parameter, indicating a possible CSRF attack.
type AuthCallbackError struct {
	// Reason describes why the callback was rejected.
	Reason string
}

// Error returns a human-readable representation of the auth callback error.
func (e *AuthCallbackError) Error() string {
	return "OAuth callback error: " + e.Reason
}
