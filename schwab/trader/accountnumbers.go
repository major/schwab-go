package trader

import (
	"context"
)

// AccountNumberHash maps a plain account number to its encrypted hash value.
// The HashValue is required for all other Trader API calls that take an account number.
type AccountNumberHash struct {
	AccountNumber string `json:"accountNumber"`
	HashValue     string `json:"hashValue"`
}

// GetAccountNumbers retrieves the mapping of account numbers to their encrypted hash values.
// This endpoint is foundational: all other Trader API calls that require an account number
// must use the HashValue returned by this method.
func (c *Client) GetAccountNumbers(ctx context.Context) ([]AccountNumberHash, error) {
	req, err := c.newRequest(ctx, "GET", "/accounts/accountNumbers", nil)
	if err != nil {
		return nil, err
	}

	var result []AccountNumberHash
	if doErr := c.do(req, &result); doErr != nil {
		return nil, doErr
	}
	return result, nil
}
