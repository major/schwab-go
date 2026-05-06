package trader

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

const defaultBaseURL = "https://api.schwabapi.com/trader/v1"

// Client is an HTTP client for the Schwab Trader API.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	token      string
}

// NewClient creates a new Trader API client with the given options.
func NewClient(opts ...schwab.Option) *Client {
	cfg := schwab.ClientConfig{
		BaseURL: &url.URL{
			Scheme: "https",
			Host:   "api.schwabapi.com",
			Path:   "/trader/v1",
		},
		HTTPClient: http.DefaultClient,
	}
	schwab.ApplyOptions(&cfg, opts)
	return &Client{
		baseURL:    cfg.BaseURL,
		httpClient: cfg.HTTPClient,
		token:      cfg.Token,
	}
}

// newRequest builds an HTTP request with the given method, path, and optional JSON body.
func (c *Client) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	u := c.baseURL.JoinPath(path)
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// do executes the request and decodes the response into out.
// If out is nil, the response body is drained and discarded.
// On HTTP errors (status >= 400), returns *schwab.APIError.
// Note: The Trader API returns empty bodies on errors; this is handled gracefully.
func (c *Client) do(req *http.Request, out any) error {
	// The request URL comes from the client's configured base URL. WithBaseURL is a documented escape hatch for tests and alternate deployments.
	//nolint:gosec // Caller-controlled base URLs are intentional for this library API.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		apiErr := &schwab.APIError{StatusCode: resp.StatusCode}
		if readErr == nil && len(bodyBytes) > 0 {
			// Try to decode structured error response (if any)
			var errResp struct {
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			if jsonErr := json.Unmarshal(bodyBytes, &errResp); jsonErr == nil && errResp.Message != "" {
				apiErr.Message = errResp.Message
			} else if jsonErr == nil && errResp.Error != "" {
				apiErr.Message = errResp.Error
			}
			apiErr.Body = string(bodyBytes)
		}
		if apiErr.Message == "" {
			apiErr.Message = http.StatusText(resp.StatusCode)
		}
		return apiErr
	}

	if out == nil {
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			return err
		}
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
