# Notebook-as-MCP-Agent Specification

**Version:** 1.0
**Date:** September 30, 2025
**Status:** Implementation Ready
**Current Status:** ~110 LOC (parser + executor + runtime)
**Target:** ~600 LOC (full bidirectional MCP)
**Remaining:** ~350 LOC

## Executive Summary

This specification describes a system where computational notebooks (`.src.md` files) become **autonomous MCP agents**. Each notebook can:

1. **Expose itself as an MCP server** - Other agents can call its cells as tools
2. **Connect to external MCP servers** - Access arxiv, exa, firecrawl, etc.
3. **Communicate peer-to-peer** - Call other notebook agents
4. **Maintain state** - Track evidence, findings, confidence using state-tracker pattern
5. **Embed resources in responses** - Tools return data + metadata in one call

## Core Principle

> **Notebooks are state trackers + executors, NOT intelligent agents.**
>
> The LLM orchestrating the agents provides all intelligence. Notebooks just:
>
> - Execute code cells
> - Track state (evidence, findings, confidence)
> - Connect agents together
> - Surface results as MCP tools

## Architecture

### High-Level Flow

```
┌─────────────────────────────────────────────────────────────┐
│  Claude Code (Orchestrator LLM)                             │
│  "Agent 1, search arxiv for X"                              │
│  "Agent 2, analyze what Agent 1 found"                      │
└──────────────┬──────────────────────────────┬───────────────┘
               │                              │
               │ MCP Protocol                 │ MCP Protocol
               │                              │
      ┌────────▼────────┐            ┌────────▼────────┐
      │  Notebook Agent 1│            │  Notebook Agent 2│
      │  (MCP Server)    │────────────│  (MCP Server)    │
      │                  │   P2P      │                  │
      │  Tools:          │            │  Tools:          │
      │  - execute_cell  │            │  - execute_cell  │
      │  - read_state    │            │  - read_state    │
      │                  │            │                  │
      │  Connects to:    │            │  Connects to:    │
      │  - arxiv-mcp     │            │  - exa-mcp       │
      │  - firecrawl-mcp │            │  - agent-1       │
      └──────────────────┘            └──────────────────┘
```

### Component Architecture

```
Notebook Agent Runtime
├── Parser (✅ Complete)
│   └── Parses .src.md → cells
│
├── Executor (✅ Complete)
│   └── Runs TypeScript cells
│
├── MCP Server (🔴 TODO: ~130 LOC)
│   ├── Exposes tools: execute_cell, read_state
│   └── Embeds resources in responses
│
├── MCP Client (🔴 TODO: ~80 LOC)
│   ├── Connects to external servers (arxiv, exa, etc.)
│   └── Connects to peer agents (HTTP transport)
│
├── Helper Injection (🔴 TODO: ~50 LOC)
│   ├── Auto-inject: state(), retrieve(), callPeer()
│   └── Available in all cells
│
└── HTTP Transport (🔴 TODO: ~40 LOC)
    └── For peer-to-peer agent communication
```

## Key Design Decisions

### 1. Resources Embedded in Tool Responses

**Decision:** Tools return embedded resources, not separate resource endpoints.

**Why:**

- Simpler for callers (one call gets everything)
- Cleaner abstraction (hide implementation details)
- Protocol supports it (CallToolResult.content is an array)
- ~50 LOC saved (no listResources/readResource handlers)

**Example:**

```typescript
// Tool call
execute_cell({
  cellId: 'literature-search',
  params: { query: 'quantum computing' }
})

// Response with embedded resources
{
  content: [
    {
      type: "text",
      text: "Cell executed successfully. Found 15 papers."
    },
    {
      type: "resource",
      resource: {
        uri: "notebook:///literature-search/output",
        mimeType: "application/json",
        text: JSON.stringify({
          papers: [...],
          totalFound: 15,
          searchTime: '2.3s'
        })
      }
    },
    {
      type: "resource",
      resource: {
        uri: "notebook:///state",
        mimeType: "application/json",
        text: JSON.stringify({
          evidenceCount: 15,
          confidence: 0.82,
          lastUpdated: '2025-09-30T01:30:00Z'
        })
      }
    }
  ]
}
```

