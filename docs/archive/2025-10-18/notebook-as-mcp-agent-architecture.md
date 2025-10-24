# Notebook-as-MCP-Agent Architecture

**Based on**: Srcbook headless runner + MCP-GCP-sweeper
**Purpose**: Multi-agent research system with bidirectional MCP communication
**Key Insight**: Notebooks are both execution environments AND communication protocols

---

## Analysis of Existing Srcbook System

### Current Architecture (from @srcbook monorepo)

**Packages**:

1. **`@srcbook/api`** - Core notebook management (sessions, cells, execution)
2. **`@srcbook/web`** - React frontend for editing notebooks
3. **`@srcbook/components`** - UI components
4. **`@srcbook/runner`** - Headless execution for Cloud Run
5. **`@srcbook/mcp-gcp-sweeper`** - MCP server exposing sweep management
6. **`@srcbook/shared`** - Common types and utilities

### What Srcbook Already Provides

✅ **Headless execution**: `@srcbook/runner` runs notebooks without UI
✅ **MCP integration**: `mcp-gcp-sweeper` exposes tools
✅ **Cell execution**: Can run TypeScript/JavaScript cells
✅ **Session management**: Track notebook state
✅ **Dependency management**: npm install for notebook deps
✅ **GCP integration**: Cloud Run + GCS for distributed execution
✅ **Parameter sweeps**: Grid search over parameter spaces

---

## What to Keep vs. Remove for Multi-Agent MCP System

### ✅ KEEP (Essential for Headless Multi-Agent)

#### From `@srcbook/api`

- **`session.mts`** - Session management, cell CRUD
- **`exec.mts`** - Process spawning (node, tsx, npm)
- **`executor.mts`** - Cell execution logic
- **`srcbook/index.mts`** - Core srcbook operations (import/export)
- **`srcmd.mts`** - Markdown parsing (.src.md ↔ cells)
- **`storage/`** - File system operations
- **Database layer** (if we need persistence across agent restarts)

**Why**: These provide headless notebook execution without UI dependency

#### From `@srcbook/runner`

- **Entire package** - Already headless!
- Parameter sweep grid logic
- GCS artifact writing
- Cloud Run integration

**Why**: Perfect for distributed agent execution

#### From `@srcbook/mcp-gcp-sweeper`

- **MCP server pattern** - How to expose notebook ops as tools
- **Tool definitions** - submit, status, logs, cancel, fetch_artifact

**Why**: Template for building bidirectional MCP servers

#### From `@srcbook/shared`

- **Type definitions** - Cell types, language types, schemas
- **Utility functions** - randomid, validation
- **Constants** - File paths, configurations

**Why**: Needed by all packages

---

### ❌ REMOVE (UI-Only, Not Needed for Headless Agents)

#### From `@srcbook/web`

- **Entire package** - React frontend
- Components for editing, rendering cells
- Websocket real-time updates

**Why**: Agents don't need visual interface

#### From `@srcbook/components`

- **Entire package** - UI components

**Why**: No rendering needed for agents

#### From `@srcbook/api`

- **`server/http.mts`** - HTTP server routes (if UI-only)
- **`ai/` directory** - AI code generation features (maybe - see below)
- **UI-focused database queries** - Listing for display purposes

**Why**: Agents interact via MCP, not HTTP UI

---

### ⚠️ MAYBE KEEP (Depends on Use Case)

#### AI Code Generation (`api/ai/`)

- **Keep if**: We want agents to generate code cells via LLM
- **Remove if**: Only LLM client (Claude Code) generates code

**My recommendation**: **KEEP** - useful for agents to modify their own notebooks

#### Database Layer

- **Keep if**: Need persistence across agent container restarts
- **Remove if**: Notebooks are ephemeral per research session

**My recommendation**: **KEEP** - enables "research memory" across runs

#### WebSocket Support

- **Keep if**: Agents stream results to each other
- **Remove if**: MCP request/response is sufficient

