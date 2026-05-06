package marketdata

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/test", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]string{"key": "value"})
	}))
	defer ts.Close()

	client := NewClient(schwab.WithHTTPClient(ts.Client()), schwab.WithBaseURL(ts.URL))
	req, err := client.newRequest(context.Background(), "/test")
	require.NoError(t, err)

	var result map[string]string
	err = client.do(req, &result)
	require.NoError(t, err)
	require.Equal(t, "value", result["key"])
}

func TestDo_ErrorWithJSONBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, map[string]string{"detail": "bad request"})
	}))
	defer ts.Close()

	client := NewClient(schwab.WithHTTPClient(ts.Client()), schwab.WithBaseURL(ts.URL))
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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := NewClient(schwab.WithHTTPClient(ts.Client()), schwab.WithBaseURL(ts.URL))
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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := NewClient(schwab.WithHTTPClient(ts.Client()), schwab.WithBaseURL(ts.URL))
	req, err := client.newRequest(context.Background(), "/test")
	require.NoError(t, err)

	err = client.do(req, nil)
	require.NoError(t, err)
}

func TestNewRequest_AuthHeader(t *testing.T) {
	client := NewClient(schwab.WithToken("test-token"))
	req, err := client.newRequest(context.Background(), "/test")
	require.NoError(t, err)
	require.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
}