### 2. State Tracker Pattern

**Decision:** Notebooks track state, don't make decisions.

**State Tracking:**

- ✅ Count evidence collected
- ✅ Track confidence scores (LLM-provided)
- ✅ Store findings
- ✅ Maintain execution history

**NOT State Tracking:**

- ❌ Evaluate quality of evidence
- ❌ Decide what to search next
- ❌ Generate hypotheses
- ❌ Make strategic decisions

**Example:**

```typescript
// In cell: literature-search
const papers = await retrieve('arxiv', { query: 'quantum computing' });

// ✅ GOOD: Store what was found
state.track('evidence', {
  source: 'arxiv',
  count: papers.length,
  confidence: 0.8  // LLM provided this
});

// ❌ BAD: Don't analyze or decide
// if (papers.length < 10) {
//   // Server shouldn't decide this needs more
//   state.needsMore = true;
// }
```

### 3. Helper Functions Auto-Injected

**Decision:** Every cell has access to `state()`, `retrieve()`, `callPeer()` without imports.

**Implementation:** Runtime injects these at execution time.

```typescript
// What the user writes in a cell:
const papers = await retrieve('arxiv', { query: 'ML' });
state.track('papers', papers.length);

// What actually executes (helpers injected):
(async function(__runtime) {
  const state = __runtime.state;
  const retrieve = __runtime.retrieve;
  const callPeer = __runtime.callPeer;

  // User's code here
  const papers = await retrieve('arxiv', { query: 'ML' });
  state.track('papers', papers.length);
})(__injectedRuntime);
```

## Implementation Phases

### Phase 1: MCP Server (~130 LOC)

**Goal:** Expose notebook as MCP server with tools.

**Tools to Implement:**

```typescript
interface NotebookTools {
  // Execute a cell and return results + embedded resources
  execute_cell(args: {
    cellId: string;
    params?: Record<string, unknown>;
  }): CallToolResult;

  // Read current state
  read_state(): CallToolResult;

  // List available cells
  list_cells(): CallToolResult;
}
```

**Files:**

```
packages/research-agent-mcp/
├── src/
│   └── server/
│       ├── mcp-server.ts          # MCP server implementation (~80 LOC)
│       ├── tool-handlers.ts       # Tool handler functions (~50 LOC)
│       └── types.ts                # TypeScript types
```

**Key Code:**

```typescript
// mcp-server.ts
import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';

export class NotebookMCPServer {
  private server: Server;
  private notebook: NotebookRuntime;

  constructor(notebookPath: string) {
    this.notebook = new NotebookRuntime(notebookPath);
    this.server = new Server({
      name: `notebook-${path.basename(notebookPath)}`,
      version: '1.0.0'
    }, {
      capabilities: {
        tools: {}
      }
    });

    this.registerHandlers();
  }

  private registerHandlers() {
    // List available tools
    this.server.setRequestHandler(ListToolsRequestSchema, async () => ({
      tools: [
        {
          name: 'execute_cell',
          description: 'Execute a notebook cell and return results',
          inputSchema: {
            type: 'object',
            properties: {
              cellId: { type: 'string' },
              params: { type: 'object' }
            },
            required: ['cellId']
          }
        },
        {
          name: 'read_state',
          description: 'Read the current state tracker data',
          inputSchema: { type: 'object', properties: {} }
        }
      ]
    }));

    // Handle tool calls
    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      switch (request.params.name) {
        case 'execute_cell':
          return this.handleExecuteCell(request.params.arguments);
        case 'read_state':
          return this.handleReadState();
        default:
          throw new Error(`Unknown tool: ${request.params.name}`);
      }
    });
  }

  private async handleExecuteCell(args: any): Promise<CallToolResult> {
    const { cellId, params } = args;

    // Execute the cell
    const result = await this.notebook.executeCell(cellId, params);

    // Return with embedded resources
    return {
      content: [
        {
          type: 'text',
          text: `Cell ${cellId} executed successfully`
        },
        {
          type: 'resource',
          resource: {
            uri: `notebook:///${cellId}/output`,
            mimeType: 'application/json',
            text: JSON.stringify(result.output)
          }
        },
        {
          type: 'resource',
          resource: {
            uri: 'notebook:///state',
            mimeType: 'application/json',
            text: JSON.stringify(this.notebook.getState())
          }
        }
      ]
    };
  }

  async start() {
    const transport = new StdioServerTransport();
    await this.server.connect(transport);
  }
}
```

### Phase 2: MCP Client (~80 LOC)

**Goal:** Connect to external MCP servers and peer agents.

**Capabilities:**

1. Connect to stdio servers (arxiv, exa, firecrawl)
2. Connect to HTTP servers (peer agents)
3. Provide `retrieve()` helper for cells

**Files:**

```
packages/research-agent-mcp/
└── src/
    └── client/
        ├── mcp-client.ts          # Client connection manager (~50 LOC)
        ├── http-transport.ts      # HTTP transport for peers (~30 LOC)
        └── types.ts
