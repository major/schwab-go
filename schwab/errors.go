package schwab

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError represents an error returned by the Schwab API.
// Body contains up to 1 MiB of the raw error response body.
// Use [errors.AsType] with *APIError (Go 1.26) to extract the typed error.
type APIError struct {
	StatusCode int
	Message    string
	Body       string
}

// Error returns a human-readable representation of the API error
// including the HTTP status code and message.
func (e *APIError) Error() string {
	return fmt.Sprintf("schwab API error %d: %s", e.StatusCode, e.Message)
}

// IsStatusCode reports whether err contains an *APIError with statusCode.
func IsStatusCode(err error, statusCode int) bool {
	apiErr, ok := errors.AsType[*APIError](err)
	return ok && apiErr != nil && apiErr.StatusCode == statusCode
}

// IsUnauthorized reports whether err contains an *APIError with HTTP 401.
func IsUnauthorized(err error) bool {
	return IsStatusCode(err, http.StatusUnauthorized)
}
