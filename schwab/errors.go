package schwab

import "fmt"

// APIError represents an error returned by the Schwab API.
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
