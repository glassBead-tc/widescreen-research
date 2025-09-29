# MCP SDK Migration Utility

A command-line tool for automatically migrating Go MCP server code from the unofficial `mark3labs/mcp-go` SDK to the official `modelcontextprotocol/go-sdk`.

## Installation

### As Project Utility

The tool is already available in this repository:

```bash
# Via Makefile (recommended)
make mcp-sdk-migrate FILE=path/to/file.go

# Build standalone binary
go build -o mcp-sdk-migrate ./cmd/mcp-sdk-migrate
./mcp-sdk-migrate -file path/to/file.go
```

### As Slash Command (Claude Code)

Available as `/mcp-sdk-migrate` in Claude Code when working in this repository:

```
/mcp-sdk-migrate pkg/mcp/server.go
/mcp-sdk-migrate pkg/mcp/server.go dry-run
```

## Usage

### Command Line

```bash
# Basic migration
./mcp-sdk-migrate -file cmd/server/main.go

# Preview changes without modifying (dry-run)
./mcp-sdk-migrate -file cmd/server/main.go -dry-run

# Custom backup extension
./mcp-sdk-migrate -file cmd/server/main.go -backup .old

# Show help
./mcp-sdk-migrate -help
```

### Makefile

```bash
# Migrate a file
make mcp-sdk-migrate FILE=pkg/mcp/server.go

# Preview changes
make mcp-sdk-migrate FILE=pkg/mcp/server.go DRY_RUN=1
```

## What Gets Migrated Automatically

The tool performs the following automatic transformations:

### 1. Import Statements

**Before:**

```go
import (
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)
```

**After:**

```go
import (
    "context"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)
```

### 2. Server Creation

**Before:**

```go
server := server.NewMCPServer("my-server", "1.0.0",
    server.WithToolCapabilities(true),
    server.WithRecovery())
```

**After:**

```go
server := mcp.NewServer(&mcp.Implementation{
    Name:    "my-server",
    Version: "1.0.0",
}, nil)
```

### 3. Tool Result Constructors

**Before:**

```go
return mcp.NewToolResultText("Success!"), nil
return mcp.NewToolResultError("Failed"), nil
```

**After:**

```go
return &mcp.CallToolResult{
    Content: []mcp.Content{
        &mcp.TextContent{Text: "Success!"},
    },
}, nil, nil

return &mcp.CallToolResult{
    IsError: true,
    Content: []mcp.Content{
        &mcp.TextContent{Text: "Failed"},
    },
}, nil, nil
```

### 4. Server Start

**Before:**

```go
if err := server.ServeStdio(server); err != nil {
    log.Fatal(err)
}
```

**After:**

```go
if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
    log.Fatalf("Server error: %v", err)
}
```

## Manual Changes Required

Some patterns require manual intervention:

### Tool Registration

**Before (mark3labs):**

```go
tool := mcp.NewTool("my_tool",
    mcp.WithDescription("My tool"),
    mcp.WithString("param", mcp.Required()))

server.AddTool(tool, handlerFunc)
```

**After (official SDK):**

```go
type MyToolArgs struct {
    Param string `json:"param" jsonschema:"required,description=Parameter description"`
}

mcp.AddTool(server, &mcp.Tool{
    Name:        "my_tool",
    Description: "My tool",
}, handlerFunc)
```

### Handler Signatures

**Before:**

```go
func handler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    param, _ := req.RequireString("param")
    // ... logic
}
```

**After:**

```go
func handler(ctx context.Context, req *mcp.CallToolRequest, args MyToolArgs) (*mcp.CallToolResult, any, error) {
    // Use args.Param directly
    // ... logic
}
```

## Backup and Safety

- **Automatic Backups**: Creates `.backup` files before modifying
- **Dry-Run Mode**: Preview changes without modifying files
- **Non-Destructive**: Original files preserved in backups

## Examples

### Migrate Single File

```bash
$ ./mcp-sdk-migrate -file pkg/mcp/server.go
🔄 Running MCP SDK migration tool...
📦 Backup created: pkg/mcp/server.go.backup
✅ Successfully migrated pkg/mcp/server.go
   4 changes applied
   1. Updated mcp import to official SDK
   2. Removed mark3labs server import
   3. Updated server creation to official SDK pattern
   4. Updated ServeStdio to Run with StdioTransport
```

### Preview Changes (Dry-Run)

```bash
$ ./mcp-sdk-migrate -file pkg/mcp/server.go -dry-run
🔍 Dry run - changes that would be made:
1. Updated mcp import to official SDK
2. Removed mark3labs server import
3. Updated server creation to official SDK pattern
4. Updated NewToolResultText calls

--- Migrated Content ---
[Shows full migrated code without modifying file]
```

## Distribution

This utility can be used in other MCP Go projects:

### Standalone Binary

```bash
# Build for current platform
go build -o mcp-sdk-migrate ./cmd/mcp-sdk-migrate

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o mcp-sdk-migrate-linux ./cmd/mcp-sdk-migrate
GOOS=darwin GOARCH=arm64 go build -o mcp-sdk-migrate-mac ./cmd/mcp-sdk-migrate
GOOS=windows GOARCH=amd64 go build -o mcp-sdk-migrate.exe ./cmd/mcp-sdk-migrate
```

### As Go Install

```bash
# Install directly from source
go install github.com/glassBead-tc/widescreen-research/cmd/mcp-sdk-migrate@latest

# Use anywhere
mcp-sdk-migrate -file server.go
```

## Related Documentation

- [Migration Specification](../../docs/mcp-server-migration-spec.md) - Complete migration guide
- [Official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) - Official SDK docs
- [Migration Examples](../../docs/mcp-server-migration-spec.md#detailed-migration-examples) - Before/after examples

## Contributing

This utility is maintained as part of the widescreen-research project. Contributions welcome!

## License

MIT License - See repository LICENSE file