**My recommendation**: **REMOVE** - MCP handles communication

---

## Proposed Headless Multi-Agent Architecture

### Package Structure

```
@research-agent-mcp/
├── core/                    # From @srcbook/api (stripped)
│   ├── session.ts          # ✅ Keep
│   ├── exec.ts             # ✅ Keep
│   ├── executor.ts         # ✅ Keep
│   ├── srcbook/            # ✅ Keep
│   │   ├── index.ts        # Import/export notebooks
│   │   ├── config.ts       # Package.json generation
│   │   └── path.ts         # File path utilities
│   ├── srcmd.ts            # ✅ Keep (.src.md parser)
│   └── storage/            # ✅ Keep (file operations)
│
├── runner/                  # From @srcbook/runner
│   ├── index.ts            # ✅ Keep (headless execution)
│   └── Dockerfile          # ✅ Keep (containerization)
│
├── mcp-server/              # NEW - Bidirectional MCP
│   ├── server.ts           # MCP server exposing notebook tools
│   ├── client.ts           # MCP client for calling other agents
│   └── tools/
│       ├── notebook-ops.ts # execute_cell, read_cell, write_cell
│       ├── agent-coord.ts  # discover_agents, call_agent_tool
│       └── state-track.ts  # record_state, check_gate, get_status
│
├── shared/                  # From @srcbook/shared
│   ├── types.ts            # ✅ Keep (cell types, schemas)
│   └── utils.ts            # ✅ Keep (validation, randomid)
│
└── examples/
    └── research-patterns/   # Research workflow notebooks
        ├── autonomous-discovery.src.md
        ├── collaborative-compounding.src.md
        └── swarm-intelligence.src.md
```

**Total removal**: ~60% of codebase (all UI packages)
**Kept**: Core execution engine + MCP integration

---

## New MCP Server Design (Bidirectional)

### Each Agent Container Runs

```typescript
// agent-mcp-server.ts

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { createSession, executeCell, addCell } from '@research-agent-mcp/core';

// ======================
// MCP SERVER (Expose capabilities to others)
// ======================

const server = new Server({
  name: `research-agent-${process.env.AGENT_ID}`,
  version: '1.0.0'
}, {
  capabilities: {
    tools: {},
    resources: {}
  }
});

// Tool: Execute notebook cell
mcp.AddTool(server, {
  name: 'execute_cell',
  description: 'Execute a specific cell in the agent notebook'
}, async (ctx, req, args: {cellId: string, params: any}) => {
  const session = await createSession(process.env.NOTEBOOK_PATH);
  const result = await executeCell(session, args.cellId, args.params);

  return {
    Content: [{
      Type: 'text',
      Text: JSON.stringify(result)
    }]
  };
});

// Tool: Write cell (for other agents to contribute)
mcp.AddTool(server, {
  name: 'write_cell',
  description: 'Write data to a cell in shared notebook'
}, async (ctx, req, args: {cellId: string, content: string}) => {
  const session = await createSession(process.env.NOTEBOOK_PATH);
  await writeCell(session, args.cellId, args.content);

  return {
    Content: [{Type: 'text', Text: 'Cell written'}]
  };
});

// Tool: Query local knowledge graph
mcp.AddTool(server, {
  name: 'query_local_graph',
  description: 'Query this agent\'s knowledge graph'
}, async (ctx, req, args: {concept: string}) => {
  // Execute the knowledge-graph.ts cell in notebook
  const session = await createSession(process.env.NOTEBOOK_PATH);
  const result = await executeCell(session, 'knowledge-graph', {
    operation: 'query',
    concept: args.concept
  });

  return {
    Content: [{Type: 'text', Text: JSON.stringify(result)}]
  };
});

// Resource: Expose notebook cells
server.setResourceHandler(async (request) => {
  const session = await createSession(process.env.NOTEBOOK_PATH);
  const cell = session.cells.find(c => c.id === request.uri.split('/').pop());

  return {
    contents: [{
      uri: request.uri,
      mimeType: 'text/plain',
      text: cell?.source || ''
    }]
  };
});

// ======================
// MCP CLIENT (Call other agents + external services)
// ======================

const client = new Client({
  name: `research-agent-client-${process.env.AGENT_ID}`,
  version: '1.0.0'
}, {});

// Discover other agents via service discovery
const otherAgents = await discoverPeerAgents(); // e.g., from env vars or k8s service discovery

// Connect to each peer agent
for (const agent of otherAgents) {
  const transport = createTransport(agent.url); // HTTP, stdio, etc.
  await client.connect(transport);

  // Now this agent can call other agents' tools!
  // e.g., await client.callTool({name: 'agent-2.execute_cell', arguments: {...}})
}

// Connect to external MCP servers
await connectToExternalServers([
  'arxiv-paper-mcp',
  'exa',
  'firecrawl'
]);

// Start MCP server
await server.connect(new StdioServerTransport());
```

