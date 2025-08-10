# Refactor Spec: Adopt Official MCP Go SDK for Widescreen Research

Reference: [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk)

## Objective
Refactor MCP-facing components to use the official Model Context Protocol Go SDK (the “SDK”) instead of the current third‑party `mcp-go` library and custom glue types, while preserving user-visible behavior:
- Orchestrator-driven research execution (topic breakdown, drone assignment)
- Drone compatibility (persistent server via HTTP, async results via Pub/Sub)
- Progress tracking and report generation (JSON artifacts + final markdown)
- Existing orchestrator tests remain green; add an end-to-end SDK client test

## Target End State
- All MCP servers/clients in this repo use the official SDK APIs.
- Tools defined with typed parameter structs using JSON Schema annotations supplied via the SDK.
- Handlers use the SDK’s typed call semantics and return SDK content payloads.
- No remaining dependencies on `github.com/mark3labs/mcp-go`.
- Domain logic remains in `orchestrator`, `operations`, and `schemas` and is decoupled from protocol details.

## Scope
- Replace MCP server in `cmd/widescreen-research-mcp/server` with the SDK server.
- Replace any MCP client usage with the SDK client.
- Keep domain types and logic intact; use SDK only at the protocol boundary.

## SDK Architectural Model (v0.2.0)
See: [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk)

- Server: `mcp.NewServer(&mcp.Implementation{Name, Version}, features)`
- Run: `server.Run(ctx, mcp.NewStdioTransport())`
- Register Tool: `mcp.AddTool(server, &mcp.Tool{Name, Description}, Handler)`
- Handler signature:
  - `func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ParamsT]) (*mcp.CallToolResultFor[any], error)`
- Client: `mcp.NewClient(...)`, `session := client.Connect(...)`, `session.CallTool(ctx, &mcp.CallToolParams{Name, Arguments})`

Note: The SDK is currently marked as unstable and subject to breaking changes. Pin to a specific tag and isolate SDK usage via small adapters.

## Design Changes

### 1) MCP Server (widescreen research)
- Replace current server wiring with SDK server:
  - `server := mcp.NewServer(&mcp.Implementation{Name: "widescreen-research", Version: "1.0.0"}, nil)`
  - Register a primary tool `widescreen_research` handling elicitation and operation dispatch via typed params.
- Tool params example (typed with jsonschema tags):
  ```go
  type WidescreenResearchParams struct {
    Operation          string         `json:"operation" jsonschema:"operation to execute"`
    SessionID          string         `json:"session_id" jsonschema:"elicitation session id"`
    Parameters         map[string]any `json:"parameters,omitempty" jsonschema:"operation params"`
    ElicitationAnswers map[string]any `json:"elicitation_answers,omitempty" jsonschema:"answers"`
  }
  ```
- Handler converts params to `schemas.WidescreenResearchInput` and routes:
  - `start` or empty → elicitation manager
  - `orchestrate-research`, `sequential-thinking`, `gcp-provision`, `analyze-findings` → corresponding domain ops
- Responses as SDK `CallToolResultFor[any]` content:
  - Text summaries (progress)
  - JSON content (`mcp.JSONContent`) for structured outputs (reports, metrics)

### 2) Tools coverage
Implement/preserve these tools (typed params + typed JSON result):
- `widescreen_research` (primary entrypoint)
- `sequential_thinking`
- `gcp_provision`
- `analyze_findings`

### 3) Prompts and Resources
- Model previous prompt/resource behaviors as explicit tools for now.
- If the SDK exposes native prompt/resource APIs later, add adapters without touching domain code.

### 4) Orchestrator and Drone Integration
- Orchestrator remains the coordinator for topic breakdown and drone assignment.
- Drone servers continue over HTTP and Pub/Sub for results; no protocol change required.
- If MCP client calls are needed, use `mcp.NewClient` + `CallTool`.

### 5) Domain Types and Boundaries
- Keep domain types in `cmd/widescreen-research-mcp/schemas` and `operations`.
- SDK types appear only at the server/client boundary; avoid leaking them into domain packages.

## Migration Plan
1) Dependencies
- Add `github.com/modelcontextprotocol/go-sdk` pinned to a specific tag (e.g., v0.2.0 or commit SHA).
- Keep `mark3labs/mcp-go` during transition; remove after parity.

2) Parallel server implementation
- Add SDK-based server behind a build tag or alternate package path.
- Smoke test with SDK example client over stdio.

3) Tool-by-tool migration
- Port `widescreen_research` first; validate typed params/JSON schema.
- Port `sequential_thinking`, `gcp_provision`, `analyze_findings` next.

4) Remove legacy MCP lib
- Remove `github.com/mark3labs/mcp-go` from `go.mod` once parity is verified.
- Delete obsolete adapters.

5) Test & docs
- Ensure orchestrator unit tests pass.
- Add an SDK client e2e test (start server on stdio, call `widescreen_research`, verify elicitation + execution + JSON content shape).
- Update README(s) and this spec.

## Acceptance Criteria
- `go build ./...` green; `go test ./...` green incl. SDK e2e.
- Tools functional: `widescreen_research`, `sequential_thinking`, `gcp_provision`, `analyze_findings`.
- Elicitation produces `ResearchConfig`; orchestrator consumes it.
- Progress tracking + report generation unchanged from user POV.
- `github.com/mark3labs/mcp-go` removed.

## Risks & Mitigations
- SDK instability: pin exact version, isolate via wrappers.
- Feature parity gaps: model prompts/resources as tools; revisit when SDK supports.
- Integration drift: keep domain boundaries; strengthen e2e tests.

## Type Mapping (Illustrative)
- Server
  - From: third‑party `server.NewMCPServer(...)`
  - To: SDK `mcp.NewServer(&mcp.Implementation{Name, Version}, nil)`
- Transport
  - From: `ServeStdio(s)`
  - To: `server.Run(ctx, mcp.NewStdioTransport())`
- Tool definition
  - From: `mcp.NewTool(...); s.AddTool(tool, handler)` (third‑party)
  - To: `mcp.AddTool(server, &mcp.Tool{Name, Description}, handler)` (SDK)
- Handler signature
  - From: `(ctx, request mcp.CallToolRequest) (*mcp.CallToolResult, error)`
  - To: `(ctx, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ParamsT]) (*mcp.CallToolResultFor[any], error)`
- Params/Schema
  - Introduce `ParamsT` with `jsonschema` tags (descriptions, enums, defaults).

## Testing Strategy
- Unit: param unmarshalling, domain path validation per tool.
- Integration: SDK client invokes `widescreen_research` and verifies
  - Elicitation loop (start → follow‑ups → ready)
  - Orchestration output structure (task breakdown, assignments)
  - Final JSON content contains links to raw artifacts and summary markdown path.

## Deliverables
- Updated server using SDK and pinned dep in `go.mod`.
- New/updated tests incl. SDK e2e.
- Removal of legacy MCP lib and dead code.
- Updated documentation (README, this spec).
