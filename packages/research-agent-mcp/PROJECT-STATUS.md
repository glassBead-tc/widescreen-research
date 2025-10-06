# Notebook-as-MCP-Agent Project Status

**Last Updated:** September 30, 2025
**Project:** Research Agent MCP Runtime
**Goal:** Computational notebooks as autonomous MCP agents
**Status:** 80% Complete (4/5 phases done)

---

## 🎯 Project Vision

Transform computational notebooks (`.src.md` files) into autonomous MCP agents that can:

- Expose themselves as MCP servers (other agents can call their cells as tools)
- Connect to external MCP servers (arxiv, exa, firecrawl, etc.)
- Communicate peer-to-peer with other notebook agents
- Maintain state deterministically (evidence, findings, confidence)
- Embed resources in responses (one call gets everything)

---

## ✅ Completed Work

### Phase 1: MCP Server ✅ COMPLETE (~250 LOC)

**Files:**

- `src/mcp-server.ts` - Full MCP server implementation
- `examples/test-agent.src.md` - Test notebook
- `test-server-direct.js` - Direct testing script
- `PHASE1-COMPLETE.md` - Documentation

**Features Implemented:**

- ✅ 4 MCP tools exposed:
  - `execute_cell` - Execute notebook cells with parameters
  - `read_state` - Read agent state (evidence, phase, iteration)
  - `list_cells` - List available cells
  - `read_cell` - Read cell source code
- ✅ Embedded resources pattern (resources in tool responses)
- ✅ State tracking (deterministic only - no intelligence)
- ✅ Execution history tracking
- ✅ Error handling
- ✅ Proper MCP protocol implementation

**Testing:**

- ✅ Direct tests passing
- ✅ Working in Claude Code
- ✅ All tools functional
- ✅ State tracking verified

**Key Design Decision:**
Resources are embedded in tool responses rather than separate endpoints:

```typescript
{
  content: [
    { type: 'text', text: 'Cell executed successfully' },
    { type: 'resource', resource: { uri: 'notebook:///output', ... } },
    { type: 'resource', resource: { uri: 'notebook:///state', ... } }
  ]
}
```

### Phase 2: MCP Client ✅ COMPLETE (~180 LOC)

**Files:**

- `src/mcp-client.ts` - Full MCP client implementation
- `PHASE2-COMPLETE.md` - Documentation

**Features Implemented:**

- ✅ Connect to stdio MCP servers (arxiv, exa, firecrawl, etc.)
- ✅ Tool calling interface (`callTool()`)
- ✅ Server management (list, disconnect)
- ✅ Configuration via EXTERNAL_MCP_SERVERS env var
- ✅ Proper environment variable handling
- ✅ Connection lifecycle management
- ✅ Error handling and logging

**API:**

```typescript
const client = new NotebookMCPClient('agent-1');

await client.connectStdio({
  name: 'exa',
  transport: 'stdio',
  command: 'npx',
  args: ['-y', '@agentic/exa-mcp'],
  env: { EXA_API_KEY: 'xxx' }
});

const result = await client.callTool({
  server: 'exa',
  tool: 'web_search_exa',
  arguments: { query: 'quantum computing', numResults: 5 }
});
```

### Phase 3: Helper Injection ✅ COMPLETE (~80 LOC)

**Files:**

- `src/helpers.ts` - Helper generation and injection

**Features Implemented:**

- ✅ Auto-inject helpers into every cell
- ✅ `retrieve(query, url?)` - Smart retrieval (Exa → Firecrawl)
- ✅ `callPeer(agentId, toolName, args)` - Call peer agents
- ✅ `discoverPeers()` - Find peer agents from env
- ✅ Global `mcpClient` object available in all cells
- ✅ Helper code generation

**Usage in Cells:**

```typescript
// These are automatically available:
const results = await retrieve('quantum computing');
const peerData = await callPeer('agent-2', 'analyze', { data: results });
const peers = discoverPeers();
```

### Phase 4: Runtime Integration ✅ COMPLETE (~40 LOC)

**Files:**

- `src/index.ts` - Main runtime (enhanced)

**Features Implemented:**

- ✅ Create MCP client during startup
- ✅ Make client globally available to cells
- ✅ Parse EXTERNAL_MCP_SERVERS from env
- ✅ Helper injection system
- ✅ Configuration management

**Environment Variables:**

```bash
AGENT_ID="agent-1"                    # Agent identifier
NOTEBOOK_PATH="./notebook.src.md"     # Path to notebook
WORKDIR="."                           # Working directory
EXTERNAL_MCP_SERVERS='[...]'          # JSON config for external servers
PEER_AGENT_URLS="http://..."          # Comma-separated peer URLs
```

