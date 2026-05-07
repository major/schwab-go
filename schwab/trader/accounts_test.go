package trader

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

func TestGetAccounts(t *testing.T) {
	fixture := []Account{
		{
			SecuritiesAccount: SecuritiesAccount{
				Type:                    "CASH",
				AccountNumber:           "123456789",
				RoundTrips:              0,
				IsDayTrader:             false,
				IsClosingOnlyRestricted: false,
				IsForeign:               true,
				PfcbFlag:                false,
				Positions: []Position{
					{
						ShortQuantity:                  0,
						AveragePrice:                   150.00,
						CurrentDayProfitLoss:           300.00,
						CurrentDayProfitLossPercentage: 2.0,
						LongQuantity:                   10,
						SettledLongQuantity:            10,
						SettledShortQuantity:           0,
						Instrument: AccountInstrument{
							AssetType:    "EQUITY",
							Cusip:        "037833100",
							Symbol:       "AAPL",
							Description:  "Apple Inc",
							InstrumentID: 1234567,
							NetChange:    1.5,
						},
						MarketValue:                 1530.00,
						MaintenanceRequirement:      459.00,
						AverageLongPrice:            150.00,
						TaxLotAverageLongPrice:      150.00,
						LongOpenProfitLoss:          30.00,
						PreviousSessionLongQuantity: 10,
						CurrentDayCost:              0,
					},
				},
				CurrentBalances: Balance{
					CashBalance:      5000.00,
					Equity:           6530.00,
					LiquidationValue: 6530.00,
					BuyingPower:      5000.00,
					AccountValue:     6530.00,
				},
			},
		},
	}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/accounts", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, fixture)
	})

	result, err := client.GetAccounts(context.Background(), "")
	require.NoError(t, err)
	require.Len(t, result, 1)

	acct := result[0].SecuritiesAccount
	assert.Equal(t, "CASH", acct.Type)
	assert.Equal(t, "123456789", acct.AccountNumber)
	assert.Equal(t, 0, acct.RoundTrips)
	assert.False(t, acct.IsDayTrader)
	assert.False(t, acct.IsClosingOnlyRestricted)
	assert.True(t, acct.IsForeign)
	assert.False(t, acct.PfcbFlag)

	// Verify position
	require.Len(t, acct.Positions, 1)
	pos := acct.Positions[0]
	assert.InDelta(t, 0.0, pos.ShortQuantity, 0.000001)
	assert.InDelta(t, 150.00, pos.AveragePrice, 0.000001)
	assert.InDelta(t, 300.00, pos.CurrentDayProfitLoss, 0.000001)
	assert.InDelta(t, 2.0, pos.CurrentDayProfitLossPercentage, 0.000001)
	assert.InDelta(t, 10.0, pos.LongQuantity, 0.000001)
	assert.InDelta(t, 10.0, pos.SettledLongQuantity, 0.000001)
	assert.InDelta(t, 0.0, pos.SettledShortQuantity, 0.000001)
	assert.InDelta(t, 1530.00, pos.MarketValue, 0.000001)
	assert.InDelta(t, 459.00, pos.MaintenanceRequirement, 0.000001)
	assert.InDelta(t, 150.00, pos.AverageLongPrice, 0.000001)
	assert.InDelta(t, 150.00, pos.TaxLotAverageLongPrice, 0.000001)
	assert.InDelta(t, 30.00, pos.LongOpenProfitLoss, 0.000001)
	assert.InDelta(t, 10.0, pos.PreviousSessionLongQuantity, 0.000001)
	assert.InDelta(t, 0.0, pos.CurrentDayCost, 0.000001)

	// Verify instrument
	inst := pos.Instrument
	assert.Equal(t, schwab.AssetType("EQUITY"), inst.AssetType)
	assert.Equal(t, "037833100", inst.Cusip)
	assert.Equal(t, "AAPL", inst.Symbol)
	assert.Equal(t, "Apple Inc", inst.Description)
	assert.Equal(t, int64(1234567), inst.InstrumentID)
	assert.InDelta(t, 1.5, inst.NetChange, 0.000001)

	// Verify balances
	bal := acct.CurrentBalances
	assert.InDelta(t, 5000.00, bal.CashBalance, 0.000001)
	assert.InDelta(t, 6530.00, bal.Equity, 0.000001)
	assert.InDelta(t, 6530.00, bal.LiquidationValue, 0.000001)
	assert.InDelta(t, 5000.00, bal.BuyingPower, 0.000001)
	assert.InDelta(t, 6530.00, bal.AccountValue, 0.000001)
}

func TestGetAccounts_WithFields(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/accounts", r.URL.Path)

		// Verify fields query param is present
		fields := r.URL.Query().Get("fields")
		assert.Equal(t, "positions", fields)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, []Account{})
	})

	_, err := client.GetAccounts(context.Background(), "positions")
	require.NoError(t, err)
}

func TestGetAccountsRaw_PreservesResponseJSON(t *testing.T) {
	payload := []map[string]any{
		{
			"securitiesAccount": map[string]any{
				"accountNumber": "123456789",
				"isForeign":     true,
			},
		},
	}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/accounts", r.URL.Path)
		assert.Equal(t, "positions", r.URL.Query().Get("fields"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, payload)
	})

	result, err := client.GetAccountsRaw(context.Background(), "positions")
	require.NoError(t, err)

	var got []map[string]any
	require.NoError(t, json.Unmarshal(result, &got))
	require.Equal(t, payload, got)
}

