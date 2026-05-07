package marketdata_test

import (
	"context"
	"net/http"
	"strconv"

	"github.com/major/schwab-go/schwab/marketdata"
)

func ExampleClient_GetMovers() {
	var method string
	var path string
	var sort string
	client, closeServer := newMarketDataExampleClient(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		sort = r.URL.Query().Get("sort")
		writeExampleJSON(w, marketdata.MoverResponse{
			Screeners: []marketdata.Screener{
				{
					Symbol:           "AAPL",
					Description:      "Apple Inc.",
					Direction:        "up",
					Last:             150.25,
					Change:           2.50,
					NetPercentChange: 1.69,
					TotalVolume:      50000000,
				},
			},
		})
	})
	defer closeServer()

	frequency := 60
	movers, err := client.GetMovers(context.Background(), "$DJI", marketdata.MoverSortVolume, &frequency)
	if err != nil {
		panic(err)
	}

	first := movers.Screeners[0]
	printExampleLines(
		"request="+method+" "+path,
		"sort="+sort,
		"movers="+strconv.Itoa(len(movers.Screeners)),
		"first="+first.Symbol+" "+first.Direction,
		"last="+formatExampleFloat(first.Last),
	)

	// Output:
	// request=GET /movers/$DJI
	// sort=VOLUME
	// movers=1
	// first=AAPL up
	// last=150.25
}
