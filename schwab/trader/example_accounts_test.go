package trader_test

import (
	"context"
	"net/http"
	"strconv"

	schwab "github.com/major/schwab-go/schwab"
	"github.com/major/schwab-go/schwab/trader"
)

func ExampleClient_GetAccounts() {
	var method string
	var path string
	client, closeServer := newOrderExampleClient(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.String()
		writeExampleJSON(w, []trader.Account{
			accountExample(),
		})
	})
	defer closeServer()

	accounts, err := client.GetAccounts(context.Background(), "positions")
	if err != nil {
		panic(err)
	}

	account := accounts[0].SecuritiesAccount
	printExampleLines(
		"request="+method+" "+path,
		"accounts="+strconv.Itoa(len(accounts)),
		"account="+account.AccountNumber+" "+account.Type,
		"positions="+strconv.Itoa(len(account.Positions)),
		"liquidationValue="+formatExampleFloat(account.CurrentBalances.LiquidationValue, 0),
	)

	// Output:
	// request=GET /accounts?fields=positions
	// accounts=1
	// account=123456789 CASH
	// positions=1
	// liquidationValue=6530
}

func ExampleClient_GetAccount() {
	var method string
	var path string
	client, closeServer := newOrderExampleClient(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.String()
		writeExampleJSON(w, accountExample())
	})
	defer closeServer()

	account, err := client.GetAccount(context.Background(), "HASH_ABC123", "positions")
	if err != nil {
		panic(err)
	}

	securitiesAccount := account.SecuritiesAccount
	position := securitiesAccount.Positions[0]
	printExampleLines(
		"request="+method+" "+path,
		"account="+securitiesAccount.AccountNumber,
		"position="+position.Instrument.Symbol+" "+formatExampleFloat(position.LongQuantity, 0),
		"cash="+formatExampleFloat(securitiesAccount.CurrentBalances.CashBalance, 0),
	)

	// Output:
	// request=GET /accounts/HASH_ABC123?fields=positions
	// account=123456789
	// position=AAPL 10
	// cash=5000
}

func ExampleAccount() {
	account := accountExample()
	securitiesAccount := account.SecuritiesAccount

	printExampleLines(
		"account="+securitiesAccount.AccountNumber,
		"type="+securitiesAccount.Type,
		"positions="+strconv.Itoa(len(securitiesAccount.Positions)),
	)

	// Output:
	// account=123456789
	// type=CASH
	// positions=1
}

func ExampleSecuritiesAccount() {
	account := accountExample().SecuritiesAccount

	printExampleLines(
		"type="+account.Type,
		"account="+account.AccountNumber,
		"roundTrips="+strconv.Itoa(account.RoundTrips),
		"dayTrader="+strconv.FormatBool(account.IsDayTrader),
		"closingOnly="+strconv.FormatBool(account.IsClosingOnlyRestricted),
		"pfcb="+strconv.FormatBool(account.PfcbFlag),
		"positions="+strconv.Itoa(len(account.Positions)),
		"currentValue="+formatExampleFloat(account.CurrentBalances.AccountValue, 0),
	)

	// Output:
	// type=CASH
	// account=123456789
	// roundTrips=0
	// dayTrader=false
	// closingOnly=false
	// pfcb=false
	// positions=1
	// currentValue=6530
}

func ExamplePosition() {
	position := accountExample().SecuritiesAccount.Positions[0]

	printExampleLines(
		"symbol="+position.Instrument.Symbol,
		"long="+formatExampleFloat(position.LongQuantity, 0),
		"short="+formatExampleFloat(position.ShortQuantity, 0),
		"average="+formatExampleFloat(position.AveragePrice, 0),
		"marketValue="+formatExampleFloat(position.MarketValue, 0),
		"dayPL="+formatExampleFloat(position.CurrentDayProfitLoss, 0),
		"dayPLPercent="+formatExampleFloat(position.CurrentDayProfitLossPercentage, 0),
	)

	// Output:
	// symbol=AAPL
	// long=10
	// short=0
	// average=150
	// marketValue=1530
	// dayPL=300
	// dayPLPercent=2
}

