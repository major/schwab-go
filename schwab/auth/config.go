package auth

import (
	"fmt"
	"net/url"
)

// Config holds the OAuth2 client credentials and endpoint configuration
// for the Schwab API. Construct it directly or load it from a JSON file
// with LoadConfig.
type Config struct {
	// ClientID is the OAuth2 application client identifier.
	ClientID string `json:"client_id"`

	// ClientSecret is the OAuth2 application client secret.
	ClientSecret string `json:"client_secret"`

	// CallbackURL is the loopback OAuth2 redirect URL. Schwab requires this
	// URL to use the 127.0.0.1 host.
	CallbackURL string `json:"callback_url"`

	// OAuthBaseURL overrides the OAuth endpoint base URL. When empty, callers
	// should use https://api.schwabapi.com/v1/oauth.
	OAuthBaseURL string `json:"oauth_base_url,omitempty"`
}

// Validate returns an error if any required Config field is missing or invalid.
func (c Config) Validate() error {
	if c.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}

	if c.ClientSecret == "" {
		return fmt.Errorf("client_secret is required")
	}

	if err := validateCallbackURL(c.CallbackURL); err != nil {
		return err
	}

	if c.OAuthBaseURL != "" {
		if _, err := parseAbsoluteURL("oauth_base_url", c.OAuthBaseURL); err != nil {
			return err
		}
	}

	return nil
}

func validateCallbackURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("callback_url is required")
	}

	callbackURL, err := parseAbsoluteURL("callback_url", rawURL)
	if err != nil {
		return err
	}

	if callbackURL.Hostname() != "127.0.0.1" {
		return fmt.Errorf("callback_url host must be 127.0.0.1")
	}

	return nil
}

func parseAbsoluteURL(fieldName, rawURL string) (*url.URL, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("%s must be a valid URL: %w", fieldName, err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("%s must include scheme and host", fieldName)
	}

	return parsedURL, nil
}
