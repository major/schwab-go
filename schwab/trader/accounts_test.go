package trader

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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
							InstrumentId: 1234567,
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

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/accounts", r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, fixture)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithToken("test-token"),
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	result, err := client.GetAccounts(context.Background(), "")
	require.NoError(t, err)
	require.Len(t, result, 1)

	acct := result[0].SecuritiesAccount
	assert.Equal(t, "CASH", acct.Type)
	assert.Equal(t, "123456789", acct.AccountNumber)
	assert.Equal(t, 0, acct.RoundTrips)
	assert.False(t, acct.IsDayTrader)
	assert.False(t, acct.IsClosingOnlyRestricted)
	assert.False(t, acct.PfcbFlag)

	// Verify position
	require.Len(t, acct.Positions, 1)
	pos := acct.Positions[0]
	assert.Equal(t, 0.0, pos.ShortQuantity)
	assert.Equal(t, 150.00, pos.AveragePrice)
	assert.Equal(t, 300.00, pos.CurrentDayProfitLoss)
	assert.Equal(t, 2.0, pos.CurrentDayProfitLossPercentage)
	assert.Equal(t, 10.0, pos.LongQuantity)
	assert.Equal(t, 10.0, pos.SettledLongQuantity)
	assert.Equal(t, 0.0, pos.SettledShortQuantity)
	assert.Equal(t, 1530.00, pos.MarketValue)
	assert.Equal(t, 459.00, pos.MaintenanceRequirement)
	assert.Equal(t, 150.00, pos.AverageLongPrice)
	assert.Equal(t, 150.00, pos.TaxLotAverageLongPrice)
	assert.Equal(t, 30.00, pos.LongOpenProfitLoss)
	assert.Equal(t, 10.0, pos.PreviousSessionLongQuantity)
	assert.Equal(t, 0.0, pos.CurrentDayCost)

	// Verify instrument
	inst := pos.Instrument
	assert.Equal(t, schwab.AssetType("EQUITY"), inst.AssetType)
	assert.Equal(t, "037833100", inst.Cusip)
	assert.Equal(t, "AAPL", inst.Symbol)
	assert.Equal(t, "Apple Inc", inst.Description)
	assert.Equal(t, int64(1234567), inst.InstrumentId)
	assert.Equal(t, 1.5, inst.NetChange)

	// Verify balances
	bal := acct.CurrentBalances
	assert.Equal(t, 5000.00, bal.CashBalance)
	assert.Equal(t, 6530.00, bal.Equity)
	assert.Equal(t, 6530.00, bal.LiquidationValue)
	assert.Equal(t, 5000.00, bal.BuyingPower)
	assert.Equal(t, 6530.00, bal.AccountValue)
}

func TestGetAccounts_WithFields(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/accounts", r.URL.Path)

		// Verify fields query param is present
		fields := r.URL.Query().Get("fields")
		require.Equal(t, "positions", fields)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, []Account{})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetAccounts(context.Background(), "positions")
	require.NoError(t, err)
}

func TestGetAccounts_NoFields(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/accounts", r.URL.Path)

		// Verify fields query param is absent
		fields := r.URL.Query().Get("fields")
		require.Empty(t, fields)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, []Account{})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetAccounts(context.Background(), "")
	require.NoError(t, err)
}

func TestGetAccount(t *testing.T) {
	fixture := Account{
		SecuritiesAccount: SecuritiesAccount{
			Type:          "MARGIN",
			AccountNumber: "987654321",
			RoundTrips:    3,
			IsDayTrader:   true,
			CurrentBalances: Balance{
				CashBalance:  10000.00,
				Equity:       25000.00,
				BuyingPower:  20000.00,
				AccountValue: 25000.00,
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify path contains the hash value
		require.Equal(t, "/accounts/HASH_ABC123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, fixture)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithToken("test-token"),
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	result, err := client.GetAccount(context.Background(), "HASH_ABC123", "")
	require.NoError(t, err)
	require.NotNil(t, result)

	acct := result.SecuritiesAccount
	assert.Equal(t, "MARGIN", acct.Type)
	assert.Equal(t, "987654321", acct.AccountNumber)
	assert.Equal(t, 3, acct.RoundTrips)
	assert.True(t, acct.IsDayTrader)
	assert.Equal(t, 10000.00, acct.CurrentBalances.CashBalance)
	assert.Equal(t, 25000.00, acct.CurrentBalances.Equity)
	assert.Equal(t, 20000.00, acct.CurrentBalances.BuyingPower)
	assert.Equal(t, 25000.00, acct.CurrentBalances.AccountValue)
}

func TestGetAccount_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetAccount(context.Background(), "HASH_ABC123", "")
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
}