```

**Key Code:**

```typescript
// mcp-client.ts
import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';

export class NotebookMCPClient {
  private clients: Map<string, Client> = new Map();

  async connectStdio(name: string, command: string, args: string[]) {
    const transport = new StdioClientTransport({ command, args });
    const client = new Client({
      name: `notebook-client-${name}`,
      version: '1.0.0'
    }, {
      capabilities: {}
    });

    await client.connect(transport);
    this.clients.set(name, client);
  }

  async connectHTTP(name: string, url: string) {
    const transport = new HTTPClientTransport(url);
    const client = new Client({
      name: `notebook-client-${name}`,
      version: '1.0.0'
    }, {
      capabilities: {}
    });

    await client.connect(transport);
    this.clients.set(name, client);
  }

  async retrieve(serverName: string, toolName: string, args: any) {
    const client = this.clients.get(serverName);
    if (!client) {
      throw new Error(`Not connected to server: ${serverName}`);
    }

    const result = await client.callTool({ name: toolName, arguments: args });
    return result;
  }
}
```

### Phase 3: Helper Injection (~50 LOC)

**Goal:** Auto-inject helpers into every cell execution.

**Helpers:**

- `state` - State tracker operations
- `retrieve()` - Call external MCP tools
- `callPeer()` - Call peer agent tools

**Files:**

```
packages/research-agent-mcp/
└── src/
    └── runtime/
        ├── helpers.ts              # Helper implementations (~30 LOC)
        └── injector.ts             # Code injection (~20 LOC)
```

**Key Code:**

```typescript
// injector.ts
export function injectHelpers(cellCode: string, runtime: NotebookRuntime): string {
  return `
(async function(__runtime) {
  // State tracker
  const state = {
    track: (key, value) => __runtime.state.set(key, value),
    get: (key) => __runtime.state.get(key),
    all: () => __runtime.state.getAll()
  };

  // Retrieve from external servers
  const retrieve = async (server, params) => {
    return await __runtime.client.retrieve(server, params);
  };

  // Call peer agents
  const callPeer = async (agentName, cellId, params) => {
    return await __runtime.client.retrieve(
      agentName,
      'execute_cell',
      { cellId, params }
    );
  };

  // User's cell code
  ${cellCode}
})(__injectedRuntime);
  `.trim();
}
```

### Phase 4: Runtime Integration (~50 LOC)

**Goal:** Wire everything together.

**Files:**

```
packages/research-agent-mcp/
└── src/
    ├── runtime.ts                 # Main runtime (~50 LOC)
    └── index.ts                   # CLI entry point
```

**Key Code:**

```typescript
// runtime.ts (enhanced)
export class NotebookRuntime {
  private cells: Map<string, Cell>;
  private state: StateTracker;
  private mcpServer: NotebookMCPServer;
  private mcpClient: NotebookMCPClient;

