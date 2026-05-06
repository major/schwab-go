package trader

import (
	"context"
	"fmt"
	"net/url"

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
// StartDate and EndDate are required by the API.
type TransactionListParams struct {
	StartDate string // required, yyyy-MM-dd format
	EndDate   string // required, yyyy-MM-dd format
	Types     string // optional, comma-separated TransactionType values
	Symbol    string // optional, filter by symbol
}

// Transaction represents a single account transaction.
type Transaction struct {
	ActivityId    int64           `json:"activityId"`
	Time          string          `json:"time"`
	Type          TransactionType `json:"type"`
	Status        string          `json:"status"`
	SubAccount    string          `json:"subAccount"`
	TradeDate     string          `json:"tradeDate"`
	PositionId    int64           `json:"positionId"`
	OrderId       int64           `json:"orderId"`
	NetAmount     float64         `json:"netAmount"`
	TransferItems []TransferItem  `json:"transferItems"`
	Description   string          `json:"description"`
	AccountNumber string          `json:"accountNumber"`
}

// TransferItem represents an item within a transaction.
type TransferItem struct {
	Instrument TransactionInstrument `json:"instrument"`
	Amount     float64               `json:"amount"`
	Cost       float64               `json:"cost"`
	Price      float64               `json:"price"`
}

// TransactionInstrument represents the financial instrument in a transaction.
type TransactionInstrument struct {
	AssetType    schwab.AssetType `json:"assetType"`
	Cusip        string           `json:"cusip"`
	Symbol       string           `json:"symbol"`
	Description  string           `json:"description"`
	InstrumentId int64            `json:"instrumentId"`
}

// GetTransactions retrieves a list of transactions for the given account.
// params must not be nil; StartDate and EndDate are required by the API.
func (c *Client) GetTransactions(ctx context.Context, accountHash string, params *TransactionListParams) ([]Transaction, error) {
	path := fmt.Sprintf("/accounts/%s/transactions", url.PathEscape(accountHash))
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("startDate", params.StartDate)
	q.Set("endDate", params.EndDate)
	if params.Types != "" {
		q.Set("types", params.Types)
	}
	if params.Symbol != "" {
		q.Set("symbol", params.Symbol)
	}
	req.URL.RawQuery = q.Encode()

	var result []Transaction
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetTransaction retrieves a single transaction by ID for the given account.
func (c *Client) GetTransaction(ctx context.Context, accountHash string, transactionID string) (*Transaction, error) {
	path := fmt.Sprintf("/accounts/%s/transactions/%s", url.PathEscape(accountHash), url.PathEscape(transactionID))
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result Transaction
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