---

## What Gets Removed vs. Kept (Summary Table)

| Component | Keep? | Reason |
|-----------|-------|--------|
| **Core notebook execution** | ✅ YES | Essential - runs .src.md files |
| **Session management** | ✅ YES | Tracks notebook state |
| **Cell CRUD operations** | ✅ YES | Agents need to read/write cells |
| **Process spawning (exec)** | ✅ YES | Runs TypeScript cells |
| **Dependency management** | ✅ YES | npm install for agent deps |
| **Srcmd parser** | ✅ YES | Parse .src.md format |
| **Storage layer** | ✅ YES | File system + GCS |
| **Parameter sweep logic** | ✅ YES | Useful for agent orchestration |
| **MCP server pattern** | ✅ YES | Template for agent MCP servers |
| **Cloud Run integration** | ✅ YES | Deploy agents to GCP |
| | | |
| **Web UI (React)** | ❌ NO | Agents don't need visual interface |
| **UI components** | ❌ NO | No rendering |
| **HTTP routes for UI** | ❌ NO | MCP replaces HTTP for agents |
| **WebSocket updates** | ❌ NO | MCP handles real-time |
| **User authentication** | ❌ NO | Agent-to-agent has different auth |
| **Database UI queries** | ⚠️ MAYBE | Only if persistence needed |
| **AI code generation** | ⚠️ MAYBE | Useful but optional |

**Net reduction**: ~60% smaller codebase
**Kept**: Pure execution engine + MCP integration

---

## My Ideal Design (Maximum Usability for Agents)

### Simplification Principles

1. **One notebook = One agent's workspace**
   - Agent owns a `.src.md` file
   - All agent's state in notebook cells
   - Agent exposes notebook operations via MCP server

2. **Bidirectional MCP = Agent communication**
   - MCP Server: Expose what this agent can do
   - MCP Client: Call other agents + external services
   - No custom protocols needed

3. **Deterministic state tracking in code cells**
   - State tracker cell: `state.ts`
   - All state mutations go through state tracker
   - Hooks validate state transitions

4. **LLM intelligence in markdown cells**
   - Research questions documented in markdown
   - Hypotheses in markdown
   - Agent execution in TypeScript cells

5. **Container = Agent runtime**
   - One Docker container per agent
   - Isolated filesystem
   - MCP server on stdio/HTTP
   - Can call peer containers via MCP

---

## Proposed Minimal Package: `@research-agent-notebook`

### Core Module (200-300 LOC)

