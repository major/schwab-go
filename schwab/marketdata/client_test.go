package marketdata

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schwab "github.com/major/schwab-go/schwab"
)

func TestNewClient_Defaults(t *testing.T) {
	client := NewClient()
	require.NotNil(t, client)
	require.Equal(t, defaultBaseURL, client.baseURL.String())
	require.NotNil(t, client.httpClient)
	require.Empty(t, client.token)
}

func TestNewClient_WithToken(t *testing.T) {
	client := NewClient(schwab.WithToken("test-token"))
	require.NotNil(t, client)
	require.Equal(t, "test-token", client.token)
}

func TestNewClient_WithBaseURL(t *testing.T) {
	customURL := "https://custom.example.com/api/v2"
	client := NewClient(schwab.WithBaseURL(customURL))
	require.NotNil(t, client)
	require.Equal(t, customURL, client.baseURL.String())
}

func TestNewClient_WithHTTPClient(t *testing.T) {
	customHTTPClient := &http.Client{}
	client := NewClient(schwab.WithHTTPClient(customHTTPClient))
	require.NotNil(t, client)
	require.Equal(t, customHTTPClient, client.httpClient)
}

func TestDo_Success(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]string{"key": "value"})
	})
	req, err := client.newRequest(context.Background(), "/test")
	require.NoError(t, err)

	var result map[string]string
	err = client.do(req, &result)
	require.NoError(t, err)
	require.Equal(t, "value", result["key"])
}

func TestDo_ErrorWithJSONBody(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, map[string]string{"detail": "bad request"})
	})
	req, err := client.newRequest(context.Background(), "/test")
	require.NoError(t, err)

	err = client.do(req, nil)
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	require.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
	require.Equal(t, "bad request", apiErr.Message)
}

func TestDo_ErrorWithEmptyBody(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	req, err := client.newRequest(context.Background(), "/test")
	require.NoError(t, err)

	err = client.do(req, nil)
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	require.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	require.Equal(t, "Internal Server Error", apiErr.Message)
}

func TestDo_NilOutput(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	req, err := client.newRequest(context.Background(), "/test")
	require.NoError(t, err)

	err = client.do(req, nil)
	require.NoError(t, err)
}

func TestDo_RedirectReturnsError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMovedPermanently)
		writeJSON(t, w, map[string]string{"detail": "moved"})
	})
	req, err := client.newRequest(context.Background(), "/test")
	require.NoError(t, err)

	err = client.do(req, nil)
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	require.Equal(t, http.StatusMovedPermanently, apiErr.StatusCode)
	require.Equal(t, "moved", apiErr.Message)
}

func TestNewRequest_AuthHeader(t *testing.T) {
	client := NewClient(schwab.WithToken("test-token"))
	req, err := client.newRequest(context.Background(), "/test")
	require.NoError(t, err)
	require.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
}

func TestDo_MalformedJSONBody(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("{invalid json"))
		assert.NoError(t, err)
	})
	req, err := client.newRequest(context.Background(), "/test")
	require.NoError(t, err)

	var result map[string]string
	err = client.do(req, &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode response body:")

	// Verify error chain: the underlying error should be a JSON syntax error
	var syntaxErr *json.SyntaxError
	require.ErrorAs(t, err, &syntaxErr)
}