  constructor(notebookPath: string) {
    this.cells = this.parseNotebook(notebookPath);
    this.state = new StateTracker();
    this.mcpServer = new NotebookMCPServer(this);
    this.mcpClient = new NotebookMCPClient();
  }

  async executeCell(cellId: string, params?: any): Promise<ExecutionResult> {
    const cell = this.cells.get(cellId);
    if (!cell) throw new Error(`Cell not found: ${cellId}`);

    // Inject helpers
    const injectedCode = injectHelpers(cell.code, this);

    // Execute with injected runtime
    const result = await this.execute(injectedCode, {
      __injectedRuntime: {
        state: this.state,
        client: this.mcpClient,
        params
      }
    });

    return result;
  }

  async connectToServer(name: string, config: ServerConfig) {
    if (config.transport === 'stdio') {
      await this.mcpClient.connectStdio(name, config.command, config.args);
    } else if (config.transport === 'http') {
      await this.mcpClient.connectHTTP(name, config.url);
    }
  }

  async start() {
    // Start MCP server
    await this.mcpServer.start();

    // Connect to configured external servers
    const config = this.loadConfig();
    for (const [name, serverConfig] of Object.entries(config.servers)) {
      await this.connectToServer(name, serverConfig);
    }
  }
}
```

### Phase 5: HTTP Transport (~40 LOC)

**Goal:** Enable peer-to-peer agent communication over HTTP.

**Files:**

```
packages/research-agent-mcp/
└── src/
    └── transport/
        ├── http-server.ts         # HTTP server transport (~20 LOC)
        └── http-client.ts         # HTTP client transport (~20 LOC)
```

**Key Code:**

```typescript
// http-server.ts
import express from 'express';

export class HTTPServerTransport {
  private app = express();

  constructor(private server: NotebookMCPServer, private port: number) {
    this.app.use(express.json());

    // MCP over HTTP endpoint
    this.app.post('/mcp', async (req, res) => {
      try {
        const response = await this.server.handleRequest(req.body);
        res.json(response);
      } catch (error) {
        res.status(500).json({ error: error.message });
      }
    });
  }

  async start() {
    return new Promise((resolve) => {
      this.app.listen(this.port, () => {
        console.log(`Agent listening on http://localhost:${this.port}`);
        resolve();
      });
    });
  }
}

// http-client.ts
export class HTTPClientTransport {
  constructor(private url: string) {}

  async send(request: any): Promise<any> {
    const response = await fetch(`${this.url}/mcp`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request)
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return response.json();
  }
}
```

## Configuration

### Notebook Configuration File

Each notebook can have a `.notebookrc.json`:

```json
{
  "name": "research-agent-1",
  "version": "1.0.0",
  "mcp": {
    "server": {
      "enabled": true,
      "transport": "stdio",
      "httpPort": 3001
    },
    "client": {
      "servers": {
        "arxiv": {
          "transport": "stdio",
          "command": "npx",
          "args": ["-y", "@agentic/arxiv-mcp"]
        },
        "exa": {
          "transport": "stdio",
          "command": "npx",
          "args": ["-y", "@agentic/exa-mcp"]
        },
        "agent-2": {
          "transport": "http",
          "url": "http://localhost:3002"
        }
      }
    }
  },
  "state": {
    "tracker": "default",
    "persistence": "memory"
  }
}
```

## Usage Examples

### Example 1: Simple Research Agent

**Notebook: `research-agent.src.md`**

````markdown
# Research Agent

## Cell: literature-search

```typescript
const papers = await retrieve('arxiv', {
  query: params.query,
  maxResults: 20
});

state.track('papers_found', papers.length);
state.track('search_query', params.query);

return { papers };
```

## Cell: analyze-findings

```typescript
const papers = state.get('papers_found');

// Ask peer agent to analyze
const analysis = await callPeer('analyst-agent', 'deep-analysis', {
  papers: papers
});

state.track('analysis_complete', true);
state.track('confidence', analysis.confidence);

