.PHONY: build run-coordinator run-drone test docker install-hooks lint lint-go lint-docker security-scan clean-hooks pre-push mcp-sdk-migrate

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

# Help
help:
	@echo "Available targets:"
	@echo "  build               - Build Go binaries"
	@echo "  run-coordinator     - Run coordinator locally"
	@echo "  run-drone           - Run drone locally"
	@echo "  test                - Run tests"
	@echo "  docker              - Build Docker image"
	@echo "  install-hooks       - Install Git hooks"
	@echo "  lint                - Run all linters"
	@echo "  lint-go             - Run Go linters"
	@echo "  lint-docker         - Run Docker linters"
	@echo "  security-scan       - Run security scans"
	@echo "  pre-push            - Run all checks before pushing"
	@echo "  clean-hooks         - Remove Git hooks"
	@echo "  mcp-sdk-migrate     - Migrate MCP server from mark3labs to official SDK"
	@echo "                        Usage: make mcp-sdk-migrate FILE=path/to/file.go"
