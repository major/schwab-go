package trader

import (
	"context"
	"net/url"

	schwab "github.com/major/schwab-go/schwab"
)

// Account wraps a securities account.
type Account struct {
	SecuritiesAccount SecuritiesAccount `json:"securitiesAccount"`
}

// SecuritiesAccount represents a brokerage account (CASH or MARGIN).
type SecuritiesAccount struct {
	Type                    string     `json:"type"` // "CASH" or "MARGIN"
	AccountNumber           string     `json:"accountNumber"`
	RoundTrips              int        `json:"roundTrips"`
	IsDayTrader             bool       `json:"isDayTrader"`
	IsClosingOnlyRestricted bool       `json:"isClosingOnlyRestricted"`
	PfcbFlag                bool       `json:"pfcbFlag"`
	Positions               []Position `json:"positions"`
	InitialBalances         Balance    `json:"initialBalances"`
	CurrentBalances         Balance    `json:"currentBalances"`
	ProjectedBalances       Balance    `json:"projectedBalances"`
}

// Position represents a held position in an account.
type Position struct {
	ShortQuantity                  float64           `json:"shortQuantity"`
	AveragePrice                   float64           `json:"averagePrice"`
	CurrentDayProfitLoss           float64           `json:"currentDayProfitLoss"`
	CurrentDayProfitLossPercentage float64           `json:"currentDayProfitLossPercentage"`
	LongQuantity                   float64           `json:"longQuantity"`
	SettledLongQuantity            float64           `json:"settledLongQuantity"`
	SettledShortQuantity           float64           `json:"settledShortQuantity"`
	AgedQuantity                   float64           `json:"agedQuantity"`
	Instrument                     AccountInstrument `json:"instrument"`
	MarketValue                    float64           `json:"marketValue"`
	MaintenanceRequirement         float64           `json:"maintenanceRequirement"`
	AverageLongPrice               float64           `json:"averageLongPrice"`
	AverageShortPrice              float64           `json:"averageShortPrice"`
	TaxLotAverageLongPrice         float64           `json:"taxLotAverageLongPrice"`
	TaxLotAverageShortPrice        float64           `json:"taxLotAverageShortPrice"`
	LongOpenProfitLoss             float64           `json:"longOpenProfitLoss"`
	ShortOpenProfitLoss            float64           `json:"shortOpenProfitLoss"`
	PreviousSessionLongQuantity    float64           `json:"previousSessionLongQuantity"`
	PreviousSessionShortQuantity   float64           `json:"previousSessionShortQuantity"`
	CurrentDayCost                 float64           `json:"currentDayCost"`
}

// AccountInstrument represents a financial instrument held in an account.
// For options, the PutCall, OptionMultiplier, OptionDeliverables, and
// UnderlyingSymbol fields are populated. For fixed income instruments,
// MaturityDate, Factor, and VariableRate are populated.
type AccountInstrument struct {
	AssetType          schwab.AssetType           `json:"assetType"`
	Cusip              string                     `json:"cusip"`
	Symbol             string                     `json:"symbol"`
	Description        string                     `json:"description"`
	InstrumentId       int64                      `json:"instrumentId"`
	NetChange          float64                    `json:"netChange"`
	Type               string                     `json:"type,omitempty"`
	PutCall            string                     `json:"putCall,omitempty"`
	OptionMultiplier   int32                      `json:"optionMultiplier,omitempty"`
	OptionDeliverables []AccountOptionDeliverable `json:"optionDeliverables,omitempty"`
	UnderlyingSymbol   string                     `json:"underlyingSymbol,omitempty"`
	MaturityDate       string                     `json:"maturityDate,omitempty"`
	Factor             float64                    `json:"factor,omitempty"`
	VariableRate       float64                    `json:"variableRate,omitempty"`
}

// AccountOptionDeliverable represents a deliverable for an account option instrument.
type AccountOptionDeliverable struct {
	Symbol           string           `json:"symbol"`
	DeliverableUnits float64          `json:"deliverableUnits"`
	APICurrencyType  string           `json:"apiCurrencyType"`
	AssetType        schwab.AssetType `json:"assetType"`
}

// Balance contains financial balance information for an account.
type Balance struct {
	AccruedInterest                  float64 `json:"accruedInterest"`
	AvailableFundsNonMarginableTrade float64 `json:"availableFundsNonMarginableTrade"`
	BondValue                        float64 `json:"bondValue"`
	BuyingPower                      float64 `json:"buyingPower"`
	CashBalance                      float64 `json:"cashBalance"`
	CashAvailableForTrading          float64 `json:"cashAvailableForTrading"`
	CashReceipts                     float64 `json:"cashReceipts"`
	DayTradingBuyingPower            float64 `json:"dayTradingBuyingPower"`
	DayTradingBuyingPowerCall        float64 `json:"dayTradingBuyingPowerCall"`
	DayTradingEquityCall             float64 `json:"dayTradingEquityCall"`
	Equity                           float64 `json:"equity"`
	EquityPercentage                 float64 `json:"equityPercentage"`
	LiquidationValue                 float64 `json:"liquidationValue"`
	LongMarginValue                  float64 `json:"longMarginValue"`
	LongOptionMarketValue            float64 `json:"longOptionMarketValue"`
	LongStockValue                   float64 `json:"longStockValue"`
	MaintenanceCall                  float64 `json:"maintenanceCall"`
	MaintenanceRequirement           float64 `json:"maintenanceRequirement"`
	Margin                           float64 `json:"margin"`
	MarginEquity                     float64 `json:"marginEquity"`
	MoneyMarketFund                  float64 `json:"moneyMarketFund"`
	MutualFundValue                  float64 `json:"mutualFundValue"`
	RegTCall                         float64 `json:"regTCall"`
	ShortMarginValue                 float64 `json:"shortMarginValue"`
	ShortOptionMarketValue           float64 `json:"shortOptionMarketValue"`
	ShortStockValue                  float64 `json:"shortStockValue"`
	TotalCash                        float64 `json:"totalCash"`
	TotalEquityValue                 float64 `json:"totalEquityValue"`
	IsInCall                         bool    `json:"isInCall"`
	PendingDeposits                  float64 `json:"pendingDeposits"`
	MarginBalance                    float64 `json:"marginBalance"`
	ShortBalance                     float64 `json:"shortBalance"`
	AccountValue                     float64 `json:"accountValue"`
}

// GetAccounts retrieves all linked brokerage accounts.
// fields: optional comma-separated list of additional fields to include (e.g., "positions");
// if empty, no fields param is sent.
func (c *Client) GetAccounts(ctx context.Context, fields string) ([]Account, error) {
	req, err := c.newRequest(ctx, "GET", "/accounts", nil)
	if err != nil {
		return nil, err
	}

	if fields != "" {
		q := req.URL.Query()
		q.Set("fields", fields)
		req.URL.RawQuery = q.Encode()
	}

	var result []Account
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetAccount retrieves a single brokerage account by its encrypted hash value.
// accountHash is the encrypted hash value from GetAccountNumbers, not the plain account number.
// fields: optional comma-separated list of additional fields to include (e.g., "positions");
// if empty, no fields param is sent.
func (c *Client) GetAccount(ctx context.Context, accountHash, fields string) (*Account, error) {
	path := "/accounts/" + url.PathEscape(accountHash)
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if fields != "" {
		q := req.URL.Query()
		q.Set("fields", fields)
		req.URL.RawQuery = q.Encode()
	}

	var result Account
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
