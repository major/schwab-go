package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schwab "github.com/major/schwab-go/schwab"
)

const relativeBaseURLError = "invalid base URL \"relative/path\": absolute URL with scheme and host required"

type responsePayload struct {
	Symbol string `json:"symbol"`
	Price  int    `json:"price"`
}

type doTestCase struct {
	name                string
	status              int
	responseBody        string
	responseContentType string
	responseLimit       int64
	extractError        func([]byte) string
	out                 any
	wantOut             any
	wantAPIError        bool
	wantStatus          int
	wantMessage         string
	wantBody            string
	wantDecodeError     bool
	wantLimitError      bool
}

func TestNewConfig(t *testing.T) {
	defaultBase, err := url.Parse("https://api.example.test/base")
	require.NoError(t, err)

	defaultClient := &http.Client{}
	customClient := &http.Client{}

	tests := []struct {
		name              string
		defaultBase       *url.URL
		defaultClient     *http.Client
		opts              []schwab.Option
		wantBase          string
		wantClient        *http.Client
		wantClientNonNil  bool
		wantToken         string
		wantOptionError   string
		wantBodyLimit     int64
		wantEmptyBase     bool
		wantDistinctClone bool
	}{
		{
			name:             "happy path uses defaults",
			defaultBase:      defaultBase,
			defaultClient:    defaultClient,
			opts:             nil,
			wantBase:         "https://api.example.test/base",
			wantClient:       defaultClient,
			wantClientNonNil: true,
			wantBodyLimit:    schwab.DefaultResponseBodyLimit,
		},
		{
			name:             "nil default base uses empty URL",
			defaultBase:      nil,
			defaultClient:    defaultClient,
			opts:             nil,
			wantClient:       defaultClient,
			wantClientNonNil: true,
			wantBodyLimit:    schwab.DefaultResponseBodyLimit,
			wantEmptyBase:    true,
		},
		{
			name:              "nil default client uses empty client",
			defaultBase:       defaultBase,
			defaultClient:     nil,
			opts:              nil,
			wantBase:          "https://api.example.test/base",
			wantClientNonNil:  true,
			wantBodyLimit:     schwab.DefaultResponseBodyLimit,
			wantDistinctClone: true,
		},
		{
			name:          "options override defaults",
			defaultBase:   defaultBase,
			defaultClient: defaultClient,
			opts: []schwab.Option{
				schwab.WithToken("tok"),
				schwab.WithBaseURL("http://example.com"),
				schwab.WithHTTPClient(customClient),
				schwab.WithResponseBodyLimit(512),
			},
			wantBase:         "http://example.com",
			wantClient:       customClient,
			wantClientNonNil: true,
			wantToken:        "tok",
			wantBodyLimit:    512,
		},
		{
			name:          "invalid base URL option preserves default and stores error",
			defaultBase:   defaultBase,
			defaultClient: defaultClient,
			opts: []schwab.Option{
				schwab.WithBaseURL("relative/path"),
			},
			wantBase:        "https://api.example.test/base",
			wantClient:      defaultClient,
			wantOptionError: relativeBaseURLError,
			wantBodyLimit:   schwab.DefaultResponseBodyLimit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig(tt.defaultBase, tt.defaultClient, tt.opts)

			require.NotNil(t, cfg.BaseURL)
			if tt.wantEmptyBase {
				assert.Empty(t, cfg.BaseURL.String())
			} else {
				assert.Equal(t, tt.wantBase, cfg.BaseURL.String())
			}

			if tt.wantClientNonNil {
				require.NotNil(t, cfg.HTTPClient)
			}
			if tt.wantClient != nil {
				assert.Same(t, tt.wantClient, cfg.HTTPClient)
			}
			if tt.wantDistinctClone {
				assert.NotSame(t, defaultClient, cfg.HTTPClient)
			}
			assert.Equal(t, tt.wantToken, cfg.Token)
			if tt.wantOptionError == "" {
				require.NoError(t, cfg.OptionError)
			} else {
				require.Error(t, cfg.OptionError)
				require.ErrorContains(t, cfg.OptionError, tt.wantOptionError)
			}
			assert.Equal(t, tt.wantBodyLimit, cfg.ResponseBodyLimit)
		})
	}
}

