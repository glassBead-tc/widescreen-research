# MCPClient Specification

## Overview

The MCPClient in the widescreen-research orchestrator needs to communicate with remote MCP servers (research drones) to coordinate distributed research tasks.

## Requirements

### Core Functionality

1. **Connection Management**: Establish and maintain connections to multiple MCP servers (drones)
2. **Tool Invocation**: Call tools on remote MCP servers with proper parameter passing
3. **Resource Access**: Access resources from remote MCP servers if needed
4. **Error Handling**: Robust error handling with retries and fallbacks
5. **Authentication**: Handle GCP service-to-service authentication for Cloud Run deployments

### Use Cases

1. **Drone Coordination**: Send research tasks to provisioned drone MCP servers
2. **Result Collection**: Gather results from completed drone tasks
3. **Health Monitoring**: Check status and availability of drone servers
4. **Dynamic Scaling**: Connect to newly provisioned drones and disconnect from terminated ones
5. **Human-in-the-Loop Research**: Use elicitation for research validation and quality control
6. **AI-Assisted Analysis**: Use sampling for content analysis and synthesis
7. **Shared Research Workspace**: Use roots for centralized file access and storage

## Technical Design

### SDK Choice

- Use the official `github.com/modelcontextprotocol/go-sdk`
- NOT the unofficial `github.com/mark3labs/mcp-go`

### Transport Layer

- **Primary**: HTTP transport for Cloud Run service-to-service communication
- **Authentication**: GCP Identity Token for authenticated requests
- **Fallback**: Consider stdio transport for local development/testing

### Interface Design

```go
type MCPClient interface {
    // Initialize sets up the client
    Initialize(ctx context.Context) error

    // ConnectToDrone establishes connection to a specific drone MCP server
    ConnectToDrone(ctx context.Context, droneURL string) error

    // CallTool invokes a tool on a specific drone
    CallTool(ctx context.Context, droneURL, toolName string, arguments map[string]interface{}) (*mcp.CallToolResult, error)

    // ListTools gets available tools from a drone
    ListTools(ctx context.Context, droneURL string) (*mcp.ListToolsResult, error)

    // DisconnectFromDrone closes connection to a specific drone
    DisconnectFromDrone(ctx context.Context, droneURL string) error

    // Shutdown closes all connections
    Shutdown() error
}
```

### Official SDK Implementation Patterns

#### Client Creation and Connection

```go
// Create MCP client with implementation info
client := mcp.NewClient(&mcp.Implementation{
    Name:    "widescreen-research-orchestrator",
    Version: "v1.0.0",
}, nil)

// For HTTP transport to Cloud Run services
transport := &mcp.StreamableClientTransport{
    Endpoint:   droneURL,
    HTTPClient: authenticatedHTTPClient,
}

// Connect and get session
session, err := client.Connect(ctx, transport, nil)
if err != nil {
    return fmt.Errorf("failed to connect to drone: %w", err)
}
```

#### Tool Calling Pattern

```go
// Call tool using session
params := &mcp.CallToolParams{
    Name:      toolName,
    Arguments: arguments,
}
result, err := session.CallTool(ctx, params)
if err != nil {
    return nil, fmt.Errorf("tool call failed: %w", err)
}
if result.IsError {
    return nil, fmt.Errorf("tool returned error")
}
```

#### Session Management

```go
// Store sessions by drone URL
sessions := make(map[string]*mcp.ClientSession)

// Clean up on shutdown
for _, session := range sessions {
    session.Close()
}
```

### Connection Pool

- Maintain a pool of connections to active drones
- Implement connection reuse and cleanup
- Handle connection failures gracefully

### Error Handling

- Retry logic for transient failures
- Circuit breaker pattern for failing drones
- Proper error propagation with context

### Authentication Flow

1. Generate GCP Identity Token for target service
2. Include token in HTTP headers
3. Handle token refresh automatically

## Implementation Plan

### Phase 1: Basic MCP Client

1. Replace current stub with official MCP Go SDK client
2. Implement HTTP transport with GCP authentication
3. Basic tool calling functionality
4. Connection management and session handling

### Phase 2: Research Query Elicitation