return { analysis };
```
````

**Usage from Claude Code:**

```typescript
// Start agent
$ notebook-agent research-agent.src.md --port 3001

// In Claude Code conversation:
User: "Search arxiv for quantum computing papers"

Claude: *calls research-agent.execute_cell*
{
  cellId: 'literature-search',
  params: { query: 'quantum computing' }
}

// Response with embedded resources:
{
  content: [
    { type: 'text', text: 'Found 20 papers' },
    {
      type: 'resource',
      resource: {
        uri: 'notebook:///literature-search/output',
        text: '{"papers": [...]}'
      }
    },
    {
      type: 'resource',
      resource: {
        uri: 'notebook:///state',
        text: '{"papers_found": 20, "search_query": "quantum computing"}'
      }
    }
  ]
}
```

### Example 2: Multi-Agent Collaboration

**Agent 1: Searcher**

```typescript
// Cell: search
const papers = await retrieve('arxiv', { query: params.query });
return { papers };
```

**Agent 2: Analyzer**

```typescript
// Cell: analyze
const searchResults = await callPeer('agent-1', 'search', {
  query: 'machine learning'
});

const analysis = await retrieve('exa', {
  url: searchResults.papers[0].url
});

return { analysis };
```

**Agent 3: Synthesizer**

```typescript
// Cell: synthesize
const agent1Results = await callPeer('agent-1', 'search', { query: 'ML' });
const agent2Results = await callPeer('agent-2', 'analyze', { papers: agent1Results });

const synthesis = {
  totalPapers: agent1Results.papers.length,
  analysis: agent2Results.analysis,
  confidence: 0.85
};

