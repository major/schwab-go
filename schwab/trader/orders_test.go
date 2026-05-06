package trader

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schwab "github.com/major/schwab-go/schwab"
)

func TestGetOrders(t *testing.T) {
	fixture := []Order{testOrderFixture()}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/accounts/HASH_ABC123/orders", r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		require.Equal(t, "2024-01-01", r.URL.Query().Get("fromEnteredTime"))
		require.Equal(t, "2024-01-31", r.URL.Query().Get("toEnteredTime"))
		require.Equal(t, "25", r.URL.Query().Get("maxResults"))
		require.Equal(t, "FILLED", r.URL.Query().Get("status"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(fixture)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithToken("test-token"),
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	result, err := client.GetOrders(context.Background(), "HASH_ABC123", &OrderListParams{
		MaxResults:      25,
		FromEnteredTime: "2024-01-01",
		ToEnteredTime:   "2024-01-31",
		Status:          "FILLED",
	})
	require.NoError(t, err)
	require.Len(t, result, 1)
	assertOrderFixture(t, result[0])
}

func TestCreateOrder(t *testing.T) {
	order := &Order{
		Session:   "NORMAL",
		Duration:  "DAY",
		OrderType: "MARKET",
		OrderLegCollection: []OrderLeg{
			{
				Instruction: "BUY",
				Quantity:    3,
				Instrument: OrderInstrument{
					AssetType: schwab.AssetTypeEquity,
					Symbol:    "MSFT",
				},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/accounts/HASH_ABC123/orders", r.URL.Path)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var got Order
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "NORMAL", got.Session)
		assert.Equal(t, "DAY", got.Duration)
		assert.Equal(t, "MARKET", got.OrderType)
		require.Len(t, got.OrderLegCollection, 1)
		assert.Equal(t, "BUY", got.OrderLegCollection[0].Instruction)
		assert.Equal(t, 3.0, got.OrderLegCollection[0].Quantity)
		assert.Equal(t, "MSFT", got.OrderLegCollection[0].Instrument.Symbol)

		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	err := client.CreateOrder(context.Background(), "HASH_ABC123", order)
	require.NoError(t, err)
}

func TestGetOrder(t *testing.T) {
	fixture := testOrderFixture()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/accounts/HASH_ABC123/orders/9001", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(fixture)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	result, err := client.GetOrder(context.Background(), "HASH_ABC123", "9001")
	require.NoError(t, err)
	require.NotNil(t, result)
	assertOrderFixture(t, *result)
}

func TestReplaceOrder(t *testing.T) {
	order := &Order{
		Session:   "NORMAL",
		Duration:  "DAY",
		OrderType: "LIMIT",
		Price:     199.50,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPut, r.Method)
		require.Equal(t, "/accounts/HASH_ABC123/orders/9001", r.URL.Path)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var got Order
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "LIMIT", got.OrderType)
		assert.Equal(t, 199.50, got.Price)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	err := client.ReplaceOrder(context.Background(), "HASH_ABC123", "9001", order)
	require.NoError(t, err)
}

func TestCancelOrder(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/accounts/HASH_ABC123/orders/9001", r.URL.Path)
		require.Empty(t, r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	err := client.CancelOrder(context.Background(), "HASH_ABC123", "9001")
	require.NoError(t, err)
}

func TestPreviewOrder(t *testing.T) {
	order := &Order{
		Session:   "NORMAL",
		Duration:  "DAY",
		OrderType: "MARKET",
	}
	fixture := PreviewOrder{
		OrderID: 9001,
		OrderStrategy: &OrderStrategy{
			AccountNumber:     "123456789",
			OrderStrategyType: "SINGLE",
			OrderType:         "MARKET",
			Session:           "NORMAL",
			Duration:          "DAY",
			Quantity:          1,
			OrderLegs: []PreviewOrderLeg{
				{
					LegID:       1,
					Instruction: "BUY",
					Quantity:    1,
					Instrument: OrderInstrument{
						AssetType: schwab.AssetTypeEquity,
						Symbol:    "AAPL",
					},
				},
			},
		},
		OrderValidationResult: &OrderValidationResult{
			Alerts: []OrderValidationDetail{
				{
					ValidationRuleName: "MarketHoursRule",
					Message:            "Market is currently closed",
				},
			},
		},
		CommissionAndFee: &CommissionAndFee{
			Commission: &Commission{
				CommissionLegs: []CommissionLeg{
					{
						CommissionValues: []CommissionValue{
							{Value: 0.00, Type: "COMMISSION"},
						},
					},
				},
			},
			Fee: &Fees{
				FeeLegs: []FeeLeg{
					{
						FeeValues: []FeeValue{
							{Value: 0.02, Type: "SEC_FEE"},
						},
					},
				},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/accounts/HASH_ABC123/previewOrder", r.URL.Path)

		var got Order
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "MARKET", got.OrderType)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(fixture)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	result, err := client.PreviewOrder(context.Background(), "HASH_ABC123", order)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(9001), result.OrderID)
	require.NotNil(t, result.OrderStrategy)
	assert.Equal(t, "SINGLE", result.OrderStrategy.OrderStrategyType)
	assert.Equal(t, "MARKET", result.OrderStrategy.OrderType)
	require.Len(t, result.OrderStrategy.OrderLegs, 1)
	assert.Equal(t, "BUY", result.OrderStrategy.OrderLegs[0].Instruction)
	require.NotNil(t, result.OrderValidationResult)
	require.Len(t, result.OrderValidationResult.Alerts, 1)
	assert.Equal(t, "MarketHoursRule", result.OrderValidationResult.Alerts[0].ValidationRuleName)
	require.NotNil(t, result.CommissionAndFee)
	require.NotNil(t, result.CommissionAndFee.Fee)
	require.Len(t, result.CommissionAndFee.Fee.FeeLegs, 1)
	assert.Equal(t, "SEC_FEE", result.CommissionAndFee.Fee.FeeLegs[0].FeeValues[0].Type)
	assert.Equal(t, 0.02, result.CommissionAndFee.Fee.FeeLegs[0].FeeValues[0].Value)
}

func TestGetAllOrders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/orders", r.URL.Path)
		require.Equal(t, "2024-01-01", r.URL.Query().Get("fromEnteredTime"))
		require.Equal(t, "2024-01-31", r.URL.Query().Get("toEnteredTime"))
		require.Empty(t, r.URL.Query().Get("maxResults"))
		require.Empty(t, r.URL.Query().Get("status"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Order{testOrderFixture()})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	result, err := client.GetAllOrders(context.Background(), &OrderListParams{
		FromEnteredTime: "2024-01-01",
		ToEnteredTime:   "2024-01-31",
	})
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, int64(9001), result[0].OrderID)
}

func TestRecursiveOrder(t *testing.T) {
	data := []byte(`{
		"orderId": 1,
		"orderType": "TRIGGER",
		"childOrderStrategies": [
			{
				"orderId": 2,
				"orderType": "LIMIT",
				"childOrderStrategies": [
					{"orderId": 3, "orderType": "STOP"}
				]
			}
		],
		"replacingOrderCollection": [
			{"orderId": 4, "orderType": "MARKET"}
		]
	}`)

	var order Order
	require.NoError(t, json.Unmarshal(data, &order))
	assert.Equal(t, int64(1), order.OrderID)
	require.Len(t, order.ChildOrderStrategies, 1)
	assert.Equal(t, int64(2), order.ChildOrderStrategies[0].OrderID)
	require.Len(t, order.ChildOrderStrategies[0].ChildOrderStrategies, 1)
	assert.Equal(t, int64(3), order.ChildOrderStrategies[0].ChildOrderStrategies[0].OrderID)
	assert.Equal(t, "STOP", order.ChildOrderStrategies[0].ChildOrderStrategies[0].OrderType)
	require.Len(t, order.ReplacingOrderCollection, 1)
	assert.Equal(t, int64(4), order.ReplacingOrderCollection[0].OrderID)
}

func TestGetOrdersError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetOrders(context.Background(), "HASH_ABC123", &OrderListParams{
		FromEnteredTime: "2024-01-01",
		ToEnteredTime:   "2024-01-31",
	})
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
}

func testOrderFixture() Order {
	return Order{
		Session:                  "NORMAL",
		Duration:                 "DAY",
		OrderType:                "LIMIT",
		ComplexOrderStrategyType: "NONE",
		Quantity:                 10,
		FilledQuantity:           10,
		RemainingQuantity:        0,
		RequestedDestination:     "AUTO",
		DestinationLinkName:      "NYSE",
		Price:                    150.25,
		StopPrice:                145.00,
		StopPriceLinkBasis:       "LAST",
		StopPriceLinkType:        "VALUE",
		StopType:                 "STANDARD",
		OrderID:                  9001,
		Cancelable:               false,
		Editable:                 false,
		Status:                   "FILLED",
		EnteredTime:              "2024-01-15T10:30:00Z",
		CloseTime:                "2024-01-15T10:31:00Z",
		Tag:                      "client-tag",
		AccountNumber:            123456789,
		StatusDescription:        "Order filled",
		OrderLegCollection: []OrderLeg{
			{
				OrderLegType:   "EQUITY",
				LegID:          1,
				Instruction:    "BUY",
				PositionEffect: "OPENING",
				Quantity:       10,
				Instrument: OrderInstrument{
					AssetType:    schwab.AssetTypeEquity,
					CUSIP:        "037833100",
					Symbol:       "AAPL",
					Description:  "Apple Inc",
					InstrumentID: 1234567,
					Type:         "COMMON_STOCK",
				},
			},
		},
		OrderActivityCollection: []OrderActivity{
			{
				ActivityType:           "EXECUTION",
				ExecutionType:          "FILL",
				Quantity:               10,
				OrderRemainingQuantity: 0,
				ExecutionLegs: []ExecutionLeg{
					{
						LegID:             1,
						Price:             150.25,
						Quantity:          10,
						MismarkedQuantity: 0,
						InstrumentID:      1234567,
						Time:              "2024-01-15T10:31:00Z",
					},
				},
			},
		},
	}
}

func assertOrderFixture(t *testing.T, order Order) {
	t.Helper()

	assert.Equal(t, "NORMAL", order.Session)
	assert.Equal(t, "DAY", order.Duration)
	assert.Equal(t, "LIMIT", order.OrderType)
	assert.Equal(t, "NONE", order.ComplexOrderStrategyType)
	assert.Equal(t, 10.0, order.Quantity)
	assert.Equal(t, 10.0, order.FilledQuantity)
	assert.Equal(t, 0.0, order.RemainingQuantity)
	assert.Equal(t, "AUTO", order.RequestedDestination)
	assert.Equal(t, "NYSE", order.DestinationLinkName)
	assert.Equal(t, 150.25, order.Price)
	assert.Equal(t, 145.00, order.StopPrice)
	assert.Equal(t, "LAST", order.StopPriceLinkBasis)
	assert.Equal(t, "VALUE", order.StopPriceLinkType)
	assert.Equal(t, "STANDARD", order.StopType)
	assert.Equal(t, int64(9001), order.OrderID)
	assert.False(t, order.Cancelable)
	assert.False(t, order.Editable)
	assert.Equal(t, "FILLED", order.Status)
	assert.Equal(t, "2024-01-15T10:30:00Z", order.EnteredTime)
	assert.Equal(t, "2024-01-15T10:31:00Z", order.CloseTime)
	assert.Equal(t, "client-tag", order.Tag)
	assert.Equal(t, int64(123456789), order.AccountNumber)
	assert.Equal(t, "Order filled", order.StatusDescription)

	require.Len(t, order.OrderLegCollection, 1)
	leg := order.OrderLegCollection[0]
	assert.Equal(t, "EQUITY", leg.OrderLegType)
	assert.Equal(t, int64(1), leg.LegID)
	assert.Equal(t, "BUY", leg.Instruction)
	assert.Equal(t, "OPENING", leg.PositionEffect)
	assert.Equal(t, 10.0, leg.Quantity)
	assert.Equal(t, schwab.AssetTypeEquity, leg.Instrument.AssetType)
	assert.Equal(t, "037833100", leg.Instrument.CUSIP)
	assert.Equal(t, "AAPL", leg.Instrument.Symbol)
	assert.Equal(t, "Apple Inc", leg.Instrument.Description)
	assert.Equal(t, int64(1234567), leg.Instrument.InstrumentID)
	assert.Equal(t, "COMMON_STOCK", leg.Instrument.Type)

	require.Len(t, order.OrderActivityCollection, 1)
	activity := order.OrderActivityCollection[0]
	assert.Equal(t, "EXECUTION", activity.ActivityType)
	assert.Equal(t, "FILL", activity.ExecutionType)
	assert.Equal(t, 10.0, activity.Quantity)
	assert.Equal(t, 0.0, activity.OrderRemainingQuantity)
	require.Len(t, activity.ExecutionLegs, 1)
	execution := activity.ExecutionLegs[0]
	assert.Equal(t, int64(1), execution.LegID)
	assert.Equal(t, 150.25, execution.Price)
	assert.Equal(t, 10.0, execution.Quantity)
	assert.Equal(t, 0.0, execution.MismarkedQuantity)
	assert.Equal(t, int64(1234567), execution.InstrumentID)
	assert.Equal(t, "2024-01-15T10:31:00Z", execution.Time)
}
