package marketdata_test

import (
	"context"
	"net/http"
	"strconv"

	schwab "github.com/major/schwab-go/schwab"
	"github.com/major/schwab-go/schwab/marketdata"
)

func ExampleClient_SearchInstruments() {
	var method string
	var path string
	var projection string
	client, closeServer := newMarketDataExampleClient(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		projection = r.URL.Query().Get("projection")
		writeExampleJSON(w, marketdata.InstrumentResponse{
			Instruments: []marketdata.Instrument{instrumentExample()},
		})
	})
	defer closeServer()

	result, err := client.SearchInstruments(context.Background(), "AAPL", marketdata.ProjectionSymbolSearch)
	if err != nil {
		panic(err)
	}

	instrument := result.Instruments[0]
	printExampleLines(
		"request="+method+" "+path,
		"projection="+projection,
		"matches="+strconv.Itoa(len(result.Instruments)),
		"symbol="+instrument.Symbol,
		"cusip="+instrument.Cusip,
	)

	// Output:
	// request=GET /instruments
	// projection=symbol-search
	// matches=1
	// symbol=AAPL
	// cusip=037833100
}

func ExampleClient_GetInstrumentByCUSIP() {
	var method string
	var path string
	client, closeServer := newMarketDataExampleClient(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		writeExampleJSON(w, marketdata.InstrumentResponse{
			Instruments: []marketdata.Instrument{instrumentExample()},
		})
	})
	defer closeServer()

	instrument, err := client.GetInstrumentByCUSIP(context.Background(), "037833100")
	if err != nil {
		panic(err)
	}

	printExampleLines(
		"request="+method+" "+path,
		"symbol="+instrument.Symbol,
		"description="+instrument.Description,
		"assetType="+string(instrument.AssetType),
	)

	// Output:
	// request=GET /instruments/037833100
	// symbol=AAPL
	// description=Apple Inc
	// assetType=EQUITY
}

func instrumentExample() marketdata.Instrument {
	return marketdata.Instrument{
		Cusip:       "037833100",
		Symbol:      "AAPL",
		Description: "Apple Inc",
		Exchange:    "NASDAQ",
		AssetType:   schwab.AssetTypeEquity,
	}
}
