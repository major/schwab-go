package auth

const (
	redactedValue         = "redacted"
	redactedSeparator     = "..."
	tokenVisiblePrefix    = 6
	tokenVisibleSuffix    = 4
	clientIDVisiblePrefix = 4
	clientIDVisibleSuffix = 4
)

// RedactToken returns a display-safe form of an OAuth token.
// Empty tokens stay empty, short tokens are fully redacted, and longer tokens
// keep a small prefix and suffix so callers can distinguish stored values
// without exposing the secret.
func RedactToken(token string) string {
	return redactSecret(token, tokenVisiblePrefix, tokenVisibleSuffix)
}

// RedactClientID returns a display-safe form of an OAuth client ID.
// Empty client IDs stay empty, short values are fully redacted, and longer
// values keep a small prefix and suffix for troubleshooting.
func RedactClientID(clientID string) string {
	return redactSecret(clientID, clientIDVisiblePrefix, clientIDVisibleSuffix)
}

func redactSecret(secret string, visiblePrefix, visibleSuffix int) string {
	if secret == "" {
		return ""
	}
	if len(secret) <= visiblePrefix+visibleSuffix {
		return redactedValue
	}

	return secret[:visiblePrefix] + redactedSeparator + secret[len(secret)-visibleSuffix:]
}
