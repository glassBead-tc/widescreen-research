# MCP Server Introspection Template

**Purpose**: A generalizable framework for AI agents to systematically explore, understand, and document any MCP server's capabilities, limitations, and behavior.

**Use Case**: When an AI agent (like Claude Code) connects to an MCP server for the first time, this template guides the agent through a structured discovery process to build a comprehensive mental model of what the server can and cannot do.

---

## Phase 1: Discovery & Enumeration

### 1.1 Server Metadata
**Objective**: Identify the server and its version.

**Checks**:
- [ ] Retrieve server name
- [ ] Retrieve server version
- [ ] Identify server implementation (official SDK, custom, etc.)
- [ ] Check for server description/documentation

**Method**:
```
Call MCP initialization/handshake to get server info
```

**Document**:
- Server name: `_______________`
- Version: `_______________`
- Implementation: `_______________`

---

### 1.2 Tool Discovery
**Objective**: Enumerate all available tools.

**Checks**:
- [ ] List all available tools
- [ ] For each tool, capture:
  - Tool name
  - Tool description
  - Input schema (required parameters)
  - Input schema (optional parameters)
  - Expected output format

**Method**:
```
Call tools/list endpoint or equivalent discovery mechanism
```

**Document**:
```
Tool Inventory:
1. tool_name_1
   - Description: _______________
   - Required params: _______________
   - Optional params: _______________
   - Output: _______________

2. tool_name_2
   ...
```

---

### 1.3 Resource Discovery (if applicable)
**Objective**: Identify available resources (files, data sources, etc.).

**Checks**:
- [ ] List available resources
- [ ] Identify resource URIs and types
- [ ] Check resource access patterns

**Method**:
```
Call resources/list endpoint
```

---

### 1.4 Prompt Discovery (if applicable)
**Objective**: Identify pre-defined prompts or templates.

**Checks**:
- [ ] List available prompts
- [ ] Understand prompt purposes and parameters

**Method**:
```
Call prompts/list endpoint
```

---

## Phase 2: Capability Mapping

### 2.1 Basic Tool Invocation
**Objective**: Verify each tool can be called successfully with minimal valid input.

**For each tool**:
- [ ] Identify minimal required parameters
- [ ] Construct minimal valid request
- [ ] Invoke tool with minimal input
- [ ] Verify successful response
- [ ] Document response structure

**Method**:
```
For tool in tools:
  - Create minimal valid input
  - Call tool
  - Verify response
  - Document behavior
```

**Document**:
```
Tool: tool_name
Minimal Input: { param1: "value1" }
Response: { ... }
Success: ✓/✗
Notes: _______________
```

---

### 2.2 Parameter Exploration
**Objective**: Understand the full range of parameters and their effects.

**For each tool**:
- [ ] Test with all required parameters
- [ ] Test with optional parameters (one at a time)
- [ ] Test with combinations of optional parameters
- [ ] Identify parameter types (string, number, object, array)
- [ ] Identify parameter constraints (min/max, enums, patterns)

**Document**:
```
Tool: tool_name
Parameter: param_name
  - Type: string/number/object/array
  - Required: yes/no
  - Constraints: _______________
  - Effect: _______________
```

---

### 2.3 Workflow Discovery
**Objective**: Identify multi-step workflows or tool sequences.

**Checks**:
- [ ] Identify tools that depend on outputs from other tools
- [ ] Map tool call sequences (e.g., init → execute → collect)
- [ ] Identify session/state management patterns
- [ ] Document typical workflows

**Method**:
```
Analyze tool descriptions and parameters for references to:
- Session IDs
- Task IDs
- State transitions
- Callback patterns
```

**Document**:
```
Workflow: workflow_name
Steps:
1. Call tool_a with params {...}
2. Extract session_id from response
3. Call tool_b with session_id
4. Poll tool_c until status = "complete"
```

---

## Phase 3: Error Boundary Testing

### 3.1 Invalid Input Handling
**Objective**: Understand how the server handles malformed requests.

**For each tool**:
- [ ] Call with missing required parameters
- [ ] Call with invalid parameter types
- [ ] Call with out-of-range values
- [ ] Call with malformed JSON
- [ ] Document error responses

**Document**:
```
Tool: tool_name
Error Scenario: Missing required param
Response: { error: "...", code: "..." }
Behavior: Graceful/Crash/Timeout
```

---

### 3.2 Edge Case Testing
**Objective**: Identify limits and edge cases.

**Checks**:
- [ ] Test with empty strings
- [ ] Test with very long strings
- [ ] Test with special characters
- [ ] Test with null/undefined values
- [ ] Test with extreme numbers (0, negative, very large)

**Document**:
```
Edge Case: Empty string for required param
Result: _______________
```

---

### 3.3 Timeout & Performance
**Objective**: Understand performance characteristics.

**Checks**:
- [ ] Identify long-running operations
- [ ] Test timeout behavior
- [ ] Measure typical response times
- [ ] Identify rate limits (if any)

**Document**:
```
Tool: tool_name
Typical Response Time: ___ ms
Timeout: ___ seconds
Rate Limit: ___ requests/minute
```

---

