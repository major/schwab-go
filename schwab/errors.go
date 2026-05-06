package schwab

import "fmt"

// APIError represents an error returned by the Schwab API.
// Use errors.AsType[*APIError](err) (Go 1.26) to extract the typed error.
type APIError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("schwab API error %d: %s", e.StatusCode, e.Message)
}
