package marketdata

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	schwab "github.com/major/schwab-go/schwab"
	"github.com/major/schwab-go/schwab/internal/httpclient"
)

const defaultBaseURL = "https://api.schwabapi.com/marketdata/v1"

// Client is an HTTP client for the Schwab Market Data API.
type Client struct {
	baseURL           *url.URL
	httpClient        *http.Client
	token             string
	optionError       error
	responseBodyLimit int64
}

// NewClient creates a new Market Data API client with the given options.
func NewClient(opts ...schwab.Option) *Client {
	defaultBase, err := url.Parse(defaultBaseURL)
	if err != nil {
		defaultBase = &url.URL{Scheme: "https", Host: "api.schwabapi.com", Path: "/marketdata/v1"}
	}
	cfg := httpclient.NewConfig(defaultBase, http.DefaultClient, opts)
	return &Client{
		baseURL:           cfg.BaseURL,
		httpClient:        cfg.HTTPClient,
		token:             cfg.Token,
		optionError:       cfg.OptionError,
		responseBodyLimit: cfg.ResponseBodyLimit,
	}
}

func (c *Client) config() httpclient.Config {
	return httpclient.Config{
		BaseURL:           c.baseURL,
		HTTPClient:        c.httpClient,
		Token:             c.token,
		OptionError:       c.optionError,
		ResponseBodyLimit: c.responseBodyLimit,
	}
}

// newRequest builds a GET request with the given path.
func (c *Client) newRequest(ctx context.Context, path string) (*http.Request, error) {
	return httpclient.NewRequest(ctx, c.config(), http.MethodGet, path, http.NoBody)
}

// do executes the request and decodes the response into out.
// If out is nil, the response body is drained and discarded.
// On non-2xx HTTP status, returns *schwab.APIError.
func (c *Client) do(req *http.Request, out any) error {
	return httpclient.Do(c.config(), req, out, c.extractError)
}

func (c *Client) extractError(body []byte) string {
	var errResp struct {
		Detail string `json:"detail"`
		Title  string `json:"title"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		return ""
	}
	if errResp.Detail != "" {
		return errResp.Detail
	}
	if errResp.Title != "" {
		return errResp.Title
	}
	return ""
}
