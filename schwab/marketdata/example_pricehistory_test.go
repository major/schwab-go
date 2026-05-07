package marketdata_test

import (
	"context"
	"net/http"
	"strconv"

	"github.com/major/schwab-go/schwab/marketdata"
)

func ExampleClient_GetPriceHistory() {
	var method string
	var path string
	var symbol string
	client, closeServer := newMarketDataExampleClient(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		symbol = r.URL.Query().Get("symbol")
		writeExampleJSON(w, marketdata.CandleList{
			Symbol: "AAPL",
			Candles: []marketdata.Candle{
				{Open: 150, High: 155, Low: 149, Close: 153, Volume: 1000000, Datetime: 1705276800000},
			},
		})
	})
	defer closeServer()

	trueValue := true
	history, err := client.GetPriceHistory(context.Background(), "AAPL", &marketdata.PriceHistoryParams{
		PeriodType:            marketdata.PeriodTypeDay,
		Period:                10,
		FrequencyType:         marketdata.FrequencyTypeMinute,
		Frequency:             5,
		NeedExtendedHoursData: &trueValue,
	})
	if err != nil {
		panic(err)
	}

	candle := history.Candles[0]
	printExampleLines(
		"request="+method+" "+path,
		"symbol="+symbol,
		"candles="+strconv.Itoa(len(history.Candles)),
		"close="+formatExampleFloat(candle.Close),
		"volume="+strconv.FormatInt(candle.Volume, 10),
	)

	// Output:
	// request=GET /pricehistory
	// symbol=AAPL
	// candles=1
	// close=153.00
	// volume=1000000
}
