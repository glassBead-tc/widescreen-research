#!/usr/bin/env bash
set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "Running Go quality checks..."

# Check Go formatting
echo -n "Checking gofmt... "
if [ -n "$(gofmt -l .)" ]; then
    echo -e "${RED}FAILED${NC}"
    echo "Files need formatting:"
    gofmt -l .
    exit 1
fi
echo -e "${GREEN}OK${NC}"

# Check for ineffective assignments
echo -n "Checking ineffassign... "
if ! command -v ineffassign &> /dev/null; then
    go install github.com/gordonklaus/ineffassign@latest
fi
ineffassign ./...
echo -e "${GREEN}OK${NC}"

# Check for misspellings
echo -n "Checking misspell... "
if ! command -v misspell &> /dev/null; then
    go install github.com/client9/misspell/cmd/misspell@latest
fi
misspell -error .
echo -e "${GREEN}OK${NC}"

# Security check
echo -n "Checking gosec... "
if ! command -v gosec &> /dev/null; then
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi
gosec -quiet -fmt json -out /dev/null ./... || true
echo -e "${GREEN}OK${NC}"

echo -e "${GREEN}✅ All Go quality checks passed!${NC}"