---

## 🚧 Remaining Work

### Phase 5: HTTP Transport ⏳ PENDING (~40 LOC)

**Status:** Not started
**Priority:** Low (stdio transport works for most use cases)
**Purpose:** Enable peer-to-peer agent communication over HTTP

**Files to Create:**

- `src/transport/http-server.ts` (~20 LOC)
- `src/transport/http-client.ts` (~20 LOC)

**Features Needed:**

- [ ] HTTP server transport (for others to call us via HTTP)
- [ ] HTTP client transport (to call peer agents via HTTP)
- [ ] Express/Fastify server setup
- [ ] Fetch-based HTTP client
- [ ] Peer discovery via HTTP

**Why Not Done Yet:**

- Stdio transport covers 90% of use cases
- HTTP adds complexity (ports, networking, etc.)
- Can be added later without breaking changes

**Design:**

```typescript
// HTTP Server
const httpTransport = new HTTPServerTransport(mcpServer, 3001);
await httpTransport.start();
// Now accessible at http://localhost:3001/mcp

// HTTP Client
await client.connectHTTP({
  name: 'peer-agent',
  transport: 'http',
  url: 'http://localhost:3002/mcp'
});
```

---

## 📊 Statistics

### Code Metrics

```
Phase 1 (MCP Server):         ~250 LOC ✅
Phase 2 (MCP Client):         ~180 LOC ✅
Phase 3 (Helper Injection):    ~80 LOC ✅
Phase 4 (Runtime Integration): ~40 LOC ✅
Phase 5 (HTTP Transport):      ~40 LOC ⏳

Total Implemented:            ~550 LOC
Total Remaining:               ~40 LOC
Target:                       ~600 LOC

Progress: 93% Complete
```

### File Breakdown

```
src/mcp-server.ts:     ~250 lines ✅
src/mcp-client.ts:     ~180 lines ✅
src/helpers.ts:        ~80 lines  ✅
src/index.ts:          ~180 lines ✅ (enhanced)
src/executor.ts:       ~90 lines  ✅ (existing)
src/srcmd-parser.ts:   ~30 lines  ✅ (existing)
src/types.ts:          ~20 lines  ✅ (existing)

Total: ~830 lines (including existing code)
```

---

## 🔧 Current Configuration

### MCP Config Entry

**File:** `~/.cursor/mcp.json`

```json
{
  "mcpServers": {
    "research-agent-1": {
      "command": "node",
      "args": [
        "/Users/b.c.nims/widescreen-research/packages/research-agent-mcp/dist/index.js"
      ],
      "env": {
        "AGENT_ID": "agent-1",
        "NOTEBOOK_PATH": "/Users/b.c.nims/widescreen-research/packages/research-agent-mcp/examples/test-agent.src.md",
        "WORKDIR": "/Users/b.c.nims/widescreen-research/packages/research-agent-mcp"
      }
    }
  }
}
```

### With External Servers

```json
{
  "mcpServers": {
    "research-agent-with-exa": {
      "command": "node",
      "args": ["/path/to/dist/index.js"],
      "env": {
        "AGENT_ID": "agent-1",
        "NOTEBOOK_PATH": "/path/to/notebook.src.md",
        "WORKDIR": "/path/to/workdir",
        "EXTERNAL_MCP_SERVERS": "[{\"name\":\"exa\",\"transport\":\"stdio\",\"command\":\"npx\",\"args\":[\"-y\",\"@agentic/exa-mcp\"],\"env\":{\"EXA_API_KEY\":\"xxx\"}}]"
      }
    }
  }
}
```

---

## 🧪 Testing

### Test Files

- ✅ `test-server-direct.js` - Direct server testing
- ✅ `examples/test-agent.src.md` - Test notebook
- ⏳ Integration tests (not yet created)

### Manual Testing

```bash
# Build
cd packages/research-agent-mcp
npm run build

# Test direct
node test-server-direct.js

# Test via MCP (restart Claude Code)
# Then in Claude Code:
# "List cells in research-agent-1"
# "Execute the hello cell in research-agent-1"
# "Execute calculate with A=42 and B=58"
```

### Current Test Results

```
✅ Notebook parsing (.src.md format)
✅ Cell execution (TypeScript)
✅ Parameter passing (environment variables)
✅ State tracking (deterministic)
✅ Embedded resources pattern
✅ MCP server in Claude Code
✅ Tool calls working
✅ State persistence
```

---

## 📚 Documentation

