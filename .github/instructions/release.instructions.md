---
applyTo: ".goreleaser.yml"
---

# Release review instructions

- Release automation must keep GoReleaser v2 behavior and keyless cosign signing intact.
- This repository is a Go library, so GoReleaser should skip binary builds and publish source archives with checksums.
- Do not add a binary platform matrix unless the repository gains a `main` package.
