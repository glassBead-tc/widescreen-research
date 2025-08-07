# Widescreen Research MCP Server

A powerful Model Context Protocol (MCP) server that enables horizontal research at scale using distributed research drones on Google Cloud Platform. This server provides comprehensive research capabilities through parallel processing, allowing agents to provision and orchestrate multiple research drones to conduct research on various topics simultaneously.

## 🌟 Overview

The Widescreen Research MCP server implements a unique architecture where:

1. **Claude Code** calls the widescreen-research server
2. The server uses **elicitation** to qualify the user's research requirements
3. An **orchestrator agent** (with bidirectional MCP capabilities) manages the research process
4. Multiple **research drones** are provisioned on GCP Cloud Run
5. Results are collected through a **queue system** and processed into comprehensive reports

## 🚀 Features

### Core Capabilities
- **Elicitation-Based Qualification**: Interactive questioning to understand research needs
- **Distributed Research**: Provision 1-100 research drones in parallel
- **Sequential Thinking**: Advanced reasoning capabilities for complex problems
- **GCP Resource Management**: Automated provisioning of Cloud Run, Pub/Sub, and Firestore
- **Data Analysis**: Comprehensive analysis of research findings with pattern detection
- **Report Generation**: AI-powered report generation from collected data

### Operations
- `orchestrate-research`: Main operation that coordinates the entire research process
- `sequential-thinking`: Performs step-by-step reasoning for complex problems
- `gcp-provision`: Provisions GCP resources (Cloud Run, Pub/Sub, Firestore)
- `analyze-findings`: Analyzes collected research data for patterns and insights

## 📋 Prerequisites

- Go 1.21 or later
- Google Cloud Platform account with:
  - Cloud Run API enabled
  - Pub/Sub API enabled
  - Firestore API enabled
  - Appropriate IAM permissions
- MCP-compatible client (Claude Desktop, etc.)
- (Optional) Claude API key for enhanced AI capabilities

## 🔧 Installation

1. **Clone the repository**:
```bash
git clone https://github.com/your-org/widescreen-research-mcp.git
cd widescreen-research-mcp
```

2. **Install dependencies**:
```bash
go mod download
```

3. **Set up environment variables**:
```bash
export GOOGLE_CLOUD_PROJECT="your-project-id"
export GOOGLE_CLOUD_REGION="us-central1"  # or your preferred region
export CLAUDE_API_KEY="your-claude-api-key"  # Optional
export ORCHESTRATOR_URL="http://localhost:8080"  # For local development
```

4. **Build the server**:
```bash
go build -o widescreen-research ./cmd/widescreen-research-mcp
```

## 🎯 Usage

### Running the Server

```bash
./widescreen-research
```

### MCP Client Configuration

#### Claude Desktop

Add to your `claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "widescreen-research": {
      "command": "/path/to/widescreen-research",
      "env": {
        "GOOGLE_CLOUD_PROJECT": "your-project-id",
        "GOOGLE_CLOUD_REGION": "us-central1",
        "CLAUDE_API_KEY": "your-claude-api-key"
      }
    }
  }
}
```

### Using the Tool

The server provides a single main tool: `widescreen-research`

#### Starting a Research Session

```json
{
  "tool": "widescreen-research",
  "arguments": {
    "operation": "start"
  }
}
```

This initiates the elicitation process with questions like:
- What would you like to perform research on?
- How many researchers do you want to provision?
- What level of research depth do you need?
- Do you have any pre-orchestrated workflows?

#### Executing Research

After elicitation is complete:

```json
{
  "tool": "widescreen-research",
  "arguments": {
    "operation": "orchestrate-research",
    "session_id": "session-uuid-here",
    "parameters": {
      "additional_config": "optional"
    }
  }
}
```

#### Other Operations

**Sequential Thinking**:
```json
{
  "tool": "widescreen-research",
  "arguments": {
    "operation": "sequential-thinking",
    "parameters": {
      "problem": "Complex problem to analyze",
      "context": "Additional context",
      "max_steps": 10
    }
  }
}
```

**GCP Resource Provisioning**:
```json
{
  "tool": "widescreen-research",
  "arguments": {
    "operation": "gcp-provision",
    "parameters": {
      "resource_type": "cloud_run",
      "count": 5,
      "region": "us-central1",
      "config": {
        "cpu": "1000m",
        "memory": "512Mi"
      }
    }
  }
}
```

