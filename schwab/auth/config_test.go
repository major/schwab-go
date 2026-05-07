package auth

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	validConfig := Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		CallbackURL:  "https://127.0.0.1:8182/callback",
		OAuthBaseURL: "https://api.schwabapi.com/v1/oauth",
	}

	tests := []struct {
		name          string
		config        Config
		wantErr       bool
		wantErrSubstr string
	}{
		{
			name:   "valid config",
			config: validConfig,
		},
		{
			name: "missing client id",
			config: Config{
				ClientSecret: "client-secret",
				CallbackURL:  "https://127.0.0.1:8182/callback",
			},
			wantErr:       true,
			wantErrSubstr: "client_id",
		},
		{
			name: "missing client secret",
			config: Config{
				ClientID:    "client-id",
				CallbackURL: "https://127.0.0.1:8182/callback",
			},
			wantErr:       true,
			wantErrSubstr: "client_secret",
		},
		{
			name: "missing callback url",
			config: Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
			},
			wantErr:       true,
			wantErrSubstr: "callback_url",
		},
		{
			name: "invalid callback url",
			config: Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				CallbackURL:  "://127.0.0.1",
			},
			wantErr: true,
		},
		{
			name: "callback url uses http scheme",
			config: Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				CallbackURL:  "http://127.0.0.1:8182/callback",
			},
			wantErr:       true,
			wantErrSubstr: "https",
		},
		{
			name: "callback url uses non-loopback host",
			config: Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				CallbackURL:  "https://localhost:8182/callback",
			},
			wantErr:       true,
			wantErrSubstr: "127.0.0.1",
		},
		{
			name: "callback url missing port",
			config: Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				CallbackURL:  "https://127.0.0.1/callback",
			},
			wantErr:       true,
			wantErrSubstr: "port",
		},
		{
			name: "invalid oauth base url",
			config: Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				CallbackURL:  "https://127.0.0.1:8182/callback",
				OAuthBaseURL: "://api.schwabapi.com/v1/oauth",
			},
			wantErr: true,
		},
		{
			name: "oauth base url uses http scheme for remote host",
			config: Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				CallbackURL:  "https://127.0.0.1:8182/callback",
				OAuthBaseURL: "http://api.schwabapi.com/v1/oauth",
			},
			wantErr:       true,
			wantErrSubstr: "https",
		},
		{
			name: "empty oauth base url",
			config: Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				CallbackURL:  "https://127.0.0.1:8182/callback",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("Validate(%+v) = %v, want nil", tt.config, err)
				}
				return
			}

			if err == nil {
				t.Fatalf("Validate(%+v) = nil, want error", tt.config)
			}
			if tt.wantErrSubstr != "" {
				if !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Fatalf("Validate(%+v) = %v, want error containing %q", tt.config, err, tt.wantErrSubstr)
				}
			}
		})
	}
}

func TestOAuthBaseURLFromAPIBaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		apiBaseURL    string
		want          string
		wantErrSubstr string
	}{
		{
			name: "empty uses production oauth base url",
			want: DefaultOAuthBaseURL,
		},
		{
			name:       "api origin adds oauth path",
			apiBaseURL: "https://api.schwabapi.com",
			want:       "https://api.schwabapi.com/v1/oauth",
		},
		{
			name:       "proxy prefix is preserved",
			apiBaseURL: "https://proxy.example.com/root/",
			want:       "https://proxy.example.com/root/v1/oauth",
		},
		{
			name:          "relative url rejected",
			apiBaseURL:    "api.schwabapi.com",
			wantErrSubstr: "scheme and host",
		},
		{
			name:          "http url rejected",
			apiBaseURL:    "http://api.schwabapi.com",
			wantErrSubstr: "https",
		},
		{
			name:          "query rejected",
			apiBaseURL:    "https://api.schwabapi.com?env=test",
			wantErrSubstr: "query or fragment",
		},
		{
			name:          "fragment rejected",
			apiBaseURL:    "https://api.schwabapi.com#oauth",
			wantErrSubstr: "query or fragment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := OAuthBaseURLFromAPIBaseURL(tt.apiBaseURL)
			if tt.wantErrSubstr == "" {
				require.NoError(t, err, "OAuthBaseURLFromAPIBaseURL(%q)", tt.apiBaseURL)
				require.Equal(t, tt.want, got, "OAuthBaseURLFromAPIBaseURL(%q)", tt.apiBaseURL)
				return
			}

			require.Error(t, err, "OAuthBaseURLFromAPIBaseURL(%q)", tt.apiBaseURL)
			require.ErrorContains(t, err, tt.wantErrSubstr, "OAuthBaseURLFromAPIBaseURL(%q)", tt.apiBaseURL)
		})
	}
}

