package trader

import (
	"context"
	"errors"
	"strconv"

	schwab "github.com/major/schwab-go/schwab"
)

// OrderListParams contains parameters for listing orders.
// FromEnteredTime and ToEnteredTime are required by the API in yyyy-MM-dd format.
type OrderListParams struct {
	MaxResults      int
	FromEnteredTime string
	ToEnteredTime   string
	Status          string
}

// Order represents a Schwab order, including recursive child and replacement orders.
type Order struct {
	Session                  string          `json:"session"`
	Duration                 string          `json:"duration"`
	OrderType                string          `json:"orderType"`
	ComplexOrderStrategyType string          `json:"complexOrderStrategyType"`
	Quantity                 float64         `json:"quantity"`
	FilledQuantity           float64         `json:"filledQuantity"`
	RemainingQuantity        float64         `json:"remainingQuantity"`
	RequestedDestination     string          `json:"requestedDestination"`
	DestinationLinkName      string          `json:"destinationLinkName"`
	Price                    float64         `json:"price"`
	StopPrice                float64         `json:"stopPrice"`
	StopPriceLinkBasis       string          `json:"stopPriceLinkBasis"`
	StopPriceLinkType        string          `json:"stopPriceLinkType"`
	StopPriceOffset          float64         `json:"stopPriceOffset"`
	StopType                 string          `json:"stopType"`
	PriceLinkBasis           string          `json:"priceLinkBasis"`
	PriceLinkType            string          `json:"priceLinkType"`
	TaxLotMethod             string          `json:"taxLotMethod"`
	ActivationPrice          float64         `json:"activationPrice"`
	SpecialInstruction       string          `json:"specialInstruction"`
	OrderStrategyType        string          `json:"orderStrategyType"`
	CancelTime               string          `json:"cancelTime"`
	ReleaseTime              string          `json:"releaseTime"`
	OrderLegCollection       []OrderLeg      `json:"orderLegCollection"`
	OrderActivityCollection  []OrderActivity `json:"orderActivityCollection"`
	ChildOrderStrategies     []*Order        `json:"childOrderStrategies"`
	ReplacingOrderCollection []*Order        `json:"replacingOrderCollection"`
	OrderID                  int64           `json:"orderId"`
	Cancelable               bool            `json:"cancelable"`
	Editable                 bool            `json:"editable"`
	Status                   string          `json:"status"`
	EnteredTime              string          `json:"enteredTime"`
	CloseTime                string          `json:"closeTime"`
	Tag                      string          `json:"tag"`
	AccountNumber            int64           `json:"accountNumber"`
	StatusDescription        string          `json:"statusDescription"`
}

// OrderLeg represents one leg within an order.
type OrderLeg struct {
	OrderLegType   string          `json:"orderLegType"`
	LegID          int64           `json:"legId"`
	Instrument     OrderInstrument `json:"instrument"`
	Instruction    string          `json:"instruction"`
	PositionEffect string          `json:"positionEffect"`
	Quantity       float64         `json:"quantity"`
	QuantityType   string          `json:"quantityType"`
	DivCapGains    string          `json:"divCapGains"`
	ToSymbol       string          `json:"toSymbol"`
}

// OrderInstrument represents an instrument within an order leg.
type OrderInstrument struct {
	AssetType    schwab.AssetType `json:"assetType"`
	CUSIP        string           `json:"cusip"`
	Symbol       string           `json:"symbol"`
	Description  string           `json:"description"`
	InstrumentID int64            `json:"instrumentId"`
	Type         string           `json:"type"`
}

// OrderActivity represents activity recorded against an order.
type OrderActivity struct {
	ActivityType           string         `json:"activityType"`
	ExecutionType          string         `json:"executionType"`
	Quantity               float64        `json:"quantity"`
	OrderRemainingQuantity float64        `json:"orderRemainingQuantity"`
	ExecutionLegs          []ExecutionLeg `json:"executionLegs"`
}

