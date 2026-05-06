package marketdata

import (
	"context"
)

// ExpirationChain is the response from GET /expirationchain.
type ExpirationChain struct {
	ExpirationList []Expiration `json:"expirationList"`
}

// Expiration represents a single option expiration date.
type Expiration struct {
	ExpirationDate   string `json:"expirationDate"`
	DaysToExpiration int    `json:"daysToExpiration"`
	ExpirationType   string `json:"expirationType"`
	SettlementType   string `json:"settlementType"`
	OptionRoots      string `json:"optionRoots"`
	Standard         bool   `json:"standard"`
}

// GetExpirationChain retrieves the option expiration chain for a given symbol.
// symbol: the underlying symbol (e.g., "AAPL")
func (c *Client) GetExpirationChain(ctx context.Context, symbol string) (*ExpirationChain, error) {
	req, err := c.newRequest(ctx, "GET", "/expirationchain", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("symbol", symbol)
	req.URL.RawQuery = q.Encode()

	var result ExpirationChain
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
