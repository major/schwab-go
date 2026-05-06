# renovate: datasource=go depName=github.com/getkin/kin-openapi
KIN_OPENAPI_VERSION := v0.137.0
# renovate: datasource=go depName=github.com/daveshanley/vacuum
VACUUM_VERSION := v0.18.0
OPENAPI_SPECS := docs/market_data.openapi.json docs/trader_api.openapi.json

.PHONY: build test lint spec-validate vuln clean release

build:
	go build ./...

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

spec-validate:
	set -e; for spec in $(OPENAPI_SPECS); do go run github.com/getkin/kin-openapi/cmd/validate@$(KIN_OPENAPI_VERSION) -examples=false -- "$$spec"; done
	go run github.com/daveshanley/vacuum@$(VACUUM_VERSION) lint --ruleset .vacuum.yaml $(OPENAPI_SPECS)

vuln:
	govulncheck ./...

clean:
	go clean
	rm -f coverage.out
	rm -rf dist/

release:
ifndef VERSION
	$(error VERSION is required. Usage: make release VERSION=v0.3.0)
endif
	@[ "$$(git branch --show-current)" = "main" ] || { echo "Error: must be on main branch"; exit 1; }
	@[ -z "$$(git status --porcelain)" ] || { echo "Error: working tree is not clean"; exit 1; }
	@$(MAKE) test
	@$(MAKE) lint
	@PREV=$$(git describe --tags --abbrev=0 2>/dev/null || true); \
	if [ -n "$$PREV" ]; then RANGE="$$PREV..HEAD"; else RANGE=""; fi; \
	FEATS=$$(git log --oneline --grep='^feat' $$RANGE | sed 's/^[a-f0-9]* /- /'); \
	FIXES=$$(git log --oneline --grep='^fix' $$RANGE | sed 's/^[a-f0-9]* /- /'); \
	OTHER=$$(git log --oneline --grep='^feat' --grep='^fix' --invert-grep $$RANGE | sed 's/^[a-f0-9]* /- /'); \
	MSG="Release $(VERSION)"; \
	if [ -n "$$FEATS" ]; then MSG="$$MSG\n\nFeatures:\n$$FEATS"; fi; \
	if [ -n "$$FIXES" ]; then MSG="$$MSG\n\nFixes:\n$$FIXES"; fi; \
	if [ -n "$$OTHER" ]; then MSG="$$MSG\n\nOther:\n$$OTHER"; fi; \
	printf '%b\n' "$$MSG" | git tag -s -F - $(VERSION)
	@echo ""
	@echo "Tag $(VERSION) created. Push with: git push origin $(VERSION)"
