package marketdata

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	schwab "github.com/major/schwab-go/schwab"
)

const defaultBaseURL = "https://api.schwabapi.com/marketdata/v1"

// Client is an HTTP client for the Schwab Market Data API.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	token      string
}

// NewClient creates a new Market Data API client with the given options.
func NewClient(opts ...schwab.Option) *Client {
	cfg := schwab.ClientConfig{
		BaseURL: &url.URL{
			Scheme: "https",
			Host:   "api.schwabapi.com",
			Path:   "/marketdata/v1",
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

// newRequest builds a GET request with the given path.
func (c *Client) newRequest(ctx context.Context, path string) (*http.Request, error) {
	u := c.baseURL.JoinPath(path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return req, nil
}

// do executes the request and decodes the response into out.
// If out is nil, the response body is drained and discarded.
// On HTTP errors (status >= 400), returns *schwab.APIError.
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
			// Try to decode structured error response
			var errResp struct {
				Detail string `json:"detail"`
				Title  string `json:"title"`
			}
			if jsonErr := json.Unmarshal(bodyBytes, &errResp); jsonErr == nil && errResp.Detail != "" {
				apiErr.Message = errResp.Detail
			} else if jsonErr == nil && errResp.Title != "" {
				apiErr.Message = errResp.Title
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