```typescript
// index.ts - Complete headless notebook-as-MCP-agent runtime

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { exec } from 'child_process';
import { promises as fs } from 'fs';

// ============================================
// SRCBOOK EXECUTION (from @srcbook/api)
// ============================================

interface Cell {
  id: string;
  type: 'markdown' | 'code' | 'package.json';
  source: string;
  filename?: string; // for code cells
}

class NotebookSession {
  constructor(
    public id: string,
    public dir: string,
    public cells: Cell[]
  ) {}

  async executeCell(cellId: string, params?: any): Promise<any> {
    const cell = this.cells.find(c => c.id === cellId);
    if (!cell || cell.type !== 'code') {
      throw new Error(`Code cell ${cellId} not found`);
    }

    // Write cell to filesystem
    const cellPath = `${this.dir}/src/${cell.filename}`;
    await fs.writeFile(cellPath, cell.source);

    // Execute with tsx
    return new Promise((resolve, reject) => {
      exec(
        `npx tsx ${cellPath}`,
        { cwd: this.dir, env: {...process.env, ...params} },
        (error, stdout, stderr) => {
          if (error) reject(error);
          resolve({ stdout, stderr });
        }
      );
    });
  }

  async writeCell(cellId: string, content: string): Promise<void> {
    const cell = this.cells.find(c => c.id === cellId);
    if (cell) {
      cell.source = content;
      if (cell.type === 'code' && cell.filename) {
        await fs.writeFile(`${this.dir}/src/${cell.filename}`, content);
      }
    }
  }

  async readCell(cellId: string): Promise<string> {
    const cell = this.cells.find(c => c.id === cellId);
    return cell?.source || '';
  }
}

// ============================================
// MCP SERVER (Expose notebook operations)
// ============================================

const agentId = process.env.AGENT_ID || 'agent-unknown';
const notebookPath = process.env.NOTEBOOK_PATH || './notebook.src.md';

const server = new Server({
  name: `research-agent-${agentId}`,
  version: '1.0.0'
}, {
  capabilities: { tools: {}, resources: {} }
});

let session: NotebookSession;

// Initialize session
async function initSession() {
  // Parse .src.md and create session
  session = await createSessionFromSrcmd(notebookPath);
}

// Tool: Execute cell
mcp.AddTool(server, {
  name: 'execute_cell',
  description: 'Execute a notebook cell with optional parameters'
}, async (ctx, req, args: {cellId: string, params?: any}) => {
  const result = await session.executeCell(args.cellId, args.params);
  return {
    Content: [{Type: 'text', Text: JSON.stringify(result)}]
  };
});

// Tool: Read cell
mcp.AddTool(server, {
  name: 'read_cell',
  description: 'Read cell content (for other agents to access data)'
}, async (ctx, req, args: {cellId: string}) => {
  const content = await session.readCell(args.cellId);
  return {
    Content: [{Type: 'text', Text: content}]
  };
});

// Tool: Write cell
mcp.AddTool(server, {
  name: 'write_cell',
  description: 'Write content to a cell (for collaboration)'
}, async (ctx, req, args: {cellId: string, content: string}) => {
  await session.writeCell(args.cellId, args.content);
  return {
    Content: [{Type: 'text', Text: 'Cell updated'}]
  };
});

// ============================================
// MCP CLIENT (Call other agents + external)
// ============================================

const client = new Client({
  name: `research-client-${agentId}`,
  version: '1.0.0'
}, {});

// Connect to peer agents (from env var: PEER_AGENT_URLS)
const peerUrls = (process.env.PEER_AGENT_URLS || '').split(',');
for (const url of peerUrls) {
  if (url) {
    await client.connect(createHttpTransport(url));
  }
}

// Connect to external MCP servers
await client.connect(createStdioTransport({
  command: 'npx',
  args: ['-y', '@smithery/cli', 'run', 'arxiv-paper-mcp']
}));

// Export for use in notebook cells
export { client };

// ============================================
// START SERVERS
// ============================================

await initSession();
await server.connect(new StdioServerTransport());
```

**Total LOC**: ~200-300 (vs. thousands in full srcbook)
**Dependencies**: MCP SDK + minimal utilities
**Functionality**: 90% of what we need for multi-agent research

---

## What This Enables

### Agent Container Architecture

