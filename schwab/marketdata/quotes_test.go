package marketdata

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schwab "github.com/major/schwab-go/schwab"
)

func TestGetQuotes(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/quotes", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "AAPL,MSFT", r.URL.Query().Get("symbols"))
		assert.Equal(t, "quote", r.URL.Query().Get("fields"))
		assert.Equal(t, "true", r.URL.Query().Get("indicative"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]QuoteEntry{
			"AAPL": equityQuoteEntry("AAPL"),
		})
	})

	quotes, quoteErr, err := client.GetQuotes(context.Background(), []string{"AAPL", "MSFT"}, "quote", true)
	require.NoError(t, err)
	require.Nil(t, quoteErr)
	require.NotNil(t, quotes)

	entry := (*quotes)["AAPL"]
	require.NotNil(t, entry)
	assert.Equal(t, schwab.AssetTypeEquity, entry.AssetMainType)
	assert.Equal(t, QuoteTypeNBBO, entry.QuoteType)
	assert.Equal(t, "AAPL", entry.Symbol)

	equity, err := entry.EquityQuote()
	require.NoError(t, err)
	assert.InDelta(t, 170.12, equity.AskPrice, 0.000001)
	assert.InDelta(t, 169.98, equity.BidPrice, 0.000001)
	assert.InDelta(t, 170.01, equity.LastPrice, 0.000001)
	assert.Equal(t, int64(12345678), equity.TotalVolume)
	assert.InDelta(t, 199.62, equity.Week52High, 0.000001)
	assert.InDelta(t, 124.17, equity.Week52Low, 0.000001)
	assert.InDelta(t, 28.5, equity.PeRatio, 0.000001)
	assert.Equal(t, "NASDAQ", equity.Exchange)
	assert.Equal(t, int64(1712345678901), equity.TradeTime)
	assert.Equal(t, int64(1712345678999), equity.QuoteTime)
	reference, err := entry.EquityReference()
	require.NoError(t, err)
	assert.Equal(t, "037833100", reference.CUSIP)
	require.NotNil(t, entry.Regular)
	assert.InDelta(t, 170.01, entry.Regular.LastPrice, 0.000001)
	require.NotNil(t, entry.Extended)
	assert.InDelta(t, 170.05, entry.Extended.LastPrice, 0.000001)
	require.NotNil(t, entry.Fundamental)
	assert.InDelta(t, 28.5, entry.Fundamental.PERatio, 0.000001)
}

func TestGetQuotesOption(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/quotes", r.URL.Path)
		assert.Equal(t, "AAPL_052424C170", r.URL.Query().Get("symbols"))
		assert.Empty(t, r.URL.Query().Get("fields"))
		assert.Empty(t, r.URL.Query().Get("indicative"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]QuoteEntry{
			"AAPL_052424C170": optionQuoteEntry("AAPL_052424C170"),
		})
	})

	quotes, quoteErr, err := client.GetQuotes(context.Background(), []string{"AAPL_052424C170"}, "", false)
	require.NoError(t, err)
	require.Nil(t, quoteErr)

	entry := (*quotes)["AAPL_052424C170"]
	require.NotNil(t, entry)
	option, err := entry.OptionQuote()
	require.NoError(t, err)
	assert.InDelta(t, 0.52, option.Delta, 0.000001)
	assert.InDelta(t, 0.03, option.Gamma, 0.000001)
	assert.InDelta(t, -0.11, option.Theta, 0.000001)
	assert.InDelta(t, 0.22, option.Vega, 0.000001)
	assert.InDelta(t, 0.02, option.Rho, 0.000001)
	assert.Equal(t, int64(1234), option.OpenInterest)
	assert.InDelta(t, 4.26, option.IndAskPrice, 0.000001)
	assert.InDelta(t, 4.14, option.IndBidPrice, 0.000001)
	assert.Equal(t, int64(1712345678998), option.IndQuoteTime)
	assert.InDelta(t, 1.25, option.MoneyIntrinsicValue, 0.000001)
	assert.InDelta(t, 170.0, option.StrikePrice, 0.000001)
	assert.InDelta(t, 171.25, option.UnderlyingPrice, 0.000001)

	reference, err := entry.OptionReference()
	require.NoError(t, err)
	assert.Equal(t, OptionContractTypeCall, reference.ContractType)
	assert.Equal(t, OptionExerciseTypeAmerican, reference.ExerciseType)
	assert.Equal(t, OptionExpirationTypeStandard, reference.ExpirationType)
	assert.Equal(t, OptionSettlementTypePM, reference.SettlementType)
	assert.InDelta(t, 170.0, reference.StrikePrice, 0.000001)
	assert.Equal(t, "AAPL", reference.Underlying)
	assert.True(t, reference.PennyPilot)
	assert.Equal(t, int64(1716508800000), reference.LastTradingDay)
}

