package resources

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// GuideResource handles embedded guide resources
type GuideResource struct {
	guides map[string]string
}

// NewGuideResource creates a new guide resource handler
func NewGuideResource() *GuideResource {
	return &GuideResource{
		guides: map[string]string{
			"main": mainGuide,
			"websets": websetsWorkflow,
			"orchestration": orchestrationWorkflow,
			"quickstart": quickstartGuide,
		},
	}
}

// RegisterWithServer registers guide resources with the MCP server
func (g *GuideResource) RegisterWithServer(server *mcpserver.MCPServer) {
	// Note: Current mcp-go doesn't support resources directly
	// This is preparation for when it does, or for migration to go-sdk
	// For now, we'll expose guides through a tool
}

// GetGuide returns guide content by name
func (g *GuideResource) GetGuide(name string) (string, error) {
	guide, exists := g.guides[name]
	if !exists {
		return "", fmt.Errorf("guide not found: %s", name)
	}
	return guide, nil
}

// ListGuides returns all available guide names
func (g *GuideResource) ListGuides() []string {
	names := make([]string, 0, len(g.guides))
	for name := range g.guides {
		names = append(names, name)
	}
	return names
}

// HandleGuideRequest handles MCP resource requests for guides
func (g *GuideResource) HandleGuideRequest(ctx context.Context, uri string) (mcp.Content, error) {
	// Parse URI to extract guide name
	// Format: embedded://guides/{name}
	
	guideName := extractGuideName(uri)
	guide, err := g.GetGuide(guideName)
	if err != nil {
		return nil, err
	}
	
	return &mcp.TextContent{Text: guide}, nil
}

func extractGuideName(uri string) string {
	// Simple extraction - in production would use proper URL parsing
	if len(uri) > 17 && uri[:17] == "embedded://guides/" {
		return uri[17:]
	}
	return "main"
}

const mainGuide = `# Widescreen Research System Guide

## Quick Start

The widescreen research system orchestrates distributed research using multiple approaches:

### Available Operations

| Operation | Description | Complexity | Time |
|-----------|-------------|------------|------|
| **websets-orchestrate** | EXA websets research | Low | 10-15 min |
| **orchestrate-research** | Full GCP orchestration | High | 30-60 min |
| **sequential-thinking** | Step-by-step reasoning | Medium | 5-10 min |
| **analyze-findings** | Analyze collected data | Low | 1-2 min |

### 1. Simple Websets Research (Fastest)

For quick research using EXA's content websets:

` + "```json" + `
{
  "operation": "websets-orchestrate",
  "parameters_json": "{\"topic\": \"AI safety research\", \"result_count\": 20}"
}
` + "```" + `

### 2. Full Orchestration (Most Comprehensive)

Start with elicitation for guided research:

` + "```json" + `
{
  "operation": "start",
  "session_id": ""
}
` + "```" + `

Then follow the interactive questions to configure your research.

### 3. Direct Analysis

For analyzing existing data:

` + "```json" + `
{
  "operation": "analyze-findings",
  "parameters_json": "{\"data\": [...], \"analysis_type\": \"summary\"}"
}
` + "```" + `

## Understanding the System

- **Websets**: Fast, focused research via EXA API
- **Orchestration**: Comprehensive research with GCP drones
- **Elicitation**: Guided configuration through Q&A
- **Analysis**: Pattern recognition and synthesis

## Need Help?

- Try "getting_started" for first-time users
- Check "websets" workflow for EXA research
- See "orchestration" workflow for full process
`

