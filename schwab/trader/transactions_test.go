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

func TestGetTransactions(t *testing.T) {
	fixture := []Transaction{
		{
			ActivityID:    1001,
			Time:          "2024-01-15T10:30:00Z",
			Type:          TransactionTypeTrade,
			Status:        "VALID",
			SubAccount:    "CASH",
			TradeDate:     "2024-01-15",
			PositionID:    5001,
			OrderID:       9001,
			NetAmount:     -1500.00,
			Description:   "Buy 10 AAPL",
			AccountNumber: "123456789",
			TransferItems: []TransferItem{
				{
					Instrument: TransactionInstrument{
						AssetType:    schwab.AssetTypeEquity,
						Cusip:        "037833100",
						Symbol:       "AAPL",
						Description:  "Apple Inc",
						InstrumentID: 1234567,
					},
					Amount: 10,
					Cost:   1500.00,
					Price:  150.00,
				},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/accounts/HASH_ABC123/transactions", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify required query params
		assert.Equal(t, "2024-01-01", r.URL.Query().Get("startDate"))
		assert.Equal(t, "2024-01-31", r.URL.Query().Get("endDate"))
		assert.Equal(t, "TRADE", r.URL.Query().Get("types"))

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

	result, err := client.GetTransactions(context.Background(), "HASH_ABC123", &TransactionListParams{
		StartDate: "2024-01-01",
		EndDate:   "2024-01-31",
		Types:     "TRADE",
	})
	require.NoError(t, err)
	require.Len(t, result, 1)

	txn := result[0]
	assert.Equal(t, int64(1001), txn.ActivityID)
	assert.Equal(t, "2024-01-15T10:30:00Z", txn.Time)
	assert.Equal(t, TransactionTypeTrade, txn.Type)
	assert.Equal(t, "VALID", txn.Status)
	assert.Equal(t, "CASH", txn.SubAccount)
	assert.Equal(t, "2024-01-15", txn.TradeDate)
	assert.Equal(t, int64(5001), txn.PositionID)
	assert.Equal(t, int64(9001), txn.OrderID)
	assert.InDelta(t, -1500.00, txn.NetAmount, 0.000001)
	assert.Equal(t, "Buy 10 AAPL", txn.Description)
	assert.Equal(t, "123456789", txn.AccountNumber)

	// Verify TransferItem and Instrument unmarshal
	require.Len(t, txn.TransferItems, 1)
	item := txn.TransferItems[0]
	assert.InDelta(t, 10.0, item.Amount, 0.000001)
	assert.InDelta(t, 1500.00, item.Cost, 0.000001)
	assert.InDelta(t, 150.00, item.Price, 0.000001)
	assert.Equal(t, schwab.AssetTypeEquity, item.Instrument.AssetType)
	assert.Equal(t, "037833100", item.Instrument.Cusip)
	assert.Equal(t, "AAPL", item.Instrument.Symbol)
	assert.Equal(t, "Apple Inc", item.Instrument.Description)
	assert.Equal(t, int64(1234567), item.Instrument.InstrumentID)
}

func TestGetTransactions_WithOptionalParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		// Verify required params
		assert.Equal(t, "2024-01-01", r.URL.Query().Get("startDate"))
		assert.Equal(t, "2024-01-31", r.URL.Query().Get("endDate"))

		// Verify optional params
		assert.Equal(t, "TRADE,DIVIDEND_OR_INTEREST", r.URL.Query().Get("types"))
		assert.Equal(t, "AAPL", r.URL.Query().Get("symbol"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, []Transaction{})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetTransactions(context.Background(), "HASH_ABC123", &TransactionListParams{
		StartDate: "2024-01-01",
		EndDate:   "2024-01-31",
		Types:     "TRADE,DIVIDEND_OR_INTEREST",
		Symbol:    "AAPL",
	})
	require.NoError(t, err)
}

func TestGetTransaction(t *testing.T) {
	fixture := Transaction{
		ActivityID:    2002,
		Time:          "2024-02-10T14:00:00Z",
		Type:          TransactionTypeDividendOrInterest,
		Status:        "VALID",
		SubAccount:    "CASH",
		TradeDate:     "2024-02-10",
		PositionID:    6001,
		OrderID:       0,
		NetAmount:     25.50,
		Description:   "DIVIDEND PAYMENT",
		AccountNumber: "123456789",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/accounts/HASH_ABC123/transactions/2002", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, []Transaction{fixture})
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithToken("test-token"),
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	result, err := client.GetTransaction(context.Background(), "HASH_ABC123", 2002)
	require.NoError(t, err)
	require.Len(t, result, 1)
	txn := result[0]

	assert.Equal(t, int64(2002), txn.ActivityID)
	assert.Equal(t, "2024-02-10T14:00:00Z", txn.Time)
	assert.Equal(t, TransactionTypeDividendOrInterest, txn.Type)
	assert.Equal(t, "VALID", txn.Status)
	assert.InDelta(t, 25.50, txn.NetAmount, 0.000001)
	assert.Equal(t, "DIVIDEND PAYMENT", txn.Description)
}

func TestGetTransactionsRequiresParams(t *testing.T) {
	client := NewClient()

	_, err := client.GetTransactions(context.Background(), "HASH_ABC123", nil)
	require.EqualError(t, err, "transaction list params are required")

	_, err = client.GetTransactions(
		context.Background(),
		"HASH_ABC123",
		&TransactionListParams{EndDate: "2024-01-31", Types: "TRADE"},
	)
	require.EqualError(t, err, "startDate is required")

	_, err = client.GetTransactions(
		context.Background(),
		"HASH_ABC123",
		&TransactionListParams{StartDate: "2024-01-01", Types: "TRADE"},
	)
	require.EqualError(t, err, "endDate is required")

	_, err = client.GetTransactions(
		context.Background(),
		"HASH_ABC123",
		&TransactionListParams{StartDate: "2024-01-01", EndDate: "2024-01-31"},
	)
	require.EqualError(t, err, "types is required")
}

func TestGetTransactions_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	client := NewClient(
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)

	_, err := client.GetTransactions(context.Background(), "HASH_ABC123", &TransactionListParams{
		StartDate: "2024-01-01",
		EndDate:   "2024-01-31",
		Types:     "TRADE",
	})
	require.Error(t, err)

	apiErr, ok := errors.AsType[*schwab.APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
}