func TestGetQuotesMixed(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "AAPL,$SPX,VFIAX,EUR/USD,/ES", r.URL.Query().Get("symbols"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]QuoteEntry{
			"AAPL":    equityQuoteEntry("AAPL"),
			"$SPX":    indexQuoteEntry("$SPX"),
			"VFIAX":   mutualFundQuoteEntry("VFIAX"),
			"EUR/USD": forexQuoteEntry("EUR/USD"),
			"/ES":     futureQuoteEntry("/ES"),
		})
	})

	quotes, quoteErr, err := client.GetQuotes(
		context.Background(), []string{"AAPL", "$SPX", "VFIAX", "EUR/USD", "/ES"}, "", false,
	)
	require.NoError(t, err)
	require.Nil(t, quoteErr)
	require.Len(t, *quotes, 5)

	indexQuote, err := (*quotes)["$SPX"].IndexQuote()
	require.NoError(t, err)
	assert.InDelta(t, 5123.45, indexQuote.LastPrice, 0.000001)

	fundQuote, err := (*quotes)["VFIAX"].MutualFundQuote()
	require.NoError(t, err)
	assert.InDelta(t, 470.23, fundQuote.ClosePrice, 0.000001)

	forexQuote, err := (*quotes)["EUR/USD"].ForexQuote()
	require.NoError(t, err)
	assert.InDelta(t, 1.08, forexQuote.LastPrice, 0.000001)

	forexReference, err := (*quotes)["EUR/USD"].ForexReference()
	require.NoError(t, err)
	assert.True(t, forexReference.Tradable)
	assert.Equal(t, "EUR/USD", forexReference.Product)
	assert.Equal(t, "24x5", forexReference.TradingHours)

	futureQuote, err := (*quotes)["/ES"].FutureQuote()
	require.NoError(t, err)
	assert.InDelta(t, 5125.25, futureQuote.SettlementPrice, 0.000001)
	assert.Equal(t, int64(987654), futureQuote.OpenInterest)

	futureReference, err := (*quotes)["/ES"].FutureReference()
	require.NoError(t, err)
	assert.Equal(t, int64(20240621), futureReference.FutureExpirationDate)
	assert.InDelta(t, 5125.25, futureReference.FutureSettlementPrice, 0.000001)
}

func TestGetQuotesFutureOption(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/quotes", r.URL.Path)
		assert.Equal(t, "./EW1H24C5000", r.URL.Query().Get("symbols"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]QuoteEntry{
			"./EW1H24C5000": futureOptionQuoteEntry("./EW1H24C5000"),
		})
	})

	quotes, quoteErr, err := client.GetQuotes(context.Background(), []string{"./EW1H24C5000"}, "", false)
	require.NoError(t, err)
	require.Nil(t, quoteErr)

	entry := (*quotes)["./EW1H24C5000"]
	require.NotNil(t, entry)

	quote, err := entry.FutureOptionQuote()
	require.NoError(t, err)
	assert.InDelta(t, 12.25, quote.SettlemetPrice, 0.000001)
	assert.InDelta(t, 12.3, quote.SettlementPrice, 0.000001)
	assert.Equal(t, int64(1500), quote.OpenInterest)

	reference, err := entry.FutureOptionReference()
	require.NoError(t, err)
	assert.Equal(t, OptionContractTypeCall, reference.ContractType)
	assert.Equal(t, int64(20240315), reference.ExpirationDate)
	assert.Equal(t, "American", reference.ExpirationStyle)
	assert.InDelta(t, 50, reference.Multiplier, 0.000001)
	assert.Equal(t, "/ES", reference.Underlying)
}

