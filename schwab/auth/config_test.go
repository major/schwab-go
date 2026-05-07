package auth

import (
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
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			if tt.wantErrSubstr != "" {
				require.ErrorContains(t, err, tt.wantErrSubstr)
			}
		})
	}
}
