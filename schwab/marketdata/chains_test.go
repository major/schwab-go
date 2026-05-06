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

func TestGetOptionChain(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/chains", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		query := r.URL.Query()
		assert.Equal(t, "SPY", query.Get("symbol"))
		assert.Equal(t, "ALL", query.Get("contractType"))
		assert.Equal(t, "2", query.Get("strikeCount"))
		assert.Equal(t, "true", query.Get("includeUnderlyingQuote"))
		assert.Equal(t, "SINGLE", query.Get("strategy"))
		assert.Equal(t, "NTM", query.Get("range"))
		assert.Equal(t, "2024-01-01", query.Get("fromDate"))
		assert.Equal(t, "2024-02-01", query.Get("toDate"))
		assert.Equal(t, "JAN", query.Get("expMonth"))
		assert.Equal(t, "ALL", query.Get("optionType"))
		assert.Equal(t, "PN", query.Get("entitlement"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, optionChainFixture())
	})

	params := &OptionChainParams{
		Symbol:                 "SPY",
		ContractType:           OptionChainContractTypeAll,
		StrikeCount:            2,
		IncludeUnderlyingQuote: true,
		Strategy:               OptionChainStrategySingle,
		Range:                  OptionChainRangeNearTheMoney,
		FromDate:               "2024-01-01",
		ToDate:                 "2024-02-01",
		ExpMonth:               ExpirationMonthJanuary,
		OptionType:             OptionChainTypeAll,
		Entitlement:            OptionEntitlementPayingNonProfessional,
	}

	result, err := client.GetOptionChain(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "SPY", result.Symbol)
	assert.Equal(t, "SUCCESS", result.Status)
	assert.Equal(t, OptionChainStrategySingle, result.Strategy)
	assert.InDelta(t, 0.0, result.Interval, 0.000001)
	assert.True(t, result.IsDelayed)
	assert.False(t, result.IsIndex)
	assert.InDelta(t, 5.0, result.DaysToExpiration, 0.000001)
	assert.InDelta(t, 5.25, result.InterestRate, 0.000001)
	assert.InDelta(t, 470.50, result.UnderlyingPrice, 0.000001)
	assert.InDelta(t, 0.19, result.Volatility, 0.000001)

	require.NotNil(t, result.Underlying)
	assert.Equal(t, "SPY", result.Underlying.Symbol)
	assert.InDelta(t, 470.50, result.Underlying.Last, 0.000001)
	assert.InDelta(t, 470.40, result.Underlying.Bid, 0.000001)
	assert.InDelta(t, 470.60, result.Underlying.Ask, 0.000001)
	assert.Equal(t, int64(54321000), result.Underlying.TotalVolume)

	callContract := result.CallExpDateMap["2024-01-19:5"]["470.0"][0]
	assert.Equal(t, OptionChainContractTypeCall, callContract.PutCall)
	assert.Equal(t, "SPY_011924C470", callContract.Symbol)
	assert.InDelta(t, 3.40, callContract.BidPrice, 0.000001)
	assert.InDelta(t, 3.45, callContract.AskPrice, 0.000001)
	assert.InDelta(t, 3.43, callContract.MarkPrice, 0.000001)
	assert.Equal(t, int64(12345), callContract.TotalVolume)
	assert.Equal(t, "20240119", callContract.TradeDate)
	assert.InDelta(t, 0.52, callContract.Delta, 0.000001)
	assert.InDelta(t, 0.08, callContract.Gamma, 0.000001)
	assert.InDelta(t, -0.12, callContract.Theta, 0.000001)
	assert.InDelta(t, 0.21, callContract.Vega, 0.000001)
	assert.InDelta(t, 0.03, callContract.Rho, 0.000001)
	assert.InDelta(t, 470.0, callContract.StrikePrice, 0.000001)
	assert.True(t, callContract.InTheMoney)
	assert.True(t, callContract.PennyPilot)

	putContract := result.PutExpDateMap["2024-01-19:5"]["470.0"][0]
	assert.Equal(t, OptionChainContractTypePut, putContract.PutCall)
	assert.Equal(t, "SPY_011924P470", putContract.Symbol)
	assert.InDelta(t, -0.48, putContract.Delta, 0.000001)
	assert.InDelta(t, 1.95, putContract.IntrinsicValue, 0.000001)
}

