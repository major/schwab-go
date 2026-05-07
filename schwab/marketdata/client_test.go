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

const relativeBaseURLError = "invalid base URL \"relative/path\": absolute URL with scheme and host required"

func TestNewClient_Defaults(t *testing.T) {
	client := NewClient()
	require.NotNil(t, client)
	require.Equal(t, defaultBaseURL+apiPathPrefix, client.baseURL.String())
	require.NotNil(t, client.httpClient)
	require.Empty(t, client.token)
	require.Equal(t, schwab.DefaultResponseBodyLimit, client.responseBodyLimit)
}

func TestNewClient_WithToken(t *testing.T) {
	client := NewClient(schwab.WithToken("test-token"))
	require.NotNil(t, client)
	require.Equal(t, "test-token", client.token)
}

func TestNewClient_WithTokenProvider(t *testing.T) {
	client := NewClient(schwab.WithTokenProvider(staticTokenProvider{token: "provider-token"}))
	require.NotNil(t, client)
	require.NotNil(t, client.config().TokenProvider)

	req, err := client.newRequest(context.Background(), "/test")
	require.NoError(t, err)
	require.Equal(t, "Bearer provider-token", req.Header.Get("Authorization"))
}

func TestNewClient_WithBaseURL(t *testing.T) {
	for _, tt := range []struct {
		name    string
		rawURL  string
		wantURL string
	}{
		{
			name:    "root",
			rawURL:  "https://custom.example.com",
			wantURL: "https://custom.example.com/marketdata/v1",
		},
		{
			name:    "root with proxy path",
			rawURL:  "https://custom.example.com/proxy",
			wantURL: "https://custom.example.com/proxy/marketdata/v1",
		},
		{
			name:    "existing market data prefix",
			rawURL:  "https://custom.example.com/marketdata/v1",
			wantURL: "https://custom.example.com/marketdata/v1",
		},
		{
			name:    "existing market data prefix with trailing slash",
			rawURL:  "https://custom.example.com/proxy/marketdata/v1/",
			wantURL: "https://custom.example.com/proxy/marketdata/v1/",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(schwab.WithBaseURL(tt.rawURL))
			require.NotNil(t, client)
			require.Equal(t, tt.wantURL, client.baseURL.String())
			require.NoError(t, client.optionError)
		})
	}
}

func TestNewClient_WithInvalidBaseURL(t *testing.T) {
	client := NewClient(schwab.WithBaseURL("relative/path"))
	require.NotNil(t, client)
	require.Equal(t, defaultBaseURL+apiPathPrefix, client.baseURL.String())
	require.Error(t, client.optionError)
	require.ErrorContains(t, client.optionError, relativeBaseURLError)
}

func TestNewClient_WithHTTPClient(t *testing.T) {
	customHTTPClient := &http.Client{}
	client := NewClient(schwab.WithHTTPClient(customHTTPClient))
	require.NotNil(t, client)
	require.Equal(t, customHTTPClient, client.httpClient)
}

func TestNewClient_WithResponseBodyLimit(t *testing.T) {
	client := NewClient(schwab.WithResponseBodyLimit(1234))
	require.NotNil(t, client)
	require.Equal(t, int64(1234), client.responseBodyLimit)
	require.Equal(t, int64(1234), client.config().ResponseBodyLimit)
}

func TestNewClient_WithHeaders(t *testing.T) {
	client := NewClient(schwab.WithUserAgent("schwab-test/1.0"), schwab.WithHeader("X-Trace-Id", "trace-123"))
	require.NotNil(t, client)
	require.Equal(t, "schwab-test/1.0", client.headers.Get("User-Agent"))
	require.Equal(t, "trace-123", client.config().Headers.Get("X-Trace-Id"))
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
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
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
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
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
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	req, err := client.newRequest(context.Background(), "/test")
	require.NoError(t, err)

	err = client.do(req, nil)
	require.NoError(t, err)
}

func TestDo_RedirectReturnsError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestNewRequest_InvalidBaseURL(t *testing.T) {
	client := NewClient(schwab.WithBaseURL("relative/path"))
	req, err := client.newRequest(context.Background(), "/test")
	require.Error(t, err)
	require.ErrorContains(t, err, relativeBaseURLError)
	require.Nil(t, req)
}

func TestPublicMethod_InvalidBaseURLDoesNotCallHTTPClient(t *testing.T) {
	client := NewClient(
		schwab.WithHTTPClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			t.Fatal("HTTP client was called for invalid base URL")
			return nil, assert.AnError
		})}),
		schwab.WithBaseURL("relative/path"),
	)

	quote, err := client.GetQuote(context.Background(), "AAPL", "quote")
	require.Error(t, err)
	require.ErrorContains(t, err, relativeBaseURLError)
	require.Nil(t, quote)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type staticTokenProvider struct {
	token string
}

func (p staticTokenProvider) Token(context.Context) (string, error) {
	return p.token, nil
}

func TestDo_ErrorWithTitleFallback(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, map[string]string{"title": "validation error"})
	})
	req, err := client.newRequest(context.Background(), "/test")
	require.NoError(t, err)

	err = client.do(req, nil)
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, "validation error", apiErr.Message)
}

func TestDo_ErrorWithMalformedJSONErrorBody(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, writeErr := w.Write([]byte("{not json"))
		assert.NoError(t, writeErr)
	})
	req, err := client.newRequest(context.Background(), "/test")
	require.NoError(t, err)

	err = client.do(req, nil)
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	// extractError returns "" for malformed JSON, so falls back to StatusText.
	assert.Equal(t, "Bad Request", apiErr.Message)
}

func TestDo_ErrorWithNoMatchingFields(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		writeJSON(t, w, map[string]string{"other": "field"})
	})
	req, err := client.newRequest(context.Background(), "/test")
	require.NoError(t, err)

	err = client.do(req, nil)
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	// Neither detail nor title present, falls back to StatusText.
	assert.Equal(t, "Forbidden", apiErr.Message)
}

func TestDo_MalformedJSONBody(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
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
