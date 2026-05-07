// Package auth implements OAuth2 authorization code flow for the
// Schwab API. It provides token acquisition, refresh, and persistence
// as a library suitable for integration into CLI tools, web servers,
// or any Go application.
//
// The package is a separate Go module (github.com/major/schwab-go/schwab/auth)
// and must be imported independently from the root schwab-go module.
//
// # Quick Start
//
// Create a config and run the login flow:
//
//	cfg := auth.Config{
//	    ClientID:     "your-app-key",
//	    ClientSecret: "your-app-secret",
//	    CallbackURL:  "https://127.0.0.1:8443/callback",
//	}
//
//	store := auth.NewFileTokenStore("/path/to/tokens.json")
//	openBrowser := func(url string) error { return exec.Command("xdg-open", url).Start() }
//	provider, err := auth.Login(ctx, cfg, store, openBrowser)
//
// Use the returned TokenProvider with schwab-go clients:
//
//	client := marketdata.NewClient(schwab.WithTokenProvider(provider))
package auth
