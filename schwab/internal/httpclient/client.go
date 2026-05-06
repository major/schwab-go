// Package httpclient provides a shared HTTP client for executing API requests
// and decoding responses used by the marketdata and trader packages.
package httpclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	schwab "github.com/major/schwab-go/schwab"
)

const (
	jsonContentType      = "application/json"
	maxAPIErrorBodyBytes = 1 << 20
)

const (
	acceptHeader        = "Accept"
	authorizationHeader = "Authorization"
	contentTypeHeader   = "Content-Type"
)

// Config holds shared HTTP client settings for Schwab API packages.
type Config struct {
	BaseURL           *url.URL
	HTTPClient        *http.Client
	Token             string
	OptionError       error
	ResponseBodyLimit int64
	Headers           http.Header
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
	cfg := schwab.ClientConfig{
		BaseURL:           defaultBase,
		HTTPClient:        defaultClient,
		ResponseBodyLimit: schwab.DefaultResponseBodyLimit,
	}
	schwab.ApplyOptions(&cfg, opts)
	return Config{
		BaseURL:           cfg.BaseURL,
		HTTPClient:        cfg.HTTPClient,
		Token:             cfg.Token,
		OptionError:       cfg.OptionError,
		ResponseBodyLimit: cfg.ResponseBodyLimit,
		Headers:           cfg.Headers.Clone(),
	}
}

// NewRequest builds an HTTP request with optional JSON request body.
func NewRequest(ctx context.Context, cfg Config, method, path string, body any) (*http.Request, error) {
	if cfg.OptionError != nil {
		return nil, cfg.OptionError
	}
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
	applyHeaders(req.Header, cfg.Headers)
	req.Header.Set(acceptHeader, jsonContentType)
	if cfg.Token != "" {
		req.Header.Set(authorizationHeader, "Bearer "+cfg.Token)
	}
	if body != nil && body != http.NoBody {
		req.Header.Set(contentTypeHeader, jsonContentType)
	}
	return req, nil
}

func applyHeaders(dst, src http.Header) {
	for name, values := range src {
		if isLibraryHeader(name) {
			continue
		}
		dst.Del(name)
		for _, value := range values {
			dst.Add(name, value)
		}
	}
}

func isLibraryHeader(name string) bool {
	switch http.CanonicalHeaderKey(name) {
	case acceptHeader, authorizationHeader, contentTypeHeader:
		return true
	default:
		return false
	}
}

// Do executes the request and decodes the response into out.
func Do(cfg Config, req *http.Request, out any, extractError func([]byte) string) error {
	//nolint:gosec // Base URLs are Schwab defaults or explicit caller-provided test/API endpoints.
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	limitedBody := http.MaxBytesReader(nil, resp.Body, responseBodyLimit(cfg))

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		return apiErrorFromResponse(resp.StatusCode, limitedBody, extractError)
	}

	if out == nil {
		if _, copyErr := io.Copy(io.Discard, limitedBody); copyErr != nil {
			return responseBodyReadError("drain response body", copyErr)
		}
		return nil
	}

	bodyReader := bufio.NewReader(limitedBody)
	if contentTypeErr := validateJSONContentType(resp.Header.Get("Content-Type"), bodyReader); contentTypeErr != nil {
		return contentTypeErr
	}
	decoder := json.NewDecoder(bodyReader)
	if decodeErr := decoder.Decode(out); decodeErr != nil {
		if limitErr := responseBodyLimitError("decode response body", decodeErr); limitErr != nil {
			return limitErr
		}
		return fmt.Errorf("decode response body: %w", decodeErr)
	}
	if _, copyErr := io.Copy(io.Discard, io.MultiReader(decoder.Buffered(), bodyReader)); copyErr != nil {
		return responseBodyReadError("drain response body", copyErr)
	}
	return nil
}

func apiErrorFromResponse(statusCode int, body io.Reader, extractError func([]byte) string) *schwab.APIError {
	bodyBytes, readErr := readAPIErrorBody(body)
	apiErr := &schwab.APIError{StatusCode: statusCode}
	if readErr != nil {
		apiErr.Message = responseBodyReadMessage("read error response body", readErr)
		return apiErr
	}
	if len(bodyBytes) > 0 {
		apiErr.Body = string(bodyBytes)
	}
	if len(bodyBytes) > 0 && extractError != nil {
		apiErr.Message = extractError(bodyBytes)
	}
	if apiErr.Message == "" {
		apiErr.Message = http.StatusText(statusCode)
	}
	return apiErr
}

func readAPIErrorBody(body io.Reader) ([]byte, error) {
	bodyBytes, readErr := io.ReadAll(io.LimitReader(body, maxAPIErrorBodyBytes))
	if readErr != nil {
		return nil, readErr
	}
	if _, drainErr := io.Copy(io.Discard, body); drainErr != nil {
		return nil, drainErr
	}
	return bodyBytes, nil
}

func validateJSONContentType(contentType string, bodyReader *bufio.Reader) error {
	if _, err := bodyReader.Peek(1); err != nil {
		if err != io.EOF {
			return fmt.Errorf("read response body prefix: %w", err)
		}
		return nil
	}
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(contentType))
	if err == nil && mediaType == jsonContentType {
		return nil
	}
	return fmt.Errorf("unexpected Content-Type %q (expected %s)", contentType, jsonContentType)
}

func responseBodyLimit(cfg Config) int64 {
	if cfg.ResponseBodyLimit > 0 {
		return cfg.ResponseBodyLimit
	}
	return schwab.DefaultResponseBodyLimit
}

func responseBodyLimitError(operation string, err error) error {
	var maxBytesErr *http.MaxBytesError
	if !errors.As(err, &maxBytesErr) {
		return nil
	}
	return fmt.Errorf(
		"%s: response body too large: configured limit is %d bytes: %w",
		operation,
		maxBytesErr.Limit,
		err,
	)
}

func responseBodyReadError(operation string, err error) error {
	if limitErr := responseBodyLimitError(operation, err); limitErr != nil {
		return limitErr
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func responseBodyReadMessage(operation string, err error) string {
	return responseBodyReadError(operation, err).Error()
}