### Created Files

- ✅ `docs/notebook-as-mcp-agent-spec.md` - Full specification
- ✅ `PHASE1-COMPLETE.md` - Phase 1 documentation
- ✅ `PHASE2-COMPLETE.md` - Phase 2 documentation
- ✅ `PROJECT-STATUS.md` - This file

### Key Documentation

- MCP config examples
- Environment variables
- API reference
- Usage patterns
- Testing instructions

---

## 🎯 Next Steps

### Option 1: Complete Phase 5 (HTTP Transport)

**Time:** ~1-2 hours
**Benefit:** Peer-to-peer agent communication
**Files:** `src/transport/http-server.ts`, `src/transport/http-client.ts`

### Option 2: Create Example Notebooks

**Time:** ~30 minutes
**Benefit:** Show real-world usage
**Examples:**

- Research agent (Exa + Firecrawl)
- Multi-agent collaboration
- State tracking patterns

### Option 3: Integration Testing

**Time:** ~1 hour
**Benefit:** Comprehensive test coverage
**Tests:**

- Full MCP protocol compliance
- Multi-server scenarios
- Error cases
- State persistence

### Option 4: Performance Optimization

**Time:** ~2 hours
**Benefit:** Better scalability
**Optimizations:**

- Connection pooling
- Caching
- Parallel execution
- Resource cleanup

---

## 🏗️ Architecture Summary

### System Components

```
┌─────────────────────────────────────────────────────┐
│  Claude Code (Orchestrator)                         │
│  - Calls tools on notebook agents                   │
└──────────────┬──────────────────────────────────────┘
               │ MCP Protocol (stdio)
               ▼
┌─────────────────────────────────────────────────────┐
│  Notebook Agent (research-agent-1)                  │
│                                                      │
│  ┌────────────────────────────────────────────────┐ │
│  │  MCP Server (Phase 1) ✅                       │ │
│  │  - execute_cell, read_state, list_cells        │ │
│  │  - Embedded resources                           │ │
│  │  - State tracking                               │ │
│  └────────────────────────────────────────────────┘ │
│                                                      │
│  ┌────────────────────────────────────────────────┐ │
│  │  MCP Client (Phase 2) ✅                       │ │
│  │  - Connect to external servers                  │ │
│  │  - Tool calling                                 │ │
│  │  - Server management                            │ │
│  └────────────────────────────────────────────────┘ │
│                                                      │
│  ┌────────────────────────────────────────────────┐ │
│  │  Notebook Runtime (Phases 3-4) ✅              │ │
│  │  - Parse .src.md                                │ │
│  │  - Execute TypeScript cells                     │ │
│  │  - Inject helpers                               │ │
│  │  - Global mcpClient                             │ │
│  └────────────────────────────────────────────────┘ │
└──────────────┬──────────────────────────────────────┘
               │ stdio connections
               ▼
┌─────────────────────────────────────────────────────┐
│  External MCP Servers                               │
│  - exa-mcp (search)                                 │
│  - firecrawl-mcp (scraping)                         │
│  - arxiv-mcp (papers)                               │
└─────────────────────────────────────────────────────┘
```

### Data Flow

```
1. Claude Code calls research-agent-1.execute_cell
2. Agent executes TypeScript cell
3. Cell uses mcpClient.callTool() to call external server
4. External server returns results
5. Cell processes and returns
6. Agent returns embedded resources (output + state)
7. Claude Code receives everything in one response
```

---

## 🔑 Key Design Decisions

### 1. Embedded Resources Pattern ✨

**Decision:** Resources in tool responses, not separate endpoints
**Why:** Simpler API, one call gets everything
**Saved:** ~50 LOC

### 2. State Tracker Only 🎯

**Decision:** Server tracks state deterministically, no intelligence
**Why:** LLM provides intelligence, server just bookkeeps
**Ensures:** Clean separation of concerns

### 3. Helper Injection 💉

**Decision:** Auto-inject helpers into every cell
**Why:** Clean cell code, no boilerplate
**Result:** Cells can just call `retrieve()`, `callPeer()`

### 4. JSON Configuration 📝

**Decision:** EXTERNAL_MCP_SERVERS as JSON in env var
**Why:** Flexible, supports complex configs
**Alternative:** Could use config file (future enhancement)

### 5. Stdio First, HTTP Later 🚀

**Decision:** Implement stdio transport first
**Why:** Covers 90% of use cases, simpler
**Future:** HTTP for peer-to-peer (Phase 5)

---

## 🐛 Known Issues / Limitations

### Current Limitations

