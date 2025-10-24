# Widescreen Research MCP Server - Verification Checklist

**Purpose**: Comprehensive verification checklist for Claude Code (or other AI agents) to validate the widescreen-research MCP server's functionality and build a complete understanding of its capabilities.

**Server**: widescreen-research MCP server
**Version**: 1.0.0
**Architecture**: Coordinator-Drone distributed research system with elicitation-based qualification

---

## Quick Start Verification

### ✓ Basic Connectivity

- [ ] Server responds to MCP handshake
- [ ] Server name is "widescreen-research"
- [ ] Server version is "1.0.0"
- [ ] Main tool "widescreen-research" is available

**Test**:

```json
{
  "tool": "widescreen-research",
  "arguments": {
    "operation": "start"
  }
}
```

**Expected**: Elicitation response with initial questions

---

## Phase 1: Elicitation Workflow Verification

### 1.1 Initial Elicitation

**Objective**: Verify the elicitation process starts correctly.

- [ ] Calling with `operation: "start"` returns elicitation questions
- [ ] Response type is "elicitation"
- [ ] Session ID is generated and returned
- [ ] Initial questions include:
  - [ ] `research_topic` (text, required)
  - [ ] `researcher_count` (number, required, 1-100)
  - [ ] `research_depth` (select, required)

**Test**:

```json
{
  "tool": "widescreen-research",
  "arguments": {
    "operation": "start"
  }
}
```

**Document**:

- Session ID received: `_______________`
- Number of initial questions: `_______________`
- Question IDs: `_______________`

---

### 1.2 Multi-Stage Elicitation

**Objective**: Verify the elicitation progresses through multiple stages.

- [ ] Answering initial questions returns next set of questions
- [ ] Session state transitions: initial → workflow → advanced → complete
- [ ] Each stage has appropriate questions
- [ ] Session ID remains consistent across stages

**Test Sequence**:

1. Start elicitation (get session_id)
2. Answer initial questions
3. Verify workflow questions are returned
4. Answer workflow questions
5. Verify advanced questions are returned
6. Answer advanced questions
7. Verify completion (type: "ready")

**Document**:

```
Stage 1 (initial): ___ questions
Stage 2 (workflow): ___ questions
Stage 3 (advanced): ___ questions
Completion message: _______________
```

---

### 1.3 Elicitation Completion

**Objective**: Verify elicitation completes and provides research config.

- [ ] Final response has type "ready"
- [ ] Research config is included in response
- [ ] Config contains all elicited parameters:
  - [ ] session_id
  - [ ] topic
  - [ ] researcher_count
  - [ ] research_depth
  - [ ] output_format
  - [ ] timeout_minutes
  - [ ] priority_level

**Test**:
Complete full elicitation workflow and verify final response.

**Document**:

```json
{
  "type": "ready",
  "session_id": "...",
  "config": {
    "topic": "...",
    "researcher_count": ...,
    ...
  }
}
```

---

## Phase 2: Operation Verification

### 2.1 Orchestrate Research

**Objective**: Verify the main research orchestration operation.

**Prerequisites**:

- [ ] GCP credentials configured
- [ ] GOOGLE_CLOUD_PROJECT environment variable set
- [ ] EXA_API_KEY environment variable set (for research drones)

**Checks**:

- [ ] Operation accepts session_id from completed elicitation
- [ ] Operation provisions research drones on Cloud Run
- [ ] Operation distributes research tasks to drones
- [ ] Operation collects results via Pub/Sub queue
- [ ] Operation generates final research report
- [ ] Operation returns ResearchResult with metrics

**Test** (requires GCP setup):

```json
{
  "tool": "widescreen-research",
  "arguments": {
    "operation": "orchestrate-research",
    "session_id": "session-uuid-from-elicitation",
    "parameters": {
      "additional_config": "optional"
    }
  }
}
```

**Expected**:

```json
{
  "session_id": "...",
  "status": "completed",
  "report_url": "...",
  "metrics": {
    "drones_provisioned": 10,
    "drones_completed": 10,
    "drones_failed": 0,
    "total_duration": "...",
    "data_points_collected": 150,
    "cost_estimate": 2.50
  }
}
```

**Document**:

- Drones provisioned: `_______________`
- Research duration: `_______________`
- Data points collected: `_______________`
- Cost estimate: `_______________`
- Report generated: Yes/No

---

### 2.2 Sequential Thinking

**Objective**: Verify the sequential thinking operation for complex reasoning.

**Checks**:

- [ ] Operation accepts problem description
- [ ] Operation accepts optional context
- [ ] Operation accepts optional max_steps parameter
- [ ] Operation returns structured thought steps
- [ ] Operation returns final solution
- [ ] Operation returns confidence score

**Test**:

```json
{
  "tool": "widescreen-research",
  "arguments": {
    "operation": "sequential-thinking",
    "parameters": {
      "problem": "How can we optimize research drone allocation for cost efficiency?",
      "context": "We have a budget of $100 and need to research 50 topics",
      "max_steps": 5
    }
  }
}
```

