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
