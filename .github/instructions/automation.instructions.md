---
applyTo: ".github/workflows/**"
---

# GitHub Actions review instructions

- Validate GitHub Actions syntax, minimum permissions, and secret handling.
- Actions should be pinned consistently with the existing workflow style.
- Avoid workflow changes that spend extra CI minutes without a clear project benefit.
- Build verification should use `go build ./...` for this library repository.