state.track('synthesis_complete', true);
return { synthesis };
```

## State Tracker Pattern Enforcement

All notebook agents follow the **state tracker pattern**:

### ✅ Notebooks Should

1. **Execute cells** - Run TypeScript code
2. **Track state** - Store evidence, findings, counts
3. **Connect agents** - Facilitate peer-to-peer calls
4. **Return results** - Embed resources in responses
5. **Handle errors** - Validate parameters, catch exceptions

### ❌ Notebooks Should NEVER

1. **Generate content** - Don't create hypotheses, ideas, or strategies
2. **Analyze semantics** - Don't evaluate quality or correctness
3. **Make decisions** - Don't choose what to do next based on content
4. **Provide intelligence** - Don't infer, deduce, or reason
5. **Transform content** - Don't modify or enhance data beyond basic formatting

### Validation

The notebook runtime includes validation:

```typescript
// In runtime.ts
function validateCellCode(code: string): void {
  const forbidden = [
    /generateHypothesis/i,
    /analyzeQuality/i,
    /decideBestStrategy/i,
    /inferConclusion/i,
  ];

  for (const pattern of forbidden) {
    if (pattern.test(code)) {
      throw new Error(
        `Cell contains forbidden pattern: ${pattern}. ` +
        'Notebooks are state trackers, not intelligence providers.'
      );
    }
  }
}
```

## File Structure

```
packages/research-agent-mcp/
├── src/
│   ├── index.ts                   # CLI entry point
│   ├── runtime.ts                 # Main runtime (enhanced)
│   ├── parser.ts                  # ✅ Complete
│   ├── executor.ts                # ✅ Complete
│   │
│   ├── server/
│   │   ├── mcp-server.ts          # MCP server (~80 LOC)
│   │   ├── tool-handlers.ts       # Tool handlers (~50 LOC)
│   │   └── types.ts
│   │
│   ├── client/
│   │   ├── mcp-client.ts          # Client manager (~50 LOC)
│   │   ├── http-transport.ts      # HTTP transport (~30 LOC)
│   │   └── types.ts
│   │
│   ├── runtime/
│   │   ├── helpers.ts             # Helper functions (~30 LOC)
│   │   ├── injector.ts            # Code injection (~20 LOC)
│   │   └── state-tracker.ts
│   │
│   └── transport/
│       ├── http-server.ts         # HTTP server (~20 LOC)
│       └── http-client.ts         # HTTP client (~20 LOC)
│
├── examples/
│   ├── research-agent.src.md
│   ├── analyst-agent.src.md
│   └── synthesizer-agent.src.md
│
├── tests/
│   ├── mcp-server.test.ts
│   ├── mcp-client.test.ts
│   └── integration.test.ts
│
├── package.json
├── tsconfig.json
└── README.md
```

## Testing Strategy

### Unit Tests

```typescript
describe('NotebookMCPServer', () => {
  test('exposes execute_cell tool', async () => {
    const server = new NotebookMCPServer('example.src.md');
    const tools = await server.listTools();
    expect(tools.find(t => t.name === 'execute_cell')).toBeDefined();
  });

  test('embeds resources in tool responses', async () => {
    const server = new NotebookMCPServer('example.src.md');
    const result = await server.callTool('execute_cell', { cellId: 'test' });

    expect(result.content).toContainEqual({
      type: 'resource',
      resource: expect.objectContaining({
        uri: expect.stringMatching(/^notebook:\/\//)
      })
    });
  });
});
```

### Integration Tests

```typescript
describe('Multi-Agent Communication', () => {
  test('agent can call peer agent', async () => {
    const agent1 = new NotebookRuntime('agent1.src.md');
    const agent2 = new NotebookRuntime('agent2.src.md');

    await agent1.start();
    await agent2.start();
    await agent1.connectToServer('agent-2', {
      transport: 'http',
      url: 'http://localhost:3002'
    });

    const result = await agent1.executeCell('call-peer', {
      agent: 'agent-2',
      cell: 'process'
    });

    expect(result.success).toBe(true);
  });
});
```

## Performance Considerations

### Latency Budget

```
User: "Agent 1, search papers"
  │
  ├─ Claude → Agent 1 MCP: ~50ms
  │
  ├─ Agent 1 execute cell: ~100ms
  │   ├─ Inject helpers: ~5ms
  │   ├─ Execute TypeScript: ~80ms
  │   └─ Collect results: ~15ms
  │
  ├─ Agent 1 → arxiv MCP: ~200ms
  │
  └─ Response to Claude: ~50ms

Total: ~400ms ✅
```

### Resource Management

- **Memory**: Each agent ~50MB baseline
- **Connections**: Max 10 external servers per agent
- **Concurrency**: Process cells sequentially within agent
- **State**: In-memory by default, optional persistence

## Migration Path

### Phase 1: Foundation (Week 1)

- [ ] Implement MCP server
- [ ] Basic tool handlers
- [ ] Test with Claude Code

### Phase 2: Client Connectivity (Week 1)

- [ ] Implement MCP client
- [ ] Connect to stdio servers
- [ ] Test with arxiv-mcp

### Phase 3: Peer-to-Peer (Week 2)

- [ ] HTTP transport
- [ ] Multi-agent tests
- [ ] Example notebooks

### Phase 4: Polish (Week 2)

- [ ] Helper injection
- [ ] Documentation
- [ ] State tracker validation

## Success Criteria

- [ ] Agent exposes itself as MCP server
- [ ] Agent can connect to external MCP servers (arxiv, exa, etc.)
- [ ] Agent can call peer agents over HTTP
- [ ] Cells have access to `state()`, `retrieve()`, `callPeer()`
- [ ] Tool responses embed resources
- [ ] Example multi-agent notebook works
- [ ] Tests pass
- [ ] ~600 LOC total

## References

- [MCP Protocol Specification](https://modelcontextprotocol.io/)
- [Claude Code MCP Docs](https://docs.claude.com/en/docs/claude-code/mcp)
- [Claude Code Sub-agents](https://docs.claude.com/en/docs/claude-code/sub-agents)
- [State Tracker Pattern](./state-tracker-pattern.md)
- [One-Step-Per-Call Implementation](../clearthought-onepointfive/docs/one-step-per-call-implementation.md)

---

**Status:** Ready for implementation
**Next Step:** Phase 1 - Implement MCP Server
**Estimated Completion:** 2 weeks