**Data Analysis**:
```json
{
  "tool": "widescreen-research",
  "arguments": {
    "operation": "analyze-findings",
    "parameters": {
      "data": [...],
      "analysis_type": "comprehensive"
    }
  }
}
```

## 🏗️ Architecture

```
┌─────────────────┐
│  Claude Code    │
│  (MCP Client)   │
└────────┬────────┘
         │
         v
┌─────────────────┐     ┌─────────────────┐
│  Widescreen     │────>│  Orchestrator   │
│  Research MCP   │     │  (MCP Client)   │
│  Server         │<────│                 │
└─────────────────┘     └────────┬────────┘
                                 │
                    ┌────────────┴────────────┐
                    │                         │
                    v                         v
           ┌─────────────────┐      ┌─────────────────┐
           │  Research       │      │  External MCP   │
           │  Drones         │      │  Servers        │
           │  (Cloud Run)    │      │  (Exa, etc.)   │
           └─────────────────┘      └─────────────────┘
                    │
                    v
           ┌─────────────────┐
           │  Result Queue   │
           │  (Pub/Sub)      │
           └─────────────────┘
```

### Components

1. **MCP Server**: Main server that handles the widescreen-research tool
2. **Elicitation Manager**: Manages user qualification through questions
3. **Orchestrator**: Bidirectional MCP agent that coordinates research
4. **Research Drones**: Lightweight Cloud Run containers that perform research
5. **Queue System**: Pub/Sub-based queue for collecting results
6. **Report Generator**: AI-powered report generation from collected data

## 🔍 Research Process

1. **Elicitation Phase**:
   - User initiates research request
   - Server asks qualification questions
   - Configuration is built based on answers

2. **Provisioning Phase**:
   - Orchestrator provisions requested number of drones
   - Cloud Run services are deployed with appropriate resources
   - Pub/Sub topics and subscriptions are created

3. **Research Phase**:
   - Drones receive research instructions
   - Parallel research is conducted
   - Results are sent to the queue

4. **Collection Phase**:
   - Orchestrator monitors queue for results
   - Data is aggregated as it arrives
   - Progress is tracked

5. **Analysis Phase**:
   - Collected data is analyzed for patterns
   - Insights are extracted
   - Statistics are calculated

6. **Report Phase**:
   - AI generates comprehensive report
   - Results are structured and formatted
   - Report is returned to user

## 🛠️ Configuration

### Environment Variables

- `GOOGLE_CLOUD_PROJECT`: GCP project ID (required)
- `GOOGLE_CLOUD_REGION`: Default region for resources (default: us-central1)
- `CLAUDE_API_KEY`: Claude API key for AI capabilities (optional)
- `ORCHESTRATOR_URL`: URL for orchestrator callbacks
- `EXA_MCP_URL`: URL for Exa research MCP server (optional)
- `WEB_RESEARCH_MCP_URL`: URL for web research MCP server (optional)

### Research Configuration

The elicitation process allows configuration of:
- **Research Topic**: What to research
- **Researcher Count**: 1-100 drones
- **Research Depth**: basic, standard, or deep
- **Output Format**: structured_json, markdown_report, executive_summary, raw_data
- **Timeout**: Maximum time for research completion
- **Priority Level**: low (cost-optimized), normal (balanced), high (performance-optimized)

## 📊 Monitoring and Logging

The server provides comprehensive logging and monitoring:

- Session status tracking
- Drone health monitoring
- Queue depth metrics
- Research progress updates
- Error tracking and reporting

## 🚧 Deployment

### Local Development

```bash
# Run the server locally
go run ./cmd/widescreen-research-mcp
```

### Google Cloud Deployment

1. **Build container**:
```bash
docker build -t gcr.io/YOUR_PROJECT_ID/widescreen-research-mcp .
docker push gcr.io/YOUR_PROJECT_ID/widescreen-research-mcp
```

2. **Deploy to Cloud Run**:
```bash
gcloud run deploy widescreen-research-mcp \
  --image gcr.io/YOUR_PROJECT_ID/widescreen-research-mcp \
  --platform managed \
  --region us-central1 \
  --set-env-vars GOOGLE_CLOUD_PROJECT=YOUR_PROJECT_ID
```

## 🔒 Security

- Uses Google Cloud IAM for authentication
- Supports Workload Identity for service accounts
- Implements least-privilege access patterns
- Automatic cleanup of resources after research

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

Apache 2.0 License - see LICENSE file for details.

---

**Built for comprehensive research at scale • Powered by MCP and Google Cloud Platform**