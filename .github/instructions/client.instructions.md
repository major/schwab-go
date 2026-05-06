---
applyTo: "schwab/**/client.go"
---

# Client review instructions

- Preserve context propagation for cancellation and request timeouts.
- Response bodies and idle connections must be closed where required.
- HTTP errors should map to typed library errors with useful caller-facing context.
- Client options should be explicit and testable. Do not add hidden environment or config-file behavior.
