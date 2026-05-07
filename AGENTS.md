# Repository instructions

- Keep this file updated whenever project commands, package layout, CI, release flow, or repo-specific conventions change.
- Module path is `github.com/major/schwab-go`; `go.mod` currently targets Go 1.26 and uses `testify` plus `kin-openapi` for tests.
- This is a library only. There is no `main` package or CLI, and GoReleaser is configured for source archives only.

## Commands

- `make test` runs the CI test command through gotestsum: `go run gotest.tools/gotestsum@v1.13.0 --junitfile junit.xml -- -v -race -coverprofile=coverage.out ./...`.
- `make build` runs `go build ./...`; use this instead of looking for a binary target.
- `make lint` runs `golangci-lint run ./...`; CI uses golangci-lint v2.12.2.
- `make spec-validate` validates `docs/*.openapi.json` with pinned `kin-openapi` and `vacuum` versions. `kin-openapi` disables example validation because Schwab's captured specs contain invalid example metadata; `.vacuum.yaml` keeps non-blocking warnings visible while disabling known upstream portal defects that should not fail local validation.
- `make vuln` runs `govulncheck ./...`.
- `make clean` removes `coverage.out`, `junit.xml`, and `dist/`.
- `make release VERSION=vX.Y.Z` must run from a clean `main` branch, runs test then lint, and creates a signed tag. Push the tag manually.

## Tool version drift

- Before trusting local tool output, compare local versions against repo-pinned versions in `go.mod`, `.github/workflows/ci.yml`, `Makefile`, and other checked-in automation.
- Warn the user when a local tool is older or newer than the repo/CI pin. Version drift can change lint findings or config compatibility.
- For golangci-lint specifically, check `golangci-lint version` against the CI `golangci/golangci-lint-action` `version:` value before running `make lint`. If they differ, say which version is local, which version CI uses, and whether the command was verified with the pinned version, the local version, or both.

## Git workflow

- Always use Conventional Commits style for commit messages.

## Layout and API boundaries

- Public packages are `schwab` for shared types/options/errors, `schwab/auth` for OAuth2 token acquisition and refresh, `schwab/marketdata` for Market Data API endpoints, and `schwab/trader` for Trader API endpoints.
- Subpackages import the root package as `schwab "github.com/major/schwab-go/schwab"`.
- Client construction uses shared functional options from `schwab/options.go`: `WithToken`, `WithTokenProvider`, `WithHTTPClient`, `WithBaseURL`, and `ApplyOptions`.
- OAuth and token refresh live in the explicit `schwab/auth` package. API clients still accept bearer tokens through `WithToken` or `WithTokenProvider`; do not add hidden environment or config-file reads.
- Default API bases are `https://api.schwabapi.com/marketdata/v1` and `https://api.schwabapi.com/trader/v1`; tests override them with `WithBaseURL(ts.URL)`.

## Implementation rules

- Library code must not call `os.Exit`, write user-facing stdout/stderr output, read hidden config files, or inspect environment variables unless a documented public option is added.
- Public request methods take `context.Context` as the first argument.
- Keep request paths, query parameter names, JSON tags, and response structs aligned with Schwab API names. Avoid speculative fields unless docs, fixtures, or observed API data prove them.
- Trader account-scoped paths use encrypted account hashes, not plain account numbers. Get hashes through `GetAccountNumbers()` and escape path segments with `url.PathEscape`.
- Build query strings through `req.URL.Query()` and `req.URL.RawQuery = q.Encode()`, not by concatenating query text into paths.
- `Content-Type: application/json` is set only when a request body is present. Success endpoints with empty bodies should call `do(req, nil)`.
- Preserve `*schwab.APIError` for HTTP errors. Market Data decodes `detail`/`title`; Trader decodes `message`/`error` and may receive empty error bodies.
- Use `errors.AsType[*schwab.APIError](err)` for typed error extraction in Go 1.26 code.
- `QuoteEntry.Quote` and `QuoteEntry.Reference` are `json.RawMessage` because quote payloads vary by asset type; use the typed accessors after checking `AssetMainType`.

## Tests and review sources

- Tests use same-package test files, `httptest.NewServer`, inline request validation, `schwab.WithHTTPClient(ts.Client())`, and `schwab.WithBaseURL(ts.URL)`.
- Use `require` for assertions that must stop the test and `assert` for field checks that can continue.
- Reusable assertion helpers should call `t.Helper()`; existing fixtures are inline Go helpers, not `testdata/` files.
- CI also runs CodeQL, OpenSSF Scorecard, `govulncheck`, `go build ./...`, and Codecov coverage/test-results uploads when `CODECOV_TOKEN` is present.
- Preserve high-signal guidance in `.github/copilot-instructions.md`, `.github/instructions/*.md`, `.coderabbit.yaml`, `Makefile`, and `.goreleaser.yml` when changing related areas.
- Keep `README.md` up to date when adding, removing, or renaming public API methods, changing package structure, updating badges, or modifying build/install requirements.

## Runnable examples

- Public API additions or behavior changes should include or update runnable Go examples when the usage is non-obvious.
- Prefer `Example...` functions in `*_test.go` files with `// Output:` comments so `go test ./...` verifies the examples.
- Examples should demonstrate public API usage from a consumer perspective, using external test packages such as `trader_test` when practical.
- Keep examples focused on common workflows. Use normal `Test...` functions for edge cases, error paths, and exact payload assertions.
- When changing request/response models, order builders, auth helpers, or client options, review existing examples for drift.
- Trader order examples must use fake data and local `httptest` servers only. Cover enum families with catalog examples when no documented payload exists, and do not present an order shape as Schwab-accepted unless it is backed by docs, fixtures, or observed API data.

## Agent research modes

### Search mode

- Maximize search effort. Launch multiple background agents in parallel:
  - Explore agents for codebase patterns, file structures, and AST-grep searches.
  - Librarian agents for remote repositories, official docs, and GitHub examples.
- Use direct tools as well: `Grep`, `rg`, and `sg` or AST-grep.
- Never stop at the first result. Search exhaustively until repeated searches stop producing new useful context.

### Analyze mode

- Gather context before diving deep:
  - Launch 1-2 explore agents for codebase patterns and implementations.
  - Launch 1-2 librarian agents when external libraries, APIs, or docs are involved.
  - Use direct tools such as `Grep`, AST-grep, and LSP for targeted searches.
- If the problem is complex, consult specialists instead of struggling alone:
  - Oracle for conventional architecture, debugging, or complex logic problems.
  - Artistry for non-conventional problems that may need a different approach.
- Synthesize findings before proceeding.

### Delegation requirements

- When delegating, always include `load_skills=[]` and `run_in_background`.
- Example: `delegate_task(subagent_type="explore", prompt="...", run_in_background=true, load_skills=[])`.
