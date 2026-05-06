package marketdata

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schwab "github.com/major/schwab-go/schwab"
)

func TestGetQuotes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/quotes", r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		require.Equal(t, "AAPL,MSFT", r.URL.Query().Get("symbols"))
		require.Equal(t, "quote", r.URL.Query().Get("fields"))
		require.Equal(t, "true", r.URL.Query().Get("indicative"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]QuoteEntry{
			"AAPL": equityQuoteEntry("AAPL"),
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithToken("test-token"),
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	quotes, quoteErr, err := client.GetQuotes(context.Background(), []string{"AAPL", "MSFT"}, "quote", true)
	require.NoError(t, err)
	require.Nil(t, quoteErr)
	require.NotNil(t, quotes)

	entry := (*quotes)["AAPL"]
	require.NotNil(t, entry)
	assert.Equal(t, schwab.AssetTypeEquity, entry.AssetMainType)
	assert.Equal(t, "AAPL", entry.Symbol)

	equity, err := entry.EquityQuote()
	require.NoError(t, err)
	assert.Equal(t, 170.12, equity.AskPrice)
	assert.Equal(t, 169.98, equity.BidPrice)
	assert.Equal(t, 170.01, equity.LastPrice)
	assert.Equal(t, int64(12345678), equity.TotalVolume)
	assert.Equal(t, 199.62, equity.Week52High)
	assert.Equal(t, 124.17, equity.Week52Low)
	assert.Equal(t, 28.5, equity.PeRatio)
	assert.Equal(t, "NASDAQ", equity.Exchange)
	assert.Equal(t, int64(1712345678901), equity.TradeTime)
	assert.Equal(t, int64(1712345678999), equity.QuoteTime)
	reference, err := entry.EquityReference()
	require.NoError(t, err)
	assert.Equal(t, "037833100", reference.CUSIP)
	require.NotNil(t, entry.Regular)
	assert.Equal(t, 170.01, entry.Regular.LastPrice)
	require.NotNil(t, entry.Extended)
	assert.Equal(t, 170.05, entry.Extended.LastPrice)
	require.NotNil(t, entry.Fundamental)
	assert.Equal(t, 28.5, entry.Fundamental.PERatio)
}

func TestGetQuotesOption(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/quotes", r.URL.Path)
		require.Equal(t, "AAPL_052424C170", r.URL.Query().Get("symbols"))
		require.Empty(t, r.URL.Query().Get("fields"))
		require.Empty(t, r.URL.Query().Get("indicative"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]QuoteEntry{
			"AAPL_052424C170": optionQuoteEntry("AAPL_052424C170"),
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	quotes, quoteErr, err := client.GetQuotes(context.Background(), []string{"AAPL_052424C170"}, "", false)
	require.NoError(t, err)
	require.Nil(t, quoteErr)

	entry := (*quotes)["AAPL_052424C170"]
	require.NotNil(t, entry)
	option, err := entry.OptionQuote()
	require.NoError(t, err)
	assert.Equal(t, 0.52, option.Delta)
	assert.Equal(t, 0.03, option.Gamma)
	assert.Equal(t, -0.11, option.Theta)
	assert.Equal(t, 0.22, option.Vega)
	assert.Equal(t, 0.02, option.Rho)
	assert.Equal(t, int64(1234), option.OpenInterest)
	assert.Equal(t, 170.0, option.StrikePrice)
	assert.Equal(t, 171.25, option.UnderlyingPrice)

	reference, err := entry.OptionReference()
	require.NoError(t, err)
	assert.Equal(t, 170.0, reference.StrikePrice)
	assert.Equal(t, "AAPL", reference.Underlying)
}

func TestGetQuotesMixed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "AAPL,$SPX,VFIAX,EUR/USD,/ES", r.URL.Query().Get("symbols"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]QuoteEntry{
			"AAPL":    equityQuoteEntry("AAPL"),
			"$SPX":    indexQuoteEntry("$SPX"),
			"VFIAX":   mutualFundQuoteEntry("VFIAX"),
			"EUR/USD": forexQuoteEntry("EUR/USD"),
			"/ES":     futureQuoteEntry("/ES"),
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	quotes, quoteErr, err := client.GetQuotes(context.Background(), []string{"AAPL", "$SPX", "VFIAX", "EUR/USD", "/ES"}, "", false)
	require.NoError(t, err)
	require.Nil(t, quoteErr)
	require.Len(t, *quotes, 5)

	indexQuote, err := (*quotes)["$SPX"].IndexQuote()
	require.NoError(t, err)
	assert.Equal(t, 5123.45, indexQuote.LastPrice)

	fundQuote, err := (*quotes)["VFIAX"].MutualFundQuote()
	require.NoError(t, err)
	assert.Equal(t, 470.23, fundQuote.ClosePrice)

	forexQuote, err := (*quotes)["EUR/USD"].ForexQuote()
	require.NoError(t, err)
	assert.Equal(t, 1.08, forexQuote.LastPrice)

	futureQuote, err := (*quotes)["/ES"].FutureQuote()
	require.NoError(t, err)
	assert.Equal(t, 5125.25, futureQuote.SettlementPrice)
	assert.Equal(t, int64(987654), futureQuote.OpenInterest)

	futureReference, err := (*quotes)["/ES"].FutureReference()
	require.NoError(t, err)
	assert.Equal(t, int64(20240621), futureReference.FutureExpirationDate)
	assert.Equal(t, 5125.25, futureReference.FutureSettlementPrice)
}

func TestGetQuotesPartialFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]any{
			"AAPL": equityQuoteEntry("AAPL"),
			"errors": QuoteError{
				InvalidCusips:  []string{"123456789"},
				InvalidSSIDs:   []int64{12345},
				InvalidSymbols: []string{"NOTREAL"},
			},
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	quotes, quoteErr, err := client.GetQuotes(context.Background(), []string{"AAPL", "NOTREAL"}, "quote", false)
	require.NoError(t, err)
	require.NotNil(t, quoteErr)
	require.Len(t, *quotes, 1)
	assert.Equal(t, []string{"NOTREAL"}, quoteErr.InvalidSymbols)
	assert.Equal(t, []string{"123456789"}, quoteErr.InvalidCusips)
	assert.Equal(t, []int64{12345}, quoteErr.InvalidSSIDs)
}

func TestGetQuote(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/$SPX/quotes", r.URL.Path)
		require.Equal(t, "quote,reference", r.URL.Query().Get("fields"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]QuoteEntry{
			"$SPX": indexQuoteEntry("$SPX"),
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	quotes, err := client.GetQuote(context.Background(), "$SPX", "quote,reference")
	require.NoError(t, err)
	require.NotNil(t, quotes)

	entry := (*quotes)["$SPX"]
	require.NotNil(t, entry)
	assert.Equal(t, schwab.AssetTypeIndex, entry.AssetMainType)
	assert.Equal(t, "$SPX", entry.Symbol)

	indexQuote, err := entry.IndexQuote()
	require.NoError(t, err)
	assert.Equal(t, 5123.45, indexQuote.LastPrice)
	assert.Equal(t, "CBOE", indexQuote.Exchange)
}

func TestGetQuotesError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, map[string]string{
			"detail": "invalid fields parameter",
		})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	quotes, quoteErr, err := client.GetQuotes(context.Background(), []string{"AAPL"}, "invalid", false)
	require.Error(t, err)
	require.Nil(t, quotes)
	require.Nil(t, quoteErr)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
	assert.Equal(t, "invalid fields parameter", apiErr.Message)
}

func TestQuoteWrongHelper(t *testing.T) {
	entry := equityQuoteEntry("AAPL")

	option, err := entry.OptionQuote()
	require.Error(t, err)
	require.Nil(t, option)
	assert.EqualError(t, err, "quote is EQUITY, not OPTION")
}

func equityQuoteEntry(symbol string) QuoteEntry {
	return QuoteEntry{
		AssetMainType: schwab.AssetTypeEquity,
		AssetSubType:  "COE",
		QuoteType:     "NBBO",
		Realtime:      true,
		SSID:          123456,
		Symbol:        symbol,
		Quote: mustRaw(map[string]any{
			"askPrice":          170.12,
			"bidPrice":          169.98,
			"lastPrice":         170.01,
			"totalVolume":       12345678,
			"highPrice":         172.25,
			"lowPrice":          168.5,
			"openPrice":         169.25,
			"closePrice":        168.75,
			"netChange":         1.26,
			"netPercentChange":  0.75,
			"52WeekHigh":        199.62,
			"52WeekLow":         124.17,
			"peRatio":           28.5,
			"divAmount":         0.24,
			"divYield":          0.45,
			"divDate":           "2024-05-10",
			"eps":               6.05,
			"exchange":          "NASDAQ",
			"exchangeName":      "NASDAQ Global Select",
			"mark":              170.04,
			"markChange":        1.29,
			"markPercentChange": 0.77,
			"bidSize":           100,
			"askSize":           200,
			"lastSize":          50,
			"tradeTime":         1712345678901,
			"quoteTime":         1712345678999,
		}),
		Reference: mustRaw(EquityReference{
			CUSIP:        "037833100",
			Description:  "Apple Inc",
			Exchange:     "Q",
			ExchangeName: "NASDAQ Global Select",
			Shortable:    true,
		}),
		Regular: &RegularMarket{
			LastPrice:     170.01,
			LastSize:      50,
			NetChange:     1.26,
			PercentChange: 0.75,
			TradeTime:     1712345678901,
		},
		Extended: &ExtendedMarket{
			BidPrice:    169.95,
			BidSize:     25,
			AskPrice:    170.15,
			AskSize:     30,
			LastPrice:   170.05,
			LastSize:    10,
			TotalVolume: 1000,
			TradeTime:   1712345680000,
		},
		Fundamental: &Fundamental{
			DivAmount: 0.24,
			DivYield:  0.45,
			EPS:       6.05,
			PERatio:   28.5,
		},
	}
}

func optionQuoteEntry(symbol string) QuoteEntry {
	entry := equityQuoteEntry(symbol)
	entry.AssetMainType = schwab.AssetTypeOption
	entry.AssetSubType = "O"
	entry.Reference = mustRaw(OptionReference{
		ContractType:     "CALL",
		CUSIP:            "037833100",
		DaysToExpiration: 30,
		Description:      "AAPL 05/24/2024 $170 Call",
		Exchange:         "OPR",
		ExchangeName:     "Options Price Reporting Authority",
		ExerciseType:     "A",
		ExpirationDay:    24,
		ExpirationMonth:  5,
		ExpirationType:   "S",
		ExpirationYear:   2024,
		Multiplier:       100,
		SettlementType:   "P",
		StrikePrice:      170,
		Underlying:       "AAPL",
	})
	entry.Quote = mustRaw(map[string]any{
		"askPrice":               4.25,
		"bidPrice":               4.15,
		"lastPrice":              4.2,
		"totalVolume":            321,
		"highPrice":              4.5,
		"lowPrice":               3.9,
		"openPrice":              4.0,
		"closePrice":             4.1,
		"netChange":              0.1,
		"netPercentChange":       2.44,
		"delta":                  0.52,
		"gamma":                  0.03,
		"theta":                  -0.11,
		"vega":                   0.22,
		"rho":                    0.02,
		"openInterest":           1234,
		"strikePrice":            170.0,
		"expirationDate":         "2024-05-24",
		"intrinsicValue":         1.25,
		"extrinsicValue":         2.95,
		"timeValue":              2.95,
		"impliedYield":           0.0,
		"theoreticalOptionValue": 4.18,
		"theoreticalVolatility":  0.31,
		"underlyingPrice":        171.25,
		"volatilityPct":          31.0,
		"exchange":               "OPR",
		"exchangeName":           "Options Price Reporting Authority",
		"mark":                   4.2,
		"markChange":             0.1,
		"markPercentChange":      2.44,
		"bidSize":                10,
		"askSize":                12,
		"lastSize":               1,
		"tradeTime":              1712345678901,
		"quoteTime":              1712345678999,
	})
	return entry
}

func indexQuoteEntry(symbol string) QuoteEntry {
	return QuoteEntry{
		AssetMainType: schwab.AssetTypeIndex,
		AssetSubType:  "I",
		QuoteType:     "NBBO",
		Realtime:      true,
		SSID:          789,
		Symbol:        symbol,
		Quote: mustRaw(map[string]any{
			"lastPrice":        5123.45,
			"netChange":        10.25,
			"openPrice":        5110.0,
			"closePrice":       5113.2,
			"highPrice":        5130.5,
			"lowPrice":         5101.0,
			"totalVolume":      0,
			"tradeTime":        1712345678901,
			"exchange":         "CBOE",
			"exchangeName":     "Cboe Global Indices",
			"netPercentChange": 0.2,
			"52WeekHigh":       5264.85,
			"52WeekLow":        4103.78,
		}),
	}
}

func mutualFundQuoteEntry(symbol string) QuoteEntry {
	return QuoteEntry{
		AssetMainType: schwab.AssetTypeMutualFund,
		Symbol:        symbol,
		Quote: mustRaw(map[string]any{
			"closePrice":       470.23,
			"netChange":        1.15,
			"totalVolume":      0,
			"tradeTime":        1712345678901,
			"exchange":         "NASDAQ",
			"exchangeName":     "NASDAQ Mutual Funds",
			"netPercentChange": 0.25,
			"52WeekHigh":       480.0,
			"52WeekLow":        390.0,
		}),
	}
}

func forexQuoteEntry(symbol string) QuoteEntry {
	return QuoteEntry{
		AssetMainType: schwab.AssetTypeForex,
		Symbol:        symbol,
		Quote: mustRaw(map[string]any{
			"askPrice":         1.0801,
			"bidPrice":         1.0799,
			"lastPrice":        1.08,
			"openPrice":        1.075,
			"closePrice":       1.076,
			"highPrice":        1.081,
			"lowPrice":         1.074,
			"netChange":        0.004,
			"netPercentChange": 0.37,
			"mark":             1.08,
			"tick":             0.0001,
			"tickAmount":       10.0,
			"exchange":         "FOREX",
			"exchangeName":     "Foreign Exchange",
			"tradeTime":        1712345678901,
		}),
	}
}

func futureQuoteEntry(symbol string) QuoteEntry {
	return QuoteEntry{
		AssetMainType: schwab.AssetTypeFuture,
		Symbol:        symbol,
		Reference: mustRaw(FutureReference{
			Description:           "E-mini S&P 500 Future",
			Exchange:              "CME",
			ExchangeName:          "Chicago Mercantile Exchange",
			FutureActiveSymbol:    "/ESM24",
			FutureExpirationDate:  20240621,
			FutureIsActive:        true,
			FutureMultiplier:      50,
			FuturePriceFormat:     "D",
			FutureSettlementPrice: 5125.25,
			FutureTradingHours:    "ETH",
			Product:               "/ES",
		}),
		Quote: mustRaw(map[string]any{
			"askPrice":         5126.0,
			"bidPrice":         5125.5,
			"lastPrice":        5125.75,
			"openPrice":        5110.0,
			"closePrice":       5120.0,
			"highPrice":        5132.0,
			"lowPrice":         5104.0,
			"netChange":        5.75,
			"netPercentChange": 0.11,
			"mark":             5125.75,
			"tick":             0.25,
			"tickAmount":       12.5,
			"exchange":         "CME",
			"exchangeName":     "Chicago Mercantile Exchange",
			"tradeTime":        1712345678901,
			"settlementPrice":  5125.25,
			"openInterest":     987654,
		}),
	}
}

func mustRaw(value any) json.RawMessage {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return raw
}
