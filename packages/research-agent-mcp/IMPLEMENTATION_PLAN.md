# Notebook-as-MCP-Agent: Complete Implementation Plan

**Current Status**: ~110 LOC (parser + executor + runtime)
**Target**: ~600 LOC (full bidirectional MCP)
**Remaining**: ~350 LOC

---

## What We Have

✅ **Core Runtime** (~110 LOC):

- types.ts (20 LOC) - Interfaces
- srcmd-parser.ts (23 LOC) - Parse .src.md files
- executor.ts (45 LOC) - Execute TypeScript cells
- index.ts (22 LOC) - Main entry point

✅ **Build System**:

- TypeScript compilation
- MCP SDK 1.17.1 dependency
- Example notebooks

---

## What We Need (Based on Claude Code Docs)

### 1. MCP Server Implementation (~130 LOC)

**Purpose**: Expose notebook operations as MCP tools

**From Claude Code docs**:

- Stdio transport for local servers
- Tool handler pattern
- Resource exposure for notebook cells
- Official SDK server pattern

**Tools to expose**:

```typescript
execute_cell(cellId: string, params?: object) → ExecutionResult
read_cell(cellId: string) → string
write_cell(cellId: string, content: string) → void
list_cells() → Cell[]
get_state() → AgentState
```

**Resources to expose**:

```
notebook:///cell-id → Cell content
notebook:///state → Current state
```

**Implementation**: Enhance existing `mcp-server.ts` (currently stubbed)

---

### 2. MCP Client Implementation (~80 LOC)

**Purpose**: Connect to external MCP servers + peer agents

**From Claude Code docs**:

- Stdio transport for local servers (arxiv, exa, firecrawl)
- HTTP transport for peer agents (agent-1:9001, agent-2:9002)
- Tool calling pattern
- Connection management

**External servers to connect**:

- arxiv-paper-mcp (stdio)
- exa (stdio)
- firecrawl (stdio)
- context7 (stdio)

**Peer agents to connect**:

- agent-1, agent-2, agent-3, etc. (HTTP on ports 9001-9005)

**Implementation**: Complete `mcp-client.ts` with transport setup

---

### 3. Helper Injection System (~50 LOC)

**Purpose**: Auto-inject helper cells into every notebook

**Cells to inject**:

**3.1 state-tracker.ts** (~30 lines):

```typescript
export const state = {evidenceCount: 0, phase: 'observe', iteration: 0};
export function recordEvidence(count: number): void;
export function checkGate(gateName: string): boolean;
```

**3.2 helpers.ts** (~40 lines):

```typescript
// Exa → Firecrawl hierarchy
export async function retrieve(query: string, url?: string): Promise<any>;
// Peer discovery
export function discoverPeers(): string[];
export async function callPeer(agentId, tool, args): Promise<any>;
```

**Implementation**: Create injection logic in `index.ts`

---

### 4. Main Runtime Integration (~50 LOC)

**Purpose**: Wire everything together

**Flow**:

```
1. Parse notebook.src.md
2. Inject helper cells
3. Create MCP server (expose tools)
4. Create MCP client (connect to external + peers)
5. Make client available to cells (global.mcpClient)
6. Start MCP server
```

**Implementation**: Update existing `index.ts`

---

### 5. HTTP Transport for Peer Agents (~40 LOC)

**Purpose**: Allow agents to call each other via HTTP

**From docs**: HTTP transport for remote servers

**Pattern**:

```typescript
// In MCP client
const httpTransport = new HttpClientTransport({
  url: 'http://agent-2:9002/mcp'
});
await client.connect(httpTransport);
```

**Implementation**: Add HTTP transport to `mcp-client.ts`

---

## Implementation Order (Parallelizable)

### Phase 1: Complete MCP Server (~130 LOC)

**File**: `src/mcp-server.ts`

**Tasks** (sequential):

1. Implement tool handlers for execute_cell, read_cell, write_cell
2. Implement resource handlers for notebook cells
3. Add proper error handling
4. Test: Can expose tools via MCP

**Estimated**: 1-2 hours

---

