# schwab-go

[![CI](https://github.com/major/schwab-go/actions/workflows/ci.yml/badge.svg)](https://github.com/major/schwab-go/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/major/schwab-go/branch/main/graph/badge.svg)](https://codecov.io/gh/major/schwab-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/major/schwab-go)](https://goreportcard.com/report/github.com/major/schwab-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/major/schwab-go.svg)](https://pkg.go.dev/github.com/major/schwab-go)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/major/schwab-go/badge)](https://scorecard.dev/viewer/?uri=github.com/major/schwab-go)
[![License](https://img.shields.io/github/license/major/schwab-go)](LICENSE)

Go client library for the [Schwab API](https://developer.schwab.com/). Covers Market Data and Trader API endpoints with typed responses, functional options, and structured error handling.

## Features

- **Market Data** - quotes, price history, option chains, instruments, market hours, movers
- **Trader** - accounts, orders (create/replace/cancel/preview), transactions, user preferences
- **Auth** (`schwab/auth`) - OAuth2 authorization code flow, token refresh, and file-based token persistence included in the core module
- **Typed quote accessors** - asset-specific quote and reference types for equities, options, indices, mutual funds, forex, futures, and future options
- **Structured errors** - `*schwab.APIError` with status code, message, and up to 1 MiB of the raw body
- **Functional options** - `WithToken`, `WithHTTPClient`, `WithTLSConfig`, `WithBaseURL`, `WithResponseBodyLimit`, `WithUserAgent`, `WithHeader`, and `WithHeaders` for flexible client configuration. Invalid base URL overrides fail when a request is created instead of falling back to the production Schwab API. Sub-clients append their own API path prefixes, so custom base URLs can point at the API root. Response bodies are capped at 10 MiB by default; non-positive custom limits are ignored.
- **Context propagation** - all request methods take `context.Context`
- **Testable** - override HTTP client and base URL for `httptest` integration
- **No runtime dependencies** - public client packages stay dependency-free; tests use `stretchr/testify` and `kin-openapi`

## Installation

```bash
go get github.com/major/schwab-go
```

Requires Go 1.26 or later.

## Quick start

### Market Data

```go
package main

import (
	"context"
	"fmt"
	"log"

	schwab "github.com/major/schwab-go/schwab"
	"github.com/major/schwab-go/schwab/marketdata"
)

func main() {
	client := marketdata.NewClient(
		schwab.WithToken("your-bearer-token"),
	)

	// Fetch quotes for multiple symbols.
	quotes, quoteErr, err := client.GetQuotes(
		context.Background(),
		[]string{"AAPL", "MSFT"},
		"quote,reference", // fields
		false,             // indicative
	)
	if err != nil {
		log.Fatal(err)
	}
	if quoteErr != nil {
		fmt.Printf("invalid symbols: %v\n", quoteErr.InvalidSymbols)
	}

	for symbol, entry := range *quotes {
		eq, err := entry.EquityQuote()
		if err != nil {
			continue
		}
		fmt.Printf("%s: $%.2f\n", symbol, eq.LastPrice)
	}
}
```

### Trader

```go
package main

import (
	"context"
	"fmt"
	"log"

	schwab "github.com/major/schwab-go/schwab"
	"github.com/major/schwab-go/schwab/trader"
)

func main() {
	client := trader.NewClient(
		schwab.WithToken("your-bearer-token"),
	)

	// Get encrypted account hashes (required for all account-scoped calls).
	accounts, err := client.GetAccountNumbers(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Fetch account details with positions.
	for _, acct := range accounts {
		detail, err := client.GetAccount(
			context.Background(),
			acct.HashValue,
			"positions",
		)
		if err != nil {
			log.Fatal(err)
		}
		sa := detail.SecuritiesAccount
		fmt.Printf("Account %s: %d positions\n", sa.AccountNumber, len(sa.Positions))
	}
}
```

### Error handling

All API errors are returned as `*schwab.APIError`, which includes the HTTP status code, a message, and up to 1 MiB of the raw response body:

```go
import "errors"

_, _, err := client.GetQuotes(ctx, []string{"AAPL"}, "", false)
if apiErr, ok := errors.AsType[*schwab.APIError](err); ok {
	fmt.Printf("HTTP %d: %s\n", apiErr.StatusCode, apiErr.Message)
}
```

Use `schwab.StatusCode(err)` or `schwab.IsUnauthorized(err)` when callers only need generic HTTP status classification and should not depend on endpoint-specific error bodies.

## Authentication

The `schwab/auth` package handles OAuth2 authorization code flow, token refresh, read-only token status inspection, and persistence as part of the core `github.com/major/schwab-go` module.

```go
import (
    "context"
    "log"
    "os/exec"

    "github.com/major/schwab-go/schwab/auth"
    schwab "github.com/major/schwab-go/schwab"
    "github.com/major/schwab-go/schwab/marketdata"
)

cfg := auth.Config{
    ClientID:     "your-app-key",
    ClientSecret: "your-app-secret",
    CallbackURL:  "https://127.0.0.1:8443/callback",
}

store := auth.NewFileTokenStore("/path/to/tokens.json")
// For tests or short-lived applications, use auth.NewMemoryTokenStore().
ctx := context.Background()

// urlHandler receives the authorize URL. Open it in a browser or print it for SSH/headless use.
openBrowser := func(url string) error { return exec.Command("xdg-open", url).Start() }
provider, err := auth.Login(ctx, cfg, store, openBrowser)
if err != nil {
    log.Fatal(err)
}

client := marketdata.NewClient(schwab.WithTokenProvider(provider))
```

For headless or SSH environments, pass a `urlHandler` that prints the URL instead of opening a browser. Use `auth.StartLogin` when an application needs to display the authorization URL before blocking for the callback and token exchange.

`auth.ConfigFromAPIBaseURL` derives OAuth endpoints from a caller-owned Schwab API root or proxy prefix, and `auth.OAuthBaseURLFromAPIBaseURL` exposes the URL conversion separately for adapters with their own config structs.

`auth.NewFileProvider` is a convenience wrapper around `auth.NewProvider(cfg, auth.NewFileTokenStore(path), httpClient)`. `Provider.Token` refreshes expired access tokens automatically and writes refreshed tokens back to the store. Use `Provider.Refresh` or `auth.RefreshTokenFile` for explicit refresh commands, and `Provider.Status` or `auth.InspectToken` for read-only status output that must not refresh or save tokens.

`auth.RedactToken` and `auth.RedactClientID` provide safe display strings for troubleshooting output. `auth.NewMemoryTokenStore` is available for tests, examples, and short-lived applications that do not need token durability across process restarts. Use `auth.IsRequired`, `auth.IsExpired`, and `auth.IsCallback` to classify auth failures at application boundaries without duplicating `errors.As` checks.

CLI applications that need `auth login`, `auth status`, `auth refresh`, a global auth gate, JSON output envelopes, or post-login default-account setup should keep that command policy in the application adapter layer. See [Auth CLI adapter pattern](docs/auth-cli-adapter.md) for Cobra-oriented guidance.

If you already have a valid bearer token, skip `schwab/auth` and pass it directly with `schwab.WithToken()`.

If you previously installed `schwab/auth` as its own module, remove the separate `require github.com/major/schwab-go/schwab/auth ...` entry and run `go get github.com/major/schwab-go@latest && go mod tidy` so auth resolves from the core module.

## API coverage

### Market Data (`schwab/marketdata`)

| Method | Description |
|--------|-------------|
| `GetQuotes` | Multi-symbol quotes with optional fields |
| `GetQuote` | Single symbol quote |
| `GetPriceHistory` | OHLCV candles with configurable period/frequency |
| `SearchInstruments` | Search instruments by symbol or name |
| `GetInstrumentByCUSIP` | Look up instrument by CUSIP |
| `GetOptionChain` | Full option chain with strikes and expirations |
| `GetExpirationChain` | Expiration dates for a symbol |
| `GetMovers` | Market movers by index |
| `GetMarketHours` | Market hours for string market IDs with date validation |
| `GetMarketHoursTyped` | Market hours for typed market IDs with date validation |
| `GetMarketHoursSingle` | Market hours for a single string market ID with date validation |
| `GetMarketHoursSingleTyped` | Market hours for a single typed market ID with date validation |

Typed market hours methods use `MarketID` constants such as `marketdata.MarketIDEquity`. All market hours methods validate supported market values and optional dates as `YYYY-MM-DD` values from today through one year out.

### Trader (`schwab/trader`)

| Method | Description |
|--------|-------------|
| `GetAccountNumbers` | Encrypted account hashes (call first) |
| `GetAccounts` / `GetAccount` | Account details with optional positions |
| `GetAccountsRaw` / `GetAccountRaw` | Raw account JSON for compatibility adapters that need exact API fields |
| `GetOrders` / `GetAllOrders` / `GetOrder` | Retrieve orders |
| `CreateOrder` / `ReplaceOrder` / `CancelOrder` | Order lifecycle |
| `CreateOrderWithResponse` / `ReplaceOrderWithResponse` | Create/replace orders and return parsed order ID from Location header |
| `PreviewOrder` / `PreviewOrderRequestBody` | Preview an order before submission |
| `GetTransactions` / `GetTransaction` / `GetTransactionByID` | Transaction history |
| `GetUserPreference` | User preferences |

## Documentation

Full API documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/major/schwab-go).

## Contributing

Contributions are welcome. Please open an issue to discuss larger changes before submitting a pull request.

## License

[Apache License 2.0](LICENSE)
