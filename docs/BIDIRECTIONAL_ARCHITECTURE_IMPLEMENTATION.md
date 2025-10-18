# Bidirectional Architecture Implementation

**Date:** October 6, 2025
**Status:** Implementation Complete
**Based on:** `cmd/widescreen-research-mcp/specs/target-bidirectional-architecture.md`

---

## Summary

This document describes the implementation of the bidirectional MCP architecture as specified in the target architecture document. The system now has three distinct layers:

1. **Client Layer**: Claude Desktop (MCP Client)
2. **Host Layer**: `cmd/widescreen-research` (Bidirectional - MCP Server + MCP Client)
3. **Orchestration Layer**: `cmd/widescreen-research-mcp` (Server-Only)

---

## Implementation Changes

### 1. New Component: `cmd/widescreen-research` (Host Layer)

Created a new bidirectional MCP entity that serves as both:

- **MCP Server** to Claude Desktop
- **MCP Client** to `widescreen-research-mcp`

#### Structure

```
cmd/widescreen-research/
├── main.go                    # Entry point with CLI flags
├── server/
│   └── server.go             # MCP server implementation
├── client/
│   └── orchestrator_client.go # MCP client to widescreen-research-mcp
└── aggregator/
    └── aggregator.go         # Report generation and aggregation
```

#### Key Features

**Exposed Tools (to Claude Desktop):**

- `orchestrate-research` - Start distributed research
- `get-report` - Retrieve completed reports
- `list-sessions` - List all research sessions

**Internal Capabilities:**

- Session state management
- Result collection from orchestrator
- Report aggregation (moved from orchestrator)
- Concurrent result processing

### 2. Refactored: `cmd/widescreen-research-mcp` (Orchestration Layer)

Modified to be server-only with GCP-specific responsibilities:

#### Changes Made

1. **Tool Rename**: `orchestrate-research` → `start-gcp-orchestration`
   - Location: `cmd/widescreen-research-mcp/server/server.go:58`
   - Clarifies this is purely for GCP orchestration

2. **Report Generation Removed**
   - Location: `cmd/widescreen-research-mcp/orchestrator/orchestrator.go:171-177`
   - Removed `generateReport()` call
   - Removed `analyzeResults()` call
   - Now returns raw `DroneResult[]` instead of `ResearchReport`

3. **Schema Updates**
   - Location: `cmd/widescreen-research-mcp/schemas/schemas.go:23-32`
   - Added `Results []DroneResult` field to `ResearchResult`
   - Marked `ReportURL` and `ReportData` as deprecated

#### Retained Responsibilities

- Cloud Run drone provisioning
- Pub/Sub topic management
- Drone task distribution
- Result collection via Pub/Sub
- Resource cleanup

### 3. New Component: Report Aggregator

Location: `cmd/widescreen-research/aggregator/aggregator.go`

**Responsibilities:**

- Aggregate drone results into master report
- Pattern detection across results
- Insight generation
- Statistics calculation
- Markdown report rendering
- Artifact storage

**Key Methods:**

- `GenerateReport()` - Main aggregation entry point
- `analyzeResults()` - Cross-drone analysis
- `extractPatterns()` - Pattern identification
- `generateInsights()` - Key finding extraction
- `saveReportArtifacts()` - Report persistence

### 4. MCP Client Implementation

Location: `cmd/widescreen-research/client/orchestrator_client.go`

**Features:**

- Stdio transport to `widescreen-research-mcp`
- Tool invocation via MCP protocol
- Result parsing and validation
- Connection lifecycle management

**API:**

- `Connect(ctx)` - Establish MCP connection
- `StartGCPOrchestration(ctx, config)` - Delegate to orchestrator
- `Close()` - Clean shutdown

---

## Data Flow

### Complete Research Workflow

```
1. Claude Desktop → widescreen-research
   Tool: orchestrate-research
   Data: {topic, researcher_count, ...}

2. widescreen-research → widescreen-research-mcp
   Tool: start-gcp-orchestration
   Data: {session_id, topic, researcher_count, ...}

3. widescreen-research-mcp
   - Provisions N Cloud Run drones
   - Sends tasks to drones via HTTP
   - Collects results via Pub/Sub

4. widescreen-research-mcp → widescreen-research
   Returns: {session_id, status, results: [DroneResult...]}

5. widescreen-research
   - Aggregates drone results
   - Generates master report
   - Stores report

6. widescreen-research → Claude Desktop
   Tool Response: {session_id, status, message}

7. Claude Desktop → widescreen-research
   Tool: get-report {session_id}

8. widescreen-research → Claude Desktop
   Returns: ResearchReport (JSON)
```

