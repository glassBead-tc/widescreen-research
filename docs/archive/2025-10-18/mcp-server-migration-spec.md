# MCP Server Migration Specification: mark3labs → Official SDK

## Executive Summary

Migrate all MCP server implementations from the unofficial `github.com/mark3labs/mcp-go` SDK to the official `github.com/modelcontextprotocol/go-sdk`. This aligns with the project's existing MCP client implementation and ensures protocol compliance, long-term support, and consistency.

## Current State Analysis

### Files Using mark3labs SDK

1. **cmd/widescreen-research-mcp/server/server.go** - Main widescreen research MCP server
2. **pkg/mcp/server.go** - Coordinator MCP server wrapper
3. **pkg/mcp/server_handlers_campaign.go** - Campaign-specific MCP handlers
4. **cmd/simple-mcp/main.go** - Simple example MCP server

### Already Using Official SDK

- **cmd/widescreen-research-mcp/orchestrator/mcp_client.go** - MCP client (✓ correct)
- **cmd/widescreen-research-mcp/orchestrator/mcp_client_test.go** - Client tests (✓ correct)

## Why Migrate?

### Technical Reasons

1. **Official Support**: `modelcontextprotocol/go-sdk` is the canonical implementation maintained by Anthropic
2. **Protocol Compliance**: Guaranteed compliance with MCP spec updates
3. **Consistency**: Client already uses official SDK - server should match
4. **Advanced Features**: Official SDK supports all MCP primitives (tools, resources, prompts, **elicitation, sampling, roots**)
5. **Type Safety**: Better type definitions and error handling
6. **Future-Proof**: Active development and community support

### Alignment with Project Goals

From `cmd/widescreen-research-mcp/orchestrator/mcp_client_spec.md`:

- Spec explicitly states: "Use the official `github.com/modelcontextprotocol/go-sdk`"
- Spec explicitly states: "NOT the unofficial `github.com/mark3labs/mcp-go`"

## Architecture Differences

### mark3labs SDK (Current)

```go
// Server creation
server := server.NewMCPServer("name", "1.0.0",
    server.WithToolCapabilities(true),
    server.WithRecovery())

// Tool definition
tool := mcp.NewTool("tool_name",
    mcp.WithDescription("desc"),
    mcp.WithString("param", mcp.Required()))

// Tool registration
server.AddTool(tool, handlerFunc)

// Start server
server.ServeStdio(server)
```

### Official SDK (Target)

```go
// Server creation
server := mcp.NewServer(&mcp.ServerOptions{
    Capabilities: mcp.ServerCapabilities{
        Tools: &mcp.ToolsCapability{},
        Resources: &mcp.ResourcesCapability{},
        Prompts: &mcp.PromptsCapability{},
    },
})

// Tool registration via handler
server.SetToolHandler(func(ctx context.Context, req *mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
    return &mcp.ListToolsResult{
        Tools: []mcp.Tool{
            {
                Name: "tool_name",
                Description: "desc",
                InputSchema: jsonschema.Object{...},
            },
        },
    }, nil
})

server.SetCallToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Handle tool calls
})

// Start server with stdio transport
transport := mcp.NewStdioServerTransport()
if err := server.Connect(transport); err != nil {
    log.Fatal(err)
}
```

## Key Pattern Differences

| Aspect | mark3labs | Official SDK |
|--------|-----------|--------------|
| **Tool Definition** | Declarative with builders | Imperative via handlers |
| **Tool Registration** | `AddTool(tool, handler)` | `SetCallToolHandler(router)` |
| **Tool Discovery** | Automatic from registered tools | Manual in `ListToolsHandler` |
| **Parameter Extraction** | `request.RequireString("param")` | Parse from `request.Params` |
| **Server Start** | `server.ServeStdio(server)` | `server.Connect(transport)` |
| **Error Handling** | Return `*mcp.CallToolResult` with errors | Return standard Go errors |
| **Transport** | Built into server | Pluggable (stdio, SSE, WebSocket) |

## Migration Plan

### Phase 1: Specification and Analysis ✓

1. Document current implementation patterns
2. Map mark3labs patterns to official SDK equivalents
3. Identify breaking changes and compatibility issues
4. Create migration specification (this document)

### Phase 2: Update Dependencies

```bash
# Update go.mod
go get github.com/modelcontextprotocol/go-sdk@latest
go get github.com/invopop/jsonschema@latest  # For input schemas

# Optional: Remove mark3labs dependency after migration
# go mod tidy
```

### Phase 3: Create Official SDK Server Implementation

**Priority Order:**

