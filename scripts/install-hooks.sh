#!/usr/bin/env bash
set -euo pipefail

echo "Installing Git hooks..."

# Install pre-commit if not installed
if ! command -v pre-commit &> /dev/null; then
    echo "Installing pre-commit..."
    if command -v brew &> /dev/null; then
        brew install pre-commit
    elif command -v pip3 &> /dev/null; then
        pip3 install --user pre-commit
    else
        echo "Please install pre-commit manually: https://pre-commit.com/"
        exit 1
    fi
fi

# Install golangci-lint if not installed
if ! command -v golangci-lint &> /dev/null; then
    echo "Installing golangci-lint..."
    if command -v brew &> /dev/null; then
        brew install golangci-lint
    else
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
    fi
fi

# Install hadolint if not installed (for Docker linting)
if ! command -v hadolint &> /dev/null; then
    echo "Installing hadolint..."
    if command -v brew &> /dev/null; then
        brew install hadolint
    fi
fi

# Install pre-commit hooks
pre-commit install --install-hooks
pre-commit install --hook-type commit-msg

# Create initial secrets baseline
if [ ! -f .secrets.baseline ]; then
    detect-secrets scan --baseline .secrets.baseline || true
fi

echo "✅ Git hooks installed successfully!"
echo ""
echo "Hooks will run automatically on git commit."
echo "To run hooks manually: pre-commit run --all-files"
echo "To update hooks: pre-commit autoupdate"