func TestNewRequest(t *testing.T) {
	baseURL, err := url.Parse("https://api.example.test/root")
	require.NoError(t, err)

	tests := []struct {
		name             string
		cfg              Config
		method           string
		path             string
		body             any
		wantMethod       string
		wantURL          string
		wantAuth         string
		wantContentType  string
		wantBody         string
		wantNoBody       bool
		wantNilBody      bool
		wantMarshalError bool
		wantOptionError  string
	}{
		{
			name:        "GET request with http NoBody",
			cfg:         Config{BaseURL: baseURL},
			method:      http.MethodGet,
			path:        "quotes",
			body:        http.NoBody,
			wantMethod:  http.MethodGet,
			wantURL:     "https://api.example.test/root/quotes",
			wantNoBody:  true,
			wantBody:    "",
			wantAuth:    "",
			wantNilBody: false,
		},
		{
			name:            "POST request with JSON body",
			cfg:             Config{BaseURL: baseURL},
			method:          http.MethodPost,
			path:            "orders",
			body:            map[string]string{"symbol": "AAPL"},
			wantMethod:      http.MethodPost,
			wantURL:         "https://api.example.test/root/orders",
			wantContentType: "application/json",
			wantBody:        `{"symbol":"AAPL"}`,
		},
		{
			name:       "authorization header with token",
			cfg:        Config{BaseURL: baseURL, Token: "tok"},
			method:     http.MethodGet,
			path:       "accounts",
			body:       http.NoBody,
			wantMethod: http.MethodGet,
			wantURL:    "https://api.example.test/root/accounts",
			wantAuth:   "Bearer tok",
			wantNoBody: true,
		},
		{
			name:       "no authorization header without token",
			cfg:        Config{BaseURL: baseURL},
			method:     http.MethodGet,
			path:       "accounts",
			body:       http.NoBody,
			wantMethod: http.MethodGet,
			wantURL:    "https://api.example.test/root/accounts",
			wantNoBody: true,
		},
		{
			name:             "unmarshalable request body returns error",
			cfg:              Config{BaseURL: baseURL},
			method:           http.MethodPost,
			path:             "orders",
			body:             make(chan int),
			wantMarshalError: true,
		},
		{
			name:        "nil body uses nil body reader",
			cfg:         Config{BaseURL: baseURL},
			method:      http.MethodGet,
			path:        "empty",
			body:        nil,
			wantMethod:  http.MethodGet,
			wantURL:     "https://api.example.test/root/empty",
			wantNilBody: true,
		},
		{
			name:            "option error returns before request construction",
			cfg:             Config{BaseURL: baseURL, OptionError: assert.AnError},
			method:          http.MethodGet,
			path:            "quotes",
			body:            http.NoBody,
			wantOptionError: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, reqErr := NewRequest(context.Background(), tt.cfg, tt.method, tt.path, tt.body)
			if tt.wantOptionError != "" {
				require.Error(t, reqErr)
				require.ErrorContains(t, reqErr, tt.wantOptionError)
				assert.Nil(t, req)
				return
			}
			if tt.wantMarshalError {
				require.Error(t, reqErr)
				require.ErrorContains(t, reqErr, "marshal request body:")
				assert.Nil(t, req)
				return
			}

			require.NoError(t, reqErr)
			require.NotNil(t, req)
			assert.Equal(t, tt.wantMethod, req.Method)
			assert.Equal(t, tt.wantURL, req.URL.String())
			assert.Equal(t, "application/json", req.Header.Get("Accept"))
			assert.Equal(t, tt.wantAuth, req.Header.Get("Authorization"))
			assert.Equal(t, tt.wantContentType, req.Header.Get("Content-Type"))

			if tt.wantNilBody {
				assert.Nil(t, req.Body)
				return
			}

			require.NotNil(t, req.Body)
			if tt.wantNoBody {
				assert.Equal(t, http.NoBody, req.Body)
			}
			bodyBytes, readErr := io.ReadAll(req.Body)
			require.NoError(t, readErr)
			if tt.wantBody == "" {
				assert.Empty(t, string(bodyBytes))
				return
			}
			assert.JSONEq(t, tt.wantBody, string(bodyBytes))
		})
	}
}

