package trader_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"

	schwab "github.com/major/schwab-go/schwab"
	"github.com/major/schwab-go/schwab/trader"
)

func ExampleClient_GetOrders() {
	client, closeServer := newOrderExampleClient(func(w http.ResponseWriter, _ *http.Request) {
		writeExampleJSON(w, []trader.Order{
			{
				OrderID:     9001,
				OrderType:   trader.OrderTypeLimit,
				Status:      trader.OrderStatusFilled,
				EnteredTime: "2024-01-15T10:30:00Z",
				OrderLegCollection: []trader.OrderLeg{
					{
						OrderLegType: trader.OrderLegTypeEquity,
						Instruction:  trader.InstructionBuy,
						Quantity:     10,
						Instrument: trader.OrderInstrument{
							AssetType: schwab.AssetTypeEquity,
							Symbol:    "AAPL",
						},
					},
				},
			},
		})
	})
	defer closeServer()

	orders, err := client.GetOrders(context.Background(), "HASH_ABC123", &trader.OrderListParams{
		MaxResults:      25,
		FromEnteredTime: "2024-01-01",
		ToEnteredTime:   "2024-01-31",
		Status:          trader.OrderStatusFilled,
	})
	if err != nil {
		panic(err)
	}

	printExampleLines(
		"request=GET /accounts/HASH_ABC123/orders?fromEnteredTime=2024-01-01&maxResults=25&status=FILLED&toEnteredTime=2024-01-31",
		"orders="+strconv.Itoa(len(orders)),
		"first="+strconv.FormatInt(orders[0].OrderID, 10)+" "+string(orders[0].Status),
	)

	// Output:
	// request=GET /accounts/HASH_ABC123/orders?fromEnteredTime=2024-01-01&maxResults=25&status=FILLED&toEnteredTime=2024-01-31
	// orders=1
	// first=9001 FILLED
}

func ExampleClient_GetAllOrders() {
	client, closeServer := newOrderExampleClient(func(w http.ResponseWriter, _ *http.Request) {
		writeExampleJSON(w, []trader.Order{{OrderID: 9001, Status: trader.OrderStatusWorking}})
	})
	defer closeServer()

	orders, err := client.GetAllOrders(context.Background(), &trader.OrderListParams{
		FromEnteredTime: "2024-01-01",
		ToEnteredTime:   "2024-01-31",
	})
	if err != nil {
		panic(err)
	}

	printExampleLines(
		"request=GET /orders?fromEnteredTime=2024-01-01&toEnteredTime=2024-01-31",
		"orders="+strconv.Itoa(len(orders)),
	)

	// Output:
	// request=GET /orders?fromEnteredTime=2024-01-01&toEnteredTime=2024-01-31
	// orders=1
}

func ExampleClient_GetOrder() {
	client, closeServer := newOrderExampleClient(func(w http.ResponseWriter, _ *http.Request) {
		writeExampleJSON(w, trader.Order{
			OrderID:   9001,
			OrderType: trader.OrderTypeLimit,
			Status:    trader.OrderStatusFilled,
			OrderActivityCollection: []trader.OrderActivity{
				{
					ActivityType:  "EXECUTION",
					ExecutionType: "FILL",
					Quantity:      10,
					ExecutionLegs: []trader.ExecutionLeg{{LegID: 1, Price: 150.25, Quantity: 10}},
				},
			},
		})
	})
	defer closeServer()

	order, err := client.GetOrder(context.Background(), "HASH_ABC123", 9001)
	if err != nil {
		panic(err)
	}

	printExampleLines(
		"request=GET /accounts/HASH_ABC123/orders/9001",
		"order="+strconv.FormatInt(order.OrderID, 10)+" "+string(order.OrderType)+" "+string(order.Status),
		"activity="+order.OrderActivityCollection[0].ExecutionType,
	)

	// Output:
	// request=GET /accounts/HASH_ABC123/orders/9001
	// order=9001 LIMIT FILLED
	// activity=FILL
}

func ExampleClient_CreateOrder() {
	var request trader.OrderRequest
	client, closeServer := newOrderExampleClient(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			panic(err)
		}
		w.WriteHeader(http.StatusCreated)
	})
	defer closeServer()

	order := &trader.OrderRequest{
		Session:   trader.SessionNormal,
		Duration:  trader.DurationDay,
		OrderType: trader.OrderTypeRequestMarket,
		OrderLegCollection: []trader.OrderLegRequest{
			equityLeg("MSFT", trader.InstructionBuy, 3),
		},
	}
	if err := client.CreateOrder(context.Background(), "HASH_ABC123", order); err != nil {
		panic(err)
	}

	printExampleLines(
		"request=POST /accounts/HASH_ABC123/orders",
		"orderType="+string(request.OrderType)+" duration="+string(request.Duration),
		"leg="+string(request.OrderLegCollection[0].Instruction)+" "+request.OrderLegCollection[0].Instrument.Symbol,
	)

	// Output:
	// request=POST /accounts/HASH_ABC123/orders
	// orderType=MARKET duration=DAY
	// leg=BUY MSFT
}