### Phase 2: Complete MCP Client (~80 LOC)

**File**: `src/mcp-client.ts`

**Tasks** (can parallelize with Phase 1):

1. Add stdio transport connections (arxiv, exa, firecrawl)
2. Add HTTP transport for peers
3. Implement connection pooling
4. Add tool calling helpers

**Estimated**: 1-2 hours

---

### Phase 3: Helper Injection (~50 LOC)

**File**: `src/helpers.ts` + update `src/index.ts`

**Tasks** (can parallelize with Phases 1 & 2):

1. Generate state-tracker code
2. Generate retrieve() helper code
3. Generate peer discovery code
4. Add injection logic to main runtime

**Estimated**: 30 min - 1 hour

---

### Phase 4: Integration & Testing (~40 LOC + testing)

**Files**: Update `src/index.ts`, create tests

**Tasks** (sequential, after 1-3):

1. Wire MCP server/client into main runtime
2. Make client available globally
3. Test with example notebook
4. Test multi-agent communication

**Estimated**: 1-2 hours

---

## Testing Strategy

### Test 1: Single Agent

```bash
AGENT_ID=agent-1 \
NOTEBOOK_PATH=examples/research-agent-example.src.md \
node dist/index.js
```

**Verify**:

- MCP server starts
- Can list tools (execute_cell, read_cell, write_cell)
- Can execute cells
- State tracker works

### Test 2: External MCP Integration

**Verify**:

- Agent can call arxiv-paper-mcp
- Agent can call exa
- Agent can call firecrawl
- retrieve() helper uses Exa → Firecrawl correctly

### Test 3: Multi-Agent

```bash
docker-compose up
```

**Verify**:

- 3 agents start
- Agent 1 can call Agent 2's tools
- Agent 2 can call Agent 3's tools
- Shared notebook collaboration works

---

## Key Decisions from Claude Code Docs

**1. Transport Choice**:

- ✅ **Stdio** for external servers (arxiv, exa, firecrawl)
- ✅ **HTTP** for peer agents (easier than stdio in containers)

**2. Tool Naming**:

- Format: `mcp__servername__toolname`
- Our agents: `mcp__agent-1__execute_cell`
- External: `mcp__exa__web_search_exa`

**3. Resource URIs**:

- Format: `notebook:///cell-id`
- Exposed via MCP resources

**4. Authentication**:

- Environment variables for API keys
- OAuth not needed for our use case

**5. Error Handling**:

- Check connection status on init
- Graceful degradation if MCP server unavailable

---

## Code Organization

```
packages/research-agent-mcp/src/
├── types.ts              # ✅ Done (20 LOC)
├── srcmd-parser.ts       # ✅ Done (23 LOC)
├── executor.ts           # ✅ Done (45 LOC)
├── mcp-server.ts         # 🚧 Update (current: stub, target: ~130 LOC)
├── mcp-client.ts         # 🚧 Update (current: stub, target: ~80 LOC)
├── helpers.ts            # 🚧 Update (current: stub, target: ~50 LOC)
└── index.ts              # 🚧 Update (current: 22 LOC, target: ~70 LOC)
```

**Total target**: ~420 LOC (types + parser + executor + server + client + helpers + index)

Plus templates and examples: ~180 LOC

**Grand total**: ~600 LOC ✅

---

## Next Immediate Steps

1. ✅ Review plan (you are here)
2. Implement complete MCP server (mcp-server.ts)
3. Implement complete MCP client (mcp-client.ts)
4. Complete helper injection (helpers.ts + index.ts)
5. Test single agent
6. Test multi-agent
7. Document and commit

---

## Success Criteria

- [ ] Agent exposes MCP server on stdio
- [ ] Agent can call external MCP servers (arxiv, exa, firecrawl)
- [ ] Agent can call peer agent tools via HTTP
- [ ] Notebook cells can use retrieve() helper
- [ ] Notebook cells can use callPeer() helper
- [ ] State tracker auto-injected and working
- [ ] Example 3-agent research completes successfully
- [ ] Total LOC ≤ 650 (including templates)

Ready to implement?
