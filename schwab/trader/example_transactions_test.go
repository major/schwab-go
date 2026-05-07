package trader_test

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	schwab "github.com/major/schwab-go/schwab"
	"github.com/major/schwab-go/schwab/trader"
)

func ExampleClient_GetTransactions() {
	var method string
	var path string
	client, closeServer := newOrderExampleClient(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.String()
		writeExampleJSON(w, []trader.Transaction{transactionExample()})
	})
	defer closeServer()

	transactions, err := client.GetTransactions(context.Background(), "HASH_ABC123", &trader.TransactionListParams{
		StartDate: "2024-01-01",
		EndDate:   "2024-01-31",
		Types:     string(trader.TransactionTypeTrade),
		Symbol:    "AAPL",
	})
	if err != nil {
		panic(err)
	}

	transaction := transactions[0]
	item := transaction.TransferItems[0]
	printExampleLines(
		"request="+method+" "+path,
		"transactions="+strconv.Itoa(len(transactions)),
		"first="+strconv.FormatInt(transaction.ActivityID, 10)+" "+string(transaction.Type),
		"symbol="+item.Instrument.Symbol,
		"amount="+formatExampleFloat(item.Amount, 0),
	)

	// Output:
	// request=GET /accounts/HASH_ABC123/transactions?endDate=2024-01-31&startDate=2024-01-01&symbol=AAPL&types=TRADE
	// transactions=1
	// first=1001 TRADE
	// symbol=AAPL
	// amount=10
}

func ExampleClient_GetTransaction() {
	var method string
	var path string
	client, closeServer := newOrderExampleClient(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.String()
		writeExampleJSON(w, []trader.Transaction{transactionExample()})
	})
	defer closeServer()

	transactions, err := client.GetTransaction(context.Background(), "HASH_ABC123", 1001)
	if err != nil {
		panic(err)
	}

	transaction := transactions[0]
	printExampleLines(
		"request="+method+" "+path,
		"transaction="+strconv.FormatInt(transaction.ActivityID, 10),
		"description="+transaction.Description,
		"netAmount="+formatExampleFloat(transaction.NetAmount, 2),
	)

	// Output:
	// request=GET /accounts/HASH_ABC123/transactions/1001
	// transaction=1001
	// description=Buy 10 AAPL
	// netAmount=-1500.00
}

func ExampleTransactionListParams() {
	params := trader.TransactionListParams{
		StartDate: "2024-01-01",
		EndDate:   "2024-01-31",
		Types: strings.Join([]string{
			string(trader.TransactionTypeTrade),
			string(trader.TransactionTypeDividendOrInterest),
		}, ","),
		Symbol: "AAPL",
	}

	printExampleLines(
		"start="+params.StartDate,
		"end="+params.EndDate,
		"types="+params.Types,
		"symbol="+params.Symbol,
	)

	// Output:
	// start=2024-01-01
	// end=2024-01-31
	// types=TRADE,DIVIDEND_OR_INTEREST
	// symbol=AAPL
}

func ExampleTransaction() {
	transaction := transactionExample()
	item := transaction.TransferItems[0]

	printExampleLines(
		"activity="+strconv.FormatInt(transaction.ActivityID, 10),
		"type="+string(transaction.Type),
		"status="+transaction.Status,
		"account="+transaction.AccountNumber,
		"transfer="+item.Instrument.Symbol+" "+formatExampleFloat(item.Amount, 0),
	)

	// Output:
	// activity=1001
	// type=TRADE
	// status=VALID
	// account=123456789
	// transfer=AAPL 10
}

