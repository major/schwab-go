package trader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schwab "github.com/major/schwab-go/schwab"
)

func TestGetOrders(t *testing.T) {
	fixture := []Order{testOrderFixture()}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/accounts/HASH_ABC123/orders", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "2024-01-01", r.URL.Query().Get("fromEnteredTime"))
		assert.Equal(t, "2024-01-31", r.URL.Query().Get("toEnteredTime"))
		assert.Equal(t, "25", r.URL.Query().Get("maxResults"))
		assert.Equal(t, "FILLED", r.URL.Query().Get("status"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, fixture)
	})

	result, err := client.GetOrders(context.Background(), "HASH_ABC123", &OrderListParams{
		MaxResults:      25,
		FromEnteredTime: "2024-01-01",
		ToEnteredTime:   "2024-01-31",
		Status:          OrderStatusFilled,
	})
	require.NoError(t, err)
	require.Len(t, result, 1)
	assertOrderFixture(t, &result[0])
}

func TestCreateOrder(t *testing.T) {
	order := &OrderRequest{
		Session:   SessionNormal,
		Duration:  DurationDay,
		OrderType: OrderTypeRequestMarket,
		OrderLegCollection: []OrderLegRequest{
			{
				Instruction: InstructionBuy,
				Quantity:    3,
				Instrument: OrderInstrument{
					AssetType: schwab.AssetTypeEquity,
					Symbol:    "MSFT",
				},
			},
		},
	}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/accounts/HASH_ABC123/orders", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err) {
			return
		}

		var got OrderRequest
		assert.NoError(t, json.NewDecoder(bytes.NewReader(body)).Decode(&got))
		assert.Equal(t, SessionNormal, got.Session)
		assert.Equal(t, DurationDay, got.Duration)
		assert.Equal(t, OrderTypeRequestMarket, got.OrderType)
		if assert.Len(t, got.OrderLegCollection, 1) {
			assert.Equal(t, InstructionBuy, got.OrderLegCollection[0].Instruction)
			assert.InDelta(t, 3.0, got.OrderLegCollection[0].Quantity, 0.000001)
			assert.Equal(t, "MSFT", got.OrderLegCollection[0].Instrument.Symbol)
		}

		var raw map[string]json.RawMessage
		assert.NoError(t, json.NewDecoder(bytes.NewReader(body)).Decode(&raw))
		assert.NotContains(t, raw, "requestedDestination")
		assert.NotContains(t, raw, "tag")

		w.WriteHeader(http.StatusCreated)
	})

	err := client.CreateOrder(context.Background(), "HASH_ABC123", order)
	require.NoError(t, err)
}

func TestGetOrder(t *testing.T) {
	fixture := testOrderFixture()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/accounts/HASH_ABC123/orders/9001", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, fixture)
	})

	result, err := client.GetOrder(context.Background(), "HASH_ABC123", 9001)
	require.NoError(t, err)
	require.NotNil(t, result)
	assertOrderFixture(t, result)
}

func TestReplaceOrder(t *testing.T) {
	order := &OrderRequest{
		Session:   SessionNormal,
		Duration:  DurationDay,
		OrderType: OrderTypeRequestLimit,
		Price:     199.50,
	}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/accounts/HASH_ABC123/orders/9001", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var got OrderRequest
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, OrderTypeRequestLimit, got.OrderType)
		assert.InDelta(t, 199.50, got.Price, 0.000001)

		w.WriteHeader(http.StatusOK)
	})

	err := client.ReplaceOrder(context.Background(), "HASH_ABC123", 9001, order)
	require.NoError(t, err)
}

func TestCancelOrder(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/accounts/HASH_ABC123/orders/9001", r.URL.Path)
		assert.Empty(t, r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusOK)
	})

	err := client.CancelOrder(context.Background(), "HASH_ABC123", 9001)
	require.NoError(t, err)
}

