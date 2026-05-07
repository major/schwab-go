package trader_test

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	schwab "github.com/major/schwab-go/schwab"
	"github.com/major/schwab-go/schwab/trader"
)

func ExampleOrderRequest_singleLegOption() {
	order := optionOrder(
		trader.OrderTypeRequestLimit,
		trader.ComplexOrderStrategyTypeNone,
		0,
		1.45,
		optionLeg("AAPL  240621C00200000", trader.InstructionBuyToOpen, trader.PositionEffectOpening, 1),
	)

	printOptionOrder("buy to open call", order)

	// Output:
	// buy to open call
	// orderType=LIMIT complexOrderStrategyType=NONE orderStrategyType=SINGLE price=1.45
	// leg 1: BUY_TO_OPEN OPENING 1 AAPL  240621C00200000
}

func ExampleOrderRequest_singleLegOptionMarket() {
	order := optionOrder(
		trader.OrderTypeRequestMarket,
		trader.ComplexOrderStrategyTypeNone,
		0,
		0,
		optionLeg("AAPL  240621C00200000", trader.InstructionBuyToOpen, trader.PositionEffectOpening, 1),
	)

	printOptionOrder("buy to open call at market", order)

	// Output:
	// buy to open call at market
	// orderType=MARKET complexOrderStrategyType=NONE orderStrategyType=SINGLE
	// leg 1: BUY_TO_OPEN OPENING 1 AAPL  240621C00200000
}

func ExampleOrderRequest_singleLegOptionShort() {
	openOrder := optionOrder(
		trader.OrderTypeRequestLimit,
		trader.ComplexOrderStrategyTypeNone,
		0,
		2.35,
		optionLeg("MSFT  240621P00320000", trader.InstructionSellToOpen, trader.PositionEffectOpening, 1),
	)
	closeOrder := optionOrder(
		trader.OrderTypeRequestLimit,
		trader.ComplexOrderStrategyTypeNone,
		0,
		1.10,
		optionLeg("MSFT  240621P00320000", trader.InstructionBuyToClose, trader.PositionEffectClosing, 1),
	)

	printOptionOrder("sell to open put", openOrder)
	printOptionOrder("buy to close put", closeOrder)

	// Output:
	// sell to open put
	// orderType=LIMIT complexOrderStrategyType=NONE orderStrategyType=SINGLE price=2.35
	// leg 1: SELL_TO_OPEN OPENING 1 MSFT  240621P00320000
	// buy to close put
	// orderType=LIMIT complexOrderStrategyType=NONE orderStrategyType=SINGLE price=1.10
	// leg 1: BUY_TO_CLOSE CLOSING 1 MSFT  240621P00320000
}

func ExampleOrderRequest_verticalBullCallSpread() {
	openOrder := optionOrder(
		trader.OrderTypeRequestNetDebit,
		trader.ComplexOrderStrategyTypeVertical,
		1,
		1.25,
		optionLeg("AAPL  240621C00195000", trader.InstructionBuyToOpen, trader.PositionEffectOpening, 1),
		optionLeg("AAPL  240621C00200000", trader.InstructionSellToOpen, trader.PositionEffectOpening, 1),
	)
	closeOrder := optionOrder(
		trader.OrderTypeRequestNetCredit,
		trader.ComplexOrderStrategyTypeVertical,
		1,
		1.80,
		optionLeg("AAPL  240621C00195000", trader.InstructionSellToClose, trader.PositionEffectClosing, 1),
		optionLeg("AAPL  240621C00200000", trader.InstructionBuyToClose, trader.PositionEffectClosing, 1),
	)

	printOptionOrder("open bull call spread", openOrder)
	printOptionOrder("close bull call spread", closeOrder)

	// Output:
	// open bull call spread
	// orderType=NET_DEBIT complexOrderStrategyType=VERTICAL orderStrategyType=SINGLE quantity=1 price=1.25
	// leg 1: BUY_TO_OPEN OPENING 1 AAPL  240621C00195000
	// leg 2: SELL_TO_OPEN OPENING 1 AAPL  240621C00200000
	// close bull call spread
	// orderType=NET_CREDIT complexOrderStrategyType=VERTICAL orderStrategyType=SINGLE quantity=1 price=1.80
	// leg 1: SELL_TO_CLOSE CLOSING 1 AAPL  240621C00195000
	// leg 2: BUY_TO_CLOSE CLOSING 1 AAPL  240621C00200000
}