1. **cmd/widescreen-research-mcp/server/** (HIGH) - Main production server
2. **pkg/mcp/server.go** (MEDIUM) - Coordinator MCP wrapper
3. **cmd/simple-mcp/main.go** (LOW) - Example server

### Phase 4: Implementation Strategy

#### For Each Server File

1. **Create new implementation alongside old** (non-breaking)
   - `server.go` → Keep temporarily
   - `server_official.go` → New implementation

2. **Implement tool handlers**
   - Create tool list
   - Create JSON schemas for tool inputs
   - Implement call handler with routing logic

3. **Implement resource handlers** (if applicable)
   - List resources
   - Read resource

4. **Implement prompt handlers** (if applicable)
   - List prompts
   - Get prompt

5. **Update main.go entry point**
   - Switch to official SDK server
   - Configure stdio transport
   - Update error handling

6. **Test with MCP client**
   - Test tool discovery
   - Test tool execution
   - Test error scenarios

7. **Remove old implementation**
   - Delete mark3labs code
   - Clean up imports

## Detailed Migration Examples

### Example 1: Simple Tool Migration

**Before (mark3labs):**

```go
tool := mcp.NewTool("get_status",
    mcp.WithDescription("Get system status"))

server.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return mcp.NewToolResultText("Status: OK"), nil
})
```

**After (Official SDK):**

```go
// In ListToolsHandler
func (s *Server) ListToolsHandler(ctx context.Context, req *mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
    return &mcp.ListToolsResult{
        Tools: []mcp.Tool{
            {
                Name: "get_status",
                Description: "Get system status",
                InputSchema: jsonschema.Object{},  // No parameters
            },
        },
    }, nil
}

// In CallToolHandler
func (s *Server) CallToolHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    switch req.Params.Name {
    case "get_status":
        return &mcp.CallToolResult{
            Content: []mcp.Content{
                mcp.TextContent{
                    Type: "text",
                    Text: "Status: OK",
                },
            },
        }, nil
    default:
        return nil, fmt.Errorf("unknown tool: %s", req.Params.Name)
    }
}
```

### Example 2: Tool with Parameters

**Before (mark3labs):**

```go
tool := mcp.NewTool("spawn_drone",
    mcp.WithDescription("Spawn a drone"),
    mcp.WithString("drone_type", mcp.Required(), mcp.Enum("research", "analysis")),
    mcp.WithString("region", mcp.DefaultString("us-central1")))

server.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    droneType, _ := req.RequireString("drone_type")
    region := req.GetString("region", "us-central1")

    result := fmt.Sprintf("Spawned %s drone in %s", droneType, region)
    return mcp.NewToolResultText(result), nil
})
```

**After (Official SDK):**

```go
// Define schema
type SpawnDroneInput struct {
    DroneType string `json:"drone_type" jsonschema:"required,enum=research|analysis,description=Type of drone to spawn"`
    Region    string `json:"region" jsonschema:"description=GCP region,default=us-central1"`
}

// In ListToolsHandler
{
    Name: "spawn_drone",
    Description: "Spawn a drone",
    InputSchema: jsonschema.Reflect(&SpawnDroneInput{}),
}

// In CallToolHandler
case "spawn_drone":
    var input SpawnDroneInput
    if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
        return nil, fmt.Errorf("invalid arguments: %w", err)
    }

    if input.Region == "" {
        input.Region = "us-central1"
    }

    result := fmt.Sprintf("Spawned %s drone in %s", input.DroneType, input.Region)
    return &mcp.CallToolResult{
        Content: []mcp.Content{
            mcp.TextContent{Type: "text", Text: result},
        },
    }, nil
```

### Example 3: Server Setup and Start

**Before (mark3labs):**

```go
func main() {
    server := server.NewMCPServer("app", "1.0.0",
        server.WithToolCapabilities(true),
        server.WithRecovery())

    // Register tools...

    if err := server.ServeStdio(server); err != nil {
        log.Fatal(err)
    }
}
```

**After (Official SDK):**

```go
func main() {
    srv := NewMyMCPServer()  // Custom server struct

    server := mcp.NewServer(&mcp.ServerOptions{
        Capabilities: mcp.ServerCapabilities{
            Tools: &mcp.ToolsCapability{},
        },
    })

    server.SetToolHandler(srv.ListToolsHandler)
    server.SetCallToolHandler(srv.CallToolHandler)

    transport := mcp.NewStdioServerTransport()
    if err := server.Connect(transport); err != nil {
        log.Fatal(err)
    }

    // Block until transport closes
    <-transport.Done()
}
```

## Testing Strategy

### Unit Tests

```go
func TestToolHandlers(t *testing.T) {
    srv := NewMyMCPServer()

    // Test ListTools
    tools, err := srv.ListToolsHandler(context.Background(), &mcp.ListToolsRequest{})
    require.NoError(t, err)
    assert.NotEmpty(t, tools.Tools)

    // Test CallTool
    result, err := srv.CallToolHandler(context.Background(), &mcp.CallToolRequest{
        Params: mcp.CallToolParams{
            Name: "spawn_drone",
            Arguments: json.RawMessage(`{"drone_type": "research"}`),
        },
    })
    require.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Integration Tests

1. **Test with mcp-agent-ts** - TypeScript client test harness
2. **Test with Claude Code** - Real-world MCP client
3. **Test stdio protocol** - Verify JSON-RPC 2.0 compliance

### Test Checklist

- [ ] Tool discovery (`tools/list`)
- [ ] Tool execution (`tools/call`)
- [ ] Error handling and validation
- [ ] Parameter parsing and defaults
- [ ] Multiple tool execution
- [ ] Resource listing (if implemented)
- [ ] Resource reading (if implemented)
- [ ] Prompt listing (if implemented)
- [ ] Prompt retrieval (if implemented)
- [ ] Server initialization
- [ ] Graceful shutdown
- [ ] Environment variable handling
- [ ] Concurrent request handling

## File-Specific Migration Plans

### 1. cmd/widescreen-research-mcp/server/server.go

**Current Tools:**

- `widescreen-research` - Main research orchestration tool

**Migration Steps:**

1. Create input schema structs for each tool
2. Implement `ListToolsHandler` returning all tool definitions
3. Implement `CallToolHandler` with routing to operation handlers
4. Update `Start()` method to use official SDK server + stdio transport
5. Preserve existing operation handlers (orchestration, elicitation, etc.)
6. Update `main.go` to use new server pattern

**Estimated Complexity:** HIGH (main production server with complex operations)

### 2. pkg/mcp/server.go

**Current Tools:**

- `spawn_drone_server`
- `list_active_drones`
- `execute_distributed_task`
- `get_drone_status`

**Migration Steps:**

1. Create schema structs for drone operations
2. Implement tool and call handlers
3. Update coordinator integration
4. Update campaign handlers in `server_handlers_campaign.go`

**Estimated Complexity:** MEDIUM (multiple tools but straightforward logic)

### 3. cmd/simple-mcp/main.go

**Current Tools:**

- `spawn_drone_server`
- `list_active_drones`
- `execute_distributed_task`
- `get_system_status`

**Migration Steps:**

1. Simplified example - good migration template
2. Migrate as reference implementation for others
3. Use as integration test baseline

**Estimated Complexity:** LOW (example code, can be rewritten cleanly)

## Rollout Strategy

### Development Phase

1. **Branch**: Create `feat/migrate-official-mcp-sdk` branch
2. **Implement**: Migrate one file at a time in priority order
3. **Test**: Comprehensive testing after each migration
4. **Review**: Code review focusing on protocol compliance

### Testing Phase

1. Run existing Go tests
2. Test with mcp-agent-ts test harness (to be created)
3. Test with Claude Code manually
4. Integration test with orchestrator MCP client

### Deployment Phase

1. Merge to main after all tests pass
2. Update documentation (CLAUDE.md, README.md, OPERATIONS.md)
3. Update MCP config examples
4. Tag release: `v1.0.0-official-sdk`

## Dependencies Update

### Add to go.mod

```go
require (
    github.com/modelcontextprotocol/go-sdk v0.8.0  // Already present
    github.com/invopop/jsonschema v0.13.0          // For input schemas
)
```

### Remove after migration complete

```go
// Can be removed once all servers migrated
github.com/mark3labs/mcp-go v0.41.0
```

## Documentation Updates

### Files to Update

1. **CLAUDE.md** - Update MCP implementation details
2. **docs/OPERATIONS.md** - Update server start instructions
3. **cmd/widescreen-research-mcp/README.md** - Update usage examples
4. **cmd/widescreen-research-mcp/orchestrator/mcp_client_spec.md** - Add server migration notes

### New Documentation

1. **docs/mcp-server-patterns.md** - Official SDK patterns and best practices
2. **docs/mcp-testing-guide.md** - Testing MCP servers

## Success Criteria

- [ ] All server files migrated to official SDK
- [ ] All existing functionality preserved
- [ ] All tests passing (unit + integration)
- [ ] Successfully connects to Claude Code
- [ ] Successfully tested with mcp-agent-ts test harness
- [ ] Documentation updated
- [ ] mark3labs dependency removed
- [ ] No protocol compliance issues
- [ ] Performance equivalent or better

## Risk Mitigation

### Potential Issues

1. **Breaking Changes**: Parameter extraction differs significantly
   - **Mitigation**: Comprehensive input validation tests

2. **Transport Setup**: Different initialization pattern
   - **Mitigation**: Test with multiple MCP clients early

3. **Error Handling**: Different error return patterns
   - **Mitigation**: Standardize error responses, test error paths

4. **Schema Definition**: More verbose schema definitions
   - **Mitigation**: Use jsonschema package with struct tags

5. **Backward Compatibility**: Existing clients may depend on server
   - **Mitigation**: No external clients yet; internal only

## Timeline Estimate

- **Specification**: 1 hour ✓ (this document)
- **simple-mcp migration**: 1-2 hours (learning + template)
- **pkg/mcp/* migration**: 2-3 hours (multiple tools)
- **widescreen-research-mcp migration**: 3-4 hours (complex operations)
- **Testing**: 2-3 hours (comprehensive validation)
- **Documentation**: 1 hour
- **Total**: ~10-13 hours

## Next Steps

1. Review and approve this specification
2. Begin with simple-mcp as learning exercise
3. Create test harness for validation
4. Proceed with pkg/mcp migration
5. Final: widescreen-research-mcp migration
6. Comprehensive testing and documentation

## References

- [Official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [MCP Specification](https://modelcontextprotocol.io/docs/specification)
- [Existing Client Spec](cmd/widescreen-research-mcp/orchestrator/mcp_client_spec.md)
- [mark3labs SDK Docs](https://github.com/mark3labs/mcp-go)
