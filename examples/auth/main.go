// Command auth-example demonstrates the schwab/auth OAuth2 login flow.
//
// It loads credentials from a JSON config file, runs the browser-based
// authorization flow (or reuses a saved token), and makes a test API
// call to verify the token works.
//
// Usage:
//
//	# 1. Copy the sample config and fill in your Schwab app credentials:
//	cp config.sample.json config.json
//	vi config.json
//
//	# 2. Run the example:
//	go run .
//
//	# The browser opens the Schwab login page. After you authorize,
//	# the token is saved to tokens.json and a test quote is fetched.
//
//	# 3. Run again to verify token reuse (no browser prompt):
//	go run .
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	schwab "github.com/major/schwab-go/schwab"
	"github.com/major/schwab-go/schwab/auth"
	"github.com/major/schwab-go/schwab/marketdata"
)

const (
	configPath = "config.json"
	tokenPath  = "tokens.json"
	testSymbol = "AAPL"
)

func main() {
	ctx := context.Background()

	cfg, err := auth.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v\n\nCopy config.sample.json to config.json and fill in your Schwab app credentials.", configPath, err)
	}

	store := auth.NewFileTokenStore(tokenPath)
	provider, err := loadOrLogin(ctx, cfg, store)
	if err != nil {
		log.Fatalf("Auth failed: %v", err)
	}

	fmt.Println("Fetching quote for", testSymbol, "...")

	client := marketdata.NewClient(schwab.WithTokenProvider(provider))
	quote, err := client.GetQuote(ctx, testSymbol, "quote")
	if err != nil {
		log.Fatalf("API call failed: %v", err)
	}

	fmt.Printf("Got quote response for %s: %+v\n", testSymbol, quote)
	fmt.Println("Auth is working.")
}

// loadOrLogin tries to reuse a saved token. If no token exists or it
// has expired beyond refresh, it runs the full login flow.
func loadOrLogin(ctx context.Context, cfg auth.Config, store auth.TokenStore) (*auth.Provider, error) {
	provider, err := auth.NewProvider(cfg, store, nil)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	// Try the existing token first.
	_, tokenErr := provider.Token(ctx)
	if tokenErr == nil {
		fmt.Println("Reusing saved token from", tokenPath)
		return provider, nil
	}

	// If the error is anything other than "needs login", surface it.
	var requiredErr *auth.AuthRequiredError
	var expiredErr *auth.AuthExpiredError
	if !errors.As(tokenErr, &requiredErr) && !errors.As(tokenErr, &expiredErr) {
		return nil, tokenErr
	}

	fmt.Println("No valid token found, starting login flow...")

	provider, err = auth.Login(ctx, cfg, store, openBrowser)
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	fmt.Println("Login successful, token saved to", tokenPath)
	return provider, nil
}

// openBrowser opens the authorization URL in the user's default browser.
func openBrowser(url string) error {
	fmt.Println("Opening browser for Schwab authorization...")
	fmt.Println("If the browser does not open, visit this URL manually:")
	fmt.Println()
	fmt.Println(" ", url)
	fmt.Println()

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		// On unsupported platforms, the user can copy the URL above.
		fmt.Fprintln(os.Stderr, "Automatic browser open not supported on", runtime.GOOS)
		return nil
	}

	return cmd.Start()
}
