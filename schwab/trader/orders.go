package trader

import (
	"context"
	"fmt"
	"net/url"
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
	StopType                 string          `json:"stopType"`
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
	OrderID           int64             `json:"orderId"`
	OrderStrategy     string            `json:"orderStrategy"`
	CommissionAndFees CommissionAndFees `json:"commissionAndFees"`
	OrderLegs         []PreviewOrderLeg `json:"orderLegs"`
}

// CommissionAndFees contains previewed commission and fee details.
type CommissionAndFees struct {
	Commission float64 `json:"commission"`
	Fees       []Fee   `json:"fees"`
}

// Fee represents a single fee in an order preview.
type Fee struct {
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
}

// PreviewOrderLeg represents one leg in an order preview response.
type PreviewOrderLeg struct {
	LegID       int64           `json:"legId"`
	Instrument  OrderInstrument `json:"instrument"`
	Instruction string          `json:"instruction"`
	Quantity    float64         `json:"quantity"`
}

// GetOrders retrieves orders for a single account.
func (c *Client) GetOrders(ctx context.Context, accountHash string, params *OrderListParams) ([]Order, error) {
	path := fmt.Sprintf("/accounts/%s/orders", url.PathEscape(accountHash))
	return c.getOrders(ctx, path, params)
}

// CreateOrder creates an order for a single account.
func (c *Client) CreateOrder(ctx context.Context, accountHash string, order *Order) error {
	path := fmt.Sprintf("/accounts/%s/orders", url.PathEscape(accountHash))
	req, err := c.newRequest(ctx, "POST", path, order)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// GetOrder retrieves a single order for a single account.
func (c *Client) GetOrder(ctx context.Context, accountHash string, orderID string) (*Order, error) {
	path := fmt.Sprintf("/accounts/%s/orders/%s", url.PathEscape(accountHash), url.PathEscape(orderID))
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result Order
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ReplaceOrder replaces an existing order for a single account.
func (c *Client) ReplaceOrder(ctx context.Context, accountHash string, orderID string, order *Order) error {
	path := fmt.Sprintf("/accounts/%s/orders/%s", url.PathEscape(accountHash), url.PathEscape(orderID))
	req, err := c.newRequest(ctx, "PUT", path, order)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// CancelOrder cancels an existing order for a single account.
func (c *Client) CancelOrder(ctx context.Context, accountHash string, orderID string) error {
	path := fmt.Sprintf("/accounts/%s/orders/%s", url.PathEscape(accountHash), url.PathEscape(orderID))
	req, err := c.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// PreviewOrder previews commissions and fees for an order without placing it.
func (c *Client) PreviewOrder(ctx context.Context, accountHash string, order *Order) (*PreviewOrder, error) {
	path := fmt.Sprintf("/accounts/%s/previewOrder", url.PathEscape(accountHash))
	req, err := c.newRequest(ctx, "POST", path, order)
	if err != nil {
		return nil, err
	}

	var result PreviewOrder
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAllOrders retrieves orders across all accounts.
func (c *Client) GetAllOrders(ctx context.Context, params *OrderListParams) ([]Order, error) {
	return c.getOrders(ctx, "/orders", params)
}

func (c *Client) getOrders(ctx context.Context, path string, params *OrderListParams) ([]Order, error) {
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("fromEnteredTime", params.FromEnteredTime)
	q.Set("toEnteredTime", params.ToEnteredTime)
	if params.MaxResults != 0 {
		q.Set("maxResults", strconv.Itoa(params.MaxResults))
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	req.URL.RawQuery = q.Encode()

	var result []Order
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result, nil
}