func ExampleOrderRequest_verticalBearCallSpread() {
	openOrder := optionOrder(
		trader.OrderTypeRequestNetCredit,
		trader.ComplexOrderStrategyTypeVertical,
		1,
		1.15,
		optionLeg("MSFT  240621C00320000", trader.InstructionSellToOpen, trader.PositionEffectOpening, 1),
		optionLeg("MSFT  240621C00325000", trader.InstructionBuyToOpen, trader.PositionEffectOpening, 1),
	)
	closeOrder := optionOrder(
		trader.OrderTypeRequestNetDebit,
		trader.ComplexOrderStrategyTypeVertical,
		1,
		0.55,
		optionLeg("MSFT  240621C00320000", trader.InstructionBuyToClose, trader.PositionEffectClosing, 1),
		optionLeg("MSFT  240621C00325000", trader.InstructionSellToClose, trader.PositionEffectClosing, 1),
	)

	printOptionOrder("open bear call spread", openOrder)
	printOptionOrder("close bear call spread", closeOrder)

	// Output:
	// open bear call spread
	// orderType=NET_CREDIT complexOrderStrategyType=VERTICAL orderStrategyType=SINGLE quantity=1 price=1.15
	// leg 1: SELL_TO_OPEN OPENING 1 MSFT  240621C00320000
	// leg 2: BUY_TO_OPEN OPENING 1 MSFT  240621C00325000
	// close bear call spread
	// orderType=NET_DEBIT complexOrderStrategyType=VERTICAL orderStrategyType=SINGLE quantity=1 price=0.55
	// leg 1: BUY_TO_CLOSE CLOSING 1 MSFT  240621C00320000
	// leg 2: SELL_TO_CLOSE CLOSING 1 MSFT  240621C00325000
}

func ExampleOrderRequest_verticalBullPutSpread() {
	openOrder := optionOrder(
		trader.OrderTypeRequestNetCredit,
		trader.ComplexOrderStrategyTypeVertical,
		1,
		1.05,
		optionLeg("SPY   240621P00530000", trader.InstructionSellToOpen, trader.PositionEffectOpening, 1),
		optionLeg("SPY   240621P00525000", trader.InstructionBuyToOpen, trader.PositionEffectOpening, 1),
	)
	closeOrder := optionOrder(
		trader.OrderTypeRequestNetDebit,
		trader.ComplexOrderStrategyTypeVertical,
		1,
		0.40,
		optionLeg("SPY   240621P00530000", trader.InstructionBuyToClose, trader.PositionEffectClosing, 1),
		optionLeg("SPY   240621P00525000", trader.InstructionSellToClose, trader.PositionEffectClosing, 1),
	)

	printOptionOrder("open bull put spread", openOrder)
	printOptionOrder("close bull put spread", closeOrder)

	// Output:
	// open bull put spread
	// orderType=NET_CREDIT complexOrderStrategyType=VERTICAL orderStrategyType=SINGLE quantity=1 price=1.05
	// leg 1: SELL_TO_OPEN OPENING 1 SPY   240621P00530000
	// leg 2: BUY_TO_OPEN OPENING 1 SPY   240621P00525000
	// close bull put spread
	// orderType=NET_DEBIT complexOrderStrategyType=VERTICAL orderStrategyType=SINGLE quantity=1 price=0.40
	// leg 1: BUY_TO_CLOSE CLOSING 1 SPY   240621P00530000
	// leg 2: SELL_TO_CLOSE CLOSING 1 SPY   240621P00525000
}

func ExampleOrderRequest_verticalBearPutSpread() {
	openOrder := optionOrder(
		trader.OrderTypeRequestNetDebit,
		trader.ComplexOrderStrategyTypeVertical,
		1,
		1.35,
		optionLeg("QQQ   240621P00450000", trader.InstructionBuyToOpen, trader.PositionEffectOpening, 1),
		optionLeg("QQQ   240621P00445000", trader.InstructionSellToOpen, trader.PositionEffectOpening, 1),
	)
	closeOrder := optionOrder(
		trader.OrderTypeRequestNetCredit,
		trader.ComplexOrderStrategyTypeVertical,
		1,
		2.10,
		optionLeg("QQQ   240621P00450000", trader.InstructionSellToClose, trader.PositionEffectClosing, 1),
		optionLeg("QQQ   240621P00445000", trader.InstructionBuyToClose, trader.PositionEffectClosing, 1),
	)

	printOptionOrder("open bear put spread", openOrder)
	printOptionOrder("close bear put spread", closeOrder)

	// Output:
	// open bear put spread
	// orderType=NET_DEBIT complexOrderStrategyType=VERTICAL orderStrategyType=SINGLE quantity=1 price=1.35
	// leg 1: BUY_TO_OPEN OPENING 1 QQQ   240621P00450000
	// leg 2: SELL_TO_OPEN OPENING 1 QQQ   240621P00445000
	// close bear put spread
	// orderType=NET_CREDIT complexOrderStrategyType=VERTICAL orderStrategyType=SINGLE quantity=1 price=2.10
	// leg 1: SELL_TO_CLOSE CLOSING 1 QQQ   240621P00450000
	// leg 2: BUY_TO_CLOSE CLOSING 1 QQQ   240621P00445000
}