// ExecutionLeg represents an execution for one order leg.
type ExecutionLeg struct {
	LegID             int64   `json:"legId"`
	Price             float64 `json:"price"`
	Quantity          float64 `json:"quantity"`
	MismarkedQuantity float64 `json:"mismarkedQuantity"`
	InstrumentID      int64   `json:"instrumentId"`
	Time              string  `json:"time"`
}

// PreviewOrder is the response returned by the order preview endpoint.
type PreviewOrder struct {
	OrderID               int64                  `json:"orderId"`
	OrderStrategy         *OrderStrategy         `json:"orderStrategy"`
	OrderValidationResult *OrderValidationResult `json:"orderValidationResult"`
	CommissionAndFee      *CommissionAndFee      `json:"commissionAndFee"`
}

// OrderStrategy contains the previewed order strategy details.
type OrderStrategy struct {
	AccountNumber            string            `json:"accountNumber"`
	AdvancedOrderType        string            `json:"advancedOrderType"`
	CloseTime                string            `json:"closeTime"`
	EnteredTime              string            `json:"enteredTime"`
	OrderBalance             *OrderBalance     `json:"orderBalance"`
	OrderStrategyType        string            `json:"orderStrategyType"`
	OrderVersion             int               `json:"orderVersion"`
	Session                  string            `json:"session"`
	Status                   string            `json:"status"`
	AllOrNone                bool              `json:"allOrNone"`
	Discretionary            bool              `json:"discretionary"`
	Duration                 string            `json:"duration"`
	FilledQuantity           float64           `json:"filledQuantity"`
	OrderType                string            `json:"orderType"`
	OrderValue               float64           `json:"orderValue"`
	Price                    float64           `json:"price"`
	Quantity                 float64           `json:"quantity"`
	RemainingQuantity        float64           `json:"remainingQuantity"`
	SellNonMarginableFirst   bool              `json:"sellNonMarginableFirst"`
	SettlementInstruction    string            `json:"settlementInstruction"`
	SpecialInstruction       string            `json:"specialInstruction"`
	ComplexOrderStrategyType string            `json:"complexOrderStrategyType"`
	TaxLotMethod             string            `json:"taxLotMethod"`
	OrderLegs                []PreviewOrderLeg `json:"orderLegs"`
	ChildOrderStrategies     []OrderStrategy   `json:"childOrderStrategies"`
}

// OrderBalance contains estimated balance details for a previewed order.
type OrderBalance struct {
	OrderValue             float64 `json:"orderValue"`
	ProjectedAvailableFund float64 `json:"projectedAvailableFund"`
	ProjectedBuyingPower   float64 `json:"projectedBuyingPower"`
	ProjectedCommission    float64 `json:"projectedCommission"`
}

// PreviewOrderLeg represents one leg in an order preview response.
type PreviewOrderLeg struct {
	LegID       int64           `json:"legId"`
	Instrument  OrderInstrument `json:"instrument"`
	Instruction string          `json:"instruction"`
	Quantity    float64         `json:"quantity"`
}

// OrderValidationResult contains validation details for a previewed order.
type OrderValidationResult struct {
	Alerts  []OrderValidationDetail `json:"alerts"`
	Accepts []OrderValidationDetail `json:"accepts"`
	Rejects []OrderValidationDetail `json:"rejects"`
	Reviews []OrderValidationDetail `json:"reviews"`
	Warns   []OrderValidationDetail `json:"warns"`
}

// OrderValidationDetail represents a single validation message.
type OrderValidationDetail struct {
	ValidationRuleName string `json:"validationRuleName"`
	Message            string `json:"message"`
	ActivityMessage    string `json:"activityMessage"`
	OriginalSeverity   string `json:"originalSeverity"`
	OverrideName       string `json:"overrideName"`
	OverrideSeverity   string `json:"overrideSeverity"`
}

