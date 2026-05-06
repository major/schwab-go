package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	schwab "github.com/major/schwab-go/schwab"
)

// QuoteResponse is the response from GET /quotes, keyed by symbol.
type QuoteResponse map[string]*QuoteEntry

// QuoteEntry contains the shared envelope for quote responses.
type QuoteEntry struct {
	AssetMainType schwab.AssetType `json:"assetMainType"`
	AssetSubType  string           `json:"assetSubType"`
	QuoteType     string           `json:"quoteType"`
	Realtime      bool             `json:"realtime"`
	SSID          int64            `json:"ssid"`
	Symbol        string           `json:"symbol"`
	Quote         json.RawMessage  `json:"quote"`
	Reference     json.RawMessage  `json:"reference"`
	Regular       json.RawMessage  `json:"regular"`
	Extended      json.RawMessage  `json:"extended"`
}

// EquityQuote decodes the quote payload for an equity quote.
func (e *QuoteEntry) EquityQuote() (*EquityQuote, error) {
	if e.AssetMainType != schwab.AssetTypeEquity {
		return nil, fmt.Errorf("quote is %s, not EQUITY", e.AssetMainType)
	}
	return decodeQuote[EquityQuote](e.Quote)
}

// OptionQuote decodes the quote payload for an option quote.
func (e *QuoteEntry) OptionQuote() (*OptionQuote, error) {
	if e.AssetMainType != schwab.AssetTypeOption {
		return nil, fmt.Errorf("quote is %s, not OPTION", e.AssetMainType)
	}
	return decodeQuote[OptionQuote](e.Quote)
}

// IndexQuote decodes the quote payload for an index quote.
func (e *QuoteEntry) IndexQuote() (*IndexQuote, error) {
	if e.AssetMainType != schwab.AssetTypeIndex {
		return nil, fmt.Errorf("quote is %s, not INDEX", e.AssetMainType)
	}
	return decodeQuote[IndexQuote](e.Quote)
}

// MutualFundQuote decodes the quote payload for a mutual fund quote.
func (e *QuoteEntry) MutualFundQuote() (*MutualFundQuote, error) {
	if e.AssetMainType != schwab.AssetTypeMutualFund {
		return nil, fmt.Errorf("quote is %s, not MUTUAL_FUND", e.AssetMainType)
	}
	return decodeQuote[MutualFundQuote](e.Quote)
}

// ForexQuote decodes the quote payload for a forex quote.
func (e *QuoteEntry) ForexQuote() (*ForexQuote, error) {
	if e.AssetMainType != schwab.AssetTypeForex {
		return nil, fmt.Errorf("quote is %s, not FOREX", e.AssetMainType)
	}
	return decodeQuote[ForexQuote](e.Quote)
}

// FutureQuote decodes the quote payload for a future quote.
func (e *QuoteEntry) FutureQuote() (*FutureQuote, error) {
	if e.AssetMainType != schwab.AssetTypeFuture {
		return nil, fmt.Errorf("quote is %s, not FUTURE", e.AssetMainType)
	}
	return decodeQuote[FutureQuote](e.Quote)
}

// FutureOptionQuote decodes the quote payload for a future option quote.
func (e *QuoteEntry) FutureOptionQuote() (*FutureOptionQuote, error) {
	if e.AssetMainType != schwab.AssetTypeFutureOption {
		return nil, fmt.Errorf("quote is %s, not FUTURE_OPTION", e.AssetMainType)
	}
	return decodeQuote[FutureOptionQuote](e.Quote)
}

func decodeQuote[T any](raw json.RawMessage) (*T, error) {
	var quote T
	if err := json.Unmarshal(raw, &quote); err != nil {
		return nil, err
	}
	return &quote, nil
}

// EquityQuote contains equity-specific quote data.
type EquityQuote struct {
	AskMICId          string  `json:"askMICId"`
	AskPrice          float64 `json:"askPrice"`
	AskSize           int64   `json:"askSize"`
	AskTime           int64   `json:"askTime"`
	BidMICId          string  `json:"bidMICId"`
	BidPrice          float64 `json:"bidPrice"`
	BidSize           int64   `json:"bidSize"`
	BidTime           int64   `json:"bidTime"`
	ClosePrice        float64 `json:"closePrice"`
	HighPrice         float64 `json:"highPrice"`
	LastMICId         string  `json:"lastMICId"`
	LastPrice         float64 `json:"lastPrice"`
	LastSize          int64   `json:"lastSize"`
	LowPrice          float64 `json:"lowPrice"`
	Mark              float64 `json:"mark"`
	MarkChange        float64 `json:"markChange"`
	MarkPercentChange float64 `json:"markPercentChange"`
	NetChange         float64 `json:"netChange"`
	NetPercentChange  float64 `json:"netPercentChange"`
	OpenPrice         float64 `json:"openPrice"`
	QuoteTime         int64   `json:"quoteTime"`
	SecurityStatus    string  `json:"securityStatus"`
	TotalVolume       int64   `json:"totalVolume"`
	TradeTime         int64   `json:"tradeTime"`
	Volatility        float64 `json:"volatility"`
	Week52High        float64 `json:"52WeekHigh"`
	Week52Low         float64 `json:"52WeekLow"`
	PeRatio           float64 `json:"peRatio"`
	DivAmount         float64 `json:"divAmount"`
	DivYield          float64 `json:"divYield"`
	DivDate           string  `json:"divDate"`
	EPS               float64 `json:"eps"`
	Exchange          string  `json:"exchange"`
	ExchangeName      string  `json:"exchangeName"`
}

