---
applyTo: "Makefile"
---

# Makefile review instructions

- Makefile targets should have matching `.PHONY` declarations when they are not real files.
- Avoid adding flags that are already defaults.
- Keep task names aligned with CI workflow usage.
- Build targets should validate the library with `go build ./...`; do not assume a CLI binary exists.
