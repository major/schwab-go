package marketdata

import (
	"context"
	"strconv"
)

// PeriodType defines the period type for price history.
type PeriodType string

const (
	PeriodTypeDay   PeriodType = "day"
	PeriodTypeMonth PeriodType = "month"
	PeriodTypeYear  PeriodType = "year"
	PeriodTypeYTD   PeriodType = "ytd"
)

// FrequencyType defines the frequency type for price history.
type FrequencyType string

const (
	FrequencyTypeMinute  FrequencyType = "minute"
	FrequencyTypeDaily   FrequencyType = "daily"
	FrequencyTypeWeekly  FrequencyType = "weekly"
	FrequencyTypeMonthly FrequencyType = "monthly"
)

// PriceHistoryParams contains optional parameters for GetPriceHistory.
type PriceHistoryParams struct {
	PeriodType            PeriodType
	Period                int
	FrequencyType         FrequencyType
	Frequency             int
	StartDate             int64 // milliseconds since epoch
	EndDate               int64 // milliseconds since epoch
	NeedExtendedHoursData *bool
	NeedPreviousClose     *bool
}

// CandleList is the response from GET /pricehistory.
type CandleList struct {
	Candles              []Candle `json:"candles"`
	Symbol               string   `json:"symbol"`
	Empty                bool     `json:"empty"`
	PreviousClose        float64  `json:"previousClose"`
	PreviousCloseDate    int64    `json:"previousCloseDate"`
	PreviousCloseDateISO string   `json:"previousCloseDateISO8601"`
}

// Candle represents a single OHLCV candle.
type Candle struct {
	Open        float64 `json:"open"`
	High        float64 `json:"high"`
	Low         float64 `json:"low"`
	Close       float64 `json:"close"`
	Volume      int64   `json:"volume"`
	Datetime    int64   `json:"datetime"`        // milliseconds since epoch
	DatetimeISO string  `json:"datetimeISO8601"` // ISO 8601 formatted datetime
}

// GetPriceHistory retrieves price history candles for a symbol.
// symbol: the stock symbol (e.g., "AAPL")
// params: optional parameters; if nil, only symbol is sent
func (c *Client) GetPriceHistory(ctx context.Context, symbol string, params *PriceHistoryParams) (*CandleList, error) {
	req, err := c.newRequest(ctx, "/pricehistory")
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("symbol", symbol)

	if params != nil {
		if params.PeriodType != "" {
			q.Set("periodType", string(params.PeriodType))
		}
		if params.Period != 0 {
			q.Set("period", strconv.Itoa(params.Period))
		}
		if params.FrequencyType != "" {
			q.Set("frequencyType", string(params.FrequencyType))
		}
		if params.Frequency != 0 {
			q.Set("frequency", strconv.Itoa(params.Frequency))
		}
		if params.StartDate != 0 {
			q.Set("startDate", strconv.FormatInt(params.StartDate, 10))
		}
		if params.EndDate != 0 {
			q.Set("endDate", strconv.FormatInt(params.EndDate, 10))
		}
		if params.NeedExtendedHoursData != nil {
			q.Set("needExtendedHoursData", strconv.FormatBool(*params.NeedExtendedHoursData))
		}
		if params.NeedPreviousClose != nil {
			q.Set("needPreviousClose", strconv.FormatBool(*params.NeedPreviousClose))
		}
	}

	req.URL.RawQuery = q.Encode()

	var result CandleList
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