1. Implement elicitation support for query refinement
2. Create structured schemas for research parameters
3. Build query processing pipeline based on elicited responses
4. Add adaptive follow-up question logic

### Phase 3: AI-Assisted Research Enhancement

1. Implement sampling support for AI analysis
2. Add research synthesis and content analysis capabilities
3. Build query expansion and quality assessment features
4. Create automated report generation using AI

### Phase 4: Shared Research Workspace

1. Implement roots support for shared file access
2. Create research workspace structure and templates
3. Add collaborative file operations for drones
4. Build research artifact archival and discovery

### Phase 5: Advanced Features

1. Retry logic and circuit breakers
2. Metrics and observability
3. Performance optimizations
4. Integration testing with all MCP primitives

## Dependencies

### Required Go Module Updates

First, add the official MCP Go SDK to go.mod:

```bash
go get github.com/modelcontextprotocol/go-sdk@latest
```

### Import Dependencies

- `github.com/modelcontextprotocol/go-sdk/mcp` - Official MCP Go SDK
- `google.golang.org/api/idtoken` - GCP Identity Token generation (for HTTP transport auth)
- Standard Go context and HTTP packages

### Remove Unofficial Dependencies

- Remove `github.com/mark3labs/mcp-go` from go.mod if present
- Update any existing imports to use the official SDK

## Migration from Unofficial to Official SDK

### Pattern Mapping

| Unofficial SDK Pattern | Official SDK Pattern |
|------------------------|---------------------|
| `client.NewSSEMCPClient(url)` | `mcp.NewClient() + mcp.StreamableClientTransport` |
| `client.NewStdioMCPClient(cmd, env, args)` | `mcp.NewClient() + mcp.CommandTransport` |
| `cli.Start(ctx)` | `client.Connect(ctx, transport, nil)` |
| `cli.Initialize(ctx, initRequest)` | Handled automatically in Connect |
| Direct tool calls on client | Tool calls through session |

### Key Differences

1. **Separation of Concerns**: Official SDK separates client creation, transport, and connection
2. **Session-based**: All operations go through session objects, not directly on client
3. **Transport Abstraction**: Transport layer is pluggable and separate from client
4. **Type Safety**: Better type definitions for requests/responses
5. **Error Handling**: More structured error handling with proper MCP error types

### Useful Patterns from Unofficial SDK

These patterns from the unofficial SDK should be implemented using official SDK primitives:

1. **Connection Pooling**: Maintain map of sessions by endpoint
2. **Retry Logic**: Implement retry wrapper around session operations
3. **Authentication**: Use custom HTTP client in StreamableClientTransport
4. **Health Checking**: Periodic ping calls through sessions
5. **Graceful Shutdown**: Proper session cleanup and connection termination

## Advanced MCP Primitives for Research Enhancement

### Elicitation - Human-in-the-Loop Research

**Purpose**: Enable structured human input during research tasks for better outcomes.

**Primary Use Case: Research Query Refinement**
When a user submits an initial research query, use elicitation to gather clarifying information before starting the research process.

**Flow**:

1. User submits initial query: "Research climate change impacts"
2. Orchestrator elicits structured clarification
3. User provides detailed parameters
4. Orchestrator generates refined research plan and coordinates drones

**Implementation**:

```go
// Initial query refinement elicitation
result, err := session.Elicit(ctx, &mcp.ElicitParams{
    Message: "I need to clarify your research request to ensure comprehensive and relevant results.",
    RequestedSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "geographic_scope": {
                "type": "string",
                "enum": []string{"global", "regional", "national", "local"},
                "description": "Geographic scope of research",
            },
            "time_horizon": {
                "type": "string",
                "enum": []string{"historical", "current", "projected", "all"},
                "description": "Time period focus",
            },
            "impact_categories": {
                "type": "array",
                "items": {"type": "string"},
                "enum": []string{"economic", "environmental", "social", "health", "infrastructure"},
                "description": "Types of impacts to focus on",
            },
            "source_preferences": {
                "type": "array",
                "items": {"type": "string"},
                "enum": []string{"peer_reviewed", "government", "industry", "NGO", "news"},
                "description": "Preferred source types",
            },
            "depth_level": {
                "type": "string",
                "enum": []string{"overview", "detailed", "comprehensive"},
                "description": "Level of detail needed",
            },
            "specific_questions": {
                "type": "array",
                "items": {"type": "string"},
                "description": "Any specific questions you want answered",
            },
            "deadline": {
                "type": "string",
                "format": "date",
                "description": "When do you need results by?",
            },
        },
        "required": ["geographic_scope", "time_horizon", "depth_level"],
    },
})

// Process elicitation response to create refined research plan
if result.Action == "accept" {
    refinedQuery := processElicitationResponse(result.Content, originalQuery)
    researchPlan := generateResearchPlan(refinedQuery)
    // Coordinate drones with specific, targeted tasks
}
```

