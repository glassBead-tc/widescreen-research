# Widescreen Research MCP Server (Go)

This repository provides a Go-based Model Context Protocol (MCP) server for orchestrated, large-scale research. The canonical server is implemented in Go under `cmd/widescreen-research-mcp/` and exposes a single MCP tool that drives an elicitation-first workflow and several operations, including distributed research on Google Cloud, analysis, and EXA Websets integration.

Note: A standalone EXA Websets MCP server is vendored under `exa-mcp-server-websets/` and is launched as a subprocess by the orchestrator when you use Websets operations.

## 🚀 Capabilities

- Elicitation-driven session setup (topic, researcher_count, depth, output format, timeouts, priority)
- Distributed research orchestration on Google Cloud (Cloud Run + Pub/Sub + Firestore)
- Data analysis and report generation (Markdown report plus structured report object)
- Sequential-thinking analysis using a Claude agent (mocked if `CLAUDE_API_KEY` is not set)
- GCP provisioning helpers for Cloud Run, Pub/Sub, and Firestore
- EXA Websets pipeline orchestration and direct Websets tool passthrough

### MCP Surface

- Tool name: `widescreen_research`
- Arguments:
  - `operation` string
  - `session_id` string (elicitation session)
  - `parameters_json` string (JSON-encoded map)
  - `elicitation_answers_json` string (JSON-encoded map)

The server currently exposes tools only (no `resources`/`prompts`) due to the `mcp-go` server API in use.

### Operations

- orchestrate-research: Provision and coordinate multiple research “drones” on Cloud Run; collect results via Pub/Sub; generate a Markdown report and structured report
- sequential-thinking: Produce structured thought steps and a recommendation for a given problem/context
- gcp-provision: Provision GCP resources
  - cloud_run: deploy lightweight services
  - pubsub: create topics and optional subscriptions
  - firestore: create collections
- analyze-findings: Analyze drone results to extract insights, patterns, statistics, and visualizations
- websets-orchestrate: Full EXA Websets pipeline (create → wait → list items → publish to Pub/Sub)
- websets-call: Direct passthrough to EXA’s `websets_manager` tool with custom arguments

Implementation references:

```1:36:cmd/widescreen-research-mcp/server/server.go
// WidescreenResearchServer is the main MCP server that provides widescreen research capabilities
type WidescreenResearchServer struct { ... }
```

```61:114:cmd/widescreen-research-mcp/server/server.go
// registerWidescreenResearchTool registers the main tool that handles all operations
func (s *WidescreenResearchServer) registerWidescreenResearchTool() { ... }
```

```247:297:cmd/widescreen-research-mcp/server/server.go
// registerOperations registers all available operations
func (s *WidescreenResearchServer) registerOperations() { ... }
```

## 📋 Prerequisites

- Go 1.23+
- Google Cloud project and credentials (Cloud Run, Pub/Sub, Firestore)
- For Websets operations: EXA API key and either the EXA Websets MCP binary or Node.js 18+ for fallback
- Optional: `CLAUDE_API_KEY` for richer sequential thinking and report generation

## 🔧 Setup

1) Clone and download deps

```bash
git clone https://github.com/your-org/widescreen-research.git
cd widescreen-research
go mod download
```

2) Build the server

```bash
go build -o widescreen-research ./cmd/widescreen-research-mcp
```

3) Install/prepare EXA Websets MCP (only if you plan to use Websets)

- Recommended: install the published binary globally:

```bash
npm i -g exa-websets-mcp-server
```

- Or build the vendored server (fallback launched as `node ./build/index.js`):

```bash
cd exa-mcp-server-websets
npm ci && npm run build
cd -
```

## ⚙️ Environment Variables

- `GOOGLE_CLOUD_PROJECT` (required): your GCP project ID
- `GOOGLE_CLOUD_REGION` (optional, default `us-central1`)
- `EXA_API_KEY` (required for Websets operations)
- `CLAUDE_API_KEY` (optional)

## ▶️ Run

Run the compiled binary. MCP clients (e.g., Claude Desktop) communicate over stdio.

```bash
./widescreen-research
```

