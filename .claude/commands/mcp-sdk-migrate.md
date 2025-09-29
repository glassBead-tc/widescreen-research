# MCP SDK Migration Command

Migrates Go MCP server code from unofficial mark3labs SDK to official modelcontextprotocol SDK.

## Usage

```
/mcp-sdk-migrate [file_path]
```

## What it does

1. Analyzes the target Go file for mark3labs MCP patterns
2. Converts to official SDK patterns
3. Updates imports
4. Transforms tool registration from declarative to AddTool pattern
5. Updates server initialization and transport setup
6. Preserves all business logic

## Examples

```bash
# Migrate a single file
/mcp-sdk-migrate cmd/my-server/main.go

# Migrate server package
/mcp-sdk-migrate pkg/mcp/server.go
```

## What gets converted

### Imports

- `github.com/mark3labs/mcp-go/mcp` → `github.com/modelcontextprotocol/go-sdk/mcp`
- `github.com/mark3labs/mcp-go/server` → (removed, use mcp package)

### Server Creation

- `server.NewMCPServer("name", "version", opts...)` → `mcp.NewServer(&mcp.Implementation{Name: "name", Version: "version"}, nil)`

### Tool Registration

- `mcp.NewTool(name, mcp.WithDescription(...), mcp.WithString(...))` → Type-safe struct with jsonschema tags
- `server.AddTool(tool, handler)` → `mcp.AddTool(server, &mcp.Tool{...}, handler)`

### Handler Signature

- `func(ctx, mcp.CallToolRequest) (*mcp.CallToolResult, error)` → `func(ctx, *mcp.CallToolRequest, Args) (*mcp.CallToolResult, any, error)`

### Server Start

- `server.ServeStdio(server)` → `server.Run(ctx, &mcp.StdioTransport{})`

### Error Handling

- `mcp.NewToolResultError(msg)` → `&mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: msg}}}`
- `mcp.NewToolResultText(msg)` → `&mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: msg}}}`

## Notes

- Creates backup with `.backup` extension
- Preserves all comments and business logic
- May require manual adjustment for complex parameter extraction
- Always test after migration