func ExampleClient_ReplaceOrder() {
	var request trader.OrderRequest
	client, closeServer := newOrderExampleClient(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			panic(err)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer closeServer()

	replacement := &trader.OrderRequest{
		Session:   trader.SessionNormal,
		Duration:  trader.DurationDay,
		OrderType: trader.OrderTypeRequestLimit,
		Price:     199.50,
		OrderLegCollection: []trader.OrderLegRequest{
			equityLeg("MSFT", trader.InstructionBuy, 1),
		},
	}
	if err := client.ReplaceOrder(context.Background(), "HASH_ABC123", 9001, replacement); err != nil {
		panic(err)
	}

	printExampleLines(
		"request=PUT /accounts/HASH_ABC123/orders/9001",
		"orderType="+string(request.OrderType)+" price="+formatExampleFloat(request.Price, 2),
	)

	// Output:
	// request=PUT /accounts/HASH_ABC123/orders/9001
	// orderType=LIMIT price=199.50
}

func ExampleClient_CancelOrder() {
	client, closeServer := newOrderExampleClient(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer closeServer()

	if err := client.CancelOrder(context.Background(), "HASH_ABC123", 9001); err != nil {
		panic(err)
	}

	printExampleLines("request=DELETE /accounts/HASH_ABC123/orders/9001")

	// Output:
	// request=DELETE /accounts/HASH_ABC123/orders/9001
}

func ExampleClient_PreviewOrder() {
	client, closeServer := newOrderExampleClient(func(w http.ResponseWriter, _ *http.Request) {
		writeExampleJSON(w, previewOrderExample())
	})
	defer closeServer()

	preview, err := client.PreviewOrder(context.Background(), "HASH_ABC123", &trader.PreviewOrderRequest{
		OrderStrategy: &trader.OrderStrategy{
			Session:           trader.SessionNormal,
			Duration:          trader.DurationDay,
			OrderType:         trader.OrderTypeMarket,
			OrderStrategyType: trader.OrderStrategyTypeSingle,
			OrderLegs: []trader.PreviewOrderLeg{
				{
					AssetType:   schwab.AssetTypeEquity,
					FinalSymbol: "AAPL",
					Instruction: trader.InstructionBuy,
					Quantity:    1,
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	printExampleLines(
		"request=POST /accounts/HASH_ABC123/previewOrder",
		"preview="+strconv.FormatInt(preview.OrderID, 10)+" "+string(preview.OrderStrategy.AmountIndicator),
		"fee="+preview.CommissionAndFee.Fee.FeeLegs[0].FeeValues[0].Type,
	)

	// Output:
	// request=POST /accounts/HASH_ABC123/previewOrder
	// preview=9001 SHARES
	// fee=SEC_FEE
}

func ExampleOrderListParams() {
	params := trader.OrderListParams{
		MaxResults:      25,
		FromEnteredTime: "2024-01-01",
		ToEnteredTime:   "2024-01-31",
		Status:          trader.OrderStatusFilled,
	}

	printExampleLines(
		"from="+params.FromEnteredTime,
		"to="+params.ToEnteredTime,
		"status="+string(params.Status),
		"maxResults="+strconv.Itoa(params.MaxResults),
	)

	// Output:
	// from=2024-01-01
	// to=2024-01-31
	// status=FILLED
	// maxResults=25
}

func ExampleOrderRequest_equityStopLimit() {
	order := &trader.OrderRequest{
		Session:            trader.SessionNormal,
		Duration:           trader.DurationGoodTillCancel,
		OrderType:          trader.OrderTypeRequestStopLimit,
		Price:              145.00,
		StopPrice:          146.00,
		StopType:           trader.StopTypeStandard,
		TaxLotMethod:       trader.TaxLotMethodFIFO,
		SpecialInstruction: trader.SpecialInstructionAllOrNone,
		OrderStrategyType:  trader.OrderStrategyTypeSingle,
		OrderLegCollection: []trader.OrderLegRequest{
			equityLeg("AAPL", trader.InstructionSell, 10),
		},
	}

	printOrderSummary("sell stop limit", order)

	// Output:
	// sell stop limit
	// orderType=STOP_LIMIT session=NORMAL duration=GOOD_TILL_CANCEL strategy=SINGLE price=145.00 stopPrice=146.00 stopType=STANDARD taxLotMethod=FIFO specialInstruction=ALL_OR_NONE
	// leg 1: EQUITY SELL 10 AAPL
}

func ExampleOrderRequest_trailingStop() {
	order := &trader.OrderRequest{
		Session:            trader.SessionNormal,
		Duration:           trader.DurationDay,
		OrderType:          trader.OrderTypeRequestTrailingStop,
		StopPriceLinkBasis: trader.PriceLinkBasisBid,
		StopPriceLinkType:  trader.PriceLinkTypePercent,
		StopPriceOffset:    5,
		StopType:           trader.StopTypeBid,
		OrderStrategyType:  trader.OrderStrategyTypeSingle,
		OrderLegCollection: []trader.OrderLegRequest{
			equityLeg("MSFT", trader.InstructionSell, 5),
		},
	}

	printOrderSummary("trailing stop", order)

	// Output:
	// trailing stop
	// orderType=TRAILING_STOP session=NORMAL duration=DAY strategy=SINGLE stopLink=BID/PERCENT offset=5.00 stopType=BID
	// leg 1: EQUITY SELL 5 MSFT
}

func ExampleOrderRequest_oco() {
	order := &trader.OrderRequest{
		OrderStrategyType: trader.OrderStrategyTypeOCO,
		ChildOrderStrategies: []*trader.OrderRequest{
			{
				Session:           trader.SessionNormal,
				Duration:          trader.DurationGoodTillCancel,
				OrderType:         trader.OrderTypeRequestLimit,
				Price:             195.00,
				OrderStrategyType: trader.OrderStrategyTypeSingle,
				OrderLegCollection: []trader.OrderLegRequest{
					equityLeg("AAPL", trader.InstructionSell, 10),
				},
			},
			{
				Session:           trader.SessionNormal,
				Duration:          trader.DurationGoodTillCancel,
				OrderType:         trader.OrderTypeRequestStop,
				StopPrice:         170.00,
				StopType:          trader.StopTypeStandard,
				OrderStrategyType: trader.OrderStrategyTypeSingle,
				OrderLegCollection: []trader.OrderLegRequest{
					equityLeg("AAPL", trader.InstructionSell, 10),
				},
			},
		},
	}

	printExampleLines(
		"strategy="+string(order.OrderStrategyType),
		"children="+strconv.Itoa(len(order.ChildOrderStrategies)),
		"child 1="+orderSummary(order.ChildOrderStrategies[0]),
		"child 2="+orderSummary(order.ChildOrderStrategies[1]),
	)

	// Output:
	// strategy=OCO
	// children=2
	// child 1=orderType=LIMIT session=NORMAL duration=GOOD_TILL_CANCEL strategy=SINGLE price=195.00
	// child 2=orderType=STOP session=NORMAL duration=GOOD_TILL_CANCEL strategy=SINGLE stopPrice=170.00 stopType=STANDARD
}

func ExampleOrderRequest_trigger() {
	order := &trader.OrderRequest{
		Session:           trader.SessionNormal,
		Duration:          trader.DurationDay,
		OrderType:         trader.OrderTypeRequestMarket,
		OrderStrategyType: trader.OrderStrategyTypeTrigger,
		OrderLegCollection: []trader.OrderLegRequest{
			equityLeg("AAPL", trader.InstructionBuy, 10),
		},
		ChildOrderStrategies: []*trader.OrderRequest{
			{
				Session:           trader.SessionNormal,
				Duration:          trader.DurationGoodTillCancel,
				OrderType:         trader.OrderTypeRequestStopLimit,
				Price:             165.00,
				StopPrice:         166.00,
				OrderStrategyType: trader.OrderStrategyTypeSingle,
				OrderLegCollection: []trader.OrderLegRequest{
					equityLeg("AAPL", trader.InstructionSell, 10),
				},
			},
		},
	}
	parent := string(order.Session) + " " + string(order.Duration) + " " +
		string(order.OrderType) + " " + string(order.OrderLegCollection[0].Instruction)
	child := string(order.ChildOrderStrategies[0].Session) + " " +
		string(order.ChildOrderStrategies[0].Duration) + " " +
		string(order.ChildOrderStrategies[0].OrderType) +
		" price=" + formatExampleFloat(order.ChildOrderStrategies[0].Price, 2) +
		" stopPrice=" + formatExampleFloat(order.ChildOrderStrategies[0].StopPrice, 2)
	childStrategy := string(order.ChildOrderStrategies[0].OrderStrategyType) +
		" leg=" + string(order.ChildOrderStrategies[0].OrderLegCollection[0].Instruction)

	printExampleLines(
		"strategy="+string(order.OrderStrategyType),
		"parent="+parent,
		"child="+child,
		"child strategy="+childStrategy,
	)

	// Output:
	// strategy=TRIGGER
	// parent=NORMAL DAY MARKET BUY
	// child=NORMAL GOOD_TILL_CANCEL STOP_LIMIT price=165.00 stopPrice=166.00
	// child strategy=SINGLE leg=SELL
}

func ExampleOrderRequest_releaseAndCancelTimes() {
	order := &trader.OrderRequest{
		Session:           trader.SessionSeamless,
		Duration:          trader.DurationDay,
		OrderType:         trader.OrderTypeRequestLimit,
		Price:             100.00,
		ActivationPrice:   101.00,
		CancelTime:        "2024-01-15T20:00:00Z",
		ReleaseTime:       "2024-01-15T14:30:00Z",
		OrderStrategyType: trader.OrderStrategyTypeSingle,
		OrderLegCollection: []trader.OrderLegRequest{
			equityLeg("AAPL", trader.InstructionBuy, 1),
		},
	}
	orderLine := string(order.Duration) + " " + string(order.OrderType) + " " +
		string(order.OrderStrategyType) + " price=" + formatExampleFloat(order.Price, 2)
	legLine := string(order.OrderLegCollection[0].Instruction) + " " +
		order.OrderLegCollection[0].Instrument.Symbol

	printExampleLines(
		"session="+string(order.Session),
		"order="+orderLine,
		"leg="+legLine,
		"activationPrice="+formatExampleFloat(order.ActivationPrice, 2),
		"releaseTime="+order.ReleaseTime,
		"cancelTime="+order.CancelTime,
	)

	// Output:
	// session=SEAMLESS
	// order=DAY LIMIT SINGLE price=100.00
	// leg=BUY AAPL
	// activationPrice=101.00
	// releaseTime=2024-01-15T14:30:00Z
	// cancelTime=2024-01-15T20:00:00Z
}

func ExampleOrderLegRequest_mutualFundExchange() {
	leg := trader.OrderLegRequest{
		OrderLegType: trader.OrderLegTypeMutualFund,
		LegID:        1,
		Instruction:  trader.InstructionExchange,
		Quantity:     2500,
		QuantityType: trader.QuantityTypeDollars,
		DivCapGains:  trader.DivCapGainsReinvest,
		ToSymbol:     "SWPPX",
		Instrument: trader.OrderInstrument{
			AssetType: schwab.AssetTypeMutualFund,
			Symbol:    "SWTSX",
		},
	}

	printExampleLines(
		"legID="+strconv.FormatInt(leg.LegID, 10),
		"orderLegType="+string(leg.OrderLegType),
		"instruction="+string(leg.Instruction),
		"quantity="+formatExampleFloat(leg.Quantity, 0)+" "+string(leg.QuantityType),
		"from="+leg.Instrument.Symbol+" to="+leg.ToSymbol,
		"divCapGains="+string(leg.DivCapGains),
	)

	// Output:
	// legID=1
	// orderLegType=MUTUAL_FUND
	// instruction=EXCHANGE
	// quantity=2500 DOLLARS
	// from=SWTSX to=SWPPX
	// divCapGains=REINVEST
}

func ExampleOrder() {
	order := trader.Order{
		Session:              trader.SessionNormal,
		Duration:             trader.DurationDay,
		OrderType:            trader.OrderTypeLimit,
		Quantity:             10,
		FilledQuantity:       10,
		RemainingQuantity:    0,
		RequestedDestination: trader.RequestedDestinationAuto,
		DestinationLinkName:  "NYSE",
		Price:                150.25,
		Status:               trader.OrderStatusFilled,
		OrderID:              9001,
		EnteredTime:          "2024-01-15T10:30:00Z",
		CloseTime:            "2024-01-15T10:31:00Z",
		AccountNumber:        123456789,
		StatusDescription:    "Order filled",
		OrderLegCollection: []trader.OrderLeg{
			{
				OrderLegType: trader.OrderLegTypeEquity,
				LegID:        1,
				Instruction:  trader.InstructionBuy,
				Quantity:     10,
				Instrument: trader.OrderInstrument{
					AssetType:    schwab.AssetTypeEquity,
					CUSIP:        "037833100",
					Symbol:       "AAPL",
					Description:  "Apple Inc",
					InstrumentID: 1234567,
					Type:         "COMMON_STOCK",
				},
			},
		},
	}
	orderLine := strconv.FormatInt(order.OrderID, 10) + " " +
		string(order.Session) + " " + string(order.Duration) + " " +
		string(order.OrderType) + " " + string(order.Status)
	quantityLine := formatExampleFloat(order.Quantity, 0) +
		" filled=" + formatExampleFloat(order.FilledQuantity, 0) +
		" remaining=" + formatExampleFloat(order.RemainingQuantity, 0)
	legLine := strconv.FormatInt(order.OrderLegCollection[0].LegID, 10) + " " +
		order.OrderLegCollection[0].Instrument.Symbol
	instrumentLine := order.OrderLegCollection[0].Instrument.CUSIP + " " +
		order.OrderLegCollection[0].Instrument.Type

	printExampleLines(
		"order="+orderLine,
		"quantity="+quantityLine,
		"destination="+string(order.RequestedDestination)+" link="+order.DestinationLinkName,
		"price="+formatExampleFloat(order.Price, 2)+" account="+strconv.FormatInt(order.AccountNumber, 10),
		"times="+order.EnteredTime+" "+order.CloseTime,
		"description="+order.StatusDescription,
		"leg="+legLine,
		"instrument="+instrumentLine,
	)

	// Output:
	// order=9001 NORMAL DAY LIMIT FILLED
	// quantity=10 filled=10 remaining=0
	// destination=AUTO link=NYSE
	// price=150.25 account=123456789
	// times=2024-01-15T10:30:00Z 2024-01-15T10:31:00Z
	// description=Order filled
	// leg=1 AAPL
	// instrument=037833100 COMMON_STOCK
}

func ExampleOrder_recursiveStrategies() {
	order := trader.Order{
		OrderID:           1,
		OrderType:         trader.OrderTypeMarket,
		OrderStrategyType: trader.OrderStrategyTypeTrigger,
		ChildOrderStrategies: []*trader.Order{
			{
				OrderID:           2,
				OrderType:         trader.OrderTypeLimit,
				OrderStrategyType: trader.OrderStrategyTypeOCO,
				ChildOrderStrategies: []*trader.Order{
					{OrderID: 3, OrderType: trader.OrderTypeStop},
					{OrderID: 4, OrderType: trader.OrderTypeLimit},
				},
			},
		},
		ReplacingOrderCollection: []*trader.Order{{OrderID: 5, OrderType: trader.OrderTypeMarket}},
	}
	parentLine := strconv.FormatInt(order.OrderID, 10) + " " +
		string(order.OrderType) + " " + string(order.OrderStrategyType)
	childLine := strconv.FormatInt(order.ChildOrderStrategies[0].OrderID, 10) + " " +
		string(order.ChildOrderStrategies[0].OrderType) + " " +
		string(order.ChildOrderStrategies[0].OrderStrategyType)
	replacesLine := strconv.FormatInt(order.ReplacingOrderCollection[0].OrderID, 10) +
		" " + string(order.ReplacingOrderCollection[0].OrderType)

	printExampleLines(
		"parent="+parentLine,
		"child="+childLine,
		"grandchildren="+strconv.Itoa(len(order.ChildOrderStrategies[0].ChildOrderStrategies)),
		"replaces="+replacesLine,
	)

	// Output:
	// parent=1 MARKET TRIGGER
	// child=2 LIMIT OCO
	// grandchildren=2
	// replaces=5 MARKET
}

func ExampleOrderActivity() {
	activity := trader.OrderActivity{
		ActivityType:           "EXECUTION",
		ExecutionType:          "FILL",
		Quantity:               10,
		OrderRemainingQuantity: 0,
		ExecutionLegs: []trader.ExecutionLeg{
			{
				LegID:             1,
				Price:             150.25,
				Quantity:          10,
				MismarkedQuantity: 0,
				InstrumentID:      1234567,
				Time:              "2024-01-15T10:31:00Z",
			},
		},
	}
	quantityLine := formatExampleFloat(activity.Quantity, 0) +
		" remaining=" + formatExampleFloat(activity.OrderRemainingQuantity, 0)
	executionLine := strconv.FormatInt(activity.ExecutionLegs[0].LegID, 10) +
		" price=" + formatExampleFloat(activity.ExecutionLegs[0].Price, 2) +
		" quantity=" + formatExampleFloat(activity.ExecutionLegs[0].Quantity, 0)
	instrumentLine := strconv.FormatInt(activity.ExecutionLegs[0].InstrumentID, 10) +
		" mismarked=" + formatExampleFloat(activity.ExecutionLegs[0].MismarkedQuantity, 0) +
		" time=" + activity.ExecutionLegs[0].Time

	printExampleLines(
		"activity="+activity.ActivityType+" "+activity.ExecutionType,
		"quantity="+quantityLine,
		"execution="+executionLine,
		"instrumentID="+instrumentLine,
	)

	// Output:
	// activity=EXECUTION FILL
	// quantity=10 remaining=0
	// execution=1 price=150.25 quantity=10
	// instrumentID=1234567 mismarked=0 time=2024-01-15T10:31:00Z
}

func ExamplePreviewOrderRequest() {
	request := &trader.PreviewOrderRequest{
		OrderID: 9001,
		OrderStrategy: &trader.OrderStrategy{
			AdvancedOrderType:     trader.AdvancedOrderTypeNone,
			OrderStrategyType:     trader.OrderStrategyTypeSingle,
			Session:               trader.SessionNormal,
			Duration:              trader.DurationDay,
			OrderType:             trader.OrderTypeMarket,
			Quantity:              1,
			SettlementInstruction: trader.SettlementInstructionRegular,
			AmountIndicator:       trader.AmountIndicatorShares,
			Strategy:              trader.ComplexOrderStrategyTypeNone,
		},
	}

	printExampleLines(
		"orderID="+strconv.FormatInt(request.OrderID, 10),
		"advancedOrderType="+string(request.OrderStrategy.AdvancedOrderType),
		"settlement="+string(request.OrderStrategy.SettlementInstruction),
		"amountIndicator="+string(request.OrderStrategy.AmountIndicator),
	)

	// Output:
	// orderID=9001
	// advancedOrderType=NONE
	// settlement=REGULAR
	// amountIndicator=SHARES
}

func ExamplePreviewOrder() {
	preview := previewOrderExample()

	printExampleLines(
		"orderID="+strconv.FormatInt(preview.OrderID, 10),
		"strategy="+string(preview.OrderStrategy.Strategy),
		"validation="+preview.OrderValidationResult.Alerts[0].ValidationRuleName,
		"commission="+preview.CommissionAndFee.Commission.CommissionLegs[0].CommissionValues[0].Type,
		"fee="+preview.CommissionAndFee.Fee.FeeLegs[0].FeeValues[0].Type,
	)

	// Output:
	// orderID=9001
	// strategy=NONE
	// validation=MarketHoursRule
	// commission=COMMISSION
	// fee=SEC_FEE
}

func ExampleOrderLeg() {
	leg := trader.OrderLeg{
		OrderLegType: trader.OrderLegTypeEquity,
		LegID:        1,
		Instruction:  trader.InstructionBuy,
		Quantity:     10,
		QuantityType: trader.QuantityTypeShares,
		Instrument:   equityInstrument("AAPL"),
	}
	instructionLine := string(leg.Instruction) +
		" quantity=" + formatExampleFloat(leg.Quantity, 0) + " " + string(leg.QuantityType)

	printExampleLines(
		"leg="+strconv.FormatInt(leg.LegID, 10)+" "+string(leg.OrderLegType),
		"instruction="+instructionLine,
		"symbol="+leg.Instrument.Symbol,
	)

	// Output:
	// leg=1 EQUITY
	// instruction=BUY quantity=10 SHARES
	// symbol=AAPL
}

func ExampleOrderInstrument() {
	instrument := equityInstrument("AAPL")

	printExampleLines(
		"assetType="+string(instrument.AssetType),
		"symbol="+instrument.Symbol,
		"cusip="+instrument.CUSIP,
		"type="+instrument.Type,
	)

	// Output:
	// assetType=EQUITY
	// symbol=AAPL
	// cusip=037833100
	// type=COMMON_STOCK
}

func ExampleExecutionLeg() {
	execution := trader.ExecutionLeg{
		LegID:             1,
		Price:             150.25,
		Quantity:          10,
		MismarkedQuantity: 0,
		InstrumentID:      1234567,
		Time:              "2024-01-15T10:31:00Z",
	}

	printExampleLines(
		"leg="+strconv.FormatInt(execution.LegID, 10),
		"price="+formatExampleFloat(execution.Price, 2)+" quantity="+formatExampleFloat(execution.Quantity, 0),
		"mismarked="+formatExampleFloat(execution.MismarkedQuantity, 0)+" time="+execution.Time,
		"instrumentID="+strconv.FormatInt(execution.InstrumentID, 10),
	)

	// Output:
	// leg=1
	// price=150.25 quantity=10
	// mismarked=0 time=2024-01-15T10:31:00Z
	// instrumentID=1234567
}

func ExampleOrderStrategy() {
	strategy := previewOrderExample().OrderStrategy

	printExampleLines(
		"account="+strategy.AccountNumber,
		"advanced="+string(strategy.AdvancedOrderType)+" status="+string(strategy.Status),
		"order="+string(strategy.Session)+" "+string(strategy.Duration)+" "+string(strategy.OrderType),
		"amount="+string(strategy.AmountIndicator)+" settlement="+string(strategy.SettlementInstruction),
	)

	// Output:
	// account=123456789
	// advanced=NONE status=ACCEPTED
	// order=NORMAL DAY MARKET
	// amount=SHARES settlement=REGULAR
}

func ExampleOrderBalance() {
	balance := trader.OrderBalance{
		OrderValue:             150.25,
		ProjectedAvailableFund: 1000,
		ProjectedBuyingPower:   1000,
		ProjectedCommission:    0,
	}

	printExampleLines(
		"orderValue="+formatExampleFloat(balance.OrderValue, 2),
		"available="+formatExampleFloat(balance.ProjectedAvailableFund, 2),
		"buyingPower="+formatExampleFloat(balance.ProjectedBuyingPower, 2),
		"commission="+formatExampleFloat(balance.ProjectedCommission, 2),
	)

	// Output:
	// orderValue=150.25
	// available=1000.00
	// buyingPower=1000.00
	// commission=0.00
}

func ExamplePreviewOrderLeg() {
	leg := previewOrderExample().OrderStrategy.OrderLegs[0]
	prices := formatExampleFloat(leg.BidPrice, 2) +
		" ask=" + formatExampleFloat(leg.AskPrice, 2) +
		" mark=" + formatExampleFloat(leg.MarkPrice, 2)

	printExampleLines(
		"leg="+formatExampleFloat(leg.LegID, 0)+" "+string(leg.AssetType),
		"symbol="+leg.FinalSymbol+" instruction="+string(leg.Instruction),
		"bid="+prices,
	)

	// Output:
	// leg=1 EQUITY
	// symbol=AAPL instruction=BUY
	// bid=150.20 ask=150.30 mark=150.25
}

func ExampleOrderValidationResult() {
	result := previewOrderExample().OrderValidationResult

	printExampleLines(
		"alerts="+strconv.Itoa(len(result.Alerts)),
		"accepts="+strconv.Itoa(len(result.Accepts)),
		"rejects="+strconv.Itoa(len(result.Rejects)),
		"reviews="+strconv.Itoa(len(result.Reviews)),
		"warns="+strconv.Itoa(len(result.Warns)),
	)

	// Output:
	// alerts=1
	// accepts=1
	// rejects=1
	// reviews=1
	// warns=1
}

func ExampleOrderValidationDetail() {
	detail := trader.OrderValidationDetail{
		ValidationRuleName: "MarketHoursRule",
		Message:            "Market is currently closed",
		ActivityMessage:    "Review market hours before placing this order",
		OriginalSeverity:   "INFO",
		OverrideName:       "NONE",
		OverrideSeverity:   "NONE",
	}

	printExampleLines(
		"rule="+detail.ValidationRuleName,
		"message="+detail.Message,
		"activity="+detail.ActivityMessage,
		"severity="+detail.OriginalSeverity+" override="+detail.OverrideName+":"+detail.OverrideSeverity,
	)

	// Output:
	// rule=MarketHoursRule
	// message=Market is currently closed
	// activity=Review market hours before placing this order
	// severity=INFO override=NONE:NONE
}

func ExampleCommissionAndFee() {
	commissionAndFee := previewOrderExample().CommissionAndFee

	printExampleLines(
		"commission="+commissionAndFee.Commission.CommissionLegs[0].CommissionValues[0].Type,
		"fee="+commissionAndFee.Fee.FeeLegs[0].FeeValues[0].Type,
		"trueCommission="+commissionAndFee.TrueCommission.CommissionLegs[0].CommissionValues[0].Type,
	)

	// Output:
	// commission=COMMISSION
	// fee=SEC_FEE
	// trueCommission=TRUE_COMMISSION
}

func ExampleCommission() {
	commission := previewOrderExample().CommissionAndFee.Commission

	printExampleLines("commissionLegs=" + strconv.Itoa(len(commission.CommissionLegs)))

	// Output:
	// commissionLegs=1
}

func ExampleCommissionLeg() {
	leg := previewOrderExample().CommissionAndFee.Commission.CommissionLegs[0]

	printExampleLines("commissionValues=" + strconv.Itoa(len(leg.CommissionValues)))

	// Output:
	// commissionValues=1
}

func ExampleCommissionValue() {
	value := previewOrderExample().CommissionAndFee.Commission.CommissionLegs[0].CommissionValues[0]

	printExampleLines("commission=" + value.Type + " value=" + formatExampleFloat(value.Value, 2))

	// Output:
	// commission=COMMISSION value=0.00
}

func ExampleFees() {
	fees := previewOrderExample().CommissionAndFee.Fee

	printExampleLines("feeLegs=" + strconv.Itoa(len(fees.FeeLegs)))

	// Output:
	// feeLegs=1
}

func ExampleFeeLeg() {
	leg := previewOrderExample().CommissionAndFee.Fee.FeeLegs[0]

	printExampleLines("feeValues=" + strconv.Itoa(len(leg.FeeValues)))

	// Output:
	// feeValues=1
}

func ExampleFeeValue() {
	value := previewOrderExample().CommissionAndFee.Fee.FeeLegs[0].FeeValues[0]

	printExampleLines("fee=" + value.Type + " value=" + formatExampleFloat(value.Value, 2))

	// Output:
	// fee=SEC_FEE value=0.02
}

func ExampleOrderType() {
	printCatalog("order types", []string{
		string(trader.OrderTypeMarket),
		string(trader.OrderTypeLimit),
		string(trader.OrderTypeStop),
		string(trader.OrderTypeStopLimit),
		string(trader.OrderTypeTrailingStop),
		string(trader.OrderTypeCabinet),
		string(trader.OrderTypeNonMarketable),
		string(trader.OrderTypeMarketOnClose),
		string(trader.OrderTypeExercise),
		string(trader.OrderTypeTrailingStopLimit),
		string(trader.OrderTypeNetDebit),
		string(trader.OrderTypeNetCredit),
		string(trader.OrderTypeNetZero),
		string(trader.OrderTypeLimitOnClose),
		string(trader.OrderTypeUnknown),
	})

	// Output:
	// order types: MARKET, LIMIT, STOP, STOP_LIMIT, TRAILING_STOP, CABINET, NON_MARKETABLE, MARKET_ON_CLOSE, EXERCISE, TRAILING_STOP_LIMIT, NET_DEBIT, NET_CREDIT, NET_ZERO, LIMIT_ON_CLOSE, UNKNOWN
}

func ExampleOrderTypeRequest() {
	printCatalog("request order types", []string{
		string(trader.OrderTypeRequestMarket),
		string(trader.OrderTypeRequestLimit),
		string(trader.OrderTypeRequestStop),
		string(trader.OrderTypeRequestStopLimit),
		string(trader.OrderTypeRequestTrailingStop),
		string(trader.OrderTypeRequestCabinet),
		string(trader.OrderTypeRequestNonMarketable),
		string(trader.OrderTypeRequestMarketOnClose),
		string(trader.OrderTypeRequestExercise),
		string(trader.OrderTypeRequestTrailingStopLimit),
		string(trader.OrderTypeRequestNetDebit),
		string(trader.OrderTypeRequestNetCredit),
		string(trader.OrderTypeRequestNetZero),
		string(trader.OrderTypeRequestLimitOnClose),
	})

	// Output:
	// request order types: MARKET, LIMIT, STOP, STOP_LIMIT, TRAILING_STOP, CABINET, NON_MARKETABLE, MARKET_ON_CLOSE, EXERCISE, TRAILING_STOP_LIMIT, NET_DEBIT, NET_CREDIT, NET_ZERO, LIMIT_ON_CLOSE
}

func ExampleComplexOrderStrategyType() {
	printCatalog("complex order strategy types", []string{
		string(trader.ComplexOrderStrategyTypeNone),
		string(trader.ComplexOrderStrategyTypeCovered),
		string(trader.ComplexOrderStrategyTypeVertical),
		string(trader.ComplexOrderStrategyTypeBackRatio),
		string(trader.ComplexOrderStrategyTypeCalendar),
		string(trader.ComplexOrderStrategyTypeDiagonal),
		string(trader.ComplexOrderStrategyTypeStraddle),
		string(trader.ComplexOrderStrategyTypeStrangle),
		string(trader.ComplexOrderStrategyTypeCollarSynthetic),
		string(trader.ComplexOrderStrategyTypeButterfly),
		string(trader.ComplexOrderStrategyTypeCondor),
		string(trader.ComplexOrderStrategyTypeIronCondor),
		string(trader.ComplexOrderStrategyTypeVerticalRoll),
		string(trader.ComplexOrderStrategyTypeCollarWithStock),
		string(trader.ComplexOrderStrategyTypeDoubleDiagonal),
		string(trader.ComplexOrderStrategyTypeUnbalancedButterfly),
		string(trader.ComplexOrderStrategyTypeUnbalancedCondor),
		string(trader.ComplexOrderStrategyTypeUnbalancedIronCondor),
		string(trader.ComplexOrderStrategyTypeUnbalancedVerticalRoll),
		string(trader.ComplexOrderStrategyTypeMutualFundSwap),
		string(trader.ComplexOrderStrategyTypeCustom),
	})

	// Output:
	// complex order strategy types: NONE, COVERED, VERTICAL, BACK_RATIO, CALENDAR, DIAGONAL, STRADDLE, STRANGLE, COLLAR_SYNTHETIC, BUTTERFLY, CONDOR, IRON_CONDOR, VERTICAL_ROLL, COLLAR_WITH_STOCK, DOUBLE_DIAGONAL, UNBALANCED_BUTTERFLY, UNBALANCED_CONDOR, UNBALANCED_IRON_CONDOR, UNBALANCED_VERTICAL_ROLL, MUTUAL_FUND_SWAP, CUSTOM
}

func ExampleSession() {
	printCatalog("sessions", []string{
		string(trader.SessionNormal),
		string(trader.SessionAM),
		string(trader.SessionPM),
		string(trader.SessionSeamless),
	})

	// Output:
	// sessions: NORMAL, AM, PM, SEAMLESS
}

func ExampleDuration() {
	printCatalog("durations", []string{
		string(trader.DurationDay),
		string(trader.DurationGoodTillCancel),
		string(trader.DurationFillOrKill),
		string(trader.DurationImmediateOrCancel),
		string(trader.DurationEndOfWeek),
		string(trader.DurationEndOfMonth),
		string(trader.DurationNextEndOfMonth),
		string(trader.DurationUnknown),
	})

	// Output:
	// durations: DAY, GOOD_TILL_CANCEL, FILL_OR_KILL, IMMEDIATE_OR_CANCEL, END_OF_WEEK, END_OF_MONTH, NEXT_END_OF_MONTH, UNKNOWN
}

func ExampleRequestedDestination() {
	printCatalog("requested destinations", []string{
		string(trader.RequestedDestinationINET),
		string(trader.RequestedDestinationECNARCA),
		string(trader.RequestedDestinationCBOE),
		string(trader.RequestedDestinationAMEX),
		string(trader.RequestedDestinationPHLX),
		string(trader.RequestedDestinationISE),
		string(trader.RequestedDestinationBOX),
		string(trader.RequestedDestinationNYSE),
		string(trader.RequestedDestinationNASDAQ),
		string(trader.RequestedDestinationBATS),
		string(trader.RequestedDestinationC2),
		string(trader.RequestedDestinationAuto),
	})

	// Output:
	// requested destinations: INET, ECN_ARCA, CBOE, AMEX, PHLX, ISE, BOX, NYSE, NASDAQ, BATS, C2, AUTO
}

func ExamplePriceLinkBasis() {
	printCatalog("price link bases", []string{
		string(trader.PriceLinkBasisManual),
		string(trader.PriceLinkBasisBase),
		string(trader.PriceLinkBasisTrigger),
		string(trader.PriceLinkBasisLast),
		string(trader.PriceLinkBasisBid),
		string(trader.PriceLinkBasisAsk),
		string(trader.PriceLinkBasisAskBid),
		string(trader.PriceLinkBasisMark),
		string(trader.PriceLinkBasisAverage),
	})

	// Output:
	// price link bases: MANUAL, BASE, TRIGGER, LAST, BID, ASK, ASK_BID, MARK, AVERAGE
}

func ExamplePriceLinkType() {
	printCatalog("price link types", []string{
		string(trader.PriceLinkTypeValue),
		string(trader.PriceLinkTypePercent),
		string(trader.PriceLinkTypeTick),
	})

	// Output:
	// price link types: VALUE, PERCENT, TICK
}

func ExampleStopType() {
	printCatalog("stop types", []string{
		string(trader.StopTypeStandard),
		string(trader.StopTypeBid),
		string(trader.StopTypeAsk),
		string(trader.StopTypeLast),
		string(trader.StopTypeMark),
	})

	// Output:
	// stop types: STANDARD, BID, ASK, LAST, MARK
}

func ExampleTaxLotMethod() {
	printCatalog("tax lot methods", []string{
		string(trader.TaxLotMethodFIFO),
		string(trader.TaxLotMethodLIFO),
		string(trader.TaxLotMethodHighCost),
		string(trader.TaxLotMethodLowCost),
		string(trader.TaxLotMethodAverageCost),
		string(trader.TaxLotMethodSpecificLot),
		string(trader.TaxLotMethodLossHarvester),
	})

	// Output:
	// tax lot methods: FIFO, LIFO, HIGH_COST, LOW_COST, AVERAGE_COST, SPECIFIC_LOT, LOSS_HARVESTER
}

func ExampleSpecialInstruction() {
	printCatalog("special instructions", []string{
		string(trader.SpecialInstructionAllOrNone),
		string(trader.SpecialInstructionDoNotReduce),
		string(trader.SpecialInstructionAllOrNoneDoNotReduce),
	})

	// Output:
	// special instructions: ALL_OR_NONE, DO_NOT_REDUCE, ALL_OR_NONE_DO_NOT_REDUCE
}

func ExampleOrderStrategyType() {
	printCatalog("order strategy types", []string{
		string(trader.OrderStrategyTypeSingle),
		string(trader.OrderStrategyTypeCancel),
		string(trader.OrderStrategyTypeRecall),
		string(trader.OrderStrategyTypePair),
		string(trader.OrderStrategyTypeFlatten),
		string(trader.OrderStrategyTypeTwoDaySwap),
		string(trader.OrderStrategyTypeBlastAll),
		string(trader.OrderStrategyTypeOCO),
		string(trader.OrderStrategyTypeTrigger),
	})

	// Output:
	// order strategy types: SINGLE, CANCEL, RECALL, PAIR, FLATTEN, TWO_DAY_SWAP, BLAST_ALL, OCO, TRIGGER
}

func ExampleOrderStatus() {
	printCatalog("order statuses", []string{
		string(trader.OrderStatusAwaitingParentOrder),
		string(trader.OrderStatusAwaitingCondition),
		string(trader.OrderStatusAwaitingStopCondition),
		string(trader.OrderStatusAwaitingManualReview),
		string(trader.OrderStatusAccepted),
		string(trader.OrderStatusAwaitingUROut),
		string(trader.OrderStatusPendingActivation),
		string(trader.OrderStatusQueued),
		string(trader.OrderStatusWorking),
		string(trader.OrderStatusRejected),
		string(trader.OrderStatusPendingCancel),
		string(trader.OrderStatusCanceled),
		string(trader.OrderStatusPendingReplace),
		string(trader.OrderStatusReplaced),
		string(trader.OrderStatusFilled),
		string(trader.OrderStatusExpired),
		string(trader.OrderStatusNew),
		string(trader.OrderStatusAwaitingReleaseTime),
		string(trader.OrderStatusPendingAcknowledgment),
		string(trader.OrderStatusPendingRecall),
		string(trader.OrderStatusUnknown),
	})

	// Output:
	// order statuses: AWAITING_PARENT_ORDER, AWAITING_CONDITION, AWAITING_STOP_CONDITION, AWAITING_MANUAL_REVIEW, ACCEPTED, AWAITING_UR_OUT, PENDING_ACTIVATION, QUEUED, WORKING, REJECTED, PENDING_CANCEL, CANCELED, PENDING_REPLACE, REPLACED, FILLED, EXPIRED, NEW, AWAITING_RELEASE_TIME, PENDING_ACKNOWLEDGEMENT, PENDING_RECALL, UNKNOWN
}

func ExampleInstruction() {
	printCatalog("instructions", []string{
		string(trader.InstructionBuy),
		string(trader.InstructionSell),
		string(trader.InstructionBuyToCover),
		string(trader.InstructionSellShort),
		string(trader.InstructionBuyToOpen),
		string(trader.InstructionBuyToClose),
		string(trader.InstructionSellToOpen),
		string(trader.InstructionSellToClose),
		string(trader.InstructionExchange),
		string(trader.InstructionSellShortExempt),
	})

	// Output:
	// instructions: BUY, SELL, BUY_TO_COVER, SELL_SHORT, BUY_TO_OPEN, BUY_TO_CLOSE, SELL_TO_OPEN, SELL_TO_CLOSE, EXCHANGE, SELL_SHORT_EXEMPT
}

func ExampleOrderLegType() {
	printCatalog("order leg types", []string{
		string(trader.OrderLegTypeEquity),
		string(trader.OrderLegTypeOption),
		string(trader.OrderLegTypeIndex),
		string(trader.OrderLegTypeMutualFund),
		string(trader.OrderLegTypeCashEquivalent),
		string(trader.OrderLegTypeFixedIncome),
		string(trader.OrderLegTypeCurrency),
		string(trader.OrderLegTypeCollectiveInvestment),
	})

	// Output:
	// order leg types: EQUITY, OPTION, INDEX, MUTUAL_FUND, CASH_EQUIVALENT, FIXED_INCOME, CURRENCY, COLLECTIVE_INVESTMENT
}

func ExamplePositionEffect() {
	printCatalog("position effects", []string{
		string(trader.PositionEffectOpening),
		string(trader.PositionEffectClosing),
		string(trader.PositionEffectAutomatic),
	})

	// Output:
	// position effects: OPENING, CLOSING, AUTOMATIC
}

func ExampleQuantityType() {
	printCatalog("quantity types", []string{
		string(trader.QuantityTypeAllShares),
		string(trader.QuantityTypeDollars),
		string(trader.QuantityTypeShares),
	})

	// Output:
	// quantity types: ALL_SHARES, DOLLARS, SHARES
}

func ExampleDivCapGains() {
	printCatalog("dividend and capital gains", []string{
		string(trader.DivCapGainsReinvest),
		string(trader.DivCapGainsPayout),
	})

	// Output:
	// dividend and capital gains: REINVEST, PAYOUT
}

func ExampleAmountIndicator() {
	printCatalog("amount indicators", []string{
		string(trader.AmountIndicatorDollars),
		string(trader.AmountIndicatorShares),
		string(trader.AmountIndicatorAllShares),
		string(trader.AmountIndicatorPercentage),
		string(trader.AmountIndicatorUnknown),
	})

	// Output:
	// amount indicators: DOLLARS, SHARES, ALL_SHARES, PERCENTAGE, UNKNOWN
}

func ExampleSettlementInstruction() {
	printCatalog("settlement instructions", []string{
		string(trader.SettlementInstructionRegular),
		string(trader.SettlementInstructionCash),
		string(trader.SettlementInstructionNextDay),
		string(trader.SettlementInstructionUnknown),
	})

	// Output:
	// settlement instructions: REGULAR, CASH, NEXT_DAY, UNKNOWN
}

func ExampleAdvancedOrderType() {
	printCatalog("advanced order types", []string{
		string(trader.AdvancedOrderTypeNone),
		string(trader.AdvancedOrderTypeOTO),
		string(trader.AdvancedOrderTypeOCO),
		string(trader.AdvancedOrderTypeOTOCO),
		string(trader.AdvancedOrderTypeOT2OCO),
		string(trader.AdvancedOrderTypeOT3OCO),
		string(trader.AdvancedOrderTypeBlastAll),
		string(trader.AdvancedOrderTypeOTA),
		string(trader.AdvancedOrderTypePair),
	})

	// Output:
	// advanced order types: NONE, OTO, OCO, OTOCO, OT2OCO, OT3OCO, BLAST_ALL, OTA, PAIR
}

func newOrderExampleClient(handler http.HandlerFunc) (*trader.Client, func()) {
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	})
	server := httptest.NewServer(http.StripPrefix("/trader/v1", wrapped))

	client := trader.NewClient(
		schwab.WithToken("example-token"),
		schwab.WithHTTPClient(server.Client()),
		schwab.WithBaseURL(server.URL),
	)
	return client, server.Close
}

func equityLeg(symbol string, instruction trader.Instruction, quantity float64) trader.OrderLegRequest {
	return trader.OrderLegRequest{
		OrderLegType: trader.OrderLegTypeEquity,
		Instruction:  instruction,
		Quantity:     quantity,
		Instrument: trader.OrderInstrument{
			AssetType: schwab.AssetTypeEquity,
			Symbol:    symbol,
		},
	}
}

func equityInstrument(symbol string) trader.OrderInstrument {
	return trader.OrderInstrument{
		AssetType:    schwab.AssetTypeEquity,
		CUSIP:        "037833100",
		Symbol:       symbol,
		Description:  "Apple Inc",
		InstrumentID: 1234567,
		Type:         "COMMON_STOCK",
	}
}

func previewOrderExample() trader.PreviewOrder {
	return trader.PreviewOrder{
		OrderID: 9001,
		OrderStrategy: &trader.OrderStrategy{
			AccountNumber:          "123456789",
			AdvancedOrderType:      trader.AdvancedOrderTypeNone,
			OrderStrategyType:      trader.OrderStrategyTypeSingle,
			OrderVersion:           1,
			Session:                trader.SessionNormal,
			Status:                 trader.OrderStatusAccepted,
			AllOrNone:              false,
			Discretionary:          false,
			Duration:               trader.DurationDay,
			OrderType:              trader.OrderTypeMarket,
			OrderValue:             150.25,
			Quantity:               1,
			RemainingQuantity:      1,
			SellNonMarginableFirst: false,
			SettlementInstruction:  trader.SettlementInstructionRegular,
			AmountIndicator:        trader.AmountIndicatorShares,
			Strategy:               trader.ComplexOrderStrategyTypeNone,
			OrderBalance: &trader.OrderBalance{
				OrderValue:             150.25,
				ProjectedAvailableFund: 1000,
				ProjectedBuyingPower:   1000,
				ProjectedCommission:    0,
			},
			OrderLegs: []trader.PreviewOrderLeg{
				{
					AskPrice:            150.30,
					AssetType:           schwab.AssetTypeEquity,
					BidPrice:            150.20,
					FinalSymbol:         "AAPL",
					Instruction:         trader.InstructionBuy,
					LastPrice:           150.25,
					LegID:               1,
					MarkPrice:           150.25,
					ProjectedCommission: 0,
					Quantity:            1,
				},
			},
		},
		OrderValidationResult: &trader.OrderValidationResult{
			Alerts: []trader.OrderValidationDetail{
				{
					ValidationRuleName: "MarketHoursRule",
					Message:            "Market is currently closed",
					OriginalSeverity:   "INFO",
				},
			},
			Accepts: []trader.OrderValidationDetail{
				{
					ValidationRuleName: "BuyingPowerRule",
					Message:            "Buying power accepted",
					OriginalSeverity:   "ACCEPT",
				},
			},
			Rejects: []trader.OrderValidationDetail{
				{
					ValidationRuleName: "ExampleRejectRule",
					ActivityMessage:    "No reject in this example",
					OverrideName:       "NONE",
				},
			},
			Reviews: []trader.OrderValidationDetail{
				{ValidationRuleName: "ReviewRule", OverrideSeverity: "REVIEW"},
			},
			Warns: []trader.OrderValidationDetail{
				{ValidationRuleName: "WarningRule", Message: "Example warning"},
			},
		},
		CommissionAndFee: &trader.CommissionAndFee{
			Commission: &trader.Commission{
				CommissionLegs: []trader.CommissionLeg{
					{CommissionValues: []trader.CommissionValue{{Value: 0, Type: "COMMISSION"}}},
				},
			},
			Fee: &trader.Fees{
				FeeLegs: []trader.FeeLeg{
					{FeeValues: []trader.FeeValue{{Value: 0.02, Type: "SEC_FEE"}}},
				},
			},
			TrueCommission: &trader.Commission{
				CommissionLegs: []trader.CommissionLeg{
					{CommissionValues: []trader.CommissionValue{{Value: 0, Type: "TRUE_COMMISSION"}}},
				},
			},
		},
	}
}

func writeExampleJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		panic(err)
	}
}

func printOrderSummary(label string, order *trader.OrderRequest) {
	if _, err := json.Marshal(order); err != nil {
		panic(err)
	}

	lines := []string{label, orderSummary(order)}
	for i, leg := range order.OrderLegCollection {
		lines = append(lines, legSummary(i+1, leg))
	}
	printExampleLines(lines...)
}

func orderSummary(order *trader.OrderRequest) string {
	var output strings.Builder
	output.WriteString("orderType=")
	output.WriteString(string(order.OrderType))
	if order.Session != "" {
		output.WriteString(" session=")
		output.WriteString(string(order.Session))
	}
	if order.Duration != "" {
		output.WriteString(" duration=")
		output.WriteString(string(order.Duration))
	}
	if order.OrderStrategyType != "" {
		output.WriteString(" strategy=")
		output.WriteString(string(order.OrderStrategyType))
	}
	if order.Price != 0 {
		output.WriteString(" price=")
		output.WriteString(formatExampleFloat(order.Price, 2))
	}
	if order.StopPrice != 0 {
		output.WriteString(" stopPrice=")
		output.WriteString(formatExampleFloat(order.StopPrice, 2))
	}
	if order.StopPriceLinkBasis != "" || order.StopPriceLinkType != "" {
		output.WriteString(" stopLink=")
		output.WriteString(string(order.StopPriceLinkBasis))
		output.WriteByte('/')
		output.WriteString(string(order.StopPriceLinkType))
	}
	if order.StopPriceOffset != 0 {
		output.WriteString(" offset=")
		output.WriteString(formatExampleFloat(order.StopPriceOffset, 2))
	}
	if order.StopType != "" {
		output.WriteString(" stopType=")
		output.WriteString(string(order.StopType))
	}
	if order.TaxLotMethod != "" {
		output.WriteString(" taxLotMethod=")
		output.WriteString(string(order.TaxLotMethod))
	}
	if order.SpecialInstruction != "" {
		output.WriteString(" specialInstruction=")
		output.WriteString(string(order.SpecialInstruction))
	}
	return output.String()
}

func legSummary(position int, leg trader.OrderLegRequest) string {
	var output strings.Builder
	output.WriteString("leg ")
	output.WriteString(strconv.Itoa(position))
	output.WriteString(": ")
	output.WriteString(string(leg.OrderLegType))
	output.WriteByte(' ')
	output.WriteString(string(leg.Instruction))
	output.WriteByte(' ')
	output.WriteString(formatExampleFloat(leg.Quantity, 0))
	output.WriteByte(' ')
	output.WriteString(leg.Instrument.Symbol)
	return output.String()
}

func printCatalog(label string, values []string) {
	printExampleLines(label + ": " + strings.Join(values, ", "))
}

func printExampleLines(lines ...string) {
	var output strings.Builder
	for _, line := range lines {
		output.WriteString(line)
		output.WriteByte('\n')
	}
	if _, err := os.Stdout.WriteString(output.String()); err != nil {
		panic(err)
	}
}

func formatExampleFloat(value float64, precision int) string {
	return strconv.FormatFloat(value, 'f', precision, 64)
}