func TestGetAccountsRaw_NoFields(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/accounts", r.URL.Path)
		assert.Empty(t, r.URL.Query().Get("fields"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, []map[string]any{})
	})

	result, err := client.GetAccountsRaw(context.Background(), "")
	require.NoError(t, err)
	assert.JSONEq(t, `[]`, string(result))
}

func TestGetAccountsRaw_PreservesOmittedAndZeroFields(t *testing.T) {
	payload := json.RawMessage(`[
		{
			"securitiesAccount": {
				"accountNumber": "123456789",
				"roundTrips": 0,
				"isDayTrader": false,
				"currentBalances": {
					"cashBalance": 0
				}
			}
		},
		{
			"securitiesAccount": {
				"accountNumber": "987654321",
				"currentBalances": {}
			}
		}
	]`)

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/accounts", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(payload)
		assert.NoError(t, err)
	})

	result, err := client.GetAccountsRaw(context.Background(), "")
	require.NoError(t, err)

	var got []map[string]map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(result, &got))
	require.Len(t, got, 2)
	assert.Contains(t, got[0]["securitiesAccount"], "roundTrips")
	assert.Contains(t, got[0]["securitiesAccount"], "isDayTrader")
	assert.Contains(t, got[0]["securitiesAccount"], "currentBalances")
	assert.NotContains(t, got[1]["securitiesAccount"], "roundTrips")
	assert.NotContains(t, got[1]["securitiesAccount"], "isDayTrader")

	var explicitBalances map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(got[0]["securitiesAccount"]["currentBalances"], &explicitBalances))
	assert.Contains(t, explicitBalances, "cashBalance")

	var omittedBalances map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(got[1]["securitiesAccount"]["currentBalances"], &omittedBalances))
	assert.NotContains(t, omittedBalances, "cashBalance")
}

func TestGetAccountsRaw_Error(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	_, err := client.GetAccountsRaw(context.Background(), "")
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
}

func TestGetAccounts_NoFields(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/accounts", r.URL.Path)

		// Verify fields query param is absent
		fields := r.URL.Query().Get("fields")
		assert.Empty(t, fields)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, []Account{})
	})

	_, err := client.GetAccounts(context.Background(), "")
	require.NoError(t, err)
}

func TestGetAccounts_Error(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	_, err := client.GetAccounts(context.Background(), "")
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
}

func TestGetAccount(t *testing.T) {
	fixture := Account{
		SecuritiesAccount: SecuritiesAccount{
			Type:          "MARGIN",
			AccountNumber: "987654321",
			RoundTrips:    3,
			IsDayTrader:   true,
			IsForeign:     true,
			CurrentBalances: Balance{
				CashBalance:  10000.00,
				Equity:       25000.00,
				BuyingPower:  20000.00,
				AccountValue: 25000.00,
			},
		},
	}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify path contains the hash value
		assert.Equal(t, "/accounts/HASH_ABC123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, fixture)
	})

	result, err := client.GetAccount(context.Background(), "HASH_ABC123", "")
	require.NoError(t, err)
	require.NotNil(t, result)

	acct := result.SecuritiesAccount
	assert.Equal(t, "MARGIN", acct.Type)
	assert.Equal(t, "987654321", acct.AccountNumber)
	assert.Equal(t, 3, acct.RoundTrips)
	assert.True(t, acct.IsDayTrader)
	assert.True(t, acct.IsForeign)
	assert.InDelta(t, 10000.00, acct.CurrentBalances.CashBalance, 0.000001)
	assert.InDelta(t, 25000.00, acct.CurrentBalances.Equity, 0.000001)
	assert.InDelta(t, 20000.00, acct.CurrentBalances.BuyingPower, 0.000001)
	assert.InDelta(t, 25000.00, acct.CurrentBalances.AccountValue, 0.000001)
}

func TestGetAccount_WithFields(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/accounts/HASH_ABC123", r.URL.Path)
		assert.Equal(t, "positions", r.URL.Query().Get("fields"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, Account{
			SecuritiesAccount: SecuritiesAccount{AccountNumber: "123"},
		})
	})

	result, err := client.GetAccount(context.Background(), "HASH_ABC123", "positions")
	require.NoError(t, err)
	assert.Equal(t, "123", result.SecuritiesAccount.AccountNumber)
}

func TestGetAccountRaw_PreservesResponseJSON(t *testing.T) {
	payload := map[string]any{
		"securitiesAccount": map[string]any{
			"accountNumber": "987654321",
			"isForeign":     true,
		},
	}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/accounts/HASH_ABC123", r.URL.Path)
		assert.Equal(t, "positions", r.URL.Query().Get("fields"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, payload)
	})

	result, err := client.GetAccountRaw(context.Background(), "HASH_ABC123", "positions")
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(result, &got))
	require.Equal(t, payload, got)
}

func TestGetAccountRaw_NoFields(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/accounts/HASH_ABC123", r.URL.Path)
		assert.Empty(t, r.URL.Query().Get("fields"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, map[string]any{"securitiesAccount": map[string]any{"accountNumber": "123"}})
	})

	result, err := client.GetAccountRaw(context.Background(), "HASH_ABC123", "")
	require.NoError(t, err)
	assert.JSONEq(t, `{"securitiesAccount":{"accountNumber":"123"}}`, string(result))
}

func TestGetAccountRaw_Error(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := client.GetAccountRaw(context.Background(), "HASH_ABC123", "")
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestGetAccount_Error(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	_, err := client.GetAccount(context.Background(), "HASH_ABC123", "")
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
}
