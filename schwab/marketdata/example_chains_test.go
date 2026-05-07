package marketdata_test

import (
	"context"
	"net/http"
	"strconv"

	"github.com/major/schwab-go/schwab/marketdata"
)

func ExampleClient_GetOptionChain() {
	var method string
	var path string
	var symbol string
	client, closeServer := newMarketDataExampleClient(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		symbol = r.URL.Query().Get("symbol")
		writeExampleJSON(w, optionChainExample())
	})
	defer closeServer()

	chain, err := client.GetOptionChain(context.Background(), &marketdata.OptionChainParams{
		Symbol:                 "SPY",
		ContractType:           marketdata.OptionChainContractTypeAll,
		StrikeCount:            2,
		IncludeUnderlyingQuote: true,
		Strategy:               marketdata.OptionChainStrategySingle,
		Range:                  marketdata.OptionChainRangeNearTheMoney,
	})
	if err != nil {
		panic(err)
	}

	call := chain.CallExpDateMap["2024-01-19:5"]["470.0"][0]
	printExampleLines(
		"request="+method+" "+path,
		"symbol="+symbol,
		"status="+chain.Status,
		"underlying="+chain.Underlying.Symbol+" "+formatExampleFloat(chain.Underlying.Last),
		"call="+call.Symbol+" bid="+formatExampleFloat(call.BidPrice),
		"volume="+strconv.FormatInt(call.TotalVolume, 10),
	)

	// Output:
	// request=GET /chains
	// symbol=SPY
	// status=SUCCESS
	// underlying=SPY 470.50
	// call=SPY_011924C470 bid=3.40
	// volume=12345
}

func optionChainExample() marketdata.OptionChain {
	return marketdata.OptionChain{
		Symbol:          "SPY",
		Status:          "SUCCESS",
		Strategy:        marketdata.OptionChainStrategySingle,
		UnderlyingPrice: 470.50,
		Underlying: &marketdata.Underlying{
			Symbol:      "SPY",
			Last:        470.50,
			TotalVolume: 54321000,
		},
		CallExpDateMap: map[string]map[string][]marketdata.OptionContract{
			"2024-01-19:5": {
				"470.0": {
					{
						PutCall:     marketdata.OptionChainContractTypeCall,
						Symbol:      "SPY_011924C470",
						Description: "SPY Jan 19 2024 470 Call",
						BidPrice:    3.40,
						AskPrice:    3.45,
						TotalVolume: 12345,
						StrikePrice: 470,
					},
				},
			},
		},
		PutExpDateMap: map[string]map[string][]marketdata.OptionContract{},
	}
}