**Additional Elicitation Use Cases**:

- **Mid-research validation**: When drones find conflicting information
- **Quality control**: Human review of findings before proceeding
- **Ethical oversight**: Approval for sensitive research topics
- **Parameter adjustment**: Refining search terms based on initial results

### Sampling - AI-Assisted Research Analysis

**Purpose**: Request AI model completions for content analysis and synthesis.

**Use Cases**:

- **Research synthesis**: Combine findings from multiple drones into coherent insights
- **Content analysis**: Extract key themes and patterns from research data
- **Query expansion**: Generate related research questions or search terms
- **Quality assessment**: Evaluate relevance and credibility of research results
- **Report generation**: Create structured summaries and recommendations

**Implementation**:

```go
// Request AI synthesis of multi-drone research results
result, err := session.CreateMessage(ctx, &mcp.CreateMessageParams{
    Messages: []mcp.SamplingMessage{
        {
            Role: "user",
            Content: mcp.TextContent{
                Text: fmt.Sprintf(`Analyze and synthesize these research findings from multiple sources:

Drone A (Academic Sources): %s
Drone B (Government Data): %s
Drone C (Industry Reports): %s

Please identify:
1. Key trends and patterns
2. Areas of consensus vs. disagreement
3. Research gaps or limitations
4. Actionable insights
5. Recommendations for further research`,
                    droneAResults, droneBResults, droneCResults),
            },
        },
    },
    SystemPrompt: "You are a research analyst specializing in synthesizing information from diverse sources. Provide objective, evidence-based analysis.",
    MaxTokens: 3000,
    ModelPreferences: &mcp.ModelPreferences{
        Hints: []mcp.ModelHint{
            {Name: "claude-3-5-sonnet"}, // Prefer models good at analysis
        },
    },
})

// Use AI to expand research queries
expansionResult, err := session.CreateMessage(ctx, &mcp.CreateMessageParams{
    Messages: []mcp.SamplingMessage{
        {
            Role: "user",
            Content: mcp.TextContent{
                Text: fmt.Sprintf("Given this research query: '%s', suggest 5 related questions that would provide additional valuable context.", originalQuery),
            },
        },
    },
    SystemPrompt: "You are a research strategist. Generate related questions that would enhance understanding of the topic.",
    MaxTokens: 500,
})
```

### Roots - Shared Research Workspace

**Purpose**: Provide drones access to shared file systems for collaborative research.

**Use Cases**:

- **Shared research workspace**: Centralized storage for research artifacts and outputs
- **Document repositories**: Access to existing research documents and datasets
- **Template library**: Standardized formats for reports, citations, and data structures
- **Collaborative storage**: Multiple drones contributing to shared research files
- **Result archival**: Long-term storage of research outputs for future reference

**Implementation**:

