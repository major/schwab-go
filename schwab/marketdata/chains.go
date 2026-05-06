package marketdata

import (
	"context"
	"errors"
	"strconv"
)

// OptionChainParams contains parameters for GetOptionChain.
type OptionChainParams struct {
	Symbol                 string
	ContractType           string
	StrikeCount            int
	IncludeUnderlyingQuote bool
	Strategy               string
	Interval               float64
	Strike                 float64
	Range                  string
	FromDate               string
	ToDate                 string
	Volatility             float64
	UnderlyingPrice        float64
	InterestRate           float64
	DaysToExpiration       int
	ExpMonth               string
	OptionType             string
	Entitlement            string
}

// OptionChain is the response from GET /chains.
type OptionChain struct {
	Symbol           string                                 `json:"symbol"`
	Status           string                                 `json:"status"`
	Strategy         string                                 `json:"strategy"`
	Interval         float64                                `json:"interval"`
	IsDelayed        bool                                   `json:"isDelayed"`
	IsIndex          bool                                   `json:"isIndex"`
	DaysToExpiration float64                                `json:"daysToExpiration"`
	InterestRate     float64                                `json:"interestRate"`
	UnderlyingPrice  float64                                `json:"underlyingPrice"`
	Volatility       float64                                `json:"volatility"`
	CallExpDateMap   map[string]map[string][]OptionContract `json:"callExpDateMap"`
	PutExpDateMap    map[string]map[string][]OptionContract `json:"putExpDateMap"`
	Underlying       *Underlying                            `json:"underlying"`
}

// OptionContract represents a single option contract within an option chain.
type OptionContract struct {
	PutCall                string              `json:"putCall"`
	Symbol                 string              `json:"symbol"`
	Description            string              `json:"description"`
	ExchangeName           string              `json:"exchangeName"`
	BidPrice               float64             `json:"bidPrice"`
	AskPrice               float64             `json:"askPrice"`
	LastPrice              float64             `json:"lastPrice"`
	MarkPrice              float64             `json:"markPrice"`
	BidSize                int                 `json:"bidSize"`
	AskSize                int                 `json:"askSize"`
	LastSize               int                 `json:"lastSize"`
	HighPrice              float64             `json:"highPrice"`
	LowPrice               float64             `json:"lowPrice"`
	OpenPrice              float64             `json:"openPrice"`
	ClosePrice             float64             `json:"closePrice"`
	TotalVolume            int64               `json:"totalVolume"`
	TradeDate              string              `json:"tradeDate"`
	TradeTimeInLong        int64               `json:"tradeTimeInLong"`
	QuoteTimeInLong        int64               `json:"quoteTimeInLong"`
	NetChange              float64             `json:"netChange"`
	PercentChange          float64             `json:"percentChange"`
	MarkChange             float64             `json:"markChange"`
	MarkPercentChange      float64             `json:"markPercentChange"`
	Volatility             float64             `json:"volatility"`
	Delta                  float64             `json:"delta"`
	Gamma                  float64             `json:"gamma"`
	Theta                  float64             `json:"theta"`
	Vega                   float64             `json:"vega"`
	Rho                    float64             `json:"rho"`
	OpenInterest           int64               `json:"openInterest"`
	TimeValue              float64             `json:"timeValue"`
	TheoreticalOptionValue float64             `json:"theoreticalOptionValue"`
	TheoreticalVolatility  float64             `json:"theoreticalVolatility"`
	StrikePrice            float64             `json:"strikePrice"`
	ExpirationDate         string              `json:"expirationDate"`
	DaysToExpiration       int                 `json:"daysToExpiration"`
	ExpirationType         string              `json:"expirationType"`
	LastTradingDay         int64               `json:"lastTradingDay"`
	Multiplier             float64             `json:"multiplier"`
	SettlementType         string              `json:"settlementType"`
	DeliverableNote        string              `json:"deliverableNote"`
	InTheMoney             bool                `json:"isInTheMoney"`
	NonStandard            bool                `json:"isNonStandard"`
	Mini                   bool                `json:"isMini"`
	PennyPilot             bool                `json:"isPennyPilot"`
	IndexOption            bool                `json:"isIndexOption"`
	OptionRoot             string              `json:"optionRoot"`
	IntrinsicValue         float64             `json:"intrinsicValue"`
	ExtrinsicValue         float64             `json:"extrinsicValue"`
	ImpliedYield           float64             `json:"impliedYield"`
	OptionDeliverablesList []OptionDeliverable `json:"optionDeliverablesList,omitempty"`
}

// OptionDeliverable represents a single deliverable in an option contract.
type OptionDeliverable struct {
	AssetType        string `json:"assetType"`
	CurrencyType     string `json:"currencyType"`
	DeliverableUnits string `json:"deliverableUnits"`
	Symbol           string `json:"symbol"`
}

// Underlying represents the underlying quote returned with an option chain.
type Underlying struct {
	Ask               float64 `json:"ask"`
	AskSize           int     `json:"askSize"`
	Bid               float64 `json:"bid"`
	BidSize           int     `json:"bidSize"`
	Change            float64 `json:"change"`
	Close             float64 `json:"close"`
	Delayed           bool    `json:"delayed"`
	Description       string  `json:"description"`
	ExchangeName      string  `json:"exchangeName"`
	FiftyTwoWeekHigh  float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow   float64 `json:"fiftyTwoWeekLow"`
	HighPrice         float64 `json:"highPrice"`
	Last              float64 `json:"last"`
	LowPrice          float64 `json:"lowPrice"`
	Mark              float64 `json:"mark"`
	MarkChange        float64 `json:"markChange"`
	MarkPercentChange float64 `json:"markPercentChange"`
	OpenPrice         float64 `json:"openPrice"`
	PercentChange     float64 `json:"percentChange"`
	QuoteTime         int64   `json:"quoteTime"`
	Symbol            string  `json:"symbol"`
	TotalVolume       int64   `json:"totalVolume"`
	TradeTime         int64   `json:"tradeTime"`
}

// GetOptionChain retrieves an option chain for a symbol.
func (c *Client) GetOptionChain(ctx context.Context, params *OptionChainParams) (*OptionChain, error) {
	if params == nil || params.Symbol == "" {
		return nil, errors.New("symbol is required")
	}

	req, err := c.newRequest(ctx, "/chains")
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("symbol", params.Symbol)
	setOptionalString(q, "contractType", params.ContractType)
	setOptionalInt(q, "strikeCount", params.StrikeCount)
	if params.IncludeUnderlyingQuote {
		q.Set("includeUnderlyingQuote", strconv.FormatBool(params.IncludeUnderlyingQuote))
	}
	setOptionalString(q, "strategy", params.Strategy)
	setOptionalFloat64(q, "interval", params.Interval)
	setOptionalFloat64(q, "strike", params.Strike)
	setOptionalString(q, "range", params.Range)
	setOptionalString(q, "fromDate", params.FromDate)
	setOptionalString(q, "toDate", params.ToDate)
	setOptionalFloat64(q, "volatility", params.Volatility)
	setOptionalFloat64(q, "underlyingPrice", params.UnderlyingPrice)
	setOptionalFloat64(q, "interestRate", params.InterestRate)
	setOptionalInt(q, "daysToExpiration", params.DaysToExpiration)
	setOptionalString(q, "expMonth", params.ExpMonth)
	setOptionalString(q, "optionType", params.OptionType)
	setOptionalString(q, "entitlement", params.Entitlement)
	req.URL.RawQuery = q.Encode()

	var result OptionChain
	if doErr := c.do(req, &result); doErr != nil {
		return nil, doErr
	}
	return &result, nil
}
