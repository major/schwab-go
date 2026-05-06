package trader

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	schwab "github.com/major/schwab-go/schwab"
	"github.com/major/schwab-go/schwab/internal/httpclient"
)

const (
	apiPathPrefix  = "/trader/v1"
	defaultBaseURL = "https://api.schwabapi.com"
)

// Client is an HTTP client for the Schwab Trader API.
type Client struct {
	baseURL           *url.URL
	httpClient        *http.Client
	token             string
	optionError       error
	responseBodyLimit int64
	headers           http.Header
}

// NewClient creates a new Trader API client with the given options.
func NewClient(opts ...schwab.Option) *Client {
	defaultBase, err := url.Parse(defaultBaseURL)
	if err != nil {
		defaultBase = &url.URL{Scheme: "https", Host: "api.schwabapi.com"}
	}
	cfg := httpclient.NewConfig(defaultBase, http.DefaultClient, opts)
	return &Client{
		baseURL:           httpclient.WithPathPrefix(cfg.BaseURL, apiPathPrefix),
		httpClient:        cfg.HTTPClient,
		token:             cfg.Token,
		optionError:       cfg.OptionError,
		responseBodyLimit: cfg.ResponseBodyLimit,
		headers:           cfg.Headers.Clone(),
	}
}

func (c *Client) config() httpclient.Config {
	return httpclient.Config{
		BaseURL:           c.baseURL,
		HTTPClient:        c.httpClient,
		Token:             c.token,
		OptionError:       c.optionError,
		ResponseBodyLimit: c.responseBodyLimit,
		Headers:           c.headers.Clone(),
	}
}

// newRequest builds an HTTP request with the given method, path, and optional JSON body.
func (c *Client) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	return httpclient.NewRequest(ctx, c.config(), method, path, body)
}

// do executes the request and decodes the response into out.
// If out is nil, the response body is drained and discarded.
// On non-2xx HTTP status, returns *schwab.APIError.
// Note: The Trader API returns empty bodies on errors; this is handled gracefully.
func (c *Client) do(req *http.Request, out any) error {
	return httpclient.Do(c.config(), req, out, c.extractError)
}

func (c *Client) extractError(body []byte) string {
	var errResp struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		return ""
	}
	if errResp.Message != "" {
		return errResp.Message
	}
	if errResp.Error != "" {
		return errResp.Error
	}
	return ""
}
