#!/bin/bash
# MCP SDK Migration Slash Command
# Converts mark3labs MCP Go SDK to official modelcontextprotocol SDK

set -e

# Accept arguments from command line or environment
FILE_PATH="${1:-$FILE}"
DRY_RUN="${2:-$DRY_RUN}"
BACKUP_EXT="${BACKUP_EXT:-.backup}"
VERBOSE="${VERBOSE:-}"

# Get the repository root (where this script is located)
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

# Check if we're in the right directory
if [ ! -f "$REPO_ROOT/go.mod" ]; then
    echo "Error: Not in a Go project root"
    exit 1
fi

if [ -z "$FILE_PATH" ]; then
    echo "Usage: /mcp-sdk-migrate <file_path> [dry-run]"
    echo ""
    echo "Arguments:"
    echo "  <file_path>  - Path to Go file to migrate (required)"
    echo "  [dry-run]    - Preview changes without modifying (optional)"
    echo ""
    echo "Environment Variables:"
    echo "  FILE         - Alternative way to specify file path"
    echo "  DRY_RUN      - Set to '1' or 'true' for dry-run mode"
    echo "  BACKUP_EXT   - Backup file extension (default: .backup)"
    echo "  VERBOSE      - Set to '1' for verbose output"
    echo ""
    echo "Examples:"
    echo "  /mcp-sdk-migrate cmd/my-server/main.go"
    echo "  /mcp-sdk-migrate pkg/mcp/server.go dry-run"
    echo "  FILE=pkg/mcp/server.go DRY_RUN=1 /mcp-sdk-migrate"
    echo "  BACKUP_EXT=.old /mcp-sdk-migrate cmd/server/main.go"
    exit 1
fi

# Make path absolute if relative
if [[ ! "$FILE_PATH" = /* ]]; then
    FILE_PATH="$REPO_ROOT/$FILE_PATH"
fi

if [ ! -f "$FILE_PATH" ]; then
    echo "Error: File not found: $FILE_PATH"
    exit 1
fi

# Build and run the migration tool
cd "$REPO_ROOT"

[ -n "$VERBOSE" ] && echo "Repository root: $REPO_ROOT"
[ -n "$VERBOSE" ] && echo "Target file: $FILE_PATH"
[ -n "$VERBOSE" ] && echo "Backup extension: $BACKUP_EXT"

echo "🔄 Running MCP SDK migration tool..."
echo ""

# Build the tool if needed
if [ ! -f "./mcp-sdk-migrate" ] || [ "./cmd/mcp-sdk-migrate/main.go" -nt "./mcp-sdk-migrate" ]; then
    echo "Building migration tool..."
    go build -o mcp-sdk-migrate ./cmd/mcp-sdk-migrate
fi

# Build command arguments
TOOL_ARGS="-file $FILE_PATH -backup $BACKUP_EXT"

# Add dry-run flag if requested
if [ "$DRY_RUN" = "dry-run" ] || [ "$DRY_RUN" = "1" ] || [ "$DRY_RUN" = "true" ]; then
    TOOL_ARGS="$TOOL_ARGS -dry-run"
fi

[ -n "$VERBOSE" ] && echo "Running: ./mcp-sdk-migrate $TOOL_ARGS"

# Run the migration
./mcp-sdk-migrate $TOOL_ARGS

echo ""
echo "📖 For complete migration guide, see: docs/mcp-server-migration-spec.md"
