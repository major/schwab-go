package marketdata_test

import (
	"context"
	"net/http"
	"strconv"

	"github.com/major/schwab-go/schwab/marketdata"
)

func ExampleClient_GetExpirationChain() {
	var method string
	var path string
	var symbol string
	client, closeServer := newMarketDataExampleClient(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		symbol = r.URL.Query().Get("symbol")
		writeExampleJSON(w, marketdata.ExpirationChain{
			ExpirationList: []marketdata.Expiration{
				{
					ExpirationDate:   "2024-01-19",
					DaysToExpiration: 4,
					ExpirationType:   marketdata.OptionExpirationTypeStandard,
					SettlementType:   marketdata.OptionSettlementTypePM,
					OptionRoots:      "AAPL",
					Standard:         true,
				},
			},
		})
	})
	defer closeServer()

	chain, err := client.GetExpirationChain(context.Background(), "AAPL")
	if err != nil {
		panic(err)
	}

	expiration := chain.ExpirationList[0]
	printExampleLines(
		"request="+method+" "+path,
		"symbol="+symbol,
		"expirations="+strconv.Itoa(len(chain.ExpirationList)),
		"first="+expiration.ExpirationDate,
		"days="+strconv.Itoa(expiration.DaysToExpiration),
	)

	// Output:
	// request=GET /expirationchain
	// symbol=AAPL
	// expirations=1
	// first=2024-01-19
	// days=4
}
