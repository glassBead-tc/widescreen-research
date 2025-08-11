# MCP Client Implementation (Go) for EXA Websets MCP Server

## Overview

Implement a Go MCP client that connects to the existing TypeScript EXA Websets MCP server as a subprocess over stdio and issues MCP `tools/call` requests to the unified `websets_manager` tool. This reuses the official Go MCP SDK in this repo (`go-sdk/`) and avoids re-implementing EXA Websets logic.

## Dependencies

- Local SDK: `github.com/modelcontextprotocol/go-sdk`
  - Use a Go module replace to point to the local `go-sdk/` folder.

```go
// in go.mod (root module)
replace github.com/modelcontextprotocol/go-sdk => ./go-sdk
```

- Node >= 18 and the EXA server binary
  - Path: `third_party/exa-mcp-server-websets` or `exa-mcp-server-websets` (embedded)
  - Entry: bin `exa-websets-mcp-server` or `node ./build/index.js`
  - Env: `EXA_API_KEY` must be set for the child process

## Client responsibilities

- Spawn the TS MCP server as a child process with stdio transport
- Initialize MCP session
- Call `tools/call` for `websets_manager` with typed arguments
- Parse `result.content` (text) payloads; surface `isError`
- Manage lifecycle: reuse single session, restart on crash, close on shutdown
- Provide convenience methods for common ops (create, status, list items)

## Public API (Go)

```go
// Package: cmd/widescreen-research-mcp/orchestrator (or pkg/mcpclient)

type WebsetsClient interface {
    Connect(ctx context.Context) error
    Call(ctx context.Context, arguments map[string]any) (string, error)
    Close() error
}
```

## Reference implementation sketch

```go
package orchestrator

import (
    "context"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "sync"

    sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type StdIOWebsetsClient struct {
    mu       sync.Mutex
    client   *sdk.Client
    session  *sdk.ClientSession
    cmd      *exec.Cmd
    started  bool
    binPath  string   // e.g., "exa-websets-mcp-server" or "node"
    binArgs  []string // e.g., ["./build/index.js"]
    env      []string // must include EXA_API_KEY
}

func NewStdIOWebsetsClient(binPath string, binArgs []string) *StdIOWebsetsClient {
    return &StdIOWebsetsClient{binPath: binPath, binArgs: binArgs}
}

func (c *StdIOWebsetsClient) Connect(ctx context.Context) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.started {
        return nil
    }

    // Ensure EXA_API_KEY present
    if os.Getenv("EXA_API_KEY") == "" {
        return errors.New("EXA_API_KEY not set in environment")
    }

    // Prepare command
    cmd := exec.CommandContext(ctx, c.binPath, c.binArgs...)
    cmd.Env = os.Environ() // inherits EXA_API_KEY

    // Create MCP client and transport
    client := sdk.NewClient(&sdk.Implementation{Name: "widescreen-client", Version: "v1.0.0"}, nil)
    transport := sdk.NewCommandTransport(cmd)

    // Connect (per go-sdk example)
    session, err := client.Connect(ctx, transport)
    if err != nil {
        return fmt.Errorf("mcp connect failed: %w", err)
    }

    c.client = client
    c.session = session
    c.cmd = cmd
    c.started = true
    return nil
}

// Call invokes the unified websets_manager tool.
func (c *StdIOWebsetsClient) Call(ctx context.Context, arguments map[string]any) (string, error) {
    c.mu.Lock()
    session := c.session
    c.mu.Unlock()
    if session == nil {
        if err := c.Connect(ctx); err != nil { return "", err }
        c.mu.Lock(); session = c.session; c.mu.Unlock()
    }

    params := &sdk.CallToolParams{
        Name:      "websets_manager",
        Arguments: arguments,
    }
    res, err := session.CallTool(ctx, params)
    if err != nil { return "", fmt.Errorf("tools/call failed: %w", err) }
    if res.IsError {
        // Extract first text content if present
        if len(res.Content) > 0 {
            if tc, ok := res.Content[0].(*sdk.TextContent); ok {
                return "", fmt.Errorf(tc.Text)
            }
        }
        return "", errors.New("tool call returned isError=true")
    }
    if len(res.Content) == 0 { return "", nil }
    if tc, ok := res.Content[0].(*sdk.TextContent); ok {
        return tc.Text, nil
    }
    return "", nil
}

func (c *StdIOWebsetsClient) Close() error {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.session != nil { _ = c.session.Close() }
    c.client = nil
    c.session = nil
    c.started = false
    return nil
}
```

## Usage patterns

- Create webset:

```go
text, err := wsClient.Call(ctx, map[string]any{
    "operation": "create_webset",
    "webset": map[string]any{
        "searchQuery": topic,
        "advanced": map[string]any{"resultCount": 50},
    },
})
```

- Poll status (loop with timeout):

```go
text, err := wsClient.Call(ctx, map[string]any{
    "operation": "get_webset_status",
    "resourceId": websetID,
})
```

- List items (paginate via the tool’s defaults):

```go
text, err := wsClient.Call(ctx, map[string]any{
    "operation": "list_content_items",
    "resourceId": websetID,
    "query": map[string]any{"limit": 100},
})
```

Parse the returned `text` JSON per the EXA tool’s responses (success, items array, pagination flags).

## Concurrency and timeouts

- Start with serialized calls per process instance (one request at a time). The SDK supports concurrent sessions if needed later.
- Wrap each `Call` with context deadlines. Long-running actions should be handled by polling, not single long calls.

## Restart policy

- If the child process exits or `CallTool` returns a transport error, mark the client as not started and attempt a one-time reconnect on next call. Surface errors after a second failure.

## Logging and metrics

- Log: start, connect, tool name, duration, errors. Avoid logging `EXA_API_KEY`.
- Optional counters: calls by operation, failures, average latency.

## Wiring into orchestrator

- Instantiate `StdIOWebsetsClient` in `orchestrator.NewOrchestrator()` and reuse across the process.
- Use it inside a new `RunWebsetsPipeline` (create → poll → list → publish items to Pub/Sub).
- Expose `websets-orchestrate` and optional `websets-call` operations in `cmd/widescreen-research-mcp/server/server.go`.

## Configuration

- Child binary: default `exa-websets-mcp-server`; fallback `node ./build/index.js` if needed.
- Env: `EXA_API_KEY` required.
- Poll interval: default 10s; timeout: default 15m. Make overrides via env or parameters.


