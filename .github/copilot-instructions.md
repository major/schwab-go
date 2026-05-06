# schwab-go review instructions

Review this repository as a Go client library for Charles Schwab APIs. The public API should be predictable for client applications: typed request parameters, typed response structs, context propagation, and explicit errors instead of process-level behavior.

Focus on bugs, security, data loss, broken API contracts, and project conventions. Do not nitpick formatting or style that `gofmt` or golangci-lint already handles.

## Project invariants

- Module path is `github.com/major/schwab-go`.
- Public packages live under `schwab`, `schwab/marketdata`, and `schwab/trader`.
- Library code must not call `os.Exit`, write user-facing output, read hidden config files, or inspect environment variables unless a public option documents that behavior.
- Public methods that perform requests take `context.Context` as the first argument.
- Request paths, query parameters, JSON tags, and response structs must match the Schwab API specification.
- Preserve typed errors from `schwab/errors.go` and wrap errors with enough context for callers to handle them.
- Keep exported identifiers documented with useful Go comments.

## Security and account safety

- Flag credential, bearer token, account hash, or secret exposure in logs, errors, tests, docs, or generated output.
- Verify account-scoped trader calls keep account hash values in path parameters only when required by the API.
- Order placement, replacement, cancellation, and preview methods must not invent safety shortcuts or silently mutate payloads.
- Avoid silent fallback behavior around HTTP status handling, body decoding, or token application. Return clear errors instead.

## Testing expectations

- Use `testify/require` for assertions that must stop a test and `testify/assert` for non-critical checks.
- Use `httptest.NewServer()` for HTTP API mocks with inline request validation.
- Mark reusable test helpers with `t.Helper()`.
- Prefer table-driven subtests with `t.Run()`.
- Keep generated response data inline unless fixtures clearly improve readability.
- Verify request methods, paths, query parameters, headers, and decoded response fields.

## Build and lint expectations

- CI runs `go test -v -race -coverprofile=coverage.out ./...`, `go build ./...`, govulncheck, CodeQL, and golangci-lint v2.
- GoReleaser is source-only because this repository is a library with no `main` package.
- Nolint directives require a specific linter name and an explanation.
- US English spelling is enforced.
