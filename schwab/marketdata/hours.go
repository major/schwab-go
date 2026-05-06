package marketdata

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const marketHoursDateLayout = "2006-01-02"

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
// Markets must use supported values (e.g., "equity", "option").
// Date is optional and must use YYYY-MM-DD format from today through one year out.
func (c *Client) GetMarketHours(ctx context.Context, markets []string, date string) (MarketHoursMap, error) {
	return c.getMarketHours(ctx, marketIDsFromStrings(markets), date)
}

// GetMarketHoursTyped retrieves market hours for multiple typed markets on a given date.
// Markets must use supported MarketID values (e.g., MarketIDEquity, MarketIDOption).
// Date is optional and must use YYYY-MM-DD format from today through one year out.
func (c *Client) GetMarketHoursTyped(ctx context.Context, markets []MarketID, date string) (MarketHoursMap, error) {
	return c.getMarketHours(ctx, markets, date)
}

func (c *Client) getMarketHours(ctx context.Context, markets []MarketID, date string) (MarketHoursMap, error) {
	if err := validateMarketIDs(markets); err != nil {
		return nil, err
	}
	if err := validateMarketHoursDate(date); err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, "/markets")
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("markets", joinMarketIDs(markets))
	if date != "" {
		q.Set("date", date)
	}
	req.URL.RawQuery = q.Encode()

	var result MarketHoursMap
	if doErr := c.do(req, &result); doErr != nil {
		return nil, doErr
	}
	return result, nil
}

// GetMarketHoursSingle retrieves market hours for a single market on a given date.
// marketID: the market identifier (e.g., "equity")
// date: optional date string (YYYY-MM-DD format); if empty, current date is used.
func (c *Client) GetMarketHoursSingle(ctx context.Context, marketID, date string) (MarketHoursMap, error) {
	return c.getMarketHoursSingle(ctx, MarketID(marketID), date)
}

// GetMarketHoursSingleTyped retrieves market hours for a single typed market on a given date.
// marketID: the market identifier (e.g., MarketIDEquity)
// date: optional date string (YYYY-MM-DD format); if empty, current date is used.
func (c *Client) GetMarketHoursSingleTyped(
	ctx context.Context,
	marketID MarketID,
	date string,
) (MarketHoursMap, error) {
	return c.getMarketHoursSingle(ctx, marketID, date)
}

func (c *Client) getMarketHoursSingle(ctx context.Context, marketID MarketID, date string) (MarketHoursMap, error) {
	if err := validateMarketID(marketID); err != nil {
		return nil, err
	}
	if err := validateMarketHoursDate(date); err != nil {
		return nil, err
	}

	path := "/markets/" + url.PathEscape(string(marketID))
	req, err := c.newRequest(ctx, path)
	if err != nil {
		return nil, err
	}

	if date != "" {
		q := req.URL.Query()
		q.Set("date", date)
		req.URL.RawQuery = q.Encode()
	}

	var result MarketHoursMap
	if doErr := c.do(req, &result); doErr != nil {
		return nil, doErr
	}
	return result, nil
}

func validateMarketIDs(markets []MarketID) error {
	if len(markets) == 0 {
		return errors.New("markets is required")
	}

	seen := make(map[MarketID]struct{}, len(markets))
	for _, marketID := range markets {
		if err := validateMarketID(marketID); err != nil {
			return err
		}
		if _, ok := seen[marketID]; ok {
			return fmt.Errorf("duplicate market %q", marketID)
		}
		seen[marketID] = struct{}{}
	}
	return nil
}

func validateMarketID(marketID MarketID) error {
	switch marketID {
	case MarketIDEquity, MarketIDOption, MarketIDBond, MarketIDFuture, MarketIDForex:
		return nil
	default:
		return fmt.Errorf("invalid market %q: expected one of equity, option, bond, future, forex", marketID)
	}
}

func validateMarketHoursDate(date string) error {
	if date == "" {
		return nil
	}
	return validateMarketHoursDateWithNow(date, time.Now())
}

func validateMarketHoursDateWithNow(date string, now time.Time) error {
	parsed, err := time.Parse(marketHoursDateLayout, date)
	if err != nil {
		return fmt.Errorf("invalid date %q: expected YYYY-MM-DD", date)
	}

	now = now.In(now.Location())
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if parsed.Before(today) {
		return fmt.Errorf("date %q is before today", date)
	}
	if parsed.After(today.AddDate(1, 0, 0)) {
		return fmt.Errorf("date %q is more than one year from today", date)
	}
	return nil
}

func marketIDsFromStrings(markets []string) []MarketID {
	marketIDs := make([]MarketID, len(markets))
	for i, market := range markets {
		marketIDs[i] = MarketID(market)
	}
	return marketIDs
}

func joinMarketIDs(markets []MarketID) string {
	marketValues := make([]string, len(markets))
	for i, marketID := range markets {
		marketValues[i] = string(marketID)
	}
	return strings.Join(marketValues, ",")
}