```go
// Client (orchestrator) exposes research workspace to drones
client.AddRoots([]mcp.Root{
    {
        URI:  "file:///shared/research/workspace",
        Name: "Research Workspace",
        Description: "Shared workspace for current research projects",
    },
    {
        URI:  "file:///shared/research/templates",
        Name: "Research Templates",
        Description: "Standardized templates for reports and data formats",
    },
    {
        URI:  "file:///shared/research/archive",
        Name: "Research Archive",
        Description: "Historical research results and datasets",
    },
})

// Drone accesses shared resources
resources, err := session.ListResources(ctx, &mcp.ListResourcesParams{})
for _, resource := range resources.Resources {
    if strings.Contains(resource.URI, "templates/report") {
        // Get report template
        content, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
            URI: resource.URI,
        })
        reportTemplate := string(content.Contents[0].(*mcp.TextContent).Text)

        // Use template for standardized reporting
        formattedReport := populateTemplate(reportTemplate, researchFindings)

        // Save completed report back to workspace
        saveToWorkspace(formattedReport, "workspace/reports/climate_analysis.md")
    }
}

// Collaborative file updates
func saveResearchArtifact(session *mcp.ClientSession, data interface{}, path string) error {
    // Convert data to appropriate format
    content := formatResearchData(data)

    // Save to shared workspace via MCP resource operations
    return updateSharedResource(session, path, content)
}
```

**Benefits of Roots Integration**:

- **Consistency**: All drones use same templates and formats
- **Collaboration**: Multiple drones can contribute to shared documents
- **Persistence**: Research artifacts survive individual drone lifecycles
- **Discoverability**: Drones can find and build upon previous research
- **Standardization**: Enforced formats improve result quality and integration

## Advanced MCP Primitives for Research Enhancement

### Elicitation - Human-in-the-Loop Research

**Purpose**: Enable drones to request structured human input during research tasks.

**Use Cases**:

- Research validation when conflicting information is found
- Parameter refinement for ambiguous queries
- Quality control and human review of findings
- Ethical oversight for sensitive research topics

**Implementation**:

```go
// Request human validation of research findings
result, err := session.Elicit(ctx, &mcp.ElicitParams{
    Message: "Found conflicting data about climate change impacts. Which source should be prioritized?",
    RequestedSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "preferred_source": {"type": "string", "enum": []string{"peer_reviewed", "government", "industry", "all"}},
            "reasoning": {"type": "string", "description": "Explain your choice"},
            "confidence": {"type": "number", "minimum": 0, "maximum": 1},
        },
        "required": []string{"preferred_source", "reasoning"},
    },
})
```

### Sampling - AI-Assisted Research Analysis

**Purpose**: Request AI model completions for content analysis and synthesis.

**Use Cases**:

- Analyze research findings and extract key insights
- Synthesize information from multiple drones
- Generate related research questions or search terms
- Evaluate relevance and quality of research results

**Implementation**:

```go
// Request AI analysis of research data
result, err := session.CreateMessage(ctx, &mcp.CreateMessageParams{
    Messages: []mcp.SamplingMessage{
        {
            Role: "user",
            Content: mcp.TextContent{
                Text: fmt.Sprintf("Analyze this research data and identify key trends: %s", researchData),
            },
        },
    },
    SystemPrompt: "You are a research analyst. Identify patterns, insights, and potential gaps in the research.",
    MaxTokens: 2000,
    ModelPreferences: &mcp.ModelPreferences{
        Hints: []mcp.ModelHint{
            {Name: "claude-3-5-sonnet"},
        },
    },
})
```

### Roots - Shared Research Workspace

**Purpose**: Provide drones access to shared file systems for collaborative research.

**Use Cases**:

- Shared research workspace for storing/retrieving artifacts
- Document repositories for content analysis
- Centralized storage for research outputs
- Access to templates and standardized formats

**Implementation**:

```go
// Client exposes research workspace
client.AddRoots([]mcp.Root{
    {
        URI:  "file:///shared/research/workspace",
        Name: "Research Workspace",
    },
    {
        URI:  "file:///shared/research/templates",
        Name: "Research Templates",
    },
})

// Drone accesses shared resources
resources, err := session.ListResources(ctx, &mcp.ListResourcesParams{})
for _, resource := range resources.Resources {
    if strings.Contains(resource.URI, "templates") {
        content, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
            URI: resource.URI,
        })
        // Use template for standardized reporting
    }
}
```

## Configuration

- Environment variables for authentication
- Configurable timeouts and retry policies
- Debug logging support

## Testing Strategy

- Unit tests with mock MCP servers
- Integration tests with real drone deployments
- Load testing for connection pooling
- Error injection testing for resilience