// CommissionAndFee contains the commission and fee breakdown for a previewed order.
type CommissionAndFee struct {
	Commission     *Commission `json:"commission"`
	Fee            *Fees       `json:"fee"`
	TrueCommission *Commission `json:"trueCommission"`
}

// Commission contains commission leg details.
type Commission struct {
	CommissionLegs []CommissionLeg `json:"commissionLegs"`
}

// CommissionLeg represents a single commission leg.
type CommissionLeg struct {
	CommissionValues []CommissionValue `json:"commissionValues"`
}

// CommissionValue represents a single commission value entry.
type CommissionValue struct {
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

// Fees contains fee leg details.
type Fees struct {
	FeeLegs []FeeLeg `json:"feeLegs"`
}

// FeeLeg represents a single fee leg.
type FeeLeg struct {
	FeeValues []FeeValue `json:"feeValues"`
}

// FeeValue represents a single fee value entry.
type FeeValue struct {
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

// GetOrders retrieves orders for a single account.
func (c *Client) GetOrders(ctx context.Context, accountHash string, params *OrderListParams) ([]Order, error) {
	return c.getOrders(ctx, accountPath(accountHash, "orders"), params)
}

// CreateOrder creates an order for a single account.
func (c *Client) CreateOrder(ctx context.Context, accountHash string, order *Order) error {
	req, err := c.newRequest(ctx, "POST", accountPath(accountHash, "orders"), order)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// GetOrder retrieves a single order for a single account.
func (c *Client) GetOrder(ctx context.Context, accountHash string, orderID int64) (*Order, error) {
	req, err := c.newRequest(ctx, "GET", accountPath(accountHash, "orders", strconv.FormatInt(orderID, 10)), nil)
	if err != nil {
		return nil, err
	}

	var result Order
	if doErr := c.do(req, &result); doErr != nil {
		return nil, doErr
	}
	return &result, nil
}

// ReplaceOrder replaces an existing order for a single account.
func (c *Client) ReplaceOrder(ctx context.Context, accountHash string, orderID int64, order *Order) error {
	req, err := c.newRequest(ctx, "PUT", accountPath(accountHash, "orders", strconv.FormatInt(orderID, 10)), order)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// CancelOrder cancels an existing order for a single account.
func (c *Client) CancelOrder(ctx context.Context, accountHash string, orderID int64) error {
	req, err := c.newRequest(ctx, "DELETE", accountPath(accountHash, "orders", strconv.FormatInt(orderID, 10)), nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// PreviewOrder previews commissions and fees for an order without placing it.
func (c *Client) PreviewOrder(ctx context.Context, accountHash string, order *Order) (*PreviewOrder, error) {
	req, err := c.newRequest(ctx, "POST", accountPath(accountHash, "previewOrder"), order)
	if err != nil {
		return nil, err
	}

	var result PreviewOrder
	if doErr := c.do(req, &result); doErr != nil {
		return nil, doErr
	}
	return &result, nil
}

// GetAllOrders retrieves orders across all accounts.
func (c *Client) GetAllOrders(ctx context.Context, params *OrderListParams) ([]Order, error) {
	return c.getOrders(ctx, "/orders", params)
}

func (c *Client) getOrders(ctx context.Context, path string, params *OrderListParams) ([]Order, error) {
	if params == nil {
		return nil, errors.New("order list params are required")
	}
	if params.FromEnteredTime == "" {
		return nil, errors.New("fromEnteredTime is required")
	}
	if params.ToEnteredTime == "" {
		return nil, errors.New("toEnteredTime is required")
	}

	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("fromEnteredTime", params.FromEnteredTime)
	q.Set("toEnteredTime", params.ToEnteredTime)
	setOptionalInt(q, "maxResults", params.MaxResults)
	setOptionalString(q, "status", params.Status)
	req.URL.RawQuery = q.Encode()

	var result []Order
	if doErr := c.do(req, &result); doErr != nil {
		return nil, doErr
	}
	return result, nil
}