// OptionQuote contains option-specific quote data.
type OptionQuote struct {
	EquityQuote
	Delta                  float64 `json:"delta"`
	Gamma                  float64 `json:"gamma"`
	Theta                  float64 `json:"theta"`
	Vega                   float64 `json:"vega"`
	Rho                    float64 `json:"rho"`
	OpenInterest           int64   `json:"openInterest"`
	StrikePrice            float64 `json:"strikePrice"`
	ExpirationDate         string  `json:"expirationDate"`
	IntrinsicValue         float64 `json:"intrinsicValue"`
	ExtrinsicValue         float64 `json:"extrinsicValue"`
	TimeValue              float64 `json:"timeValue"`
	ImpliedYield           float64 `json:"impliedYield"`
	TheoreticalOptionValue float64 `json:"theoreticalOptionValue"`
	TheoreticalVolatility  float64 `json:"theoreticalVolatility"`
	UnderlyingPrice        float64 `json:"underlyingPrice"`
	VolatilityPct          float64 `json:"volatilityPct"`
}

// IndexQuote contains index-specific quote data.
type IndexQuote struct {
	ClosePrice       float64 `json:"closePrice"`
	HighPrice        float64 `json:"highPrice"`
	LastPrice        float64 `json:"lastPrice"`
	LowPrice         float64 `json:"lowPrice"`
	NetChange        float64 `json:"netChange"`
	NetPercentChange float64 `json:"netPercentChange"`
	OpenPrice        float64 `json:"openPrice"`
	SecurityStatus   string  `json:"securityStatus"`
	TotalVolume      int64   `json:"totalVolume"`
	TradeTime        int64   `json:"tradeTime"`
	Week52High       float64 `json:"52WeekHigh"`
	Week52Low        float64 `json:"52WeekLow"`
	Exchange         string  `json:"exchange"`
	ExchangeName     string  `json:"exchangeName"`
}

// MutualFundQuote contains mutual fund-specific quote data.
type MutualFundQuote struct {
	ClosePrice       float64 `json:"closePrice"`
	NAV              float64 `json:"nAV"`
	NetChange        float64 `json:"netChange"`
	NetPercentChange float64 `json:"netPercentChange"`
	SecurityStatus   string  `json:"securityStatus"`
	TotalVolume      int64   `json:"totalVolume"`
	TradeTime        int64   `json:"tradeTime"`
	Week52High       float64 `json:"52WeekHigh"`
	Week52Low        float64 `json:"52WeekLow"`
	Exchange         string  `json:"exchange"`
	ExchangeName     string  `json:"exchangeName"`
}

// ForexQuote contains forex-specific quote data.
type ForexQuote struct {
	AskPrice             float64 `json:"askPrice"`
	AskSize              int64   `json:"askSize"`
	BidPrice             float64 `json:"bidPrice"`
	BidSize              int64   `json:"bidSize"`
	ClosePrice           float64 `json:"closePrice"`
	FuturePercentChange  float64 `json:"futurePercentChange"`
	HighPrice            float64 `json:"highPrice"`
	LastPrice            float64 `json:"lastPrice"`
	LastSize             int64   `json:"lastSize"`
	LowPrice             float64 `json:"lowPrice"`
	Mark                 float64 `json:"mark"`
	MarkChange           float64 `json:"markChange"`
	MarkPercentChange    float64 `json:"markPercentChange"`
	NetChange            float64 `json:"netChange"`
	NetPercentChange     float64 `json:"netPercentChange"`
	OpenPrice            float64 `json:"openPrice"`
	QuoteTime            int64   `json:"quoteTime"`
	SecurityStatus       string  `json:"securityStatus"`
	Tick                 float64 `json:"tick"`
	TickAmount           float64 `json:"tickAmount"`
	TotalVolume          int64   `json:"totalVolume"`
	TradeTime            int64   `json:"tradeTime"`
	Week52High           float64 `json:"52WeekHigh"`
	Week52Low            float64 `json:"52WeekLow"`
	Exchange             string  `json:"exchange"`
	ExchangeName         string  `json:"exchangeName"`
}

