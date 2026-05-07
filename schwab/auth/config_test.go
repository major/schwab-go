package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		})
	}
}
