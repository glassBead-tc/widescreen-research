# Implementation Plan: Notebook-as-MCP-Agent System
**Feature**: Minimal bidirectional MCP runtime for research agent notebooks
**Target**: ~600 LOC + templates
**Timeline**: 1-2 days for MVP
**Complexity**: Medium (new paradigm, but small scope)

---

## Problem Statement

**Current**: Multi-agent research requires complex orchestration frameworks
**Goal**: Notebooks ARE agents - each .src.md file becomes a runnable MCP server/client
**Non-Goal**: Full srcbook UI, complex session management, multi-user support

---

## Acceptance Criteria

- [ ] Parse .src.md files into executable cells
- [ ] Execute TypeScript cells in isolation
- [ ] Expose MCP server (execute_cell, read_cell, write_cell tools)
- [ ] Include MCP client (call peers + external services)
- [ ] Auto-inject state-tracker cell into every notebook
- [ ] Auto-inject retrieve() helper (Exa → Firecrawl hierarchy)
- [ ] Auto-inject discoverPeers() for zero-config networking
- [ ] Docker container runs notebook as MCP server
- [ ] Example: 3-agent collaborative research working end-to-end

---

## Implementation Phases (Parallelizable Where Noted)

### Phase 1: Core Notebook Runtime (Day 1, Morning)

**Files to create**:
```
packages/research-agent-mcp/
├── src/
│   ├── index.ts              # Main runtime
│   ├── srcmd-parser.ts       # Parse .src.md → cells
│   ├── executor.ts           # Execute TypeScript cells
│   └── types.ts              # Core interfaces
```

**Tasks** (can parallelize):

**Task 1.1**: Srcmd Parser (~100 LOC)
```typescript
// srcmd-parser.ts
export function parseSrcmd(content: string): Notebook {
  // Regex to find cells: ###### cellname.ts
  // Extract code blocks
  // Return: {cells: [{id, filename, source}]}
}
```

**Task 1.2**: Cell Executor (~100 LOC)
```typescript
// executor.ts
export async function executeCell(cell: Cell, params?: any): Promise<any> {
  // Write cell to temp file
  // Run: npx tsx temp-file.ts
  // Capture stdout/stderr
  // Parse JSON result if present
}
```

**Gate**: Can parse .src.md and execute cells ✅

---

### Phase 2: MCP Server Integration (Day 1, Afternoon)

**Parallel with Phase 1.2**

**Files**:
```
src/
├── mcp-server.ts            # Expose notebook operations
└── tools.ts                 # Tool definitions
```

**Tasks**:

**Task 2.1**: MCP Server (~100 LOC)
```typescript
// mcp-server.ts
import { mcp } from '@modelcontextprotocol/go-sdk';

export function createAgentServer(notebook: Notebook) {
  const server = mcp.NewServer({name: 'agent', version: '1.0.0'}, null);

  // Tool: execute_cell
  mcp.AddTool(server, {name: 'execute_cell'}, async (ctx, req, args) => {
    const result = await executeCell(notebook.cells[args.cellId], args.params);
    return {Content: [{Type: 'text', Text: JSON.stringify(result)}]};
  });

  // Tool: read_cell, write_cell
  // Resource: Expose all cells as resources

  return server;
}
```

**Gate**: MCP server exposes notebook operations ✅

---

### Phase 3: MCP Client Integration (Day 1, Afternoon)

**Parallel with Phase 2**

**Files**:
```
src/
├── mcp-client.ts           # Call peers + external
└── discovery.ts            # Find peer agents
```

**Tasks**:

**Task 3.1**: MCP Client (~100 LOC)
```typescript
// mcp-client.ts
export async function createAgentClient() {
  const client = mcp.NewClient({name: 'client'}, null);

  // Connect to external services
  await client.connect(stdioTransport('arxiv-paper-mcp'));
  await client.connect(stdioTransport('exa'));

  // Connect to peer agents (from env PEER_AGENT_URLS)
  const peers = process.env.PEER_AGENT_URLS.split(',');
  for (const url of peers) {
    await client.connect(httpTransport(url));
  }

  return client;
}
```

**Task 3.2**: Peer Discovery (~50 LOC)
```typescript
// discovery.ts
export async function discoverPeers(): Promise<string[]> {
  // Docker compose: agent-1, agent-2, etc. via DNS
  // Return: ['http://agent-1:9001', 'http://agent-2:9002', ...]
}
```

**Gate**: Agent can call peers and external services ✅

---

### Phase 4: Auto-Injected Helper Cells (Day 1, Evening)

**Parallel with Phase 3.2**

**Files**:
```
templates/
├── cells/
│   ├── state-tracker.ts     # Auto-injected
│   ├── retrieve-helper.ts   # Auto-injected
│   └── peer-discovery.ts    # Auto-injected
```

**Tasks**:

**Task 4.1**: State Tracker Template (~50 LOC)
```typescript
// Injected into every notebook as first cell
export const state = {
  evidenceCount: 0,
  phase: 'observe',
  iteration: 0
};

// Deterministic only
export function recordEvidence(count: number) {
  state.evidenceCount += count;
}

export function checkGate(gate: string): boolean {
  if (gate === 'evidence') return state.evidenceCount >= 20;
  return false;
}
```

**Task 4.2**: Retrieve Helper (~50 LOC)
```typescript
// Exa → Firecrawl hierarchy enforced
export async function retrieve(query: string, url?: string) {
  if (url) {
    return await firecrawl_scrape({url, formats: ['markdown']});
  }

  const discovered = await exa_web_search({query, numResults: 5});
  return await Promise.all(
    discovered.results.map(r => firecrawl_scrape({url: r.url}))
  );
}
```