**Expected**:

```json
{
  "thoughts": [
    {
      "step": 1,
      "thought": "...",
      "reasoning": "...",
      "confidence": 0.85
    },
    ...
  ],
  "solution": "...",
  "confidence": 0.90
}
```

**Document**:

- Number of thought steps: `_______________`
- Solution quality: Good/Fair/Poor
- Confidence score: `_______________`

---

### 2.3 GCP Provisioning

**Objective**: Verify GCP resource provisioning operation.

**Prerequisites**:

- [ ] GCP credentials configured
- [ ] Appropriate IAM permissions

**Checks**:

- [ ] Operation provisions Cloud Run services
- [ ] Operation provisions Pub/Sub topics
- [ ] Operation provisions Firestore collections
- [ ] Operation returns resource details (ID, URL, status)
- [ ] Operation handles region specification

**Test** (requires GCP setup):

```json
{
  "tool": "widescreen-research",
  "arguments": {
    "operation": "gcp-provision",
    "parameters": {
      "resource_type": "cloud_run",
      "count": 3,
      "region": "us-central1",
      "config": {
        "cpu": "1000m",
        "memory": "512Mi"
      }
    }
  }
}
```

**Expected**:

```json
{
  "resources": [
    {
      "id": "drone-1",
      "type": "cloud_run",
      "url": "https://...",
      "status": "ready",
      "region": "us-central1",
      "created_at": "..."
    },
    ...
  ],
  "status": "success"
}
```

**Document**:

- Resources provisioned: `_______________`
- Provisioning time: `_______________`
- Resource URLs: `_______________`

---

### 2.4 Analyze Findings

**Objective**: Verify data analysis operation.

**Checks**:

- [ ] Operation accepts array of DroneResult objects
- [ ] Operation accepts analysis_type parameter
- [ ] Operation returns summary
- [ ] Operation returns insights
- [ ] Operation returns patterns
- [ ] Operation returns statistics

**Test**:

```json
{
  "tool": "widescreen-research",
  "arguments": {
    "operation": "analyze-findings",
    "parameters": {
      "data": [
        {
          "drone_id": "drone-1",
          "status": "completed",
          "data": {"findings": "..."},
          "completed_at": "...",
          "processing_time": "30s"
        }
      ],
      "analysis_type": "comprehensive"
    }
  }
}
```

**Expected**:

```json
{
  "summary": "...",
  "insights": ["...", "..."],
  "patterns": [
    {
      "name": "...",
      "description": "...",
      "frequency": 5,
      "confidence": 0.85
    }
  ],
  "statistics": {
    "total_data_points": 100,
    "avg_processing_time": "25s"
  }
}
```

**Document**:

- Insights generated: `_______________`
- Patterns detected: `_______________`
- Analysis quality: Good/Fair/Poor

---

## Phase 3: Error Handling Verification

### 3.1 Invalid Operation

**Objective**: Verify handling of unknown operations.

**Test**:

```json
{
  "tool": "widescreen-research",
  "arguments": {
    "operation": "invalid-operation"
  }
}
```

**Expected**: Error response with clear message

**Document**:

- Error message: `_______________`
- Error code: `_______________`
- Graceful handling: Yes/No

---

### 3.2 Missing Required Parameters

**Objective**: Verify parameter validation.

**Tests**:

- [ ] Missing session_id for orchestrate-research
- [ ] Missing problem for sequential-thinking
- [ ] Missing resource_type for gcp-provision
- [ ] Missing data for analyze-findings

**Document**:

```
Test: Missing session_id
Response: _______________
Graceful: Yes/No
```

---

### 3.3 Invalid Session ID

**Objective**: Verify session management.

**Test**:

```json
{
  "tool": "widescreen-research",
  "arguments": {
    "operation": "orchestrate-research",
    "session_id": "invalid-session-id"
  }
}
```

**Expected**: Error indicating invalid or expired session

**Document**:

- Error handling: Graceful/Crash
- Error message clarity: Good/Fair/Poor

---

## Phase 4: Integration Verification

### 4.1 Orchestrator Integration

**Objective**: Verify the orchestrator component functions correctly.

**Checks**:

- [ ] Orchestrator can act as MCP client
- [ ] Orchestrator can connect to drone MCP servers
- [ ] Orchestrator can call tools on drones
- [ ] Orchestrator can collect results from drones
- [ ] Orchestrator handles drone failures gracefully

**Method**: Review orchestrator logs during research operation

**Document**:

- Drone connections: Successful/Failed
- Tool calls to drones: `_______________`
- Result collection: Complete/Partial

---

### 4.2 GCP Integration

**Objective**: Verify GCP service integrations.

**Checks**:

- [ ] Cloud Run deployment works
- [ ] Pub/Sub topic creation works
- [ ] Pub/Sub message publishing works
- [ ] Pub/Sub message consumption works
- [ ] Firestore document storage works
- [ ] Firestore document retrieval works
- [ ] Secret Manager integration works (for API keys)

