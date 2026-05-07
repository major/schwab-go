package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	callbackLoopbackHost      = "127.0.0.1"
	callbackSuccessHTML       = "<html><body><p>Login successful. You can close this tab.</p></body></html>"
	callbackReadHeaderTimeout = 10 * time.Second
	callbackShutdownTimeout   = 5 * time.Second
)

// CallbackResult contains the OAuth2 authorization code and state returned to
// the loopback callback server.
type CallbackResult struct {
	Code  string
	State string
}

// StartCallbackServer starts a loopback-only HTTPS server for receiving the
// OAuth2 authorization code callback.
//
//nolint:nonamedreturns // Named returns document the result channel, error channel, shutdown function, and setup error.
func StartCallbackServer(
	ctx context.Context,
	callbackURL string,
) (results <-chan CallbackResult, errs <-chan error, shutdown func(), err error) {
	addr, callbackPath, err := callbackListenAddress(callbackURL)
	if err != nil {
		return nil, nil, nil, err
	}

	certificate, err := generateSelfSignedCert()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate callback TLS certificate: %w", err)
	}

	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to listen for OAuth callback: %w", err)
	}

	resultCh := make(chan CallbackResult, 1)
	errCh := make(chan error, 1)
	var sendOnce sync.Once

	mux := http.NewServeMux()
	mux.Handle(callbackPath, callbackHandler(callbackPath, resultCh, errCh, &sendOnce))

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: callbackReadHeaderTimeout,
	}

	tlsListener := tls.NewListener(listener, &tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS12,
	})

	var shutdownOnce sync.Once
	shutdownFunc := func() {
		shutdownOnce.Do(func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), callbackShutdownTimeout)
			defer cancel()
			if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
				sendOnce.Do(func() {
					errCh <- fmt.Errorf("failed to shut down OAuth callback server: %w", shutdownErr)
				})
			}
		})
	}

	go func() {
		<-ctx.Done()
		shutdownFunc()
	}()

	go func() {
		serveErr := server.Serve(tlsListener)
		if serveErr != nil && serveErr != http.ErrServerClosed {
			sendOnce.Do(func() {
				errCh <- fmt.Errorf("OAuth callback server failed: %w", serveErr)
			})
		}
	}()

	return resultCh, errCh, shutdownFunc, nil
}

func callbackListenAddress(callbackURL string) (string, string, error) {
	parsedURL, err := url.Parse(callbackURL)
	if err != nil {
		return "", "", &AuthCallbackError{Msg: fmt.Sprintf("invalid callback URL: %v", err), Code: 0}
	}

	if parsedURL.Scheme != "https" {
		return "", "", &AuthCallbackError{Msg: "callback URL must use https", Code: 0}
	}

	host, port, err := net.SplitHostPort(parsedURL.Host)
	if err != nil {
		return "", "", &AuthCallbackError{
			Msg:  fmt.Sprintf("callback URL must include host and port: %v", err),
			Code: 0,
		}
	}

	if host != callbackLoopbackHost {
		return "", "", &AuthCallbackError{Msg: "callback server must bind to 127.0.0.1 only", Code: 0}
	}

	callbackPath := parsedURL.EscapedPath()
	if callbackPath == "" {
		callbackPath = "/"
	}

	return net.JoinHostPort(host, port), callbackPath, nil
}

func callbackHandler(
	callbackPath string,
	results chan<- CallbackResult,
	errs chan<- error,
	sendOnce *sync.Once,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != callbackPath {
			http.NotFound(w, r)
			return
		}

		query := r.URL.Query()
		code := query.Get("code")
		state := query.Get("state")

		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			sendOnce.Do(func() {
				errs <- &AuthCallbackError{Msg: "OAuth callback missing code", Code: http.StatusBadRequest}
			})
			return
		}

		if state == "" {
			http.Error(w, "missing state", http.StatusBadRequest)
			sendOnce.Do(func() {
				errs <- &AuthCallbackError{Msg: "OAuth callback missing state", Code: http.StatusBadRequest}
			})
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, callbackSuccessHTML)

		sendOnce.Do(func() {
			results <- CallbackResult{Code: code, State: state}
		})
	})
}
