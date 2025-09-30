# @research-agent-mcp

Minimal bidirectional MCP runtime for research agent notebooks (~600 LOC).

**Philosophy**: Notebooks ARE agents - each .src.md file becomes a runnable MCP server/client

## Features

- ✅ Parse .src.md notebooks into executable cells
- ✅ Execute TypeScript cells in isolation
- ✅ MCP Server: Expose notebook operations (execute_cell, read_cell, write_cell)
- ✅ MCP Client: Call external services (arxiv, exa, firecrawl) and peer agents
- ✅ Auto-inject helpers: state-tracker, retrieve(), discoverPeers()
- ✅ Docker deployment for multi-agent systems
- ✅ Zero-config peer discovery

## Quick Start

### Local Development

```bash
# Install dependencies
npm install

# Run single agent
AGENT_ID=agent-1 NOTEBOOK_PATH=./examples/research-agent-example.src.md npm run dev
```

### Multi-Agent with Docker

```bash
# Build
docker build -t research-agent .

# Run 3-agent system
docker-compose up
```

## Architecture

```
research-agent.src.md
  ↓ parse
Notebook (cells)
  ↓ inject helpers
Notebook + state-tracker + retrieve() + peers()
  ↓ create runtime
MCP Server (expose) + MCP Client (call)
  ↓ connect
Agent ready for bidirectional MCP communication
```

## Usage

### Create Research Agent

```markdown
<!-- srcbook:{"language":"typescript"} -->

# My Research Agent

###### package.json
\`\`\`json
{
  "type": "module",
  "dependencies": {
    "@modelcontextprotocol/sdk": "^1.17.1"
  }
}
\`\`\`

###### search.ts
\`\`\`typescript
import { retrieve } from './helpers.js';
import { recordEvidence } from './state-tracker.js';

const results = await retrieve("my research query");
recordEvidence(results.length);

export { results };
\`\`\`
```

### Run Agent

```bash
AGENT_ID=my-agent NOTEBOOK_PATH=./my-agent.src.md research-agent
```

## Auto-Injected Helpers

Every notebook automatically gets:

### state-tracker.ts
```typescript
export const state = {evidenceCount: 0, phase: 'observe', iteration: 0};
export function recordEvidence(count: number): void;
export function checkGate(gateName: string): boolean;
```

### helpers.ts
```typescript
// Smart retrieval: Exa → Firecrawl hierarchy
export async function retrieve(query: string, url?: string): Promise<any>;

// Peer discovery
export function discoverPeers(): string[];
export async function callPeer(agentId: string, tool: string, args: any): Promise<any>;
```

## Multi-Agent Patterns

### Pattern: Collaborative Research

```yaml
services:
  agent-1:  # Literature scout
  agent-2:  # Technical analyst
  agent-3:  # Synthesis specialist
```

Each agent:
- Runs same or different notebooks
- Can call peer agents via MCP
- Shares findings through notebook cells
- State tracked independently

### Pattern: Federated Knowledge Graph

Agents build local graphs, query peers' graphs via MCP:

```typescript
// In agent-1
const local = myGraph.query("concept");
const fromAgent2 = await callPeer('agent-2', 'query_graph', {concept: "concept"});
```

## MCP Tools Exposed

Each agent exposes:
- `execute_cell(cellId, params)` - Run a specific cell
- `read_cell(cellId)` - Get cell content
- `write_cell(cellId, content)` - Update cell (for collaboration)

## MCP Resources Exposed

All code cells available as resources:
- `notebook:///cell-id`

## Environment Variables

- `AGENT_ID` - Unique agent identifier
- `AGENT_ROLE` - Agent role/specialization
- `NOTEBOOK_PATH` - Path to .src.md file
- `WORKDIR` - Working directory
- `PEER_AGENT_URLS` - Comma-separated peer URLs
- `EXA_API_KEY` - Exa API key
- `FIRECRAWL_API_KEY` - Firecrawl API key

## Implementation Size

- srcmd-parser.ts: ~100 LOC
- executor.ts: ~100 LOC
- mcp-server.ts: ~130 LOC
- mcp-client.ts: ~80 LOC
- helpers.ts: ~90 LOC
- index.ts: ~150 LOC

**Total**: ~650 LOC (+ templates)

## License

MIT