func TestPreviewOrder(t *testing.T) {
	order := &PreviewOrderRequest{
		OrderStrategy: &OrderStrategy{
			Session:   SessionNormal,
			Duration:  DurationDay,
			OrderType: OrderTypeMarket,
		},
	}
	fixture := PreviewOrder{
		OrderID: 9001,
		OrderStrategy: &OrderStrategy{
			AccountNumber:     "123456789",
			OrderStrategyType: OrderStrategyTypeSingle,
			OrderType:         OrderTypeMarket,
			Session:           SessionNormal,
			Duration:          DurationDay,
			Quantity:          1,
			AmountIndicator:   AmountIndicatorShares,
			Strategy:          ComplexOrderStrategyTypeNone,
			OrderLegs: []PreviewOrderLeg{
				{
					LegID:       1,
					AssetType:   schwab.AssetTypeEquity,
					FinalSymbol: "AAPL",
					Instruction: InstructionBuy,
					Quantity:    1,
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

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/accounts/HASH_ABC123/previewOrder", r.URL.Path)

		var got PreviewOrderRequest
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		if !assert.NotNil(t, got.OrderStrategy) {
			return
		}
		assert.Equal(t, OrderTypeMarket, got.OrderStrategy.OrderType)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, fixture)
	})

	result, err := client.PreviewOrder(context.Background(), "HASH_ABC123", order)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(9001), result.OrderID)
	require.NotNil(t, result.OrderStrategy)
	assert.Equal(t, OrderStrategyTypeSingle, result.OrderStrategy.OrderStrategyType)
	assert.Equal(t, OrderTypeMarket, result.OrderStrategy.OrderType)
	assert.Equal(t, AmountIndicatorShares, result.OrderStrategy.AmountIndicator)
	assert.Equal(t, ComplexOrderStrategyTypeNone, result.OrderStrategy.Strategy)
	require.Len(t, result.OrderStrategy.OrderLegs, 1)
	assert.Equal(t, InstructionBuy, result.OrderStrategy.OrderLegs[0].Instruction)
	assert.Equal(t, schwab.AssetTypeEquity, result.OrderStrategy.OrderLegs[0].AssetType)
	assert.Equal(t, "AAPL", result.OrderStrategy.OrderLegs[0].FinalSymbol)
	require.NotNil(t, result.OrderValidationResult)
	require.Len(t, result.OrderValidationResult.Alerts, 1)
	assert.Equal(t, "MarketHoursRule", result.OrderValidationResult.Alerts[0].ValidationRuleName)
	require.NotNil(t, result.CommissionAndFee)
	require.NotNil(t, result.CommissionAndFee.Fee)
	require.Len(t, result.CommissionAndFee.Fee.FeeLegs, 1)
	assert.Equal(t, "SEC_FEE", result.CommissionAndFee.Fee.FeeLegs[0].FeeValues[0].Type)
	assert.InDelta(t, 0.02, result.CommissionAndFee.Fee.FeeLegs[0].FeeValues[0].Value, 0.000001)
}

func TestGetAllOrders(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/orders", r.URL.Path)
		assert.Equal(t, "2024-01-01", r.URL.Query().Get("fromEnteredTime"))
		assert.Equal(t, "2024-01-31", r.URL.Query().Get("toEnteredTime"))
		assert.Empty(t, r.URL.Query().Get("maxResults"))
		assert.Empty(t, r.URL.Query().Get("status"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, []Order{testOrderFixture()})
	})

	result, err := client.GetAllOrders(context.Background(), &OrderListParams{
		FromEnteredTime: "2024-01-01",
		ToEnteredTime:   "2024-01-31",
	})
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, int64(9001), result[0].OrderID)
}

func TestGetOrdersRequiresParams(t *testing.T) {
	client := NewClient()

	_, err := client.GetOrders(context.Background(), "HASH_ABC123", nil)
	require.EqualError(t, err, "order list params are required")

	_, err = client.GetAllOrders(context.Background(), &OrderListParams{ToEnteredTime: "2024-01-31"})
	require.EqualError(t, err, "fromEnteredTime is required")

	_, err = client.GetAllOrders(context.Background(), &OrderListParams{FromEnteredTime: "2024-01-01"})
	require.EqualError(t, err, "toEnteredTime is required")
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
	assert.Equal(t, OrderTypeStop, order.ChildOrderStrategies[0].ChildOrderStrategies[0].OrderType)
	require.Len(t, order.ReplacingOrderCollection, 1)
	assert.Equal(t, int64(4), order.ReplacingOrderCollection[0].OrderID)
}