```
Docker Container: research-agent-1
  ├── Notebook: autonomous-discovery.src.md
  │   ├── Cell: state-tracker.ts (deterministic state)
  │   ├── Cell: literature-search.ts (uses MCP client → arxiv)
  │   ├── Cell: knowledge-graph.ts (stores findings)
  │   └── Cell: synthesis.ts (LLM-generated insights)
  │
  ├── MCP Server (exposes):
  │   - execute_cell(cellId, params)
  │   - read_cell(cellId) → returns data
  │   - write_cell(cellId, content)
  │   - query_knowledge_graph(concept)
  │
  └── MCP Client (calls):
      - arxiv-paper-mcp (external)
      - exa (external)
      - agent-2.execute_cell() (peer)
      - agent-3.query_knowledge_graph() (peer)
```

**Each agent**:

- ✅ Has own isolated notebook workspace
- ✅ Exposes capabilities via MCP server
- ✅ Can call peers + external services via MCP client
- ✅ Containerized for isolation + deployment

---

## Communication Patterns

### Pattern 1: Direct Peer Call

```typescript
// In Agent 1's notebook cell: literature-search.ts

import { client } from '../mcp-client.js';

// Agent 1 finds paper mentioning "Llama 3.3"
const papers = await client.callTool({
  name: 'arxiv_search', // External MCP
  arguments: {keyword: 'Llama architecture'}
});

// Agent 1 calls Agent 2 to analyze architecture
const architecture = await client.callTool({
  name: 'agent-2.extract_architecture', // Peer MCP
  arguments: {model: 'Llama 3.3', source: papers[0].url}
});

// Write to own notebook
await writeCell('findings-iter0', JSON.stringify({papers, architecture}));
```

### Pattern 2: Shared Notebook Collaboration

```typescript
// Agent 1 writes to shared notebook cell

await client.callTool({
  name: 'shared-notebook.write_cell',
  arguments: {
    cellId: 'shared-findings',
    content: JSON.stringify(myFindings)
  }
});

// Agent 2 reads from shared notebook

const sharedData = await client.callTool({
  name: 'shared-notebook.read_cell',
  arguments: {cellId: 'shared-findings'}
});

// Agent 2 builds on Agent 1's work
const synthesis = analyzeFin dings(JSON.parse(sharedData));
```

### Pattern 3: Knowledge Graph Federation

```typescript
// Agent 1's knowledge-graph.ts cell

export async function queryGraph(concept: string) {
  // Query local graph built from literature
  const localResults = graph.query(concept);

  // Ask peer agents if they have related concepts
  const agent2Results = await client.callTool({
    name: 'agent-2.query_local_graph',
    arguments: {concept}
  });

  const agent3Results = await client.callTool({
    name: 'agent-3.query_local_graph',
    arguments: {concept}
  });

  // Merge results from federated knowledge graphs
  return {
    local: localResults,
    agent2: JSON.parse(agent2Results),
    agent3: JSON.parse(agent3Results)
  };
}
```

---

## What I'd Design Differently (For Maximum Usability)

### Improvement 1: Convention Over Configuration

**Current srcbook**: Requires setup of directories, configs, etc.

**My version**:

```bash
# Single command creates everything
research-agent-create autonomous-discovery

# Generates:
# - autonomous-discovery.src.md (from template)
# - Dockerfile (pre-configured)
# - docker-compose.yml (with peer agent networking)
# - .env.example (with required vars)
```

**Benefit**: Get started in seconds, not minutes

---

### Improvement 2: Built-in State Tracker Pattern

**Current srcbook**: No state tracking conventions

**My version**: Every notebook template includes `state-tracker.ts` cell:

```typescript
// state-tracker.ts (injected into every research notebook)

export interface AgentState {
  id: string;
  role: string;
  phase: string;
  evidenceCount: number;
  status: 'idle' | 'working' | 'complete' | 'error';
  lastUpdate: string;
}

// DETERMINISTIC state mutations only
export function recordEvidence(count: number): void {
  state.evidenceCount += count;
  state.lastUpdate = new Date().toISOString();
}

export function checkGate(gateName: string): boolean {
  // Return boolean based on state (no intelligence)
  switch(gateName) {
    case 'evidence-gathering':
      return state.evidenceCount >= 20;
    case 'synthesis':
      return state.evidenceCount >= 20 && allAgentsComplete();
    default:
      return false;
  }
}

// NO intelligence, NO decisions, ONLY state tracking
```

**Benefit**: State tracker pattern enforced by default

---

### Improvement 3: MCP Client Registry in Notebook

**Problem**: Hard to track which MCP servers are available

**My solution**: Auto-generated `mcp-registry.ts` cell:

```typescript
// mcp-registry.ts (auto-updated as agents connect)

export const mcpServers = {
  // External services
  external: {
    arxiv: { url: 'stdio://arxiv-paper-mcp', tools: ['search_papers', 'get_paper_info'] },
    exa: { url: 'stdio://exa', tools: ['web_search_exa', 'get_code_context_exa'] },
    firecrawl: { url: 'stdio://firecrawl', tools: ['firecrawl_scrape', 'firecrawl_crawl'] }
  },

  // Peer agents
  peers: {
    'agent-1': { url: 'http://agent-1:9001', tools: ['execute_cell', 'read_cell', 'query_graph'] },
    'agent-2': { url: 'http://agent-2:9002', tools: ['extract_architecture', 'analyze_code'] },
    'agent-3': { url: 'http://agent-3:9003', tools: ['synthesize_findings', 'identify_patterns'] }
  }
};

// Helper function
export async function callAgent(agentId: string, toolName: string, args: any) {
  return await client.callTool({
    name: `${agentId}.${toolName}`,
    arguments: args
  });
}
```

**Benefit**: Agents can discover and call each other easily

---

### Improvement 4: Exa/Firecrawl Hierarchy Helper

**Problem**: Easy to misuse Exa vs. Firecrawl

**My solution**: Helper cell that enforces hierarchy:

```typescript
// retrieval-helpers.ts (in every notebook)

export async function smartRetrieve(query: string, knownUrl?: string) {
  if (knownUrl) {
    // Have URL → use Firecrawl
    console.log('Using Firecrawl (known URL)');
    return await client.callTool({
      name: 'firecrawl_scrape',
      arguments: { url: knownUrl, formats: ['markdown'] }
    });
  } else {
    // No URL → discovery with Exa first
    console.log('Step 1: Discovery with Exa');
    const results = await client.callTool({
      name: 'exa_web_search',
      arguments: { query, numResults: 5 }
    });

    // Step 2: Extract with Firecrawl
    console.log('Step 2: Extraction with Firecrawl');
    const contents = [];
    for (const result of results.results) {
      const content = await client.callTool({
        name: 'firecrawl_scrape',
        arguments: { url: result.url }
      });
      contents.push(content);
    }

    return contents;
  }
}

// Usage in agent cells:
// const papers = await smartRetrieve("subliminal learning");
```

**Benefit**: Impossible to misuse tools, optimal pattern enforced

---

### Improvement 5: Validation Hooks Built-In

**Current**: No validation of agent behavior

**My version**: Hook system integrated with notebook execution:

```typescript
// hooks.ts (runs before/after cell execution)

export async function beforeCellExecution(cellId: string, params: any) {
  // Check state gates
  const state = await readCell('state-tracker');
  const currentPhase = JSON.parse(state).phase;

  if (cellId.includes('synthesis') && currentPhase !== 'synthesis') {
    throw new Error(`Phase violation: Cannot run synthesis in ${currentPhase} phase`);
  }

  // Check tool usage
  if (cellId.includes('firecrawl') && !params.url) {
    throw new Error(`Tool hierarchy violation: firecrawl requires URL. Use exa first for discovery.`);
  }
}

export async function afterCellExecution(cellId: string, result: any) {
  // Update state tracker automatically
  if (cellId.includes('search')) {
    await executeCell('state-tracker', {
      operation: 'recordEvidence',
      count: result.count || 1
    });
  }

  // Check gates after execution
  await executeCell('gate-checker', {});
}
```

