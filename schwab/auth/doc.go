// Package auth implements OAuth2 authorization code flow for the Schwab API. It
// provides token acquisition, refresh, status inspection, and persistence as a
// library suitable for integration into CLI tools, web servers, or any Go
// application.
//
// # Quick Start
//
// Create a config and run the login flow. The callback URL must exactly match
// the URL configured in the Schwab developer portal, including the path and
// explicit port.
//
//	cfg := auth.Config{
//	    ClientID:     "your-app-key",
//	    ClientSecret: "your-app-secret",
//	    CallbackURL:  "https://127.0.0.1:8443/callback",
//	}
//
//	store := auth.NewFileTokenStore("/path/to/tokens.json")
//	// For tests or short-lived applications, use auth.NewMemoryTokenStore().
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//
//	openBrowser := func(url string) error { return exec.Command("xdg-open", url).Start() }
//	provider, err := auth.Login(ctx, cfg, store, openBrowser)
//
// Use the returned TokenProvider with schwab-go clients:
//
//	client := marketdata.NewClient(schwab.WithTokenProvider(provider))
//
// Provider.Token returns an existing access token until it is near expiry, then
// refreshes and saves a replacement through the configured TokenStore. Use
// Provider.Refresh when an application needs an explicit refresh command, and
// Provider.Status or InspectToken when it needs read-only token lifecycle state
// without refreshing or saving tokens.
//
// For CLI applications, keep command policy outside this package: explicit
// config and token paths, environment-variable precedence, JSON output
// envelopes, exit-code mappings, browser selection, and post-login default
// account setup belong in the application adapter layer. See
// docs/auth-cli-adapter.md in the repository for a Cobra-oriented pattern.
package auth