// FutureQuote contains future-specific quote data.
type FutureQuote struct {
	AskMICId            string  `json:"askMICId"`
	AskPrice            float64 `json:"askPrice"`
	AskSize             int64   `json:"askSize"`
	AskTime             int64   `json:"askTime"`
	BidMICId            string  `json:"bidMICId"`
	BidPrice            float64 `json:"bidPrice"`
	BidSize             int64   `json:"bidSize"`
	BidTime             int64   `json:"bidTime"`
	ClosePrice          float64 `json:"closePrice"`
	FuturePercentChange float64 `json:"futurePercentChange"`
	HighPrice           float64 `json:"highPrice"`
	LastMICId           string  `json:"lastMICId"`
	LastPrice           float64 `json:"lastPrice"`
	LastSize            int64   `json:"lastSize"`
	LowPrice            float64 `json:"lowPrice"`
	Mark                float64 `json:"mark"`
	NetChange           float64 `json:"netChange"`
	NetPercentChange    float64 `json:"netPercentChange"`
	OpenInterest        int64   `json:"openInterest"`
	OpenPrice           float64 `json:"openPrice"`
	QuoteTime           int64   `json:"quoteTime"`
	QuotedInSession     bool    `json:"quotedInSession"`
	SecurityStatus      string  `json:"securityStatus"`
	SettlementPrice     float64 `json:"settlementPrice"`
	SettleTime          int64   `json:"settleTime"`
	Tick                float64 `json:"tick"`
	TickAmount          float64 `json:"tickAmount"`
	TotalVolume         int64   `json:"totalVolume"`
	TradeTime           int64   `json:"tradeTime"`
	Exchange            string  `json:"exchange"`
	ExchangeName        string  `json:"exchangeName"`
}

// FutureOptionQuote contains future option-specific quote data.
type FutureOptionQuote struct {
	AskMICId         string  `json:"askMICId"`
	AskPrice         float64 `json:"askPrice"`
	AskSize          int64   `json:"askSize"`
	BidMICId         string  `json:"bidMICId"`
	BidPrice         float64 `json:"bidPrice"`
	BidSize          int64   `json:"bidSize"`
	ClosePrice       float64 `json:"closePrice"`
	HighPrice        float64 `json:"highPrice"`
	LastMICId        string  `json:"lastMICId"`
	LastPrice        float64 `json:"lastPrice"`
	LastSize         int64   `json:"lastSize"`
	LowPrice         float64 `json:"lowPrice"`
	Mark             float64 `json:"mark"`
	MarkChange       float64 `json:"markChange"`
	NetChange        float64 `json:"netChange"`
	NetPercentChange float64 `json:"netPercentChange"`
	OpenInterest     int64   `json:"openInterest"`
	OpenPrice        float64 `json:"openPrice"`
	QuoteTime        int64   `json:"quoteTime"`
	SecurityStatus   string  `json:"securityStatus"`
	SettlementPrice  float64 `json:"settlementPrice"`
	Tick             float64 `json:"tick"`
	TickAmount       float64 `json:"tickAmount"`
	TotalVolume      int64   `json:"totalVolume"`
	TradeTime        int64   `json:"tradeTime"`
}

// QuoteError contains invalid identifiers from a partial quote response.
type QuoteError struct {
	InvalidCusips  []string `json:"invalidCusips"`
	InvalidSSIDs   []int64  `json:"invalidSSIDs"`
	InvalidSymbols []string `json:"invalidSymbols"`
}

// GetQuotes retrieves quotes for multiple symbols.
func (c *Client) GetQuotes(ctx context.Context, symbols []string, fields string, indicative bool) (*QuoteResponse, *QuoteError, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/quotes", nil)
	if err != nil {
		return nil, nil, err
	}

	q := req.URL.Query()
	q.Set("symbols", strings.Join(symbols, ","))
	if fields != "" {
		q.Set("fields", fields)
	}
	if indicative {
		q.Set("indicative", "true")
	}
	req.URL.RawQuery = q.Encode()

	var raw map[string]json.RawMessage
	if err := c.do(req, &raw); err != nil {
		return nil, nil, err
	}

	result := make(QuoteResponse)
	var quoteErr *QuoteError
	for symbol, message := range raw {
		var errProbe struct {
			InvalidSymbols []string `json:"invalidSymbols"`
		}
		if err := json.Unmarshal(message, &errProbe); err == nil && errProbe.InvalidSymbols != nil {
			var parsedErr QuoteError
			if err := json.Unmarshal(message, &parsedErr); err != nil {
				return nil, nil, fmt.Errorf("decode quote error: %w", err)
			}
			quoteErr = &parsedErr
			continue
		}

		var entry QuoteEntry
		if err := json.Unmarshal(message, &entry); err != nil {
			return nil, nil, fmt.Errorf("decode quote %s: %w", symbol, err)
		}
		result[symbol] = &entry
	}

	return &result, quoteErr, nil
}

// GetQuote retrieves a quote for a single symbol identifier.
func (c *Client) GetQuote(ctx context.Context, symbolID string, fields string) (*QuoteEntry, error) {
	path := fmt.Sprintf("/%s/quotes", url.PathEscape(symbolID))
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	if fields != "" {
		q := req.URL.Query()
		q.Set("fields", fields)
		req.URL.RawQuery = q.Encode()
	}

	var result QuoteEntry
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
