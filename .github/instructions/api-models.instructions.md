---
applyTo: "schwab/**/*.go"
---

# Schwab API model review instructions

- Structs should match Schwab API JSON field names and avoid silently dropping important response fields.
- Public request parameter structs should model documented query and path parameters accurately.
- Changes to request payload structs need tests that prove the emitted JSON or query shape.
- Avoid speculative fields unless Schwab API behavior, documentation, or fixtures show they are needed.