func TestOptionContractTradeDateStringUnmarshal(t *testing.T) {
	var contract OptionContract
	err := json.Unmarshal([]byte(`{"tradeDate":"2024-01-19"}`), &contract)
	require.NoError(t, err)
	assert.Equal(t, "2024-01-19", contract.TradeDate)
}

func TestGetOptionChainMinimal(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/chains", r.URL.Path)

		query := r.URL.Query()
		assert.Equal(t, "AAPL", query.Get("symbol"))
		assertParamAbsent(t, r, "contractType")
		assertParamAbsent(t, r, "strikeCount")
		assertParamAbsent(t, r, "includeUnderlyingQuote")
		assertParamAbsent(t, r, "strategy")
		assertParamAbsent(t, r, "interval")
		assertParamAbsent(t, r, "strike")
		assertParamAbsent(t, r, "range")
		assertParamAbsent(t, r, "fromDate")
		assertParamAbsent(t, r, "toDate")
		assertParamAbsent(t, r, "volatility")
		assertParamAbsent(t, r, "underlyingPrice")
		assertParamAbsent(t, r, "interestRate")
		assertParamAbsent(t, r, "daysToExpiration")
		assertParamAbsent(t, r, "expMonth")
		assertParamAbsent(t, r, "optionType")
		assertParamAbsent(t, r, "entitlement")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]any{
			"symbol":         "AAPL",
			"status":         "SUCCESS",
			"callExpDateMap": map[string]any{},
			"putExpDateMap":  map[string]any{},
		})
	})

	result, err := client.GetOptionChain(context.Background(), &OptionChainParams{Symbol: "AAPL"})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "AAPL", result.Symbol)
}

func TestGetOptionChainRequiresSymbol(t *testing.T) {
	client := NewClient()

	_, err := client.GetOptionChain(context.Background(), nil)
	require.EqualError(t, err, "symbol is required")

	_, err = client.GetOptionChain(context.Background(), &OptionChainParams{})
	require.EqualError(t, err, "symbol is required")
}

func TestGetOptionChainAnalytical(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/chains", r.URL.Path)

		query := r.URL.Query()
		assert.Equal(t, "MSFT", query.Get("symbol"))
		assert.Equal(t, "ANALYTICAL", query.Get("strategy"))
		assert.Equal(t, "5.5", query.Get("interval"))
		assert.Equal(t, "420.25", query.Get("strike"))
		assert.Equal(t, "0.24", query.Get("volatility"))
		assert.Equal(t, "421.75", query.Get("underlyingPrice"))
		assert.Equal(t, "5.125", query.Get("interestRate"))
		assert.Equal(t, "30", query.Get("daysToExpiration"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]any{
			"symbol":           "MSFT",
			"status":           "SUCCESS",
			"strategy":         "ANALYTICAL",
			"interval":         5.5,
			"volatility":       0.24,
			"underlyingPrice":  421.75,
			"interestRate":     5.125,
			"daysToExpiration": 30.0,
			"callExpDateMap":   map[string]any{},
			"putExpDateMap":    map[string]any{},
		})
	})

	params := &OptionChainParams{
		Symbol:           "MSFT",
		Strategy:         OptionChainStrategyAnalytical,
		Interval:         5.5,
		Strike:           420.25,
		Volatility:       0.24,
		UnderlyingPrice:  421.75,
		InterestRate:     5.125,
		DaysToExpiration: 30,
	}

	result, err := client.GetOptionChain(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, OptionChainStrategyAnalytical, result.Strategy)
	assert.InDelta(t, 5.5, result.Interval, 0.000001)
	assert.InDelta(t, 0.24, result.Volatility, 0.000001)
	assert.InDelta(t, 421.75, result.UnderlyingPrice, 0.000001)
	assert.InDelta(t, 5.125, result.InterestRate, 0.000001)
	assert.InDelta(t, 30.0, result.DaysToExpiration, 0.000001)
}

func TestGetOptionChainEmpty(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/chains", r.URL.Path)
		assert.Equal(t, "TSLA", r.URL.Query().Get("symbol"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]any{
			"symbol":         "TSLA",
			"status":         "SUCCESS",
			"callExpDateMap": map[string]any{},
			"putExpDateMap":  map[string]any{},
		})
	})

	result, err := client.GetOptionChain(context.Background(), &OptionChainParams{Symbol: "TSLA"})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.CallExpDateMap)
	require.NotNil(t, result.PutExpDateMap)
	assert.Empty(t, result.CallExpDateMap)
	assert.Empty(t, result.PutExpDateMap)
}

