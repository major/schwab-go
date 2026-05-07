package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
)

const (
	// DefaultOAuthBaseURL is the default Schwab OAuth2 base URL.
	DefaultOAuthBaseURL = "https://api.schwabapi.com/v1/oauth"

	// httpsScheme is the required URL scheme for callback and OAuth URLs.
	httpsScheme = "https"
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
	// should use DefaultOAuthBaseURL.
	OAuthBaseURL string `json:"oauth_base_url,omitempty"`
}

// Validate returns an error if any required Config field is missing or invalid.
func (c Config) Validate() error {
	if c.ClientID == "" {
		return errors.New("client_id is required")
	}

	if c.ClientSecret == "" {
		return errors.New("client_secret is required")
	}

	err := validateCallbackURL(c.CallbackURL)
	if err != nil {
		return err
	}

	if c.OAuthBaseURL != "" {
		err = validateOAuthBaseURL(c.OAuthBaseURL)
		if err != nil {
			return err
		}
	}

	return nil
}

// LoadConfig reads a JSON config file from path, validates it, and returns the
// loaded Config.
func LoadConfig(path string) (Config, error) {
	rawConfig, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg Config
	err = json.Unmarshal(rawConfig, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("parse config %q: %w", path, err)
	}

	err = cfg.Validate()
	if err != nil {
		return Config{}, fmt.Errorf("validate config %q: %w", path, err)
	}

	return cfg, nil
}

func validateCallbackURL(rawURL string) error {
	if rawURL == "" {
		return errors.New("callback_url is required")
	}

	callbackURL, err := parseAbsoluteURL("callback_url", rawURL)
	if err != nil {
		return err
	}

	if callbackURL.Scheme != httpsScheme {
		return errors.New("callback_url scheme must be https")
	}

	if callbackURL.Hostname() != "127.0.0.1" {
		return errors.New("callback_url host must be 127.0.0.1")
	}

	if callbackURL.Port() == "" {
		return errors.New("callback_url must include an explicit port")
	}

	return nil
}

func validateOAuthBaseURL(rawURL string) error {
	oauthBaseURL, err := parseAbsoluteURL("oauth_base_url", rawURL)
	if err != nil {
		return err
	}

	if oauthBaseURL.Scheme == httpsScheme {
		return nil
	}

	return errors.New("oauth_base_url scheme must be https")
}

// oauthBaseURL returns OAuthBaseURL if set, or DefaultOAuthBaseURL.
func (c Config) oauthBaseURL() string {
	if c.OAuthBaseURL != "" {
		return c.OAuthBaseURL
	}

	return DefaultOAuthBaseURL
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
