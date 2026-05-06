.PHONY: build test lint vuln clean release

build:
	go build ./...

test:
	go run gotest.tools/gotestsum@v1.13.0 --junitfile junit.xml -- -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

vuln:
	govulncheck ./...

clean:
	go clean
	rm -f coverage.out junit.xml
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
