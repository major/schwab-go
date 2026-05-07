package trader_test

import (
	"context"
	"net/http"
	"strconv"

	"github.com/major/schwab-go/schwab/trader"
)

func ExampleClient_GetUserPreference() {
	var method string
	var path string
	client, closeServer := newOrderExampleClient(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.String()
		writeExampleJSON(w, []trader.UserPreference{
			{
				Accounts: []trader.PreferenceAccount{
					{
						AccountNumber:      "123456789",
						PrimaryAccount:     true,
						Type:               "BROKERAGE",
						NickName:           "Main brokerage",
						AccountColor:       "Blue",
						DisplayAcctID:      "...6789",
						AutoPositionEffect: false,
					},
				},
				StreamerInfo: []trader.StreamerInfo{
					{
						StreamerSocketURL:      "wss://streamer.example.test/ws",
						SchwabClientCustomerID: "customer-123",
						SchwabClientCorrelID:   "correl-456",
						SchwabClientChannel:    "IO",
						SchwabClientFunctionID: "APIAPP",
					},
				},
				Offers: []trader.Offer{
					{Level2Permissions: true, MktDataPermission: "NP"},
				},
			},
		})
	})
	defer closeServer()

	preferences, err := client.GetUserPreference(context.Background())
	if err != nil {
		panic(err)
	}

	preference := preferences[0]
	account := preference.Accounts[0]
	streamer := preference.StreamerInfo[0]
	offer := preference.Offers[0]
	printExampleLines(
		"request="+method+" "+path,
		"preferences="+strconv.Itoa(len(preferences)),
		"primary="+account.AccountNumber+" "+account.DisplayAcctID,
		"streamerChannel="+streamer.SchwabClientChannel,
		"level2="+strconv.FormatBool(offer.Level2Permissions),
	)

	// Output:
	// request=GET /userPreference
	// preferences=1
	// primary=123456789 ...6789
	// streamerChannel=IO
	// level2=true
}

func ExampleUserPreference() {
	preference := trader.UserPreference{
		Accounts: []trader.PreferenceAccount{
			{AccountNumber: "123456789", PrimaryAccount: true},
		},
		StreamerInfo: []trader.StreamerInfo{
			{SchwabClientChannel: "IO", SchwabClientFunctionID: "APIAPP"},
		},
		Offers: []trader.Offer{
			{Level2Permissions: true, MktDataPermission: "NP"},
		},
	}

	printExampleLines(
		"accounts="+strconv.Itoa(len(preference.Accounts)),
		"streamers="+strconv.Itoa(len(preference.StreamerInfo)),
		"offers="+strconv.Itoa(len(preference.Offers)),
	)

	// Output:
	// accounts=1
	// streamers=1
	// offers=1
}

func ExamplePreferenceAccount() {
	account := trader.PreferenceAccount{
		AccountNumber:      "123456789",
		PrimaryAccount:     true,
		Type:               "BROKERAGE",
		NickName:           "Main brokerage",
		AccountColor:       "Blue",
		DisplayAcctID:      "...6789",
		AutoPositionEffect: false,
	}

	printExampleLines(
		"account="+account.AccountNumber,
		"primary="+strconv.FormatBool(account.PrimaryAccount),
		"type="+account.Type,
		"nickName="+account.NickName,
		"color="+account.AccountColor,
		"display="+account.DisplayAcctID,
		"autoPositionEffect="+strconv.FormatBool(account.AutoPositionEffect),
	)

	// Output:
	// account=123456789
	// primary=true
	// type=BROKERAGE
	// nickName=Main brokerage
	// color=Blue
	// display=...6789
	// autoPositionEffect=false
}

func ExampleStreamerInfo() {
	streamer := trader.StreamerInfo{
		StreamerSocketURL:      "wss://streamer.example.test/ws",
		SchwabClientCustomerID: "customer-123",
		SchwabClientCorrelID:   "correl-456",
		SchwabClientChannel:    "IO",
		SchwabClientFunctionID: "APIAPP",
	}

	printExampleLines(
		"socket="+streamer.StreamerSocketURL,
		"customer="+streamer.SchwabClientCustomerID,
		"correl="+streamer.SchwabClientCorrelID,
		"channel="+streamer.SchwabClientChannel,
		"function="+streamer.SchwabClientFunctionID,
	)

	// Output:
	// socket=wss://streamer.example.test/ws
	// customer=customer-123
	// correl=correl-456
	// channel=IO
	// function=APIAPP
}

func ExampleOffer() {
	offer := trader.Offer{
		Level2Permissions: true,
		MktDataPermission: "NP",
	}

	printExampleLines(
		"level2="+strconv.FormatBool(offer.Level2Permissions),
		"marketData="+offer.MktDataPermission,
	)

	// Output:
	// level2=true
	// marketData=NP
}
