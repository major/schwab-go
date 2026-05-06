package marketdata

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// MoverSort defines the sort order for movers.
type MoverSort string

const (
	MoverSortVolume            MoverSort = "VOLUME"
	MoverSortTrades            MoverSort = "TRADES"
	MoverSortPercentChangeUp   MoverSort = "PERCENT_CHANGE_UP"
	MoverSortPercentChangeDown MoverSort = "PERCENT_CHANGE_DOWN"
)

// MoverResponse is the response from GET /movers/{symbol_id}.
type MoverResponse struct {
	Screeners []Screener `json:"screeners"`
}

// Screener represents a single mover entry.
type Screener struct {
	Symbol           string  `json:"symbol"`
	Description      string  `json:"description"`
	Direction        string  `json:"direction"`
	Last             float64 `json:"last"`
	Change           float64 `json:"change"`
	NetPercentChange float64 `json:"netPercentChange"`
	MarketShare      float64 `json:"marketShare"`
	TotalVolume      int64   `json:"totalVolume"`
	Trades           int64   `json:"trades"`
}

// GetMovers retrieves market movers for a given symbol.
// symbolID: the symbol identifier (e.g., "$DJI", "$COMPX", "$SPX")
// sort: optional sort order; if empty, no sort param is sent
// frequency: optional frequency; if zero, no frequency param is sent
func (c *Client) GetMovers(ctx context.Context, symbolID string, sort MoverSort, frequency int) (*MoverResponse, error) {
	path := fmt.Sprintf("/movers/%s", url.PathEscape(symbolID))
	req, err := c.newRequest(ctx, path)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if sort != "" {
		q.Set("sort", string(sort))
	}
	if frequency != 0 {
		q.Set("frequency", strconv.Itoa(frequency))
	}
	req.URL.RawQuery = q.Encode()

	var result MoverResponse
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
