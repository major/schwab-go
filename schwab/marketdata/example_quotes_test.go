package marketdata_test

import (
	"context"
	"net/http"
	"strconv"

	schwab "github.com/major/schwab-go/schwab"
	"github.com/major/schwab-go/schwab/marketdata"
)

func ExampleClient_GetQuotes() {
	var method string
	var path string
	var symbols string
	client, closeServer := newMarketDataExampleClient(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		symbols = r.URL.Query().Get("symbols")
		writeExampleJSON(w, map[string]*marketdata.QuoteEntry{
			"AAPL": equityQuoteExample("AAPL"),
			"MSFT": equityQuoteExample("MSFT"),
		})
	})
	defer closeServer()

	quotes, quoteErr, err := client.GetQuotes(
		context.Background(),
		[]string{"AAPL", "MSFT"},
		"quote,reference",
		false,
	)
	if err != nil {
		panic(err)
	}
	if quoteErr != nil {
		panic("unexpected quote error")
	}

	entry := (*quotes)["AAPL"]
	quote, err := entry.EquityQuote()
	if err != nil {
		panic(err)
	}
	reference, err := entry.EquityReference()
	if err != nil {
		panic(err)
	}

	printExampleLines(
		"request="+method+" "+path,
		"symbols="+symbols,
		"quotes="+strconv.Itoa(len(*quotes)),
		"first="+entry.Symbol+" "+reference.Description,
		"last="+formatExampleFloat(quote.LastPrice),
	)

	// Output:
	// request=GET /quotes
	// symbols=AAPL,MSFT
	// quotes=2
	// first=AAPL Apple Inc
	// last=170.01
}

func ExampleClient_GetQuote() {
	var method string
	var path string
	client, closeServer := newMarketDataExampleClient(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		writeExampleJSON(w, marketdata.QuoteResponse{
			"AAPL": equityQuoteExample("AAPL"),
		})
	})
	defer closeServer()

	quotes, err := client.GetQuote(context.Background(), "AAPL", "quote,reference")
	if err != nil {
		panic(err)
	}

	entry := (*quotes)["AAPL"]
	quote, err := entry.EquityQuote()
	if err != nil {
		panic(err)
	}

	printExampleLines(
		"request="+method+" "+path,
		"symbol="+entry.Symbol,
		"last="+formatExampleFloat(quote.LastPrice),
	)

	// Output:
	// request=GET /AAPL/quotes
	// symbol=AAPL
	// last=170.01
}

func equityQuoteExample(symbol string) *marketdata.QuoteEntry {
	return &marketdata.QuoteEntry{
		AssetMainType: schwab.AssetTypeEquity,
		QuoteType:     marketdata.QuoteTypeNBBO,
		Realtime:      true,
		Symbol:        symbol,
		Quote: mustRawExample(marketdata.EquityQuote{
			AskPrice:    170.12,
			BidPrice:    169.98,
			LastPrice:   170.01,
			TotalVolume: 12345678,
			Exchange:    "NASDAQ",
		}),
		Reference: mustRawExample(marketdata.EquityReference{
			CUSIP:        "037833100",
			Description:  "Apple Inc",
			Exchange:     "Q",
			ExchangeName: "NASDAQ",
		}),
	}
}