func TestGetOptionChainError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/chains", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, map[string]any{
			"detail": "invalid option chain request",
		})
	})

	_, err := client.GetOptionChain(context.Background(), &OptionChainParams{Symbol: "INVALID"})
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
	assert.Equal(t, "invalid option chain request", apiErr.Message)
}

func assertParamAbsent(t *testing.T, r *http.Request, name string) {
	t.Helper()
	_, ok := r.URL.Query()[name]
	assert.False(t, ok, "expected %q query param to be absent", name)
}

func optionChainFixture() map[string]any {
	return map[string]any{
		"symbol":           "SPY",
		"status":           "SUCCESS",
		"strategy":         "SINGLE",
		"interval":         0.0,
		"isDelayed":        true,
		"isIndex":          false,
		"daysToExpiration": 5.0,
		"interestRate":     5.25,
		"underlyingPrice":  470.50,
		"volatility":       0.19,
		"callExpDateMap": map[string]any{
			"2024-01-19:5": map[string]any{
				"470.0": []map[string]any{optionContractFixture("CALL", "SPY_011924C470", 0.52, 0.0)},
			},
		},
		"putExpDateMap": map[string]any{
			"2024-01-19:5": map[string]any{
				"470.0": []map[string]any{optionContractFixture("PUT", "SPY_011924P470", -0.48, 1.95)},
			},
		},
		"underlying": map[string]any{
			"ask":               470.60,
			"askSize":           100,
			"bid":               470.40,
			"bidSize":           100,
			"change":            1.25,
			"close":             469.25,
			"delayed":           true,
			"description":       "SPDR S&P 500 ETF TRUST",
			"exchangeName":      "NYSE Arca",
			"fiftyTwoWeekHigh":  480.00,
			"fiftyTwoWeekLow":   380.00,
			"highPrice":         471.00,
			"last":              470.50,
			"lowPrice":          468.00,
			"mark":              470.50,
			"markChange":        1.25,
			"markPercentChange": 0.27,
			"openPrice":         469.00,
			"percentChange":     0.27,
			"quoteTime":         int64(1705276800000),
			"symbol":            "SPY",
			"totalVolume":       int64(54321000),
			"tradeTime":         int64(1705276801000),
		},
	}
}

func optionContractFixture(putCall, symbol string, delta, intrinsicValue float64) map[string]any {
	return map[string]any{
		"putCall":                putCall,
		"symbol":                 symbol,
		"description":            symbol + " description",
		"exchangeName":           "OPR",
		"bidPrice":               3.40,
		"askPrice":               3.45,
		"lastPrice":              3.42,
		"markPrice":              3.43,
		"bidSize":                25,
		"askSize":                30,
		"lastSize":               5,
		"highPrice":              4.10,
		"lowPrice":               2.95,
		"openPrice":              3.10,
		"closePrice":             3.25,
		"totalVolume":            int64(12345),
		"tradeDate":              int64(20240119),
		"tradeTimeInLong":        int64(1705276800000),
		"quoteTimeInLong":        int64(1705276805000),
		"netChange":              0.17,
		"volatility":             0.22,
		"delta":                  delta,
		"gamma":                  0.08,
		"theta":                  -0.12,
		"vega":                   0.21,
		"rho":                    0.03,
		"openInterest":           int64(67890),
		"timeValue":              1.48,
		"theoreticalOptionValue": 3.44,
		"theoreticalVolatility":  0.23,
		"strikePrice":            470.0,
		"expirationDate":         "2024-01-19T21:00:00.000+00:00",
		"daysToExpiration":       5,
		"expirationType":         "R",
		"lastTradingDay":         int64(1705708800000),
		"multiplier":             100.0,
		"settlementType":         "P",
		"isInTheMoney":           true,
		"isNonStandard":          false,
		"isMini":                 false,
		"isPennyPilot":           true,
		"intrinsicValue":         intrinsicValue,
		"extrinsicValue":         1.48,
		"impliedYield":           0.0,
	}
}
