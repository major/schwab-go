package trader

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	schwab "github.com/major/schwab-go/schwab"
)

// TransactionType defines the type of transaction.
type TransactionType string

const (
	TransactionTypeTrade              TransactionType = "TRADE"
	TransactionTypeReceiveAndDeliver  TransactionType = "RECEIVE_AND_DELIVER"
	TransactionTypeDividendOrInterest TransactionType = "DIVIDEND_OR_INTEREST"
	TransactionTypeACHReceipt         TransactionType = "ACH_RECEIPT"
	TransactionTypeACHDisbursement    TransactionType = "ACH_DISBURSEMENT"
	TransactionTypeCashReceipt        TransactionType = "CASH_RECEIPT"
	TransactionTypeCashDisbursement   TransactionType = "CASH_DISBURSEMENT"
	TransactionTypeElectronicFund     TransactionType = "ELECTRONIC_FUND"
	TransactionTypeWireIn             TransactionType = "WIRE_IN"
	TransactionTypeWireOut            TransactionType = "WIRE_OUT"
	TransactionTypeJournal            TransactionType = "JOURNAL"
	TransactionTypeMemorandum         TransactionType = "MEMORANDUM"
	TransactionTypeMarginCall         TransactionType = "MARGIN_CALL"
	TransactionTypeMoneyMarket        TransactionType = "MONEY_MARKET"
	TransactionTypeSMAadjustment      TransactionType = "SMA_ADJUSTMENT"
)

// TransactionListParams contains parameters for listing transactions.
// StartDate, EndDate, and Types are required by the API.
type TransactionListParams struct {
	StartDate string // required, yyyy-MM-dd format
	EndDate   string // required, yyyy-MM-dd format
	Types     string // required, comma-separated TransactionType values
	Symbol    string // optional, filter by symbol
}

// UserDetails contains user information associated with a transaction.
type UserDetails struct {
	CDDomainID     string `json:"cdDomainId"`
	Login          string `json:"login"`
	Type           string `json:"type"`
	UserID         int64  `json:"userId"`
	SystemUserName string `json:"systemUserName"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	BrokerRepCode  string `json:"brokerRepCode"`
}

// Transaction represents a single account transaction.
type Transaction struct {
	ActivityID     int64           `json:"activityId"`
	Time           string          `json:"time"`
	User           *UserDetails    `json:"user"`
	Type           TransactionType `json:"type"`
	Status         string          `json:"status"`
	SubAccount     string          `json:"subAccount"`
	TradeDate      string          `json:"tradeDate"`
	SettlementDate string          `json:"settlementDate"`
	PositionID     int64           `json:"positionId"`
	OrderID        int64           `json:"orderId"`
	NetAmount      float64         `json:"netAmount"`
	ActivityType   string          `json:"activityType"`
	TransferItems  []TransferItem  `json:"transferItems"`
	Description    string          `json:"description"`
	AccountNumber  string          `json:"accountNumber"`
}

// TransferItem represents an item within a transaction.
type TransferItem struct {
	Instrument     TransactionInstrument `json:"instrument"`
	Amount         float64               `json:"amount"`
	Cost           float64               `json:"cost"`
	Price          float64               `json:"price"`
	FeeType        string                `json:"feeType"`
	PositionEffect string                `json:"positionEffect"`
}

// TransactionInstrument represents the financial instrument in a transaction.
// For options, the ExpirationDate, OptionDeliverables, OptionPremiumMultiplier,
// PutCall, StrikePrice, Type, UnderlyingSymbol, and UnderlyingCusip fields
// are populated. For fixed income instruments, MaturityDate, Factor,
// Multiplier, and VariableRate are populated.
type TransactionInstrument struct {
	AssetType               schwab.AssetType               `json:"assetType"`
	Cusip                   string                         `json:"cusip"`
	Symbol                  string                         `json:"symbol"`
	Description             string                         `json:"description"`
	InstrumentID            int64                          `json:"instrumentId"`
	ExpirationDate          string                         `json:"expirationDate,omitempty"`
	OptionDeliverables      []TransactionOptionDeliverable `json:"optionDeliverables,omitempty"`
	OptionPremiumMultiplier int64                          `json:"optionPremiumMultiplier,omitempty"`
	PutCall                 string                         `json:"putCall,omitempty"`
	StrikePrice             float64                        `json:"strikePrice,omitempty"`
	Type                    string                         `json:"type,omitempty"`
	UnderlyingSymbol        string                         `json:"underlyingSymbol,omitempty"`
	UnderlyingCusip         string                         `json:"underlyingCusip,omitempty"`
	MaturityDate            string                         `json:"maturityDate,omitempty"`
	Factor                  float64                        `json:"factor,omitempty"`
	Multiplier              float64                        `json:"multiplier,omitempty"`
	VariableRate            float64                        `json:"variableRate,omitempty"`
}

// TransactionOptionDeliverable represents a deliverable for a transaction option instrument.
type TransactionOptionDeliverable struct {
	RootSymbol        string           `json:"rootSymbol"`
	StrikePercent     int64            `json:"strikePercent"`
	DeliverableNumber int64            `json:"deliverableNumber"`
	DeliverableUnits  float64          `json:"deliverableUnits"`
	AssetType         schwab.AssetType `json:"assetType"`
}

// GetTransactions retrieves a list of transactions for the given account.
// params must not be nil; StartDate and EndDate are required by the API.
func (c *Client) GetTransactions(
	ctx context.Context,
	accountHash string,
	params *TransactionListParams,
) ([]Transaction, error) {
	if params == nil {
		return nil, errors.New("transaction list params are required")
	}
	if params.StartDate == "" {
		return nil, errors.New("startDate is required")
	}
	if params.EndDate == "" {
		return nil, errors.New("endDate is required")
	}
	if params.Types == "" {
		return nil, errors.New("types is required")
	}

	path := fmt.Sprintf("/accounts/%s/transactions", url.PathEscape(accountHash))
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("startDate", params.StartDate)
	q.Set("endDate", params.EndDate)
	q.Set("types", params.Types)
	if params.Symbol != "" {
		q.Set("symbol", params.Symbol)
	}
	req.URL.RawQuery = q.Encode()

	var result []Transaction
	if doErr := c.do(req, &result); doErr != nil {
		return nil, doErr
	}
	return result, nil
}

// GetTransaction retrieves transactions matching one transaction ID for the given account.
func (c *Client) GetTransaction(ctx context.Context, accountHash string, transactionID int64) ([]Transaction, error) {
	path := fmt.Sprintf(
		"/accounts/%s/transactions/%s",
		url.PathEscape(accountHash),
		strconv.FormatInt(transactionID, 10),
	)
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result []Transaction
	if doErr := c.do(req, &result); doErr != nil {
		return nil, doErr
	}
	return result, nil
}
