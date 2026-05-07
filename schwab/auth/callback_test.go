package auth

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const callbackTestTimeout = 5 * time.Second

func TestCallbackServer(t *testing.T) {
	t.Parallel()

	t.Run("server starts and listens", func(t *testing.T) {
		t.Parallel()

		callbackURL := newCallbackTestURL(t)
		_, _, shutdown, err := StartCallbackServer(context.Background(), callbackURL)
		require.NoError(t, err)
		t.Cleanup(shutdown)

		resp := getCallback(t, callbackURL, url.Values{"code": {"auth-code"}, "state": {"csrf-state"}})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, readResponseBody(t, resp), "Login successful. You can close this tab.")
	})

	t.Run("callback with code and state sends result", func(t *testing.T) {
		t.Parallel()

		callbackURL := newCallbackTestURL(t)
		results, _, shutdown, err := StartCallbackServer(context.Background(), callbackURL)
		require.NoError(t, err)
		t.Cleanup(shutdown)

		resp := getCallback(t, callbackURL, url.Values{"code": {"auth-code"}, "state": {"csrf-state"}})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, readResponseBody(t, resp), "Login successful. You can close this tab.")

		select {
		case got := <-results:
			assert.Equal(t, CallbackResult{Code: "auth-code", State: "csrf-state"}, got)
		case <-time.After(callbackTestTimeout):
			t.Fatal("StartCallbackServer() did not send callback result")
		}
	})

	t.Run("missing code sends error", func(t *testing.T) {
		t.Parallel()

		callbackURL := newCallbackTestURL(t)
		_, errs, shutdown, err := StartCallbackServer(context.Background(), callbackURL)
		require.NoError(t, err)
		t.Cleanup(shutdown)

		resp := getCallback(t, callbackURL, url.Values{"state": {"csrf-state"}})
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		_ = readResponseBody(t, resp)

		var callbackErr *AuthCallbackError
		select {
		case gotErr := <-errs:
			require.ErrorAs(t, gotErr, &callbackErr)
			assert.Contains(t, callbackErr.Msg, "missing code")
			assert.Equal(t, http.StatusBadRequest, callbackErr.Code)
		case <-time.After(callbackTestTimeout):
			t.Fatal("StartCallbackServer() did not send missing code error")
		}
	})

	t.Run("missing state sends error", func(t *testing.T) {
		t.Parallel()

		callbackURL := newCallbackTestURL(t)
		_, errs, shutdown, err := StartCallbackServer(context.Background(), callbackURL)
		require.NoError(t, err)
		t.Cleanup(shutdown)

		resp := getCallback(t, callbackURL, url.Values{"code": {"auth-code"}})
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		_ = readResponseBody(t, resp)

		var callbackErr *AuthCallbackError
		select {
		case gotErr := <-errs:
			require.ErrorAs(t, gotErr, &callbackErr)
			assert.Contains(t, callbackErr.Msg, "missing state")
			assert.Equal(t, http.StatusBadRequest, callbackErr.Code)
		case <-time.After(callbackTestTimeout):
			t.Fatal("StartCallbackServer() did not send missing state error")
		}
	})

	t.Run("OAuth error query param sends error", func(t *testing.T) {
		t.Parallel()

		callbackURL := newCallbackTestURL(t)
		_, errs, shutdown, err := StartCallbackServer(context.Background(), callbackURL)
		require.NoError(t, err)
		t.Cleanup(shutdown)

		resp := getCallback(t, callbackURL, url.Values{
			"error":             {"access_denied"},
			"error_description": {"user declined"},
		})
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		_ = readResponseBody(t, resp)

		var callbackErr *AuthCallbackError
		select {
		case gotErr := <-errs:
			require.ErrorAs(t, gotErr, &callbackErr)
			assert.Contains(t, callbackErr.Msg, "access_denied")
			assert.Contains(t, callbackErr.Msg, "user declined")
			assert.Equal(t, http.StatusBadRequest, callbackErr.Code)
		case <-time.After(callbackTestTimeout):
			t.Fatal("StartCallbackServer() did not send OAuth error")
		}
	})

	t.Run("shutdown via context stops accepting connections", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		callbackURL := newCallbackTestURL(t)
		_, _, shutdown, err := StartCallbackServer(ctx, callbackURL)
		require.NoError(t, err)
		t.Cleanup(shutdown)

		resp := getCallback(t, callbackURL, url.Values{"code": {"auth-code"}, "state": {"csrf-state"}})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_ = readResponseBody(t, resp)

		cancel()
		require.Eventually(t, func() bool {
			_, getErr := callbackTestClient().Get(callbackURL)
			return getErr != nil
		}, callbackTestTimeout, 25*time.Millisecond)
	})
}

func newCallbackTestURL(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()
	require.NoError(t, listener.Close())

	return "https://" + addr + "/callback"
}

func getCallback(t *testing.T, callbackURL string, query url.Values) *http.Response {
	t.Helper()

	parsedURL, err := url.Parse(callbackURL)
	require.NoError(t, err)
	parsedURL.RawQuery = query.Encode()

	resp, err := callbackTestClient().Get(parsedURL.String())
	require.NoError(t, err)
	return resp
}

func callbackTestClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			// Tests use an in-memory self-signed loopback certificate.
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: callbackTestTimeout,
	}
}

func readResponseBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(body)
}