**Benefit**: State tracker pattern enforced at runtime, not just convention

---

## Deployment Architecture

### Docker Compose for 5-Agent System

```yaml
version: '3.8'

services:
  # Shared notebook server (optional - for shared workspace pattern)
  shared-notebook:
    build: ./packages/research-agent-notebook
    volumes:
      - ./notebooks/shared-research.src.md:/workspace/notebook.src.md
    environment:
      - AGENT_ID=shared-notebook
      - MODE=server-only
    ports:
      - "9000:9000"

  # Agent 1: Literature Scout
  agent-1:
    build: ./packages/research-agent-notebook
    volumes:
      - ./notebooks/literature-scout.src.md:/workspace/notebook.src.md
    environment:
      - AGENT_ID=agent-1
      - AGENT_ROLE=literature-scout
      - NOTEBOOK_PATH=/workspace/notebook.src.md
      - PEER_AGENT_URLS=http://agent-2:9002,http://agent-3:9003,http://agent-4:9004,http://agent-5:9005
      - SHARED_NOTEBOOK_URL=http://shared-notebook:9000
    ports:
      - "9001:9001"

  # Agents 2-5: Similar structure
  agent-2:
    # ... technical analyst
  agent-3:
    # ... synthesis specialist
  agent-4:
    # ... hypothesis generator
  agent-5:
    # ... validation specialist
```

**Start all**:

```bash
docker-compose up -d

# All 5 agents running, each with:
# - Own notebook workspace
# - MCP server exposing tools
# - MCP client calling peers + external
# - Isolated execution environment
```

---

## Summary: What to Keep

### Absolute Essentials (Can't Remove)

1. ✅ **Notebook parsing** (`srcmd.ts`) - Parse .src.md files
2. ✅ **Cell execution** (`executor.ts`, `exec.ts`) - Run TypeScript cells
3. ✅ **Session management** (`session.ts`) - Track notebook state
4. ✅ **File operations** (`storage/`) - Read/write cells to disk

### Highly Valuable (Should Keep)

5. ✅ **MCP server pattern** (`mcp-gcp-sweeper`) - Template for agent servers
6. ✅ **Headless runner** (`srcbook-runner`) - Cloud deployment
7. ✅ **Shared types** (`@srcbook/shared`) - Type safety
8. ✅ **Parameter sweep** (`grid.ts`) - Useful for hypothesis testing

### Optional (Nice to Have)

9. ⚠️ **AI code generation** - Agents could self-modify
10. ⚠️ **Database** - Persistence across sessions
11. ⚠️ **GCP integration** - If deploying to Cloud Run

### Definitely Remove

12. ❌ **Web UI** (`@srcbook/web`) - No visual interface needed
13. ❌ **UI Components** (`@srcbook/components`) - No rendering
14. ❌ **HTTP Server** (UI routes) - MCP replaces HTTP

---

## Estimated Complexity

**Minimal viable implementation**:

- ~300 LOC for core execution
- ~200 LOC for MCP server/client
- ~100 LOC for utilities
- **Total: ~600 LOC**

**Full-featured implementation**:

- Keep all core API packages: ~3,000 LOC
- Add bidirectional MCP: ~500 LOC
- Remove UI packages: -10,000 LOC
- **Net: ~3,500 LOC** (vs. 13,000+ current)

**My recommendation**: Start with **minimal 600 LOC version**, add features as needed.

---

## Next Steps

Want me to:

1. **Build the minimal version** (~600 LOC) as `@research-agent-notebook`?
2. **Extract from existing srcbook** (keep core, remove UI, add bidirectional MCP)?
3. **Create example notebooks** for each research pattern with this system?
4. **Write the Docker/deployment config** for 5-agent setup?