func TestConfigFromAPIBaseURL(t *testing.T) {
	t.Parallel()

	got, err := ConfigFromAPIBaseURL(
		"client-id",
		"client-secret",
		"https://127.0.0.1:8182/callback",
		"https://proxy.example.com/root",
	)

	require.NoError(t, err)
	want := Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		CallbackURL:  "https://127.0.0.1:8182/callback",
		OAuthBaseURL: "https://proxy.example.com/root/v1/oauth",
	}
	require.Equal(t, want, got)
}

func TestConfigFromAPIBaseURL_InvalidConfig(t *testing.T) {
	t.Parallel()

	_, err := ConfigFromAPIBaseURL("", "client-secret", "https://127.0.0.1:8182/callback", "")

	require.Error(t, err)
	require.ErrorContains(t, err, "client_id")
}

func TestConfigFromAPIBaseURL_InvalidAPIBaseURL(t *testing.T) {
	t.Parallel()

	_, err := ConfigFromAPIBaseURL(
		"client-id",
		"client-secret",
		"https://127.0.0.1:8182/callback",
		"http://api.schwabapi.com",
	)

	require.Error(t, err)
	require.ErrorContains(t, err, "api_base_url")
}

//nolint:gocognit // Table-driven test with many cases; splitting would reduce readability.
func TestLoadConfig(t *testing.T) {
	t.Parallel()

	validConfig := Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		CallbackURL:  "https://127.0.0.1:8182/callback",
		OAuthBaseURL: "https://api.schwabapi.com/v1/oauth",
	}

	tests := []struct {
		name            string
		contents        string
		wantConfig      Config
		wantErr         bool
		wantErrContains []string
	}{
		{
			name:       "valid config",
			contents:   `{"client_id":"client-id","client_secret":"client-secret","callback_url":"https://127.0.0.1:8182/callback","oauth_base_url":"https://api.schwabapi.com/v1/oauth"}`,
			wantConfig: validConfig,
		},
		{
			name:     "oauth base url omitted",
			contents: `{"client_id":"client-id","client_secret":"client-secret","callback_url":"https://127.0.0.1:8182/callback"}`,
			wantConfig: Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				CallbackURL:  "https://127.0.0.1:8182/callback",
			},
		},
		{
			name:            "missing file",
			wantErr:         true,
			wantErrContains: []string{"missing.json", "read config"},
		},
		{
			name:            "invalid json",
			contents:        `{"client_id":"client-id",`,
			wantErr:         true,
			wantErrContains: []string{"parse config"},
		},
		{
			name:            "missing required fields",
			contents:        `{}`,
			wantErr:         true,
			wantErrContains: []string{"validate config", "client_id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "config.json")
			if tt.name == "missing file" {
				path = filepath.Join(tmpDir, "missing.json")
			} else {
				if err := os.WriteFile(path, []byte(tt.contents), 0o600); err != nil {
					t.Fatalf("os.WriteFile(%q) = %v, want nil", path, err)
				}
			}

			got, err := LoadConfig(path)
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("LoadConfig(%q) = %v, want nil", path, err)
				}
				if got != tt.wantConfig {
					t.Fatalf("LoadConfig(%q) = %+v, want %+v", path, got, tt.wantConfig)
				}
				return
			}

			if err == nil {
				t.Fatalf("LoadConfig(%q) = nil, want error", path)
			}
			for _, substr := range tt.wantErrContains {
				if !strings.Contains(err.Error(), substr) {
					t.Fatalf("LoadConfig(%q) = %v, want error containing %q", path, err, substr)
				}
			}
			if tt.name == "missing file" {
				var authRequired *AuthRequiredError
				if !errors.As(err, &authRequired) {
					t.Fatalf("LoadConfig(%q) = %T, want *AuthRequiredError", path, err)
				}
				if !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("LoadConfig(%q) = %v, want os.ErrNotExist in error chain", path, err)
				}
			}
		})
	}
}