func TestGetOrdersError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

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
		Session:                  SessionNormal,
		Duration:                 DurationDay,
		OrderType:                OrderTypeLimit,
		ComplexOrderStrategyType: ComplexOrderStrategyTypeNone,
		Quantity:                 10,
		FilledQuantity:           10,
		RemainingQuantity:        0,
		RequestedDestination:     RequestedDestinationAuto,
		DestinationLinkName:      "NYSE",
		Price:                    150.25,
		StopPrice:                145.00,
		StopPriceLinkBasis:       PriceLinkBasisLast,
		StopPriceLinkType:        PriceLinkTypeValue,
		StopType:                 StopTypeStandard,
		OrderID:                  9001,
		Cancelable:               false,
		Editable:                 false,
		Status:                   OrderStatusFilled,
		EnteredTime:              "2024-01-15T10:30:00Z",
		CloseTime:                "2024-01-15T10:31:00Z",
		Tag:                      "client-tag",
		AccountNumber:            123456789,
		StatusDescription:        "Order filled",
		OrderLegCollection: []OrderLeg{
			{
				OrderLegType:   OrderLegTypeEquity,
				LegID:          1,
				Instruction:    InstructionBuy,
				PositionEffect: PositionEffectOpening,
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

func assertOrderFixture(t *testing.T, order *Order) {
	t.Helper()

	assert.Equal(t, SessionNormal, order.Session)
	assert.Equal(t, DurationDay, order.Duration)
	assert.Equal(t, OrderTypeLimit, order.OrderType)
	assert.Equal(t, ComplexOrderStrategyTypeNone, order.ComplexOrderStrategyType)
	assert.InDelta(t, 10.0, order.Quantity, 0.000001)
	assert.InDelta(t, 10.0, order.FilledQuantity, 0.000001)
	assert.InDelta(t, 0.0, order.RemainingQuantity, 0.000001)
	assert.Equal(t, RequestedDestinationAuto, order.RequestedDestination)
	assert.Equal(t, "NYSE", order.DestinationLinkName)
	assert.InDelta(t, 150.25, order.Price, 0.000001)
	assert.InDelta(t, 145.00, order.StopPrice, 0.000001)
	assert.Equal(t, PriceLinkBasisLast, order.StopPriceLinkBasis)
	assert.Equal(t, PriceLinkTypeValue, order.StopPriceLinkType)
	assert.Equal(t, StopTypeStandard, order.StopType)
	assert.Equal(t, int64(9001), order.OrderID)
	assert.False(t, order.Cancelable)
	assert.False(t, order.Editable)
	assert.Equal(t, OrderStatusFilled, order.Status)
	assert.Equal(t, "2024-01-15T10:30:00Z", order.EnteredTime)
	assert.Equal(t, "2024-01-15T10:31:00Z", order.CloseTime)
	assert.Equal(t, "client-tag", order.Tag)
	assert.Equal(t, int64(123456789), order.AccountNumber)
	assert.Equal(t, "Order filled", order.StatusDescription)

	require.Len(t, order.OrderLegCollection, 1)
	leg := order.OrderLegCollection[0]
	assert.Equal(t, OrderLegTypeEquity, leg.OrderLegType)
	assert.Equal(t, int64(1), leg.LegID)
	assert.Equal(t, InstructionBuy, leg.Instruction)
	assert.Equal(t, PositionEffectOpening, leg.PositionEffect)
	assert.InDelta(t, 10.0, leg.Quantity, 0.000001)
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
	assert.InDelta(t, 10.0, activity.Quantity, 0.000001)
	assert.InDelta(t, 0.0, activity.OrderRemainingQuantity, 0.000001)
	require.Len(t, activity.ExecutionLegs, 1)
	execution := activity.ExecutionLegs[0]
	assert.Equal(t, int64(1), execution.LegID)
	assert.InDelta(t, 150.25, execution.Price, 0.000001)
	assert.InDelta(t, 10.0, execution.Quantity, 0.000001)
	assert.InDelta(t, 0.0, execution.MismarkedQuantity, 0.000001)
	assert.Equal(t, int64(1234567), execution.InstrumentID)
	assert.Equal(t, "2024-01-15T10:31:00Z", execution.Time)
}