func ExampleAccountInstrument() {
	instrument := trader.AccountInstrument{
		AssetType:        schwab.AssetTypeOption,
		Cusip:            "OPTIONCUSIP",
		Symbol:           "AAPL  240621C00200000",
		Description:      "AAPL Jun 21 2024 200 Call",
		InstrumentID:     7654321,
		NetChange:        0.15,
		Type:             "VANILLA",
		PutCall:          "CALL",
		OptionMultiplier: 100,
		OptionDeliverables: []trader.AccountOptionDeliverable{
			{
				Symbol:           "AAPL",
				DeliverableUnits: 100,
				APICurrencyType:  "USD",
				AssetType:        schwab.AssetTypeEquity,
			},
		},
		UnderlyingSymbol: "AAPL",
	}

	printExampleLines(
		"symbol="+instrument.Symbol,
		"assetType="+string(instrument.AssetType),
		"cusip="+instrument.Cusip,
		"description="+instrument.Description,
		"instrumentID="+strconv.FormatInt(instrument.InstrumentID, 10),
		"netChange="+formatExampleFloat(instrument.NetChange, 2),
		"type="+instrument.Type,
		"putCall="+instrument.PutCall,
		"multiplier="+strconv.Itoa(int(instrument.OptionMultiplier)),
		"deliverable="+instrument.OptionDeliverables[0].Symbol,
		"underlying="+instrument.UnderlyingSymbol,
	)

	// Output:
	// symbol=AAPL  240621C00200000
	// assetType=OPTION
	// cusip=OPTIONCUSIP
	// description=AAPL Jun 21 2024 200 Call
	// instrumentID=7654321
	// netChange=0.15
	// type=VANILLA
	// putCall=CALL
	// multiplier=100
	// deliverable=AAPL
	// underlying=AAPL
}

func ExampleAccountOptionDeliverable() {
	deliverable := trader.AccountOptionDeliverable{
		Symbol:           "AAPL",
		DeliverableUnits: 100,
		APICurrencyType:  "USD",
		AssetType:        schwab.AssetTypeEquity,
	}

	printExampleLines(
		"symbol="+deliverable.Symbol,
		"units="+formatExampleFloat(deliverable.DeliverableUnits, 0),
		"currency="+deliverable.APICurrencyType,
		"assetType="+string(deliverable.AssetType),
	)

	// Output:
	// symbol=AAPL
	// units=100
	// currency=USD
	// assetType=EQUITY
}

func ExampleBalance() {
	balance := trader.Balance{
		CashBalance:      5000,
		BuyingPower:      5000,
		Equity:           6530,
		LiquidationValue: 6530,
		LongStockValue:   1530,
		AccountValue:     6530,
	}

	printExampleLines(
		"cash="+formatExampleFloat(balance.CashBalance, 0),
		"buyingPower="+formatExampleFloat(balance.BuyingPower, 0),
		"equity="+formatExampleFloat(balance.Equity, 0),
		"liquidation="+formatExampleFloat(balance.LiquidationValue, 0),
		"longStock="+formatExampleFloat(balance.LongStockValue, 0),
		"accountValue="+formatExampleFloat(balance.AccountValue, 0),
	)

	// Output:
	// cash=5000
	// buyingPower=5000
	// equity=6530
	// liquidation=6530
	// longStock=1530
	// accountValue=6530
}

func accountExample() trader.Account {
	return trader.Account{
		SecuritiesAccount: trader.SecuritiesAccount{
			Type:                    "CASH",
			AccountNumber:           "123456789",
			RoundTrips:              0,
			IsDayTrader:             false,
			IsClosingOnlyRestricted: false,
			PfcbFlag:                false,
			Positions: []trader.Position{
				{
					AveragePrice:                   150,
					CurrentDayProfitLoss:           300,
					CurrentDayProfitLossPercentage: 2,
					LongQuantity:                   10,
					SettledLongQuantity:            10,
					Instrument: trader.AccountInstrument{
						AssetType:    schwab.AssetTypeEquity,
						Cusip:        "037833100",
						Symbol:       "AAPL",
						Description:  "Apple Inc",
						InstrumentID: 1234567,
						NetChange:    1.5,
					},
					MarketValue:                 1530,
					MaintenanceRequirement:      459,
					AverageLongPrice:            150,
					TaxLotAverageLongPrice:      150,
					LongOpenProfitLoss:          30,
					PreviousSessionLongQuantity: 10,
				},
			},
			CurrentBalances: trader.Balance{
				CashBalance:      5000,
				BuyingPower:      5000,
				Equity:           6530,
				LiquidationValue: 6530,
				LongStockValue:   1530,
				AccountValue:     6530,
			},
		},
	}
}