---

## Configuration

### Running the Host

```bash
# Stdio mode (default) - connects to orchestrator via stdio
go run ./cmd/widescreen-research

# Specify orchestrator location
go run ./cmd/widescreen-research --orchestrator stdio://widescreen-research-mcp

# HTTP mode for serving MCP over HTTP
go run ./cmd/widescreen-research --http --port 8080
```

### Environment Variables

**Host Application:**

- `MCP_TRANSPORT` - Transport mode (stdio or http)
- `PORT` - HTTP port (if --http)
- `ORCHESTRATOR_URL` - URL to widescreen-research-mcp (stdio:// or http://)

**Orchestrator (unchanged):**

- `GOOGLE_CLOUD_PROJECT` - GCP project ID
- `GOOGLE_CLOUD_REGION` - GCP region
- `EXA_API_KEY` - Exa AI API key
- `LOG_LEVEL` - Logging verbosity

---

## Build Verification

Both components build successfully:

```bash
$ go build ./cmd/widescreen-research
# Success (no output)

$ go build ./cmd/widescreen-research-mcp
# Success (no output)
```

---

## Architectural Benefits

### 1. Separation of Concerns

- **Host**: User interaction, session management, report generation
- **Orchestrator**: GCP resource management, distributed execution
- **Drones**: Individual research tasks

### 2. Scalability

- Report aggregation happens in host (not in GCP)
- Reduces orchestrator load
- Better cost management (host can run locally)

### 3. Flexibility

- Host can connect to different orchestrators
- Orchestrator can be swapped/upgraded independently
- Multiple hosts can share one orchestrator

### 4. Protocol Clarity

- `orchestrate-research` → User-facing research initiation
- `start-gcp-orchestration` → Infrastructure provisioning
- Clear intent from tool names

---

## Future Enhancements

1. **Resource Protocol**
   - Implement `widescreen://reports/{session_id}` resources
   - Allow direct resource access from Claude Desktop

2. **HTTP Client Transport**
   - Implement HTTP client transport for remote orchestrators
   - Currently only stdio is supported

3. **Streaming Results**
   - Stream drone results as they complete
   - Progressive report updates

4. **AI-Powered Aggregation**
   - Use LLM for intelligent result synthesis
   - Advanced pattern detection

5. **Multi-Orchestrator Support**
   - Route sessions to different orchestrators
   - Load balancing across GCP regions

---

## Testing Recommendations

### Unit Testing

```bash
# Test orchestrator (GCP provisioning)
go test ./cmd/widescreen-research-mcp/orchestrator/...

# Test aggregator (report generation)
go test ./cmd/widescreen-research/aggregator/...

# Test MCP client
go test ./cmd/widescreen-research/client/...
```

### Integration Testing

1. **Local Mode**: Test without GCP
   - Mock drone responses
   - Verify report aggregation

2. **GCP Mode**: Full end-to-end
   - Provision real drones
   - Verify Pub/Sub flow
   - Check report quality

### MCP Protocol Testing

```bash
# Test host as MCP server
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  go run ./cmd/widescreen-research

# Test orchestrator as MCP server
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  go run ./cmd/widescreen-research-mcp
```

---

## Migration Notes

### Breaking Changes

1. **Tool Name Change**
   - `orchestrate-research` on widescreen-research-mcp is now `start-gcp-orchestration`
   - Direct users of widescreen-research-mcp must update tool calls

2. **Response Format**
   - widescreen-research-mcp now returns `results` array instead of `report_data`
   - Clients expecting reports must use the new host layer

### Backward Compatibility

- Old `ReportURL` and `ReportData` fields kept in schemas (deprecated)
- widescreen-research-mcp can still run standalone (but returns raw results)

---

## Related Documentation

- Target Spec: `cmd/widescreen-research-mcp/specs/target-bidirectional-architecture.md`
- Project README: `README.md`
- Claude Instructions: `CLAUDE.md`
- Operations Guide: `docs/OPERATIONS.md` (needs update)

---

## Implementation Checklist

- [x] Create `cmd/widescreen-research` structure
- [x] Implement MCP server (host)
- [x] Implement MCP client (to orchestrator)
- [x] Create report aggregator
- [x] Refactor orchestrator to remove report generation
- [x] Rename tool: orchestrate-research → start-gcp-orchestration
- [x] Update ResearchResult schema
- [x] Build verification
- [ ] Integration tests
- [ ] Update CLAUDE.md
- [ ] Update README.md

---

## Contact

For questions or issues with this implementation, refer to the project specification or open an issue in the repository.