func ExampleOrderRequest_customOptionSpread() {
	order := optionOrder(
		trader.OrderTypeRequestMarket,
		trader.ComplexOrderStrategyTypeCustom,
		0,
		0,
		optionLeg("TSLA  240621C00180000", trader.InstructionBuyToOpen, trader.PositionEffectOpening, 1),
		optionLeg("TSLA  240621C00190000", trader.InstructionSellToOpen, trader.PositionEffectOpening, 2),
	)

	printOptionOrder("custom one-by-two call spread", order)

	// Output:
	// custom one-by-two call spread
	// orderType=MARKET complexOrderStrategyType=CUSTOM orderStrategyType=SINGLE
	// leg 1: BUY_TO_OPEN OPENING 1 TSLA  240621C00180000
	// leg 2: SELL_TO_OPEN OPENING 2 TSLA  240621C00190000
}

func optionLeg(
	symbol string,
	instruction trader.Instruction,
	positionEffect trader.PositionEffect,
	quantity float64,
) trader.OrderLegRequest {
	return trader.OrderLegRequest{
		OrderLegType:   trader.OrderLegTypeOption,
		Instruction:    instruction,
		PositionEffect: positionEffect,
		Quantity:       quantity,
		Instrument: trader.OrderInstrument{
			AssetType: schwab.AssetTypeOption,
			Symbol:    symbol,
		},
	}
}

func optionOrder(
	orderType trader.OrderTypeRequest,
	complexOrderStrategyType trader.ComplexOrderStrategyType,
	quantity float64,
	price float64,
	legs ...trader.OrderLegRequest,
) *trader.OrderRequest {
	return &trader.OrderRequest{
		Session:                  trader.SessionNormal,
		Duration:                 trader.DurationDay,
		OrderType:                orderType,
		ComplexOrderStrategyType: complexOrderStrategyType,
		Quantity:                 quantity,
		Price:                    price,
		OrderStrategyType:        trader.OrderStrategyTypeSingle,
		OrderLegCollection:       legs,
	}
}

func printOptionOrder(label string, order *trader.OrderRequest) {
	if _, err := json.Marshal(order); err != nil {
		panic(err)
	}

	var output strings.Builder
	output.WriteString(label)
	output.WriteByte('\n')
	output.WriteString("orderType=")
	output.WriteString(string(order.OrderType))
	output.WriteString(" complexOrderStrategyType=")
	output.WriteString(string(order.ComplexOrderStrategyType))
	output.WriteString(" orderStrategyType=")
	output.WriteString(string(order.OrderStrategyType))
	if order.Quantity != 0 {
		output.WriteString(" quantity=")
		output.WriteString(strconv.FormatFloat(order.Quantity, 'f', 0, 64))
	}
	if order.Price != 0 {
		output.WriteString(" price=")
		output.WriteString(strconv.FormatFloat(order.Price, 'f', 2, 64))
	}
	output.WriteByte('\n')

	for i, leg := range order.OrderLegCollection {
		output.WriteString("leg ")
		output.WriteString(strconv.Itoa(i + 1))
		output.WriteString(": ")
		output.WriteString(string(leg.Instruction))
		output.WriteByte(' ')
		output.WriteString(string(leg.PositionEffect))
		output.WriteByte(' ')
		output.WriteString(strconv.FormatFloat(leg.Quantity, 'f', 0, 64))
		output.WriteByte(' ')
		output.WriteString(leg.Instrument.Symbol)
		output.WriteByte('\n')
	}

	if _, err := os.Stdout.WriteString(output.String()); err != nil {
		panic(err)
	}
}