### Claude Desktop configuration

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "widescreen-research": {
      "command": "/absolute/path/to/widescreen-research",
      "env": {
        "GOOGLE_CLOUD_PROJECT": "your-project-id",
        "GOOGLE_CLOUD_REGION": "us-central1",
        "EXA_API_KEY": "your-exa-api-key",
        "CLAUDE_API_KEY": "optional"
      }
    }
  }
}
```

### Inspect with MCP Inspector

```bash
npx @modelcontextprotocol/inspector /absolute/path/to/widescreen-research
```

## 🎯 Usage Examples

### Start elicitation

```json
{
  "tool": "widescreen_research",
  "arguments": { "operation": "start" }
}
```

You’ll receive an `elicitation` response with `session_id` and questions. Keep sending answers via `elicitation_answers_json` until the server returns `type=ready`.

### Orchestrate research

```json
{
  "tool": "widescreen_research",
  "arguments": {
    "operation": "orchestrate-research",
    "session_id": "<from-elicitation>",
    "parameters_json": "{}"
  }
}
```

Result contains `report_url` (e.g., `reports/report_<session>.md`) and structured `report_data`.

### Sequential thinking

```json
{
  "tool": "widescreen_research",
  "arguments": {
    "operation": "sequential-thinking",
    "parameters_json": "{\"problem\":\"Complex problem\",\"context\":\"Optional\",\"max_steps\":10}"
  }
}
```

### GCP provision (Cloud Run)

```json
{
  "tool": "widescreen_research",
  "arguments": {
    "operation": "gcp-provision",
    "parameters_json": "{\"resource_type\":\"cloud_run\",\"count\":1,\"region\":\"us-central1\",\"config\":{\"image\":\"gcr.io/cloudrun/hello\",\"cpu\":\"1000m\",\"memory\":\"512Mi\"}}"
  }
}
```

### Analyze findings

```json
{
  "tool": "widescreen_research",
  "arguments": {
    "operation": "analyze-findings",
    "parameters_json": "{\"analysis_type\":\"comprehensive\",\"data\":[] }"
  }
}
```

### Websets pipeline (EXA)

```json
{
  "tool": "widescreen_research",
  "arguments": {
    "operation": "websets-orchestrate",
    "parameters_json": "{\"topic\":\"AI safety\",\"result_count\":100}"
  }
}
```

### Direct Websets call (passthrough)

```json
{
  "tool": "widescreen_research",
  "arguments": {
    "operation": "websets-call",
    "parameters_json": "{\"operation\":\"create_webset\",\"webset\":{\"searchQuery\":\"AI safety\"}}"
  }
}
``;

## 📦 Outputs

- ElicitationResponse: `type` (elicitation|ready), questions, `session_id`, and derived `config`
- ResearchResult: `status`, `report_url`, `report_data` (structured), and `metrics`

Key schema types are defined in `cmd/widescreen-research-mcp/schemas/schemas.go`.

## 🏗️ Architecture

```
Claude (client) ⇄ widescreen-research (Go MCP server) ⇄ Orchestrator
                                              ├─ GCP: Cloud Run, Pub/Sub, Firestore
                                              └─ Subprocess MCP: EXA Websets server
```

## 📚 Repository Map (relevant)

- `cmd/widescreen-research-mcp/`
  - `server/` MCP tool and operation registration, elicitation manager
  - `orchestrator/` GCP provisioning, research coordination, EXA Websets client
  - `operations/` `gcp_provisioner.go`, `data_analyzer.go`, `sequential_thinking.go`
  - `schemas/` Request/response/report types
- `exa-mcp-server-websets/` Vendored EXA Websets MCP server (Node/TypeScript)
- `pkg/mcp/` Alternate MCP surface for coordinator (drone fleet mgmt; separate from this server)

## ⚠️ Notes & Limitations

- Only tools are exposed (no MCP `resources` or `prompts`) with the current `mcp-go` server API in use.
- Websets operations require `EXA_API_KEY` and the EXA Websets server to be available on PATH (or Node fallback built locally).

## 📄 License

Apache 2.0 License - see LICENSE.

