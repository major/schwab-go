package trader_test

import (
	"context"
	"net/http"
	"strconv"

	"github.com/major/schwab-go/schwab/trader"
)

func ExampleClient_GetAccountNumbers() {
	var method string
	var path string
	client, closeServer := newOrderExampleClient(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.String()
		writeExampleJSON(w, []trader.AccountNumberHash{
			{AccountNumber: "123456789", HashValue: "HASH_ABC123"},
			{AccountNumber: "987654321", HashValue: "HASH_DEF456"},
		})
	})
	defer closeServer()

	accounts, err := client.GetAccountNumbers(context.Background())
	if err != nil {
		panic(err)
	}

	printExampleLines(
		"request="+method+" "+path,
		"accounts="+strconv.Itoa(len(accounts)),
		"first="+accounts[0].AccountNumber+" -> "+accounts[0].HashValue,
	)

	// Output:
	// request=GET /accounts/accountNumbers
	// accounts=2
	// first=123456789 -> HASH_ABC123
}

func ExampleAccountNumberHash() {
	account := trader.AccountNumberHash{
		AccountNumber: "123456789",
		HashValue:     "HASH_ABC123",
	}

	printExampleLines(
		"accountNumber="+account.AccountNumber,
		"hashValue="+account.HashValue,
		"useHashValueForTraderCalls=true",
	)

	// Output:
	// accountNumber=123456789
	// hashValue=HASH_ABC123
	// useHashValueForTraderCalls=true
}