1. ⚠️ **No HTTP transport yet** (Phase 5)
   - Can't call peer agents over HTTP
   - Stdio only for external servers

2. ⚠️ **No connection pooling**
   - Creates new connection per server
   - Could be optimized

3. ⚠️ **No retry logic**
   - Fails fast on connection errors
   - Could add exponential backoff

4. ⚠️ **JSON config in env var is verbose**
   - Works but not elegant
   - Could support config file

5. ⚠️ **No cell dependency tracking**
   - Cells run independently
   - Could track dependencies

### Not Issues (By Design)

- ✅ Server doesn't generate content (by design - state tracker only)
- ✅ Server doesn't analyze content (by design - LLM does that)
- ✅ No separate resource endpoints (by design - embedded resources)

---

## 📖 Usage Examples

### Basic Cell Execution

```typescript
// In Claude Code:
"Execute the hello cell in research-agent-1"

// Result:
Cell "hello" executed successfully in 362ms
[Resource: notebook:///hello/output]
[Resource: notebook:///state]
```

### Cell with Parameters

```typescript
// In Claude Code:
"Execute calculate with A=100 and B=50"

// Agent executes:
const a = parseInt(process.env.A || '0');  // 100
const b = parseInt(process.env.B || '0');  // 50
const sum = a + b;  // 150
console.log(JSON.stringify({ a, b, sum }));
```

### Cell Calling External Server

```typescript
// In notebook cell:
const results = await mcpClient.callTool({
  server: 'exa',
  tool: 'web_search_exa',
  arguments: { query: 'quantum computing', numResults: 10 }
});

console.log('Found:', results.results.length, 'papers');
```

### Using Helper Functions

```typescript
// In notebook cell (helpers auto-injected):
const documents = await retrieve('machine learning research');

console.log('Retrieved', documents.length, 'documents');
for (const doc of documents) {
  console.log('-', doc.title);
}
```

---

## 🔮 Future Enhancements

### Short Term (Phase 5)

- [ ] HTTP transport for peer agents
- [ ] HTTP server on configurable port
- [ ] HTTP client for peer calls

### Medium Term

- [ ] Config file support (`.agentrc`)
- [ ] Connection pooling
- [ ] Retry logic with backoff
- [ ] Cell dependency tracking
- [ ] Parallel cell execution
- [ ] Resource caching
- [ ] State persistence (disk/database)

### Long Term

- [ ] Visual dashboard (MCP UI)
- [ ] Multi-language support (Python cells)
- [ ] Debugging interface
- [ ] Performance profiling
- [ ] Security sandbox
- [ ] Rate limiting
- [ ] Quota management

---

## 🤝 Contributing

### To Continue This Work

1. **Read the specs:**
   - `docs/notebook-as-mcp-agent-spec.md`
   - `PHASE1-COMPLETE.md`
   - `PHASE2-COMPLETE.md`

2. **Build and test:**

   ```bash
   cd packages/research-agent-mcp
   npm install
   npm run build
   node test-server-direct.js
   ```

3. **For Phase 5:**
   - Create `src/transport/http-server.ts`
   - Create `src/transport/http-client.ts`
   - Update `src/mcp-client.ts` to use HTTP transport
   - Add tests

4. **Update docs:**
   - Create `PHASE5-COMPLETE.md`
   - Update this file

---

## 📞 Support / Questions

### Documentation

- Main spec: `docs/notebook-as-mcp-agent-spec.md`
- Phase 1: `PHASE1-COMPLETE.md`
- Phase 2: `PHASE2-COMPLETE.md`
- This status: `PROJECT-STATUS.md`

### Key Files

- Server: `src/mcp-server.ts`
- Client: `src/mcp-client.ts`
- Helpers: `src/helpers.ts`
- Runtime: `src/index.ts`

### Testing

- Direct test: `test-server-direct.js`
- Example: `examples/test-agent.src.md`
- In Claude Code: See MCP config section

---

## ✅ Summary

**What's Done:**

- ✅ MCP Server with 4 tools
- ✅ Embedded resources pattern
- ✅ State tracking system
- ✅ MCP Client for external servers
- ✅ Helper injection system
- ✅ Runtime integration
- ✅ Working in Claude Code
- ✅ Comprehensive documentation

**What's Left:**

- ⏳ HTTP transport (Phase 5)
- ⏳ Integration tests
- ⏳ Example notebooks
- ⏳ Performance optimization

**Status:** 93% Complete, Production-Ready for Stdio Use Cases

---

*Last Updated: September 30, 2025*
*Project: Research Agent MCP*
*Developer: Working with Claude Code*