func TestDo(t *testing.T) {
	tests := []doTestCase{
		{
			name:                "success 200 decodes JSON body",
			status:              http.StatusOK,
			responseBody:        `{"symbol":"AAPL","price":185}`,
			responseContentType: "application/json",
			out:                 &responsePayload{},
			wantOut:             &responsePayload{Symbol: "AAPL", Price: 185},
		},
		{
			name:         "non 2xx JSON error body uses extractError message",
			status:       http.StatusBadRequest,
			responseBody: `{"message":"bad request detail"}`,
			extractError: func(body []byte) string {
				var payload struct {
					Message string `json:"message"`
				}
				if err := json.Unmarshal(body, &payload); err != nil {
					return ""
				}
				return payload.Message
			},
			out:          &responsePayload{},
			wantAPIError: true,
			wantStatus:   http.StatusBadRequest,
			wantMessage:  "bad request detail",
			wantBody:     `{"message":"bad request detail"}`,
		},
		{
			name:         "non 2xx empty body falls back to status text",
			status:       http.StatusBadGateway,
			responseBody: "",
			out:          &responsePayload{},
			wantAPIError: true,
			wantStatus:   http.StatusBadGateway,
			wantMessage:  http.StatusText(http.StatusBadGateway),
		},
		{
			name:         "non 2xx empty extracted message falls back to status text",
			status:       http.StatusInternalServerError,
			responseBody: `{"message":"ignored"}`,
			extractError: func([]byte) string {
				return ""
			},
			out:          &responsePayload{},
			wantAPIError: true,
			wantStatus:   http.StatusInternalServerError,
			wantMessage:  http.StatusText(http.StatusInternalServerError),
			wantBody:     `{"message":"ignored"}`,
		},
		{
			name:         "nil out with 200 drains body",
			status:       http.StatusOK,
			responseBody: `{"ignored":true}`,
			out:          nil,
		},
		{
			name:         "nil out with 204 drains body",
			status:       http.StatusNoContent,
			responseBody: "",
			out:          nil,
		},
		{
			name:                "success JSON body at limit decodes",
			status:              http.StatusOK,
			responseBody:        `{"symbol":"AAPL","price":185}`,
			responseContentType: "application/json",
			responseLimit:       int64(len(`{"symbol":"AAPL","price":185}`)),
			out:                 &responsePayload{},
			wantOut:             &responsePayload{Symbol: "AAPL", Price: 185},
		},
		{
			name:                "success JSON body over limit returns limit error",
			status:              http.StatusOK,
			responseBody:        `{"symbol":"AAPL","price":185}`,
			responseContentType: "application/json",
			responseLimit:       int64(len(`{"symbol":"AAPL","price":185}`) - 1),
			out:                 &responsePayload{},
			wantLimitError:      true,
		},
		{
			name:                "success JSON with trailing over limit body returns limit error",
			status:              http.StatusOK,
			responseBody:        `{"symbol":"AAPL","price":185}` + strings.Repeat(" ", 9),
			responseContentType: "application/json",
			responseLimit:       int64(len(`{"symbol":"AAPL","price":185}`)),
			out:                 &responsePayload{},
			wantLimitError:      true,
		},
		{
			name:           "non 2xx body over limit preserves API error",
			status:         http.StatusBadRequest,
			responseBody:   strings.Repeat("x", 9),
			responseLimit:  8,
			out:            &responsePayload{},
			wantAPIError:   true,
			wantStatus:     http.StatusBadRequest,
			wantLimitError: true,
		},
		{
			name:           "nil out body over limit returns limit error",
			status:         http.StatusOK,
			responseBody:   strings.Repeat("x", 9),
			responseLimit:  8,
			out:            nil,
			wantLimitError: true,
		},
		{
			name:                "malformed JSON response returns wrapped decode error",
			status:              http.StatusOK,
			responseBody:        `{broken`,
			responseContentType: "application/json",
			out:                 &responsePayload{},
			wantDecodeError:     true,
		},
		{
			name:         "status 301 returns API error",
			status:       http.StatusMovedPermanently,
			responseBody: `{"message":"moved"}`,
			extractError: func(body []byte) string {
				var payload map[string]string
				if err := json.Unmarshal(body, &payload); err != nil {
					return ""
				}
				return payload["message"]
			},
			out:          &responsePayload{},
			wantAPIError: true,
			wantStatus:   http.StatusMovedPermanently,
			wantMessage:  "moved",
			wantBody:     `{"message":"moved"}`,
		},
		{
			name:         "non 2xx caps raw error body",
			status:       http.StatusBadRequest,
			responseBody: strings.Repeat("a", maxAPIErrorBodyBytes+1),
			extractError: func([]byte) string {
				return ""
			},
			out:          &responsePayload{},
			wantAPIError: true,
			wantStatus:   http.StatusBadRequest,
			wantMessage:  http.StatusText(http.StatusBadRequest),
			wantBody:     strings.Repeat("a", maxAPIErrorBodyBytes),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/test", r.URL.Path)
				if tt.responseContentType != "" {
					w.Header().Set("Content-Type", tt.responseContentType)
				}
				w.WriteHeader(tt.status)
				_, writeErr := io.WriteString(w, tt.responseBody)
				assert.NoError(t, writeErr)
			}))
			defer ts.Close()

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/test", http.NoBody)
			require.NoError(t, err)

			cfg := Config{HTTPClient: ts.Client(), ResponseBodyLimit: tt.responseLimit}
			err = Do(cfg, req, tt.out, tt.extractError)

			assertDoResult(t, err, tt)
		})
	}
}

func TestDoValidatesSuccessContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		wantError   string
	}{
		{
			name:        "accepts JSON content type with charset",
			contentType: "application/json; charset=utf-8",
			body:        `{"symbol":"AAPL","price":185}`,
		},
		{
			name:        "rejects HTML content type without body preview",
			contentType: "text/html; charset=utf-8",
			body:        "<html><body>login required</body></html>",
			wantError:   `unexpected Content-Type "text/html; charset=utf-8" (expected application/json)`,
		},
		{
			name:      "rejects sniffed non JSON content type without body preview",
			body:      `{"symbol":"AAPL","price":185}`,
			wantError: `unexpected Content-Type "text/plain; charset=utf-8" (expected application/json)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/test", r.URL.Path)
				if tt.contentType != "" {
					w.Header().Set("Content-Type", tt.contentType)
				}
				_, writeErr := io.WriteString(w, tt.body)
				assert.NoError(t, writeErr)
			}))
			defer ts.Close()

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/test", http.NoBody)
			require.NoError(t, err)

			var got responsePayload
			err = Do(Config{HTTPClient: ts.Client()}, req, &got, nil)
			if tt.wantError != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantError)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, responsePayload{Symbol: "AAPL", Price: 185}, got)
		})
	}
}

func assertDoResult(t *testing.T, err error, tt doTestCase) {
	t.Helper()

	if tt.wantLimitError {
		require.Error(t, err)
		require.ErrorContains(t, err, "response body too large")
		require.ErrorContains(t, err, "configured limit")
		if tt.wantAPIError {
			apiErr, ok := errors.AsType[*schwab.APIError](err)
			require.True(t, ok)
			assert.Equal(t, tt.wantStatus, apiErr.StatusCode)
			assert.Contains(t, apiErr.Message, "response body too large")
			assert.Contains(t, apiErr.Message, "configured limit")
		} else {
			var maxBytesErr *http.MaxBytesError
			require.ErrorAs(t, err, &maxBytesErr)
		}
		return
	}

	if tt.wantAPIError {
		require.Error(t, err)
		apiErr, ok := errors.AsType[*schwab.APIError](err)
		require.True(t, ok)
		assert.Equal(t, tt.wantStatus, apiErr.StatusCode)
		assert.Equal(t, tt.wantMessage, apiErr.Message)
		assert.Equal(t, tt.wantBody, apiErr.Body)
		return
	}

	if tt.wantDecodeError {
		require.Error(t, err)
		assert.ErrorContains(t, err, "decode response body:")
		return
	}

	require.NoError(t, err)
	if tt.wantOut != nil {
		assert.Equal(t, tt.wantOut, tt.out)
	}
}

func TestDoReturnsHTTPClientError(t *testing.T) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://127.0.0.1:1/test", http.NoBody)
	require.NoError(t, err)

	client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, assert.AnError
	})}

	err = Do(Config{HTTPClient: client}, req, nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestReadAPIErrorBodyCapsStoredBodyAndDrainsRemainder(t *testing.T) {
	body := strings.NewReader(strings.Repeat("a", maxAPIErrorBodyBytes+1))

	bodyBytes, err := readAPIErrorBody(body)

	require.NoError(t, err)
	assert.Len(t, bodyBytes, maxAPIErrorBodyBytes)
	remainingBytes, readErr := io.ReadAll(body)
	require.NoError(t, readErr)
	assert.Empty(t, remainingBytes)
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