func TestGetQuotesPartialFailure(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
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
	})

	quotes, quoteErr, err := client.GetQuotes(context.Background(), []string{"AAPL", "NOTREAL"}, "quote", false)
	require.NoError(t, err)
	require.NotNil(t, quoteErr)
	require.Len(t, *quotes, 1)
	assert.Equal(t, []string{"NOTREAL"}, quoteErr.InvalidSymbols)
	assert.Equal(t, []string{"123456789"}, quoteErr.InvalidCusips)
	assert.Equal(t, []int64{12345}, quoteErr.InvalidSSIDs)
}

func TestGetQuotesQuoteDecodeFailure(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return a raw JSON string value for "BAD" instead of a JSON object.
		// This can't be unmarshaled into a QuoteEntry struct.
		_, err := w.Write([]byte(`{"AAPL": "not-a-quote-object"}`))
		assert.NoError(t, err)
	})

	quotes, quoteErr, err := client.GetQuotes(context.Background(), []string{"AAPL"}, "", false)
	require.Error(t, err)
	require.Nil(t, quotes)
	require.Nil(t, quoteErr)
	assert.Contains(t, err.Error(), "decode quote AAPL:")
}

func TestGetQuotesQuoteErrorDecodeFailure(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// invalidSymbols probes successfully as []string, but invalidSSIDs as a
		// string can't be decoded into []int64 when the full QuoteError is parsed.
		_, err := w.Write([]byte(
			`{"errors": {"invalidSymbols": ["NOTREAL"], "invalidSSIDs": "bad"}}`,
		))
		assert.NoError(t, err)
	})

	quotes, quoteErr, err := client.GetQuotes(context.Background(), []string{"NOTREAL"}, "", false)
	require.Error(t, err)
	require.Nil(t, quotes)
	require.Nil(t, quoteErr)
	assert.Contains(t, err.Error(), "decode quote error:")
}

func TestGetQuote(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/$SPX/quotes", r.URL.Path)
		assert.Equal(t, "quote,reference", r.URL.Query().Get("fields"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]QuoteEntry{
			"$SPX": indexQuoteEntry("$SPX"),
		})
	})

	quotes, err := client.GetQuote(context.Background(), "$SPX", "quote,reference")
	require.NoError(t, err)
	require.NotNil(t, quotes)

	entry := (*quotes)["$SPX"]
	require.NotNil(t, entry)
	assert.Equal(t, schwab.AssetTypeIndex, entry.AssetMainType)
	assert.Equal(t, "$SPX", entry.Symbol)

	indexQuote, err := entry.IndexQuote()
	require.NoError(t, err)
	assert.InDelta(t, 5123.45, indexQuote.LastPrice, 0.000001)
	assert.Equal(t, "CBOE", indexQuote.Exchange)
}

func TestGetQuote_Error(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		writeJSON(t, w, map[string]string{"detail": "symbol not found"})
	})

	quotes, err := client.GetQuote(context.Background(), "NOTREAL", "")
	require.Error(t, err)
	require.Nil(t, quotes)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	assert.Equal(t, "symbol not found", apiErr.Message)
}

func TestGetQuotesError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, map[string]string{
			"detail": "invalid fields parameter",
		})
	})

	quotes, quoteErr, err := client.GetQuotes(context.Background(), []string{"AAPL"}, "invalid", false)
	require.Error(t, err)
	require.Nil(t, quotes)
	require.Nil(t, quoteErr)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
	assert.Equal(t, "invalid fields parameter", apiErr.Message)
}