func ExampleTransactionInstrument() {
	instrument := trader.TransactionInstrument{
		AssetType:               schwab.AssetTypeOption,
		Cusip:                   "OPTIONCUSIP",
		Symbol:                  "AAPL  240621C00200000",
		Description:             "AAPL Jun 21 2024 200 Call",
		InstrumentID:            7654321,
		ExpirationDate:          "2024-06-21",
		OptionPremiumMultiplier: 100,
		PutCall:                 "CALL",
		StrikePrice:             200,
		Type:                    "VANILLA",
		UnderlyingSymbol:        "AAPL",
		UnderlyingCusip:         "037833100",
		OptionDeliverables: []trader.TransactionOptionDeliverable{
			{
				RootSymbol:        "AAPL",
				DeliverableNumber: 1,
				DeliverableUnits:  100,
				AssetType:         schwab.AssetTypeEquity,
			},
		},
	}

	printExampleLines(
		"symbol="+instrument.Symbol,
		"assetType="+string(instrument.AssetType),
		"description="+instrument.Description,
		"type="+instrument.Type,
		"expiration="+instrument.ExpirationDate,
		"multiplier="+strconv.FormatInt(instrument.OptionPremiumMultiplier, 10),
		"putCall="+instrument.PutCall,
		"strike="+formatExampleFloat(instrument.StrikePrice, 0),
		"underlying="+instrument.UnderlyingSymbol,
		"underlyingCusip="+instrument.UnderlyingCusip,
		"cusip="+instrument.Cusip,
		"instrumentID="+strconv.FormatInt(instrument.InstrumentID, 10),
		"deliverable="+formatExampleFloat(instrument.OptionDeliverables[0].DeliverableUnits, 0),
	)

	// Output:
	// symbol=AAPL  240621C00200000
	// assetType=OPTION
	// description=AAPL Jun 21 2024 200 Call
	// type=VANILLA
	// expiration=2024-06-21
	// multiplier=100
	// putCall=CALL
	// strike=200
	// underlying=AAPL
	// underlyingCusip=037833100
	// cusip=OPTIONCUSIP
	// instrumentID=7654321
	// deliverable=100
}

func ExampleTransactionOptionDeliverable() {
	deliverable := trader.TransactionOptionDeliverable{
		RootSymbol:        "AAPL",
		StrikePercent:     100,
		DeliverableNumber: 1,
		DeliverableUnits:  100,
		Deliverable: &trader.TransactionInstrument{
			AssetType:   schwab.AssetTypeEquity,
			Symbol:      "AAPL",
			Description: "Apple Inc",
		},
		AssetType: schwab.AssetTypeOption,
	}

	printExampleLines(
		"root="+deliverable.RootSymbol,
		"strikePercent="+strconv.FormatInt(deliverable.StrikePercent, 10),
		"number="+strconv.FormatInt(deliverable.DeliverableNumber, 10),
		"units="+formatExampleFloat(deliverable.DeliverableUnits, 0),
		"assetType="+string(deliverable.AssetType),
		"deliverable="+deliverable.Deliverable.Symbol,
	)

	// Output:
	// root=AAPL
	// strikePercent=100
	// number=1
	// units=100
	// assetType=OPTION
	// deliverable=AAPL
}

func ExampleTransactionType() {
	printCatalog("transaction types", []string{
		string(trader.TransactionTypeTrade),
		string(trader.TransactionTypeDividendOrInterest),
		string(trader.TransactionTypeACHReceipt),
		string(trader.TransactionTypeCashDisbursement),
		string(trader.TransactionTypeWireOut),
		string(trader.TransactionTypeMoneyMarket),
	})

	// Output:
	// transaction types: TRADE, DIVIDEND_OR_INTEREST, ACH_RECEIPT, CASH_DISBURSEMENT, WIRE_OUT, MONEY_MARKET
}

func transactionExample() trader.Transaction {
	return trader.Transaction{
		ActivityID:     1001,
		Time:           "2024-01-15T10:30:00Z",
		Type:           trader.TransactionTypeTrade,
		Status:         "VALID",
		SubAccount:     "CASH",
		TradeDate:      "2024-01-15",
		SettlementDate: "2024-01-17",
		PositionID:     5001,
		OrderID:        9001,
		NetAmount:      -1500,
		ActivityType:   "ACTIVITY_CORRECTION",
		Description:    "Buy 10 AAPL",
		AccountNumber:  "123456789",
		TransferItems: []trader.TransferItem{
			{
				Instrument: trader.TransactionInstrument{
					AssetType:    schwab.AssetTypeEquity,
					Cusip:        "037833100",
					Symbol:       "AAPL",
					Description:  "Apple Inc",
					InstrumentID: 1234567,
				},
				Amount:         10,
				Cost:           1500,
				Price:          150,
				PositionEffect: "OPENING",
			},
		},
	}
}