**Task 4.3**: Peer Discovery Template (~30 LOC)
```typescript
export async function discoverPeers() {
  // Returns list of peer agent URLs
  const urls = process.env.PEER_AGENT_URLS?.split(',') || [];
  return urls.filter(u => u);
}

export async function callPeer(agentId: string, tool: string, args: any) {
  return await client.callTool({name: `${agentId}.${tool}`, arguments: args});
}
```

**Gate**: Templates ready for injection ✅

---

### Phase 5: Docker Deployment (Day 2, Morning)

**Can start in parallel once Phase 1 complete**

**Files**:
```
templates/
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

**Tasks**:

**Task 5.1**: Dockerfile (~20 lines)
```dockerfile
FROM node:20-alpine
WORKDIR /workspace
COPY package.json ./
RUN npm install
COPY src/ ./src/
COPY notebook.src.md ./
CMD ["node", "dist/index.js"]
```

**Task 5.2**: Docker Compose for 5 Agents (~50 lines)
```yaml
services:
  agent-1:
    build: .
    environment:
      - AGENT_ID=agent-1
      - PEER_AGENT_URLS=http://agent-2:9002,http://agent-3:9003
    ports: ["9001:9001"]

  agent-2:
    # ...similar
```

**Gate**: Can deploy 5 agents in containers ✅

---

### Phase 6: Example Research Pattern (Day 2, Afternoon)

**Depends on Phases 1-4 complete**

**Files**:
```
examples/
└── collaborative-research.src.md
```

**Tasks**:

**Task 6.1**: Create Example Notebook
- 3-agent collaborative research on "neural network capacity"
- Agent 1: Literature search (arxiv + exa)
- Agent 2: Architecture extraction (exa + firecrawl + context7)
- Agent 3: Synthesis (reads agent 1 & 2, calculates)

**Task 6.2**: End-to-End Test
- Run docker-compose up
- Verify agents can call each other
- Verify research completes
- Check notebook has final results

**Gate**: Example working end-to-end ✅

---

## Detailed Task Breakdown

### Core Runtime (index.ts) - ~200 LOC

```typescript
// index.ts - Complete agent runtime

import { parseSrcmd } from './srcmd-parser.js';
import { executeCell } from './executor.js';
import { createAgentServer } from './mcp-server.js';
import { createAgentClient } from './mcp-client.js';
import { injectHelperCells } from './helpers.js';

async function main() {
  // 1. Load notebook
  const notebookPath = process.env.NOTEBOOK_PATH || './notebook.src.md';
  const notebook = parseSrcmd(readFileSync(notebookPath, 'utf8'));

  // 2. Inject helper cells (state-tracker, retrieve, peers)
  injectHelperCells(notebook);

  // 3. Create MCP server (expose to peers)
  const server = createAgentServer(notebook);

  // 4. Create MCP client (call peers + external)
  const client = await createAgentClient();

  // 5. Make client available to notebook cells
  global.mcpClient = client;

  // 6. Start MCP server
  await server.Run(context.Background(), mcp.StdioTransport());
}

main();
```

**Parallelizable**: Tasks 2, 3, 4 can be worked on simultaneously

---

## Parallel Execution Strategy

### Day 1 Morning (3 parallel tracks):

**Track A**: Srcmd parser (1 person/agent)
**Track B**: Cell executor (1 person/agent)
**Track C**: Type definitions (1 person/agent)

All merge to: Working notebook runtime

### Day 1 Afternoon (2 parallel tracks):

**Track A**: MCP server implementation
**Track B**: MCP client + discovery

Both use types from Track C morning

### Day 1 Evening (3 parallel tracks):

**Track A**: State tracker template
**Track B**: Retrieve helper template
**Track C**: Docker configuration

All independent, can work in parallel

### Day 2 (Integration):

**Sequential** (dependencies):
1. Integrate all pieces
2. Create example notebook
3. Test end-to-end
4. Document

---

## Risks & Mitigations

**Risk 1**: MCP bidirectional communication complex
- **Mitigation**: Start with stdio only, add HTTP later
- **Fallback**: Use simple HTTP if MCP too complex

**Risk 2**: srcmd parsing edge cases
- **Mitigation**: Start with simple notebooks, add features incrementally
- **Fallback**: Use JSON format if markdown too fragile

**Risk 3**: Agent coordination bugs
- **Mitigation**: Extensive logging, state tracker validation
- **Fallback**: Start with 2 agents, scale to 5

---

## Success Metrics

**Quantitative**:
- [ ] <600 LOC for core runtime
- [ ] <5 dependencies (MCP SDK + minimal utils)
- [ ] <30 seconds cold start for agent container
- [ ] Supports 5 parallel agents

**Qualitative**:
- [ ] Agents can discover each other's tools
- [ ] Agents can call each other via MCP
- [ ] State tracker pattern enforced
- [ ] Exa → Firecrawl hierarchy automatic
- [ ] Example research completes successfully

---

## Next Immediate Steps

1. Create package structure
2. Implement srcmd parser (simplest version)
3. Implement cell executor (tsx wrapper)
4. Test: Can we run a .src.md file headlessly?
5. Add MCP server (expose execute_cell)
6. Add MCP client (call arxiv-paper-mcp)
7. Test: Can agent use external MCP?
8. Add peer discovery
9. Test: Can 2 agents call each other?
10. Scale to 5 agents

Want me to start implementing this minimal version right now?