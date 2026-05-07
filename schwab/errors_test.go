package schwab

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		StatusCode: 404,
		Message:    "not found",
		Body:       `{"error": "resource not found"}`,
	}
	require.Equal(t, "schwab API error 404: not found", err.Error())
}

func TestAPIError_AsType_Direct(t *testing.T) {
	apiErr := &APIError{
		StatusCode: 500,
		Message:    "internal server error",
		Body:       "",
	}
	var err error = apiErr

	extracted, ok := errors.AsType[*APIError](err)
	require.True(t, ok)
	require.Equal(t, apiErr, extracted)
}

func TestAPIError_AsType_Wrapped(t *testing.T) {
	apiErr := &APIError{
		StatusCode: 403,
		Message:    "forbidden",
		Body:       "",
	}
	wrapped := fmt.Errorf("request failed: %w", apiErr)

	extracted, ok := errors.AsType[*APIError](wrapped)
	require.True(t, ok)
	require.Equal(t, apiErr, extracted)
}

func TestAPIError_AsType_NotFound(t *testing.T) {
	err := errors.New("other error")

	extracted, ok := errors.AsType[*APIError](err)
	require.False(t, ok)
	require.Nil(t, extracted)
}

func TestIsStatusCode(t *testing.T) {
	t.Parallel()

	apiErr := &APIError{StatusCode: 401, Message: "unauthorized"}
	wrapped := fmt.Errorf("get accounts: %w", apiErr)

	require.True(t, IsStatusCode(wrapped, 401))
	require.False(t, IsStatusCode(wrapped, 403))
	require.False(t, IsStatusCode(errors.New("other error"), 401))

	var nilAPIErr *APIError
	require.False(t, IsStatusCode(nilAPIErr, 401))
}

func TestStatusCode(t *testing.T) {
	t.Parallel()

	apiErr := &APIError{StatusCode: 429, Message: "too many requests"}
	wrapped := fmt.Errorf("get quotes: %w", apiErr)

	statusCode, ok := StatusCode(wrapped)
	require.True(t, ok)
	require.Equal(t, 429, statusCode)

	statusCode, ok = StatusCode(errors.New("other error"))
	require.False(t, ok)
	require.Equal(t, 0, statusCode)

	var nilAPIErr *APIError
	statusCode, ok = StatusCode(nilAPIErr)
	require.False(t, ok)
	require.Equal(t, 0, statusCode)
}

func TestIsUnauthorized(t *testing.T) {
	t.Parallel()

	apiErr := &APIError{StatusCode: 401, Message: "unauthorized"}
	wrapped := fmt.Errorf("get accounts: %w", apiErr)

	require.True(t, IsUnauthorized(wrapped))
	require.False(t, IsUnauthorized(&APIError{StatusCode: 403, Message: "forbidden"}))
}