**Method**: Monitor GCP console during operations

**Document**:

- Cloud Run services created: `_______________`
- Pub/Sub topics created: `_______________`
- Firestore collections used: `_______________`

---

### 4.3 External API Integration

**Objective**: Verify external API integrations (Exa AI, etc.).

**Checks**:

- [ ] Exa API key is properly configured
- [ ] Research drones can call Exa API
- [ ] API responses are properly parsed
- [ ] API errors are handled gracefully

**Method**: Review drone logs during research

**Document**:

- API calls successful: `_______________`
- API errors encountered: `_______________`
- Error handling: Graceful/Crash

---

## Phase 5: Performance & Scale Verification

### 5.1 Small Scale (1-5 drones)

**Objective**: Verify performance with small drone fleet.

**Test**: Run research with 1-5 drones

**Metrics**:

- [ ] Provisioning time: `_______________`
- [ ] Research completion time: `_______________`
- [ ] Cost estimate: `_______________`
- [ ] Success rate: `_______________`

---

### 5.2 Medium Scale (10-20 drones)

**Objective**: Verify performance with medium drone fleet.

**Test**: Run research with 10-20 drones

**Metrics**:

- [ ] Provisioning time: `_______________`
- [ ] Research completion time: `_______________`
- [ ] Cost estimate: `_______________`
- [ ] Success rate: `_______________`

---

### 5.3 Large Scale (50-100 drones)

**Objective**: Verify performance at maximum scale.

**Test**: Run research with 50-100 drones (if budget allows)

**Metrics**:

- [ ] Provisioning time: `_______________`
- [ ] Research completion time: `_______________`
- [ ] Cost estimate: `_______________`
- [ ] Success rate: `_______________`
- [ ] Resource limits hit: Yes/No

---

## Phase 6: Capability Summary

### What Widescreen-Research CAN Do

- [ ] **Elicitation-based research qualification**: Interactive questioning to understand research needs
- [ ] **Distributed research orchestration**: Provision and manage 1-100 research drones in parallel
- [ ] **Sequential thinking**: Advanced reasoning for complex problems
- [ ] **GCP resource management**: Automated provisioning of Cloud Run, Pub/Sub, Firestore
- [ ] **Data analysis**: Comprehensive analysis of research findings with pattern detection
- [ ] **Report generation**: AI-powered report generation from collected data
- [ ] **Bidirectional MCP**: Acts as both MCP server (to Claude Code) and MCP client (to drones)
- [ ] **Queue-based result collection**: Asynchronous result aggregation via Pub/Sub
- [ ] **Cost estimation**: Provides cost estimates for research operations
- [ ] **Progress tracking**: Real-time monitoring of research progress

### What Widescreen-Research CANNOT Do

- [ ] **Synchronous research**: All research is asynchronous via drone provisioning
- [ ] **Non-GCP deployment**: Requires GCP infrastructure (Cloud Run, Pub/Sub, Firestore)
- [ ] **Offline operation**: Requires internet connectivity for GCP and external APIs
- [ ] **Real-time streaming**: Results are collected asynchronously, not streamed
- [ ] **Custom drone types**: Limited to pre-defined drone types (research, scraper, etc.)
- [ ] **Direct data access**: Cannot directly access user's local files or databases
- [ ] **Long-term storage**: Research results are temporary unless explicitly saved

### Dependencies

- [ ] **Required**:
  - Google Cloud Platform account
  - GOOGLE_CLOUD_PROJECT environment variable
  - GCP APIs enabled: Cloud Run, Pub/Sub, Firestore
  - IAM permissions for resource provisioning
  - EXA_API_KEY for research capabilities

- [ ] **Optional**:
  - CLAUDE_API_KEY for enhanced AI capabilities
  - Custom MCP server URLs (EXA_MCP_URL, WEB_RESEARCH_MCP_URL)

### Best Use Cases

- [ ] Large-scale parallel research (10-100 topics)
- [ ] Comprehensive market research
- [ ] Competitive analysis
- [ ] Data collection from multiple sources
- [ ] Research requiring horizontal scaling

### Not Suited For

- [ ] Single-topic deep research (use simpler tools)
- [ ] Real-time data streaming
- [ ] Low-latency requirements (provisioning takes 1-2 minutes)
- [ ] Budget-constrained scenarios (minimum cost ~$0.50 per research session)
- [ ] Offline or air-gapped environments

---

## Verification Completion Checklist

- [ ] All elicitation stages tested
- [ ] All operations tested (orchestrate-research, sequential-thinking, gcp-provision, analyze-findings)
- [ ] Error handling verified
- [ ] GCP integration verified
- [ ] External API integration verified
- [ ] Performance at different scales measured
- [ ] Capabilities documented
- [ ] Limitations documented
- [ ] Dependencies documented
- [ ] Use cases identified

---

**Verification Status**: ☐ Not Started | ☐ In Progress | ☐ Complete

**Verified By**: `_______________`
**Date**: `_______________`
**Notes**: `_______________`
