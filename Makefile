.PHONY: build run-coordinator run-drone test docker install-hooks lint lint-go lint-docker security-scan clean-hooks pre-push ci ci-quick verify-hooks mcp-sdk-migrate

# Build targets
build:
	go build ./cmd/coordinator ./cmd/drone

# MCP Utilities
mcp-sdk-migrate:
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make mcp-sdk-migrate FILE=path/to/file.go [DRY_RUN=1]"; \
		echo ""; \
		echo "Migrates Go MCP server code from mark3labs to official SDK"; \
		echo ""; \
		echo "Examples:"; \
		echo "  make mcp-sdk-migrate FILE=pkg/mcp/server.go"; \
		echo "  make mcp-sdk-migrate FILE=pkg/mcp/server.go DRY_RUN=1"; \
		exit 1; \
	fi
	@go run ./cmd/mcp-sdk-migrate -file $(FILE) $(if $(DRY_RUN),-dry-run,)

run-coordinator:
	LOG_LEVEL=debug \
	GOOGLE_CLOUD_PROJECT=your-project-id \
	GOOGLE_CLOUD_REGION=us-central1 \
	go run ./cmd/coordinator

run-drone:
	LOG_LEVEL=debug DRONE_TYPE=research EXA_API_KEY=dummy \
	DRONE_ID=local-drone-1 \
	GOOGLE_CLOUD_PROJECT=your-project-id \
	PUBSUB_TOPIC=drone-results \
	go run ./cmd/drone

test:
	go test -race -cover ./...

docker:
	docker build -t widescreen/coordinator:dev .

# Quality enforcement
install-hooks:
	@bash scripts/install-hooks.sh

lint: lint-go lint-docker
	@pre-commit run --all-files

lint-go:
	@echo "Running Go linters..."
	@golangci-lint run --fix

lint-docker:
	@echo "Running Docker linters..."
	@hadolint Dockerfile* 2>/dev/null || echo "No Dockerfiles to lint"

security-scan:
	@echo "Running security scans..."
	@gosec ./...
	@detect-secrets scan --baseline .secrets.baseline

clean-hooks:
	@pre-commit uninstall
	@rm -rf .git/hooks/pre-commit
	@rm -rf .git/hooks/commit-msg

# Run before pushing
pre-push: lint test security-scan
	@echo "✅ Ready to push!"

# Local CI - matches GitHub CI workflow exactly
ci:
	@echo "════════════════════════════════════════"
	@echo "Running Local CI (matches GitHub CI)"
	@echo "════════════════════════════════════════"
	@echo ""
	@echo "→ Step 1/6: Checking Go formatting..."
	@bash -c 'fmt_output=$$(gofmt -s -l .); if [ -n "$$fmt_output" ]; then echo "❌ Files not formatted:"; echo "$$fmt_output"; echo ""; echo "Run: gofmt -s -w ."; exit 1; else echo "✅ Formatting OK"; fi'
	@echo ""
	@echo "→ Step 2/6: Running go vet..."
	@go vet ./...
	@echo "✅ go vet passed"
	@echo ""
	@echo "→ Step 3/6: Running staticcheck..."
	@staticcheck ./... || (echo "❌ staticcheck failed" && exit 1)
	@echo "✅ staticcheck passed"
	@echo ""
	@echo "→ Step 4/6: Running govulncheck..."
	@bash -c 'if ! command -v govulncheck &> /dev/null; then echo "Installing govulncheck..."; go install golang.org/x/vuln/cmd/govulncheck@latest; fi'
	@$$(go env GOPATH)/bin/govulncheck ./...
	@echo "✅ govulncheck passed"
	@echo ""
	@echo "→ Step 5/6: Building binaries..."
	@go build ./cmd/coordinator ./cmd/drone
	@echo "✅ Build succeeded"
	@echo ""
	@echo "→ Step 6/6: Running tests with race detection..."
	@go test -race -cover ./...
	@echo "✅ Tests passed"
	@echo ""
	@echo "════════════════════════════════════════"
	@echo "🎉 All CI checks passed! Safe to push."
	@echo "════════════════════════════════════════"

# Quick CI - faster version without govulncheck
ci-quick:
	@echo "Running quick CI checks..."
	@gofmt -s -l . | (! grep .) || (echo "❌ Run: gofmt -s -w ." && exit 1)
	@go vet ./...
	@staticcheck ./...
	@go build ./cmd/coordinator ./cmd/drone
	@go test -race -cover ./...
	@echo "✅ Quick CI passed"

# Verify git hooks are installed and working
verify-hooks:
	@echo "Verifying git hooks installation..."
	@test -f .git/hooks/pre-commit || (echo "❌ pre-commit hook missing - run 'make install-hooks'" && exit 1)
	@test -x .git/hooks/pre-commit || (echo "❌ pre-commit hook not executable" && exit 1)
	@grep -q "pre-commit" .git/hooks/pre-commit || (echo "❌ pre-commit hook invalid" && exit 1)
	@test -f .git/hooks/commit-msg || (echo "❌ commit-msg hook missing - run 'make install-hooks'" && exit 1)
	@echo "✅ Git hooks installed and active"
	@echo ""
	@echo "Installed hooks:"
	@echo "  • pre-commit: formats code, checks secrets, validates YAML/JSON"
	@echo "  • commit-msg: enforces conventional commit format"
	@echo ""
	@echo "To test hooks, try committing unformatted code (it will be blocked)"

# Help
help:
	@echo "Available targets:"
	@echo "  build               - Build Go binaries"
	@echo "  run-coordinator     - Run coordinator locally"
	@echo "  run-drone           - Run drone locally"
	@echo "  test                - Run tests"
	@echo "  docker              - Build Docker image"
	@echo "  install-hooks       - Install Git hooks"
	@echo "  verify-hooks        - Verify git hooks are installed and active"
	@echo "  ci                  - Run full CI checks (matches GitHub CI)"
	@echo "  ci-quick            - Run quick CI checks (no govulncheck)"
	@echo "  lint                - Run all linters"
	@echo "  lint-go             - Run Go linters"
	@echo "  lint-docker         - Run Docker linters"
	@echo "  security-scan       - Run security scans"
	@echo "  pre-push            - Run all checks before pushing"
	@echo "  clean-hooks         - Remove Git hooks"
	@echo "  mcp-sdk-migrate     - Migrate MCP server from mark3labs to official SDK"
	@echo "                        Usage: make mcp-sdk-migrate FILE=path/to/file.go"
