package marketdata_test

import (
	"context"
	"net/http"
	"strconv"

	"github.com/major/schwab-go/schwab/marketdata"
)

func ExampleClient_GetMarketHours() {
	var method string
	var path string
	var markets string
	client, closeServer := newMarketDataExampleClient(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		markets = r.URL.Query().Get("markets")
		writeExampleJSON(w, marketHoursExample())
	})
	defer closeServer()

	hours, err := client.GetMarketHours(context.Background(), []string{"equity", "option"}, "")
	if err != nil {
		panic(err)
	}

	equity := hours["equity"]["EQ"]
	printExampleLines(
		"request="+method+" "+path,
		"markets="+markets,
		"product="+equity.Product,
		"open="+strconv.FormatBool(equity.IsOpen),
	)

	// Output:
	// request=GET /markets
	// markets=equity,option
	// product=EQ
	// open=true
}

func ExampleClient_GetMarketHoursTyped() {
	client, closeServer := newMarketDataExampleClient(func(w http.ResponseWriter, _ *http.Request) {
		writeExampleJSON(w, marketHoursExample())
	})
	defer closeServer()

	hours, err := client.GetMarketHoursTyped(
		context.Background(),
		[]marketdata.MarketID{marketdata.MarketIDEquity, marketdata.MarketIDOption},
		"",
	)
	if err != nil {
		panic(err)
	}

	equity := hours["equity"]["EQ"]
	printExampleLines(
		"product="+equity.Product,
		"session="+equity.SessionHours["regularMarket"][0].Start,
	)

	// Output:
	// product=EQ
	// session=2024-01-15T09:30:00-05:00
}

func ExampleClient_GetMarketHoursSingle() {
	var method string
	var path string
	client, closeServer := newMarketDataExampleClient(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		writeExampleJSON(w, marketHoursExample())
	})
	defer closeServer()

	hours, err := client.GetMarketHoursSingle(context.Background(), "equity", "")
	if err != nil {
		panic(err)
	}

	equity := hours["equity"]["EQ"]
	printExampleLines(
		"request="+method+" "+path,
		"productName="+equity.ProductName,
		"exchange="+equity.Exchange,
	)

	// Output:
	// request=GET /markets/equity
	// productName=equity
	// exchange=NYSE
}

func ExampleClient_GetMarketHoursSingleTyped() {
	client, closeServer := newMarketDataExampleClient(func(w http.ResponseWriter, _ *http.Request) {
		writeExampleJSON(w, marketHoursExample())
	})
	defer closeServer()

	hours, err := client.GetMarketHoursSingleTyped(context.Background(), marketdata.MarketIDEquity, "")
	if err != nil {
		panic(err)
	}

	equity := hours["equity"]["EQ"]
	printExampleLines(
		"marketType="+equity.MarketType,
		"category="+equity.Category,
	)

	// Output:
	// marketType=EQUITY
	// category=EQUITY
}

func marketHoursExample() marketdata.MarketHoursMap {
	return marketdata.MarketHoursMap{
		"equity": {
			"EQ": {
				Date:        "2024-01-15",
				MarketType:  "EQUITY",
				Exchange:    "NYSE",
				Category:    "EQUITY",
				Product:     "EQ",
				ProductName: "equity",
				IsOpen:      true,
				SessionHours: map[string][]marketdata.SessionHours{
					"regularMarket": {
						{Start: "2024-01-15T09:30:00-05:00", End: "2024-01-15T16:00:00-05:00"},
					},
				},
			},
		},
	}
}