const websetsWorkflow = `# Websets Research Workflow

## Overview

Websets provide a streamlined research approach using EXA's content aggregation.

## Step-by-Step Process

### 1. Create Webset

` + "```json" + `
{
  "operation": "websets-orchestrate",
  "parameters_json": "{
    \"topic\": \"quantum computing breakthroughs 2024\",
    \"result_count\": 50
  }"
}
` + "```" + `

### 2. What Happens Behind the Scenes

1. **Webset Creation** (1-2 min)
   - Sends request to EXA API
   - Creates content webset with your topic

2. **Status Polling** (10-15 min)
   - Checks status every 10 seconds
   - Waits for completion

3. **Content Retrieval**
   - Fetches all matched items
   - Extracts titles, URLs, content

4. **Publishing to Pub/Sub**
   - Creates topic for results
   - Publishes items for processing

### 3. Direct Websets Calls

For custom operations, use websets-call:

` + "```json" + `
{
  "operation": "websets-call",
  "parameters_json": "{
    \"operation\": \"create_webset\",
    \"webset\": {
      \"searchQuery\": \"your query\",
      \"advanced\": {
        \"resultCount\": 100,
        \"domainFilter\": [\"edu\", \"org\"]
      }
    }
  }"
}
` + "```" + `

## Benefits

✅ No GCP setup required
✅ Fast results (10-15 minutes)
✅ Lower complexity
✅ Good for focused research

## Limitations

❌ Less comprehensive than full orchestration
❌ Limited to EXA's index
❌ No custom drone logic
`

const orchestrationWorkflow = `# Full Orchestration Workflow

## Overview

The complete orchestration process for comprehensive research.

## Phases

### Phase 1: Elicitation

Start the conversation:

` + "```json" + `
{
  "operation": "start",
  "session_id": ""
}
` + "```" + `

Answer questions about:
- Research scope
- Depth required
- Time constraints
- Output format

### Phase 2: Configuration

System determines:
- Number of drones needed
- GCP resources required
- Research strategies
- Aggregation approach

### Phase 3: Provisioning

Automatic setup of:
- Cloud Run services
- Pub/Sub topics
- Firestore collections
- IAM permissions

### Phase 4: Research Execution

Parallel execution:
- Drones deployed
- Research conducted
- Results published
- Real-time aggregation

### Phase 5: Analysis

Final processing:
- Pattern recognition
- Data synthesis
- Report generation
- Confidence scoring

## Example Timeline

- **0-5 min**: Elicitation
- **5-10 min**: Provisioning
- **10-40 min**: Research
- **40-45 min**: Analysis
- **45 min**: Results ready

## When to Use

✅ Complex, multi-faceted topics
✅ Need comprehensive coverage
✅ Have GCP resources available
✅ Time for deeper investigation

## Requirements

- GCP Project configured
- Sufficient quota
- API keys set
- 30-60 minutes available
`

const quickstartGuide = `# Quick Start Guide

## First Time Setup

### 1. Check Connection

Your MCP connection is working if you see this!

### 2. Try Your First Research

**Simplest approach - Websets:**

` + "```json" + `
{
  "operation": "websets-orchestrate",
  "parameters_json": "{\"topic\": \"latest AI news\", \"result_count\": 10}"
}
` + "```" + `

### 3. Understanding Results

Results include:
- **sessionId**: Unique identifier
- **status**: Current state
- **reportData**: The actual findings
- **metrics**: Performance data

### Common Operations Examples

**Research a Company:**
` + "```json" + `
{
  "operation": "websets-orchestrate",
  "parameters_json": "{\"topic\": \"OpenAI company updates 2024\", \"result_count\": 20}"
}
` + "```" + `

**Academic Research:**
` + "```json" + `
{
  "operation": "websets-orchestrate",
  "parameters_json": "{\"topic\": \"machine learning papers arxiv 2024\", \"result_count\": 30}"
}
` + "```" + `

**Market Analysis:**
` + "```json" + `
{
  "operation": "websets-orchestrate",
  "parameters_json": "{\"topic\": \"EV market trends 2024\", \"result_count\": 40}"
}
` + "```" + `

## Tips

1. Start with websets for quick results
2. Use elicitation for complex research
3. Check guides for detailed workflows
4. Monitor progress through status

## Troubleshooting

- **"GOOGLE_CLOUD_PROJECT not set"**: Using websets-orchestrate avoids this
- **Timeout errors**: Normal for 10-15 min operations
- **No results**: Try broader search terms
`