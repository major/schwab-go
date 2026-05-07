# Auth CLI adapter pattern

This package owns Schwab OAuth2 primitives: authorization URLs, loopback callbacks, code exchange, token refresh, token persistence, and `schwab.WithTokenProvider` integration. CLI applications own command shape, config policy, output envelopes, exit codes, browser behavior, and post-login setup.

This document is an adapter recipe, not a finished `schwab-agent` migration. Keep the concrete Cobra commands, JSON envelopes, exit-code constants, environment variable names, and default-account persistence in `schwab-agent` or another CLI so this library stays reusable and free of hidden config or stdout/stderr behavior.

## Command layout

Model auth commands as a thin adapter over `schwab/auth`:

```text
auth login [--config PATH] [--token PATH] [--no-browser]
auth status [--config PATH] [--token PATH]
auth refresh [--config PATH] [--token PATH]
```

`auth login` should load application config, create `auth.NewFileTokenStore(tokenPath)`, pass a `urlHandler` to `auth.Login`, then run any application setup. `auth status` should create a provider with `auth.NewProvider(cfg, store, httpClient)` and call `Provider.Status(ctx, now)` to report token age and expiry without refreshing or saving. `auth refresh` should create a provider and call `Provider.Refresh(ctx)` so an explicit refresh command always refreshes with the stored refresh token and persists the result.

## Global auth gate

Cobra CLIs usually enforce auth in `PersistentPreRunE` for commands that call Schwab APIs:

```go
func requireAuth(ctx context.Context, cfg auth.Config, tokenPath string) (*auth.Provider, string, error) {
	store := auth.NewFileTokenStore(tokenPath)
	provider, err := auth.NewProvider(cfg, store, nil)
	if err != nil {
		return nil, "", err
	}

	accessToken, err := provider.Token(ctx)
	if err != nil {
		return nil, "", err
	}

	return provider, accessToken, nil
}
```

Commands that do not need a bearer token should opt out before the global gate runs. Common bypasses are `help`, shell completion, symbol parsing/building commands that do not call Schwab, and the `auth` command group itself.

## Config and token paths

Keep config precedence in the CLI adapter. The library intentionally does not inspect environment variables or hidden config files.

Recommended order for CLI apps:

1. Start with explicit flags such as `--config` and `--token`.
2. Load file config when the file exists. If environment variables are allowed to create a config from scratch, read JSON into your own app config type instead of calling `auth.LoadConfig` first, because `auth.LoadConfig` validates immediately.
3. Apply app-owned environment overrides such as `SCHWAB_CLIENT_ID`, `SCHWAB_CLIENT_SECRET`, or app-specific base URL settings if your CLI documents them.
4. Convert the final app config into `auth.Config` and validate it before calling `auth.NewProvider`, `auth.Login`, or `auth.NewFileTokenStore`.

This keeps reusable OAuth logic in `schwab/auth` while preserving each application’s config policy.

## Browser and headless login

For an explicit `auth refresh` command, call `Provider.Refresh(ctx)`. It always refreshes with the stored refresh token, saves through the configured token store, and preserves the original `TokenFile.CreationTimestamp` so refresh-token age stays accurate. Keep direct `auth.RefreshAccessToken` calls for low-level integrations that intentionally manage token loading, saving, and timestamp preservation themselves. `Provider.Token(ctx)` is better for a global auth gate because it returns an existing access token until it is inside the expiry buffer.

`auth.Login` accepts a `urlHandler func(string) error`. Browser CLIs can open the URL with `xdg-open`, `open`, `rundll32`, or a user-selected browser command. Headless CLIs should emit the URL through their normal output path and let the user open it elsewhere.

Use a context timeout for callback waits:

```go
ctx, cancel := context.WithTimeout(parentCtx, 5*time.Minute)
defer cancel()

provider, err := auth.Login(ctx, cfg, store, handleAuthorizeURL)
```

The callback URL must exactly match the Schwab developer portal value, use `https`, include an explicit port, and bind to `127.0.0.1`. Local browser flows use a self-signed loopback certificate, so browsers may show a certificate warning before redirecting back to the CLI.

## Status reporting

`Provider.Status(ctx, now)` loads the configured token store and returns `auth.TokenStatus` without refreshing or saving. Missing token files map to `TokenStatus{LoginRequired: true}` with no error so CLIs can render normal status output for an unauthenticated installation. Other token-store load errors are returned so the adapter can report parse or filesystem failures.

`auth.InspectToken(tokenFile, now)` is the pure helper behind `Provider.Status`. It reports `AccessTokenExpiresAt`, `AccessTokenExpired`, `RefreshTokenCreatedAt`, `RefreshTokenExpiresAt`, `RefreshTokenStale`, `CanRefresh`, and `LoginRequired`. Use those fields to build the CLI's existing JSON envelope or table output, but keep output names and exit-code policy in the adapter.

Suggested CLI interpretation:

| Status field | CLI meaning |
|---|---|
| `LoginRequired` | The user needs `auth login` before API calls can succeed. |
| `AccessTokenExpired` with `CanRefresh` | API calls through `Provider.Token` can refresh automatically; an explicit status command should report the expiry without mutating the token file. |
| `RefreshTokenStale` | The refresh token is too old to rely on; prompt for login. |
| `CanRefresh` | An explicit `auth refresh` command can call `Provider.Refresh(ctx)`. |

## Structured output and exit codes

`schwab/auth` returns Go errors. CLIs should translate those errors into stable command output and exit codes at the adapter boundary.

Example JSON error envelope:

```json
{
  "ok": false,
  "error": {
    "code": "auth_required",
    "message": "Run auth login before calling Schwab APIs."
  }
}
```

Suggested mappings:

| Error | CLI code | Remediation |
|---|---|---|
| `*auth.AuthRequiredError` from missing config or token file | `auth_required` | Run `auth login` or configure explicit paths. |
| Config validation error from `auth.LoadConfig` | `auth_config_invalid` | Fix client credentials or callback URL. |
| `*auth.AuthExpiredError` | `auth_expired` | Run `auth login` again because the refresh token is stale or revoked. |
| `*auth.AuthCallbackError` | `auth_callback_failed` | Retry login and confirm the callback URL matches the Schwab app. |
| Other refresh or exchange error | `auth_refresh_failed` | Show the status/message and preserve logs for debugging. |

Use `errors.As` to classify the auth error while `errors.Is` can inspect joined or wrapped underlying causes when the error includes one.

## Post-login setup

After successful login, the CLI can run application-specific setup with the returned provider. For schwab-agent parity, create a Trader client, call `GetAccountNumbers`, and persist a default account only when Schwab returns exactly one account:

```go
provider, err := auth.Login(ctx, cfg, store, handleAuthorizeURL)
if err != nil {
	return err
}

client := trader.NewClient(schwab.WithTokenProvider(provider))
accounts, err := client.GetAccountNumbers(ctx)
if err != nil {
	return err
}
if len(accounts) == 1 {
	appConfig.DefaultAccount = accounts[0].HashValue
	return saveAppConfig(appConfigPath, appConfig)
}
```

Keep this hook in the CLI application because default-account storage, JSON envelopes, and safety prompts are application policy, not Schwab OAuth behavior.
