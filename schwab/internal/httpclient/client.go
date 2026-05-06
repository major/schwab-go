// Package httpclient provides a shared HTTP client for executing API requests
// and decoding responses used by the marketdata and trader packages.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	schwab "github.com/major/schwab-go/schwab"
)

// Config holds shared HTTP client settings for Schwab API packages.
type Config struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
	Token      string
}

// NewConfig applies shared Schwab client options to default HTTP settings.
// It defensively nil-guards defaultBase and defaultClient to prevent panics in NewRequest or Do.
func NewConfig(defaultBase *url.URL, defaultClient *http.Client, opts []schwab.Option) Config {
	// Provide safe defaults if nil values are passed.
	if defaultBase == nil {
		defaultBase = &url.URL{}
	}
	if defaultClient == nil {
		defaultClient = &http.Client{}
	}
	cfg := schwab.ClientConfig{BaseURL: defaultBase, HTTPClient: defaultClient}
	schwab.ApplyOptions(&cfg, opts)
	return Config{BaseURL: cfg.BaseURL, HTTPClient: cfg.HTTPClient, Token: cfg.Token}
}

// NewRequest builds an HTTP request with optional JSON request body.
func NewRequest(ctx context.Context, cfg Config, method, path string, body any) (*http.Request, error) {
	u := cfg.BaseURL.JoinPath(path)
	var bodyReader io.Reader
	if body != nil && body != http.NoBody {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	} else if body == http.NoBody {
		bodyReader = http.NoBody
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.Token)
	}
	if body != nil && body != http.NoBody {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// Do executes the request and decodes the response into out.
func Do(cfg Config, req *http.Request, out any, extractError func([]byte) string) error {
	//nolint:gosec // Base URLs are Schwab defaults or explicit caller-provided test/API endpoints.
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		apiErr := &schwab.APIError{StatusCode: resp.StatusCode}
		if readErr == nil && len(bodyBytes) > 0 {
			apiErr.Message = extractError(bodyBytes)
			// Body intentionally omitted to avoid leaking sensitive API response data.
		}
		if apiErr.Message == "" {
			apiErr.Message = http.StatusText(resp.StatusCode)
		}
		return apiErr
	}

	if out == nil {
		if _, copyErr := io.Copy(io.Discard, resp.Body); copyErr != nil {
			return copyErr
		}
		return nil
	}

	if decodeErr := json.NewDecoder(resp.Body).Decode(out); decodeErr != nil {
		return fmt.Errorf("decode response body: %w", decodeErr)
	}
	return nil
}
