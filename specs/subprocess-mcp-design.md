# Subprocess MCP Design Pattern (Go host ↔ TypeScript server)

## Intent
Run a TypeScript MCP server (EXA Websets) as a child process of a Go application, speaking Model Context Protocol over stdio JSON‑RPC. This bridges language boundaries without bindings and adheres to MCP’s standard tool invocation.

### Architecture overview (high level)

```mermaid
graph TD
  A[Go Host Process<br/>MCP Client (SDK)] -- stdio JSON-RPC --> B[TypeScript MCP Server<br/>EXA Websets Tools]
  B --> C[External Services<br/>EXA API, Web, etc.]
  A --> D[Widescreen Orchestrator]
  D --> E[Drone Fleet (HTTP)]
  E --> F[Pub/Sub<br/>Async Results]
  D --> G[Progress + Reports]
```

## Why this pattern
- Keeps the TS code (websets logic, polling, pagination) intact
- Avoids re-implementing APIs in Go
- Uses MCP’s first‑class stdio transport and tools/call semantics
- Swappable server: any MCP server can be dropped in under the same client

## Roles (per MCP spec)
- Host/Client (Go): initiates connection, lists/calls tools
- Server (TS): exposes `websets_manager` and related tools via MCP SDK
- Transport: stdio (stdin/stdout) using JSON‑RPC 2.0 message frames

## Message flow
1) initialize: client → server (capabilities)
2) tools/list (optional sanity)
3) tools/call name=`websets_manager` with arguments
4) result: server → client `result.content[]` (usually first item is text)

### Sequence diagram (protocol)

```mermaid
sequenceDiagram
  participant Go as Go Host (MCP Client)
  participant TS as TS MCP Server (Websets)

  Go->>TS: initialize (features/capabilities)
  TS-->>Go: initialized
  Go->>TS: tools/list (optional)
  TS-->>Go: tools (websets_manager, ...)
  Go->>TS: tools/call websets_manager { args }
  TS-->>Go: result.content[] (text, json)
```

## Process model
- Go creates a long‑lived child: `exa-websets-mcp-server` (bin) or `node build/index.js`
- Connection established via SDK `NewCommandTransport(exec.Command(...))`
- One session reused across calls; restart on crash
- Graceful shutdown ties to Go process lifecycle

### Lifecycle flow

```mermaid
flowchart TD
  S[Start Host] --> P[Spawn TS MCP Server Child]
  P --> C[Connect via SDK CommandTransport]
  C -->|Success| Sess[Create/Reuse MCP Session]
  C -->|Failure| R1[Backoff & Retry]
  R1 --> C
  Sess --> Call[Call Tools as Needed]
  Call --> Sess
  Sess --> H[Shutdown Signal]
  H --> G[Graceful Close Session]
  G --> K[Kill Child if Running]
  K --> End[End]
```

## Concurrency
- Start serialized calls per server instance
- If needed, scale by creating multiple child processes (pool) and shard calls
- MCP JSON‑RPC supports concurrent IDs, but pool-based scaling simplifies backpressure

### Concurrency options

```mermaid
flowchart LR
  Q[Work Queue] -->|Dispatch| PM[Pool Manager]
  PM --> S1[Server #1]
  PM --> S2[Server #2]
  PM --> S3[Server #N]
  S1 --> R1[Result]
  S2 --> R2[Result]
  S3 --> RN[Result]
  R1 & R2 & RN --> Collate[Collate/Reduce]
```

## Timeouts and retries
- Apply per‑call context deadlines
- Use polling patterns for long operations (e.g., get_webset_status)
- Backoff and retry transient failures; one process restart attempt

### Error handling and retries

```mermaid
sequenceDiagram
  participant Go as Go Client
  participant TS as TS Server
  Go->>TS: tools/call (ctx with deadline)
  Note over Go: Start timer
  TS-->>Go: (no response yet)
  Go-->>Go: timeout reached
  Go-->>Go: backoff (exponential)
  Go->>TS: retry tools/call
  alt transient failure persists
    Go-->>Go: restart child process
    Go->>TS: reconnect session
    Go->>TS: retry tools/call
  end
```

## Error surfaces
- JSON‑RPC error → transport/protocol failure
- Tool-level isError: true → business error returned in result.content

## Observability
- Log child start/stop, calls (tool name, duration), and errors
- Optional metrics: per‑operation counters/latency, restart counts

### Observability signals

```mermaid
graph LR
  A[Host Logger] --> L1[Child Start/Stop]
  A --> L2[Tool Calls (name, duration)]
  A --> L3[Errors]
  M[Metrics] --> C1[Per-op Counters]
  M --> C2[Latency Histograms]
  M --> C3[Restart Counts]
```

## Security
- Pass EXA_API_KEY only through child env; never log
- Constrain child binary path and arguments

## Failure modes and handling
- Child exit: mark client unhealthy; next call attempts reconnect
- Stuck child: CallTool context deadline cancels; consider kill/replace
- Tool busy/limits: backoff; reduce concurrency; respect EXA rate limits

## Alternatives considered
- Native Go client to EXA Websets API: single language but re‑implements TS server logic
- HTTP facade on TS server: adds surface area; MCP already solves the contract
- In‑process JS engine: complex and non‑standard for production

## Fit within Widescreen Research
- The orchestrator already owns async orchestration and Pub/Sub; subprocess MCP feeds that pipeline with Websets items
- Minimal edits: implement WebsetsClient, add RunWebsetsPipeline, register operations

## Rollout
1) Add client and wire into orchestrator
2) Integrate new operation(s) in server
3) Local smoke test with a real EXA_API_KEY
4) Stage rollout; monitor restart rates, call success, item throughput