## Phase 4: Limitation Identification

### 4.1 Functional Limitations
**Objective**: Document what the server CANNOT do.

**Checks**:
- [ ] Identify missing capabilities (compared to similar servers)
- [ ] Identify unsupported operations
- [ ] Identify data format limitations
- [ ] Identify scale limitations (max items, max size, etc.)

**Document**:
```
Limitations:
- Cannot: _______________
- Does not support: _______________
- Maximum: _______________
```

---

### 4.2 Dependency Mapping
**Objective**: Identify external dependencies and requirements.

**Checks**:
- [ ] Identify required environment variables
- [ ] Identify required external services (APIs, databases, etc.)
- [ ] Identify authentication requirements
- [ ] Identify network/firewall requirements

**Document**:
```
Dependencies:
- Environment: REQUIRED_VAR_1, OPTIONAL_VAR_2
- External Services: service_name (purpose)
- Authentication: API key / OAuth / None
- Network: Outbound access to _______________
```

---

### 4.3 State & Persistence
**Objective**: Understand state management.

**Checks**:
- [ ] Identify stateful vs. stateless operations
- [ ] Identify session management patterns
- [ ] Identify data persistence mechanisms
- [ ] Identify cleanup/garbage collection behavior

**Document**:
```
State Management:
- Stateful operations: _______________
- Session lifetime: _______________
- Persistence: In-memory / Database / Filesystem
- Cleanup: Automatic / Manual
```

---

## Phase 5: Integration Mapping

### 5.1 Bidirectional Capabilities
**Objective**: Identify if the server can act as both client and server.

**Checks**:
- [ ] Can the server call other MCP servers?
- [ ] Can the server expose resources to clients?
- [ ] Can the server use sampling (AI assistance)?
- [ ] Can the server use elicitation (user input)?

**Document**:
```
Bidirectional Capabilities:
- Acts as MCP client: Yes/No
- Exposes resources: Yes/No
- Uses sampling: Yes/No
- Uses elicitation: Yes/No
```

---

### 5.2 External Integrations
**Objective**: Map integrations with external systems.

**Checks**:
- [ ] Identify cloud provider integrations (GCP, AWS, Azure)
- [ ] Identify third-party API integrations
- [ ] Identify database integrations
- [ ] Identify messaging/queue integrations

**Document**:
```
External Integrations:
- Cloud: GCP (Cloud Run, Pub/Sub, Firestore)
- APIs: Exa AI, OpenAI, etc.
- Database: PostgreSQL, Firestore, etc.
- Messaging: Pub/Sub, RabbitMQ, etc.
```

---

## Phase 6: Synthesis & Documentation

### 6.1 Capability Summary
**Objective**: Create a high-level summary of what the server does.

**Template**:
```
# [Server Name] Capability Summary

## Primary Purpose
[One-sentence description]

## Core Capabilities
1. [Capability 1]
2. [Capability 2]
...

## Key Workflows
1. [Workflow 1]: [Description]
2. [Workflow 2]: [Description]

## Limitations
- [Limitation 1]
- [Limitation 2]

## Dependencies
- [Dependency 1]
- [Dependency 2]

## Best Suited For
[Use cases where this server excels]

## Not Suited For
[Use cases where this server is not appropriate]
```

---

### 6.2 Mental Model
**Objective**: Create a conceptual model of how the server works.

**Template**:
```
# Mental Model: [Server Name]

## Architecture Pattern
[e.g., Coordinator-Worker, Pipeline, Request-Response, Event-Driven]

## Data Flow
[Describe how data flows through the system]

## Key Abstractions
- [Abstraction 1]: [Explanation]
- [Abstraction 2]: [Explanation]

## Typical Interaction Pattern
1. [Step 1]
2. [Step 2]
3. [Step 3]
```

---

### 6.3 Quick Reference
**Objective**: Create a cheat sheet for common operations.

**Template**:
```
# Quick Reference: [Server Name]

## Common Operations

### [Operation 1]
```json
{
  "tool": "tool_name",
  "arguments": {
    "param1": "value1"
  }
}
```

### [Operation 2]
...
```

---

## Appendix: Verification Checklist

Use this checklist to ensure comprehensive introspection:

- [ ] All tools discovered and documented
- [ ] All parameters understood (required, optional, types, constraints)
- [ ] All workflows mapped
- [ ] Error handling verified
- [ ] Edge cases tested
- [ ] Limitations identified
- [ ] Dependencies documented
- [ ] Integration points mapped
- [ ] Capability summary created
- [ ] Mental model documented
- [ ] Quick reference created

---

## Notes for AI Agents

When using this template:

1. **Be Systematic**: Work through each phase in order
2. **Document Everything**: Capture all observations, even if they seem minor
3. **Test Incrementally**: Start with simple cases, then increase complexity
4. **Ask Questions**: If something is unclear, note it for clarification
5. **Build Confidence**: Start with low-risk operations before attempting complex workflows
6. **Respect Limits**: Don't overwhelm the server with rapid-fire requests
7. **Verify Assumptions**: Don't assume behavior—test it
8. **Update Continuously**: As you learn more, update your documentation

---

**End of Template**

