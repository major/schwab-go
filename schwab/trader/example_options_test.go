package trader_test

import (
	"encoding/json"
	"fmt"
	"log"

	schwab "github.com/major/schwab-go/schwab"
	"github.com/major/schwab-go/schwab/trader"
)

func ExampleOrderRequest_singleLegOption() {
	order := &trader.OrderRequest{
		Session:           trader.SessionNormal,
		Duration:          trader.DurationDay,
		OrderType:         trader.OrderTypeRequestLimit,
		Price:             1.45,
		OrderStrategyType: trader.OrderStrategyTypeSingle,
		OrderLegCollection: []trader.OrderLegRequest{
			{
				OrderLegType:   trader.OrderLegTypeOption,
				Instruction:    trader.InstructionBuyToOpen,
				PositionEffect: trader.PositionEffectOpening,
				Quantity:       1,
				Instrument: trader.OrderInstrument{
					AssetType: schwab.AssetTypeOption,
					Symbol:    "AAPL  240621C00200000",
				},
			},
		},
	}

	payload, err := json.MarshalIndent(order, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(payload))

	// Output:
	// {
	//   "session": "NORMAL",
	//   "duration": "DAY",
	//   "orderType": "LIMIT",
	//   "price": 1.45,
	//   "orderStrategyType": "SINGLE",
	//   "orderLegCollection": [
	//     {
	//       "orderLegType": "OPTION",
	//       "instrument": {
	//         "assetType": "OPTION",
	//         "cusip": "",
	//         "symbol": "AAPL  240621C00200000",
	//         "description": "",
	//         "instrumentId": 0,
	//         "type": ""
	//       },
	//       "instruction": "BUY_TO_OPEN",
	//       "positionEffect": "OPENING",
	//       "quantity": 1
	//     }
	//   ]
	// }
}

func ExampleOrderRequest_verticalOptionSpread() {
	order := &trader.OrderRequest{
		Session:                  trader.SessionNormal,
		Duration:                 trader.DurationDay,
		OrderType:                trader.OrderTypeRequestNetDebit,
		ComplexOrderStrategyType: trader.ComplexOrderStrategyTypeVertical,
		Quantity:                 1,
		Price:                    1.25,
		OrderStrategyType:        trader.OrderStrategyTypeSingle,
		OrderLegCollection: []trader.OrderLegRequest{
			{
				OrderLegType:   trader.OrderLegTypeOption,
				Instruction:    trader.InstructionBuyToOpen,
				PositionEffect: trader.PositionEffectOpening,
				Quantity:       1,
				Instrument: trader.OrderInstrument{
					AssetType: schwab.AssetTypeOption,
					Symbol:    "AAPL  240621C00195000",
				},
			},
			{
				OrderLegType:   trader.OrderLegTypeOption,
				Instruction:    trader.InstructionSellToOpen,
				PositionEffect: trader.PositionEffectOpening,
				Quantity:       1,
				Instrument: trader.OrderInstrument{
					AssetType: schwab.AssetTypeOption,
					Symbol:    "AAPL  240621C00200000",
				},
			},
		},
	}

	payload, err := json.MarshalIndent(order, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(payload))

	// Output:
	// {
	//   "session": "NORMAL",
	//   "duration": "DAY",
	//   "orderType": "NET_DEBIT",
	//   "complexOrderStrategyType": "VERTICAL",
	//   "quantity": 1,
	//   "price": 1.25,
	//   "orderStrategyType": "SINGLE",
	//   "orderLegCollection": [
	//     {
	//       "orderLegType": "OPTION",
	//       "instrument": {
	//         "assetType": "OPTION",
	//         "cusip": "",
	//         "symbol": "AAPL  240621C00195000",
	//         "description": "",
	//         "instrumentId": 0,
	//         "type": ""
	//       },
	//       "instruction": "BUY_TO_OPEN",
	//       "positionEffect": "OPENING",
	//       "quantity": 1
	//     },
	//     {
	//       "orderLegType": "OPTION",
	//       "instrument": {
	//         "assetType": "OPTION",
	//         "cusip": "",
	//         "symbol": "AAPL  240621C00200000",
	//         "description": "",
	//         "instrumentId": 0,
	//         "type": ""
	//       },
	//       "instruction": "SELL_TO_OPEN",
	//       "positionEffect": "OPENING",
	//       "quantity": 1
	//     }
	//   ]
	// }
}
