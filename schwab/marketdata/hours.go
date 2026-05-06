package marketdata

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// MarketHoursMap is the response from GET /markets and GET /markets/{market_id}.
// Outer key: market type (e.g., "equity"), inner key: product (e.g., "EQ").
type MarketHoursMap map[string]map[string]MarketHours

// MarketHours contains trading hours for a specific market product.
type MarketHours struct {
	Date         string                    `json:"date"`
	MarketType   string                    `json:"marketType"`
	Exchange     string                    `json:"exchange"`
	Category     string                    `json:"category"`
	Product      string                    `json:"product"`
	ProductName  string                    `json:"productName"`
	IsOpen       bool                      `json:"isOpen"`
	SessionHours map[string][]SessionHours `json:"sessionHours"`
}

// SessionHours represents a trading session time window.
type SessionHours struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// GetMarketHours retrieves market hours for multiple markets on a given date.
// markets: comma-separated list of market IDs (e.g., ["equity", "option"])
// date: optional date string (YYYY-MM-DD format); if empty, current date is used
func (c *Client) GetMarketHours(ctx context.Context, markets []string, date string) (MarketHoursMap, error) {
	if len(markets) == 0 {
		return nil, fmt.Errorf("markets is required")
	}

	req, err := c.newRequest(ctx, "GET", "/markets", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("markets", strings.Join(markets, ","))
	if date != "" {
		q.Set("date", date)
	}
	req.URL.RawQuery = q.Encode()

	var result MarketHoursMap
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetMarketHoursSingle retrieves market hours for a single market on a given date.
// marketID: the market identifier (e.g., "equity")
// date: optional date string (YYYY-MM-DD format); if empty, current date is used
func (c *Client) GetMarketHoursSingle(ctx context.Context, marketID string, date string) (MarketHoursMap, error) {
	path := fmt.Sprintf("/markets/%s", url.PathEscape(marketID))
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if date != "" {
		q := req.URL.Query()
		q.Set("date", date)
		req.URL.RawQuery = q.Encode()
	}

	var result MarketHoursMap
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result, nil
}