func TestQuoteWrongAssetType(t *testing.T) {
	// Use an equity entry for all tests so the AssetMainType is always wrong.
	entry := equityQuoteEntry("AAPL")

	tests := []struct {
		name    string
		call    func() (any, error)
		wantErr string
	}{
		{
			"OptionQuote",
			func() (any, error) { return entry.OptionQuote() },
			"quote is EQUITY, not OPTION",
		},
		{
			"OptionReference",
			func() (any, error) { return entry.OptionReference() },
			"quote is EQUITY, not OPTION",
		},
		{
			"IndexQuote",
			func() (any, error) { return entry.IndexQuote() },
			"quote is EQUITY, not INDEX",
		},
		{
			"IndexReference",
			func() (any, error) { return entry.IndexReference() },
			"quote is EQUITY, not INDEX",
		},
		{
			"MutualFundQuote",
			func() (any, error) { return entry.MutualFundQuote() },
			"quote is EQUITY, not MUTUAL_FUND",
		},
		{
			"MutualFundReference",
			func() (any, error) { return entry.MutualFundReference() },
			"quote is EQUITY, not MUTUAL_FUND",
		},
		{
			"ForexQuote",
			func() (any, error) { return entry.ForexQuote() },
			"quote is EQUITY, not FOREX",
		},
		{
			"ForexReference",
			func() (any, error) { return entry.ForexReference() },
			"quote is EQUITY, not FOREX",
		},
		{
			"FutureQuote",
			func() (any, error) { return entry.FutureQuote() },
			"quote is EQUITY, not FUTURE",
		},
		{
			"FutureReference",
			func() (any, error) { return entry.FutureReference() },
			"quote is EQUITY, not FUTURE",
		},
		{
			"FutureOptionQuote",
			func() (any, error) { return entry.FutureOptionQuote() },
			"quote is EQUITY, not FUTURE_OPTION",
		},
		{
			"FutureOptionReference",
			func() (any, error) { return entry.FutureOptionReference() },
			"quote is EQUITY, not FUTURE_OPTION",
		},
	}

	// Also test EquityQuote/EquityReference with a wrong type (INDEX entry).
	indexEntry := indexQuoteEntry("$DJI")
	tests = append(tests,
		struct {
			name    string
			call    func() (any, error)
			wantErr string
		}{
			"EquityQuote",
			func() (any, error) { return indexEntry.EquityQuote() },
			"quote is INDEX, not EQUITY",
		},
		struct {
			name    string
			call    func() (any, error)
			wantErr string
		}{
			"EquityReference",
			func() (any, error) { return indexEntry.EquityReference() },
			"quote is INDEX, not EQUITY",
		},
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.call()
			require.Error(t, err)
			require.Nil(t, result)
			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestQuoteDecodeError(t *testing.T) {
	invalidJSON := json.RawMessage(`{not valid json`)

	tests := []struct {
		name string
		call func() (any, error)
	}{
		{"EquityQuote", func() (any, error) {
			e := equityQuoteEntry("AAPL")
			e.Quote = invalidJSON
			return e.EquityQuote()
		}},
		{"EquityReference", func() (any, error) {
			e := equityQuoteEntry("AAPL")
			e.Reference = invalidJSON
			return e.EquityReference()
		}},
		{"OptionQuote", func() (any, error) {
			e := optionQuoteEntry("AAPL_C170")
			e.Quote = invalidJSON
			return e.OptionQuote()
		}},
		{"OptionReference", func() (any, error) {
			e := optionQuoteEntry("AAPL_C170")
			e.Reference = invalidJSON
			return e.OptionReference()
		}},
		{"IndexQuote", func() (any, error) {
			e := indexQuoteEntry("$SPX")
			e.Quote = invalidJSON
			return e.IndexQuote()
		}},
		{"IndexReference", func() (any, error) {
			e := indexQuoteEntry("$SPX")
			e.Reference = invalidJSON
			return e.IndexReference()
		}},
		{"MutualFundQuote", func() (any, error) {
			e := mutualFundQuoteEntry("VFIAX")
			e.Quote = invalidJSON
			return e.MutualFundQuote()
		}},
		{"MutualFundReference", func() (any, error) {
			e := mutualFundQuoteEntry("VFIAX")
			e.Reference = invalidJSON
			return e.MutualFundReference()
		}},
		{"ForexQuote", func() (any, error) {
			e := forexQuoteEntry("EUR/USD")
			e.Quote = invalidJSON
			return e.ForexQuote()
		}},
		{"ForexReference", func() (any, error) {
			e := forexQuoteEntry("EUR/USD")
			e.Reference = invalidJSON
			return e.ForexReference()
		}},
		{"FutureQuote", func() (any, error) {
			e := futureQuoteEntry("/ES")
			e.Quote = invalidJSON
			return e.FutureQuote()
		}},
		{"FutureReference", func() (any, error) {
			e := futureQuoteEntry("/ES")
			e.Reference = invalidJSON
			return e.FutureReference()
		}},
		{"FutureOptionQuote", func() (any, error) {
			e := futureOptionQuoteEntry("./EW1H24C5000")
			e.Quote = invalidJSON
			return e.FutureOptionQuote()
		}},
		{"FutureOptionReference", func() (any, error) {
			e := futureOptionQuoteEntry("./EW1H24C5000")
			e.Reference = invalidJSON
			return e.FutureOptionReference()
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.call()
			require.Error(t, err)
			require.Nil(t, result)
			assert.Contains(t, err.Error(), "decode quote payload:")
		})
	}
}

func TestIndexReferenceHappyPath(t *testing.T) {
	entry := QuoteEntry{
		AssetMainType: schwab.AssetTypeIndex,
		Symbol:        "$SPX",
		Reference: mustRaw(IndexReference{
			Description:  "S&P 500 Index",
			Exchange:     "IND",
			ExchangeName: "Indices",
		}),
	}

	ref, err := entry.IndexReference()
	require.NoError(t, err)
	assert.Equal(t, "S&P 500 Index", ref.Description)
	assert.Equal(t, "IND", ref.Exchange)
}

func TestMutualFundReferenceHappyPath(t *testing.T) {
	entry := QuoteEntry{
		AssetMainType: schwab.AssetTypeMutualFund,
		Symbol:        "VFIAX",
		Reference: mustRaw(MutualFundReference{
			CUSIP:        "922908710",
			Description:  "Vanguard 500 Index Fund",
			Exchange:     "NASDAQ",
			ExchangeName: "NASDAQ Mutual Funds",
		}),
	}

	ref, err := entry.MutualFundReference()
	require.NoError(t, err)
	assert.Equal(t, "922908710", ref.CUSIP)
	assert.Equal(t, "Vanguard 500 Index Fund", ref.Description)
}

func equityQuoteEntry(symbol string) QuoteEntry {
	return QuoteEntry{
		AssetMainType: schwab.AssetTypeEquity,
		AssetSubType:  "COE",
		QuoteType:     QuoteTypeNBBO,
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
		ContractType:     OptionContractTypeCall,
		CUSIP:            "037833100",
		DaysToExpiration: 30,
		Description:      "AAPL 05/24/2024 $170 Call",
		Exchange:         "OPR",
		ExchangeName:     "Options Price Reporting Authority",
		ExerciseType:     OptionExerciseTypeAmerican,
		ExpirationDay:    24,
		ExpirationMonth:  5,
		ExpirationType:   OptionExpirationTypeStandard,
		ExpirationYear:   2024,
		Multiplier:       100,
		SettlementType:   OptionSettlementTypePM,
		StrikePrice:      170,
		Underlying:       "AAPL",
		PennyPilot:       true,
		LastTradingDay:   1716508800000,
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
		"indAskPrice":            4.26,
		"indBidPrice":            4.14,
		"indQuoteTime":           1712345678998,
		"moneyIntrinsicValue":    1.25,
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
		Reference: mustRaw(ForexReference{
			Description:  "Euro/US Dollar",
			Exchange:     "FOREX",
			ExchangeName: "Foreign Exchange",
			Tradable:     true,
			MarketMaker:  "FX",
			Product:      "EUR/USD",
			TradingHours: "24x5",
		}),
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

func futureOptionQuoteEntry(symbol string) QuoteEntry {
	return QuoteEntry{
		AssetMainType: schwab.AssetTypeFutureOption,
		Symbol:        symbol,
		Reference: mustRaw(FutureOptionReference{
			Description:     "E-mini S&P 500 Weekly Call",
			ContractType:    OptionContractTypeCall,
			Exchange:        "CME",
			ExchangeName:    "Chicago Mercantile Exchange",
			ExpirationDate:  20240315,
			ExpirationStyle: "American",
			Multiplier:      50,
			StrikePrice:     5000,
			Underlying:      "/ES",
		}),
		Quote: mustRaw(map[string]any{
			"askPrice":        12.5,
			"bidPrice":        12.0,
			"lastPrice":       12.25,
			"highPrice":       13.0,
			"lowPrice":        11.5,
			"mark":            12.25,
			"netChange":       0.5,
			"openInterest":    1500,
			"settlemetPrice":  12.25,
			"settlementPrice": 12.3,
			"totalVolume":     25,
			"tradeTime":       1712345678901,
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
